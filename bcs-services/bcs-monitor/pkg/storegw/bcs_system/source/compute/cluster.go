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

package compute

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/prompb"
)

// Compute xxx
type Compute struct {
}

// NewCompute xxx
func NewCompute() *Compute {
	return &Compute{}
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (p *Compute) GetClusterCPUTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPUUsed 获取CPU核心使用量
func (p *Compute) GetClusterCPUUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPUUsage 获取CPU核心使用量
func (p *Compute) GetClusterCPUUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPURequest 获取CPU Request
func (p *Compute) GetClusterCPURequest(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPURequestUsage 获取CPU核心装箱率
func (p *Compute) GetClusterCPURequestUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryTotal 获取集群内存总量
func (p *Compute) GetClusterMemoryTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryUsed 获取内存使用量
func (p *Compute) GetClusterMemoryUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryUsage 获取内存使用率
func (p *Compute) GetClusterMemoryUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryRequest 获取内存 Request
func (p *Compute) GetClusterMemoryRequest(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryRequestUsage 获取内存装箱率
func (p *Compute) GetClusterMemoryRequestUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskTotal 获取集群磁盘总量
func (p *Compute) GetClusterDiskTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskUsed 获取集群磁盘使用量
func (p *Compute) GetClusterDiskUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskUsage 获取CPU核心使用量
func (p *Compute) GetClusterDiskUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioUsage 集群磁盘IO使用率
func (p *Compute) GetClusterDiskioUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioUsed 集群磁盘IO使用量
func (p *Compute) GetClusterDiskioUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioTotal 集群磁盘IO
func (m *Compute) GetClusterDiskioTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodUsed 获取集群pod使用量
func (p *Compute) GetClusterPodUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodTotal 获取集群最大允许pod数
func (p *Compute) GetClusterPodTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodUsage 获取集群pod使用率
func (p *Compute) GetClusterPodUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}
