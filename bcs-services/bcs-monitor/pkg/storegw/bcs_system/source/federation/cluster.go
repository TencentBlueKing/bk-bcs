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

package federation

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/prompb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// Federation xxx
type Federation struct {
	dispatch map[string]clientutil.DispatchConf
}

// GetClusterPodUsed 获取集群pod使用量
func (p *Federation) GetClusterPodUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodTotal 获取集群最大允许pod数
func (p *Federation) GetClusterPodTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodUsage 获取集群pod使用率
func (p *Federation) GetClusterPodUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// NewFederation xxx
func NewFederation(dispatch map[string]clientutil.DispatchConf) *Federation {
	return &Federation{dispatch: dispatch}
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (p *Federation) GetClusterCPUTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPUUsed 获取CPU核心使用量
func (p *Federation) GetClusterCPUUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPUUsage 获取CPU核心使用量
func (p *Federation) GetClusterCPUUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPURequest 获取CPU Request
func (p *Federation) GetClusterCPURequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterCPURequestUsage 获取CPU核心装箱率
func (p *Federation) GetClusterCPURequestUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryTotal 获取集群内存总量
func (p *Federation) GetClusterMemoryTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryUsed 获取内存使用量
func (p *Federation) GetClusterMemoryUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryUsage 获取内存使用率
func (p *Federation) GetClusterMemoryUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryRequest 获取内存 Request
func (p *Federation) GetClusterMemoryRequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterMemoryRequestUsage 获取内存装箱率
func (p *Federation) GetClusterMemoryRequestUsage(ctx context.Context, projectID, clusterID string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskTotal 获取集群磁盘总量
func (p *Federation) GetClusterDiskTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskUsed 获取集群磁盘使用量
func (p *Federation) GetClusterDiskUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskUsage 获取CPU核心使用量
func (p *Federation) GetClusterDiskUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioUsage 集群磁盘IO使用率
func (p *Federation) GetClusterDiskioUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioUsed 集群磁盘IO使用量
func (p *Federation) GetClusterDiskioUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioTotal 集群磁盘IO
func (p *Federation) GetClusterDiskioTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}
