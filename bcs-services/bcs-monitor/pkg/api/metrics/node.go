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

// GetNodeInfo 节点信息
func GetNodeInfo(c *rest.Context) (interface{}, error) {
	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"ip":        c.Param("ip"),
	}

	promql := `bcs:node:info{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	labelSet, err := bcsmonitor.QueryLabelSet(c.Context, c.ProjectId, promql, params, time.Now())
	if err != nil {
		return nil, err
	}
	return labelSet, nil
}

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

// GetNodeOverview 查询节点概览
func GetNodeOverview(c *rest.Context) (interface{}, error) {
	params := map[string]interface{}{
		"clusterId": c.ClusterId,
		"ip":        c.Param("ip"),
	}
	overview := &NodeOveriewMetric{}

	cpuPromQL := `bcs:node:cpu:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	cpuUsage, err := bcsmonitor.QueryValue(c.Context, c.ProjectId, cpuPromQL, params, time.Now())
	if err != nil {
		return nil, err
	}
	overview.CPUUsage = cpuUsage

	memoryPromQL := `bcs:node:memory:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	memory, err := bcsmonitor.QueryValue(c.Context, c.ProjectId, memoryPromQL, params, time.Now())
	if err != nil {
		return nil, err
	}
	overview.MemoryUsage = memory

	diskPromQL := `bcs:node:disk:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	disk, err := bcsmonitor.QueryValue(c.Context, c.ProjectId, diskPromQL, params, time.Now())
	if err != nil {
		return nil, err
	}
	overview.DiskUsage = disk

	diskioPromQL := `bcs:node:diskio:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	diskio, err := bcsmonitor.QueryValue(c.Context, c.ProjectId, diskioPromQL, params, time.Now())
	if err != nil {
		return nil, err
	}
	overview.DiskioUsage = diskio

	containerCountPromQL := `bcs:node:container_count{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	containerCount, err := bcsmonitor.QueryValue(c.Context, c.ProjectId, containerCountPromQL, params, time.Now())
	if err != nil {
		return nil, err
	}
	overview.ContainerCount = containerCount

	podCountPromQL := `bcs:node:pod_count{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	podCount, err := bcsmonitor.QueryValue(c.Context, c.ProjectId, podCountPromQL, params, time.Now())
	if err != nil {
		return nil, err
	}
	overview.PodCount = podCount

	return overview, nil
}

// GetNodeCPUUsage 查询 CPU 使用率
func GetNodeCPUUsage(c *rest.Context) (interface{}, error) {
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

	promql := `bcs:node:cpu:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	vector, _, err := bcsmonitor.QueryRangeF(c.Context, c.ProjectId, promql, params, queryTime.Start, queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return vector, nil
}

// GetNodeMemoryUsage 节点内存使用率
func GetNodeMemoryUsage(c *rest.Context) (interface{}, error) {
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

	promql := `bcs:node:memory:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	vector, _, err := bcsmonitor.QueryRangeF(c.Context, c.ProjectId, promql, params, queryTime.Start, queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return vector, nil
}

// GetNodeNetworkTransmitUsage 节点网络发送
func GetNodeNetworkTransmitUsage(c *rest.Context) (interface{}, error) {
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

	promql := `bcs:node:network_transmit{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	vector, _, err := bcsmonitor.QueryRangeF(c.Context, c.ProjectId, promql, params, queryTime.Start, queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return vector, nil
}

// GetNodeNetworkReceiveUsage 节点网络接收
func GetNodeNetworkReceiveUsage(c *rest.Context) (interface{}, error) {
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

	promql := `bcs:node:network_receive{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	vector, _, err := bcsmonitor.QueryRangeF(c.Context, c.ProjectId, promql, params, queryTime.Start, queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return vector, nil
}

// GetNodeDiskioUsage 节点磁盘IO
func GetNodeDiskioUsage(c *rest.Context) (interface{}, error) {
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

	promql := `bcs:node:diskio:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}`
	vector, _, err := bcsmonitor.QueryRangeF(c.Context, c.ProjectId, promql, params, queryTime.Start, queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return vector, nil
}
