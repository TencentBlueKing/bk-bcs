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

type handleGroupMetricFunc func(handler base.MetricHandler, ctx context.Context, projectID, clusterID, group string,
	start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)

func (m *Federation) handleGroupMetric(ctx context.Context, projectID, clusterID, group string,
	start, end time.Time, step time.Duration, fn handleGroupMetricFunc) (
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
				blog.Warnf("get client in cluster %s error, %s", clusterID, err)
				return nil
			}
			s, err := fn(client, ctx, projectID, clusterID, group, start, end, step)
			if err != nil {
				blog.Warnf("handle group metrics in cluster %s error, %s", clusterID, err)
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

// GetClusterGroupNodeNum 集群节点池数目
func (m *Federation) GetClusterGroupNodeNum(ctx context.Context, projectID, clusterID, group string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetClusterGroupNodeNum
	return m.handleGroupMetric(ctx, projectID, clusterID, group, start, end, step, fn)
}

// GetClusterGroupMaxNodeNum 集群最大节点池数目
func (m *Federation) GetClusterGroupMaxNodeNum(ctx context.Context, projectID, clusterID, group string, start,
	end time.Time, step time.Duration) ([]*prompb.TimeSeries, error) {
	fn := base.MetricHandler.GetClusterGroupMaxNodeNum
	return m.handleGroupMetric(ctx, projectID, clusterID, group, start, end, step, fn)
}
