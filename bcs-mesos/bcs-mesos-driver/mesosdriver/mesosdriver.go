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

package mesosdriver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	rd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	commtype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/backend"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/backend/v4http"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/config"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/filter"

	restful "github.com/emicklei/go-restful"
)

//MesosDriver is data struct of mesos driver
type MesosDriver struct {
	config        *config.MesosDriverConfig
	etcdRegistry  registry.Registry
	httpServ      *httpserver.HttpServer
	v4Scheduler   backend.Scheduler
	CurrScheduler string
	bcsClusterID  string

	inited bool
}

// IsHealthy healthy interface for bcs-health(deprecated)
func (m *MesosDriver) IsHealthy() (bool, string) {

	if !m.inited {
		return false, "in starting"
	}

	if m.bcsClusterID == "" {
		return false, "fail to get clusterid"
	}

	if m.v4Scheduler.GetHost() == "" {
		return false, "cannot connect to any scheduler"
	}

	return true, "run ok"
}

// RunMetric run metric feaature
//deprecated, change to promethus collector
func (m *MesosDriver) RunMetric() {

	conf := metric.Config{
		RunMode:     metric.Master_Master_Mode,
		ModuleName:  commtype.BCS_MODULE_MESOSDRIVER,
		MetricPort:  m.config.MetricPort,
		IP:          m.config.Address,
		ClusterID:   m.bcsClusterID,
		SvrCaFile:   m.config.ServCert.CAFile,
		SvrCertFile: m.config.ServCert.CertFile,
		SvrKeyFile:  m.config.ServCert.KeyFile,
		SvrKeyPwd:   m.config.ServCert.CertPasswd,
	}

	healthFunc := func() metric.HealthMeta {
		ok, msg := m.IsHealthy()
		return metric.HealthMeta{
			CurrentRole: metric.MasterRole,
			IsHealthy:   ok,
			Message:     msg,
		}
	}

	if err := metric.NewMetricController(
		conf,
		healthFunc); nil != err {
		blog.Errorf("run metric fail: %s", err.Error())
	}

	blog.Infof("run metric ok")
}

// NewMesosDriverServer create mesosdriver according config
func NewMesosDriverServer(conf *config.MesosDriverConfig) (*MesosDriver, error) {

	blog.Infof("mesos-driver loading configuration: %+v", conf)

	m := &MesosDriver{}

	//config
	m.config = conf

	//http server
	m.httpServ = httpserver.NewHttpServer(m.config.Port, m.config.Address, "")
	if m.config.ServCert.IsSSL {
		blog.Info("mesos driver server is SSL: CA:%s, Cert:%s, Key:%s",
			m.config.ServCert.CAFile, m.config.ServCert.CertFile, m.config.ServCert.KeyFile)
		m.httpServ.SetSsl(m.config.ServCert.CAFile, m.config.ServCert.CertFile, m.config.ServCert.KeyFile, m.config.ServCert.CertPasswd)
	}

	//v4 scheduler
	m.v4Scheduler = v4http.NewScheduler()
	m.v4Scheduler.InitConfig(m.config)

	m.bcsClusterID = conf.Cluster
	m.inited = true

	return m, nil
}

// Stop driver stop
func (m *MesosDriver) Stop() error {
	return nil
}

// Start start mesosdriver
func (m *MesosDriver) Start() error {

	blog.Info("mesos driver %s run for cluster %s", m.config.Address, m.bcsClusterID)

	m.RunMetric()

	go m.DiscvScheduler()

	if m.config.RegisterWithWebsocket {
		err := m.buildWebsocketToAPI()
		if err != nil {
			blog.Fatalf("err when register with websocket: %s", err.Error())
			os.Exit(1)
		}
	} else {
		go m.RegDiscover()
		if err := m.etcdRegistryFeature(); err != nil {
			os.Exit(1)
		}
		defer m.etcdRegistryClean()
	}

	go m.registerMesosZkEndpoints()

	chErr := make(chan error, 1)

	generalFilter := filter.NewFilter()
	//admission webhook filter
	if m.config.AdmissionWebhook {
		admissionFilter, err := filter.NewAdmissionWebhookFilter(m.v4Scheduler, m.config.KubeConfig)
		if err != nil {
			blog.Errorf(err.Error())
			os.Exit(1)
		}
		generalFilter.AppendFilter(admissionFilter)
		blog.Infof("mesosdriver add admission webhook filter")
	}
	//check http head valid filter
	generalFilter.AppendFilter(filter.NewHeaderValidFilter(m.config))
	blog.Infof("mesosdriver add header valid filter")

	// register filter middleware
	filters := []restful.FilterFunction{}
	filters = append(filters, generalFilter.Filter)

	//register actions
	blog.Info("mesos driver begin register v4 api")
	m.httpServ.RegisterWebServer("/mesosdriver/v4", filters, m.v4Scheduler.Actions())

	go func() {
		err := m.httpServ.ListenAndServe()
		blog.Warn("http listen and service end! err:%s", err.Error())
		chErr <- err
	}()

	select {
	case err := <-chErr:
		blog.Error("exit!, err: %s", err.Error())
		return err
	}
}

