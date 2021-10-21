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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/register/apisix/admin"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/utils"
)

//New create Register implementation for apisix
// return empty
func New(addr []string, config *tls.Config, token string) (register.Register, error) {
	opt := &admin.Option{
		AdminToken: token,
		Addrs:      addr,
		//TLSConfig:  config,
	}
	blog.Infof("apisix config details: %+v", opt)
	reg := &apiRegister{
		apisixClient: admin.NewClient(opt),
	}
	return reg, nil
}

//apiRegister apisix register implementation
type apiRegister struct {
	apisixClient admin.Client
}

//CreateService create Service interface, if service already exists, return error
// create service include three operations:
// 1. create specified service information, including plugins
// 2. create service relative route rules, including plugins
// 3. create service relative Upstream
// proxy rules from bcs-api-gateway:
// we authenticate in stage of route, then post to backend service when authentication success.
// in stage of proxy, we clean original Authorization information and switch to inner
// authentication token for different bkbcs modules
func (r *apiRegister) CreateService(svc *register.Service) error {
	var (
		started = time.Now()
		err     error
	)
	defer reportRegisterAPISixMetrics("CreateService", err, started)

	if err = svc.Valid(); err != nil {
		blog.Errorf("service %s is invalid, %s", svc.Name, err.Error())
		return err
	}

	//create specified upstream information
	upstream := apisixUpstreamConversion(svc)
	if err = r.apisixClient.CreateUpstream(upstream); err != nil {
		blog.Errorf("apisix register create service %s Upstream failed, %s. upstream details: %+v",
			svc.Name, err.Error(), upstream)
		return err
	}
	//create specified service information
	service := apisixServiceConversion(svc)
	if err = r.apisixClient.CreateService(service); err != nil {
		blog.Errorf("apisix register create Service %s failed, %s. service details: %+v",
			svc.Name, err.Error(), service)
		//create service failed, ready to clean dirty upstream data
		if streamErr := r.apisixClient.DeleteUpstream(upstream.ID); streamErr != nil {
			blog.Errorf("apisix register clean service %s dirty Upstream data failed, %s", svc.Name, streamErr.Error())
		}
		return err
	}
	// 2. create service relative route rules
	var routes []*admin.Route
	failed := false
	for _, innerroute := range svc.Routes {
		route := apisixRouteConversion(svc, &innerroute)
		if err = r.apisixClient.CreateRoute(route); err != nil {
			blog.Errorf("apisix register create service %s route failed, %s. route details: %+v",
				svc.Name, err.Error(), route)
			failed = true
			break
		}
		routes = append(routes, route)
	}
	if failed {
		//clean relative dirty data
		for _, route := range routes {
			if err := r.apisixClient.DeleteRoute(route.ID); err != nil {
				blog.Errorf("apisix clean service %s dirty route failed, %s", svc.Name, err.Error())
			}
		}
		//create service failed, ready to clean dirty upstream data
		if err := r.apisixClient.DeleteService(service.ID); err != nil {
			blog.Errorf("apisix register clean dirty service %s data failed, %s", service.ID, err.Error())
		}
		//create service failed, ready to clean dirty upstream data
		if err := r.apisixClient.DeleteUpstream(upstream.ID); err != nil {
			blog.Errorf("apisix register clean service %s dirty Upstream data failed, %s", service.ID, err.Error())
		}
		return err
	}
	return nil
}

//UpdateService update specified Service, if service does not exist, return error
func (r *apiRegister) UpdateService(svc *register.Service) error {
	return fmt.Errorf("Not Implemented")
}

//GetService get specified service by name, if no service, return nil
func (r *apiRegister) GetService(svc string) (*register.Service, error) {
	var (
		started = time.Now()
		err     error
	)
	defer reportRegisterAPISixMetrics("GetService", err, started)

	var service *admin.Service
	service, err = r.apisixClient.GetService(svc)
	if err != nil {
		blog.Errorf("apisix register get service %s failed, %s", svc, err.Error())
		return nil, err
	}
	if service == nil {
		blog.Warnf("apisix register get no Service named %s", svc)
		return nil, nil
	}

	var upstream *admin.Upstream
	upstream, err = r.apisixClient.GetUpstream(svc)
	if err != nil {
		blog.Errorf("apisix register get service %s relative upstream failed, %s", svc, err.Error())
		return nil, err
	}
	if upstream == nil {
		blog.Errorf("apisix register get service %s err, Upsteram Not Found", svc)
		return nil, fmt.Errorf("Upstream Not Found")
	}

	var route *admin.Route
	route, err = r.apisixClient.GetRoute(svc)
	if err != nil {
		blog.Errorf("apisix register get service %s relative route failed, %s", svc, err.Error())
		return nil, err
	}
	if route == nil {
		blog.Errorf("apisix register get service %s err, Route Not Found", svc)
		return nil, fmt.Errorf("Route Not Found")
	}
	//convert data structure
	return innerServiceConvert(service, route, upstream), nil
}

