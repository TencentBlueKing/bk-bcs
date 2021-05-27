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

package v1http

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/emicklei/go-restful"
)

//CreateRegisterToken http handler for register specified cluster token
func CreateRegisterToken(request *restful.Request, response *restful.Response) {
	start := time.Now()

	clusterID := request.PathParameter("cluster_id")
	clusterInDb := sqlstore.GetCluster(clusterID)
	if clusterInDb == nil {
		metrics.ReportRequestAPIMetrics("CreateRegisterToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("create register_token failed, cluster [%s] not exist", clusterID)
		message := fmt.Sprintf("errcode: %d, create register_token failed, cluster [%s] not exist",
			common.BcsErrApiBadRequest, clusterID)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	err := sqlstore.CreateRegisterToken(clusterID)
	if err != nil {
		metrics.ReportRequestAPIMetrics("CreateRegisterToken", request.Request.Method, metrics.ErrStatus, start)
		blog.Errorf("failed to create register_token for cluster [%s]: %s", clusterID, err.Error())
		message := fmt.Sprintf("errcode: %d, can not create register token: %s",
			common.BcsErrApiBadRequest, err.Error())
		utils.WriteServerError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", sqlstore.GetRegisterToken(clusterID))
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("CreateRegisterToken", request.Request.Method, metrics.SucStatus, start)
}

//GetRegisterToken http handler for search specified cluster token
//it's served for bcs-gateway-discovery for cluster service discovery
func GetRegisterToken(request *restful.Request, response *restful.Response) {
	start := time.Now()

	clusterID := request.PathParameter("cluster_id")
	token := sqlstore.GetRegisterToken(clusterID)
	if token == nil {
		metrics.ReportRequestAPIMetrics("GetRegisterToken", request.Request.Method, metrics.ErrStatus, start)
		message := fmt.Sprintf("errcode: %d, register token not found", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	data := utils.CreateResponeData(nil, "success", token)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetRegisterToken", request.Request.Method, metrics.ErrStatus, start)
}
