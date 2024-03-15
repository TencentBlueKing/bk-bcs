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

// Package bkmonitor bk monitor
package bkmonitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chonla/format"
	"github.com/prometheus/prometheus/prompb"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

const (
	// DisFstype xxx
	DisFstype = "ext[234]|btrfs|xfs|zfs" // 磁盘统计 允许的文件系统
	// DiskMountPoint xxx
	DiskMountPoint = "/data" // 磁盘统计 允许的挂载目录
	// PROVIDER xxx
	PROVIDER = `provider="BK_MONITOR"`

	// 默认查询分片，每10个节点一组并发查询，限制并发为 20，防止对数据源造成压力
	scale                  = 10
	defaultQueryConcurrent = 20
)

// BKMonitor :
type BKMonitor struct{}

// NewBKMonitor :
func NewBKMonitor() *BKMonitor {
	return &BKMonitor{}
}

// HandleBKMonitorClusterMetric bkmonitor metrics 处理
func HandleBKMonitorClusterMetric(ctx context.Context, projectID, clusterID string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	nodeSlice, ok := base.GetNodeMatchWithScaleIngErr(ctx, clusterID, scale)
	if !ok {
		return nil, nil
	}
	if len(nodeSlice) == 0 {
		return nil, nil
	}

	// get cluster info
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}
	provider := fmt.Sprintf(`bk_biz_id="%s"`, cluster.BKBizID)
	clusterMatch := fmt.Sprintf(`bcs_cluster_id="%s"`, clusterID)

	series := make([]*prompb.TimeSeries, 0)
	var mtx sync.Mutex
	var wg errgroup.Group
	wg.SetLimit(defaultQueryConcurrent)
	for _, res := range nodeSlice {
		res := res
		wg.Go(func() error {
			params := map[string]interface{}{
				"cluster":    clusterMatch,
				"node":       res.NodeNameMatch,
				"instance":   res.NodeMatch,
				"fstype":     DisFstype,
				"mountpoint": DiskMountPoint,
				"provider":   provider,
			}

			step := int64(clientutil.MinStepSeconds)
			// 直接查询 bk_monitor，数据源进行聚合查询
			rawQL := format.Sprintf(promql, params)
			s, err := bkmonitor_client.QueryByPromQL(ctx, config.G.BKMonitor.URL, cluster.BKBizID,
				start.Unix(), end.Unix(), step, nil, rawQL)
			if err != nil {
				return err
			}
			mtx.Lock()
			series = append(series, s...)
			mtx.Unlock()
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, err
	}
	// 对分片数据进行合并返回
	return base.MergeSameSeries(series), nil
}

// handleClusterMetric Cluster 处理公共函数
func (m *BKMonitor) handleClusterMetric(ctx context.Context, projectID, clusterID string, promql string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	return HandleBKMonitorClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPUTotal 获取集群CPU核心总量
func (m *BKMonitor) GetClusterCPUTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(avg_over_time(kube_node_status_allocatable_cpu_cores{%<cluster>s, ` +
			`job="kube-state-metrics", node=~"%<node>s", %<provider>s}[1m]))`
	// NOCC:goconst/string(设计如此)
	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPUUsed 获取CPU核心使用量
func (m *BKMonitor) GetClusterCPUUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(irate(node_cpu_seconds_total{%<cluster>s, mode!="idle", bk_instance=~"%<instance>s", %<provider>s}[2m]))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPUUsage 获取CPU核心使用率
func (m *BKMonitor) GetClusterCPUUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetClusterCPUUsed(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetClusterCPUTotal(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetClusterPodUsed 获取集群pod使用量
func (m *BKMonitor) GetClusterPodUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 获取pod使用率
	promql :=
		`sum (kubelet_running_pods{%<cluster>s, node=~"%<node>s", %<provider>s})`
	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterPodTotal 获取集群最大允许pod数
func (m *BKMonitor) GetClusterPodTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// 获取集群中最大可用pod数
	nodes, err := k8sclient.GetClusterNodeList(ctx, clusterID, true)
	if err != nil {
		return nil, err
	}
	var pod int64
	for _, node := range nodes {
		pod += node.Status.Allocatable.Pods().Value()
	}
	return base.GetSameSeries(start, end, step, float64(pod), nil), nil
}

// GetClusterPodUsage 获取集群pod使用率
func (m *BKMonitor) GetClusterPodUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	usedSeries, err := m.GetClusterPodUsed(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}

	nodes, err := k8sclient.GetClusterNodeList(ctx, clusterID, true)
	if err != nil {
		return nil, err
	}
	var pod int64
	for _, node := range nodes {
		pod += node.Status.Allocatable.Pods().Value()
	}

	return base.DivideSeriesByValue(usedSeries, float64(pod)), nil
}

// GetClusterCPURequest 获取CPU Rquest
func (m *BKMonitor) GetClusterCPURequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		sum(avg_over_time(kube_pod_container_resource_requests_cpu_cores{%<cluster>s, job="kube-state-metrics", ` +
		`node=~"%<node>s", %<provider>s}[1m]))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterCPURequestUsage 获取CPU核心装箱率
func (m *BKMonitor) GetClusterCPURequestUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetClusterCPURequest(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetClusterCPUTotal(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetClusterMemoryTotal 获取集群CPU核心总量
func (m *BKMonitor) GetClusterMemoryTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	// NOCC:goconst/string(设计如此)
	promql :=
		`sum(avg_over_time(kube_node_status_allocatable_memory_bytes{%<cluster>s, ` +
			`job="kube-state-metrics", node=~"%<node>s", %<provider>s}[1m]))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryUsed 获取集群内存使用量
func (m *BKMonitor) GetClusterMemoryUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`(sum(node_memory_MemTotal_bytes{%<cluster>s, bk_instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_MemFree_bytes{%<cluster>s, bk_instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_Buffers_bytes{%<cluster>s, bk_instance=~"%<instance>s", %<provider>s}) - ` +
			`sum(node_memory_Cached_bytes{%<cluster>s, bk_instance=~"%<instance>s", %<provider>s}) + ` +
			`sum(node_memory_Shmem_bytes{%<cluster>s, bk_instance=~"%<instance>s", %<provider>s}))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryUsage 获取内存使用率
func (m *BKMonitor) GetClusterMemoryUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetClusterMemoryUsed(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetClusterMemoryTotal(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetClusterMemoryRequest 获取内存 Request
func (m *BKMonitor) GetClusterMemoryRequest(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql := `
		sum(avg_over_time(kube_pod_container_resource_requests_memory_bytes{%<cluster>s, job="kube-state-metrics", ` +
		`node=~"%<node>s", %<provider>s}[1m]))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterMemoryRequestUsage 获取内存装箱率
func (m *BKMonitor) GetClusterMemoryRequestUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	seriesA, err := m.GetClusterMemoryRequest(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.GetClusterMemoryTotal(ctx, projectID, clusterID, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetClusterDiskTotal 集群磁盘总量
func (m *BKMonitor) GetClusterDiskTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_filesystem_size_bytes{%<cluster>s, bk_instance=~"%<instance>s", fstype=~"%<fstype>s", ` + // nolint
			`mountpoint=~"%<mountpoint>s", %<provider>s})` // nolint

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskUsed 集群磁盘使用
func (m *BKMonitor) GetClusterDiskUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(node_filesystem_size_bytes{%<cluster>s, bk_instance=~"%<instance>s", fstype=~"%<fstype>s", ` +
			`mountpoint=~"%<mountpoint>s", %<provider>s}) -
		sum(node_filesystem_free_bytes{%<cluster>s, bk_instance=~"%<instance>s", fstype=~"%<fstype>s", ` +
			`mountpoint=~"%<mountpoint>s", %<provider>s})`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskUsage 集群磁盘使用率
func (m *BKMonitor) GetClusterDiskUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promqlA :=
		`(sum(node_filesystem_size_bytes{%<cluster>s, bk_instance=~"%<instance>s", fstype=~"%<fstype>s", ` +
			`mountpoint=~"%<mountpoint>s", %<provider>s}) -
		sum(node_filesystem_free_bytes{%<cluster>s, bk_instance=~"%<instance>s", fstype=~"%<fstype>s", ` +
			`mountpoint=~"%<mountpoint>s", %<provider>s}))`
	promqlB :=
		`sum(node_filesystem_size_bytes{%<cluster>s, bk_instance=~"%<instance>s", fstype=~"%<fstype>s", ` +
			`mountpoint=~"%<mountpoint>s", %<provider>s})`

	seriesA, err := m.handleClusterMetric(ctx, projectID, clusterID, promqlA, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.handleClusterMetric(ctx, projectID, clusterID, promqlB, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetClusterDiskioUsage 集群磁盘IO使用率
func (m *BKMonitor) GetClusterDiskioUsage(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promqlA :=
		`sum(max by(bk_instance) (rate(node_disk_io_time_seconds_total{%<cluster>s, bk_instance=~"%<instance>s", ` +
			`%<provider>s}[2m])))` // nolint
	promqlB :=
		`count(max by(bk_instance) (rate(node_disk_io_time_seconds_total{%<cluster>s, bk_instance=~"%<instance>s", ` +
			`%<provider>s}[2m])))`

	seriesA, err := m.handleClusterMetric(ctx, projectID, clusterID, promqlA, start, end, step)
	if err != nil {
		return nil, err
	}
	seriesB, err := m.handleClusterMetric(ctx, projectID, clusterID, promqlB, start, end, step)
	if err != nil {
		return nil, err
	}

	return base.DivideSeries(seriesA, seriesB), nil
}

// GetClusterDiskioUsed 集群磁盘IO使用量
func (m *BKMonitor) GetClusterDiskioUsed(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`sum(max by(bk_instance) (rate(node_disk_io_time_seconds_total{%<cluster>s, bk_instance=~"%<instance>s", ` +
			`%<provider>s}[2m])))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}

// GetClusterDiskioTotal 集群磁盘IO
func (m *BKMonitor) GetClusterDiskioTotal(ctx context.Context, projectID, clusterID string, start, end time.Time,
	step time.Duration) ([]*prompb.TimeSeries, error) {
	promql :=
		`count(max by(bk_instance) (rate(node_disk_io_time_seconds_total{%<cluster>s, bk_instance=~"%<instance>s", ` +
			`%<provider>s}[2m])))`

	return m.handleClusterMetric(ctx, projectID, clusterID, promql, start, end, step)
}
