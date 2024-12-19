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

// Package bcssystem bcs system
package bcssystem

import (
	"context"
	"math"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/store"
	"github.com/thanos-io/thanos/pkg/store/labelpb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// Config 配置
type Config struct {
	Dispatch []clientutil.DispatchConf `yaml:"dispatch"`
}

// BCSSystemStore implements the store node API on top of the Prometheus remote read API.
type BCSSystemStore struct { // nolint
	config   *Config
	dispatch map[string]clientutil.DispatchConf
}

// NewBCSSystemStore :
func NewBCSSystemStore(conf []byte) (*BCSSystemStore, error) { // nolint
	var config Config
	if err := yaml.UnmarshalStrict(conf, &config); err != nil {
		return nil, errors.Wrap(err, "parsing bcs_system store config")
	}

	store := &BCSSystemStore{
		config:   &config,
		dispatch: make(map[string]clientutil.DispatchConf, 0),
	}

	for _, d := range config.Dispatch {
		store.dispatch[d.ClusterID] = d
	}
	return store, nil
}

// Info 返回元数据信息
func (s *BCSSystemStore) Info(ctx context.Context, r *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	// 默认配置
	lsets := []labelpb.ZLabelSet{}

	for _, m := range AvailableNodeMetrics {
		labelSets := labels.FromMap(map[string]string{"provider": "BCS_SYSTEM", "__name__": m})
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
func (s *BCSSystemStore) LabelNames(ctx context.Context, r *storepb.LabelNamesRequest) (*storepb.LabelNamesResponse,
	error) {
	names := []string{"__name__"}
	return &storepb.LabelNamesResponse{Names: names}, nil
}

// LabelValues 返回 label values 列表
func (s *BCSSystemStore) LabelValues(ctx context.Context, r *storepb.LabelValuesRequest) (*storepb.LabelValuesResponse,
	error) {
	values := []string{}
	if r.Label == "__name__" {
		values = append(values, AvailableNodeMetrics...)
	}

	return &storepb.LabelValuesResponse{Values: values}, nil
}

// Series 返回时序数据
// NOCC:golint/fnsize(设计如此)
func (s *BCSSystemStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error { // nolint
	ctx := srv.Context()
	blog.Infow(clientutil.DumpPromQL(r), "request_id", store.RequestIDValue(ctx), "minTime=",
		r.MinTime, "maxTime", r.MaxTime, "step", r.QueryHints.StepMillis)

	// series 数据, 这里只查询最近 SeriesStepDeltaSeconds
	if r.SkipChunks {
		r.MaxTime = time.Now().UnixMilli()
		r.MinTime = r.MaxTime - clientutil.SeriesStepDeltaSeconds*1000
	}

	startTime := time.UnixMilli(r.MinTime)
	endTime := time.UnixMilli(r.MaxTime)
	// step 固定1分钟
	stepDuration := time.Second * time.Duration(clientutil.MinStepSeconds)

	metricName, err := clientutil.GetLabelMatchValue("__name__", r.Matchers)
	if err != nil {
		return err
	}
	if metricName == "" {
		// return errors.New("metric name is required")
		return nil
	}

	clusterID, err := clientutil.GetLabelMatchValue("cluster_id", r.Matchers)
	if err != nil {
		return err
	}

	if clusterID == "" {
		return nil
		// return errors.New("cluster_id is required")
	}

	node, err := clientutil.GetLabelMatchValue("node", r.Matchers)
	if err != nil {
		return err
	}

	namespace, err := clientutil.GetLabelMatchValue("namespace", r.Matchers)
	if err != nil {
		return err
	}

	group, err := clientutil.GetLabelMatchValue("group", r.Matchers)
	if err != nil {
		return err
	}

	podNameList, err := clientutil.GetLabelMatchValues("pod_name", r.Matchers)
	if err != nil {
		return err
	}

	podName, err := clientutil.GetLabelMatchValue("pod_name", r.Matchers)
	if err != nil {
		return err
	}

	containerNameList, err := clientutil.GetLabelMatchValues("container_name", r.Matchers)
	if err != nil {
		return err
	}

	cluster, err := bcs.GetCluster(clusterID)
	if err != nil {
		return err
	}

	client, err := source.ClientFactory(ctx, cluster.ClusterID, s.dispatch[clusterID].SourceType, s.dispatch,
		cluster.IsVirtual())
	if err != nil {
		return err
	}

	params := metricsParams{
		client:         client,
		ctx:            ctx,
		projectID:      cluster.ProjectID,
		clusterID:      cluster.ClusterID,
		namespace:      namespace,
		group:          group,
		podName:        podName,
		podNames:       podNameList,
		containerNames: containerNameList,
		node:           node,
		startTime:      startTime,
		endTime:        endTime,
		stepDuration:   stepDuration,
	}
	metricsFn, ok := metricsMaps[metricName]
	if !ok {
		return nil
	}
	promSeriesSet, promErr := metricsFn(params)
	if promErr != nil {
		return promErr
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
