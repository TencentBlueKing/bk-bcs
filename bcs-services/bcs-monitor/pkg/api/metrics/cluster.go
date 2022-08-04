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
	"fmt"
	"time"

	"github.com/prometheus/common/model"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// Usage 使用量
type Usage struct {
	Used  string `json:"used"`
	Total string `json:"total"`
}

// UsageByte 使用量, bytes单位
type UsageByte struct {
	UsedByte  string `json:"used_bytes"`
	TotalByte string `json:"total_bytes"`
}

// ClusterOverviewMetric 集群概览接口
type ClusterOverviewMetric struct {
	CPUUsage    *Usage     `json:"cpu_usage"`
	DiskUsage   *UsageByte `json:"disk_usage"`
	MemoryUsage *UsageByte `json:"memory_usage"`
}

// GetFirstValue 获取第一个值
func GetFirstValue(vector model.Vector) string {
	if len(vector) == 0 {
		return "0"
	}
	return vector[0].Value.String()
}

// GetClusterOverview 集群概览数据
func GetClusterOverview(c *rest.Context) (interface{}, error) {
	promqlUsed := fmt.Sprintf(`bcs:cluster:cpu:used{cluster_id="%s"}`, c.ClusterId)
	usedVector, _, err := bcsmonitor.QueryInstant(c.Context, c.ProjectId, promqlUsed, time.Now())
	if err != nil {
		return nil, err
	}

	promqlTotal := fmt.Sprintf(`bcs:cluster:cpu:total{cluster_id="%s"}`, c.ClusterId)
	totalVector, _, err := bcsmonitor.QueryInstant(c.Context, c.ProjectId, promqlTotal, time.Now())
	if err != nil {
		return nil, err
	}

	memoryTotal := fmt.Sprintf(`bcs:cluster:memory:total{cluster_id="%s"}`, c.ClusterId)
	memoryTotalVector, _, err := bcsmonitor.QueryInstant(c.Context, c.ProjectId, memoryTotal, time.Now())
	if err != nil {
		return nil, err
	}

	memoryUsed := fmt.Sprintf(`bcs:cluster:memory:used{cluster_id="%s"}`, c.ClusterId)
	memoryUsedVector, _, err := bcsmonitor.QueryInstant(c.Context, c.ProjectId, memoryUsed, time.Now())
	if err != nil {
		return nil, err
	}

	diskTotal := fmt.Sprintf(`bcs:cluster:disk:total{cluster_id="%s"}`, c.ClusterId)
	diskTotalVector, _, err := bcsmonitor.QueryInstant(c.Context, c.ProjectId, diskTotal, time.Now())
	if err != nil {
		return nil, err
	}

	diskUsed := fmt.Sprintf(`bcs:cluster:disk:used{cluster_id="%s"}`, c.ClusterId)
	diskUsedVector, _, err := bcsmonitor.QueryInstant(c.Context, c.ProjectId, diskUsed, time.Now())
	if err != nil {
		return nil, err
	}

	m := ClusterOverviewMetric{
		CPUUsage: &Usage{
			Used:  GetFirstValue(usedVector),
			Total: GetFirstValue(totalVector),
		},
		MemoryUsage: &UsageByte{
			UsedByte:  GetFirstValue(memoryUsedVector),
			TotalByte: GetFirstValue(memoryTotalVector),
		},
		DiskUsage: &UsageByte{
			UsedByte:  GetFirstValue(diskUsedVector),
			TotalByte: GetFirstValue(diskTotalVector),
		},
	}

	return m, nil
}

// ClusterCPUUsage 集群 CPU 使用率
func ClusterCPUUsage(c *rest.Context) (interface{}, error) {
	promql := fmt.Sprintf(`bcs:cluster:cpu:usage{cluster_id="%s"}`, c.ClusterId)
	end := time.Now()
	start := end.Add(-time.Hour)
	vector, _, err := bcsmonitor.QueryRange(c.Context, c.ProjectId, promql, start, end, time.Minute)
	if err != nil {
		return nil, err
	}
	return vector, nil
}

// ClusterMemoryUsage 集群 内存 使用率
func ClusterMemoryUsage(c *rest.Context) (interface{}, error) {
	promql := fmt.Sprintf(`bcs:cluster:memory:usage{cluster_id="%s"}`, c.ClusterId)
	end := time.Now()
	start := end.Add(-time.Hour)
	vector, _, err := bcsmonitor.QueryRange(c.Context, c.ProjectId, promql, start, end, time.Minute)
	if err != nil {
		return nil, err
	}
	return vector, nil
}

// ClusterDiskUsage 集群磁盘使用率
func ClusterDiskUsage(c *rest.Context) (interface{}, error) {
	promql := fmt.Sprintf(`bcs:cluster:disk:usage{cluster_id="%s"}`, c.ClusterId)
	end := time.Now()
	start := end.Add(-time.Hour)
	vector, _, err := bcsmonitor.QueryRange(c.Context, c.ProjectId, promql, start, end, time.Minute)
	if err != nil {
		return nil, err
	}
	return vector, nil
}
