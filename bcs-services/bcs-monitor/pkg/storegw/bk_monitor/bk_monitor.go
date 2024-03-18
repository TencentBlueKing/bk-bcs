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
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/store"
	"github.com/thanos-io/thanos/pkg/store/labelpb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// Config 配置
type Config struct {
	Dispatch []clientutil.DispatchConf `yaml:"dispatch"`
}

// BKMonitorStore implements the store node API on top of the Prometheus remote read API.
type BKMonitorStore struct { // nolint
	config   *Config
	dispatch map[string]clientutil.DispatchConf
}

// NewBKMonitorStore xxx
func NewBKMonitorStore(conf []byte) (*BKMonitorStore, error) {
	var config Config
	if err := yaml.UnmarshalStrict(conf, &config); err != nil {
		return nil, errors.Wrap(err, "parsing bk_monitor store config")
	}

	store := &BKMonitorStore{
		config:   &config,
		dispatch: make(map[string]clientutil.DispatchConf, 0),
	}

	for _, d := range config.Dispatch {
		store.dispatch[d.ClusterID] = d
	}
	return store, nil
}

// Info 返回元数据信息
func (s *BKMonitorStore) Info(ctx context.Context, r *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	// 默认配置
	var lsets []labelpb.ZLabelSet

	clusterMap, err := bcs.GetClusterMap()
	if err != nil {
		return nil, err
	}

	// NOCC:ineffassign/assign(误报)
	var grayClusterMap map[string]struct{}
	if config.G.BKMonitor.EnableGrey {
		grayClusterMap, err = bkmonitor_client.QueryGrayClusterMap(ctx, config.G.BKMonitor.MetadataURL)
		if err != nil {
			klog.Errorf("query bk_monitor cluster list error, %s", err)
		}
		lsets = make([]labelpb.ZLabelSet, 0, len(grayClusterMap))
		for clusterID := range grayClusterMap {
			// 不存在的，或者已经删除的集群，需要过滤
			if _, ok := clusterMap[clusterID]; !ok {
				continue
			}
			labelSets := labels.FromMap(map[string]string{"provider": "BK_MONITOR", "cluster_id": clusterID})
			lsets = append(lsets, labelpb.ZLabelSet{Labels: labelpb.ZLabelsFromPromLabels(labelSets)})
		}
	} else {
		lsets = make([]labelpb.ZLabelSet, 0, len(clusterMap))
		for clusterID := range clusterMap {
			labelSets := labels.FromMap(map[string]string{"provider": "BK_MONITOR", "cluster_id": clusterID})
			lsets = append(lsets, labelpb.ZLabelSet{Labels: labelpb.ZLabelsFromPromLabels(labelSets)})
		}
	}

	for _, m := range AvailableNodeMetrics {
		labelSets := labels.FromMap(map[string]string{"provider": "BK_MONITOR", "__name__": m})
		lsets = append(lsets, labelpb.ZLabelSet{Labels: labelpb.ZLabelsFromPromLabels(labelSets)})
	}

	res := &storepb.InfoResponse{
		StoreType: component.Store.ToProto(),
		MinTime:   math.MinInt64,
		MaxTime:   math.MaxInt64,
		LabelSets: lsets,
	}
	return res, nil
}

// LabelNames 返回 labels 列表
func (s *BKMonitorStore) LabelNames(ctx context.Context, r *storepb.LabelNamesRequest) (*storepb.LabelNamesResponse,
	error) {
	names := []string{"__name__"}
	return &storepb.LabelNamesResponse{Names: names}, nil
}

// LabelValues 返回 label values 列表
func (s *BKMonitorStore) LabelValues(ctx context.Context, r *storepb.LabelValuesRequest) (*storepb.LabelValuesResponse,
	error) {
	klog.InfoS(storepb.MatchersToString(r.Matchers...), "request_id", store.RequestIDValue(ctx))
	values := []string{}
	if r.Label != "__name__" {
		return &storepb.LabelValuesResponse{Values: values}, nil
	}

	// get cluster id
	clusterID, err := clientutil.GetLabelMatchValue("cluster_id", r.Matchers)
	if err != nil {
		return nil, err
	}

	scopeClusterID := store.ClusterIDValue(ctx)
	if clusterID == "" && scopeClusterID == "" {
		return &storepb.LabelValuesResponse{Values: values}, nil
	}

	// 优先使用 clusterID
	if scopeClusterID != "" {
		clusterID = scopeClusterID
	}
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}

	// get bk_monitor metrics
	metrics, err := bkmonitor_client.GetMetricsList(ctx, config.G.BKMonitor.MetadataURL, clusterID, cluster.BKBizID)
	if err != nil {
		return nil, err
	}
	for _, v := range metrics {
		values = append(values, v.Metric)
	}
	values = append(values, AvailableNodeMetrics...)

	return &storepb.LabelValuesResponse{Values: values}, nil
}

