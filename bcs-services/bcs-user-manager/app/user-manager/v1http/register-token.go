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

	clusterId := request.PathParameter("cluster_id")
	clusterInDb := sqlstore.GetCluster(clusterId)
	if clusterInDb == nil {
		metrics.RequestErrorCount.WithLabelValues("register-token", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("register-token", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Warnf("create register_token failed, cluster [%s] not exist", clusterId)
		message := fmt.Sprintf("errcode: %d, create register_token failed, cluster [%s] not exist", common.BcsErrApiBadRequest, clusterId)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	err := sqlstore.CreateRegisterToken(clusterId)
	if err != nil {
		metrics.RequestErrorCount.WithLabelValues("register-token", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("register-token", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Errorf("failed to create register_token for cluster [%s]: %s", clusterId, err.Error())
		message := fmt.Sprintf("errcode: %d, can not create register token: %s", common.BcsErrApiBadRequest, err.Error())
		utils.WriteServerError(response, common.BcsErrApiBadRequest, message)
		return
	}

	data := utils.CreateResponeData(nil, "success", sqlstore.GetRegisterToken(clusterId))
	response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("register-token", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("register-token", request.Request.Method).Observe(time.Since(start).Seconds())
}

//GetRegisterToken http handler for search specified cluster token
//it's served for bcs-gateway-discovery for cluster service discovery
func GetRegisterToken(request *restful.Request, response *restful.Response) {
	start := time.Now()

	clusterId := request.PathParameter("cluster_id")
	token := sqlstore.GetRegisterToken(clusterId)
	if token == nil {
		metrics.RequestErrorCount.WithLabelValues("register-token", request.Request.Method).Inc()
		metrics.RequestErrorLatency.WithLabelValues("register-token", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, register token not found", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	data := utils.CreateResponeData(nil, "success", token)
	response.Write([]byte(data))

	metrics.RequestCount.WithLabelValues("register-token", request.Request.Method).Inc()
	metrics.RequestLatency.WithLabelValues("register-token", request.Request.Method).Observe(time.Since(start).Seconds())
}
