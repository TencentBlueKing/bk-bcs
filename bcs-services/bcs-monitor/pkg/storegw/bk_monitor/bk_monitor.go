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
	"net/url"

	"github.com/pkg/errors"
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

func (s *BKMonitorStore) Info(context.Context, *storepb.InfoRequest) (*storepb.InfoResponse, error) {
	return nil, nil
}

func (s *BKMonitorStore) LabelNames(ctx context.Context, r *storepb.LabelNamesRequest) (*storepb.LabelNamesResponse, error) {
	return nil, nil
}

func (s *BKMonitorStore) LabelValues(ctx context.Context, r *storepb.LabelValuesRequest) (*storepb.LabelValuesResponse, error) {
	return nil, nil
}

func (s *BKMonitorStore) Series(r *storepb.SeriesRequest, srv storepb.Store_SeriesServer) error {
	return nil
}
