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

package bcscheck

import (
	rd "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/metric"
	commtype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/config"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/manager"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/task"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"os"
	"runtime"
	"time"
)

type HealthCheckServer struct {
	conf config.HealthCheckConfig

	manager manager.Manager

	task task.TaskManager

	isMaster bool
	Role     metric.RoleType

	cxt    context.Context
	cancel context.CancelFunc
}

func NewHealthCheckServer(conf config.HealthCheckConfig) *HealthCheckServer {
	cxt, cancel := context.WithCancel(context.Background())

	by, _ := json.Marshal(conf)
	blog.V(3).Infof("NewHealthCheckServer conf %s", string(by))

	s := &HealthCheckServer{
		conf:   conf,
		cxt:    cxt,
		cancel: cancel,
	}

	return s
}

func (s *HealthCheckServer) initManager() {
	s.manager = manager.NewManager(s.cxt, s.conf)
}

func (s *HealthCheckServer) initTaskManeger() {
	s.task = task.NewTaskManager(s.cxt, s.conf, s.manager)
}

func (s *HealthCheckServer) initClusterId() {

	if s.conf.Cluster == "" {
		blog.Error("healthcheck cluster unknown")
		os.Exit(1)
	}

	blog.Infof("healthcheck run for cluster %s", s.conf.Cluster)

}

func (s *HealthCheckServer) Run() {
	s.initClusterId()
	go s.regDiscover()
	go s.regBcsDiscover()
	go s.runMetric()
}

func (s *HealthCheckServer) Stop() {
	blog.Info("HealthCheckServer stopped")
	s.cancel()
}

func (s *HealthCheckServer) masterStart() {
	blog.Infof("health check role change: slave --> master")

	s.isMaster = true

	blog.Info("Manager run ...")
	s.initManager()
	go s.manager.Run()

	blog.Info("TaskManager run ...")
	s.initTaskManeger()
	go s.task.Run()
}

func (s *HealthCheckServer) slaveStop() {
	blog.Infof("health check role change: master --> slave")

	s.isMaster = false
	s.Stop()

	cxt, cancel := context.WithCancel(context.Background())
	s.cxt = cxt
	s.cancel = cancel
}

