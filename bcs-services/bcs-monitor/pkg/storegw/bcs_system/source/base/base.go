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

package base

import (
	"context"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

// MetricHandler
type MetricHandler interface {
	GetClusterCPUTotal(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterCPUUsed(ctx context.Context, projectId, clusterId string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	// GetClusterCPUUsage(ctx context.Context, clusterId string) (*prompb.TimeSeries, error)
}

// GetNodeMatch 按集群node节点正则匹配
func GetNodeMatch(ctx context.Context, clusterId string) (string, error) {
	nodeList, err := k8sclient.GetNodeList(ctx, clusterId, true)
	if err != nil {
		return "", err
	}

	instanceList := make([]string, 0, len(nodeList))
	for _, node := range nodeList {
		instanceList = append(instanceList, node+`:.*`)

	}
	return strings.Join(instanceList, "|"), nil
}

func sampleStreamToSeries(m *model.SampleStream) *prompb.TimeSeries {
	series := &prompb.TimeSeries{}
	for k, v := range m.Metric {
		series.Labels = append(series.Labels, prompb.Label{
			Name:  string(k),
			Value: string(v),
		})
	}
	for _, v := range m.Values {
		series.Samples = append(series.Samples, prompb.Sample{
			Timestamp: v.Timestamp.Time().UnixMilli(),
			Value:     float64(v.Value),
		})
	}
	return series
}

// MatrixToSeries prom返回转换为时序对象
func MatrixToSeries(matrix model.Matrix) []*prompb.TimeSeries {
	series := make([]*prompb.TimeSeries, 0, len(matrix))
	for _, m := range matrix {
		series = append(series, sampleStreamToSeries(m))
	}
	return series
}
