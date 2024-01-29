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
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
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

// UsageQuery 节点查询
type UsageQuery struct {
	StartAt string `json:"start_at" form:"start_at"` // 必填参数`
	EndAt   string `json:"end_at" form:"end_at"`
}

// parseTime 兼容前端多个格式
func parseTime(rawTime string) (time.Time, error) {
	// 和前端约定, 只支持这种带时区的格式
	formats := []string{
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
		queryTime.End = utils.GetNowQueryTime()
	} else {
		t, err := parseTime(q.EndAt)
		if err != nil {
			return nil, err
		}
		queryTime.End = t.Add(-utils.QueryFallbackTime)
	}

	if q.StartAt == "" {
		queryTime.Start = queryTime.End.Add(-time.Hour)
	} else {
		t, err := parseTime(q.StartAt)
		if err != nil {
			return nil, err
		}
		queryTime.Start = t.Add(-utils.QueryFallbackTime)
	}

	// 默认只返回 60 个点
	queryTime.Step = queryTime.End.Sub(queryTime.Start) / 60

	return queryTime, nil
}

// Config 配置
type Config struct {
	Dispatch []clientutil.DispatchConf `yaml:"dispatch"`
}
