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

// Package bkmonitor bk_monitor query
package bkmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/chonla/format"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"
	"github.com/thanos-io/thanos/pkg/store/storepb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const (
	defaultQueryPath = "/query/ts/promql"
)

// Sample 返回的点
type Sample struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// MarshalJSON 序列化interface
func (s Sample) MarshalJSON() ([]byte, error) {
	values := [2]interface{}{s.Timestamp, s.Value}

	return json.Marshal(values)
}

// UnmarshalJSON 反序列化interface
func (s *Sample) UnmarshalJSON(b []byte) error {
	values := [2]json.Number{}

	if err := json.Unmarshal(b, &values); err != nil {
		return err
	}

	t, err := values[0].Int64()
	if err != nil {
		return errors.Errorf("timestamp %s is invalid", values[0])
	}

	v, err := values[1].Float64()
	if err != nil {
		return errors.Errorf("value %s is invalid", values[1])
	}

	s.Timestamp = t
	s.Value = v

	return nil
}

// Series 蓝鲸监控返回数据
type Series struct {
	Name        string    `json:"name"`
	Columns     []string  `json:"columns"`
	Types       []string  `json:"types"`
	GroupKeys   []string  `json:"group_keys"`
	GroupValues []string  `json:"group_values"`
	Values      []*Sample `json:"values"`
}

// ToPromSeries 转换为 prom 时序
func (s *Series) ToPromSeries() (*prompb.TimeSeries, error) {
	if len(s.GroupValues) < len(s.GroupKeys) {
		return nil, errors.Errorf("len GroupValues(%d) < GroupKeys(%d)", len(s.GroupValues), len(s.GroupKeys))
	}

	labels := make([]prompb.Label, 0, len(s.GroupKeys))
	for idx, key := range s.GroupKeys {
		labels = append(labels, prompb.Label{
			Name:  key,
			Value: s.GroupValues[idx],
		})
	}

	samples := make([]prompb.Sample, 0, len(s.Values))

	for _, value := range s.Values {
		samples = append(samples, prompb.Sample{
			Timestamp: value.Timestamp,
			Value:     value.Value,
		})
	}

	promSeries := &prompb.TimeSeries{
		Labels:  labels,
		Samples: samples,
	}
	return promSeries, nil
}

// BKUnifyQueryResult 蓝鲸监控UnifyQuery返回结果
type BKUnifyQueryResult struct {
	Series []*Series `json:"series"`
}

// ToPromSeriesSet 转换为 prom 时序
func (r *BKUnifyQueryResult) ToPromSeriesSet() ([]*prompb.TimeSeries, error) {
	promSeriesSet := make([]*prompb.TimeSeries, 0, len(r.Series))
	for _, series := range r.Series {
		promSeries, err := series.ToPromSeries()
		if err != nil {
			return nil, err
		}
		promSeriesSet = append(promSeriesSet, promSeries)
	}
	return promSeriesSet, nil
}

// getQueryURL 兼容网关/内部k8s service的场景
func getQueryURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Path == "" {
		u.Path = defaultQueryPath
	}

	return u.String(), nil
}

// QueryByPromQLRaw unifyquery 查询, promql 语法
func QueryByPromQLRaw(ctx context.Context, rawURL, bkBizID string, start, end, step int64,
	labelMatchers []storepb.LabelMatcher, rawPromql string) (*BKUnifyQueryResult, error) {
	url, err := getQueryURL(rawURL)
	if err != nil {
		return nil, err
	}

	promql := storepb.MatchersToString(labelMatchers...)
	if rawPromql != "" {
		promql = rawPromql
	}

	// 步长, 单位秒
	stepSecond := fmt.Sprintf("%ss", strconv.FormatInt(step, 10))
	body := map[string]string{
		"promql": promql,
		"step":   stepSecond,
		"start":  strconv.FormatInt(start, 10),
		"end":    strconv.FormatInt(end, 10),
	}

	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}

	resp, err := component.GetNoTraceClient().R().
		SetContext(ctx).
		SetBody(body).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetHeader("X-Bk-Scope-Space-Uid", fmt.Sprintf("bkcc__%s", bkBizID)). // 支持空间参数
		SetHeaders(utils.GetLaneIDByCtx(ctx)).                               // 泳道特性
		Post(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := new(BKUnifyQueryResult)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryByPromQL unifyquery 查询, promql 语法
// start, end, step 单位秒
func QueryByPromQL(ctx context.Context, rawURL, bkBizID string, start, end, step int64,
	labelMatchers []storepb.LabelMatcher, rawPromql string) ([]*prompb.TimeSeries, error) {
	result, err := QueryByPromQLRaw(ctx, rawURL, bkBizID, start, end, step, labelMatchers, rawPromql)
	if err != nil {
		return nil, err
	}
	return result.ToPromSeriesSet()
}

// QueryMultiValues unifyquery 查询, promql 语法
func QueryMultiValues(ctx context.Context, rawURL, bkBizID string, start int64, promqlMap map[string]string,
	params map[string]interface{}) (map[string]string, error) {
	var (
		wg  sync.WaitGroup
		mtx sync.Mutex
	)

	defaultValue := ""

	resultMap := map[string]string{}

	// promql 数量已知, 不控制并发数量
	for k, v := range promqlMap {
		wg.Add(1)
		go func(key, promql string) {
			defer wg.Done()

			promql = format.Sprintf(promql, params)
			resutl, err := QueryByPromQLRaw(ctx, rawURL, bkBizID, start, start, 60, nil, promql)
			mtx.Lock()
			defer mtx.Unlock()

			// 多个查询不报错, 有默认值
			if err != nil {
				blog.Warnf("query_multi_values %s error, %s", promql, err)
				resultMap[key] = defaultValue
			} else {
				resultMap[key] = GetFirstValue(resutl.Series)
			}
		}(k, v)
	}

	wg.Wait()

	return resultMap, nil
}

// GetFirstValue 获取第一个值
func GetFirstValue(series []*Series) string {
	if len(series) == 0 {
		return ""
	}
	if len(series[0].Values) == 0 {
		return ""
	}
	return strconv.FormatFloat(series[0].Values[0].Value, 'f', -1, 64)
}

// BaseResponse base response
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  bool   `json:"result"`
}

