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

//RegisterAllocator ip resource allocation
func RegisterAllocator(httpSvr *HTTPService, logic *netservice.NetService) *Allocator {
	handler := &Allocator{
		netSvr: logic,
	}
	webSvr := new(restful.WebService)
	//add http handler
	webSvr.Path("/v1/allocator").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	webSvr.Route(webSvr.POST("").To(handler.Add))
	webSvr.Route(webSvr.DELETE("").To(handler.Delete))
	webSvr.Route(webSvr.DELETE("/host/{hostip}").To(handler.HostVIPRelease))
	//list all allocation ip resource by net
	//webSvr.Route(webSvr.GET("/{ip}").To(handler.ListByID))
	httpSvr.Register(webSvr)
	return handler
}

//Allocator ip resource lean & release
type Allocator struct {
	netSvr *netservice.NetService
}

//Add iplease
func (allo *Allocator) Add(request *restful.Request, response *restful.Response) {
	started := time.Now()
	netReq := &types.NetRequest{}
	if err := request.ReadEntity(netReq); err != nil {
		blog.Errorf("Allocator decode json request falied, %s", err.Error())
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("iplease", "4xx", started)
		return
	}
	netRes := &types.NetResponse{
		Type: types.ResponseType_LEASE,
	}
	if netReq.Type != types.RequestType_LEASE || netReq.Lease == nil {
		blog.Errorf("Allocator Lost IPLease data in request")
		netRes.Code = 1
		netRes.Message = "lost ip lease data in restful request"
		response.WriteEntity(netRes)
		reportMetrics("iplease", "4xx", started)
		return
	}
	//check container id & host ip
	if netReq.Lease.Host == "" || netReq.Lease.Container == "" {
		blog.Errorf("Allocator lost Host/Container info in IPLease")
		netRes.Code = 1
		netRes.Message = "Host or Container info lost in IPLease"
		response.WriteEntity(netRes)
		reportMetrics("iplease", "4xx", started)
		return
	}
	netRes.Lease = netReq.Lease
	netRes.Data = netReq.Lease
	info, err := allo.netSvr.IPLean(netReq.Lease)
	if err != nil {
		blog.Errorf("Allocator lease ip for host %s to container %s failed, %v", netReq.Lease.Host, netReq.Lease.Container, err)
		netRes.Code = 2
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("iplease", "5xx", started)
		return
	}
	blog.Infof("Allocator lease ip [%s] for Host %s container %s success.", info.IPAddr, netRes.Lease.Host, netRes.Lease.Container)
	netRes.Info = append(netRes.Info, info)
	netRes.Code = 0
	netRes.Message = SUCCESS
	response.WriteEntity(netRes)
	reportMetrics("iplease", "2xx", started)
}

//Delete relesase ip address
func (allo *Allocator) Delete(request *restful.Request, response *restful.Response) {
	started := time.Now()
	netReq := &types.NetRequest{}
	if err := request.ReadEntity(netReq); err != nil {
		blog.Errorf("Allocator #DELETE# release ip address failed: %s", err.Error())
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("ipRelease", "4xx", started)
		return
	}
	//check data needed
	netRes := &types.NetResponse{
		Type: types.ResponseType_RELEASE,
	}
	if netReq.Type != types.RequestType_RELEASE || netReq.Release == nil {
		blog.Errorf("Allocator Release ip info failed, type or iprelease data lost")
		netRes.Code = 1
		netRes.Message = "request type or ip release data lost"
		response.WriteEntity(netRes)
		reportMetrics("ipRelease", "4xx", started)
		return
	}
	if netReq.Release.Host == "" || netReq.Release.Container == "" {
		blog.Errorf("Allocator lost host/container info, release ip address failed")
		netRes.Code = 2
		netRes.Message = "host/container info lost, release ip address failed"
		response.WriteEntity(netRes)
		reportMetrics("ipRelease", "4xx", started)
		return
	}
	if err := allo.netSvr.IPRelease(netReq.Release); err != nil {
		blog.Errorf("Allocator release container %s ip failed, %s", netReq.Release.Container, err.Error())
		netRes.Code = 2
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("ipRelease", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = SUCCESS
	netRes.Release = netReq.Release
	netRes.Data = netReq.Release
	response.WriteEntity(netRes)
	reportMetrics("ipRelease", "2xx", started)
}

//Update update pool by ip segment
func (allo *Allocator) Update(request *restful.Request, response *restful.Response) {
	blog.Warn("#######Allocator [Update] Not implemented#######")
	response.AddHeader("Content-Type", "text/plain")
	response.WriteErrorString(http.StatusForbidden, "Not implemented")
}

//List list all ip address under active
func (allo *Allocator) List(request *restful.Request, response *restful.Response) {
	//list all active ip address
	blog.Warn("#######Allocator [GET] Not implemented#######")
	response.AddHeader("Content-Type", "text/plain")
	response.WriteErrorString(http.StatusForbidden, "Not implemented")
}

//HostVIPRelease release all the vip in the host
func (allo *Allocator) HostVIPRelease(request *restful.Request, response *restful.Response) {
	started := time.Now()
	hostIP := request.PathParameter("hostip")
	netRes := &types.NetResponse{
		Type: types.ResponseType_RELEASE,
	}
	if err := allo.netSvr.HostVIPRelease(hostIP); err != nil {
		blog.Errorf("Allocator release host %s vip failed, %s", hostIP, err.Error())
		netRes.Code = 2
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("hostRelease", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = SUCCESS
	response.WriteEntity(netRes)
	reportMetrics("hostRelease", "2xx", started)
}
