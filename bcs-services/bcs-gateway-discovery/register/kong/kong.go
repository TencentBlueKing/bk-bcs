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

package kong

import (
	"bcs/control-common/blog"
	"crypto/tls"

	"bk-bcs/bcs-services/bcs-gateway-discovery/register"

	"github.com/DeveloperJim/gokong"
)

//New create Register implementation for kong
// return empty
func New(addr []string, config *tls.Config) (register.Register, error) {
	kcfg := &gokong.Config{
		HostAddress: addr[0],
	}

	reg := &kRegister{
		kAddrs:  addr,
		kClient: gokong.NewClient(kcfg),
	}
	return reg, nil
}

//kRegister kong register implementation
type kRegister struct {
	kAddrs  []string
	kClient *gokong.KongAdminClient
}

//CreateService create Service interface, if service already exists, return error
// create service include three operations:
// 1. create specified service information, including plugins
// 2. create service relative route rules, including plugins
// 3. create service relative Upstream & targets
// proxy rules from bcs-api-gateway:
// we authenticate in stage of route, then post to service when authentication success.
// in stage of service, we clean original Authorization information and switch to inner
// authentication token for different bkbcs modules
func (r *kRegister) CreateService(svc *register.Service) error {
	kreq := kongServiceRequestConvert(svc)
	// 1. create specified service information
	ksvc, err := r.kClient.Services().Create(kreq)
	if err != nil {
		blog.Errorf("kong register create Service %s[%s] failed, %s", svc.Name, svc.Host, err.Error())
		return err
	}
	// create service plugins
	if svc.HeadOption != nil {
		blog.Infof("kong register create plugin request-transformer for service %s", svc.Name)
		pReq := kongRequestTransformerConvert(svc.HeadOption, ksvc.Id)
		splugin, perr := r.kClient.Plugins().Create(pReq)
		if perr != nil {
			//todo(DeveloperJim): shall we clean created service, that we can retry in next data synchronization
			blog.Errorf("kong register create plugin request-transformer for Service %s failed, %s", svc.Name, perr.Error())
			return err
		}
		blog.Infof("kong register create plugin request-transformer for service %s successfully, plugin ID: %s", svc.Name, splugin.Id)
	}
	// 2. create service relative route rules
	kr := kongRouteConvert()
	r.kClient.Routes().Create()
	return nil
}

//UpdateService update specifed Service, if service does not exist, return error
func (r *kRegister) UpdateService(svc *register.Service) error {
	return nil
}

//GetService get specified service by name, if no service, return nil
func (r *kRegister) GetService(svc string) (*register.Service, error) {
	kSvc, err := r.kClient.Services().GetServiceByName(svc)
	if err != nil {
		blog.Errorf("kong register get service %s failed, %s", svc, err.Error())
		return nil, err
	}
	if kSvc == nil {
		blog.Warnf("kong register get no Service named %s", svc)
		return nil, nil
	}
	//convert data structure
	return innerServiceConvert(kSvc), nil
}

//DeleteService delete specified service, success even if no such service
func (r *kRegister) DeleteService(svc string) error {
	err := r.kClient.Services().DeleteServiceByName(svc)
	if err != nil {
		blog.Errorf("kong register delete service by name %s failed, %s", svc, err.Error())
		return err
	}
	return nil
}

//ListServices get all existence services
func (r *kRegister) ListServices() ([]*register.Service, error) {
	query := &gokong.ServiceQueryString{
		Size: 200,
	}
	kSvcs, err := r.kClient.Services().GetServices(query)
	if err != nil {
		blog.Errorf("kong register list all services failed, %s", err.Error())
		return nil, err
	}
	if len(kSvcs) == 0 {
		blog.Warnf("kong register list no services")
		return nil, nil
	}
	var inDatas []*register.Service
	for _, ksvc := range kSvcs {
		s := innerServiceConvert(ksvc)
		inDatas = append(inDatas, s)
	}
	return inDatas, nil
}

//GetTargetByService get service relative backends
func (r *kRegister) GetTargetByService(svc *register.Service) ([]register.Backend, error) {
	return nil, nil
}

//UpdateTargetByService replace specified service backend list
// so we don't care what original backend list are
func (r *kRegister) UpdateTargetByService(svc *register.Service, backends []register.Backend) error {
	return nil
}

//DeleteTargetByService clean all backend list for service
func (r *kRegister) DeleteTargetByService(svc *register.Service) error {
	return nil
}

//innerServiceConvert convert kong service to inner service definition
func innerServiceConvert(ksvc *gokong.Service) *register.Service {
	svc := &register.Service{
		Name:     *ksvc.Name,
		Protocol: *ksvc.Protocol,
		Host:     *ksvc.Host,
		Port:     uint(*ksvc.Port),
		Path:     *ksvc.Path,
	}
	return svc
}

//kongServiceConvert convert inner service to kong service
func kongServiceRequestConvert(svc *register.Service) *gokong.ServiceRequest {
	ksvc := &gokong.ServiceRequest{
		Name:     &svc.Name,
		Protocol: &svc.Protocol,
		Host:     &svc.Host,
		Path:     gokong.String(svc.Path),
		Retries:  gokong.Int(svc.Retries),
	}
	return ksvc
}

//kongRouteConvert convert inner service to kong Route
func kongRouteConvert(route *register.Route) *gokong.RouteRequest {
	kr := &gokong.RouteRequest{
		Name:      &route.Name,
		Protocols: []*string{gokong.String(route.Protocol)},
		Paths:     gokong.StringSlice(route.Paths),
		StripPath: gokong.Bool(route.PathRewrite),
	}
	if len(route.Header) != 0 {
		//setting header filter
	}
	if len(route.Labels) != 0 {
		//setting route tags
	}

	return kr
}

//kongRouteConvert convert inner service to kong Route
func kongRequestTransformerConvert(option *register.HeaderOption, ID *string) *gokong.PluginRequest {
	pr := &gokong.PluginRequest{
		Name:      "request-transformer",
		ServiceId: gokong.ToId(*ID),
	}
	//
	return pr
}
