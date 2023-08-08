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

package vcluster

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/prometheus/prometheus/prompb"
)

// GetNodeInfo 节点信息
func (m *VCluster) GetNodeInfo(ctx context.Context, projectID, clusterID, nodeName string, t time.Time) (*base.NodeInfo,
	error) {
	return nil, nil
}

// GetNodeCPUTotal 节点CPU总量
func (m *VCluster) GetNodeCPUTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeCPURequest 节点CPU请求量
func (m *VCluster) GetNodeCPURequest(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeCPUUsed 节点CPU使用量
func (m *VCluster) GetNodeCPUUsed(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeCPUUsage 节点CPU使用率
func (m *VCluster) GetNodeCPUUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeCPURequestUsage 节点CPU装箱率
func (m *VCluster) GetNodeCPURequestUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeMemoryTotal 节点Memory总量
func (m *VCluster) GetNodeMemoryTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeMemoryRequest 节点Memory请求量
func (m *VCluster) GetNodeMemoryRequest(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeMemoryUsed 节点Memory使用量
func (m *VCluster) GetNodeMemoryUsed(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeMemoryUsage 内存使用率
func (m *VCluster) GetNodeMemoryUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeMemoryRequestUsage 内存装箱率
func (m *VCluster) GetNodeMemoryRequestUsage(ctx context.Context, projectID, clusterID, node string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeDiskTotal 节点磁盘总量
func (m *VCluster) GetNodeDiskTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeDiskUsed 节点磁盘使用量
func (m *VCluster) GetNodeDiskUsed(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeDiskUsage 节点磁盘使用率
func (m *VCluster) GetNodeDiskUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeDiskioUsage 节点磁盘IO使用率
func (m *VCluster) GetNodeDiskioUsage(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodePodCount PodCount
func (m *VCluster) GetNodePodCount(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodePodTotal PodTotal
func (m *VCluster) GetNodePodTotal(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeContainerCount 容器Count
func (m *VCluster) GetNodeContainerCount(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeNetworkTransmit 节点网络发送量
func (m *VCluster) GetNodeNetworkTransmit(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetNodeNetworkReceive 节点网络接收
func (m *VCluster) GetNodeNetworkReceive(ctx context.Context, projectID, clusterID, node string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}
