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

// UseByte 使用量, bytes单位
type UseByte struct {
	UsedByte  string `json:"used_bytes"`
	TotalByte string `json:"total_bytes"`
}

// ClusterOverviewMetric 集群概览接口
type ClusterOverviewMetric struct {
	CPUUsage    *Usage   `json:"cpu_usage"`
	DiskUsage   *UseByte `json:"disk_usage"`
	MemoryUsage *UseByte `json:"memory_usage"`
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

	m := ClusterOverviewMetric{
		CPUUsage: &Usage{
			Used:  GetFirstValue(usedVector),
			Total: GetFirstValue(totalVector),
		},
	}

	return m, nil
}
