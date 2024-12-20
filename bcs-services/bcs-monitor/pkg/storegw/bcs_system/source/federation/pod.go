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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/prometheus/prometheus/prompb"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
)

// handlePodMetricFunc
type handlePodMetricFunc func(handler base.MetricHandler, ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)

func (m *Federation) handlePodMetric(ctx context.Context, projectID, clusterID, namespace string, podNameList []string,
	start, end time.Time, step time.Duration, fn handlePodMetricFunc) ([]*prompb.TimeSeries, error) {
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
				blog.Warnf("get client in cluster %s error, %s", clusterID, err)
				return nil
			}
			s, err := fn(client, ctx, projectID, clusterID, namespace, podNameList, start, end, step)
			if err != nil {
				blog.Warnf("handle container metrics in cluster %s error, %s", clusterID, err)
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

// GetPodCPUUsage POD 使用率
func (m *Federation) GetPodCPUUsage(ctx context.Context, projectID, clusterID, namespace string, podNameList []string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetPodCPUUsage
	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, start, end, step, fn)
}

// GetPodCPULimitUsage POD CPU Limit 使用率
func (m *Federation) GetPodCPULimitUsage(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetPodCPULimitUsage
	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, start, end, step, fn)
}

// GetPodCPURequestUsage POD CPU Request 使用率
func (m *Federation) GetPodCPURequestUsage(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetPodCPURequestUsage
	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, start, end, step, fn)
}

// GetPodMemoryUsed 内存使用量
func (m *Federation) GetPodMemoryUsed(ctx context.Context, projectID, clusterID, namespace string, podNameList []string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetPodMemoryUsed
	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, start, end, step, fn)
}

// GetPodNetworkReceive 网络接收
func (m *Federation) GetPodNetworkReceive(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetPodNetworkReceive
	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, start, end, step, fn)
}

// GetPodNetworkTransmit 网络发送
func (m *Federation) GetPodNetworkTransmit(ctx context.Context, projectID, clusterID, namespace string,
	podNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetPodNetworkTransmit
	return m.handlePodMetric(ctx, projectID, clusterID, namespace, podNameList, start, end, step, fn)
}
