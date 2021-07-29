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

package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/utils"
)

//NewClient create apisix admin api client
func NewClient(option *Option) Client {
	c := &client{
		option: option,
	}
	if option.TLSConfig != nil {
		c.client = restclient.NewRESTClientWithTLS(option.TLSConfig)
	} else {
		c.client = restclient.NewRESTClient()
	}
	return c
}

func apisixSetting(req *restclient.Request, option *Option) *restclient.Request {
	header := make(http.Header)
	if len(option.AdminToken) != 0 {
		header.Add("X-API-KEY", option.AdminToken)
	}
	if len(header) != 0 {
		return req.WithHeaders(header)
	}
	return req
}

// apisix admin api client implementation
type client struct {
	option *Option
	client *restclient.RESTClient
}

// GetUpstream implementation
func (c *client) GetUpstream(id string) (*Upstream, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("upstream id required")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "GetUpstream",
		Method:  http.MethodGet,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()
	var response Basic
	err := apisixSetting(c.client.Get(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		SubPathf("/apisix/admin/upstreams/%s", id).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return nil, err
	}
	if response.Data == nil || response.Data.Value == nil {
		metricData.Status = utils.SucStatus
		// no exact data
		return nil, nil
	}
	upstream := new(Upstream)
	if err := json.Unmarshal(response.Data.Value, upstream); err != nil {
		return nil, fmt.Errorf("upstream decode err: %s", err.Error())
	}
	if len(upstream.ID) == 0 {
		return nil, fmt.Errorf("upstream data err")
	}
	return upstream, nil
}

// GetUpstream implementation
func (c *client) CreateUpstream(upstr *Upstream) error {
	if upstr == nil || len(upstr.Nodes) == 0 {
		return fmt.Errorf("upstream nodes is empty")
	}
	if len(upstr.ID) == 0 {
		return fmt.Errorf("upstream ID required")
	}
	if !(upstr.Type == BalanceTypeRoundrobin || upstr.Type == BalanceTypeChash) {
		return fmt.Errorf(
			"upstream type err, only [%s, %s] are available",
			BalanceTypeRoundrobin,
			BalanceTypeChash)
	}
	if upstr.Retries == 0 {
		upstr.Retries = 1
	}
	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "CreateUpstream",
		Method:  http.MethodPost,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()
	var response Basic
	err := apisixSetting(c.client.Put(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/apisix/admin/upstreams/" + upstr.ID).
		WithJSON(upstr).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		// no exact data
		return fmt.Errorf(response.Err)
	}
	return nil
}

// ListUpstream implementation
func (c *client) ListUpstream() ([]*Upstream, error) {
	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "CreateUpstream",
		Method:  http.MethodPost,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Get(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/apisix/admin/upstreams").
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return nil, err
	}
	if response.Count == "1" || response.Data == nil ||
		!response.Data.Directory || response.Data.Nodes == nil {
		// no exact data
		return nil, nil
	}
	//Unmarshal response.Data.Nodes to slice
	var dataNodes []*Node
	if err := json.Unmarshal(response.Data.Nodes, &dataNodes); err != nil {
		return nil, fmt.Errorf("Nodes data is not slice")
	}
	var ups []*Upstream
	for _, node := range dataNodes {
		upstream := new(Upstream)
		if err := json.Unmarshal(node.Value, upstream); err != nil {
			return nil, fmt.Errorf("upstream decode err: %s", err.Error())
		}
		if len(upstream.ID) == 0 {
			return nil, fmt.Errorf("upstream data err")
		}
		ups = append(ups, upstream)
	}
	return ups, nil
}

// GetUpstream implementation
func (c *client) UpdateUpstream(upstr *Upstream) error {
	if upstr == nil || len(upstr.Nodes) == 0 {
		return fmt.Errorf("upstream nodes is empty")
	}
	if !(upstr.Type == BalanceTypeRoundrobin || upstr.Type == BalanceTypeChash) {
		return fmt.Errorf(
			"upstream type err, only [%s, %s] are available",
			BalanceTypeRoundrobin,
			BalanceTypeChash)
	}
	if upstr.Retries == 0 {
		upstr.Retries = 1
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "UpdateUpstream",
		Method:  http.MethodPut,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Put(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/apisix/admin/upstreams/" + upstr.ID).
		WithJSON(upstr).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		// some logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}

// GetUpstream implementation
func (c *client) DeleteUpstream(id string) error {
	if len(id) == 0 {
		return fmt.Errorf("upstream id required")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "DeleteUpstream",
		Method:  http.MethodDelete,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()
	var response Basic
	err := apisixSetting(c.client.Delete(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		SubPathf("/apisix/admin/upstreams/%s", id).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		//logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}

// GetService implementation
func (c *client) GetService(id string) (*Service, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("service id required")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "GetService",
		Method:  http.MethodGet,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Get(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		SubPathf("/apisix/admin/services/%s", id).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return nil, err
	}
	if response.Data == nil || response.Data.Value == nil {
		// no exact data
		return nil, nil
	}
	service := new(Service)
	if err := json.Unmarshal(response.Data.Value, service); err != nil {
		return nil, fmt.Errorf("service decode err: %s", err.Error())
	}
	if len(service.ID) == 0 {
		return nil, fmt.Errorf("service data err")
	}
	return service, nil
}

// ListService implementation
func (c *client) ListService() ([]*Service, error) {
	var response Basic

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "ListService",
		Method:  http.MethodGet,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	err := apisixSetting(c.client.Get(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/apisix/admin/services").
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return nil, err
	}
	if response.Count == "1" || response.Data == nil ||
		!response.Data.Directory || response.Data.Nodes == nil {
		// no exact data
		return nil, nil
	}
	//Unmarshal response.Data.Nodes to slice
	var dataNodes []*Node
	if err := json.Unmarshal(response.Data.Nodes, &dataNodes); err != nil {
		return nil, fmt.Errorf("Nodes data is not slice")
	}
	var svcs []*Service
	for _, node := range dataNodes {
		service := new(Service)
		if err := json.Unmarshal(node.Value, service); err != nil {
			return nil, fmt.Errorf("service decode err: %s", err.Error())
		}
		if len(service.ID) == 0 {
			return nil, fmt.Errorf("service data err")
		}
		svcs = append(svcs, service)
	}
	return svcs, nil
}

// CreateService implementation
func (c *client) CreateService(svc *Service) error {
	if svc == nil || len(svc.ID) == 0 {
		return fmt.Errorf("service is empty")
	}
	if svc.Upstream == nil && len(svc.UpstreamID) == 0 {
		return fmt.Errorf("service lost upstream information")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "CreateService",
		Method:  http.MethodPost,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Put(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/apisix/admin/services/" + svc.ID).
		WithJSON(svc).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		// some logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}

// UpdateService implementation
func (c *client) UpdateService(svc *Service) error {
	if svc == nil || len(svc.ID) == 0 {
		return fmt.Errorf("service is empty")
	}
	if svc.Upstream == nil && len(svc.UpstreamID) == 0 {
		return fmt.Errorf("service lost upstream information")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "UpdateService",
		Method:  http.MethodPut,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Put(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/apisix/admin/services/" + svc.ID).
		WithJSON(svc).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		// some logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}

// DeleteService implementation
func (c *client) DeleteService(id string) error {
	if len(id) == 0 {
		return fmt.Errorf("service id required")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "DeleteService",
		Method:  http.MethodDelete,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Delete(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		SubPathf("/apisix/admin/services/%s", id).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		//logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}

// GetRoute implementation
func (c *client) GetRoute(id string) (*Route, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("route id required")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "GetRoute",
		Method:  http.MethodGet,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Get(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		SubPathf("/apisix/admin/routes/%s", id).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return nil, err
	}
	if response.Data == nil || response.Data.Value == nil {
		// no exact data
		return nil, nil
	}
	route := new(Route)
	if err := json.Unmarshal(response.Data.Value, route); err != nil {
		return nil, fmt.Errorf("route decode err: %s", err.Error())
	}
	if len(route.ID) == 0 {
		return nil, fmt.Errorf("route data err")
	}
	return route, nil
}

// ListRoute implementation
func (c *client) ListRoute() ([]*Route, error) {
	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "ListRoute",
		Method:  http.MethodGet,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Get(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		WithBasePath("/apisix/admin/routes").
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return nil, err
	}
	if response.Count == "1" || response.Data == nil ||
		!response.Data.Directory || response.Data.Nodes == nil {
		// no exact data
		return nil, nil
	}
	//Unmarshal response.Data.Nodes to slice
	var dataNodes []*Node
	if err := json.Unmarshal(response.Data.Nodes, &dataNodes); err != nil {
		return nil, fmt.Errorf("Nodes data is not slice")
	}
	var routes []*Route
	for _, node := range dataNodes {
		route := new(Route)
		if err := json.Unmarshal(node.Value, route); err != nil {
			return nil, fmt.Errorf("route decode err: %s", err.Error())
		}
		if len(route.ID) == 0 {
			return nil, fmt.Errorf("route data err")
		}
		routes = append(routes, route)
	}
	return routes, nil
}

// CreateRoute implementation
func (c *client) CreateRoute(route *Route) error {
	if route == nil || len(route.ID) == 0 {
		return fmt.Errorf("route is empty")
	}
	if route.Upstream == nil && len(route.UpstreamID) == 0 &&
		route.Service == nil && len(route.ServiceID) == 0 {
		return fmt.Errorf("route lost service/upstream information")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "CreateRoute",
		Method:  http.MethodPost,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Put(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		WithBasePath("/apisix/admin/routes/" + route.ID).
		WithJSON(route).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		// some logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}

// UpdateRoute implementation
func (c *client) UpdateRoute(route *Route) error {
	if route == nil || len(route.ID) == 0 {
		return fmt.Errorf("route is empty")
	}
	if route.Upstream == nil && len(route.UpstreamID) == 0 &&
		route.Service == nil && len(route.ServiceID) == 0 {
		return fmt.Errorf("route lost service/upstream information")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "UpdateRoute",
		Method:  http.MethodPut,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Put(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		WithBasePath("/apisix/admin/routes/" + route.ID).
		WithJSON(route).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		// some logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}

// DeleteRoute implementation
func (c *client) DeleteRoute(id string) error {
	if len(id) == 0 {
		return fmt.Errorf("route id required")
	}

	metricData := utils.APIMetricsMeta{
		System:  ApisixAdmin,
		Handler: "DeleteRoute",
		Method:  http.MethodDelete,
		Status:  utils.SucStatus,
		Started: time.Now(),
	}
	defer func() {
		utils.ReportBcsGatewayAPIMetrics(metricData)
	}()

	var response Basic
	err := apisixSetting(c.client.Delete(), c.option).
		WithEndpoints(c.option.Addrs).
		WithBasePath("/").
		SubPathf("/apisix/admin/routes/%s", id).
		Do().
		Into(&response)
	if err != nil {
		metricData.Status = utils.ErrStatus
		return err
	}
	if len(response.Err) != 0 {
		metricData.Status = utils.ErrStatus
		//logic error
		return fmt.Errorf(response.Err)
	}
	return nil
}
