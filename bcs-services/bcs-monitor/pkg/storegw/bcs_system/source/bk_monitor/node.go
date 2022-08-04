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

// GetNodeInfo 节点信息
func (m *BKMonitor) GetNodeInfo(ctx context.Context, projectId, clusterId, ip string, t time.Time) (*base.NodeInfo, error) {
	params := map[string]interface{}{
		"clusterId":  clusterId,
		"ip":         ip,
		"deviceType": IGNORE_DEVICE_TYPE,
	}

	info := &base.NodeInfo{}

	// 节点信息
	infoPromql := `cadvisor_version_info{cluster_id="%<clusterId>s", bk_instance=~"%<ip>s:.*", provider="BK_MONITOR"}`
	infoLabelSet, err := bcsmonitor.QueryLabelSet(ctx, projectId, infoPromql, params, t)
	if err != nil {
		return nil, err
	}
	info.DockerVersion = infoLabelSet["dockerVersion"]
	info.Release = infoLabelSet["kernelVersion"]
	info.Sysname = infoLabelSet["osVersion"]
	info.Provider = provider

	// CPU 核心
	coreCountPromQL := `count(bkmonitor:system:cpu_detail:usage{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BK_MONITOR"})`
	coreCount, err := bcsmonitor.QueryValue(ctx, projectId, coreCountPromQL, params, t)
	if err != nil {
		return nil, err
	}
	info.CPUCount = coreCount

	memoryPromQL := `sum(bkmonitor:system:mem:total{cluster_id="%<clusterId>s", ip="%<ip>s", provider="BK_MONITOR"})`
	memory, err := bcsmonitor.QueryValue(ctx, projectId, memoryPromQL, params, t)
	if err != nil {
		return nil, err
	}
	info.Memory = memory

	diskPromQL := `sum(bkmonitor:system:disk:total{cluster_id="%<clusterId>s", ip="%<ip>s", device_type!~"%<deviceType>s", provider="BK_MONITOR"})`
	disk, err := bcsmonitor.QueryValue(ctx, projectId, diskPromQL, params, t)
	if err != nil {
		return nil, err
	}
	info.Disk = disk

	return info, nil
}

// GetNodeCPUUsage 节点CPU使用率
func (m *BKMonitor) GetNodeCPUUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `sum(bkmonitor:system:cpu_detail:usage{cluster_id="%<clusterId>s", ip="%<ip>s"}) / count(bkmonitor:system:cpu_detail:usage{cluster_id="%<clusterId>s", ip="%<ip>s"})`
	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeMemoryUsage 内存使用率
func (m *BKMonitor) GetNodeMemoryUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `
		(sum(bkmonitor:system:mem:used{cluster_id="%<clusterId>s", ip="%<ip>s"}) /
		sum(bkmonitor:system:mem:total{cluster_id="%<clusterId>s", ip="%<ip>s"})) *
		100`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeDiskUsage 节点磁盘使用率
func (m *BKMonitor) GetNodeDiskUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId":   clusterId,
		"ip":          ip,
		"device_type": IGNORE_DEVICE_TYPE,
	}

	promql := `
		(sum(bkmonitor:system:disk:used{cluster_id="%<clusterId>s", ip="%<ip>s", device_type!~"%<device_type>s"}) /
		sum(bkmonitor:system:disk:total{cluster_id="%<clusterId>s", ip="%<ip>s", device_type!~"%<device_type>s"})) *
		100`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeDiskioUsage 节点磁盘IO使用率
func (m *BKMonitor) GetNodeDiskioUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max(bkmonitor:system:io:util{cluster_id="%<clusterId>s", ip="%<ip>s"}) * 100`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodePodCount PodCount
func (m *BKMonitor) GetNodePodCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	// 注意 k8s 1.19 版本以前的 metrics 是 kubelet_running_container_count
	promql := `max by (bk_instance) (kubelet_running_pods{cluster_id="%<clusterId>s", bk_instance=~"%<ip>s:.*"})`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeContainerCount 容器Count
func (m *BKMonitor) GetNodeContainerCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max by (bk_instance) (kubelet_running_containers{cluster_id="%<clusterId>s", container_state="running", bk_instance=~"%<ip>s:.*"})`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeNetworkTransmit 节点网络发送量
func (m *BKMonitor) GetNodeNetworkTransmit(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max(bkmonitor:system:net:speed_sent{cluster_id="%<clusterId>s", ip="%<ip>s"})`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}

// GetNodeNetworkReceive 节点网络接收
func (m *BKMonitor) GetNodeNetworkReceive(ctx context.Context, projectId, clusterId, ip string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	params := map[string]interface{}{
		"clusterId": clusterId,
		"ip":        ip,
	}

	promql := `max(bkmonitor:system:net:speed_recv{cluster_id="%<clusterId>s", ip="%<ip>s"})`

	matrix, _, err := bcsmonitor.QueryRangeF(ctx, projectId, promql, params, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.MatrixToSeries(matrix), nil
}
