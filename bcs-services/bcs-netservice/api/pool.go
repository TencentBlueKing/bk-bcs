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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/netservice"
	"net/http"
	"time"

	restful "github.com/emicklei/go-restful"
)

//RegisterPoolHandler create pool handler,
//url link : v1/pool
func RegisterPoolHandler(httpSvr *HTTPService, logic *netservice.NetService) *PoolHandler {
	handler := &PoolHandler{
		netSvr: logic,
	}
	webSvr := new(restful.WebService)
	//add http handler
	webSvr.Path("/v1/pool").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	//Add pool
	webSvr.Route(webSvr.POST("").To(handler.Add))
	//get all pool info
	webSvr.Route(webSvr.GET("").To(handler.List))
	//Delete by pool net
	webSvr.Route(webSvr.DELETE("/{cluster}/{net}").To(handler.Delete))
	//update by pool net
	webSvr.Route(webSvr.PUT("/{cluster}/{net}").To(handler.Update))
	//get one pool by net
	webSvr.Route(webSvr.GET("/{cluster}/{net}").To(handler.ListByID))
	//query cluster info, query parameter
	//info:
	//    static(default): static info for cluster
	//    detail: all pool info under cluster
	//sort: (todo feature)
	webSvr.Route(webSvr.GET("/{cluster}").To(handler.Query))

	httpSvr.Register(webSvr)

	return handler
}

//PoolHandler http request handler
type PoolHandler struct {
	netSvr *netservice.NetService
}

//Add add new pool
func (pool *PoolHandler) Add(request *restful.Request, response *restful.Response) {
	started := time.Now()
	netReq := &types.NetRequest{}
	if err := request.ReadEntity(netReq); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		blog.Errorf("PoolHandler[Add] json decode Err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("createIPPool", "4xx", started)
		return
	}
	netRes := &types.NetResponse{
		Type: types.ResponseType_POOL,
	}
	if netReq.Type != types.RequestType_POOL || netReq.Pool == nil {
		netRes.Code = 1
		netRes.Message = "Request Type Err or Pool lost"
		blog.Errorf("PoolHandler check Pool request, but got unexpect type %d, pool %v", netReq.Type, netReq.Pool)
		response.WriteEntity(netRes)
		reportMetrics("createIPPool", "4xx", started)
		return
	}
	if !netReq.Pool.IsValid() {
		netRes.Code = 1
		netRes.Message = "Request Pool data lost"
		blog.Errorf("PoolHandler check pool data err, data lost, Net: %s, Mask: %d, Gateway: %s", netReq.Pool.Net, netReq.Pool.Mask, netReq.Pool.Gateway)
		response.WriteEntity(netRes)
		reportMetrics("createIPPool", "4xx", started)
		return
	}
	netRes.Pool = append(netRes.Pool, netReq.Pool)
	if err := pool.netSvr.AddPool(netReq.Pool); err != nil {
		netRes.Code = 2
		netRes.Message = err.Error()
		blog.Errorf("PoolHandler add pool Err: %s", err.Error())
		response.WriteEntity(netRes)
		reportMetrics("createIPPool", "5xx", started)
		return
	}
	blog.Info("NetPool %s/%s mask %d gateway %s add succ", netReq.Pool.Cluster, netReq.Pool.Net, netReq.Pool.Mask, netReq.Pool.Gateway)
	netRes.Code = 0
	netRes.Message = "success"
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("PoolHandler reply client POST request Err: %v", err)
	}
	reportMetrics("createIPPool", "2xx", started)
}

