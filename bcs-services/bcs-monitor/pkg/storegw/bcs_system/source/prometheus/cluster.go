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

package prometheus

import (
	"context"
	"fmt"
	"time"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/prometheus/prometheus/prompb"
)

const (
	DISK_FSTYPE     = "ext[234]|btrfs|xfs|zfs" // 磁盘统计 允许的文件系统
	DISK_MOUNTPOINT = "/|/data"                // 磁盘统计 允许的挂载目录
)

// Prometheus
type Prometheus struct {
}

// NewPrometheus
func NewPrometheus() *Prometheus {
	return &Prometheus{}
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (p *Prometheus) GetClusterCPUTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%s", job="node-exporter", mode="idle", instance=~"%s"}))`,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUUsed 获取CPU核心使用量
func (p *Prometheus) GetClusterCPUUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(irate(node_cpu_seconds_total{cluster_id="%s", job="node-exporter", mode!="idle", instance=~"%s"}[2m]))`,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUUsage 获取CPU核心使用量
func (p *Prometheus) GetClusterCPUUsage(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(irate(node_cpu_seconds_total{cluster_id="%s", job="node-exporter", mode!="idle", instance=~"%s"}[2m])) /
        sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%s", job="node-exporter", mode="idle", instance=~"%s"})) *
        100`,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterMemoryTotal 获取集群内存总量
func (p *Prometheus) GetClusterMemoryTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(node_memory_MemTotal_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"})`,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterMemoryUsed 获取内存使用量
func (p *Prometheus) GetClusterMemoryUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`(sum(node_memory_MemTotal_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) -
        sum(node_memory_MemFree_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) -
        sum(node_memory_Buffers_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) -
        sum(node_memory_Cached_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) +
        sum(node_memory_Shmem_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}))`,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterMemoryUsage 获取内存使用率
func (p *Prometheus) GetClusterMemoryUsage(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`(sum(node_memory_MemTotal_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) -
        sum(node_memory_MemFree_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) -
        sum(node_memory_Buffers_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) -
        sum(node_memory_Cached_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) +
        sum(node_memory_Shmem_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"})) /
        sum(node_memory_MemTotal_bytes{cluster_id="%s", job="node-exporter", instance=~"%s"}) *
        100`,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
		clusterId,
		instanceMatch,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterDiskTotal 获取集群磁盘总量
func (p *Prometheus) GetClusterDiskTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(node_filesystem_size_bytes{cluster_id="%s", job="node-exporter", instance=~"%s", fstype=~"%s", mountpoint=~"%s"})`,
		clusterId,
		instanceMatch,
		DISK_FSTYPE,
		DISK_MOUNTPOINT,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterDiskUsed 获取集群磁盘使用量
func (p *Prometheus) GetClusterDiskUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`sum(node_filesystem_size_bytes{cluster_id="%s", job="node-exporter", instance=~"%s", fstype=~"%s", mountpoint=~"%s"}) -
        sum(node_filesystem_free_bytes{cluster_id="%s", job="node-exporter", instance=~"%s", fstype=~"%s", mountpoint=~"%s"})`,
		clusterId,
		instanceMatch,
		DISK_FSTYPE,
		DISK_MOUNTPOINT,
		clusterId,
		instanceMatch,
		DISK_FSTYPE,
		DISK_MOUNTPOINT,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUUsed 获取CPU核心使用量
func (p *Prometheus) GetClusterDiskUsage(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	instanceMatch, err := base.GetNodeMatch(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	promql := fmt.Sprintf(
		`(sum(node_filesystem_size_bytes{cluster_id="%s", job="node-exporter", instance=~"%s", fstype=~"%s", mountpoint=~"%s"}) -
		sum(node_filesystem_free_bytes{cluster_id="%s", job="node-exporter", instance=~"%s", fstype=~"%s", mountpoint=~"%s"})) /
        sum(node_filesystem_size_bytes{cluster_id="%s", job="node-exporter", instance=~"%s", fstype=~"%s", mountpoint=~"%s"}) *
        100`,
		clusterId,
		instanceMatch,
		DISK_FSTYPE,
		DISK_MOUNTPOINT,
		clusterId,
		instanceMatch,
		DISK_FSTYPE,
		DISK_MOUNTPOINT,
		clusterId,
		instanceMatch,
		DISK_FSTYPE,
		DISK_MOUNTPOINT,
	)

	matrix, _, err := bcsmonitor.QueryRange(ctx, projectId, promql, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}
