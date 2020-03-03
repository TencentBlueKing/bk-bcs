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

package regdiscv

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-services/bcs-api/config"

	"golang.org/x/net/context"
)

var rd *RDiscover

func GetRDiscover() (*RDiscover, error) {
	if rd == nil {
		return nil, fmt.Errorf("RDiscover is't init")
	}

	return rd, nil
}

//RDiscover route register and discover
type RDiscover struct {
	sync.RWMutex

	rd     *RegisterDiscover.RegDiscover
	conf   *config.ApiServConfig
	cliTls *tls.Config

	rootCxt context.Context
	cancel  context.CancelFunc

	servers map[string][]interface{} //key=BCS_MODULE_K8SAPISERVER...; value=[]*types.BcsK8sApiserverInfo...
	events  chan *RegisterDiscover.DiscoverEvent

	k8sapiClustersNodes   map[string]struct{}
	mesosapiClustersNodes map[string]struct{}
}

//NewRDiscover create a object of RDiscover
func RunRDiscover(zkserv string, conf *config.ApiServConfig) error {
	rd = &RDiscover{
		rd:   RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
		conf: conf,
		servers: map[string][]interface{}{
			types.BCS_MODULE_STORAGE:          make([]interface{}, 0),
			types.BCS_MODULE_NETSERVICE:       make([]interface{}, 0),
			types.BCS_MODULE_MESOSAPISERVER:   make([]interface{}, 0),
			types.BCS_MODULE_K8SAPISERVER:     make([]interface{}, 0),
			types.BCS_MODULE_METRICSERVICE:    make([]interface{}, 0),
			types.BCS_MODULE_CLUSTERKEEPER:    make([]interface{}, 0),
			types.BCS_MODULE_NETWORKDETECTION: make([]interface{}, 0),
		},
		events:                make(chan *RegisterDiscover.DiscoverEvent, 1024),
		k8sapiClustersNodes:   make(map[string]struct{}),
		mesosapiClustersNodes: make(map[string]struct{}),
	}

	if conf.ClientCert.IsSSL {
		cliTls, err := ssl.ClientTslConfVerity(conf.ClientCert.CAFile, conf.ClientCert.CertFile, conf.ClientCert.KeyFile, conf.ClientCert.CertPasswd)
		if err != nil {
			blog.Errorf("set client tls config error %s", err.Error())
		} else {
			rd.cliTls = cliTls
			blog.Infof("set client tls config success")
		}
	}

	return rd.start()
}

//input: types.BCS_MODULE_STORAGE...
func (r *RDiscover) GetModuleServers(moduleName string) (interface{}, error) {
	r.RLock()
	defer r.RUnlock()

	servs, ok := r.servers[moduleName]
	if !ok {
		return nil, fmt.Errorf("Module %s is invalid", moduleName)
	}

	if len(servs) == 0 {
		return nil, fmt.Errorf("Module %s have no servers endpoints", moduleName)
	}

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	serv := servs[rand.Intn(len(servs))]
	return serv, nil
}

func (r *RDiscover) GetClientTls() (*tls.Config, error) {
	if r.cliTls == nil {
		return nil, fmt.Errorf("client tls is empty")
	}

	return r.cliTls, nil
}

//Start the rdiscover
func (r *RDiscover) start() error {
	//create root context
	r.rootCxt, r.cancel = context.WithCancel(context.Background())

	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}

	//register apiserver
	err := r.registerAPI()
	if err != nil {
		blog.Errorf("register apiserver error %s", err.Error())
		return err
	}

	//discover other bcs service
	for k := range r.servers {
		go r.discoverServices(k)
	}

	return nil
}

func (r *RDiscover) discoverServices(k string) {
	key := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, k)
	blog.Infof("start discover service key %s", key)
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to register discover for api. err:%s", err.Error())
		r.cancel()
		os.Exit(1)
	}

	for {
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.Errorf("discover zk key %s error %s", key, eve.Err.Error())
				time.Sleep(time.Second)
				go r.discoverServices(k)
				return
			}

			switch path.Base(eve.Key) {
			//storage
			case types.BCS_MODULE_STORAGE:
				r.discoverStorageServ(eve.Server)
			// mesos apiserver
			case types.BCS_MODULE_MESOSAPISERVER:
				r.discoverMesosApiserver(eve.Nodes)
			// netservice
			case types.BCS_MODULE_NETSERVICE:
				r.discoverNetserviceServ(eve.Server)
			// metric service
			case types.BCS_MODULE_METRICSERVICE:
				r.discoverMetricServer(eve.Server)
			//k8s apiserver
			case types.BCS_MODULE_K8SAPISERVER:
				r.discoverK8sApiserver(eve.Nodes)
			//cluster keeper
			case types.BCS_MODULE_CLUSTERKEEPER:
				r.discoverClusterkeeper(eve.Server)
			// network detection
			case types.BCS_MODULE_NETWORKDETECTION:
				r.discoverDetectionServ(eve.Server)
			}

		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

//Stop the rdiscover
func (r *RDiscover) stop() error {
	r.cancel()

	r.rd.Stop()

	return nil
}

