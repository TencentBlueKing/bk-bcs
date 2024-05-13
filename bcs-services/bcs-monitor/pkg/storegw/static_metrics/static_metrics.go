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

// Package staticmetrics ...
package staticmetrics

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/store/storepb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// StaticMetricsStore 只提供有限 metrics store
type StaticMetricsStore struct { // nolint
	Reg             *prometheus.Registry
	ExternalLabels  labels.Labels
	OutageTolerance time.Duration
	MetricNames     []string
}

// UpdateSeriesFunc 发送数据时, 转换数据函数
type UpdateSeriesFunc func(series *prompb.TimeSeries) *prompb.TimeSeries

// LabelSetFunc 发送数据时, 转换labels函数
type LabelSetFunc func(series *prompb.TimeSeries) []storepb.Label

// Info 实现方法
func (s *StaticMetricsStore) Info(_ context.Context, _ *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	// 只允许查询最近 2 个小时数据
	minTime := time.Now().Add(-s.OutageTolerance).UnixNano() / int64(time.Millisecond)

	labelSets := clientutil.ExtendLabelSetByNames(s.ExternalLabels, s.MetricNames)

	res := &storepb.InfoResponse{
		StoreType: component.Store.ToProto(),
		MinTime:   minTime,
		MaxTime:   math.MaxInt64,
		LabelSets: labelSets,
	}
	return res, nil
}

// GetMatcherSeries : 获取 匹配的 series
func (s *StaticMetricsStore) GetMatcherSeries(rawSeries []*prompb.TimeSeries, matchers []*labels.Matcher, minTime int64,
	maxTime int64) []*prompb.TimeSeries {
	seriesMap := map[uint64]*prompb.TimeSeries{}

	for _, raw := range rawSeries {
		if !clientutil.ValidateSeries(raw, minTime, maxTime) {
			continue
		}

		builder := labels.NewBuilder(nil)
		for _, l := range raw.Labels {
			builder.Set(l.Name, l.Value)
		}
		lb := builder.Labels()

		// bcs 租户匹配
		// if tenantsMatcher != nil && !tenantsMatcher.MatchLabels(lb) {
		// 	continue
		// }

		series, ok := seriesMap[lb.Hash()]
		if !ok {
			series = &prompb.TimeSeries{Labels: raw.Labels}
			seriesMap[lb.Hash()] = series
		}

		//  用户 promql label 匹配
		if clientutil.MatchLabels(matchers, lb, s.MetricNames) {
			series.Samples = append(series.Samples, raw.Samples...)
		}
	}

	// map 转换为列表返回
	seriesList := make([]*prompb.TimeSeries, 0, len(seriesMap))
	for _, v := range seriesMap {
		seriesList = append(seriesList, v)
	}

	return seriesList
}

// DefaultUpdateSeriesFunc 默认原因返回
func DefaultUpdateSeriesFunc(series *prompb.TimeSeries) *prompb.TimeSeries {
	return series
}

// DefaultLabelSetFunc 默认原样返回
func DefaultLabelSetFunc(series *prompb.TimeSeries) []storepb.Label {
	lset := make([]storepb.Label, 0, len(series.Labels))

	for _, v := range series.Labels {
		lset = append(lset, storepb.Label{Name: v.Name, Value: v.Value})
	}

	return lset
}

// MakeLabelSetFuncByName labels 中添加 __name__ label
func MakeLabelSetFuncByName(metricName string) LabelSetFunc {
	labelSetFunc := func(series *prompb.TimeSeries) []storepb.Label {
		lset := make([]storepb.Label, 0, len(series.Labels)+1)
		lset = append(lset, storepb.Label{Name: labels.MetricName, Value: metricName})
		for _, v := range series.Labels {
			lset = append(lset, storepb.Label{Name: v.Name, Value: v.Value})
		}
		return lset
	}
	return labelSetFunc
}

// SendSeries : 给 metricName 发送数据(流)
func (s *StaticMetricsStore) SendSeries(
	series *prompb.TimeSeries,
	srv storepb.Store_SeriesServer,
	updateSeriesFunc UpdateSeriesFunc,
	labelSetFunc LabelSetFunc,
) error {
	if len(series.Samples) == 0 {
		return nil
	}

	// 返回的点需要按时间排序
	sort.Slice(series.Samples, func(i, j int) bool {
		return series.Samples[i].Timestamp < series.Samples[j].Timestamp
	})

	if updateSeriesFunc != nil {
		series = updateSeriesFunc(series)
	} else {
		series = DefaultUpdateSeriesFunc(series)
	}

	aggregatedChunks, err := clientutil.ChunkSamples(series)
	if err != nil {
		return err
	}

	var lset []storepb.Label
	if labelSetFunc != nil {
		lset = labelSetFunc(series)
	} else {
		lset = DefaultLabelSetFunc(series)
	}

	resp := &storepb.Series{Labels: lset, Chunks: aggregatedChunks}
	return srv.Send(storepb.NewSeriesResponse(resp))
}

// Series 需要继承函数实现
func (s *StaticMetricsStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error {
	blog.Infow("Series func not implemented", "PromQL", clientutil.DumpPromQL(r), "minTime", r.MinTime, "maxTime",
		r.MaxTime)
	return nil
}

// LabelNames 只返回__name__
func (s *StaticMetricsStore) LabelNames(ctx context.Context, r *storepb.LabelNamesRequest) (*storepb.LabelNamesResponse,
	error) {
	names := &storepb.LabelNamesResponse{Names: []string{labels.MetricName}}
	return names, nil
}

// LabelValues 只返回metrics names
func (s *StaticMetricsStore) LabelValues(ctx context.Context, r *storepb.LabelValuesRequest) (
	*storepb.LabelValuesResponse, error) {
	values := &storepb.LabelValuesResponse{}
	if r.Label == labels.MetricName {
		values.Values = s.MetricNames
	}

	return values, nil
}