// Series 返回时序数据
func (s *BKMonitorStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error {
	ctx := srv.Context()
	klog.InfoS(clientutil.DumpPromQL(r), "request_id", store.RequestIDValue(ctx), "minTime", r.MinTime, "maxTime",
		r.MaxTime, "step", r.QueryHints.StepMillis)

	// step 固定1分钟
	// 注意: 目前实现的 aggrChunk 为 Raw 格式, 不支持降采样, 支持参考 https://thanos.io/tip/components/compact.md/
	step := int64(clientutil.MinStepSeconds)

	// 毫秒转换为秒
	start := time.UnixMilli(r.MinTime).Unix()
	end := time.UnixMilli(r.MaxTime).Unix()

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
	if clusterID == "" && scopeClusterID == "" {
		return nil
	}

	// 优先使用 clusterID
	if scopeClusterID != "" {
		clusterID = scopeClusterID
	}
	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return err
	}

	// series 数据, 这里只查询最近 SeriesStepDeltaSeconds
	if r.SkipChunks {
		// end = time.Now().Unix()
		// start = end - clientutil.SeriesStepDeltaSeconds
		return s.getMatcherSeries(r, srv, clusterID, cluster.BKBizID)
	}

	newMatchers := getMatcher(r.Matchers, metricName, cluster)
	// 必须的参数 bk_biz_id, 单独拎出来处理

	r.Matchers = newMatchers
	pql := ""
	if r.QueryHints != nil && r.QueryHints.Func != nil &&
		utils.StringInSlice(r.QueryHints.Func.Name, AvailableFuncNames) {
		// 传递函数到底层数据源，来实现特定的特性，如：把 avg_over_time 之类的时间函数传递到底层数据源，可以忽略 prometheus 回朔特性
		pql = r.ToPromQL()
	}
	bkmonitorURL := config.G.BKMonitor.URL
	if url, ok := s.dispatch[clusterID]; ok {
		bkmonitorURL = url.URL
	}
	promSeriesSet, err := bkmonitor_client.QueryByPromQL(srv.Context(), bkmonitorURL, cluster.BKBizID,
		start, end, step, newMatchers, pql)
	if err != nil {
		return err
	}

	for _, promSeries := range promSeriesSet {
		series := &clientutil.TimeSeries{TimeSeries: promSeries}
		series = series.AddLabel("__name__", metricName)
		series = series.AddLabel("cluster_id", clusterID)
		series = series.RenameLabel("bk_namespace", "namespace")
		series = series.RenameLabel("bk_pod", "pod")

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

func getMatcher(matchers []storepb.LabelMatcher, metricName string, cluster *bcs.Cluster) []storepb.LabelMatcher {
	newMatchers := make([]storepb.LabelMatcher, 0)
	for _, m := range matchers {
		if m.Name == "provider" {
			continue
		}

		// 集群Id转换为 bcs 的规范
		if m.Name == "cluster_id" {
			// 对 bkmonitor: 为 蓝鲸监控主机的数据, 不能添加集群过滤
			if strings.HasPrefix(metricName, "bkmonitor:") {
				continue
			}
			newMatchers = append(newMatchers, storepb.LabelMatcher{Name: "bcs_cluster_id", Value: m.Value})
			continue
		}

		newMatchers = append(newMatchers, m)
	}
	newMatchers = append(newMatchers, storepb.LabelMatcher{Name: "bcs_cluster_id", Value: cluster.ClusterID})
	return newMatchers
}

func (s *BKMonitorStore) getMatcherSeries(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer,
	clusterID, bizID string) error {
	ctx := srv.Context()
	klog.InfoS(clientutil.DumpPromQL(r), "request_id", store.RequestIDValue(ctx), "clusterID", clusterID)
	series, err := bkmonitor_client.GetMetricsSeries(ctx, config.G.BKMonitor.MetadataURL, clusterID, bizID)
	if err != nil {
		return err
	}
	klog.InfoS("series", "request_id", store.RequestIDValue(ctx), "clusterID", clusterID, "len", len(series))

	metricsLabel := clientutil.GetLabelMatch("__name__", r.Matchers)
	if metricsLabel == nil || metricsLabel.Value == "" {
		return nil
	}

	for _, promSeries := range series {
		name := clientutil.GetLabelMatchValueFromSeries("__name__", promSeries.Labels)
		match := filterMetrics(name, metricsLabel.Value, metricsLabel.Type)
		if !match {
			continue
		}
		series := &clientutil.TimeSeries{TimeSeries: promSeries}
		series = series.AddLabel("cluster_id", clusterID)
		series = series.RenameLabel("bk_namespace", "namespace")
		series = series.RenameLabel("bk_pod", "pod")

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

func filterMetrics(name, queryName string, matchType storepb.LabelMatcher_Type) bool {
	switch matchType {
	case storepb.LabelMatcher_EQ:
		return name == queryName
	case storepb.LabelMatcher_NEQ:
		return name != queryName
	case storepb.LabelMatcher_RE:
		reg, _ := regexp.Compile(queryName)
		return reg.MatchString(name)
	case storepb.LabelMatcher_NRE:
		reg, _ := regexp.Compile(queryName)
		return !reg.MatchString(name)
	}
	return false
}