//DeleteService delete specified service, success even if no such service
// @param service: at least setting Name & Host for deletion
func (r *apiRegister) DeleteService(svc *register.Service) error {
	return fmt.Errorf("Not Implemented")
}

//ListServices get all existence services
func (r *apiRegister) ListServices() ([]*register.Service, error) {
	var (
		started = time.Now()
		err     error
	)
	defer reportRegisterAPISixMetrics("ListServices", err, started)

	var allServices []*admin.Service
	allServices, err = r.apisixClient.ListService()
	if err != nil {
		blog.Errorf("apisix register list all service failed, %s", err.Error())
		return nil, err
	}
	if len(allServices) == 0 {
		return nil, nil
	}
	var services []*register.Service
	for _, service := range allServices {
		svc := simpleInnerServiceConversion(service)
		services = append(services, svc)
	}
	return services, nil
}

//GetTargetByService get service relative backends
func (r *apiRegister) GetTargetByService(svc *register.Service) ([]register.Backend, error) {
	if svc == nil || len(svc.Name) == 0 {
		return nil, fmt.Errorf("necessary service info lost")
	}

	var (
		err      error
		started  = time.Now()
		upstream *admin.Upstream
	)
	defer reportRegisterAPISixMetrics("GetTargetByService", err, started)

	upstream, err = r.apisixClient.GetUpstream(svc.Name)
	if err != nil {
		blog.Errorf("apisix register get targets by service %s failed, %s", svc.Name, err.Error())
		return nil, err
	}
	if upstream == nil {
		blog.Errorf("apisix register GetTargetByService %s err, Upsteram Not Found", svc)
		return nil, fmt.Errorf("Upstream Not Found")
	}

	var backends []register.Backend
	for target, weight := range upstream.Nodes {
		backend := register.Backend{
			Target: target,
			Weight: weight,
		}
		backends = append(backends, backend)
	}
	return backends, nil
}

//ReplaceTargetByService replace specified service backend list
// so we don't care what original backend list are
func (r *apiRegister) ReplaceTargetByService(svc *register.Service, backends []register.Backend) error {
	//get original targets
	if svc.Name == "" {
		return fmt.Errorf("service info lost Name or Host")
	}
	if len(backends) == 0 {
		return fmt.Errorf("lost backends list")
	}

	var (
		err     error
		started = time.Now()
	)
	defer reportRegisterAPISixMetrics("ReplaceTargetByService", err, started)

	var upstream *admin.Upstream
	upstream, err = r.apisixClient.GetUpstream(svc.Name)
	if err != nil {
		blog.Errorf("apisix register get upstream %s targets failed, %s", svc.Name, err.Error())
		return err
	}
	if upstream == nil {
		blog.Errorf("apisix register logic error, service %s lost upstream.", svc.Name)
		return fmt.Errorf("service Lost upstream")
	}
	destBackends := make(map[string]int)
	newBackends := make(map[string]int)
	for _, backend := range backends {
		destBackends[backend.Target] = backend.Weight
		oldWeight, ok := upstream.Nodes[backend.Target]
		if ok && oldWeight == backend.Weight {
			delete(upstream.Nodes, backend.Target)
			continue
		}
		newBackends[backend.Target] = backend.Weight
	}
	if len(upstream.Nodes) == 0 && len(newBackends) == 0 {
		blog.Infof("service %s upstream no changed", svc.Name)
		return nil
	}
	blog.Infof("apisix register service %s operation: delete node %+v, add node %+v", svc.Name, upstream.Nodes, newBackends)
	upstream.Nodes = destBackends
	if err = r.apisixClient.UpdateUpstream(upstream); err != nil {
		blog.Errorf("apisix register update stream %+v, failed, %s", upstream, err.Error())
		return err
	}
	return nil
}

//DeleteTargetByService clean all backend list for service
func (r *apiRegister) DeleteTargetByService(svc *register.Service) error {
	return fmt.Errorf("Not Implemented")
}

func simpleInnerServiceConversion(svc *admin.Service) *register.Service {
	innerService := &register.Service{
		Name:      svc.ID,
		Protocol:  "https",
		Retries:   1,
		Algorithm: admin.BalanceTypeRoundrobin,
	}
	return innerService
}

