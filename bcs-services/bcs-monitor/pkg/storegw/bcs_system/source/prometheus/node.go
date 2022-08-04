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

const provider = "Prometheus"

// GetNodeInfo 节点信息
func (m *Prometheus) GetNodeInfo(ctx context.Context, projectId, clusterId, innerIP string, t time.Time) (*base.NodeInfo, error) {
	params := map[string]interface{}{
		"clusterId":  clusterId,
		"innerIP":    innerIP,
		"fstype":     DISK_FSTYPE,
		"mountpoint": DISK_MOUNTPOINT,
	}

	info := &base.NodeInfo{}

	// 节点信息
	infoPromql := `cadvisor_version_info{cluster_id="%<clusterId>s", instance=~"%<innerIP>s:.*"}`
	infoLabelSet, err := bcsmonitor.QueryLabelSet(ctx, projectId, infoPromql, params, t)
	if err != nil {
		return nil, err
	}
	info.DockerVersion = infoLabelSet["dockerVersion"]
	info.Release = infoLabelSet["kernelVersion"]
	info.Sysname = infoLabelSet["osVersion"]
	info.Provider = provider

	coreCountPromQL := `sum by (instance) (count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", mode="idle", instance=~"%<innerIP>s:\\d+"}))`
	coreCount, err := bcsmonitor.QueryValue(ctx, projectId, coreCountPromQL, params, t)
	if err != nil {
		return nil, err
	}
	info.CPUCount = coreCount

	memoryPromQL := `sum by (instance) (node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<innerIP>s:\\d+"})`
	memory, err := bcsmonitor.QueryValue(ctx, projectId, memoryPromQL, params, t)
	if err != nil {
		return nil, err
	}
	info.Memory = memory

	diskPromQL := `sum by (instance) (node_filesystem_size_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<innerIP>s:\\d+", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"})`
	disk, err := bcsmonitor.QueryValue(ctx, projectId, diskPromQL, params, t)
	if err != nil {
		return nil, err
	}
	info.Disk = disk

	return info, nil
}

// GetNodeCPUUsage 节点CPU使用率
func (m *Prometheus) GetNodeCPUUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `
		sum(irate(node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", mode!="idle", instance="%<ip>s:9100"}[2m])) /
		sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", mode="idle", instance="%<ip>s:9100"})) *
		100`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeMemoryUsage 节点内存使用率
func (m *Prometheus) GetNodeMemoryUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `
		(sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance="%<ip>s:9100"}) -
        sum(node_memory_MemFree_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance="%<ip>s:9100"}) -
        sum(node_memory_Buffers_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance="%<ip>s:9100"}) -
        sum(node_memory_Cached_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance="%<ip>s:9100"}) +
        sum(node_memory_Shmem_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance="%<ip>s:9100"})) /
        sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance="%<ip>s:9100"}) *
        100`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeDiskUsage 节点磁盘使用率
func (m *Prometheus) GetNodeDiskUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId":  clusterId,
		"ip":         ip,
		"fstype":     DISK_FSTYPE,
		"mountpoint": DISK_MOUNTPOINT,
	}

	promql := `
		(sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", instance="%<ip>s:9100", job="node-exporter", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}) -
        sum(node_filesystem_free_bytes{cluster_id="%<clusterId>s", instance="%<ip>s:9100", job="node-exporter", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"})) /
        sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", instance="%<ip>s:9100", job="node-exporter", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s"}) *
        100`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeDiskioUsage 接触磁盘IO使用率
func (m *Prometheus) GetNodeDiskioUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max(rate(node_disk_io_time_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", instance="%<ip>s:9100"}[2m]) * 100)`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodePodCount PodCount
func (m *Prometheus) GetNodePodCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max by (instance) (kubelet_running_pod_count{cluster_id="%<clusterId>s", instance=~"%<ip>s:\\d+"})`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeContainerCount 容器数量
func (m *Prometheus) GetNodeContainerCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max by (instance) (kubelet_running_container_count{cluster_id="%<clusterId>s", instance=~"%<ip>s:\\d+"})`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeNetworkTransmit 网络发送量
func (m *Prometheus) GetNodeNetworkTransmit(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max(rate(node_network_transmit_bytes_total{cluster_id="%<clusterId>s", instance="%<ip>s:9100", job="node-exporter"}[5m]))`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeNetworkReceive 节点网络接收量
func (m *Prometheus) GetNodeNetworkReceive(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max(rate(node_network_receive_bytes_total{cluster_id="%<clusterId>s", instance="%<ip>s:9100", job="node-exporter"}[5m]))`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}
