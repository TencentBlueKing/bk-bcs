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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
	"github.com/pkg/errors"
)

// NodeOveriewMetric 节点概览
type NodeOveriewMetric struct {
	ContainerCount string `json:"container_count"`
	PodCount       string `json:"pod_count"`
	CPUUsage       string `json:"cpu_usage"`
	DiskUsage      string `json:"disk_usage"`
	DiskioUsage    string `json:"diskio_usage"`
	MemoryUsage    string `json:"memory_usage"`
}

// UsageQuery 节点查询
type UsageQuery struct {
	StartAt string `json:"start_at" form:"start_at"` // 必填参数`
	EndAt   string `json:"end_at" form:"end_at"`
}

// parseTime 兼容前端多个格式
func parseTime(rawTime string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		t, err := time.Parse(format, rawTime)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.Errorf("invalid datetime %s", rawTime)
}

// GetQueryTime 转换为 promql 查询时间
func (q *UsageQuery) GetQueryTime() (*clientutil.PromQueryTime, error) {
	queryTime := &clientutil.PromQueryTime{}

	if q.EndAt == "" {
		queryTime.End = time.Now()
	} else {
		t, err := parseTime(q.EndAt)
		if err != nil {
			return nil, err
		}
		queryTime.End = t
	}

	if q.StartAt == "" {
		queryTime.Start = queryTime.End.Add(-time.Hour)
	} else {
		t, err := parseTime(q.StartAt)
		if err != nil {
			return nil, err
		}
		queryTime.Start = t
	}

	// 默认只返回 60 个点
	queryTime.Step = queryTime.End.Sub(queryTime.Start) / 60

	return queryTime, nil
}

// handleNodeMetric Node 处理公共函数
func handleNodeMetric(c *rest.Context, promql string) (interface{}, error) {
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
		"ip":        c.Param("ip"),
	}

	result, err := bcsmonitor.QueryRange(c.Context, c.ProjectCode, promql, params, queryTime.Start, queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetNodeInfo 节点信息
// @Summary  节点信息
// @Tags     Metrics
// @Success  200  {string}  string
// @Router   /nodes/:ip/info [get]
func GetNodeInfo(c *rest.Context) (interface{}, error) {
	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"ip":        c.Param("ip"),
	}

	promql := `bcs:node:info{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`
	labelSet, err := bcsmonitor.QueryLabelSet(c.Context, c.ProjectId, promql, params, time.Now())
	if err != nil {
		return nil, err
	}
	return labelSet, nil
}

// GetNodeOverview 查询节点概览
// @Summary  查询节点概览
// @Tags     Metrics
// @Success  200  {string}  string
// @Router   /nodes/:ip/overview [get]
func GetNodeOverview(c *rest.Context) (interface{}, error) {
	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"ip":        c.Param("ip"),
	}

	promqlMap := map[string]string{
		"cpu":             `bcs:node:cpu:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`,
		"memory":          `bcs:node:memory:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`,
		"disk":            `bcs:node:disk:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`,
		"diskio":          `bcs:node:diskio:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`,
		"container_count": `bcs:node:container_count{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`,
		"pod_count":       `bcs:node:pod_count{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`,
	}

	result, err := bcsmonitor.QueryMultiValues(c.Context, c.ProjectId, promqlMap, params, time.Now())
	if err != nil {
		return nil, err
	}

	overview := &NodeOveriewMetric{
		CPUUsage:       result["cpu"],
		MemoryUsage:    result["memory"],
		DiskUsage:      result["disk"],
		DiskioUsage:    result["diskio"],
		ContainerCount: result["container_count"],
		PodCount:       result["pod_count"],
	}

	return overview, nil
}

// GetNodeCPUUsage 查询 CPU 使用率
// @Summary  查询 CPU 使用率
// @Tags     Metrics
// @Success  200  {string}  string
// @Router   /nodes/:ip/cpu_usage [get]
func GetNodeCPUUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:node:cpu:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`

	return handleNodeMetric(c, promql)

}

// GetNodeMemoryUsage 节点内存使用率
// @Summary  节点内存使用率
// @Tags     Metrics
// @Success  200  {string}  string
// @Router   /nodes/:ip/memory_usage [get]
func GetNodeMemoryUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:node:memory:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`

	return handleNodeMetric(c, promql)

}

// GetNodeNetworkTransmitUsage 节点网络发送
// @Summary  节点网络发送
// @Tags     Metrics
// @Success  200  {string}  string
// @Router   /nodes/:ip/network_receive [get]
func GetNodeNetworkTransmitUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:node:network_transmit{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`

	return handleNodeMetric(c, promql)

}

// GetNodeNetworkReceiveUsage 节点网络接收
// @Summary  节点网络接收
// @Tags     Metrics
// @Success  200  {string}  string
// @Router   /nodes/:ip/network_transmit [get]
func GetNodeNetworkReceiveUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:node:network_receive{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`

	return handleNodeMetric(c, promql)

}

// GetNodeDiskioUsage 节点磁盘IO
// @Summary  节点磁盘IO
// @Tags     Metrics
// @Success  200  {string}  string
// @Router   /nodes/:ip/diskio_usage [get]
func GetNodeDiskioUsage(c *rest.Context) (interface{}, error) {
	promql := `bcs:node:diskio:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BCS_SYSTEM"}`

	return handleNodeMetric(c, promql)
}
