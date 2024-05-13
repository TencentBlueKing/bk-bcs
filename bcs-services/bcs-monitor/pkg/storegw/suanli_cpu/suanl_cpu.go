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

// Package suanlicpu implements the store node API on top of the Prometheus remote read API.
package suanlicpu

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/thanos-io/thanos/pkg/exthttp"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"github.com/thanos-io/thanos/pkg/tracing"
	"github.com/thanos-io/thanos/pkg/tracing/client"
	yaml "gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
	staticmetrics "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/static_metrics"
)

// Store implements the store node API on top of the Prometheus remote read API.
type Store struct {
	staticmetrics.StaticMetricsStore
	reg    *prometheus.Registry
	base   *url.URL
	config *Config
}

// Config 配置
type Config struct {
	Name          string   `yaml:"name"`
	URL           string   `yaml:"url"`
	ClusterID     string   `yaml:"cluster_id"`               // 对应的集群
	ExtNamespaces []string `yaml:"ext_namespaces,omitempty"` // 额外的namespaces 对应算力实际的namespace
}

// NewSuanLiCPUStore returns a new SuanLiCPUStore that uses the given HTTP client
// to talk to Prometheus.
// It attaches the provided external labels to all results.
func NewSuanLiCPUStore(reg *prometheus.Registry, conf []byte) (*Store, error) {
	var config Config
	if err := yaml.UnmarshalStrict(conf, &config); err != nil {
		return nil, errors.Wrap(err, "parsing prometheus config")
	}

	baseURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	t := exthttp.NewTransport()
	t.MaxIdleConnsPerHost = 100
	t.MaxIdleConns = 0

	staticStore := staticmetrics.StaticMetricsStore{
		Reg:             reg,
		ExternalLabels:  labels.FromMap(map[string]string{"provider": provider, "cluster_id": config.ClusterID}),
		MetricNames:     MetricNames,
		OutageTolerance: time.Hour * 24 * 15,
	}

	p := &Store{
		StaticMetricsStore: staticStore,
		reg:                reg,
		config:             &config,
		base:               baseURL,
	}
	return p, nil
}

// MakeNamespaceMatcher 把集群namspace转换为label match规则
func (p *Store) MakeNamespaceMatcher(ctx context.Context) (*storepb.LabelMatcher, error) {
	namespaces, err := k8sclient.GetNamespaces(ctx, p.config.ClusterID)
	if err != nil {
		return nil, errors.Wrap(err, "list namespace")
	}
	namespaces = append(namespaces, p.config.ExtNamespaces...)

	matcher := &storepb.LabelMatcher{
		Name:  "namespace",
		Value: strings.Join(namespaces, "|"),
		Type:  storepb.LabelMatcher_RE,
	}
	return matcher, nil
}

// Series returns all series for a requested time range and label matcher.
func (p *Store) Series(r *storepb.SeriesRequest, s storepb.Store_SeriesServer) error {
	metricName, err := clientutil.GetLabelMatchValue("__name__", r.Matchers)
	if err != nil {
		return err
	}

	// 只转合法的 metrics 数据
	if !strings.HasPrefix(metricName, "k8s_container_bs_") {
		return nil
	}

	ctx := tracing.ContextWithTracer(s.Context(), client.NoopTracer())

	podType, err := clientutil.GetLabelMatchValue("pod_type", r.Matchers)
	if err != nil {
		return err
	}

	// GPU, 需要做 label 转换
	if podType == "GPU" {
		return p.FetchAndSendGPUSeries(ctx, r, s)
	}

	return p.FetchAndSendPromSeries(ctx, r, s)
}

// FetchAndSendPromSeries 算力原始数据
func (p *Store) FetchAndSendPromSeries(ctx context.Context, r *storepb.SeriesRequest,
	s storepb.Store_SeriesServer) error {
	matchers := make([]storepb.LabelMatcher, 0, len(r.Matchers))
	ignoreLabels := map[string]string{
		"cluster_display_name": "",
		"cluster_id":           "",
	}

	for _, m := range r.Matchers {
		if _, ok := ignoreLabels[m.Name]; ok {
			continue
		}
		matchers = append(matchers, m)
	}

	namespaceMatcher, err := p.MakeNamespaceMatcher(ctx)
	if err != nil {
		return err
	}

	// 添加算力集群ID限制， 使用正则表达式
	matchers = append(matchers, *namespaceMatcher)
	data, _, err := p.QueryRangeInGRPC(ctx, p.base, matchers, clientutil.GetPromQueryTime(r))
	if err != nil {
		return err
	}

	// 删除老的label, 替换为BCS的集群ID
	appendLabels := map[string]string{
		"cluster_id": p.config.ClusterID,
		"provider":   provider,
	}
	series := clientutil.SampleStreamToSeries(data, ignoreLabels, appendLabels)
	for _, serie := range series {
		if err := p.SendSeries(serie, s, nil, nil); err != nil {
			return err
		}
	}

	return nil
}

// QueryRangeInGRPC query range in grpc
func (p *Store) QueryRangeInGRPC(ctx context.Context, base *url.URL, matchers []storepb.LabelMatcher,
	queryTime *clientutil.PromQueryTime) (model.Matrix, []string, error) {
	query := clientutil.LabelMatcherToString(matchers)
	return promclient.QueryRangeMatrix(ctx, base.String(), http.Header{}, query, queryTime.Start, queryTime.End,
		queryTime.Step)
}
