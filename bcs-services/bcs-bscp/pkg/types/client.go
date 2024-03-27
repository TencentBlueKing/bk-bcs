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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
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

// ClientComponentVersionFormat 客户端组件版本分布格式
type ClientComponentVersionFormat interface {
	Format(items []*table.Client) []interface{}
}

// SunburstFormatter 树状结构
type SunburstFormatter struct{}

// Format 格式成树状结构
func (s SunburstFormatter) Format(items []*table.Client) []interface{} {
	var charts []interface{}

	typeCountByVersion := make(map[string]map[string]int)
	for _, v := range items {
		if typeCountByVersion[string(v.Spec.ClientType)] == nil {
			typeCountByVersion[string(v.Spec.ClientType)] = make(map[string]int)
		}
		typeCountByVersion[string(v.Spec.ClientType)][v.Spec.ClientVersion]++
	}

	var totalCount int
	for _, versionCounts := range typeCountByVersion {
		for _, count := range versionCounts {
			totalCount += count
		}
	}
	for category, versionCounts := range typeCountByVersion {
		categoryCount := 0
		var children []interface{}
		for version, count := range versionCounts {
			categoryCount += count
			versionPercent := float64(count) / float64(totalCount) * 100
			children = append(children, map[string]interface{}{"name": version, "value": count, "percent": versionPercent})
		}
		categoryPercent := float64(categoryCount) / float64(totalCount) * 100
		charts = append(charts, map[string]interface{}{"name": category, "value": categoryCount,
			"percent": categoryPercent, "children": children})
	}

	return charts
}

// BarFormatter 柱状结构
type BarFormatter struct{}

// Format 格式成柱状结构
func (b BarFormatter) Format(items []*table.Client) []interface{} {
	var charts []interface{}

	counts := make(map[string]int)
	for _, item := range items {
		counts[string(item.Spec.ClientType)+"_"+item.Spec.ClientVersion]++
	}

	totalCount := 0
	for _, count := range counts {
		totalCount += count
	}

	var data []map[string]interface{}
	for _, item := range items {
		key := string(item.Spec.ClientType) + "_" + item.Spec.ClientVersion
		count := counts[key]
		percent := float64(count) / float64(totalCount) * 100
		data = append(data, map[string]interface{}{
			"client_type":    string(item.Spec.ClientType),
			"client_version": item.Spec.ClientVersion,
			"percent":        percent,
			"value":          count,
		})
	}

	filteredOutputData := make(map[string][]map[string]interface{})
	for _, item := range data {
		key := item["client_type"].(string) + "_" + item["client_version"].(string)
		filteredOutputData[key] = append(filteredOutputData[key], item)
	}

	for _, data := range filteredOutputData {
		charts = append(charts, data[0])
	}

	return charts
}
