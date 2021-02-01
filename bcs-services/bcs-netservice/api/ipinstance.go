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

//RegisterIPInstanceHandler create pool handler,
//url link : v1/ipinstanace
func RegisterIPInstanceHandler(httpSvr *HTTPService, logic *netservice.NetService) *IPInstanceHandler {
	handler := &IPInstanceHandler{
		netSvr: logic,
	}
	webSvr := new(restful.WebService)
	//add http handler
	webSvr.Path("/v1/ipinstance").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	//update ipinstanace
	webSvr.Route(webSvr.PUT("").To(handler.Update))
	webSvr.Route(webSvr.PUT("status").To(handler.TransferIPAttr))
	httpSvr.Register(webSvr)

	return handler
}

//IPInstanceHandler http request handler
type IPInstanceHandler struct {
	netSvr *netservice.NetService
}

//Update update pool by ip segment
func (inst *IPInstanceHandler) Update(request *restful.Request, response *restful.Response) {
	started := time.Now()
	netReq := &types.IPInst{}
	if err := request.ReadEntity(netReq); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		blog.Errorf("IPInstance [Update] json decode Err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("updateAvailableIP", "4xx", started)
		return
	}
	netRes := &types.SvcResponse{}
	if err := inst.netSvr.UpdateAvailableIPInstance(netReq); err != nil {
		netRes.Code = 2
		netRes.Message = err.Error()
		blog.Errorf("IPInstance Update %s", err.Error())
		response.WriteEntity(netRes)
		reportMetrics("updateAvailableIP", "5xx", started)
		return
	}
	blog.Info("IPInstance %s %s Update available %s succ", netReq.Cluster, netReq.Pool, netReq.IPAddr)
	netRes.Code = 0
	netRes.Message = SUCCESS
	if err := response.WriteEntity(netRes); err != nil {
		blog.Errorf("IPInstance Update reply client POST request Err: %v", err)
	}
	reportMetrics("updateAvailableIP", "2xx", started)
}

//TransferIPAttr transfer ip attr from available/reserved to reserved/available
func (inst *IPInstanceHandler) TransferIPAttr(request *restful.Request, response *restful.Response) {
	started := time.Now()
	tranInput := &types.TranIPAttrInput{}
	if err := request.ReadEntity(tranInput); err != nil {
		response.AddHeader("Content-Type", "text/plain")
		blog.Errorf("TransferIPAttr [Update] json decode Err: %s", err.Error())
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		reportMetrics("transferIPAttr", "4xx", started)
		return
	}
	tranOutput := &types.TranIPAttrOutput{}
	if !tranInput.IsValid() {
		tranOutput.Code = 1
		tranOutput.Message = "invalid param,please check net and cluster and iplist can not be empty,src and dest must be available/reserved"
		blog.Errorf("invalid param:%v", tranInput)
		response.WriteEntity(tranOutput)
		reportMetrics("transferIPAttr", "4xx", started)
	}
	if failedCode, err := inst.netSvr.TransferIPAttribute(tranInput); err != nil {
		tranOutput.Code = failedCode
		tranOutput.Message = err.Error()
		blog.Errorf("TransferIPAttribute failed:%s", err.Error())
		response.WriteEntity(tranOutput)
		reportMetrics("transferIPAttr", "5xx", started)
		return
	}
	tranOutput.Code = 0
	tranOutput.Message = SUCCESS
	if err := response.WriteEntity(tranOutput); err != nil {
		blog.Errorf("TransferIPAttr client POST request Err: %v", err)
		return
	}
	reportMetrics("transferIPAttr", "2xx", started)
	blog.Infof("TransferIPAttr %v from %s to %s success", tranInput.IPList, tranInput.SrcStatus, tranInput.DestStatus)
}
