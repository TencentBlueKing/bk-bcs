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

package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/netservice"
	"net/http"
	"time"

	restful "github.com/emicklei/go-restful"
)

//RegisterResourceHandler init host url info
func RegisterResourceHandler(httpSvr *HTTPService, logic *netservice.NetService) *ResourceHandler {
	handler := &ResourceHandler{
		netSvr: logic,
	}
	webSvr := new(restful.WebService)
	//add http handler
	webSvr.Path("/v1/resource").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	//get all host info
	webSvr.Route(webSvr.POST("").To(handler.List))
	httpSvr.Register(webSvr)
	return handler
}

//ResourceHandler for http host handler
type ResourceHandler struct {
	netSvr *netservice.NetService
}

//List list all pools
func (resource *ResourceHandler) List(request *restful.Request, response *restful.Response) {
	started := time.Now()
	//request json decode
	req := &types.ResourceRequest{}
	if err := request.ReadEntity(req); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		blog.Errorf("ResourceHandler [List] json decode Err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("hostResource", "4xx", started)
		return
	}
	res := &types.ResourceResponse{
		HostResource: make(map[string]int),
	}
	if !req.IsValid() {
		res.Code = 1
		res.Message = "resource request is invalid"
		response.WriteEntity(res)
		reportMetrics("hostResource", "4xx", started)
		return
	}
	for _, host := range req.Hosts {
		//default num for error
		res.HostResource[host] = -1
	}
	//host: cluster/pool
	hostMap := make(map[string]string)
	clusterMap := make(map[string]int)
	for k := range res.HostResource {
		//get HostInfo without container info
		host, err := resource.netSvr.GetHostInfo(k)
		if err != nil {
			blog.Errorf("Get Host %s Node info failed in ResourceHandler, skip", k)
			reportMetrics("hostResource", "5xx", started)
			continue
		}
		if host == nil {
			continue
		}
		key := host.Cluster + "/" + host.Pool
		hostMap[k] = key
		clusterMap[key] = -1
	}
	//get cluster
	for k := range clusterMap {
		pool, err := resource.netSvr.GetPoolAvailable(k)
		if err != nil {
			blog.Errorf("Get PoolAvailable for Pool %s failed in ResourceHandler, skip", k)
			reportMetrics("hostResource", "5xx", started)
			continue
		}
		clusterMap[k] = len(pool.Available)
	}
	for k, v := range hostMap {
		res.HostResource[k] = clusterMap[v]
	}
	res.Code = 0
	res.Message = SUCCESS
	blog.Infof("Resourcehandler client %s Get IP resource under host success", request.Request.RemoteAddr)
	if err := response.WriteEntity(res); err != nil {
		blog.Errorf("Resourcehandler reply client %s under List request Err: %v", request.Request.RemoteAddr, err)
	}
	reportMetrics("hostResource", "2xx", started)
}
