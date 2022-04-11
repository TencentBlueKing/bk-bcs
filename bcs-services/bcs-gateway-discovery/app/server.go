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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	discoverys "github.com/Tencent/bk-bcs/bcs-common/pkg/module-discovery"
	modulediscovery "github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register/apisix"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register/kong"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/utils"

	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
)

//New create
func New(root context.Context) *DiscoveryServer {
	cxt, cfunc := context.WithCancel(root)
	s := &DiscoveryServer{
		exitCancel: cfunc,
		exitCxt:    cxt,
		evtCh:      make(chan *ModuleEvent, 12),
		clusterID:  make(map[string]string),
	}
	return s
}

//ModuleEvent event
type ModuleEvent struct {
	// Module name
	Module string
	// GoMicro flag for go-micro registry
	GoMicro bool
	// flag for delete
	Deletion bool
	// Svc api-gateway service definition
	Svc *register.Service
}

//DiscoveryServer holds all resources for services discovery
type DiscoveryServer struct {
	option *ServerOptions
	//manager for gateway information register
	regMgr register.Register
	//adapter for service structure conversion
	adapter *Adapter
	//bk-bcs modules discovery for backend service list
	discovery discoverys.ModuleDiscovery
	//go micro version discovery
	microDiscovery modulediscovery.Discovery
	//exit func
	exitCancel context.CancelFunc
	//exit context
	exitCxt context.Context
	//Event channel for module-discovery callback
	evtCh chan *ModuleEvent
	// clusterID to prevent zookeeper discovery
	clusterID   map[string]string
	clusterLock sync.RWMutex
	clusterCli  cm.ClusterManagerClient
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
	//init gateway manager
	gatewayAddrs := strings.Split(option.AdminAPI, ",")
	tlsConfig, err := option.GetClientTLS()
	if err != nil {
		return err
	}
	if option.AdminType == "kong" {
		s.regMgr, err = kong.New(gatewayAddrs, tlsConfig)
		if err != nil {
			blog.Errorf("gateway init kong admin api register implementation failed, %s", err.Error())
			return err
		}
	} else {
		s.regMgr, err = apisix.New(gatewayAddrs, tlsConfig, option.AdminToken, option.GatewayMetricsEnabled)
		if err != nil {
			blog.Errorf("gateway init apisix admin api register implementation failed, %s", err.Error())
			return err
		}
	}

	//init etcd registry feature with modulediscovery base on micro.Registry
	if err := s.turnOnEtcdFeature(option); err != nil {
		return err
	}

	defaultModules = append(defaultModules, strings.Split(option.Modules, ",")...)
	//init service data adapter
	s.adapter = NewAdapter(option)
	//init module disovery
	s.discovery, err = discoverys.NewDiscoveryV2(option.ZkConfig.BCSZk, defaultModules)
	if err != nil {
		blog.Errorf("gateway init services discovery failed, %s", err.Error())
		return err
	}
	s.discovery.RegisterEventFunc(s.moduleEventNotifycation)
	return nil
}

func (s *DiscoveryServer) turnOnEtcdFeature(option *ServerOptions) error {
	blog.Infof("gateway-discovery check etcd registry feature turn on, try to initialize etcd registry")
	etcdTLSConfig, err := option.GetEtcdRegistryTLS()
	if err != nil {
		blog.Errorf("gateway init etcd registry feature failed, no tlsConfig parsed correctlly, %s", err.Error())
		return err
	}
	//initialize micro registry
	addrs := strings.Split(option.Etcd.Address, ",")
	mregistry := etcd.NewRegistry(
		registry.Addrs(addrs...),
		registry.TLSConfig(etcdTLSConfig),
	)
	if err := mregistry.Init(); err != nil {
		blog.Errorf("gateway init etcd registry feature failed, %s", err.Error())
		return err
	}
	//clean duplicated watch module for registry
	noDuplicated := make(map[string]string)
	for _, v := range strings.Split(option.Etcd.GrpcModules, ",") {
		key := strings.ToLower(v)
		defaultGrpcModules[key] = v
		noDuplicated[key] = key
	}
	for _, v := range strings.Split(option.Etcd.HTTPModules, ",") {
		key := strings.ToLower(v)
		defaultHTTPModules[key] = v
		noDuplicated[key] = key
	}
	var modules []string
	for k := range noDuplicated {
		modules = append(modules, k)
	}
	s.microDiscovery = modulediscovery.NewDiscovery(modules, s.microModuleEvent, mregistry)
	blog.Infof("gateway init etcd registry success, try to init bkbcs module watch")
	return s.microDiscovery.Start()
}

//Run running all necessary conversion logic, block
func (s *DiscoveryServer) Run() error {
	//check master status first
	if err := s.dataSynchronization(); err != nil {
		blog.Errorf("gateway-discovery first data synchronization failed, %s", err.Error())
		return err
	}
	tick := time.NewTicker(time.Second * 60)
	defer tick.Stop()
	for {
		select {
		case <-s.exitCxt.Done():
			s.discovery.Stop()
			s.microDiscovery.Stop()
			close(s.evtCh)
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
			utils.ReportDiscoveryEventChanLengthDec()
			blog.Infof("gateway-discovery got module %s changed event", evt.Module)
			//ready to update specified module proxy rules
			if evt.GoMicro {
				s.handleMicroChange(evt)
			} else {
				s.handleModuleChange(evt)
			}
		}
	}
}