func (s *HealthCheckServer) regDiscover() {

	blog.Info("HealthCheckServer to do register ...")

	// register service
	regDiscv := rd.NewRegDiscoverEx(s.conf.RegDiscvSvr, time.Second*10)
	if regDiscv == nil {
		blog.Error("NewRegDiscover(%s) return nil, redo after 3 second ...", s.conf.RegDiscvSvr)
		time.Sleep(3 * time.Second)

		go s.regDiscover()
		return
	}
	blog.Info("NewRegDiscover(%s) succ", s.conf.RegDiscvSvr)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("regDiscv start error(%s), redo after 3 second ...", err.Error())
		time.Sleep(3 * time.Second)

		go s.regDiscover()
		return
	}
	blog.Info("RegDiscover start succ")

	defer func() {
		err = regDiscv.Stop()
		if err != nil {
			blog.Errorf("regDiscv stop error %s", err.Error())
		}
	}()

	host, err := os.Hostname()
	if err != nil {
		blog.Error("health check get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	var regInfo commtype.BcsCheckInfo
	regInfo.ServerInfo.Cluster = s.conf.Cluster
	regInfo.ServerInfo.IP = s.conf.Address
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	regInfo.ServerInfo.MetricPort = s.conf.MetricPort

	key := fmt.Sprintf("%s/%s/%s.%d", commtype.BCS_SERV_BASEPATH, commtype.BCS_MODULE_Check,
		s.conf.Address, os.Getpid())

	//key = commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_Check + "/" + regInfo.Cluster + "/" + s.conf.Address

	data, err := json.Marshal(regInfo)
	if err != nil {
		blog.Error("json Marshal error(%s)", err.Error())
		return
	}

	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("RegisterService(%s) error(%s), redo after 3 second ...", key, err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscover()
		return
	}

	blog.Info("RegisterService(%s:%s) succ", key, data)

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_Check

	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("DiscoverService(%s) error(%s), redo after 3 second ...", discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscover()
		return
	}

	blog.Info("DiscoverService(%s) succ", discvPath)

	tick := time.NewTicker(180 * time.Second)
	for {
		select {
		case <-tick.C:
			blog.Info("tick: health check(cluster:%s %s) running, current goroutine num (%d)",
				s.conf.Cluster, s.conf.Address, runtime.NumGoroutine())

		case event := <-discvEvent:
			blog.Info("get discover event")

			if event.Err != nil {
				blog.Error("get discover event err:%s,  redo after 3 second ...", event.Err.Error())
				time.Sleep(3 * time.Second)
				go s.regDiscover()
				return
			}

			isRegstered := false
			isMaster := false

			for i, server := range event.Server {
				blog.Info("discovered : server[%d]: %s %s", i, event.Key, server)
				if server == string(data) {
					blog.Info("discovered : server[%d] is myself", i)
					isRegstered = true
				}

				if i == 0 && server == string(data) {
					isMaster = true
					blog.Info("discoved : I am master")
				}
			}

			if isRegstered == false {
				blog.Warn("drive is not regestered in zk, do register after 3 second ...")
				time.Sleep(3 * time.Second)
				s.Role = metric.UnknownRole
				go s.regDiscover()
				return
			}

			//from slave change to master
			if isMaster && !s.isMaster {
				s.Role = metric.MasterRole
				s.masterStart()
			}

			if !isMaster && s.isMaster {
				s.Role = metric.SlaveRole
				s.slaveStop()
			}

		} // end select
	} // end for
}

func (s *HealthCheckServer) regBcsDiscover() {

	blog.Info("HealthCheckServer to do bcs register ...")

	clusterId := s.conf.Cluster

	// register service
	regDiscv := rd.NewRegDiscoverEx(s.conf.BcsDiscvSvr, time.Second*10)
	if regDiscv == nil {
		blog.Error("NewRegDiscover(%s) return nil, redo after 3 second ...", s.conf.BcsDiscvSvr)
		time.Sleep(3 * time.Second)

		go s.regDiscover()
		return
	}
	blog.Info("NewRegDiscover(%s) succ", s.conf.BcsDiscvSvr)

	err := regDiscv.Start()
	if err != nil {
		blog.Error("regDiscv start error(%s), redo after 3 second ...", err.Error())
		time.Sleep(3 * time.Second)

		go s.regDiscover()
		return
	}
	blog.Info("RegDiscover start succ")

	defer func() {
		err = regDiscv.Stop()
		if err != nil {
			blog.Errorf("regDiscv stop error %s", err.Error())
		}
	}()

	host, err := os.Hostname()
	if err != nil {
		blog.Error("health check get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	var regInfo commtype.BcsCheckInfo
	regInfo.ServerInfo.Cluster = s.conf.Cluster
	regInfo.ServerInfo.IP = s.conf.Address
	regInfo.ServerInfo.HostName = host
	regInfo.ServerInfo.Pid = os.Getpid()
	regInfo.ServerInfo.Version = version.GetVersion()
	regInfo.ServerInfo.MetricPort = s.conf.MetricPort

	key := fmt.Sprintf("%s/%s/%s/%s.%d", commtype.BCS_SERV_BASEPATH, commtype.BCS_MODULE_Check, clusterId,
		s.conf.Address, os.Getpid())

	//key = commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_Check + "/" + regInfo.Cluster + "/" + s.conf.Address

	data, err := json.Marshal(regInfo)
	if err != nil {
		blog.Error("json Marshal error(%s)", err.Error())
		return
	}

	err = regDiscv.RegisterService(key, []byte(data))
	if err != nil {
		blog.Error("RegisterService(%s) error(%s), redo after 3 second ...", key, err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscover()
		return
	}

	blog.Info("RegisterService(%s:%s) succ", key, data)

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_Check + "/" + clusterId

	discvEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Error("DiscoverService(%s) error(%s), redo after 3 second ...", discvPath, err.Error())
		time.Sleep(3 * time.Second)
		go s.regDiscover()
		return
	}

	blog.Info("DiscoverService(%s) succ", discvPath)

	tick := time.NewTicker(180 * time.Second)
	for {
		select {
		case <-tick.C:
			blog.Info("tick: health check(cluster:%s %s) running, current goroutine num (%d)",
				s.conf.Cluster, s.conf.Address, runtime.NumGoroutine())

		case event := <-discvEvent:
			blog.Info("get discover event")

			if event.Err != nil {
				blog.Error("get discover event err:%s,  redo after 3 second ...", event.Err.Error())
				time.Sleep(3 * time.Second)
				go s.regDiscover()
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

				go s.regDiscover()
				return
			}

		} // end select
	} // end for

}

func (s *HealthCheckServer) runMetric() {

	metricConf := metric.Config{
		RunMode:     metric.Master_Slave_Mode,
		ModuleName:  commtype.BCS_MODULE_Check,
		MetricPort:  s.conf.MetricPort,
		IP:          s.conf.Address,
		SvrCaFile:   s.conf.ServCert.CAFile,
		SvrCertFile: s.conf.ServCert.CertFile,
		SvrKeyFile:  s.conf.ServCert.KeyFile,
		SvrKeyPwd:   s.conf.ServCert.CertPasswd,
		ClusterID:   s.conf.Cluster,
	}

	healthFunc := func() metric.HealthMeta {
		var isHealthy bool
		var msg string
		if s.Role != metric.UnknownRole {
			isHealthy = true
		} else {
			msg = fmt.Sprintf("bcs-check %s register zk %s failed", s.conf.Address, s.conf.RegDiscvSvr)
		}

		return metric.HealthMeta{
			CurrentRole: s.Role,
			IsHealthy:   isHealthy,
			Message:     msg,
		}
	}

	if err := metric.NewMetricController(
		metricConf,
		healthFunc); nil != err {
		blog.Errorf("run metric fail: %s", err.Error())
	}

	blog.Infof("run metric ok")
}
