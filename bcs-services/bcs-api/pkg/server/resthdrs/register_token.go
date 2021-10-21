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

package resthdrs

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/metric"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/filters"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"

	"github.com/emicklei/go-restful"
	"time"
)

func ListRegisterTokens(request *restful.Request, response *restful.Response) {

	start := time.Now()

	cluster := filters.GetCluster(request)

	token := sqlstore.GetRegisterToken(cluster.ID)
	if token == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, register token not found", common.BcsErrApiBadRequest)
		WriteClientError(response, "RTOKEN_NOT_FOUND", message)
		return
	}
	response.WriteEntity([]*m.RegisterToken{token})

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}

func CreateRegisterToken(request *restful.Request, response *restful.Response) {

	start := time.Now()

	cluster := filters.GetCluster(request)
	clusterId := cluster.ID

	err := sqlstore.CreateRegisterToken(clusterId)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, can not create register token: %s", common.BcsErrApiBadRequest, err.Error())
		WriteServerError(response, "CANNOT_CREATE_RTOKEN", message)
		return
	}
	response.WriteEntity([]*m.RegisterToken{
		sqlstore.GetRegisterToken(clusterId),
	})

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}