// Filter filte request not belong to local cluster
func (m *MesosDriver) Filter(req *restful.Request, resp *restful.Response, filterChain *restful.FilterChain) {
	clusterID := req.Request.Header.Get("BCS-ClusterID")
	if clusterID != m.bcsClusterID {
		msg := fmt.Sprintf("ClusterId %s is invalid", clusterID)
		blog.Error(msg)
		resp.WriteHeaderAndEntity(http.StatusBadRequest, commhttp.APIRespone{Result: false, Code: 1, Message: msg, Data: nil})
		return
	}

	filterChain.ProcessFilter(req, resp)

}

func (m *MesosDriver) etcdRegistryFeature() error {
	if m.config.Etcd.Feature {
		blog.Infof("etcd registry information: %+v", m.config.Etcd)
		tlsCfg, err := m.config.Etcd.GetTLSConfig()
		if err != nil {
			blog.Errorf("turn on etcd registry feature but configuration not correct, %s", err.Error())
			return err
		}
		clusterID := strings.Split(m.config.Cluster, "-")
		if len(clusterID) != 3 {
			blog.Errorf("clusterID in configuration is not formation expected! expect: BCS-MESOS-XXXXX, actual: %s", m.config.Cluster)
			return fmt.Errorf("error clusterID formation")
		}
		// init go-micro registry
		eoption := &registry.Options{
			Name:         clusterID[2] + "." + "mesosdriver.bkbcs.tencent.com",
			Version:      version.BcsVersion,
			RegistryAddr: strings.Split(m.config.Etcd.Address, ","),
			RegAddr:      fmt.Sprintf("%s:%d", m.config.Address, m.config.Port),
			Config:       tlsCfg,
		}
		m.etcdRegistry = registry.NewEtcdRegistry(eoption)
		if err := m.etcdRegistry.Register(); err != nil {
			blog.Errorf("etcd registry feature turn on but register failed, %s", err.Error())
			return err
		}
	}
	return nil
}

func (m *MesosDriver) etcdRegistryClean() {
	if m.config.Etcd.Feature {
		m.etcdRegistry.Deregister()
	}
}

