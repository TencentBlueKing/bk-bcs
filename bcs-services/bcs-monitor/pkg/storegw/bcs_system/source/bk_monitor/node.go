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
	"time"

	"github.com/prometheus/prometheus/prompb"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
)

// 和现在规范的名称保持一致
const provider = "BK-Monitor"

// handleNodeMetric xxx
func (m *BKMonitor) handleNodeMetric(ctx context.Context, projectID, clusterID, nodeName string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeMatch, _, ok := base.GetNodeMatchByNameIngErr(ctx, clusterID, nodeName)
	if !ok {
		return nil, nil
	}
	params := map[string]interface{}{
		"clusterID":  clusterID,
		"ip":         nodeMatch,
		"node":       nodeName,
		"fstype":     DisFstype,
		"mountpoint": DiskMountPoint,
		"provider":   PROVIDER,
	}

	matrix, _, err := bcsmonitor.QueryRangeMatrix(ctx, projectID, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeInfo 节点信息
func (m *BKMonitor) GetNodeInfo(ctx context.Context, projectID, clusterID, nodeName string, t time.Time) (
	*base.NodeInfo, error) {
	nodeMatch, ips, ok := base.GetNodeMatchByNameIngErr(ctx, clusterID, nodeName)
	if !ok {
		return nil, nil
	}
	params := map[string]interface{}{
		"clusterID":  clusterID,
		"ip":         nodeMatch,
		"node":       nodeName,
		"fstype":     DisFstype,
		"mountpoint": DiskMountPoint,
		"provider":   PROVIDER,
	}

	// 获取容器运行时的版本
	version, ok := base.GetNodeCRVersionByNameIngErr(ctx, clusterID, nodeName)
	if !ok {
		return nil, nil
	}

	info := &base.NodeInfo{}

	// 节点信息
	infoPromql := `cadvisor_version_info{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", %<provider>s}`
	infoLabelSet, err := bcsmonitor.QueryLabelSet(ctx, projectID, infoPromql, params, t)
	if err != nil {
		return nil, err
	}

	info.DockerVersion = infoLabelSet["dockerVersion"]
	info.Release = infoLabelSet["kernelVersion"]
	info.Sysname = infoLabelSet["osVersion"]
	info.Provider = provider
	info.IP = ips
	info.ContainerRuntimeVersion = version

	promqlMap := map[string]string{
		"coreCount": `sum by (instance) (count without(cpu, mode) (node_cpu_seconds_total{cluster_id="%<clusterID>s` +
			`", mode="idle", bk_instance=~"%<ip>s", %<provider>s}))`,
		"memory": `sum by (instance) (node_memory_MemTotal_bytes{cluster_id="%<clusterID>s", ` +
			`bk_instance=~"%<ip>s", %<provider>s})`,
		"disk": `sum by (instance) (node_filesystem_size_bytes{cluster_id="%<clusterID>s", ` +
			`bk_instance=~"%<ip>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})`,
	}

	result, err := bcsmonitor.QueryMultiValues(ctx, projectID, promqlMap, params, time.Now())
	if err != nil {
		return nil, err
	}

	info.CPUCount = result["coreCount"]
	info.Memory = result["memory"]
	info.Disk = result["disk"]

	return info, nil
}

// GetNodeCPUTotal 节点CPU总量
func (m *BKMonitor) GetNodeCPUTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(kube_node_status_allocatable_cpu_cores{cluster_id="%<clusterID>s", node="%<node>s", ` +
		`%<provider>s})` // nolint

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeCPURequest 节点CPU请求量
func (m *BKMonitor) GetNodeCPURequest(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(kube_pod_container_resource_requests_cpu_cores{cluster_id="%<clusterID>s", node="%<node>s", ` +
		`%<provider>s})` // nolint

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeCPUUsed 节点CPU使用量
func (m *BKMonitor) GetNodeCPUUsed(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(irate(node_cpu_seconds_total{cluster_id="%<clusterID>s", mode!="idle", bk_instance=~"%<ip>s", ` +
		`%<provider>s}[2m]))` // nolint

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeCPUUsage 节点CPU使用率
func (m *BKMonitor) GetNodeCPUUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetNodeCPUUsed(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetNodeCPUTotal(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetNodeCPURequestUsage 节点CPU装箱率
func (m *BKMonitor) GetNodeCPURequestUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetNodeCPURequest(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetNodeCPUTotal(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetNodeMemoryTotal 节点Memory总量
func (m *BKMonitor) GetNodeMemoryTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(kube_node_status_allocatable_memory_bytes{cluster_id="%<clusterID>s", node="%<node>s", ` +
		`%<provider>s})`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeMemoryRequest 节点Memory请求量
func (m *BKMonitor) GetNodeMemoryRequest(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(kube_pod_container_resource_requests_memory_bytes{cluster_id="%<clusterID>s", node="%<node>s", ` +
		`%<provider>s})`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeMemoryUsed 节点Memory使用量
func (m *BKMonitor) GetNodeMemoryUsed(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(node_memory_MemTotal_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", %<provider>s}) - ` +
		`sum(node_memory_MemFree_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", %<provider>s}) - ` +
		`sum(node_memory_Buffers_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", %<provider>s}) - ` +
		`sum(node_memory_Cached_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", %<provider>s}) + ` +
		`sum(node_memory_Shmem_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeMemoryUsage 内存使用率
func (m *BKMonitor) GetNodeMemoryUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetNodeMemoryUsed(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetNodeMemoryTotal(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetNodeMemoryRequestUsage 内存装箱率
func (m *BKMonitor) GetNodeMemoryRequestUsage(ctx context.Context, projectID, clusterID, node string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetNodeMemoryRequest(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetNodeMemoryTotal(ctx, projectID, clusterID, node, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetNodeDiskTotal 节点磁盘总量
func (m *BKMonitor) GetNodeDiskTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", fstype=~"%<fstype>s", ` +
		`mountpoint=~"%<mountpoint>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeDiskUsed 节点磁盘使用量
func (m *BKMonitor) GetNodeDiskUsed(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", ` +
		`fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s}) - ` +
		`sum(node_filesystem_free_bytes{cluster_id="%<clusterID>s", ` +
		`bk_instance=~"%<ip>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeDiskUsage 节点磁盘使用率
func (m *BKMonitor) GetNodeDiskUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `(sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", ` +
		`fstype=~"%<fstype>s", ` +
		`mountpoint=~"%<mountpoint>s", %<provider>s}) - sum(node_filesystem_free_bytes{cluster_id="%<clusterID>s", ` +
		`bk_instance=~"%<ip>s", fstype=~"%<fstype>s", mountpoint=~"%<mountpoint>s", %<provider>s})) / ` +
		`sum(node_filesystem_size_bytes{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", fstype=~"%<fstype>s", ` +
		`mountpoint=~"%<mountpoint>s", %<provider>s}) * 100`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeDiskioUsage 节点磁盘IO使用率
func (m *BKMonitor) GetNodeDiskioUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max(rate(node_disk_io_time_seconds_total{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", ` +
			`%<provider>s}[2m]) * 100)`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodePodCount PodCount
func (m *BKMonitor) GetNodePodCount(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 注意 k8s 1.19 版本以前的 metrics 是 kubelet_running_pod_count
	promql :=
		`max by (instance) (kubelet_running_pods{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodePodTotal PodTotal
func (m *BKMonitor) GetNodePodTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 获取集群中最大可用pod数
	nodes, ok := base.GetNodeInfoIngoreErr(ctx, clusterID, node)
	if !ok {
		return nil, nil
	}
	return base.GetSameSeries(start, end, step, float64(nodes.Status.Allocatable.Pods().Value()), nil), nil
}

// GetNodeContainerCount 容器Count
func (m *BKMonitor) GetNodeContainerCount(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 注意 k8s 1.19 版本以前的 metrics 是 kubelet_running_container_count
	// https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.16.md
	// 添加 running/exited/created/unknown label
	// https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.19.md

	// container_state 常量定义
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/container/runtime.go#L258
	// 使用不等于, 兼容高低版本
	promql :=
		`max by (instance) (kubelet_running_containers{cluster_id="%<clusterID>s", ` +
			`container_state!="exited|created|unknown", bk_instance=~"%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeNetworkTransmit 节点网络发送量
func (m *BKMonitor) GetNodeNetworkTransmit(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max(rate(node_network_transmit_bytes_total{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", ` +
			`%<provider>s}[5m]))`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}

// GetNodeNetworkReceive 节点网络接收
func (m *BKMonitor) GetNodeNetworkReceive(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`max(rate(node_network_receive_bytes_total{cluster_id="%<clusterID>s", bk_instance=~"%<ip>s", ` +
			`%<provider>s}[5m]))`

	return m.handleNodeMetric(ctx, projectID, clusterID, node, promql, start, end, step)
}
