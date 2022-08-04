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

package bkmonitor

import (
	"context"
	"fmt"
	"time"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/prometheus/prometheus/prompb"
)

const (
	IGNORE_DEVICE_TYPE = "iso9660|tmpfs|udf" // 磁盘统计 忽略的设备类型, 数据来源蓝鲸监控主机查询规则
	DISK_MOUNTPOINT    = "/data"             // 数据盘目录
)

// BKMonitor :
type BKMonitor struct{}

// NewBKMonitor :
func NewBKMonitor() *BKMonitor {
	return &BKMonitor{}
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (m *BKMonitor) GetClusterCPUTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}
	cpuTotalPromQuery := fmt.Sprintf(`count(bkmonitor:system:cpu_detail:usage{cluster_id="%s", ip=~"%s", provider="BK_MONITOR"})`,
		clusterId,
		nodeMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, cpuTotalPromQuery, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUUsed 获取CPU核心使用量
func (m *BKMonitor) GetClusterCPUUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}

	cpuTotalPromQuery := fmt.Sprintf(`sum(bkmonitor:system:cpu_detail:usage{cluster_id="%s", ip=~"%s", provider="BK_MONITOR"}) / 100`, clusterId, nodeMatch)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, cpuTotalPromQuery, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUUsed 获取CPU核心使用率
func (m *BKMonitor) GetClusterCPUUsage(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}

	cpuTotalPromQuery := fmt.Sprintf(`sum(bkmonitor:system:cpu_detail:usage{cluster_id="%s", ip=~"%s", provider="BK_MONITOR"}) / 100`, clusterId, nodeMatch)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, cpuTotalPromQuery, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (m *BKMonitor) GetClusterMemoryTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}
	cpuTotalPromQuery := fmt.Sprintf(`sum(bkmonitor:system:mem:total{cluster_id="%s", ip=~"%s", provider="BK_MONITOR"})`,
		clusterId,
		nodeMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, cpuTotalPromQuery, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterMemoryUsed 获取集群内存使用量
func (m *BKMonitor) GetClusterMemoryUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}

	cpuTotalPromQuery := fmt.Sprintf(`sum(bkmonitor:system:mem:used{cluster_id="%s", ip=~"%s", provider="BK_MONITOR"})`, clusterId, nodeMatch)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, cpuTotalPromQuery, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterMemoryUsage 获取内存使用率
func (m *BKMonitor) GetClusterMemoryUsage(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`(sum(bkmonitor:system:mem:used{cluster_id="%s", ip=~"%s", provider="BK_MONITOR"}) / sum(bkmonitor:system:mem:total{cluster_id="%s", ip=~"%s", provider="BK_MONITOR"})) * 100`,
		clusterId,
		nodeMatch,
		clusterId,
		nodeMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterDiskTotal 集群磁盘总量
func (m *BKMonitor) GetClusterDiskTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}
	promql := fmt.Sprintf(`sum(bkmonitor:system:disk:total{cluster_id="%s", ip=~"%s", device_type!~"%s", provider="BK_MONITOR"})`,
		clusterId,
		nodeMatch,
		IGNORE_DEVICE_TYPE,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterDiskUsed 集群磁盘使用
func (m *BKMonitor) GetClusterDiskUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(`sum(bkmonitor:system:disk:used{cluster_id="%s", ip=~"%s", device_type!~"%s", provider="BK_MONITOR"})`,
		clusterId,
		nodeMatch,
		IGNORE_DEVICE_TYPE,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterDiskUsage 集群磁盘使用率
func (m *BKMonitor) GetClusterDiskUsage(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, err := base.GetNodeMatch(ctx, clusterId, false)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(bkmonitor:system:disk:used{cluster_id="%s", ip=~"%s", device_type!~"%s", provider="BK_MONITOR"}) /
		sum(bkmonitor:system:disk:total{cluster_id="%s",  ip=~"%s", device_type!~"%s", provider="BK_MONITOR"}) * 100`,
		clusterId,
		nodeMatch,
		IGNORE_DEVICE_TYPE,
		clusterId,
		nodeMatch,
		IGNORE_DEVICE_TYPE,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}
