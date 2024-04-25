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

// Package prometheus prometheus
package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/prompb"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
)

const (
	// DiskFstype xxx
	DiskFstype = "ext[234]|btrfs|xfs|zfs" // 磁盘统计 允许的文件系统
	// DiskMountpoint xxx
	DiskMountpoint = "/data" // 磁盘统计 允许的挂载目录
	// PROVIDER xxx 查询限制
	PROVIDER = `prometheus=~"thanos/po-kube-prometheus-stack-prometheus|thanos/po-prometheus-operator-prometheus"`
)

// Prometheus xxx
type Prometheus struct {
}

// NewPrometheus xxx
func NewPrometheus() *Prometheus {
	return &Prometheus{}
}

// handleClusterMetric Cluster 处理公共函数
func (m *Prometheus) handleClusterMetric(ctx context.Context, projectID, clusterID string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, nodeNameMatch, ok := base.GetNodeMatchIgnoreErr(ctx, clusterID)
	if !ok {
		return nil, nil
	}

	params := map[string]interface{}{
		"clusterID":  clusterID,
		"instance":   nodeMatch,
		"node":       nodeNameMatch,
		"fstype":     DiskFstype,
		"mountpoint": DiskMountpoint,
		"provider":   PROVIDER,
	}

	matrix, _, err := bcsmonitor.QueryRangeMatrix(ctx, projectID, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (m *Prometheus) GetClusterCPUTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterID>s", ` +
			`job="node-exporter", mode="idle", instance=~"%<instance>s", %<provider>s}))` // nolint

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPUUsed 获取CPU核心使用量
func (m *Prometheus) GetClusterCPUUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(irate(node_cpu_seconds_total{cluster_id="%<clusterID>s", job="node-exporter", mode!="idle", ` + // nolint
			`instance=~"%<instance>s", %<provider>s}[2m]))`
	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterPodUsed 获取集群pod使用量
func (m *Prometheus) GetClusterPodUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 获取pod使用率
	promql :=
		`sum (kubelet_running_pod_count{cluster_id="%<clusterID>s", instance=~"%<instance>s", %<provider>s})`
	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterPodTotal 获取集群最大允许pod数
func (m *Prometheus) GetClusterPodTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 获取集群中最大可用pod数
	nodes, err := k8sclient.GetClusterNodeList(ctx, clusterID, true)
	if err != nil {
		return nil, err
	}
	var pod int64
	for _, node := range nodes {
		pod += node.Status.Allocatable.Pods().Value()
	}
	return base.GetSameSeries(start, end, step, float64(pod), nil), nil
}

// GetClusterPodUsage 获取集群pod使用率
func (m *Prometheus) GetClusterPodUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	usedSeries, err := m.GetClusterPodUsed(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}

	nodes, err := k8sclient.GetClusterNodeList(ctx, clusterID, true)
	if err != nil {
		return nil, err
	}
	var pod int64
	for _, node := range nodes {
		pod += node.Status.Allocatable.Pods().Value()
	}

	return base.DivideSeriesByValue(usedSeries, float64(pod)), nil
}

// GetClusterCPUUsage 获取CPU核心使用量
func (m *Prometheus) GetClusterCPUUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(irate(node_cpu_seconds_total{cluster_id="%<clusterID>s", job="node-exporter", mode!="idle", ` +
			`instance=~"%<instance>s", %<provider>s}[2m])) / sum(count without(cpu, mode) ` +
			`(node_cpu_seconds_total{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`mode="idle", instance=~"%<instance>s", %<provider>s})) * 100`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPURequest 获取CPU Request使用量
func (m *Prometheus) GetClusterCPURequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_cpu_cores{cluster_id="%<clusterID>s", job="kube-state-metrics", ` +
			`node=~"%<node>s", %<provider>s})`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPURequestUsage 获取CPU核心装箱率
func (m *Prometheus) GetClusterCPURequestUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_cpu_cores{cluster_id="%<clusterID>s", job="kube-state-metrics", ` +
			`node=~"%<node>s", %<provider>s}) / sum(count without(cpu, mode) ` +
			`(node_cpu_seconds_total{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`mode="idle", instance=~"%<instance>s", %<provider>s})) * 100`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryTotal 获取集群内存总量
func (m *Prometheus) GetClusterMemoryTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_memory_MemTotal_bytes{cluster_id="%<clusterID>s", job="node-exporter", instance=~"%<instance>s", ` +
			`%<provider>s})`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryUsed 获取内存使用量
func (m *Prometheus) GetClusterMemoryUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`(sum(node_memory_MemTotal_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_MemFree_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_Buffers_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) - ` + // nolint
			`sum(node_memory_Cached_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) + ` +
			`sum(node_memory_Shmem_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryUsage 获取内存使用率
func (m *Prometheus) GetClusterMemoryUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`(sum(node_memory_MemTotal_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_MemFree_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_Buffers_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_Cached_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) + ` +
			`sum(node_memory_Shmem_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s})) / ` +
			`sum(node_memory_MemTotal_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) * 100`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryRequest 获取内存 Request
func (m *Prometheus) GetClusterMemoryRequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_memory_bytes{cluster_id="%<clusterID>s", ` + // nolint
			`job="kube-state-metrics", node=~"%<node>s", %<provider>s})`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryRequestUsage 获取内存装箱率
func (m *Prometheus) GetClusterMemoryRequestUsage(ctx context.Context, projectID, clusterID string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_memory_bytes{cluster_id="%<clusterID>s", ` +
			`job="kube-state-metrics", node=~"%<node>s", %<provider>s}) / ` +
			`sum(node_memory_MemTotal_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) * 100`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskTotal 获取集群磁盘总量
func (m *Prometheus) GetClusterDiskTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})` // nolint

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskUsed 获取集群磁盘使用量
func (m *Prometheus) GetClusterDiskUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) - ` + // nolint
			`sum(node_filesystem_free_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskUsage 获取CPU核心使用量
func (m *Prometheus) GetClusterDiskUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`(sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) - ` +
			`sum(node_filesystem_free_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})) / ` +
			`sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", job="node-exporter", ` +
			`instance=~"%<instance>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) * 100`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskioUsage 集群磁盘IO使用率
func (m *Prometheus) GetClusterDiskioUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(max by(instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterID>s", ` +
			`job="node-exporter", instance=~"%<instance>s", %<provider>s}[2m]))) / ` +
			`count(max by(instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterID>s", ` +
			`job="node-exporter", instance=~"%<instance>s", %<provider>s}[2m]))) * 100)`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskioUsed 集群磁盘IO使用量
func (m *Prometheus) GetClusterDiskioUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(max by(instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterID>s", ` +
			`instance=~"%<instance>s", %<provider>s}[2m])))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskioTotal 集群磁盘IO
func (m *Prometheus) GetClusterDiskioTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`count(max by(instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterID>s", ` +
			`instance=~"%<instance>s", %<provider>s}[2m])))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}
