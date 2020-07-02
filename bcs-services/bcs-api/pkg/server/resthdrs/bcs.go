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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/metric"
	"github.com/emicklei/go-restful"
	"time"
)

type QueryBCSCredentialsForm struct {
	ProjectID string `json:"project_id" validate:"required"`
	ClusterID string `json:"cluster_id" validate:"required"`
}

type QueryBCSCredentialsByClusterIdForm struct {
	ClusterID string `json:"cluster_id" validate:"required"`
}

// QueryBCSClusterByID query for bke cluster info by given BCS "project_id + cluster_id"
func QueryBCSClusterByID(request *restful.Request, response *restful.Response) {

	start := time.Now()

	// request.readEntity can not read from GET parameters, so we will build the struct on our own
	form := QueryBCSCredentialsForm{
		ProjectID: request.QueryParameter("project_id"),
		ClusterID: request.QueryParameter("cluster_id"),
	}
	err := validate.Struct(&form)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Debug(fmt.Sprintf("QueryBCSClusterByID form validate failed, %s", err))
		response.WriteEntity(FormatValidationError(err))
		return
	}

	// get cluster
	cluster := sqlstore.GetClusterByBCSInfo(form.ProjectID, form.ClusterID)
	if cluster == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("errcode: %d, cluster with project_id=%s cluster_id=%s not found", common.BcsErrApiBadRequest, form.ProjectID, form.ClusterID)
		blog.Warnf(message)
		WriteNotFoundError(response, "CLUSTER_NOT_FOUND", message)
		return
	}

	// Force a permission sync
	/*
		user := filters.GetUser(request)
		err = auth.SyncUserClusterPerms(user, cluster)
		if err != nil {
			WriteServerError(response, "SERVER_ERROR", fmt.Sprintf("permission check error: can not sync permission from external service"))
			return
		}

		// Check cluster permission
		if !auth.UserHasClusterPerm(user, cluster, m.ClusterPermNameView) {
			WriteForbiddenError(response, "PERMISION_DENIED", fmt.Sprintf("current user is not authorized to perform this action"))
			return
		}
	*/

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())

	response.WriteEntity(cluster)
}

// QueryBCSClusterByClusterID query for bke cluster info by given BCS cluster_id
func QueryBCSClusterByClusterID(request *restful.Request, response *restful.Response) {

	start := time.Now()

	// request.readEntity can not read from GET parameters, so we will build the struct on our own
	form := QueryBCSCredentialsByClusterIdForm{
		ClusterID: request.QueryParameter("cluster_id"),
	}
	err := validate.Struct(&form)
	if err != nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		blog.Debug(fmt.Sprintf("QueryBCSClusterByClusterID form validate failed, %s", err))
		response.WriteEntity(FormatValidationError(err))
		return
	}

	// get cluster
	cluster := sqlstore.GetClusterByBCSInfo("", form.ClusterID)
	if cluster == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
		message := fmt.Sprintf("cluster with cluster_id=%s not found", form.ClusterID)
		blog.Warnf(message)
		WriteNotFoundError(response, "CLUSTER_NOT_FOUND", message)
		return
	}

	response.WriteEntity(cluster)

	metric.RequestCount.WithLabelValues("k8s_rest", request.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("k8s_rest", request.Request.Method).Observe(time.Since(start).Seconds())
}
