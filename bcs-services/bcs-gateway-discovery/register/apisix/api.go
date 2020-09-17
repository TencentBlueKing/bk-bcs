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

package apisix

import (
	"crypto/tls"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"

	"github.com/DeveloperJim/gokong"
)

const (
	protocolHTTP = "http"
	protocolGRPC = "grpc"

	protocolHTTPS = "https"
	protocolGRPCS = "grpcs"
)

//New create Register implementation for kong
// return empty
func New(addr []string, config *tls.Config) (register.Register, error) {
	kcfg := &gokong.Config{
		HostAddress: addr[0],
	}

	reg := &apiRegister{
		kAddrs:  addr,
		kClient: gokong.NewClient(kcfg),
	}
	return reg, nil
}

//apiRegister apisix register implementation
type apiRegister struct {
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
func (r *apiRegister) CreateService(svc *register.Service) error {
	var err error
	kreq := kongServiceRequestConvert(svc)
	// 1. create specified service information
	ksvc, err := r.kClient.Services().Create(kreq)
	if err != nil {
		blog.Errorf("kong register create Service %s[%s] failed, %s", svc.Name, svc.Host, err.Error())
		return err
	}
	// create service plugins
	if svc.Plugin != nil {
		pReqs := kongPluginConvert(svc.Plugin, ksvc.Id, "service")
		for _, pluReq := range pReqs {
			splugin, err := r.kClient.Plugins().Create(pluReq)
			if err != nil {
				//todo(DeveloperJim): shall we clean created service, that we can retry in next data synchronization
				blog.Errorf("kong register create plugin %s for Service %s failed, %s", pluReq.Name, svc.Name, err.Error())
				return err
			}
			blog.Infof("kong register create plugin for service %s successfully, plugin ID: %s/%s", svc.Name, splugin.Id, splugin.Name)
		}
	}
	// 2. create service relative route rules
	for _, route := range svc.Routes {
		kr := kongRouteConvert(&route, ksvc.Id)
		kroute, err := r.kClient.Routes().Create(kr)
		if err != nil {
			blog.Errorf("kong register create route for Service %s failed, %s", svc.Name, err.Error())
			return err
		}
		if route.Plugin != nil {
			rReqs := kongPluginConvert(route.Plugin, kroute.Id, "route")
			for _, pluReq := range rReqs {
				rplugin, err := r.kClient.Plugins().Create(pluReq)
				if err != nil {
					//todo(DeveloperJim): roll back discussion
					blog.Errorf("kong register create plugin %s for route %s failed, %s", pluReq.Name, route.Name, err.Error())
					return err
				}
				blog.Infof("kong register create plugins for route %s successfully, pluginID: %s/%s", route.Name, rplugin.Id, rplugin.Name)
			}
		}
	}
	// 3. create service relative Upstream & targets
	kupstrreq := &gokong.UpstreamRequest{
		Name: svc.Host,
	}
	//setting tags
	if len(svc.Labels) != 0 {
		for _, v := range svc.Labels {
			kupstrreq.Tags = append(kupstrreq.Tags, gokong.String(v))
		}
	}
	kUpstream, err := r.kClient.Upstreams().Create(kupstrreq)
	if err != nil {
		blog.Errorf("kong register create upstream %s for service %s failed, %s", svc.Host, svc.Name, err.Error())
		return err
	}
	blog.Infof("kong register create upstream %s [%s] successfully", kUpstream.Name, kUpstream.Id)
	//create targets for upstream
	for _, backend := range svc.Backends {
		targetReq := &gokong.TargetRequest{
			Target: backend.Target,
			Weight: backend.Weight,
		}
		ktarget, err := r.kClient.Targets().CreateFromUpstreamName(kUpstream.Name, targetReq)
		if err != nil {
			blog.Errorf("kong register create target %s for upstream %s failed, %s. try next one ", targetReq.Target, kUpstream.Name, err.Error())
			continue
		}
		blog.Infof("kong register create target %s[%s] for upstream %s successfully", targetReq.Target, *ktarget.Id, kUpstream.Name)
	}
	return nil
}

//UpdateService update specifed Service, if service does not exist, return error
func (r *apiRegister) UpdateService(svc *register.Service) error {
	return fmt.Errorf("not implemented")
}

//GetService get specified service by name, if no service, return nil
func (r *apiRegister) GetService(svc string) (*register.Service, error) {
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
// @param service: at least setting Name & Host for deletion
func (r *apiRegister) DeleteService(svc *register.Service) error {
	if svc.Host == "" || svc.Name == "" {
		return fmt.Errorf("service lost Name or Host")
	}
	var err error
	//clean route, route name is same with service
	if err = r.kClient.Routes().DeleteByName(svc.Name); err != nil {
		blog.Errorf("kong register delete service %s relative route failed, %s", svc, err.Error())
		return err
	}
	blog.V(3).Infof("kong register delete route %s success", svc.Name)
	err = r.kClient.Services().DeleteServiceByName(svc.Name)
	if err != nil {
		blog.Errorf("kong register delete service by name %s failed, %s", svc, err.Error())
		return err
	}
	blog.V(3).Infof("kong register delete service %s success", svc.Name)
	//* clean upstream
	if err = r.kClient.Upstreams().DeleteByName(svc.Host); err != nil {
		blog.Errorf("kong register delete service %s relative Upstream %s failed, %s", svc.Name, svc.Host, err.Error())
		return err
	}
	blog.Infof("kong register delete service %s upstream %s success", svc.Name, svc.Host)
	return nil
}

//ListServices get all existence services
func (r *apiRegister) ListServices() ([]*register.Service, error) {
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
func (r *apiRegister) GetTargetByService(svc *register.Service) ([]register.Backend, error) {
	if svc == nil || len(svc.Host) == 0 {
		return nil, fmt.Errorf("neccessary service info lost")
	}
	ktargets, err := r.kClient.Targets().GetTargetsFromUpstreamId(svc.Host)
	if err != nil {
		blog.Errorf("kong register get targets by service %s failed, %s", svc.Host, err.Error())
		return nil, err
	}
	var backends []register.Backend
	for _, target := range ktargets {
		backend := register.Backend{
			Target: *target.Target,
			Weight: *target.Weight,
		}
		backends = append(backends, backend)
	}
	return backends, nil
}

//ReplaceTargetByService replace specified service backend list
// so we don't care what original backend list are
func (r *apiRegister) ReplaceTargetByService(svc *register.Service, backends []register.Backend) error {
	//get original targets
	if svc.Name == "" || svc.Host == "" {
		return fmt.Errorf("service info lost Name or Host")
	}
	if len(backends) == 0 {
		return fmt.Errorf("lost backends list")
	}
	targets, err := r.kClient.Targets().GetTargetsFromUpstreamId(svc.Host)
	if err != nil {
		blog.Errorf("kong register get upstream %s targets failed, %s", svc.Host, err.Error())
		return err
	}
	cleanTargets := make(map[string]*gokong.Target)
	for _, target := range targets {
		cleanTargets[*target.Target] = target
	}
	addTargets := make(map[string]*gokong.TargetRequest)
	for _, backend := range backends {
		_, ok := cleanTargets[backend.Target]
		if ok {
			blog.V(3).Infof("upstream %s target %s already exist, skip Replace", svc.Host, backend.Target)
			delete(cleanTargets, backend.Target)
			continue
		}
		//this is new Target we need to add
		addTargets[backend.Target] = &gokong.TargetRequest{
			Target: backend.Target,
			Weight: backend.Weight,
		}
	}
	if len(addTargets) != 0 {
		for k, v := range addTargets {
			ktarget, err := r.kClient.Targets().CreateFromUpstreamName(svc.Host, v)
			if err != nil {
				blog.Errorf("kong add New target %s for upstream %s failed, %s", k, svc.Host, err.Error())
				continue
			}
			blog.Infof("kong add new target %s[%d] for upstream %s success", k, ktarget.Id, svc.Host)
		}
	}
	if len(cleanTargets) != 0 {
		for k, v := range cleanTargets {
			if err := r.kClient.Targets().DeleteFromUpstreamById(svc.Host, *v.Id); err != nil {
				blog.Errorf("kong clean out-of-dated target %s for upstream %s failed, %s", k, svc.Host, err.Error())
				continue
			}
			blog.Infof("kong clean out-of-dated target %s[%d] for upstream %s success", k, v.Id, svc.Host)
		}
	}
	return nil
}

//DeleteTargetByService clean all backend list for service
func (r *apiRegister) DeleteTargetByService(svc *register.Service) error {
	return fmt.Errorf("not implemented")
}

func (r *apiRegister) deletePlugins(resource string, plugins []*gokong.Plugin) error {
	for _, plugin := range plugins {
		if err := r.kClient.Plugins().DeleteById(plugin.Id); err != nil {
			blog.Errorf("kong register delete resource %s plugin %s[%s] failed, %s", resource, plugin.Name, plugin.Id, err.Error())
			return err
		}
	}
	blog.V(3).Infof("kong register clean %s all plugins success", resource)
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
		Path:     &svc.Path,
		Retries:  gokong.Int(svc.Retries),
	}
	if len(svc.Labels) != 0 {
		for _, v := range svc.Labels {
			ksvc.Tags = append(ksvc.Tags, gokong.String(v))
		}
	}
	return ksvc
}

//kongRouteConvert convert inner service to kong Route, tls feature supported in default.
//args: inner route definition; kong service Id
func kongRouteConvert(route *register.Route, ID *string) *gokong.RouteRequest {
	var protocols []*string
	if route.Protocol == protocolHTTP {
		protocols = []*string{gokong.String(protocolHTTP), gokong.String(protocolHTTPS)}
	} else if route.Protocol == protocolGRPC {
		protocols = []*string{gokong.String(protocolGRPC), gokong.String(protocolGRPCS)}
	}
	kr := &gokong.RouteRequest{
		Name:      &route.Name,
		Protocols: protocols,
		Paths:     gokong.StringSlice(route.Paths),
		StripPath: gokong.Bool(route.PathRewrite),
	}
	if len(route.Header) != 0 {
		//setting header filter
		kr.Header = make(map[string][]*string)
		for k, v := range route.Header {
			kr.Header[k] = []*string{gokong.String(v)}
		}
	}
	if len(route.Labels) != 0 {
		//setting route tags
		for _, v := range route.Labels {
			kr.Tags = append(kr.Tags, gokong.String(v))
		}
	}
	kr.Service = gokong.ToId(*ID)
	return kr
}

//kongPluginConvert convert inner service request plugin to request-transformer
func kongPluginConvert(plugin *register.Plugins, ID *string, tys string) []*gokong.PluginRequest {
	var plugins []*gokong.PluginRequest
	if plugin.HeadOption != nil {
		plu := kongReqTransformerConvert(plugin.HeadOption, *ID, tys)
		plugins = append(plugins, plu)
	}
	if plugin.AuthOption != nil {
		plu := kongBKBCSAuthConvert(plugin.AuthOption, *ID, tys)
		plugins = append(plugins, plu)
	}
	return plugins
}

//kongReqTransformerConvert convert inner service request plugin to request-transformer
func kongReqTransformerConvert(option *register.HeaderOption, ID string, tys string) *gokong.PluginRequest {
	pr := &gokong.PluginRequest{
		Name: "request-transformer",
	}
	if tys == "service" {
		pr.ServiceId = gokong.ToId(ID)
	} else {
		pr.RouteId = gokong.ToId(ID)
	}
	//setting clean operation
	pr.Config = make(map[string]interface{})
	if len(option.Clean) != 0 {
		pr.Config["remove"] = &httpTransformer{
			Body:     []*string{},
			Headers:  gokong.StringSlice(option.Clean),
			QueryStr: []*string{},
		}
	}
	//add operation
	if len(option.Add) != 0 {
		var values []string
		for k, v := range option.Add {
			value := fmt.Sprintf("%s: %s", k, v)
			values = append(values, value)
		}
		pr.Config["add"] = &httpTransformer{
			Body:     []*string{},
			Headers:  gokong.StringSlice(values),
			QueryStr: []*string{},
		}
	}
	//replace operation
	if len(option.Replace) != 0 {
		var values []string
		for k, v := range option.Replace {
			value := fmt.Sprintf("%s: %s", k, v)
			values = append(values, value)
		}
		pr.Config["replace"] = &httpTransformer{
			Body:     []*string{},
			Headers:  gokong.StringSlice(values),
			QueryStr: []*string{},
		}
	}
	return pr
}

//kongBKBCSAuthConvert convert inner service request plugin to request-transformer
func kongBKBCSAuthConvert(option *register.BCSAuthOption, id string, tys string) *gokong.PluginRequest {
	pr := &gokong.PluginRequest{
		Name: option.Name,
	}
	if tys == "service" {
		pr.ServiceId = gokong.ToId(id)
	} else {
		pr.RouteId = gokong.ToId(id)
	}
	pr.Config = make(map[string]interface{})
	//setting clean operation
	pr.Config["bkbcs_auth_endpoints"] = option.AuthEndpoints
	pr.Config["module"] = option.Module
	pr.Config["token"] = option.AuthToken
	return pr
}

// httpTransformer holder for http plugins
type httpTransformer struct {
	Body     []*string `json:"body" yaml:"body"`
	Headers  []*string `json:"headers" yaml:"headers"`
	QueryStr []*string `json:"querystring" yaml:"querystring"`
}
