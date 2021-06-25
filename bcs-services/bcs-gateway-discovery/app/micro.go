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
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/utils"
)

// this file contains all new features about etcd registry
// * handle etcd registry event stream
// * convert etcd registry data structure to local register definition

// getMicroModuleName get specified bkbcs module name from
func getMicroModuleName(fullName string) string {
	shortName := strings.ReplaceAll(fullName, defaultDomain, "")
	IDName := strings.Split(shortName, ".")
	return IDName[len(IDName)-1]
}

// getMicroModuleClusterID get clusterID from module host
// like 100032.mesosdriver.bkbcs.tencent.com
func getMicroModuleClusterID(host string) string {
	IDName := strings.Split(host, ".")
	return IDName[0]
}

//microModuleEvent event notification callback
func (s *DiscoveryServer) microModuleEvent(module string) {
	//get event notification
	event := &ModuleEvent{
		// module info: 100032.mesosdriver, storage, meshmanager
		Module:  module,
		GoMicro: true,
	}
	s.evtCh <- event
	utils.ReportDiscoveryEventChanLengthInc()
}

func (s *DiscoveryServer) handleMicroChange(event *ModuleEvent) {
	module := getMicroModuleName(event.Module)
	//check grpc service information registration
	if _, ok := defaultGrpcModules[module]; ok {
		//get specified module info and construct data for refresh
		svcs, err := s.formatEtcdInfo(event.Module, false)
		if err != nil {
			blog.Errorf("discovery formate module %s grpc service failed, %s", event.Module, err.Error())
			return
		}
		event.Svc = svcs
		if err := s.gatewayServiceSync(event); err != nil {
			blog.Errorf("discovery sync go micro grpc service %s failed, %s", event.Module, err.Error())
			return
		}
	}
	//check http service information registration
	if _, ok := defaultHTTPModules[module]; ok {
		if module == modules.BCSModuleMesosdriver {
			id := getMicroModuleClusterID(event.Module)
			if !s.isClusterRestriction(id) {
				s.clusterRestricted(id)
				blog.Warnf("cluster mesosdriver %s registry changed!!!!!", event.Module)
			}
		}

		//http service create/update
		svcs, err := s.formatEtcdInfo(event.Module, true)
		if err != nil {
			blog.Errorf("discovery formate module %s http service failed, %s", event.Module, err.Error())
			return
		}
		event.Svc = svcs
		if err := s.gatewayServiceSync(event); err != nil {
			blog.Errorf("discovery sync go micro grpc service %s failed, %s", event.Module, err.Error())
			return
		}
	}
}

// formatEtcdInfo format internal service info according module info
//@param: module, bkbcs module info, like 10032.mesosdriver, storage, meshsmanager
//@param: http, flag for http route conversion
func (s *DiscoveryServer) formatEtcdInfo(module string, http bool) (*register.Service, error) {
	service, err := s.microDiscovery.GetModuleServer(module)
	if err != nil {
		blog.Errorf("get module %s information from micro-discovery failed, %s", module, err.Error())
		return nil, err
	}
	if service == nil {
		//no service in local cache, it means deletion event
		//in this stage, we suppose bkbcs need all modules for long time.
		//we don't delete service in synchronization and reserves service until that module registes back.
		//then we can replace upstream target with new information simplly.
		//if api-gateway routable rules change, we change api-gateway by release maintenance
		blog.Warnf("module %s is not in micro-discovery cache", module)
		return nil, fmt.Errorf("no module in cache")
	}
	blog.V(5).Infof("get module %s string detail: %+v", module, service)
	if len(service.Nodes) == 0 {
		blog.Errorf("micro-discovery has no available node of %s, pay more attention", module)
		return nil, fmt.Errorf("no module node in cache")
	}
	var rSvcs *register.Service
	bkbcsName := getMicroModuleName(module)
	if http {
		//data structure conversion
		rSvcs, err = s.adapter.GetHTTPService(bkbcsName, service)
		if err != nil {
			blog.Errorf("converts micro http module %s registry to api-gateway info failed, %s", service.Name, err.Error())
			return nil, err
		}
	} else {
		//grpc data structure conversion
		rSvcs, err = s.adapter.GetGrpcService(bkbcsName, service)
		if err != nil {
			blog.Errorf("converts micro grpc module %s registry to api-gateway info failed, %s", service.Name, err.Error())
			return nil, err
		}

	}
	return rSvcs, nil
}

// formatMultiEtcdService use for data synchronization
func (s *DiscoveryServer) formatMultiEtcdService() ([]*register.Service, error) {
	svcs, err := s.microDiscovery.ListAllServer()
	if err != nil {
		blog.Errorf("discovery server list all registry service failed, %s", err.Error())
		return nil, fmt.Errorf("list all micro registry service err")
	}
	if len(svcs) == 0 {
		blog.Warnf("no module in etcd registry...")
		return nil, nil
	}
	var allServices []*register.Service
	for _, svc := range svcs {
		module := getMicroModuleName(svc.Name)
		//check grpc route conversion
		if _, ok := defaultGrpcModules[module]; ok {
			rsvc, err := s.adapter.GetGrpcService(module, svc)
			if err != nil {
				blog.Errorf("converts module %s grpc registry info to api-gateway info failed, %s", module, err.Error())
				continue
			}
			allServices = append(allServices, rsvc)
			blog.V(5).Infof("etcd registry module %s[%s] grpc conversion successfully", svc.Name, module)
		}
		//check http route rules conversion
		//! pay more attention, modules that don't support grpc must be compatible in http conversion
		if _, ok := defaultHTTPModules[module]; ok {
			rsvc, err := s.adapter.GetHTTPService(module, svc)
			if err != nil {
				blog.Errorf("converts module %s http registry info to api-gateway info failed, %s", svc.Name, err.Error())
				continue
			}
			allServices = append(allServices, rsvc)
			//! compatible discovery for mesosdriver, mesosdriver should support zookeeper registry
			//! & etcd registry. but actually, it's hard to update all cluster mesos driver at the
			//! same time. so when discovery find that cluster mesosdriver update to etcd registry
			//! version, discovery will restrict that only retreve discovery information from etcd
			//! registry and ignore same information from zookeeper.
			if module == modules.BCSModuleMesosdriver {
				id := getMicroModuleClusterID(svc.Name)
				if !s.isClusterRestriction(id) {
					s.clusterRestricted(id)
				}
			}
			blog.V(5).Infof("etcd registry module %s http conversion successfully", svc.Name)
		}
	}
	return allServices, nil
}
