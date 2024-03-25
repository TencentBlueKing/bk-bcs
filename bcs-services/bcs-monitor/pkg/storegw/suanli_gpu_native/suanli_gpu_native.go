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

// Package suanligpunative implements the store node API on top of the Prometheus remote read API.
package suanligpunative

import (
	"context"
	"fmt"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	yaml "gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/suanliclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/clientutil"
	staticmetrics "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/static_metrics"
)

// SeriesQuery series query
type SeriesQuery struct {
	metricName string
	namespace  string
	podName    string
	podType    string
	vmId       string
}

// String string
func (s *SeriesQuery) String() string {
	if s.vmId != "" {
		return fmt.Sprintf("%s{vm_id=\"%s\"}", s.metricName, s.vmId)
	}
	return fmt.Sprintf("%s{namespace=\"%s\", pod_name=\"%s\"}", s.metricName, s.namespace, s.podName)
}

// Config 配置
type Config struct {
	suanliclient.SuanLiConfig `yaml:",inline"`
	Name                      string `yaml:"name"`
	ClusterID                 string `yaml:"cluster_id"` // 对应的集群
}

// Store implements the store node API on top of the Prometheus remote read API.
type Store struct {
	staticmetrics.StaticMetricsStore
	reg          *prometheus.Registry
	config       *Config
	suanLiClient *suanliclient.SuanLiClient
}

// NewSuanLiGPUNativeStore returns a new SuanLiGPUNativeStore that uses the given HTTP client
// to talk to Prometheus.
// It attaches the provided external labels to all results.
func NewSuanLiGPUNativeStore(reg *prometheus.Registry, conf []byte) (*Store, error) {
	var config Config
	if err := yaml.UnmarshalStrict(conf, &config); err != nil {
		return nil, errors.Wrap(err, "parsing prometheus config")
	}

	suanLiClient, err := suanliclient.NewSuanLiClient(&config.SuanLiConfig)
	if err != nil {
		return nil, err
	}

	p := &Store{
		StaticMetricsStore: staticmetrics.StaticMetricsStore{
			Reg:             reg,
			ExternalLabels:  labels.FromMap(map[string]string{"provider": provider, "cluster_id": config.ClusterID}),
			MetricNames:     MetricNames,
			OutageTolerance: time.Hour * 24 * 15,
		},
		reg:          reg,
		config:       &config,
		suanLiClient: suanLiClient,
	}
	return p, nil
}

// GetSeriesQueryList get series query list
func (p *Store) GetSeriesQueryList(ctx context.Context, r *storepb.SeriesRequest) ([]*SeriesQuery,
	error) {
	seriesQueryList := []*SeriesQuery{}

	// 支持的过滤 labels
	metricName, err := clientutil.GetLabelMatchValue("__name__", r.Matchers)
	if err != nil {
		return nil, err
	}
	podType, err := clientutil.GetLabelMatchValue("pod_type", r.Matchers)
	if err != nil {
		return nil, err
	}
	namespace, err := clientutil.GetLabelMatchValue("namespace", r.Matchers)
	if err != nil {
		return nil, err
	}
	podNames, err := clientutil.GetLabelMatchValues("pod_name", r.Matchers)
	if err != nil {
		return nil, err
	}
	vmIds, err := clientutil.GetLabelMatchValues("vm_id", r.Matchers)
	if err != nil {
		return nil, err
	}

	if len(vmIds) > 0 {
		for _, vmId := range vmIds {
			seriesQueryList = append(seriesQueryList, &SeriesQuery{
				podType:    podType,
				metricName: metricName,
				vmId:       vmId,
			})
		}
		return seriesQueryList, nil
	}

	if namespace != "" && len(podNames) > 0 {
		for _, podName := range podNames {
			seriesQueryList = append(seriesQueryList, &SeriesQuery{
				podType:    podType,
				metricName: metricName,
				namespace:  namespace,
				podName:    podName,
			})
		}
		return seriesQueryList, nil
	}

	return nil, errors.Errorf("namespace,pod_name or vm_id label not found")

}

// FetchAndSendSeries fetch and send series
func (p *Store) FetchAndSendSeries(r *storepb.SeriesRequest, s storepb.Store_SeriesServer,
	query *SeriesQuery) error {
	tagFilters := []*suanliclient.TagFilter{}
	pbLabels := []prompb.Label{
		{Name: "__name__", Value: query.metricName},
		{Name: "cluster_id", Value: p.config.ClusterID},
		{Name: "provider", Value: provider},
	}

	if query.namespace == "" || query.podName == "" {
		return errors.Errorf("namespace,pod_name label not found")
	}
	tagFilters = append(tagFilters, &suanliclient.TagFilter{Key: "pod_name", Value: []string{query.podName}})
	tagFilters = append(tagFilters, &suanliclient.TagFilter{Key: "namespace", Value: []string{query.namespace}})
	pbLabels = append(pbLabels, []prompb.Label{
		{Name: "namespace", Value: query.namespace},
		{Name: "pod_name", Value: query.podName},
	}...)

	data, err := p.suanLiClient.QueryInfo(s.Context(), query.metricName, tagFilters, r.MinTime, r.MaxTime)
	if err != nil {
		return err
	}

	for _, v := range data.Data.ChartInfoList {
		series, _ := v.ToSeries(pbLabels, p.suanLiClient.Loc())
		if err := p.SendSeries(series, s, nil, nil); err != nil {
			return err
		}
	}

	return nil
}

// Series returns all series for a requested time range and label matcher.
func (p *Store) Series(r *storepb.SeriesRequest, s storepb.Store_SeriesServer) error {
	seriesQueryList, err := p.GetSeriesQueryList(s.Context(), r)
	if err != nil {
		return err
	}

	var (
		wg              sync.WaitGroup
		seriesQueryChan = make(chan *SeriesQuery)
		multiErrors     *multierror.Error
		mtx             sync.Mutex
	)

	expectConcurrency := concurrency
	if expectConcurrency > len(seriesQueryList) {
		expectConcurrency = len(seriesQueryList)
	}

	for i := 0; i < expectConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for seriesQuery := range seriesQueryChan {
				err := p.FetchAndSendSeries(r, s, seriesQuery)
				if err != nil {
					mtx.Lock()
					multiErrors = multierror.Append(multiErrors, errors.Wrapf(err,
						fmt.Sprintf("fetch %s", seriesQuery)))
					mtx.Unlock()
				}
			}
		}()
	}

	for _, seriesQuery := range seriesQueryList {
		seriesQueryChan <- seriesQuery
	}

	close(seriesQueryChan)
	wg.Wait()

	// 如果全部错误, 直接返回异常请求
	if len(multiErrors.WrappedErrors()) == len(seriesQueryList) {
		return multiErrors.ErrorOrNil()
	}

	// 部分错误, 返回 warning 信息
	if len(multiErrors.WrappedErrors()) > 0 {
		_ = s.Send(storepb.NewWarnSeriesResponse(multiErrors.ErrorOrNil()))
	}

	return nil
}
