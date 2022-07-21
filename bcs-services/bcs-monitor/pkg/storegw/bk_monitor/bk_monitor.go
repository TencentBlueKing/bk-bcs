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
 *
 */

package bk_monitor

import (
	"context"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/store/labelpb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
)

// Config 配置
type Config struct {
	URL         string `yaml:"url" mapstructure:"metadata_url"`          // unify-query 访问地址
	MetadataURL string `yaml:"metadata_url" mapstructure:"metadata_url"` // 元数据地址, 目前只包含白名单
}

// BKMonitorStore implements the store node API on top of the Prometheus remote read API.
type BKMonitorStore struct {
	config      *Config
	baseURL     *url.URL
	metadataURL *url.URL
}

// NewBKMonitorStore
func NewBKMonitorStore(conf []byte) (*BKMonitorStore, error) {
	var config Config
	if err := yaml.UnmarshalStrict(conf, &config); err != nil {
		return nil, errors.Wrap(err, "parsing bkmonitor stor config")
	}

	baseURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	metadataURL, err := url.Parse(config.MetadataURL)
	if err != nil {
		return nil, err
	}

	store := &BKMonitorStore{
		config:      &config,
		baseURL:     baseURL,
		metadataURL: metadataURL,
	}
	return store, nil
}

// Info 返回元数据信息
func (s *BKMonitorStore) Info(ctx context.Context, r *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	labelSets := labels.FromMap(map[string]string{"provider": "BK_MONITOR"})

	zset := labelpb.ZLabelSet{Labels: labelpb.ZLabelsFromPromLabels(labelSets)}

	// 默认配置
	lsets := []labelpb.ZLabelSet{zset}

	clusterMap, err := bcs.GetClusterMap(ctx, config.G.BCS)
	if err != nil {
		return nil, err
	}

	clusterList, err := bkmonitor_client.QueryClusterList(ctx, s.config.MetadataURL)
	if err != nil {
		klog.Errorf("query bk_monitor cluster list error, %s", err)
	} else if clusterList.Enabled {
		lsets = make([]labelpb.ZLabelSet, 0, len(clusterList.ClusterIdList))
		for _, clusterId := range clusterList.ClusterIdList {
			// 不存在的，或者已经删除的集群，需要过滤
			if _, ok := clusterMap[clusterId]; !ok {
				continue
			}

			labelSets := labels.FromMap(map[string]string{"provider": "BK_MONITOR", "cluster_id": clusterId})
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
func (s *BKMonitorStore) LabelNames(ctx context.Context, r *storepb.LabelNamesRequest) (*storepb.LabelNamesResponse, error) {
	names := []string{"__name__"}
	return &storepb.LabelNamesResponse{Names: names}, nil
}

// LabelValues 返回 label values 列表
func (s *BKMonitorStore) LabelValues(ctx context.Context, r *storepb.LabelValuesRequest) (*storepb.LabelValuesResponse, error) {
	values := []string{}
	if r.Label == "__name__" {
		values = []string{"container_network_receive_bytes_total"}
	}
	values = append(values, AvailableNodeMetrics...)

	return &storepb.LabelValuesResponse{Values: values}, nil
}

// Series 返回时序数据
func (s *BKMonitorStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error {
	logger.Infow(clientutil.DumpPromQL(r), "minTime", r.MinTime, "maxTime", r.MaxTime)

	if r.Step < 60 {
		r.Step = 60
	}

	// 毫秒转换为秒
	start := time.UnixMilli(r.MinTime).Unix()
	end := time.UnixMilli(r.MaxTime).Unix()

	// series 数据, 这里只查询最近1分钟
	if r.SkipChunks {
		end = time.Now().Unix()
		start = end - 60
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

	clusterId, err := clientutil.GetLabelMatchValue("cluster_id", r.Matchers)
	if err != nil {
		return err
	}

	if clusterId == "" {
		return nil
		// return errors.New("cluster_id is required")
	}

	newMatchers := make([]storepb.LabelMatcher, 0, len(r.Matchers))
	for _, m := range r.Matchers {
		// 集群Id转换为 bcs 的规范
		if m.Name == "cluster_id" {
			// 对 bkmonitor: 为 蓝鲸监控主机的数据, 不能添加集群过滤
			if strings.HasPrefix(metricName, "bkmonitor:") {
				continue
			}
			newMatchers = append(newMatchers, storepb.LabelMatcher{Name: "bcs_cluster_id", Value: m.Value})
		} else {
			newMatchers = append(newMatchers, m)
		}
	}

	bcsConf := k8sclient.GetBCSConfByClusterId(clusterId)
	cluster, err := bcs.GetCluster(srv.Context(), bcsConf, clusterId)
	if err != nil {
		return err
	}

	promSeriesSet, err := bkmonitor_client.QueryByPromQL(srv.Context(), s.config.URL, cluster.BKBizID, start, end, r.Step, newMatchers)
	if err != nil {
		return err
	}

	for _, promSeries := range promSeriesSet {
		series := &clientutil.TimeSeries{TimeSeries: promSeries}
		series = series.AddLabel("__name__", metricName)
		series = series.AddLabel("cluster_id", clusterId)
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
