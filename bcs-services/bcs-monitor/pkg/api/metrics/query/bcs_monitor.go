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

package query

import (
	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const (
	// PROVIDER provider
	PROVIDER = `provider="BCS_SYSTEM"`
)

// BCSMonitorHandler metric handler
type BCSMonitorHandler struct {
}

// handleClusterMetric Cluster 处理公共函数
func handleClusterMetric(c *rest.Context, promql string, query UsageQuery) (promclient.ResultData, error) {

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return promclient.ResultData{}, err
	}
	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"provider":  PROVIDER,
	}

	result, err := bcsmonitor.QueryRange(c.Request.Context(), c.ProjectCode, promql, params, queryTime.Start,
		queryTime.End, queryTime.Step)
	if err != nil {
		return promclient.ResultData{}, err
	}
	return result.Data, nil
}

// GetClusterOverview 获取集群概览
func (h BCSMonitorHandler) GetClusterOverview(c *rest.Context) (ClusterOverviewMetric, error) {
	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"provider":  PROVIDER,
	}

	promqlMap := map[string]string{
		"cpu_used":       `bcs:cluster:cpu:used{cluster_id="%<clusterId>s", %<provider>s}`,
		"cpu_request":    `bcs:cluster:cpu:request{cluster_id="%<clusterId>s", %<provider>s}`,
		"cpu_total":      `bcs:cluster:cpu:total{cluster_id="%<clusterId>s", %<provider>s}`,
		"memory_used":    `bcs:cluster:memory:used{cluster_id="%<clusterId>s", %<provider>s}`,
		"memory_request": `bcs:cluster:memory:request{cluster_id="%<clusterId>s", %<provider>s}`,
		"memory_total":   `bcs:cluster:memory:total{cluster_id="%<clusterId>s", %<provider>s}`,
		"disk_used":      `bcs:cluster:disk:used{cluster_id="%<clusterId>s", %<provider>s}`,
		"disk_total":     `bcs:cluster:disk:total{cluster_id="%<clusterId>s", %<provider>s}`,
		"diskio_used":    `bcs:cluster:diskio:used{cluster_id="%<clusterId>s", %<provider>s}`,
		"diskio_total":   `bcs:cluster:diskio:total{cluster_id="%<clusterId>s", %<provider>s}`,
		"pod_used":       `bcs:cluster:pod:used{cluster_id="%<clusterId>s", %<provider>s}`,
		"pod_total":      `bcs:cluster:pod:total{cluster_id="%<clusterId>s", %<provider>s}`,
	}

	result, err := bcsmonitor.QueryMultiValues(c.Request.Context(), c.ProjectId, promqlMap, params,
		utils.GetNowQueryTime())
	if err != nil {
		return ClusterOverviewMetric{}, err
	}

	m := ClusterOverviewMetric{
		CPUUsage: &Usage{
			Used:    result["cpu_used"],
			Request: result["cpu_request"],
			Total:   result["cpu_total"],
		},
		MemoryUsage: &UsageByte{
			UsedByte:    result["memory_used"],
			RequestByte: result["memory_request"],
			TotalByte:   result["memory_total"],
		},
		DiskUsage: &UsageByte{
			UsedByte:  result["disk_used"],
			TotalByte: result["disk_total"],
		},
		DiskIOUsage: &Usage{
			Used:  result["diskio_used"],
			Total: result["diskio_total"],
		},
		PodUsage: &Usage{
			Used:  result["pod_used"],
			Total: result["pod_total"],
		},
	}

	return m, nil
}

// ClusterPodUsage implements Handler.
func (BCSMonitorHandler) ClusterPodUsage(c *rest.Context, query UsageQuery) (promclient.ResultData, error) {
	promql := `bcs:cluster:pod:usage{cluster_id="%<clusterId>s", %<provider>s}`
	return handleClusterMetric(c, promql, query)
}

// ClusterCPUUsage implements Handler.
func (BCSMonitorHandler) ClusterCPUUsage(c *rest.Context, query UsageQuery) (promclient.ResultData, error) {
	promql := `bcs:cluster:cpu:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql, query)
}

// ClusterCPURequestUsage implements Handler.
func (BCSMonitorHandler) ClusterCPURequestUsage(c *rest.Context, query UsageQuery) (promclient.ResultData, error) {
	promql := `bcs:cluster:cpu_request:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql, query)
}

// ClusterMemoryUsage implements Handler.
func (BCSMonitorHandler) ClusterMemoryUsage(c *rest.Context, query UsageQuery) (promclient.ResultData, error) {
	promql := `bcs:cluster:memory:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql, query)
}

// ClusterMemoryRequestUsage implements Handler.
func (BCSMonitorHandler) ClusterMemoryRequestUsage(c *rest.Context, query UsageQuery) (promclient.ResultData, error) {
	promql := `bcs:cluster:memory_request:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql, query)
}

// ClusterDiskUsage implements Handler.
func (BCSMonitorHandler) ClusterDiskUsage(c *rest.Context, query UsageQuery) (promclient.ResultData, error) {
	promql := `bcs:cluster:disk:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql, query)
}

// ClusterDiskioUsage implements Handler.
func (BCSMonitorHandler) ClusterDiskioUsage(c *rest.Context, query UsageQuery) (promclient.ResultData, error) {
	promql := `bcs:cluster:diskio:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql, query)
}

// NewBCSMonitorHandler new handler
func NewBCSMonitorHandler() *BCSMonitorHandler {
	return &BCSMonitorHandler{}
}

var _ Handler = BCSMonitorHandler{}
