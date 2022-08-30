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
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/thanos-io/thanos/pkg/store/storepb"
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

// QueryByPromQL unifyquery 查询, promql 语法
// start, end, step 单位秒
func QueryByPromQL(ctx context.Context, host string, bkBizId string, start, end, step int64,
	labelMatchers []storepb.LabelMatcher) ([]*prompb.TimeSeries, error) {
	url := fmt.Sprintf("%s/query/ts/promql", host)

	// 必须的参数 bk_biz_id, 单独拎出来处理
	bkBizIdMatcher := storepb.LabelMatcher{
		Type:  storepb.LabelMatcher_EQ,
		Name:  "bk_biz_id",
		Value: bkBizId,
	}
	promql := storepb.MatchersToString(append(labelMatchers, bkBizIdMatcher)...)

	// 步长, 单位秒
	stepSecond := fmt.Sprintf("%ss", strconv.FormatInt(step, 10))
	body := map[string]string{
		"promql": promql,
		"step":   stepSecond,
		"start":  strconv.FormatInt(start, 10),
		"end":    strconv.FormatInt(end, 10),
	}

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetBody(body).
		SetQueryParam("bk_app_code", config.G.Base.AppCode).
		SetQueryParam("bk_app_secret", config.G.Base.AppSecret).
		Post(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	// 部分接口，如 usermanager 返回的content-type不是json, 需要手动Unmarshal
	result := new(BKUnifyQueryResult)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return nil, err
	}

	return result.ToPromSeriesSet()
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

// QueryClusterList 查询已经接入蓝鲸监控的集群列表
func QueryClusterList(ctx context.Context, host string) (*GrayClusterList, error) {
	cacheKey := "bcs.QueryClusterList"
	if cacheResult, ok := storage.LocalCache.Slot.Get(cacheKey); ok {
		return cacheResult.(*GrayClusterList), nil
	}

	url := fmt.Sprintf("%s/get_bcs_gray_cluster_list", host)

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetQueryParam("bk_app_code", config.G.Base.AppCode).
		SetQueryParam("bk_app_secret", config.G.Base.AppSecret).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	bkMonitorResult := &BKMonitorResult{}
	if err := json.Unmarshal(resp.Body(), bkMonitorResult); err != nil {
		return nil, err
	}

	if !bkMonitorResult.Result {
		return nil, errors.Errorf("result = %t, shoud be true", bkMonitorResult.Result)
	}

	bkMonitorResult.Data.initClusterMap()
	storage.LocalCache.Slot.Set(cacheKey, bkMonitorResult.Data, time.Minute*10)

	return bkMonitorResult.Data, nil
}