// RegDiscover register local service information
func (m *MesosDriver) RegDiscover() {

	blog.Info("driver to do register ...")

	if m.config.Cluster == "" {
		blog.Error("RegDiscover clusterid is nil, redo after 3 second ...")
		time.Sleep(3 * time.Second)
		go m.RegDiscover()
		return
	}

	// register service
	regDiscv := rd.NewRegDiscoverEx(m.config.RegDiscvSvr, time.Second*10)
	if regDiscv == nil {
		blog.Error("NewRegDiscover(%s) return nil, redo after 3 second ...", m.config.RegDiscvSvr)
		time.Sleep(3 * time.Second)
		go m.RegDiscover()
		return
	}
	blog.Info("NewRegDiscover(%s) succ", m.config.RegDiscvSvr)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("regDiscv start error(%s), redo after 3 second ...", err.Error())
		time.Sleep(3 * time.Second)
		go m.RegDiscover()
		return
	}
	blog.Info("RegDiscover start succ")

	defer regDiscv.Stop()

	host, err := os.Hostname()
	if err != nil {
		blog.Error("mesos driver get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	var regInfo commtype.MesosDriverServInfo
	regInfo.ServerInfo.Cluster = m.config.Cluster
	regInfo.ServerInfo.IP = m.config.Address
	regInfo.ServerInfo.Port = m.config.Port
	regInfo.ServerInfo.ExternalIp = m.config.ExternalIp
	regInfo.ServerInfo.ExternalPort = m.config.ExternalPort
	regInfo.ServerInfo.MetricPort = m.config.MetricPort
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Scheme = "http"
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	if m.config.ServCert.IsSSL {
		regInfo.ServerInfo.Scheme = "https"
	}

	key := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDRIVER + "/" + regInfo.Cluster + "/" + m.config.Address
	data, err := json.Marshal(regInfo)
	if err != nil {
		blog.Error("json Marshal error(%s)", err.Error())
		return
	}
	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("RegisterService(%s) error(%s), redo after 3 second ...", key, err.Error())
		time.Sleep(3 * time.Second)
		go m.RegDiscover()
		return
	}
	blog.Info("RegisterService(%s:%s) succ", key, data)

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDRIVER + "/" + regInfo.Cluster
	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("DiscoverService(%s) error(%s), redo after 3 second ...", discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go m.RegDiscover()
		return
	}
	blog.Info("DiscoverService(%s) succ", discvPath)

	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("tick: driver(cluster:%s %s:%d) running, scheduler(%s), current goroutine num (%d)",
				m.config.Cluster, m.config.Address, m.config.Port, m.CurrScheduler, runtime.NumGoroutine())

		case event := <-discvEvent:
			blog.Info("get discover event")
			if event.Err != nil {
				blog.Error("get discover event err:%s,  redo after 3 second ...", event.Err.Error())
				time.Sleep(3 * time.Second)
				go m.RegDiscover()
				return
			}

			isRegstered := false
			for i, server := range event.Server {
				blog.Info("discovered : server[%d]: %s %s", i, event.Key, server)
				if server == string(data) {
					blog.Info("discovered : server[%d] is myself", i)
					isRegstered = true
				}
			}

			if isRegstered == false {
				blog.Warn("drive is not regestered in zk, do register after 3 second ...")
				time.Sleep(3 * time.Second)
				go m.RegDiscover()
				return
			}
		} // end select
	} // end for
}

// DiscvScheduler discovery scheduler master
func (m *MesosDriver) DiscvScheduler() {
	blog.Infof("begin to discover scheduler from (%s), curr goroutine num(%d)", m.config.SchedDiscvSvr, runtime.NumGoroutine())
	MesosDiscv := m.config.SchedDiscvSvr
	regDiscv := rd.NewRegDiscover(MesosDiscv)
	if regDiscv == nil {
		blog.Errorf("new scheduler discover(%s) return nil", MesosDiscv)
		time.Sleep(3 * time.Second)
		go m.DiscvScheduler()
		return
	}
	blog.Infof("new scheduler discover(%s) succ, current goroutine num(%d)", MesosDiscv, runtime.NumGoroutine())

	err := regDiscv.Start()
	if err != nil {
		blog.Errorf("scheduler discover start error(%s)", err.Error())
		time.Sleep(3 * time.Second)
		go m.DiscvScheduler()
		return
	}
	blog.Infof("scheduler discover start succ, current goroutine num(%d)", runtime.NumGoroutine())
	//defer regDiscv.Stop()

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_SCHEDULER
	discvMesosEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Errorf("watch scheduler under (%s: %s) error(%s)", MesosDiscv, discvPath, err.Error())
		regDiscv.Stop()
		time.Sleep(3 * time.Second)
		go m.DiscvScheduler()
		return
	}
	blog.Infof("watch scheduler under (%s: %s), current goroutine num(%d)", MesosDiscv, discvPath, runtime.NumGoroutine())

	tick := time.NewTicker(180 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Infof("scheduler discover(%s:%s), curr scheduler:%s", MesosDiscv, discvPath, m.CurrScheduler)
		case event := <-discvMesosEvent:
			blog.Infof("discover event for scheduler")
			if event.Err != nil {
				blog.Errorf("get scheduler discover event err:%s", event.Err.Error())
				regDiscv.Stop()
				time.Sleep(3 * time.Second)
				go m.DiscvScheduler()
				return
			}
			currSched := ""
			blog.Infof("get scheduler node num(%d)", len(event.Server))
			for i, server := range event.Server {
				blog.Infof("get scheduler: server[%d]: %s %s", i, event.Key, server)
				var serverInfo commtype.SchedulerServInfo
				if err = json.Unmarshal([]byte(server), &serverInfo); err != nil {
					blog.Errorf("fail to unmarshal scheduler(%s), err:%s", string(server), err.Error())
				}
				if i == 0 {
					currSched = serverInfo.ServerInfo.Scheme + "://" + serverInfo.ServerInfo.IP + ":" + strconv.Itoa(int(serverInfo.ServerInfo.Port))
				}
			}
			if currSched != m.CurrScheduler {
				blog.Infof("scheduler changed(%s-->%s)", m.CurrScheduler, currSched)
				m.CurrScheduler = currSched
				m.v4Scheduler.SetHost([]string{currSched})
			}
		} // select
	} // for
}