// IsBKMonitorEnabled 集群是否接入到蓝鲸监控
func IsBKMonitorEnabled(clusterID string) bool {
	return !utils.StringInSlice(clusterID, config.G.BKMonitor.DisableClusters)
}

// MetricsListResult metrics 列表
type MetricsListResult struct {
	BaseResponse
	Data []MetricList `json:"data"`
}

// MetricList metrics list
type MetricList struct {
	Metric string        `json:"field_name"`
	Labels []MetricLabel `json:"dimensions"`
}

// MetricListSlice metrics list slice
type MetricListSlice []MetricList

// ToSeries trans metrics to series
func (m MetricListSlice) ToSeries() []*prompb.TimeSeries {
	series := make([]*prompb.TimeSeries, 0)
	for _, v := range m {
		labels := make([]prompb.Label, 0)
		labels = append(labels, prompb.Label{
			Name:  "__name__",
			Value: v.Metric,
		})
		for _, lb := range v.Labels {
			if lb.Value == "" {
				lb.Value = "string"
			}
			labels = append(labels, prompb.Label{
				Name:  lb.Key,
				Value: lb.Value,
			})
		}
		series = append(series, &prompb.TimeSeries{
			Labels: labels,
		})
	}
	return series
}

// MetricLabel metrics label
type MetricLabel struct {
	Key   string `json:"field_name"`
	Value string `json:"type"`
}

// GetMetricsList 获取 metrics 列表
func GetMetricsList(ctx context.Context, host, clusterID, bizID string) ([]MetricList, error) {
	cacheKey := fmt.Sprintf("bcs.QueryGrayClusterMap.%s", clusterID)
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.([]MetricList), nil
	}

	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/query_bcs_metrics", host)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetQueryParam("cluster_ids", clusterID).
		SetQueryString(fmt.Sprintf("bk_biz_ids=0&bk_biz_ids=%s", bizID)).
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}
	result := new(MetricsListResult)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return nil, err
	}
	storage.LocalCache.Slot.Set(cacheKey, result.Data, time.Minute*10)
	return result.Data, nil
}

// GetMetricsSeries 获取 metrics series 列表
func GetMetricsSeries(ctx context.Context, host, clusterID, bizID string) ([]*prompb.TimeSeries, error) {
	metrics, err := GetMetricsList(ctx, host, clusterID, bizID)
	if err != nil {
		return nil, err
	}
	return MetricListSlice(metrics).ToSeries(), nil
}

// ClusterDataIDResult 集群数据ID结果
type ClusterDataIDResult struct {
	BaseResponse
	Data map[string]*ClusterDataID `json:"data"`
}

// ClusterDataID 集群数据ID
type ClusterDataID struct {
	BkDataID        int    `json:"bk_data_id"`
	DataName        string `json:"data_name"`
	ResultTableID   string `json:"result_table_id"`
	VmResultTableID string `json:"vm_result_table_id"`
}

// GetClusterEventDataID 获取集群事件数据ID
func GetClusterEventDataID(ctx context.Context, host, clusterID string) (*ClusterDataID, error) {
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/metadata_query_bcs_related_data_link_info", host)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetQueryParam("bcs_cluster_id", clusterID).
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}
	result := new(ClusterDataIDResult)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 || result.Data["K8SEvent"] == nil {
		return nil, errors.New("no data")
	}

	return result.Data["K8SEvent"], nil
}
