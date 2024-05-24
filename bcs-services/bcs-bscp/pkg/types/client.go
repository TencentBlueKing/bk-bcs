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

package types

import (
	"time"
)

// ChartType xxx
type ChartType string

const (
	// Sunburst 树状结构
	Sunburst ChartType = "sunburst"
	// Bar 柱状结构
	Bar ChartType = "bar"
)

// ClientConfigVersionChart 客户端配置图表
type ClientConfigVersionChart struct {
	CurrentReleaseID uint32 `json:"current_release_id"`
	Count            int    `json:"count"`
}

// ChangeStatusChart 客户端变更状态图表
type ChangeStatusChart struct {
	ReleaseChangeStatus string `json:"release_change_status"`
	Count               int    `json:"count"`
}

// FailedReasonChart 客户端变更失败原因图表
type FailedReasonChart struct {
	ReleaseChangeFailedReason string `json:"release_change_failed_reason"`
	Count                     int    `json:"count"`
}

// MinMaxAvgTimeChart 拉取平均耗时
type MinMaxAvgTimeChart struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
}

// ResourceUsage 资源使用率
type ResourceUsage struct {
	CpuMaxUsage    float64 `json:"cpu_max_usage"`
	MemoryMaxUsage float64 `json:"memory_max_usage"`
	CpuAvgUsage    float64 `json:"cpu_avg_usage"`
	MemoryAvgUsage float64 `json:"memory_avg_usage"`
	CpuMinUsage    float64 `json:"cpu_min_usage"`
	MemoryMinUsage float64 `json:"memory_min_usage"`
}

// PullTrend 拉取趋势
type PullTrend struct {
	ClientID uint32    `json:"client_id"`
	PullTime time.Time `json:"pull_time"`
	Count    int       `json:"count"`
}

// PrimaryAndForeign 定义一组主键和外键的结构体
type PrimaryAndForeign struct {
	PrimaryKey, ForeignKey string
	PrimaryVal, ForeignVal string
	Count                  int
}
