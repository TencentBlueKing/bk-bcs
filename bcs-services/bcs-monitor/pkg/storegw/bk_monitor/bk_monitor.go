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

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/store/labelpb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"gopkg.in/yaml.v2"
)

// Config 配置
type Config struct {
	Host string `yaml:"host"`
}

// BKMonitorStore implements the store node API on top of the Prometheus remote read API.
type BKMonitorStore struct {
	config *Config
	Base   *url.URL
}

// NewBKMonitorStore
func NewBKMonitorStore(conf []byte) (*BKMonitorStore, error) {
	var config Config
	if err := yaml.UnmarshalStrict(conf, &config); err != nil {
		return nil, errors.Wrap(err, "parsing bkmonitor stor config")
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	store := &BKMonitorStore{config: &config, Base: baseURL}
	return store, nil
}

// Info 返回元数据信息
func (s *BKMonitorStore) Info(context.Context, *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	labelSets := labels.FromMap(map[string]string{"provider": "BK_MONITOR"})

	zset := labelpb.ZLabelSet{Labels: labelpb.ZLabelsFromPromLabels(labelSets)}

	res := &storepb.InfoResponse{
		StoreType: component.Store.ToProto(),
		MinTime:   math.MinInt64,
		MaxTime:   math.MaxInt64,
		LabelSets: []labelpb.ZLabelSet{zset},
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
	return &storepb.LabelValuesResponse{Values: values}, nil
}

// Series 返回时序数据
func (s *BKMonitorStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error {
	return nil
}
