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
 */

package metrics

import (
	"context"
	"strings"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// PodUsageQuery Pod 查询
type PodUsageQuery struct {
	UsageQuery  `json:",inline"`
	PodNameList []string `json:"pod_name_list"`
}

// handlePodMetric Pod 处理公共函数
func handlePodMetric(c *rest.Context, promql string, query *PodUsageQuery) (*promclient.ResultData, error) {

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"clusterId":   c.ClusterId,
		"namespace":   query.Namespace,
		"podNameList": strings.Join(query.PodNameList, "|"),
		"provider":    PROVIDER,
	}

	result, err := bcsmonitor.QueryRange(c.Request.Context(), c.ProjectCode, promql, params, queryTime.Start,
		queryTime.End, queryTime.Step)

	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// PodCPUUsage Pod CPU使用率
// @Summary Pod CPU使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/cpu_usage [POST]
func PodCPUUsage(c context.Context, req *PodUsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql :=
		`bcs:pod:cpu_usage{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}` // nolint

	return handlePodMetric(rctx, promql, req)
}

// PodCPULimitUsage Pod Limit CPU使用率
// @Summary Pod Limit CPU使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/cpu_limit_usage [POST]
func PodCPULimitUsage(c context.Context, req *PodUsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql :=
		`bcs:pod:cpu_limit_usage{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}` // nolint

	return handlePodMetric(rctx, promql, req)
}

// PodCPURequestUsage Pod Request CPU使用率
// @Summary Pod Request CPU使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/cpu_request_usage [POST]
func PodCPURequestUsage(c context.Context, req *PodUsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql :=
		`bcs:pod:cpu_request_usage{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}` // nolint

	return handlePodMetric(rctx, promql, req)
}

// PodMemoryUsed Pod 内存使用量
// @Summary Pod 内存使用量
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/memory_used [POST]
func PodMemoryUsed(c context.Context, req *PodUsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql :=
		`bcs:pod:memory_used{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}` // nolint

	return handlePodMetric(rctx, promql, req)
}

// PodNetworkReceive 网络接收量
// @Summary 网络接收量
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/network_receive [POST]
func PodNetworkReceive(c context.Context, req *PodUsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql :=
		`bcs:pod:network_receive{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}` // nolint

	return handlePodMetric(rctx, promql, req)
}

// PodNetworkTransmit Pod 网络发送量
// @Summary Pod 网络发送量
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /namespaces/:namespace/pods/network_transmit [POST]
func PodNetworkTransmit(c context.Context, req *PodUsageQuery) (*promclient.ResultData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	promql :=
		`bcs:pod:network_transmit{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podNameList>s", %<provider>s}` // nolint

	return handlePodMetric(rctx, promql, req)
}
