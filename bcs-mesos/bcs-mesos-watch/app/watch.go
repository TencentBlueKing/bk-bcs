/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package app

import (
	rd "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	commtype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/cluster/etcd"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/cluster/mesos"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/service"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/storage"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
)

var (
	storageState = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "bkbcs_datawatch",
		Subsystem: "mesos",
		Name:      "storage_state",
		Help:      "The state of bcs-storage that watch detected",
	})
	clusterState = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "bkbcs_datawatch",
		Subsystem: "mesos",
		Name:      "cluster_state",
		Help:      "The state of mesos watch cluster state",
	})
	roleState = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "bkbcs_datawatch",
		Subsystem: "mesos",
		Name:      "role_state",
		Help:      "The role of meoss watch",
	})
)

const (
	roleStateMaster float64 = 1
	roleStateSlave  float64 = 0
	stateErr        float64 = 0
	stateOK         float64 = 1
	stateRegisteErr float64 = 2
	stateDiscvErr   float64 = 3
	stateNotRun     float64 = 4
	stateRoleErr    float64 = 5
)

func runMetric(cfg *types.CmdConfig) {

	blog.Infof("run metric: port(%d)", cfg.MetricPort)
	prometheus.MustRegister(storageState)
	prometheus.MustRegister(clusterState)
	prometheus.MustRegister(roleState)
	http.Handle("/metrics", promhttp.Handler())
	addr := cfg.Address + ":" + strconv.Itoa(int(cfg.MetricPort))
	go http.ListenAndServe(addr, nil)

	blog.Infof("run metric ok")
}

//Run running watch
func Run(cfg *types.CmdConfig) error {

	if cfg.ClusterID == "" {
		blog.Error("datawatcher cluster unknown")
		return fmt.Errorf("datawatcher cluster unknown")
	}
	blog.Info("datawatcher run for cluster %s", cfg.ClusterID)

	//create root context for exit
	rootCxt, rootCancel := context.WithCancel(context.Background())
	interupt := make(chan os.Signal, 10)
	signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	signalCxt, _ := context.WithCancel(rootCxt)
	go handleSysSignal(signalCxt, interupt, rootCancel)

	runMetric(cfg)

	//create storage
	ccStorage, ccErr := storage.NewCCStorage(cfg)
	if ccErr != nil {
		blog.Error("Create CCStorage Err: %s", ccErr.Error())
		return ccErr
	}
	var DCHosts []string
	ccStorage.SetDCAddress(DCHosts)
	//servermetric.SetDCStatus(false)
	clusterState.Set(stateErr)

	ccCxt, _ := context.WithCancel(rootCxt)
	go RefreshDCHost(ccCxt, cfg, ccStorage)
	time.Sleep(2 * time.Second)

	retryNum := 0
	for {
		if retryNum == 10 {
			blog.Errorf("storage address is empty, datawatcher exited")
			os.Exit(1)
		}

		if ccStorage.GetDCAddress() == "" {
			blog.Warn("storage address is empty, mesos datawatcher cannot run")
			time.Sleep(2 * time.Second)
		} else {
			break
		}
		retryNum++
	}

	ccStorage.Run(ccCxt)
	rdCxt, _ := context.WithCancel(rootCxt)

	go func() {
		retry, rdErr := registerZkEndpoints(rdCxt, cfg)
		for retry == true {
			if rdErr != nil {
				blog.Error("registerZkEndpoints err: %s", rdErr.Error())
			}
			time.Sleep(3 * time.Second)
			blog.Info("retry registerZkEndpoints...")
			retry, rdErr = registerZkEndpoints(rdCxt, cfg)
		}
		if rdErr != nil {
			blog.Error("registerZkEndpoints err: %s, and exit", rdErr.Error())
		}
	}()

	// watch netservice servers from ZK.
	netservice, netserviceZKRD, err := GetNetService(cfg)
	if err != nil {
		blog.Error("watch netservice servers failed, %+v", err)
		return err
	}

	blog.Info("after storage created, to run server...")
	retry, rdErr := runServer(rdCxt, cfg, ccStorage, netservice)
	for retry == true {
		if rdErr != nil {
			blog.Error("run server err: %s", rdErr.Error())
		}
		time.Sleep(3 * time.Second)
		blog.Info("retry run server...")
		retry, rdErr = runServer(rdCxt, cfg, ccStorage, netservice)
	}
	if rdErr != nil {
		blog.Error("run server err: %s", rdErr.Error())
	}

	blog.Info("to cancel root after runServer returned")
	netserviceZKRD.Stop()
	rootCancel()

	return rdErr
}