func (r *RDiscover) registerAPI() error {
	apiServInfo := new(types.APIServInfo)

	apiServInfo.IP = r.conf.LocalIp
	apiServInfo.Port = r.conf.InsecurePort
	apiServInfo.Scheme = "http"
	apiServInfo.MetricPort = r.conf.MetricPort
	if r.conf.ServCert.IsSSL {
		apiServInfo.Scheme = "https"
		apiServInfo.Port = r.conf.Port
	}
	apiServInfo.Version = version.GetVersion()
	apiServInfo.Pid = os.Getpid()

	data, err := json.Marshal(apiServInfo)
	if err != nil {
		blog.Error("fail to marshal apiservInfo to json. err:%s", err.Error())
		return err
	}

	path := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_APISERVER + "/" + r.conf.LocalIp

	blog.Infof("register key %s apiserver %s", path, string(data))
	return r.rd.RegisterAndWatchService(path, data)
}

func (r *RDiscover) discoverStorageServ(servInfos []string) error {
	blog.Info("discover storage(%v)", servInfos)

	storages := make([]interface{}, 0)
	for _, serv := range servInfos {
		storage := new(types.BcsStorageInfo)
		if err := json.Unmarshal([]byte(serv), storage); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		storages = append(storages, storage)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_STORAGE] = storages
	r.Unlock()

	return nil
}

func (r *RDiscover) discoverDetectionServ(servInfos []string) error {
	blog.Info("discover network detection(%v)", servInfos)

	detections := make([]interface{}, 0)
	for _, serv := range servInfos {
		detection := new(types.NetworkDetectionServInfo)
		if err := json.Unmarshal([]byte(serv), detection); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		detections = append(detections, detection)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_NETWORKDETECTION] = detections
	r.Unlock()

	return nil
}

func (r *RDiscover) discoverNetserviceServ(servInfos []string) error {
	blog.Info("discover netservice(%v)", servInfos)

	netservices := make([]interface{}, 0)
	for _, serv := range servInfos {
		netservice := new(types.NetServiceInfo)
		if err := json.Unmarshal([]byte(serv), netservice); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		netservices = append(netservices, netservice)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_NETSERVICE] = netservices
	r.Unlock()

	return nil
}

func (r *RDiscover) discoverClusterkeeper(servInfos []string) error {
	blog.Info("discover clusterkeeper(%v)", servInfos)

	clusters := make([]interface{}, 0)
	for _, serv := range servInfos {
		cluster := new(types.ClusterKeeperServInfo)
		if err := json.Unmarshal([]byte(serv), cluster); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		clusters = append(clusters, cluster)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_CLUSTERKEEPER] = clusters
	r.Unlock()

	return nil
}

func (r *RDiscover) discoverMesosApiserver(nodes []string) error {

	for _, node := range nodes {
		_, ok := r.mesosapiClustersNodes[node]
		if ok {
			continue
		}

		r.mesosapiClustersNodes[node] = struct{}{}
		blog.Infof("start discover cluster %s mesosapi", node)
		key := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_MESOSAPISERVER, node)

		go r.discoverClusterMesosApiserver(key)
	}

	return nil
}

func (r *RDiscover) discoverClusterMesosApiserver(key string) {
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to discover service %s err:%s", key, err.Error())
		time.Sleep(time.Second)
		go r.discoverClusterMesosApiserver(key)
		return
	}

	for {
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.Errorf("discover zk key %s error %s", key, err.Error())
				time.Sleep(time.Second)
				go r.discoverClusterMesosApiserver(key)
				return
			}

			mesosApis := make([]interface{}, 0)
			for _, serv := range eve.Server {
				blog.Infof("discover key %s mesos apiserver %s", key, serv)

				api := new(types.BcsMesosApiserverInfo)
				if err := json.Unmarshal([]byte(serv), api); err != nil {
					blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
					continue
				}

				mesosApis = append(mesosApis, api)
			}

			r.Lock()
			r.servers[fmt.Sprintf("%s/%s", types.BCS_MODULE_MESOSAPISERVER, path.Base(key))] = mesosApis
			r.Unlock()
		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

func (r *RDiscover) discoverK8sApiserver(nodes []string) error {

	for _, node := range nodes {
		_, ok := r.k8sapiClustersNodes[node]
		if ok {
			continue
		}

		r.k8sapiClustersNodes[node] = struct{}{}
		blog.Infof("start discover cluster %s k8sapi", node)
		key := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_K8SAPISERVER, node)

		go r.discoverClusterK8sApiserver(key)
	}

	return nil
}

func (r *RDiscover) discoverClusterK8sApiserver(key string) {
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to discover service %s err:%s", key, err.Error())
		time.Sleep(time.Second)
		go r.discoverClusterK8sApiserver(key)
		return
	}

	for {
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.Errorf("discover zk key %s error %s", key, err.Error())
				time.Sleep(time.Second)
				go r.discoverClusterK8sApiserver(key)
				return
			}

			k8sApis := make([]interface{}, 0)
			for _, serv := range eve.Server {
				blog.Infof("discover cluster %s k8s apiserver %s", key, serv)

				api := new(types.BcsK8sApiserverInfo)
				if err := json.Unmarshal([]byte(serv), api); err != nil {
					blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
					continue
				}

				k8sApis = append(k8sApis, api)
			}

			r.Lock()
			r.servers[fmt.Sprintf("%s/%s", types.BCS_MODULE_K8SAPISERVER, path.Base(key))] = k8sApis
			r.Unlock()
		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

func (r *RDiscover) discoverMetricServer(servInfos []string) error {
	blog.Info("discover metricservice(%v)", servInfos)

	clusters := make([]interface{}, 0)
	for _, serv := range servInfos {
		cluster := new(types.MetricServiceInfo)
		if err := json.Unmarshal([]byte(serv), cluster); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		clusters = append(clusters, cluster)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_METRICSERVICE] = clusters
	r.Unlock()

	return nil
}
