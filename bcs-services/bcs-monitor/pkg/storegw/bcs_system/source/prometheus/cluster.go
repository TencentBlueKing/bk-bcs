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
	"time"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/prometheus/prometheus/prompb"
)

const (
	// DISK_FSTYPE xxx
	DISK_FSTYPE = "ext[234]|btrfs|xfs|zfs" // 磁盘统计 允许的文件系统
	// DISK_MOUNTPOINT xxx
	DISK_MOUNTPOINT = "/|/data" // 磁盘统计 允许的挂载目录
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
func (m *Prometheus) handleClusterMetric(ctx context.Context, projectId, clusterId string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, nodeNameMatch, err := base.GetNodeMatch(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"clusterId":  clusterId,
		"instance":   nodeMatch,
		"node":       nodeNameMatch,
		"fstype":     DISK_FSTYPE,
		"mountpoint": DISK_MOUNTPOINT,
		"provider":   PROVIDER,
	}

	matrix, _, err := bcsmonitor.QueryRangeMatrix(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (p *Prometheus) GetClusterCPUTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`mode="idle", instance=~"%<instance>s", %<provider>s}))`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterCPUUsed 获取CPU核心使用量
func (p *Prometheus) GetClusterCPUUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(irate(node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", mode!="idle", ` +
			`instance=~"%<instance>s", %<provider>s}[2m]))`
	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterPodUsed 获取集群pod使用量
func (p *Prometheus) GetClusterPodUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 获取pod使用率
	promql :=
		`sum (kubelet_running_pods{cluster_id="%<clusterId>s", %<provider>s})`
	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterPodTotal 获取集群最大允许pod数
func (p *Prometheus) GetClusterPodTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPUUsage 获取CPU核心使用量
func (p *Prometheus) GetClusterCPUUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(irate(node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", mode!="idle", ` +
			`instance=~"%<instance>s", %<provider>s}[2m])) /
        sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", 
		` + `mode="idle", instance=~"%<instance>s", %<provider>s})) *
        100`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterCPURequest 获取CPU Request使用量
func (p *Prometheus) GetClusterCPURequest(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_cpu_cores{cluster_id="%<clusterId>s", job="kube-state-metrics", ` +
			`node=~"%<node>s", %<provider>s})`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterCPURequestUsage 获取CPU核心装箱率
func (p *Prometheus) GetClusterCPURequestUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_cpu_cores{cluster_id="%<clusterId>s", job="kube-state-metrics", ` +
			`node=~"%<node>s", %<provider>s}) /
		sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`mode="idle", instance=~"%<instance>s", %<provider>s})) *
		100`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterMemoryTotal 获取集群内存总量
func (p *Prometheus) GetClusterMemoryTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<instance>s", ` +
			`%<provider>s})`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterMemoryUsed 获取内存使用量
func (p *Prometheus) GetClusterMemoryUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`(sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) -
        sum(node_memory_MemFree_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) -
        sum(node_memory_Buffers_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) -
        sum(node_memory_Cached_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) +
        sum(node_memory_Shmem_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}))`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterMemoryUsage 获取内存使用率
func (p *Prometheus) GetClusterMemoryUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`(sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) -
        sum(node_memory_MemFree_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) -
        sum(node_memory_Buffers_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) -
        sum(node_memory_Cached_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) +
        sum(node_memory_Shmem_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s})) /
        sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) *
        100`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterMemoryRequest 获取内存 Request
func (p *Prometheus) GetClusterMemoryRequest(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_memory_bytes{cluster_id="%<clusterId>s", ` +
			`job="kube-state-metrics", node=~"%<node>s", %<provider>s})`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterMemoryRequestUsage 获取内存装箱率
func (p *Prometheus) GetClusterMemoryRequestUsage(ctx context.Context, projectId, clusterId string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(kube_pod_container_resource_requests_memory_bytes{cluster_id="%<clusterId>s", ` +
			`job="kube-state-metrics", node=~"%<node>s", %<provider>s}) /
		sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<instance>s", %<provider>s}) * 100`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterDiskTotal 获取集群磁盘总量
func (p *Prometheus) GetClusterDiskTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterDiskUsed 获取集群磁盘使用量
func (p *Prometheus) GetClusterDiskUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) -
        sum(node_filesystem_free_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterDiskUsage 获取CPU核心使用量
func (p *Prometheus) GetClusterDiskUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`(sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) -
		sum(node_filesystem_free_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})) /
        sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<instance>s", ` +
			`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) *
        100`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterDiskioUsage 集群磁盘IO使用率
func (p *Prometheus) GetClusterDiskioUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(max by(instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterId>s", ` +
			`job="node-exporter", instance=~"%<instance>s", %<provider>s}[2m]))) /
		count(max by(instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterId>s", ` +
			`job="node-exporter", instance=~"%<instance>s", %<provider>s}[2m]))) /
		* 100)`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterDiskioUsed 集群磁盘IO使用量
func (p *Prometheus) GetClusterDiskioUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(max by(bk_instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterId>s", ` +
			`instance=~"%<instance>s", %<provider>s}[2m])))`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}

// GetClusterDiskioTotal 集群磁盘IO
func (p *Prometheus) GetClusterDiskioTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`count(max by(bk_instance) (rate(node_disk_io_time_seconds_total{cluster_id="%<clusterId>s", ` +
			`instance=~"%<instance>s", %<provider>s}[2m])))`

	return p.handleClusterMetric(ctx, projectId, clusterId, promql, start, end, step)
}
