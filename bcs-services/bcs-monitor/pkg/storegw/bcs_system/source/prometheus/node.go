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

// handleNodeMetric xxx
func (m *Prometheus) handleNodeMetric(ctx context.Context, projectId, clusterId, nodeName string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, _, err := base.GetNodeMatchByName(ctx, clusterId, nodeName)
	if err != nil {
		return nil, err
	}
	params := map[string]interface{}{
		"clusterId":  clusterId,
		"ip":         nodeMatch,
		"node":       nodeName,
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

// GetNodeInfo 节点信息
func (m *Prometheus) GetNodeInfo(ctx context.Context, projectId, clusterId, nodeName string, t time.Time) (
	*base.NodeInfo, error) {
	nodeMatch, ips, err := base.GetNodeMatchByName(ctx, clusterId, nodeName)
	if err != nil {
		return nil, err
	}
	params := map[string]interface{}{
		"clusterId":  clusterId,
		"ip":         nodeMatch,
		"node":       nodeName,
		"fstype":     DISK_FSTYPE,
		"mountpoint": DISK_MOUNTPOINT,
		"provider":   PROVIDER,
	}

	info := &base.NodeInfo{}

	// 节点信息
	infoPromql := `cadvisor_version_info{cluster_id="%<clusterId>s", instance=~"%<ip>s", %<provider>s}`
	infoLabelSet, err := bcsmonitor.QueryLabelSet(ctx, projectId, infoPromql, params, t)
	if err != nil {
		return nil, err
	}
	info.DockerVersion = infoLabelSet["dockerVersion"]
	info.Release = infoLabelSet["kernelVersion"]
	info.Sysname = infoLabelSet["osVersion"]
	info.Provider = provider
	info.IP = ips

	promqlMap := map[string]string{
		"coreCount": `sum by (instance) (count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s",` +
			` job="node-exporter", mode="idle", instance=~"%<ip>s", %<provider>s}))`,
		"memory": `sum by (instance) (node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<ip>s", %<provider>s})`,
		"disk": `sum by (instance) (node_filesystem_size_bytes{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<ip>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})`,
	}

	result, err := bcsmonitor.QueryMultiValues(ctx, projectId, promqlMap, params, time.Now())
	if err != nil {
		return nil, err
	}

	info.CPUCount = result["coreCount"]
	info.Memory = result["memory"]
	info.Disk = result["disk"]

	return info, nil
}

// GetNodeCPUUsage 节点CPU使用率
func (m *Prometheus) GetNodeCPUUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		sum(irate(node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", mode!="idle", ` +
		`instance=~"%<ip>s", %<provider>s}[2m])) /
		sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", ` +
		`mode="idle", instance=~"%<ip>s", %<provider>s})) *
		100`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeCPURequestUsage 节点CPU装箱率
func (m *Prometheus) GetNodeCPURequestUsage(ctx context.Context, projectId, clusterId, nodeName string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		sum(kube_pod_container_resource_requests_cpu_cores{cluster_id="%<clusterId>s", job="kube-state-metrics", ` +
		`node="%<node>s", %<provider>s}) /
		sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", ` +
		`mode="idle", instance=~"%<ip>s", %<provider>s})) *
		100`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeMemoryUsage 节点内存使用率
func (m *Prometheus) GetNodeMemoryUsage(ctx context.Context, projectId, clusterId, nodeName string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		(sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<ip>s", ` +
		`%<provider>s}) -
        sum(node_memory_MemFree_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<ip>s", ` +
		`%<provider>s}) -
        sum(node_memory_Buffers_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<ip>s", ` +
		`%<provider>s}) -
        sum(node_memory_Cached_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<ip>s", ` +
		`%<provider>s}) +
        sum(node_memory_Shmem_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<ip>s", ` +
		`%<provider>s})) /
        sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<ip>s", ` +
		`%<provider>s}) *
        100`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeMemoryRequestUsage 节点内存装箱率
func (m *Prometheus) GetNodeMemoryRequestUsage(ctx context.Context, projectId, clusterId, nodeName string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		sum(kube_pod_container_resource_requests_memory_bytes{cluster_id="%<clusterId>s", job="kube-state-metrics", ` +
		`node="%<node>s", %<provider>s}) /
		sum(node_memory_MemTotal_bytes{cluster_id="%<clusterId>s", job="node-exporter", instance=~"%<ip>s", ` +
		`%<provider>s}) * 100`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeDiskUsage 节点磁盘使用率
func (m *Prometheus) GetNodeDiskUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		(sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", instance=~"%<ip>s", job="node-exporter", ` +
		`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) -
        sum(node_filesystem_free_bytes{cluster_id="%<clusterId>s", instance=~"%<ip>s", job="node-exporter", ` +
		`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})) /
        sum(node_filesystem_size_bytes{cluster_id="%<clusterId>s", instance=~"%<ip>s", job="node-exporter", ` +
		`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) *
        100`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeDiskioUsage 接触磁盘IO使用率
func (m *Prometheus) GetNodeDiskioUsage(ctx context.Context, projectId, clusterId, nodeName string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max(rate(node_disk_io_time_seconds_total{cluster_id="%<clusterId>s", job="node-exporter", ` +
			`instance=~"%<ip>s", %<provider>s}[2m]) * 100)`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodePodCount PodCount
func (m *Prometheus) GetNodePodCount(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max by (instance) (kubelet_running_pod_count{cluster_id="%<clusterId>s", instance=~"%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeContainerCount 容器数量
func (m *Prometheus) GetNodeContainerCount(ctx context.Context, projectId, clusterId, nodeName string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max by (instance) (kubelet_running_container_count{cluster_id="%<clusterId>s", ` +
			`container_state!="exited|created|unknown", instance=~"%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeNetworkTransmit 网络发送量
func (m *Prometheus) GetNodeNetworkTransmit(ctx context.Context, projectId, clusterId, nodeName string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max(rate(node_network_transmit_bytes_total{cluster_id="%<clusterId>s", instance=~"%<ip>s", ` +
			`job="node-exporter", %<provider>s}[5m]))`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}

// GetNodeNetworkReceive 节点网络接收量
func (m *Prometheus) GetNodeNetworkReceive(ctx context.Context, projectId, clusterId, nodeName string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max(rate(node_network_receive_bytes_total{cluster_id="%<clusterId>s", instance=~"%<ip>s", ` +
			`job="node-exporter", %<provider>s}[5m]))`

	return m.handleNodeMetric(ctx, projectId, clusterId, nodeName, promql, start, end, step)
}
