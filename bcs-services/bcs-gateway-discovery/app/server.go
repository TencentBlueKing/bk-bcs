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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-common/pkg/master"
	discoverys "bk-bcs/bcs-common/pkg/module-discovery"
	"bk-bcs/bcs-services/bcs-gateway-discovery/register"
	"bk-bcs/bcs-services/bcs-gateway-discovery/register/kong"
)

func New() *DiscoveryServer {
	cxt, cfunc := context.WithCancel(context.Background())
	s := &DiscoveryServer{
		exitCancel: cfunc,
		exitCxt:    cxt,
		evtCh:      make(chan *ModuleEvent, 12),
	}
	return s
}

type ModuleEvent struct {
	Module string
	Svc    *register.Service
}

//DiscoveryServer holds all resources for services discovery
type DiscoveryServer struct {
	option *ServerOptions
	//manager for gateway information register
	regMgr register.Register
	//adapter for service structure convertion
	adapter *Adapter
	//bk-bcs modules discovery for backend service list
	discovery discoverys.ModuleDiscovery
	//self node registe & master node discovery
	bcsRegister master.Master
	//exit func
	exitCancel context.CancelFunc
	//exit context
	exitCxt context.Context
	//Event channel for module-discovery callback
	evtCh chan *ModuleEvent
}

//Init init all running resources, including
// 1. configuration validation
// 2. connecting gateway admin api
// 3. init backend service information adapter
func (s *DiscoveryServer) Init(option *ServerOptions) error {
	if option == nil {
		return fmt.Errorf("Lost ServerOptions")
	}
	s.option = option
	if err := option.Valid(); err != nil {
		return err
	}
	//init gateway master discovery
	if err := s.selfRegister(); err != nil {
		return err
	}
	//init gateway manager
	gatewayAddrs := strings.Split(option.AdminAPI, ",")
	tlsConfig, err := option.GetClientTLS()
	if err != nil {
		return err
	}
	s.regMgr, err = kong.New(gatewayAddrs, tlsConfig)
	if err != nil {
		blog.Errorf("gateway init kong admin api register implementation failed, %s", err.Error())
		return err
	}
	//init service data adapter
	s.adapter = NewAdapter(option.Modules)
	//init module disovery
	allModules := append(defaultModules, option.Modules...)
	s.discovery, err = discoverys.NewDiscoveryV2(option.ZkConfig.BCSZk, allModules)
	if err != nil {
		blog.Errorf("gateway init services discovery failed, %s", err.Error())
		return err
	}
	s.discovery.RegisterEventFunc(s.moduleEventNotifycation)
	return nil
}

//Run running all necessary convertion logic, block
func (s *DiscoveryServer) Run() error {
	//check master status first
	if err := s.dataSynchronization(); err != nil {
		blog.Errorf("gateway-discovery first data synchronization failed, %s", err.Error())
		return err
	}
	tick := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-s.exitCxt.Done():
			blog.Infof("gateway-discovery asked to exit")
			return nil
		case <-tick.C:
			blog.Infof("gateway-discovery time to verify data synchronization....")
			s.dataSynchronization()
		case evt := <-s.evtCh:
			if evt == nil {
				blog.Errorf("module-discovery event channel closed, gateway-discovery error exit")
				return fmt.Errorf("module-discover channel closed")
			}
			blog.Infof("gateway-discovery got module %s changed event", evt.Module)
			//ready to update specified module proxy rules
			s.handleModuleChange(evt)
		}
	}
}

//Stop all backgroup routines
func (s *DiscoveryServer) Stop() {
	s.bcsRegister.Clean()
	s.bcsRegister.Finit()
	s.exitCancel()
}

//selfRegister
func (s *DiscoveryServer) selfRegister() error {
	zkAddrs := strings.Split(s.option.BCSZk, ",")
	selfPath := filepath.Join(types.BCS_SERV_BASEPATH, types.BCS_MODULE_GATEWAYDISCOVERY)
	//self node information
	hostname, _ := os.Hostname()
	self := &types.ServerInfo{
		IP:         s.option.ServiceConfig.Address,
		Port:       s.option.ServiceConfig.Port,
		Pid:        os.Getpid(),
		HostName:   hostname,
		Scheme:     "https",
		Version:    version.BcsVersion,
		MetricPort: s.option.MetricConfig.MetricPort,
	}
	var err error
	s.bcsRegister, err = master.NewZookeeperMaster(zkAddrs, selfPath, self)
	if err != nil {
		blog.Errorf("gateway-discovery init zookeeper master machinery failed, %s", err.Error())
		return err
	}
	//ready to start
	if err = s.bcsRegister.Init(); err != nil {
		blog.Errorf("gateway-discovery start master machinery failed, %s", err.Error())
		return err
	}
	if err = s.bcsRegister.Register(); err != nil {
		blog.Errorf("gateway-discvovery register local service instance failed, %s", err.Error())
		return err
	}
	//time for registe & master ready
	time.Sleep(time.Second)
	return nil
}

//service event notification
func (s *DiscoveryServer) moduleEventNotifycation(module string) {
	if !s.bcsRegister.IsMaster() {
		blog.Infof("gateway-discovery instance is not master, skip module %s event notification", module)
		return
	}
	//get event notification
	event := &ModuleEvent{
		Module: module,
	}
	s.evtCh <- event
}