//Stop all backgroup routines
func (s *DiscoveryServer) Stop() {
	s.exitCancel()
}

//dataSynchronization sync all data from bk bcs service discovery to gateway
func (s *DiscoveryServer) dataSynchronization() error {
	blog.V(3).Infof("gateway-discovery instance is master, ready to sync all datas")
	//first get all gateway route information
	regisetedService, err := s.regMgr.ListServices()
	if err != nil {
		blog.Errorf("gateway-discovery get all registered Service from Register failed, %s. wait for next tick", err.Error())
		return err
	}
	regisetedMap := make(map[string]*register.Service)
	if len(regisetedService) == 0 {
		blog.Warnf("gateway-discovery finds no registered service from Register, maybe this is first synchronization.")
	} else {
		for _, srv := range regisetedService {
			blog.V(3).Infof("gateway-discovery check Service %s is under regiseted", srv.Name)
			regisetedMap[srv.Name] = srv
		}
	}

	var allCaches []*register.Service
	//* module step: check etcd registry feature, if feature is on,
	// get all module information from etcd discovery
	if s.option.Etcd.Feature {
		etcdModules, err := s.formatMultiEtcdService()
		if err != nil {
			blog.Errorf("discovery format etcd service info when in Synchronization, %s", err.Error())
			return err
		}
		if len(etcdModules) == 0 {
			blog.Warnf("gateway-discovery finds no bk-bcs service in Micro-discovery, please check bk-bcs discovery machinery")
		} else {
			allCaches = append(allCaches, etcdModules...)
		}
	}
	//* module step: get all register module information from zookeeper discovery
	localCaches, err := s.formatMultiServerInfo(defaultModules)
	if err != nil {
		blog.Errorf("disovery formate zookeeper Service info when in Synchronization, %s", err.Error())
		return err
	}
	//check zookeeper module info
	if len(localCaches) == 0 {
		blog.Warnf("gateway-discovery finds no bk-bcs service in module-discovery, please check bk-bcs discovery machinery")
	} else {
		allCaches = append(allCaches, localCaches...)
	}
	//udpate datas in gateway
	for _, local := range allCaches {
		svc, ok := regisetedMap[local.Name]
		if ok {
			//service reigsted, we affirm that proxy rule is correct
			// so just update backend targets info, if rules of plugins & routes
			// change frequently, we need to verify all changes between oldSvc & newSvc.
			// but now, we confirm that rules are stable. operations can be done quickly by manually
			if err := s.regMgr.ReplaceTargetByService(local, local.Backends); err != nil {
				blog.Errorf("gateway-discovery update Service %s backend failed in synchronization, %s. backend %+v", svc.Name, err.Error(), local.Backends)
				continue
			}
			blog.V(5).Infof("Update serivce %s backend %+v successfully", svc.Name, local.Backends)
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
	//we don't clean additional datas in api-gateway,
	// because we allow registe service information in api-gateway manually
	return nil
}

func (s *DiscoveryServer) gatewayServiceSync(event *ModuleEvent) error {
	//update service route
	exist, err := s.regMgr.GetService(event.Svc.Name)
	if err != nil {
		blog.Errorf("gateway-discovery get register Service %s failed, %s", event.Module, err.Error())
		return err
	}
	if exist == nil {
		blog.Infof("gateway-discovery find no %s module in api-gateway, try to create...", event.Module)
		if err := s.regMgr.CreateService(event.Svc); err != nil {
			blog.Errorf("gateway-discovery create Service %s to api-gateway failed, %s", event.Module, err.Error())
			return err
		}
		blog.Infof("gateway-discovery create Service %s successfully, serviceName: %s", event.Module, event.Svc.Name)
	} else {
		//only update Target for Service
		if err := s.regMgr.ReplaceTargetByService(event.Svc, event.Svc.Backends); err != nil {
			blog.Errorf("gateway-discovery update Service %s Target failed, %s", event.Svc.Name, err.Error())
			return err
		}
		blog.Infof("gateway-discovery update Target for Service %s in api-gateway successfully, serviceName: %s", event.Module, event.Svc.Name)
	}
	return nil
}

//detailServiceVerification all information including service/plugin/target check
func (s *DiscoveryServer) detailServiceVerification(newSvc *register.Service, oldSvc *register.Service) {
	//todo(DeveloperJim): we need complete verification if plugin & route rules changed frequently, not now
}

func (s *DiscoveryServer) isClusterRestriction(clusterID string) bool {
	cluster := clusterID
	if strings.Contains(clusterID, "-") {
		items := strings.Split(clusterID, "-")
		cluster = items[len(items)-1]
	}
	s.clusterLock.RLock()
	defer s.clusterLock.RUnlock()
	if _, ok := s.clusterID[cluster]; ok {
		return true
	}
	return false
}

func (s *DiscoveryServer) clusterRestricted(clusterID string) {
	blog.Infof("cluster %s is ready to restricted discovery machinery to etcd registry", clusterID)
	cluster := clusterID
	if strings.Contains(clusterID, "-") {
		items := strings.Split(clusterID, "-")
		cluster = items[len(items)-1]
	}
	s.clusterLock.Lock()
	defer s.clusterLock.Unlock()
	s.clusterID[cluster] = clusterID
}
