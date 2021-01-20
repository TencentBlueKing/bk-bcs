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

//RegisterHostHandler init host url info
func RegisterHostHandler(httpSvr *HTTPService, logic *netservice.NetService) *HostHandler {
	handler := &HostHandler{
		netSvr: logic,
	}
	webSvr := new(restful.WebService)
	//add http handler
	webSvr.Path("/v1/host").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	webSvr.Route(webSvr.POST("").To(handler.Add))
	webSvr.Route(webSvr.DELETE("/{host}").To(handler.Delete))
	//update by host ip
	webSvr.Route(webSvr.PUT("/{host}").To(handler.Update))
	//get all host info
	webSvr.Route(webSvr.GET("").To(handler.List))
	//get host by ip
	webSvr.Route(webSvr.GET("/{host}").To(handler.ListByID))

	httpSvr.Register(webSvr)
	return handler
}

//HostHandler for http host handler
type HostHandler struct {
	netSvr *netservice.NetService
}

//Add add host to ip pool
func (host *HostHandler) Add(request *restful.Request, response *restful.Response) {
	started := time.Now()
	netReq := &types.NetRequest{}
	if err := request.ReadEntity(netReq); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		blog.Errorf("HostHandler [Add] json decode Err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("addHostToIPPool", "4xx", started)
		return
	}
	netRes := &types.NetResponse{
		Type: types.RequestType_HOST,
	}
	if netReq.Type != types.RequestType_HOST || netReq.Host == nil {
		netRes.Code = 1
		netRes.Message = "Request type err or Host info lost"
		blog.Errorf("HostHandler POST check request type/info, but got unexpect type %d", netReq.Type)
		response.WriteEntity(netRes)
		reportMetrics("addHostToIPPool", "4xx", started)
		return
	}
	//check host data
	if !netReq.Host.IsValid() {
		netRes.Code = 1
		netRes.Message = "Request host data lost"
		blog.Errorf("HostHandler Post check request data lost! host %s, pool %s", netReq.Host.IPAddr, netReq.Host.Pool)
		response.WriteEntity(netRes)
		reportMetrics("addHostToIPPool", "4xx", started)
		return
	}
	if err := host.netSvr.AddHost(netReq.Host); err != nil {
		netRes.Code = 1
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("addHostToIPPool", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = SUCCESS
	response.WriteEntity(netRes)
	blog.Infof("HostHandler Post %s success.", netReq.Host.IPAddr)
	reportMetrics("addHostToIPPool", "2xx", started)
}

//Delete delete specified by host IPaddress, also clean IPAddress assign to this host
func (host *HostHandler) Delete(request *restful.Request, response *restful.Response) {
	started := time.Now()
	hostIP := request.PathParameter("host")
	netReq := &types.NetRequest{}
	if err := request.ReadEntity(netReq); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		blog.Errorf("HostHandler [Delete] json decode Err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("deleteHostFromIPPool", "4xx", started)
		return
	}
	netRes := &types.NetResponse{
		Type: types.ResponseType_HOST,
	}
	if err := host.netSvr.DeleteHost(hostIP, netReq.IPs); err != nil {
		netRes.Code = 1
		netRes.Message = err.Error()
		blog.Errorf("HostHandler DELETE host %s failed, %s", hostIP, err)
		response.WriteEntity(netRes)
		reportMetrics("deleteHostFromIPPool", "5xx", started)
		return
	}
	blog.Infof("HostHandler Delete host %s success.", hostIP)
	netRes.Code = 0
	netRes.Message = SUCCESS
	response.WriteEntity(netRes)
	reportMetrics("deleteHostFromIPPool", "2xx", started)
}

//Update update pool by ip segment
func (host *HostHandler) Update(request *restful.Request, response *restful.Response) {
	blog.Warn("#######HostHandler [Update] Not implemented#######")
	response.AddHeader("Content-Type", "text/plain")
	response.WriteErrorString(http.StatusForbidden, "Not implemented")
}

//List list all pools
func (host *HostHandler) List(request *restful.Request, response *restful.Response) {
	started := time.Now()
	//list all hosts
	netRes := &types.NetResponse{
		Type: types.ResponseType_HOST,
	}
	allHosts, err := host.netSvr.ListHost()
	if err != nil {
		blog.Errorf("HostHandler List all request err: %s", err.Error())
		netRes.Code = 1
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("listHostFromIPPool", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = SUCCESS
	netRes.Host = allHosts
	netRes.Data = netRes.Host
	blog.Infof("Hosthandler client %s Get all host success", request.Request.RemoteAddr)
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("HostHandler reply client GET request Err: %v", err)
	}
	reportMetrics("listHostFromIPPool", "2xx", started)
}

//ListByID list host by ip
func (host *HostHandler) ListByID(request *restful.Request, response *restful.Response) {
	started := time.Now()
	//list host by host ip
	ip := request.PathParameter("host")
	netRes := &types.NetResponse{
		Type: types.ResponseType_HOST,
	}
	hostInfo, err := host.netSvr.ListHostByKey(ip)
	if err != nil {
		blog.Errorf("HostHandler list host %s request err: %s", ip, err.Error())
		netRes.Code = 1
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("listHostInfoFromIPPool", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = SUCCESS
	netRes.Host = append(netRes.Host, hostInfo)
	netRes.Data = netRes.Host
	blog.Infof("Hosthandler client %s Get host %s success", request.Request.RemoteAddr, ip)
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("HostHandler reply client GET/%s request Err: %v", ip, err)
	}
	reportMetrics("listHostInfoFromIPPool", "2xx", started)
}