//dataSynchronization sync all data from bk bcs service discovery to gateway
func (s *DiscoveryServer) dataSynchronization() error {
	if !s.bcsRegister.IsMaster() {
		blog.Infof("gateway-discovery instance is not master, skip data synchronization")
		return nil
	}
	//first get all gateway route information
	regisetedService, err := s.regMgr.ListServices()
	if err != nil {
		blog.Errorf("gateway-discovery get all registed Service from Register failed, %s. wait for next tick", err.Error())
		return err
	}
	regisetedMap := make(map[string]*register.Service)
	if len(regisetedService) == 0 {
		blog.Warnf("gateway-discovery finds no registed service from Register, maybe this is first synchronization.")
	} else {
		for _, srv := range regisetedService {
			blog.V(3).Infof("gateway-discovery check Service %s is under regiseted", srv.Name)
			regisetedMap[srv.Name] = srv
		}
	}
	//get all register module information
	allModules := append(defaultModules, s.option.Modules...)
	var localCaches []*register.Service
	for _, m := range allModules {
		svcs, err := s.formatBCSServerInfo(m)
		if err != nil {
			blog.Errorf("gateway-discovery even get Module %s from cache failed in synchronization, continue")
			continue
		}
		if len(svcs) == 0 {
			blog.Warnf("gateway-discovery get no %s ServerInfo from module-discovery synchronization, try next modules", m)
			continue
		}
		//data structure conversion
		rSvcs, err := s.adapter.GetService(m, svcs)
		if err != nil {
			blog.Errorf("gateway-discovery converts module %s ServerInfo to api-gateway info failed in synchronization, %s", m, err.Error())
			continue
		}
		localCaches = append(localCaches, rSvcs)
	}
	//differ datas
	if len(localCaches) == 0 {
		blog.Warnf("gateway-discovery finds no bk-bcs service in module-discovery, please check bk-bcs discovery machinery")
		return nil
	}
	//udpate datas in gateway
	for _, local := range localCaches {
		svc, ok := regisetedMap[local.Name]
		if ok {
			//service reigsted, we affirm that proxy rule is correct
			// so just update backend targets info
			if err := s.regMgr.ReplaceTargetByService(svc, local.Backends); err != nil {
				blog.Errorf("gateway-discovery update Service %s backend failed in synchronization, %s. backend %v", svc.Name, local.Backends)
				continue
			}
			blog.V(3).Infof("Update serivce %s backend %v successfully", svc.Name, local.Backends)
		} else {
			blog.Infof("Service %s is Not affective in api-gateway when synchronization, try creation", local.Name)
			//create service in api-gateway
			if err := s.regMgr.CreateService(local); err != nil {
				blog.Errorf("discovery create Service %s failed in synchronization, %s. details: %+v", local.Name, err.Error(), local)
				continue
			}
			blog.Infof("discovery create %s Service successfully", local.Name)
			blog.V(3).Infof("Service Creation details: %+v", local)
		}
	}
	blog.Infof("gateway-discovery data synchroniztion finish")
	//todo(DevelperJim): try to fix this feature if we don't allow edit api-gateway configuration manually
	//we don't clean additional datas in api-gateway,
	// because we allow registe service information in api-gateway manually
	return nil
}

func (s *DiscoveryServer) handleModuleChange(event *ModuleEvent) error {
	//get specified module info and construct data for refresh
	svcs, err := s.formatBCSServerInfo(event.Module)
	if err != nil {
		return err
	}
	if len(svcs) == 0 {
		blog.Errorf("gateway-discovery get no %s ServerInfo from module-discovery when in ModuleChanged Event", event.Module)
		return fmt.Errorf("Lost %s ServerInfo after Module Event", event.Module)
	}
	//data structure conversion
	event.Svc, err = s.adapter.GetService(event.Module, svcs)
	if err != nil {
		blog.Errorf("gateway-discovery converts module %s ServerInfo to api-gateway info failed, %s", event.Module, err.Error())
		return err
	}
	//update service route
	exist, err := s.regMgr.GetService(event.Svc.Name)
	if err != nil {
		blog.Errorf("gateway-discovery get register Service %s failed in ModuleChanged Event, %s. it can only recover in dataSynchronization", event.Module, err.Error())
		return err
	}
	if exist == nil {
		blog.Infof("gateway-discovery find no %s module in api-gateway in ModuleChanged, try to create...", event.Module)
		if err := s.regMgr.CreateService(event.Svc); err != nil {
			blog.Errorf("gateway-discovery create Service %s to api-gateway in ModuleChanged Event failed, %s. it can only recover in dataSynchronization", event.Module, err.Error())
			return err
		}
		blog.Infof("gateway-discovery create Service %s in ModuleChanged Event successfully, serviceName: %s", event.Module, event.Svc.Name)
	} else {
		//only update Target for Service
		//todo(DeveloperJim): discovery needs to check service plugins changed when version updates
		if err := s.regMgr.ReplaceTargetByService(event.Svc, event.Svc.Backends); err != nil {
			blog.Errorf("gateway-discovery update Service %s Target failed, %s", event.Svc.Name, err.Error())
			return err
		}
		blog.Infof("gateway-discovery update Target for Service %s in api-gateway successfully, serviceName: %s", event.Module, event.Svc.Name)
	}
	return nil
}

func (s *DiscoveryServer) formatBCSServerInfo(module string) ([]*types.ServerInfo, error) {
	originals, err := s.discovery.GetModuleServers(module)
	if err != nil {
		blog.Errorf("gateway-discovery get module %s information from module-discovery failed, %s", module, err.Error())
		return nil, err
	}
	var svcs []*types.ServerInfo
	for _, info := range originals {
		data := info.(string)
		var svc *types.ServerInfo
		if err := json.Unmarshal([]byte(data), svc); err != nil {
			blog.Errorf("gateway-discovery handle module %s json unmarshal failed, %s", module, err.Error())
			continue
		}
		svcs = append(svcs, svc)
	}
	return svcs, nil
}