//innerServiceConvert convert apisix service/route/upstream to inner service definition
func innerServiceConvert(svc *admin.Service, route *admin.Route, upstream *admin.Upstream) *register.Service {
	innerService := &register.Service{
		Name:      svc.ID,
		Protocol:  route.ServiceProtocol,
		Retries:   upstream.Retries,
		Algorithm: upstream.Type,
	}
	//complicated conversion begin
	for target, weight := range upstream.Nodes {
		backend := register.Backend{
			Target: target,
			Weight: weight,
		}
		innerService.Backends = append(innerService.Backends, backend)
	}

	return innerService
}

// apisixUpstreamConversion convert to apisix upstream information
func apisixUpstreamConversion(svc *register.Service) *admin.Upstream {
	up := &admin.Upstream{
		ID:      svc.Name,
		Name:    svc.Name,
		Type:    admin.BalanceTypeRoundrobin,
		Retries: svc.Retries,
		Nodes:   make(map[string]int),
	}
	for _, backend := range svc.Backends {
		up.Nodes[backend.Target] = backend.Weight
	}
	return up
}

//apisixServiceConversion convert inner service to kong service
func apisixServiceConversion(svc *register.Service) *admin.Service {
	service := &admin.Service{
		ID:         svc.Name,
		UpstreamID: svc.Name,
		Websocket:  true,
		Plugins:    make(map[string]interface{}),
	}
	name, plugin := apisixLimitRequestPlugin()
	service.Plugins[name] = plugin
	return service
}

//apisixRouteConversion convert inner service to apisix Route, tls feature supported in default.
func apisixRouteConversion(svc *register.Service, route *register.Route) *admin.Route {
	r := &admin.Route{
		ID:              route.Name,
		Name:            route.Name,
		Websocket:       true,
		ServiceProtocol: "http",
		ServiceID:       svc.Name,
		Plugins:         make(map[string]interface{}),
	}
	if route.Protocol == "grpc" {
		r.ServiceProtocol = route.Protocol
	}
	if route.Plugin != nil && route.Plugin.AuthOption != nil {
		bcsAuth, authPlugin := apisixBKBCSAuthConversion(route.Plugin.AuthOption)
		r.Plugins[bcsAuth] = authPlugin
	}
	reqID, reqPlugin := apisixRequestIDPlugin()
	r.Plugins[reqID] = reqPlugin
	//setting route path, end with * means wildcard
	r.URI = route.Paths[0] + "*"
	proxyPlugin := make(map[string]interface{})
	r.Plugins["proxy-rewrite"] = proxyPlugin
	proxyPlugin["scheme"] = register.ProtocolHTTPS
	proxyPlugin["host"] = svc.Host
	if route.PathRewrite {
		var regexURI []string
		regexURI = append(regexURI, route.Paths[0]+"(.*)")
		regexURI = append(regexURI, svc.Path+"$1")
		proxyPlugin["regex_uri"] = regexURI
	}
	if svc.Plugin != nil && svc.Plugin.HeadOption != nil {
		//setting header authorization
		header := make(map[string]string)
		for key, value := range svc.Plugin.HeadOption.Add {
			header[key] = value
		}
		proxyPlugin["headers"] = header
	}
	//header filter
	for key, value := range route.Header {
		var filter []string
		filter = append(filter, "http_"+key)
		filter = append(filter, "==")
		filter = append(filter, value)
		r.Vars = append(r.Vars, filter)
	}
	return r
}

func apisixLimitRequestPlugin() (string, map[string]interface{}) {
	plgn := make(map[string]interface{})
	plgn["conn"] = 1000
	plgn["burst"] = 500
	plgn["rejected_code"] = 429
	plgn["key"] = "remote_addr"
	plgn["default_conn_delay"] = 0.1
	return "limit-conn", plgn
}

func apisixRequestIDPlugin() (string, map[string]interface{}) {
	plgn := make(map[string]interface{})
	plgn["include_in_response"] = true
	return "request-id", plgn
}

//apisixBKBCSAuthConvert convert inner service request plugin to request-transformer
func apisixBKBCSAuthConversion(option *register.BCSAuthOption) (string, map[string]interface{}) {
	auth := make(map[string]interface{})
	auth["token"] = option.AuthToken
	auth["bkbcs_auth_endpoints"] = option.AuthEndpoints
	auth["module"] = option.Module
	return option.Name, auth
}

func reportRegisterAPISixMetrics(handler string, err error, started time.Time) {
	metricData := utils.APIMetricsMeta{
		System:  admin.ApisixAdmin,
		Handler: handler,
		Status:  utils.SucStatus,
		Started: started,
	}
	if err != nil {
		metricData.Status = utils.ErrStatus
	}
	utils.ReportBcsGatewayRegistryMetrics(metricData)
}
