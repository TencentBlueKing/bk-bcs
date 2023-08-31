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
	"time"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

const (
	// PROVIDER provider
	PROVIDER = `provider="BCS_SYSTEM"`
)

// Usage 使用量
type Usage struct {
	Used    string `json:"used"`
	Request string `json:"request"`
	Total   string `json:"total"`
}

// UsageByte 使用量, bytes单位
type UsageByte struct {
	UsedByte    string `json:"used_bytes"`
	RequestByte string `json:"request_bytes"`
	TotalByte   string `json:"total_bytes"`
}

// ClusterOverviewMetric 集群概览接口
type ClusterOverviewMetric struct {
	CPUUsage    *Usage     `json:"cpu_usage"`
	DiskUsage   *UsageByte `json:"disk_usage"`
	MemoryUsage *UsageByte `json:"memory_usage"`
	DiskIOUsage *Usage     `json:"diskio_usage"`
	PodUsage    *Usage     `json:"pod_usage"`
}

// handleClusterMetric Cluster 处理公共函数
func handleClusterMetric(c *rest.Context, promql string) (interface{}, error) {
	query := &UsageQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		return nil, err
	}

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return nil, err
	}
	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"provider":  PROVIDER,
	}

	result, err := bcsmonitor.QueryRange(c.Request.Context(), c.ProjectCode, promql, params, queryTime.Start,
		queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetClusterOverview 集群概览数据
// @Summary 集群概览数据
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /overview [get]
func GetClusterOverview(c *rest.Context) (interface{}, error) {
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

	result, err := bcsmonitor.QueryMultiValues(c.Request.Context(), c.ProjectId, promqlMap, params, time.Now())
	if err != nil {
		return nil, err
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

// ClusterPodUsage 集群 POD 使用率
// @Summary 集群 POD 使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /pod_usage [get]
func ClusterPodUsage(c *rest.Context) (interface{}, error) {
	// 获取集群中节点列表
	promql := `bcs:cluster:pod:usage{cluster_id="%<clusterId>s", %<provider>s}`
	return handleClusterMetric(c, promql)
}

// ClusterCPUUsage 集群 CPU 使用率
// @Summary 集群 CPU 使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /cpu_usage [get]
func ClusterCPUUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:cpu:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql)

}

// ClusterCPURequestUsage 集群 CPU 装箱率
// @Summary 集群 CPU 装箱率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /cpu_request_usage [get]
func ClusterCPURequestUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:cpu_request:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql)

}

// ClusterMemoryUsage 集群内存使用率
// @Summary 集群内存使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /memory_usage [get]
func ClusterMemoryUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:memory:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql)
}

// ClusterMemoryRequestUsage 集群内存装箱率
// @Summary 集群内存装箱率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /memory_request_usage [get]
func ClusterMemoryRequestUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:memory_request:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql)
}

// ClusterDiskUsage 集群磁盘使用率
// @Summary 集群磁盘使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /disk_usage [get]
func ClusterDiskUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:disk:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql)
}

// ClusterDiskioUsage 集群磁盘IO使用率
// @Summary 集群磁盘IO使用率
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /diskio_usage [get]
func ClusterDiskioUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:cluster:diskio:usage{cluster_id="%<clusterId>s", %<provider>s}`

	return handleClusterMetric(c, promql)
}