func handleSysSignal(exitCxt context.Context, signalChan <-chan os.Signal, cancel context.CancelFunc) {
	select {
	case s := <-signalChan:
		blog.V(3).Infof("watch Get singal %s, exit!", s.String())
		cancel()
		time.Sleep(2 * time.Second)
		return
	case <-exitCxt.Done():
		blog.V(3).Infof("Signal Handler asked to exit")
		return
	}
}

func runServer(rdCxt context.Context, cfg *types.CmdConfig, storage storage.Storage, netservice *service.InnerService) (bool, error) {

	// servermetric.SetClusterStatus(false, "begin run server")
	// servermetric.SetRole(metric.SlaveRole)
	clusterState.Set(stateErr)
	roleState.Set(roleStateSlave)

	regDiscv := rd.NewRegDiscoverEx(cfg.RegDiscvSvr, time.Second*10)
	if regDiscv == nil {
		// servermetric.SetClusterStatus(false, "register error")
		clusterState.Set(stateRegisteErr)
		return false, fmt.Errorf("NewRegDiscover(%s) return nil", cfg.RegDiscvSvr)
	}
	blog.Info("NewRegDiscover(%s) succ", cfg.RegDiscvSvr)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("regDisv start error(%s)", err.Error())
		// servermetric.SetClusterStatus(false, "register error:"+err.Error())
		clusterState.Set(stateRegisteErr)
		return false, err
	}
	blog.Info("RegDiscover start succ")

	blog.Infof("ApplicationThreadNum: %d, TaskgroupThreadNum: %d, ExportserviceThreadNum: %d",
		cfg.ApplicationThreadNum, cfg.TaskgroupThreadNum, cfg.ExportserviceThreadNum)

	host, err := os.Hostname()
	if err != nil {
		blog.Error("mesoswatcher get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	var regInfo commtype.MesosDataWatchServInfo
	regInfo.ServerInfo.Cluster = cfg.ClusterID
	regInfo.ServerInfo.IP = cfg.Address
	regInfo.ServerInfo.Port = 0
	regInfo.ServerInfo.MetricPort = cfg.MetricPort
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Scheme = cfg.ServerSchem
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	data, err := json.Marshal(regInfo)
	key := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDATAWATCH + "/" + cfg.ClusterID + "/" + cfg.Address
	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDATAWATCH + "/" + cfg.ClusterID

	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("RegisterService(%s) error(%s)", key, err.Error())
		//servermetric.SetClusterStatus(false, "register error:"+err.Error())
		clusterState.Set(stateRegisteErr)
		regDiscv.Stop()
		return true, err
	}
	blog.Info("RegisterService(%s:%s) succ", key, data)

	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("DiscoverService(%s) error(%s)", discvPath, err.Error())
		//servermetric.SetClusterStatus(false, "discove error:"+err.Error())
		clusterState.Set(stateDiscvErr)
		regDiscv.Stop()
		return true, err
	}
	blog.Info("DiscoverService(%s) succ", discvPath)

	// init, slave, master
	var clusterCancel context.CancelFunc
	var currCluster cluster.Cluster
	clusterCancel = nil
	currCluster = nil

	appRole := "slave"
	tick := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-rdCxt.Done():
			blog.V(3).Infof("runServer asked to exit")
			regDiscv.Stop()
			if currCluster != nil {
				currCluster.Stop()
				currCluster = nil
			}
			if clusterCancel != nil {
				clusterCancel()
			}
			return false, nil
		case <-tick.C:
			blog.V(3).Infof("tick: runServer is alive, current goroutine num (%d)", runtime.NumGoroutine())
			if currCluster != nil && currCluster.GetClusterStatus() != "running" {
				blog.V(3).Infof("tick: current cluster status(%s), to rebuild cluster", currCluster.GetClusterStatus())
				// servermetric.SetClusterStatus(false, "cluster status not running")
				clusterState.Set(stateNotRun)
				regDiscv.Stop()
				if currCluster != nil {
					currCluster.Stop()
					currCluster = nil
				}
				if clusterCancel != nil {
					clusterCancel()
				}
				return true, nil
			}
		case event := <-discvEvent:
			blog.Info("get discover event")
			if event.Err != nil {
				blog.Error("get discover event err:%s", event.Err.Error())
				// servermetric.SetClusterStatus(false, "get discove error:"+event.Err.Error())
				clusterState.Set(stateDiscvErr)
				regDiscv.Stop()
				if currCluster != nil {
					currCluster.Stop()
					currCluster = nil
				}
				if clusterCancel != nil {
					clusterCancel()
				}
				return true, event.Err
			}

			currRole := ""
			for i, server := range event.Server {
				blog.Info("get discover event: server[%d]: %s %s", i, event.Key, server)
				if currRole == "" && i == 0 && server == string(data) {
					currRole = "master"
					// servermetric.SetRole(metric.MasterRole)
					// servermetric.SetClusterStatus(true, "master run ok")
					roleState.Set(roleStateMaster)
					clusterState.Set(stateOK)
				}
				if currRole == "" && i != 0 && server == string(data) {
					currRole = "slave"
					// servermetric.SetRole(metric.SlaveRole)
					// servermetric.SetClusterStatus(true, "slave run ok")
					roleState.Set(roleStateSlave)
					clusterState.Set(stateOK)
				}
			}
			if currRole == "" {
				blog.Infof("get discover event, server list len(%d), but cannot find myself", len(event.Server))
				regDiscv.Stop()
				if currCluster != nil {
					currCluster.Stop()
					currCluster = nil
				}
				if clusterCancel != nil {
					clusterCancel()
				}
				//servermetric.SetClusterStatus(false, "role error")
				clusterState.Set(stateRoleErr)

				return true, fmt.Errorf("currRole is nil")
			}

			blog.Info("get discover event, curr role: %s", currRole)

			if currRole != appRole {
				blog.Info("role changed: from %s to %s", appRole, currRole)
				appRole = currRole
				if appRole == "master" {
					blog.Info("become to master: to new and run cluster...")
					var cluster cluster.Cluster
					if cfg.StoreDriver == "etcd" {
						cluster = etcd.NewEtcdCluster(cfg, storage, netservice)
					} else {
						cluster = mesos.NewMesosCluster(cfg, storage, netservice)
					}

					if cluster == nil {
						blog.Error("Create Cluster Error.")
						regDiscv.Stop()
						//servermetric.SetClusterStatus(false, "master create cluster error")
						clusterState.Set(stateRoleErr)
						return true, fmt.Errorf("cluster create failed")
					}
					currCluster = cluster
					clusterCxt, cancel := context.WithCancel(rdCxt)
					clusterCancel = cancel
					go cluster.Run(clusterCxt)
				} else {
					blog.Infof("become to slave: to cancel cluster...")
					if currCluster != nil {
						currCluster.Stop()
						currCluster = nil
					}
					if clusterCancel != nil {
						clusterCancel()
						clusterCancel = nil
					}
				}
			} // end role change
		} // end select
	} // end for

}