func (m *MesosDriver) registerMesosZkEndpoints() {
	blog.Info("registerMesosZkEndpoints driver to do register ...")
	// register service
	regDiscv := rd.NewRegDiscoverEx(m.config.SchedDiscvSvr, time.Second*10)
	if regDiscv == nil {
		blog.Error("registerMesosZkEndpoints(%s) return nil, redo after 3 second ...", m.config.SchedDiscvSvr)
		time.Sleep(3 * time.Second)
		go m.registerMesosZkEndpoints()
		return
	}
	blog.Info("registerMesosZkEndpoints(%s) succ", m.config.SchedDiscvSvr)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("registerMesosZkEndpoints regDiscv start error(%s), redo after 3 second ...", err.Error())
		time.Sleep(3 * time.Second)
		go m.registerMesosZkEndpoints()
		return
	}
	blog.Info("registerMesosZkEndpoints start succ")
	defer regDiscv.Stop()

	host, err := os.Hostname()
	if err != nil {
		blog.Error("registerMesosZkEndpoints mesos driver get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	var regInfo commtype.MesosDriverServInfo
	regInfo.ServerInfo.Cluster = m.config.Cluster
	regInfo.ServerInfo.IP = m.config.Address
	regInfo.ServerInfo.Port = m.config.Port
	regInfo.ServerInfo.MetricPort = m.config.MetricPort
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Scheme = "http"
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	if m.config.ServCert.IsSSL {
		regInfo.ServerInfo.Scheme = "https"
	}

	key := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDRIVER + "/" + m.config.Address
	data, err := json.Marshal(regInfo)
	if err != nil {
		blog.Error("registerMesosZkEndpoints json Marshal error(%s)", err.Error())
		return
	}
	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("registerMesosZkEndpoints(%s) error(%s), redo after 3 second ...", key, err.Error())
		time.Sleep(3 * time.Second)
		go m.registerMesosZkEndpoints()
		return
	}
	blog.Info("registerMesosZkEndpoints(%s:%s) succ", key, data)

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_MESOSDRIVER
	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("registerMesosZkEndpoints(%s) error(%s), redo after 3 second ...", discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go m.registerMesosZkEndpoints()
		return
	}
	blog.Info("registerMesosZkEndpoints(%s) succ", discvPath)

	for {
		select {
		case event := <-discvEvent:
			blog.Info("registerMesosZkEndpoints get discover event")
			if event.Err != nil {
				blog.Error("registerMesosZkEndpoints get discover event err:%s,  redo after 3 second ...", event.Err.Error())
				time.Sleep(3 * time.Second)
				go m.registerMesosZkEndpoints()
				return
			}

			isRegstered := false
			for i, server := range event.Server {
				blog.Info("registerMesosZkEndpoints discovered : server[%d]: %s %s", i, event.Key, server)
				if server == string(data) {
					blog.Info("registerMesosZkEndpoints discovered : server[%d] is myself", i)
					isRegstered = true
				}
			}

			if isRegstered == false {
				blog.Warn("registerMesosZkEndpoints drive is not regestered in zk, do register after 3 second ...")
				time.Sleep(3 * time.Second)
				go m.registerMesosZkEndpoints()
				return
			}
		} // end select
	} // end for
}
