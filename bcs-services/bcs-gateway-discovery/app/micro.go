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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"
)

// this file contains all new features about etcd registry
// * handle etcd registry event stream
// * convert etcd registry data structure to local register definition

//microModuleEvent event notification callback
func (s *DiscoveryServer) microModuleEvent(module string) {
	if !s.bcsRegister.IsMaster() {
		blog.Infof("gateway-discovery instance is not master, skip module %s event notification for micro registry", module)
		return
	}
	//get event notification
	event := &ModuleEvent{
		Module:  module,
		GoMicro: true,
	}
	s.evtCh <- event
}

func (s *DiscoveryServer) handleMicroChange(event *ModuleEvent) {
	//check grpc service information registration
	if strings.Contains(s.option.Etcd.GrpcModules, event.Module) {
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
	if strings.Contains(s.option.Etcd.HTTPModules, event.Module) {
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
		//if api-gateway routable rules change, we change api-gateway throught release maintenance
		blog.Warnf("module %s is not in micro-discovery cache", module)
		return nil, fmt.Errorf("no module in cache")
	}
	blog.V(5).Infof("get module %s string detail: %+v", module, service)
	if len(service.Nodes) == 0 {
		blog.Errorf("micro-discovery has no available node of %s, pay more attention", module)
		return nil, fmt.Errorf("no module node in cache")
	}
	var rSvcs *register.Service
	if http {
		//data structure conversion
		rSvcs, err = s.adapter.GetHTTPService(module, service)
		if err != nil {
			blog.Errorf("converts module %s ServerInfo to api-gateway info failed, %s", module, err.Error())
			return nil, err
		}
	} else {
		//grpc data structure conversion
		rSvcs, err = s.adapter.GetGrpcService(module, service)
		if err != nil {
			blog.Errorf("converts micro module %s to api-gateway info failed, %s", module, err.Error())
			return nil, err
		}

	}
	return rSvcs, nil
}

// formatMultiEtcdService use for data synchronization
func (s *DiscoveryServer) formatMultiEtcdService() ([]*register.Service, error) {
	var allServices []*register.Service
	//grpc
	for name := range defaultGrpcModules {
		svc, err := s.formatEtcdInfo(name, false)
		if err != nil {
			return nil, err
		}
		allServices = append(allServices, svc)
	}

	//http
	for _, name := range defaultHTTPModules {
		svc, err := s.formatEtcdInfo(name, true)
		if err != nil {
			return nil, err
		}
		allServices = append(allServices, svc)
	}

	return allServices, nil
}