// GetNetService returns netservice InnerService object for discovery.
func GetNetService(cfg *types.CmdConfig) (*service.InnerService, *rd.RegDiscover, error) {
	discovery := rd.NewRegDiscoverEx(cfg.NetServiceZK, 10*time.Second)
	if err := discovery.Start(); err != nil {
		return nil, nil, fmt.Errorf("get netservice from ZK failed, %+v", err)
	}

	// zknode: bcs/services/endpoints/netservice
	path := fmt.Sprintf("%s/%s", commtype.BCS_SERV_BASEPATH, commtype.BCS_MODULE_NETSERVICE)
	eventChan, err := discovery.DiscoverService(path)
	if err != nil {
		discovery.Stop()
		return nil, nil, fmt.Errorf("discover netservice failed, %+v", err)
	}

	netService := service.NewInnerService(commtype.BCS_MODULE_NETSERVICE, eventChan)
	go netService.Watch(cfg)

	return netService, discovery, nil
}

//RefreshDCHost update bcs-storage info
func RefreshDCHost(rfCxt context.Context, cfg *types.CmdConfig, storage storage.Storage) {
	blog.Info("mesos data watcher to refresh DCHost ...")
	// register service
	regDiscv := rd.NewRegDiscoverEx(cfg.RegDiscvSvr, time.Second*10)
	if regDiscv == nil {
		blog.Error("NewRegDiscover(%s) return nil", cfg.RegDiscvSvr)
		return
	}
	blog.Info("NewRegDiscover(%s) succ", cfg.RegDiscvSvr)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("regDiscv start error(%s)", err.Error())
		return
	}
	blog.Info("RegDiscover start succ")

	defer regDiscv.Stop()

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_STORAGE
	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("DiscoverService(%s) error(%s)", discvPath, err.Error())
		return
	}
	blog.Info("DiscoverService(%s) succ", discvPath)

	tick := time.NewTicker(120 * time.Second)
	for {
		select {
		case <-tick.C:
			blog.Info("refresh DCHost is running")
			continue
		case <-rfCxt.Done():
			blog.V(3).Infof("refresh DCHost asked to exit")
			return
		case event := <-discvEvent:
			blog.Info("refresh DCHost get discover event")
			if event.Err != nil {
				blog.Error("DCHost discover err:%s", event.Err.Error())
				continue
			}
			blog.Infof("get DCHost node num(%d)", len(event.Server))
			var DCHost string
			var DCHosts []string
			for i, server := range event.Server {
				blog.Infof("get DCHost: server[%d]: %s", i, server)
				var serverInfo commtype.BcsStorageInfo
				if err = json.Unmarshal([]byte(server), &serverInfo); err != nil {
					blog.Errorf("fail to unmarshal DCHost(%s), err:%s", string(server), err.Error())
					continue
				}
				if !cfg.IsExternal {
					DCHost = serverInfo.ServerInfo.Scheme + "://" + serverInfo.ServerInfo.IP + ":" + strconv.Itoa(int(serverInfo.ServerInfo.Port))
				} else {
					DCHost = serverInfo.ServerInfo.Scheme + "://" + serverInfo.ServerInfo.ExternalIp + ":" + strconv.Itoa(int(serverInfo.ServerInfo.ExternalPort))
				}

				blog.Infof("get DCHost(%s)", DCHost)
				DCHosts = append(DCHosts, DCHost)
			}

			storage.SetDCAddress(DCHosts)
			if len(DCHosts) > 0 {
				//servermetric.SetDCStatus(true)
				storageState.Set(stateOK)
			} else {
				//servermetric.SetDCStatus(false)
				storageState.Set(stateErr)
			}
		} // end select
	} // end for
}

