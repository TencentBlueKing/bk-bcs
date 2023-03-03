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

package bkmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/thanos-io/thanos/pkg/store/storepb"
)

var tracer = otel.Tracer("bk_monitor_client")

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

// QueryByPromQL unifyquery 查询, promql 语法
// start, end, step 单位秒
func QueryByPromQL(ctx context.Context, rawURL, bkBizID string, start, end, step int64,
	labelMatchers []storepb.LabelMatcher, rawPromql string) ([]*prompb.TimeSeries, error) {
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
	bodyStr, _ := json.Marshal(body)
	commonAttrs := []attribute.KeyValue{
		attribute.String("body", string(bodyStr)),
	}
	ctx, span := tracer.Start(ctx, "QueryByPromQL", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetBody(body).
		SetHeader("X-Bk-Scope-Space-Uid", fmt.Sprintf("bkcc__%s", bkBizID)). // 支持空间参数
		SetQueryParam("bk_app_code", config.G.Base.AppCode).
		SetQueryParam("bk_app_secret", config.G.Base.AppSecret).
		Post(url)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		err := errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// 部分接口，如 usermanager 返回的content-type不是json, 需要手动Unmarshal
	result := new(BKUnifyQueryResult)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	resultSet, err := result.ToPromSeriesSet()
	resultSetStr, _ := json.Marshal(resultSet)
	// 设置额外标签
	span.SetAttributes(attribute.String("resultSet", string(resultSetStr)))

	return resultSet, err
}

// BKMonitorResult 蓝鲸监控返回的结构体, 和component下的BKResult数据接口规范不一致, 重新定义一份
type BKMonitorResult struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Result  bool             `json:"result"`
	Data    *GrayClusterList `json:"data"`
}

// GrayClusterList 灰度列表
type GrayClusterList struct {
	Enabled       bool                `json:"enable_bsc_gray_cluster"`
	ClusterIdList []string            `json:"bcs_gray_cluster_id_list"`
	ClusterMap    map[string]struct{} `json:"-"`
}

func (c *GrayClusterList) initClusterMap() {
	c.ClusterMap = map[string]struct{}{}
	for _, id := range c.ClusterIdList {
		c.ClusterMap[id] = struct{}{}
	}
}

// queryClusterList 查询已经接入蓝鲸监控的集群列表
func queryClusterList(ctx context.Context, host string) (*GrayClusterList, error) {
	url := fmt.Sprintf("%s/get_bcs_gray_cluster_list", host)
	commonAttrs := []attribute.KeyValue{
		attribute.String("url", url),
	}
	ctx, span := tracer.Start(ctx, "queryClusterList", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetQueryParam("bk_app_code", config.G.Base.AppCode).
		SetQueryParam("bk_app_secret", config.G.Base.AppSecret).
		Get(url)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		err := errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	bkMonitorResult := &BKMonitorResult{}
	if err := json.Unmarshal(resp.Body(), bkMonitorResult); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if !bkMonitorResult.Result {
		err := errors.Errorf("result = %t, shoud be true", bkMonitorResult.Result)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	bkMonitorResult.Data.initClusterMap()
	reslutStr, _ := json.Marshal(bkMonitorResult.Data)
	// 设置额外标签
	span.SetAttributes(attribute.String("bkMonitorResult.Data", string(reslutStr)))
	return bkMonitorResult.Data, nil
}

// QueryGrayClusterMap 查询灰度集群, 有缓存
func QueryGrayClusterMap(ctx context.Context, host string) (map[string]struct{}, error) {
	cacheKey := "bcs.QueryGrayClusterMap"
	commonAttrs := []attribute.KeyValue{
		attribute.String("cacheKey", cacheKey),
	}
	ctx, span := tracer.Start(ctx, "QueryGrayClusterMap", trace.WithSpanKind(trace.SpanKindInternal), trace.WithAttributes(commonAttrs...))
	defer span.End()
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		resultStr, _ := json.Marshal(cacheResult)
		span.SetAttributes(attribute.Key("cacheResult").String(string(resultStr)))
		return cacheResult.(map[string]struct{}), nil
	}

	clusterList, err := queryClusterList(ctx, host)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	grepClusterMap := map[string]struct{}{}

	for _, clusterId := range clusterList.ClusterIdList {
		grepClusterMap[clusterId] = struct{}{}
	}

	storage.LocalCache.Slot.Set(cacheKey, grepClusterMap, time.Minute*10)
	grepClusterMapStr, _ := json.Marshal(grepClusterMap)
	span.SetAttributes(attribute.Key("grepClusterMap").String(string(grepClusterMapStr)))
	return grepClusterMap, nil
}

// IsBKMonitorEnabled 集群是否接入到蓝鲸监控
func IsBKMonitorEnabled(ctx context.Context, clusterId string) (bool, error) {
	// 不配置则全量接入
	if len(config.G.BKMonitor.MetadataURL) == 0 {
		return true, nil
	}
	grayClusterMap, err := QueryGrayClusterMap(ctx, config.G.BKMonitor.MetadataURL)
	if err != nil {
		return false, err
	}

	_, ok := grayClusterMap[clusterId]
	return ok, nil
}