//Delete delete pool by ip segment
func (pool *PoolHandler) Delete(request *restful.Request, response *restful.Response) {
	started := time.Now()
	netKey := request.PathParameter("net")
	netCluster := request.PathParameter("cluster")
	netRes := &types.NetResponse{
		Type: types.ResponseType_POOL,
	}
	if len(netKey) == 0 || len(netCluster) == 0 {
		netRes.Code = 1
		netRes.Message = "Lost param needed"
		response.WriteEntity(netRes)
		reportMetrics("deleteIPPool", "4xx", started)
		return
	}
	if err := pool.netSvr.DeletePool(netCluster + "/" + netKey); err != nil {
		blog.Errorf("NetPool Delete %s/%s request err: %s", netCluster, netKey, err.Error())
		netRes.Code = 1
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("deleteIPPool", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = SUCCESS
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("PoolHandler reply client DELETE request Err: %v", err)
	}
	reportMetrics("deleteIPPool", "2xx", started)
}

//Update update pool by ip segment
func (pool *PoolHandler) Update(request *restful.Request, response *restful.Response) {
	started := time.Now()
	netReq := &types.NetRequest{}
	if err := request.ReadEntity(netReq); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		blog.Errorf("PoolHandler[Update] json decode Err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("updateIPPool", "4xx", started)
		return
	}
	netKey := request.PathParameter("net")
	netCluster := request.PathParameter("cluster")
	netRes := &types.NetResponse{
		Type: types.ResponseType_POOL,
	}
	if netReq.Type != types.RequestType_POOL || netReq.Pool == nil {
		netRes.Code = 1
		netRes.Message = "Request Type Err or Pool lost"
		blog.Errorf("PoolHandler check Pool request, but got unexpect type %d, pool %v", netReq.Type, netReq.Pool)
		response.WriteEntity(netRes)
		reportMetrics("updateIPPool", "4xx", started)
		return
	}
	if !netReq.Pool.IsValid() {
		netRes.Code = 1
		netRes.Message = "Request Pool data lost"
		blog.Errorf(
			"PoolHandler check pool data err, data lost, Cluster: %s, Net: %s, Mask: %d, Gateway: %s",
			netReq.Pool.Cluster,
			netReq.Pool.Net,
			netReq.Pool.Mask,
			netReq.Pool.Gateway,
		)
		response.WriteEntity(netRes)
		reportMetrics("updateIPPool", "4xx", started)
		return
	}

	netRes.Pool = append(netRes.Pool, netReq.Pool)
	if err := pool.netSvr.UpdatePool(netReq.Pool, netCluster+"/"+netKey); err != nil {
		netRes.Code = 2
		netRes.Message = err.Error()
		blog.Errorf("PoolHandler add pool Err: %s", err.Error())
		response.WriteEntity(netRes)
		reportMetrics("updateIPPool", "5xx", started)
		return
	}
	blog.Info("NetPool %s/%s mask %d gateway %s update succ", netReq.Pool.Cluster, netReq.Pool.Net, netReq.Pool.Mask, netReq.Pool.Gateway)
	netRes.Code = 0
	netRes.Message = SUCCESS
	if len(netReq.Pool.Available) == 0 && len(netReq.Pool.Reserved) == 0 {
		netRes.Message = "update nothing"
	}
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("PoolHandler reply client POST request Err: %v", err)
	}
	reportMetrics("updateIPPool", "2xx", started)
}

//List list all pools
func (pool *PoolHandler) List(request *restful.Request, response *restful.Response) {
	started := time.Now()
	//list all pools
	netRes := &types.NetResponse{
		Type: types.ResponseType_POOL,
	}
	allPools, err := pool.netSvr.ListPool()
	if err != nil {
		blog.Errorf("NetPool List all request err: %s", err.Error())
		netRes.Code = 1
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("listIPPool", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = SUCCESS
	netRes.Pool = allPools
	netRes.Data = netRes.Pool
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("PoolHandler reply client GET request Err: %v", err)
	}
	reportMetrics("listIPPool", "2xx", started)
}

//ListByID list all pools
func (pool *PoolHandler) ListByID(request *restful.Request, response *restful.Response) {
	started := time.Now()
	//or list pool by pool id
	netKey := request.PathParameter("net")
	netCluster := request.PathParameter("cluster")
	netRes := &types.NetResponse{
		Type: types.ResponseType_POOL,
	}
	p, err := pool.netSvr.ListPoolByKey(netCluster + "/" + netKey)
	if err != nil {
		blog.Errorf("NetPool list pool %s/%s request err: %s", netCluster, netKey, err.Error())
		netRes.Code = 1
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("listIPPoolByID", "5xx", started)
		return
	}
	netRes.Code = 0
	netRes.Message = "success"
	netRes.Pool = append(netRes.Pool, p)
	netRes.Data = netRes.Pool
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("PoolHandler reply client GET %s/%s request Err: %v", netCluster, netKey, err)
	}
	reportMetrics("listIPPoolByID", "2xx", started)
}

//Query list all pools
func (pool *PoolHandler) Query(request *restful.Request, response *restful.Response) {
	started := time.Now()
	cluster := request.PathParameter("cluster")
	info := request.QueryParameter("info")
	netRes := &types.NetResponse{
		Type: types.ResponseType_PSTATIC,
	}
	pools, err := pool.netSvr.ListPoolByCluster(cluster)
	if err != nil {
		blog.Errorf("NetPool List cluster %s request err %v", cluster, err)
		netRes.Code = 1
		netRes.Message = err.Error()
		response.WriteEntity(netRes)
		reportMetrics("queryIPPool", "5xx", started)
		return
	}
	var statis types.NetStatic
	if info == "detail" {
		//reply detail data
		netRes.Code = 0
		netRes.Message = fmt.Sprintf("query net pool under %s succ", cluster)
		netRes.Type = types.ResponseType_POOL
		netRes.Pool = pools
		netRes.Data = netRes.Pool
	} else {
		//static info
		netRes.Code = 0
		netRes.Message = "query statistic info succ"
		statis.PoolNum = len(pools)
		for _, p := range pools {
			statis.ActiveIP += len(p.Active)
			statis.AvailableIP += len(p.Available)
			statis.ReservedIP += len(p.Reserved)
		}
		netRes.PStatic = &statis
		netRes.Data = netRes.PStatic
	}
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("PoolHandler reply client Query %s request Err: %v", cluster, err)
	}
	reportMetrics("queryIPPool", "2xx", started)
}
