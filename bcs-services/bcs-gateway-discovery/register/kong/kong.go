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
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/utils"

	"github.com/kevholditch/gokong"
)

const (
	protocolHTTP = "http"
	protocolGRPC = "grpc"

	protocolHTTPS = "https"
	protocolGRPCS = "grpcs"

	// kongAdmin system kong
	kongAdmin = "kong_admin"
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
	var (
		err        error
		startedAll = time.Now()
	)

	defer reportRegisterKongMetrics("CreateService", err, startedAll)
	kreq := kongServiceRequestConvert(svc)
	// 1. create specified service information
	started := time.Now()
	ksvc, err := r.kClient.Services().Create(kreq)
	if err != nil {
		reportKongAPIMetrics("CreateServices", http.MethodPost, utils.ErrStatus, started)
		blog.Errorf("kong register create Service %s[%s] failed, %s", svc.Name, svc.Host, err.Error())
		return err
	}
	reportKongAPIMetrics("CreateServices", http.MethodPost, utils.SucStatus, started)

	// create service plugins
	if svc.Plugin != nil {
		pReqs := kongPluginConvert(svc.Plugin, ksvc.Id, "service")
		for _, pluReq := range pReqs {
			startedCreate := time.Now()
			var splugin *gokong.Plugin
			splugin, err = r.kClient.Plugins().Create(pluReq)
			if err != nil {
				reportKongAPIMetrics("CreatePlugins", http.MethodPost, utils.ErrStatus, startedCreate)
				//todo(DeveloperJim): shall we clean created service, that we can retry in next data synchronization
				blog.Errorf("kong register create plugin %s for Service %s failed, %s", pluReq.Name, svc.Name, err.Error())
				return err
			}
			reportKongAPIMetrics("CreatePlugins", http.MethodPost, utils.SucStatus, startedCreate)
			blog.Infof("kong register create plugin for service %s successfully, plugin ID: %s/%s", svc.Name, splugin.Id, splugin.Name)
		}
	}
	// 2. create service relative route rules
	for _, route := range svc.Routes {
		kr := kongRouteConvert(&route, ksvc.Id)
		startedRoute := time.Now()
		var kroute *gokong.Route
		kroute, err = r.kClient.Routes().Create(kr)
		if err != nil {
			reportKongAPIMetrics("CreateRoutes", http.MethodPost, utils.ErrStatus, startedRoute)
			blog.Errorf("kong register create route for Service %s failed, %s", svc.Name, err.Error())
			return err
		}
		reportKongAPIMetrics("CreateRoutes", http.MethodPost, utils.SucStatus, startedRoute)

		if route.Plugin != nil {
			rReqs := kongPluginConvert(route.Plugin, kroute.Id, "route")
			for _, pluReq := range rReqs {
				startedPluginsCreate := time.Now()
				var rplugin *gokong.Plugin
				rplugin, err = r.kClient.Plugins().Create(pluReq)
				if err != nil {
					reportKongAPIMetrics("CreatePlugins", http.MethodPost, utils.ErrStatus, startedPluginsCreate)
					//todo(DeveloperJim): roll back discussion
					blog.Errorf("kong register create plugin %s for route %s failed, %s", pluReq.Name, route.Name, err.Error())
					return err
				}
				reportKongAPIMetrics("CreatePlugins", http.MethodPost, utils.SucStatus, startedPluginsCreate)
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
	started = time.Now()
	kUpstream, err := r.kClient.Upstreams().Create(kupstrreq)
	if err != nil {
		reportKongAPIMetrics("CreateUpstreams", http.MethodPost, utils.ErrStatus, started)
		blog.Errorf("kong register create upstream %s for service %s failed, %s", svc.Host, svc.Name, err.Error())
		return err
	}
	reportKongAPIMetrics("CreateUpstreams", http.MethodPost, utils.SucStatus, started)
	blog.Infof("kong register create upstream %s [%s] successfully", kUpstream.Name, kUpstream.Id)
	//create targets for upstream
	for _, backend := range svc.Backends {
		targetReq := &gokong.TargetRequest{
			Target: backend.Target,
			Weight: backend.Weight,
		}
		startedCreateUpstream := time.Now()
		var ktarget *gokong.Target
		ktarget, err = r.kClient.Targets().CreateFromUpstreamName(kUpstream.Name, targetReq)
		if err != nil {
			reportKongAPIMetrics("CreateTargets", http.MethodPost, utils.ErrStatus, startedCreateUpstream)
			blog.Errorf("kong register create target %s for upstream %s failed, %s. try next one ", targetReq.Target, kUpstream.Name, err.Error())
			continue
		}
		reportKongAPIMetrics("CreateTargets", http.MethodPost, utils.SucStatus, startedCreateUpstream)
		blog.Infof("kong register create target %s[%s] for upstream %s successfully", targetReq.Target, *ktarget.Id, kUpstream.Name)
	}
	return nil
}

//UpdateService update specified Service, if service does not exist, return error
func (r *kRegister) UpdateService(svc *register.Service) error {
	return r.ReplaceTargetByService(svc, svc.Backends)
}

//GetService get specified service by name, if no service, return nil
func (r *kRegister) GetService(svc string) (*register.Service, error) {
	var (
		err     error
		started = time.Now()
		kSvc    *gokong.Service
	)
	defer reportRegisterKongMetrics("GetService", err, started)

	kSvc, err = r.kClient.Services().GetServiceByName(svc)
	if err != nil {
		reportKongAPIMetrics("GetServices", http.MethodGet, utils.ErrStatus, started)
		blog.Errorf("kong register get service %s failed, %s", svc, err.Error())
		return nil, err
	}
	if kSvc == nil {
		reportKongAPIMetrics("GetServices", http.MethodGet, utils.SucStatus, started)
		blog.Warnf("kong register get no Service named %s", svc)
		return nil, nil
	}
	//convert data structure
	registryService := innerServiceConvert(kSvc)
	reportKongAPIMetrics("GetServices", http.MethodGet, utils.SucStatus, started)
	return registryService, nil
}

//DeleteService delete specified service, success even if no such service
// @param service: at least setting Name & Host for deletion
func (r *kRegister) DeleteService(svc *register.Service) error {
	if svc.Host == "" || svc.Name == "" {
		return fmt.Errorf("service lost Name or Host")
	}
	var (
		err        error
		startedAll = time.Now()
	)
	defer reportRegisterKongMetrics("DeleteService", err, startedAll)

	started := time.Now()
	//clean route, route name is same with service
	if err = r.kClient.Routes().DeleteByName(svc.Name); err != nil {
		reportKongAPIMetrics("DeleteRoutes", http.MethodDelete, utils.ErrStatus, started)
		blog.Errorf("kong register delete service %s relative route failed, %s", svc, err.Error())
		return err
	}
	reportKongAPIMetrics("DeleteRoutes", http.MethodDelete, utils.SucStatus, started)
	blog.V(3).Infof("kong register delete route %s success", svc.Name)

	started = time.Now()
	err = r.kClient.Services().DeleteServiceByName(svc.Name)
	if err != nil {
		reportKongAPIMetrics("DeleteServices", http.MethodDelete, utils.ErrStatus, started)
		blog.Errorf("kong register delete service by name %s failed, %s", svc, err.Error())
		return err
	}
	reportKongAPIMetrics("DeleteServices", http.MethodDelete, utils.ErrStatus, started)
	blog.V(3).Infof("kong register delete service %s success", svc.Name)

	started = time.Now()
	//* clean upstream
	if err = r.kClient.Upstreams().DeleteByName(svc.Host); err != nil {
		reportKongAPIMetrics("DeleteUpstreams", http.MethodDelete, utils.ErrStatus, started)
		blog.Errorf("kong register delete service %s relative Upstream %s failed, %s", svc.Name, svc.Host, err.Error())
		return err
	}
	reportKongAPIMetrics("DeleteUpstreams", http.MethodDelete, utils.SucStatus, started)
	blog.Infof("kong register delete service %s upstream %s success", svc.Name, svc.Host)
	return nil
}

//ListServices get all existence services
func (r *kRegister) ListServices() ([]*register.Service, error) {
	query := &gokong.ServiceQueryString{
		Size: 200,
	}

	var (
		err     error
		started = time.Now()
		kSvcs   []*gokong.Service
	)
	defer reportRegisterKongMetrics("ListServices", err, started)

	kSvcs, err = r.kClient.Services().GetServices(query)
	if err != nil {
		reportKongAPIMetrics("GetServices", http.MethodGet, utils.ErrStatus, started)
		blog.Errorf("kong register list all services failed, %s", err.Error())
		return nil, err
	}
	reportKongAPIMetrics("GetServices", http.MethodGet, utils.SucStatus, started)

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
	if svc == nil || len(svc.Host) == 0 {
		return nil, fmt.Errorf("necessary service info lost")
	}

	var (
		err      error
		started  = time.Now()
		kTargets []*gokong.Target
	)
	defer reportRegisterKongMetrics("GetTargetByService", err, started)

	kTargets, err = r.kClient.Targets().GetTargetsFromUpstreamId(svc.Host)
	if err != nil {
		reportKongAPIMetrics("GetTargets", http.MethodGet, utils.ErrStatus, started)
		blog.Errorf("kong register get targets by service %s failed, %s", svc.Host, err.Error())
		return nil, err
	}
	reportKongAPIMetrics("GetTargets", http.MethodGet, utils.SucStatus, started)

	var backends []register.Backend
	for _, target := range kTargets {
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
func (r *kRegister) ReplaceTargetByService(svc *register.Service, backends []register.Backend) error {
	//get original targets
	if svc.Name == "" || svc.Host == "" {
		return fmt.Errorf("service info lost Name or Host")
	}
	if len(backends) == 0 {
		return fmt.Errorf("lost backends list")
	}
	var (
		startedAll = time.Now()
		err        error
	)
	defer reportRegisterKongMetrics("ReplaceTargetByService", err, startedAll)

	started := time.Now()
	targets, err := r.kClient.Targets().GetTargetsFromUpstreamId(svc.Host)
	if err != nil {
		reportKongAPIMetrics("GetTargets", http.MethodGet, utils.ErrStatus, started)
		blog.Errorf("kong register get upstream %s targets failed, %s", svc.Host, err.Error())
		return err
	}
	reportKongAPIMetrics("GetTargets", http.MethodGet, utils.SucStatus, started)

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
			started := time.Now()
			var ktarget *gokong.Target
			ktarget, err = r.kClient.Targets().CreateFromUpstreamName(svc.Host, v)
			if err != nil {
				reportKongAPIMetrics("CreateTargets", http.MethodPost, utils.ErrStatus, started)
				blog.Errorf("kong add New target %s for upstream %s failed, %s", k, svc.Host, err.Error())
				continue
			}
			reportKongAPIMetrics("CreateTargets", http.MethodPost, utils.SucStatus, started)
			blog.Infof("kong add new target %s[%d] for upstream %s success", k, ktarget.Id, svc.Host)
		}
	}
	if len(cleanTargets) != 0 {
		for k, v := range cleanTargets {
			started := time.Now()
			if err = r.kClient.Targets().DeleteFromUpstreamById(svc.Host, *v.Id); err != nil {
				reportKongAPIMetrics("DeleteTargets", http.MethodDelete, utils.ErrStatus, started)
				blog.Errorf("kong clean out-of-dated target %s for upstream %s failed, %s", k, svc.Host, err.Error())
				continue
			}
			reportKongAPIMetrics("DeleteTargets", http.MethodDelete, utils.SucStatus, started)
			blog.Infof("kong clean out-of-dated target %s[%d] for upstream %s success", k, v.Id, svc.Host)
		}
	}
	return nil
}

//DeleteTargetByService clean all backend list for service
func (r *kRegister) DeleteTargetByService(svc *register.Service) error {
	return fmt.Errorf("not implemented")
}

func (r *kRegister) deletePlugins(resource string, plugins []*gokong.Plugin) error {
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
	}
	//path will be empty when rewrite feature turns off
	if ksvc.Path != nil {
		svc.Path = *ksvc.Path
	}
	return svc
}

//kongServiceConvert convert inner service to kong service
func kongServiceRequestConvert(svc *register.Service) *gokong.ServiceRequest {
	ksvc := &gokong.ServiceRequest{
		Name:     &svc.Name,
		Protocol: &svc.Protocol,
		Host:     &svc.Host,
		Retries:  gokong.Int(svc.Retries),
	}
	if len(svc.Path) != 0 {
		ksvc.Path = &svc.Path
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
	//no matter what protocol it is, service only support tls
	//route supports double protocols
	if route.Protocol == protocolHTTP || route.Protocol == protocolHTTPS {
		protocols = []*string{gokong.String(protocolHTTP), gokong.String(protocolHTTPS)}
	} else if route.Protocol == protocolGRPC || route.Protocol == protocolGRPCS {
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

func reportKongAPIMetrics(handler, method, status string, started time.Time) {
	metricData := utils.APIMetricsMeta{
		System:  kongAdmin,
		Handler: handler,
		Method:  method,
		Status:  status,
		Started: started,
	}
	utils.ReportBcsGatewayAPIMetrics(metricData)
}

func reportRegisterKongMetrics(handler string, err error, started time.Time) {
	metricData := utils.APIMetricsMeta{
		System:  kongAdmin,
		Handler: handler,
		Status:  utils.SucStatus,
		Started: started,
	}
	if err != nil {
		metricData.Status = utils.ErrStatus
	}
	utils.ReportBcsGatewayRegistryMetrics(metricData)
}
