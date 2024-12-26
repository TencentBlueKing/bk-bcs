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

// Package prometheus ...
package prometheus

import (
	"context"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/store"
	"github.com/thanos-io/thanos/pkg/store/labelpb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// Config 配置
type Config struct {
	URL            string            `yaml:"url" mapstructure:"url"` // prometheus api url
	ClusterID      string            `yaml:"cluster_id" mapstructure:"cluster_id"`
	ExternalLabels map[string]string `yaml:"external_labels" mapstructure:"external_labels"`
}

// PromStore implements the store node API on top of the  remote read API.
type PromStore struct {
	config  *Config
	promURL *url.URL
}

// NewPromStore xxx
func NewPromStore(conf []byte) (*PromStore, error) {
	var config Config
	if err := yaml.UnmarshalStrict(conf, &config); err != nil {
		return nil, errors.Wrap(err, "parsing prometheus store config")
	}

	promURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	store := &PromStore{
		config:  &config,
		promURL: promURL,
	}
	return store, nil
}

// Info 返回元数据信息
func (s *PromStore) Info(ctx context.Context, r *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	lsets := make([]labelpb.ZLabelSet, 0)

	lsets = append(lsets, labelpb.ZLabelSet{Labels: labelpb.ZLabelsFromPromLabels(
		labels.FromMap(map[string]string{
			"provider":   "PROMETHEUS",
			"cluster_id": s.config.ClusterID,
		}))})

	res := &storepb.InfoResponse{
		StoreType: component.Store.ToProto(),
		MinTime:   math.MinInt64,
		MaxTime:   math.MaxInt64,
		LabelSets: lsets,
	}
	return res, nil
}

// LabelNames 返回 labels 列表
func (s *PromStore) LabelNames(ctx context.Context, r *storepb.LabelNamesRequest) (*storepb.LabelNamesResponse,
	error) {
	labels, err := promclient.QueryLabels(ctx, s.promURL.String(), nil, r)
	if err != nil {
		return &storepb.LabelNamesResponse{Names: labels}, err
	}
	return &storepb.LabelNamesResponse{Names: labels}, nil
}

// LabelValues 返回 label values 列表
func (s *PromStore) LabelValues(ctx context.Context, r *storepb.LabelValuesRequest) (*storepb.LabelValuesResponse,
	error) {
	blog.Infow(storepb.MatchersToString(r.Matchers...), "request_id", store.RequestIDValue(ctx))
	values, err := promclient.QueryLabelValues(ctx, s.promURL.String(), nil, r)
	if err != nil {
		return &storepb.LabelValuesResponse{Values: values}, err
	}

	return &storepb.LabelValuesResponse{Values: values}, nil
}

// Series 返回时序数据
func (s *PromStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error {
	ctx := srv.Context()
	blog.Infow(clientutil.DumpPromQL(r), "request_id", store.RequestIDValue(ctx), "minTime", r.MinTime, "maxTime",
		r.MaxTime, "step", r.QueryHints.StepMillis)

	// step 固定1分钟
	// 注意: 目前实现的 aggrChunk 为 Raw 格式, 不支持降采样, 支持参考 https://thanos.io/tip/components/compact.md/
	step := int64(clientutil.MinStepSeconds)

	// 毫秒转换为秒
	start := time.UnixMilli(r.MinTime)
	end := time.UnixMilli(r.MaxTime)

	// series 数据, 这里只查询最近 SeriesStepDeltaSeconds
	if r.SkipChunks {
		end = time.Now()
		start = end.Add(-(time.Second * clientutil.SeriesStepDeltaSeconds))
	}

	metricName, err := clientutil.GetLabelMatchValue("__name__", r.Matchers)
	if err != nil {
		return err
	}
	if metricName == "" {
		return nil
		// return errors.New("metric name is required")
	}

	// bcs 聚合 metrics 忽略
	if strings.HasPrefix(metricName, "bcs:") {
		return nil
	}

	clusterID, err := clientutil.GetLabelMatchValue("cluster_id", r.Matchers)
	if err != nil {
		return err
	}

	scopeClusterID := store.ClusterIDValue(ctx)
	if clusterID == "" {
		clusterID = scopeClusterID
	}

	// invalid clusterID
	if clusterID == "" || clusterID != s.config.ClusterID {
		return nil
	}

	newMatchers := make([]storepb.LabelMatcher, 0)
	newMatchers = append(newMatchers, storepb.LabelMatcher{Name: "bcs_cluster_id", Value: clusterID})
	for _, m := range r.Matchers {
		if m.Name == "provider" {
			continue
		}
		if m.Name == "cluster_id" || m.Name == "bcs_cluster_id" {
			continue
		}
		newMatchers = append(newMatchers, m)
	}

	r.Matchers = newMatchers
	matrix, _, err := promclient.QueryRangeMatrix(ctx, s.promURL.String(), nil, r.ToPromQL(), start, end,
		time.Second*time.Duration(step))
	if err != nil {
		return err
	}

	promSeriesSet := base.MatrixToSeries(matrix)
	for _, promSeries := range promSeriesSet {
		series := &clientutil.TimeSeries{TimeSeries: promSeries}

		s, err := series.ToThanosSeries(r.SkipChunks)
		if err != nil {
			return err
		}
		if err := srv.Send(storepb.NewSeriesResponse(s)); err != nil {
			return err
		}
	}

	return nil
}
