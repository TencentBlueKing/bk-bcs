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
	"sync"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
)

// handlePodMetricFunc
type handleContainerMetricFunc func(handler base.MetricHandler, ctx context.Context, projectID, clusterID, namespace,
	podname string, containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)

func (m *Federation) handleContainerMetric(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration, fn handleContainerMetricFunc) (
	[]*prompb.TimeSeries, error) {
	// get managed clusters
	clusters, err := k8sclient.GetManagedClusterList(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	series := make([]*prompb.TimeSeries, 0)
	eg := errgroup.Group{}
	eg.SetLimit(20)
	mux := sync.Mutex{}
	for _, cls := range clusters {
		clusterID := cls
		groupFunc := func() error {
			client, err := ClientFactory(ctx, clusterID, m.dispatch[clusterID].SourceType,
				m.dispatch[clusterID].MetricsPrefix)
			if err != nil {
				klog.Warningf("get client in cluster %s error, %s", clusterID, err)
				return nil
			}
			s, err := fn(client, ctx, projectID, clusterID, namespace, podname, containerNameList, start, end, step)
			if err != nil {
				klog.Warningf("handle container metrics in cluster %s error, %s", clusterID, err)
				return nil
			}
			mux.Lock()
			series = append(series, s...)
			mux.Unlock()
			return nil
		}
		eg.Go(groupFunc)
	}
	_ = eg.Wait()
	return series, nil
}

// GetContainerCPUUsage 容器CPU使用率
func (m *Federation) GetContainerCPUUsage(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerCPUUsage
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerMemoryUsed 容器内存使用率
func (m *Federation) GetContainerMemoryUsed(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerMemoryUsed
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerCPULimit 容器CPU限制
func (m *Federation) GetContainerCPULimit(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerCPULimit
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerMemoryLimit 容器内存限制
func (m *Federation) GetContainerMemoryLimit(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerMemoryLimit
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerGPUMemoryUsage 容器GPU显卡使用率
func (m *Federation) GetContainerGPUMemoryUsage(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerGPUMemoryUsage
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerGPUUsed 容器GPU使用量
func (m *Federation) GetContainerGPUUsed(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerGPUUsed
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerGPUUsage 容器GPU使用率
func (m *Federation) GetContainerGPUUsage(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerGPUUsage
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerDiskReadTotal 容器磁盘读
func (m *Federation) GetContainerDiskReadTotal(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerDiskReadTotal
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}

// GetContainerDiskWriteTotal 容器磁盘写
func (m *Federation) GetContainerDiskWriteTotal(ctx context.Context, projectID, clusterID, namespace, podname string,
	containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetContainerDiskWriteTotal
	return m.handleContainerMetric(ctx, projectID, clusterID, namespace, podname, containerNameList,
		start, end, step, fn)
}
