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

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/prometheus/prometheus/prompb"
)

// 和现在规范的名称保持一致
const provider = "BK-Monitor"

// handleNodeMetric xxx
func (m *BKMonitor) handleNodeMetric(ctx context.Context, projectId, clusterId, ip string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId":  clusterId,
		"ip":         ip,
		"deviceType": IGNORE_DEVICE_TYPE,
		"provider":   PROVIDER,
	}

	matrix, _, err := bcsmonitor.QueryRangeMatrix(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeInfo 节点信息
func (m *BKMonitor) GetNodeInfo(ctx context.Context, projectId, clusterId, ip string, t time.Time) (*base.NodeInfo,
	error) {
	params := map[string]interface{}{
		"clusterId":  clusterId,
		"ip":         ip,
		"deviceType": IGNORE_DEVICE_TYPE,
		"provider":   PROVIDER,
	}

	info := &base.NodeInfo{}

	// 节点信息
	infoPromql := `cadvisor_version_info{cluster_id="%<clusterId>s", bk_instance=~"%<ip>s:.*", %<provider>s}`
	infoLabelSet, err := bcsmonitor.QueryLabelSet(ctx, projectId, infoPromql, params, t)
	if err != nil {
		return nil, err
	}
	info.DockerVersion = infoLabelSet["dockerVersion"]
	info.Release = infoLabelSet["kernelVersion"]
	info.Sysname = infoLabelSet["osVersion"]
	info.Provider = provider

	promqlMap := map[string]string{
		"coreCount": `count(bkmonitor:system:cpu_detail:usage{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s})`,
		"memory":    `sum(bkmonitor:system:mem:total{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s})`,
		"disk":      `sum(bkmonitor:system:disk:total{cluster_id="%<clusterId>s", ip="%<ip>s", device_type!~"%<deviceType>s", %<provider>s})`,
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
func (m *BKMonitor) GetNodeCPUUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(bkmonitor:system:cpu_detail:usage{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s}) / count(bkmonitor:system:cpu_detail:usage{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}

// GetNodeMemoryUsage 内存使用率
func (m *BKMonitor) GetNodeMemoryUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		(sum(bkmonitor:system:mem:used{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s}) /
		sum(bkmonitor:system:mem:total{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s})) *
		100`

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}

// GetNodeDiskUsage 节点磁盘使用率
func (m *BKMonitor) GetNodeDiskUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		(sum(bkmonitor:system:disk:used{cluster_id="%<clusterId>s", ip="%<ip>s", device_type!~"%<deviceType>s", %<provider>s}) /
		sum(bkmonitor:system:disk:total{cluster_id="%<clusterId>s", ip="%<ip>s", device_type!~"%<deviceType>s", %<provider>s})) *
		100`

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}

// GetNodeDiskioUsage 节点磁盘IO使用率
func (m *BKMonitor) GetNodeDiskioUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `max(bkmonitor:system:io:util{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s}) * 100`

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}

// GetNodePodCount PodCount
func (m *BKMonitor) GetNodePodCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 注意 k8s 1.19 版本以前的 metrics 是 kubelet_running_pod_count
	promql :=
		`max by (bk_instance) (kubelet_running_pods{cluster_id="%<clusterId>s", bk_instance=~"%<ip>s:.*", %<provider>s})`

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}

// GetNodeContainerCount 容器Count
func (m *BKMonitor) GetNodeContainerCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 注意 k8s 1.19 版本以前的 metrics 是 kubelet_running_container_count
	// https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.16.md 添加 running/exited/created/unknown label
	// https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.19.md

	// container_state 常量定义 https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/container/runtime.go#L258
	// 使用不等于, 兼容高低版本
	promql :=
		`max by (bk_instance) (kubelet_running_containers{cluster_id="%<clusterId>s", container_state!="exited|created|unknown", bk_instance=~"%<ip>s:.*", %<provider>s})`

	// 按版本兼容逻辑
	// if k8sclient.K8SLessThan(ctx, clusterId, "v1.19") {
	// 	promql = `max by (bk_instance) (kubelet_running_container_count{cluster_id="%<clusterId>s", bk_instance=~"%<ip>s:.*", %<provider>s})`
	// }

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}

// GetNodeNetworkTransmit 节点网络发送量
func (m *BKMonitor) GetNodeNetworkTransmit(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `max(bkmonitor:system:net:speed_sent{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}

// GetNodeNetworkReceive 节点网络接收
func (m *BKMonitor) GetNodeNetworkReceive(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `max(bkmonitor:system:net:speed_recv{cluster_id="%<clusterId>s", ip="%<ip>s", %<provider>s})`

	return m.handleNodeMetric(ctx, projectId, clusterId, ip, promql, start, end, step)
}