func registerZkEndpoints(rdCxt context.Context, cfg *types.CmdConfig) (bool, error) {
	clusterinfo := strings.Split(cfg.ClusterInfo, "/")
	regDiscv := rd.NewRegDiscoverEx(clusterinfo[0], time.Second*10)
	if regDiscv == nil {
		return false, fmt.Errorf("registerZkEndpoints(%s) return nil", clusterinfo[0])
	}
	blog.Info("registerZkEndpoints(%s) succ", clusterinfo[0])

	err := regDiscv.Start()
	if err != nil {
		blog.Error("registerZkEndpoints regDisv start error(%s)", err.Error())
		return false, err
	}
	blog.Info("registerZkEndpoints RegDiscover start succ")

	host, err := os.Hostname()
	if err != nil {
		blog.Error("registerZkEndpoints mesoswatcher get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	var regInfo commtype.MesosDataWatchServInfo
	regInfo.ServerInfo.Cluster = cfg.ClusterID
	regInfo.ServerInfo.IP = cfg.Address
	regInfo.ServerInfo.Port = 0
	regInfo.ServerInfo.MetricPort = cfg.MetricPort
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Scheme = cfg.ServerSchem
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	data, err := json.Marshal(regInfo)
	key := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDATAWATCH + "/" + cfg.Address
	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDATAWATCH

	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("registerZkEndpoints RegisterService(%s) error(%s)", key, err.Error())
		regDiscv.Stop()
		return true, err
	}
	blog.Info("registerZkEndpoints RegisterService(%s:%s) succ", key, data)

	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("registerZkEndpoints DiscoverService(%s) error(%s)", discvPath, err.Error())
		regDiscv.Stop()
		return true, err
	}
	blog.Info("registerZkEndpoints DiscoverService(%s) succ", discvPath)

	for {
		select {
		case <-rdCxt.Done():
			blog.V(3).Infof("registerZkEndpoints asked to exit")
			regDiscv.Stop()
			return false, nil
		case event := <-discvEvent:
			blog.Info("registerZkEndpoints get discover event")
			if event.Err != nil {
				blog.Error("registerZkEndpoints get discover event err:%s", event.Err.Error())
				regDiscv.Stop()
				return true, event.Err
			}

			registerd := false
			for i, server := range event.Server {
				blog.Info("registerZkEndpoints get discover event: server[%d]: %s %s", i, event.Key, server)
				if server == string(data) {
					blog.Infof("registerZkEndpoints get discover event, and myself is registered")
					registerd = true
					break
				}
			}

			if !registerd {
				blog.Infof("registerZkEndpoints get discover event, server list len(%d), but cannot find myself", len(event.Server))
				regDiscv.Stop()
				return true, fmt.Errorf("current server is nil")
			}
		} // end select
	} // end for
}
