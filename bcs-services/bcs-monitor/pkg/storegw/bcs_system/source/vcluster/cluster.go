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

	"github.com/prometheus/prometheus/prompb"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	bkmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/bk_monitor"
)

const (
	// PROVIDER BK_MONITOR
	PROVIDER = `provider="BK_MONITOR"`
)

// VCluster :
type VCluster struct{}

// NewVCluster :
func NewVCluster() *VCluster {
	return &VCluster{}
}

// handleClusterMetric Cluster 处理公共函数
func (m *VCluster) handleClusterMetric(ctx context.Context, projectID, clusterID string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	return bkmonitor.HandleBKMonitorClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (m *VCluster) GetClusterCPUTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	res := resource.MustParse(cluster.VclusterInfo.Quota.CPULimits)
	return base.GetSameSeries(start, end, step, res.AsApproximateFloat64(), nil), nil
}

// GetClusterCPUUsed 获取CPU核心使用量
func (m *VCluster) GetClusterCPUUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(irate(container_cpu_usage_seconds_total{%<cluster>s, container_name!="", ` +
			`container_name!="POD", %<provider>s}[2m]))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPUUsage 获取CPU核心使用率
func (m *VCluster) GetClusterCPUUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	used :=
		`sum(irate(container_cpu_usage_seconds_total{%<cluster>s, container_name!="", ` +
			`container_name!="POD", %<provider>s}[2m]))`
	seriesA, err := m.handleClusterMetric(ctx, projectID, clusterID, used, start, end, step)
	if err != nil {
		return nil, err
	}
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	res := resource.MustParse(cluster.VclusterInfo.Quota.CPULimits)
	return base.DivideSeriesByValue(seriesA, res.AsApproximateFloat64()), nil
}

// GetClusterCPURequest 获取CPU Rquest
func (m *VCluster) GetClusterCPURequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(avg_over_time(kube_pod_container_resource_requests_cpu_cores{%<cluster>s, job="kube-state-metrics", ` +
			`%<provider>s}[1m]))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPURequestUsage 获取CPU核心装箱率
func (m *VCluster) GetClusterCPURequestUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promqlA := `sum(avg_over_time(kube_pod_container_resource_requests_cpu_cores{%<cluster>s, ` +
		`job="kube-state-metrics", %<provider>s}[1m]))`

	seriesA, err := m.handleClusterMetric(ctx, projectID, clusterID, promqlA, start, end, step)
	if err != nil {
		return nil, err
	}
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	res := resource.MustParse(cluster.VclusterInfo.Quota.CPULimits)
	return base.DivideSeriesByValue(seriesA, res.AsApproximateFloat64()), nil
}

// GetClusterMemoryTotal 获取集群CPU核心总量
func (m *VCluster) GetClusterMemoryTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	res := resource.MustParse(cluster.VclusterInfo.Quota.MemoryLimits)
	return base.GetSameSeries(start, end, step, res.AsApproximateFloat64(), nil), nil
}

// GetClusterMemoryUsed 获取集群内存使用量
func (m *VCluster) GetClusterMemoryUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum by (cluster_id) (container_memory_working_set_bytes{%<cluster>s, container_name!="", ` +
			`container_name!="POD", %<provider>s})`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryUsage 获取内存使用率
func (m *VCluster) GetClusterMemoryUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promqlA :=
		`sum by (cluster_id) (container_memory_working_set_bytes{%<cluster>s, container_name!="", ` +
			`container_name!="POD", %<provider>s})`

	seriesA, err := m.handleClusterMetric(ctx, projectID, clusterID, promqlA, start, end, step)
	if err != nil {
		return nil, err
	}
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	res := resource.MustParse(cluster.VclusterInfo.Quota.MemoryLimits)
	return base.DivideSeriesByValue(seriesA, res.AsApproximateFloat64()), nil
}

// GetClusterMemoryRequest 获取内存 Request
func (m *VCluster) GetClusterMemoryRequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(avg_over_time(kube_pod_container_resource_requests_memory_bytes{%<cluster>s, job="kube-state-metrics", ` +
			`%<provider>s}[1m]))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryRequestUsage 获取内存装箱率
func (m *VCluster) GetClusterMemoryRequestUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promqlA := `sum(avg_over_time(kube_pod_container_resource_requests_memory_bytes{%<cluster>s, ` +
		`job="kube-state-metrics", %<provider>s}[1m]))`

	seriesA, err := m.handleClusterMetric(ctx, projectID, clusterID, promqlA, start, end, step)
	if err != nil {
		return nil, err
	}
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	res := resource.MustParse(cluster.VclusterInfo.Quota.MemoryLimits)
	return base.DivideSeriesByValue(seriesA, res.AsApproximateFloat64()), nil
}

// GetClusterDiskTotal 集群磁盘总量
func (m *VCluster) GetClusterDiskTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskUsed 集群磁盘使用
func (m *VCluster) GetClusterDiskUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskUsage 集群磁盘使用率
func (m *VCluster) GetClusterDiskUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioUsage 集群磁盘IO使用率
func (m *VCluster) GetClusterDiskioUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioUsed 集群磁盘IO使用量
func (m *VCluster) GetClusterDiskioUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterDiskioTotal 集群磁盘IO
func (m *VCluster) GetClusterDiskioTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodUsed 获取集群pod使用量
func (m *VCluster) GetClusterPodUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodTotal 获取集群最大允许pod数
func (m *VCluster) GetClusterPodTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}

// GetClusterPodUsage 获取集群pod使用率
func (m *VCluster) GetClusterPodUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	return nil, nil
}
