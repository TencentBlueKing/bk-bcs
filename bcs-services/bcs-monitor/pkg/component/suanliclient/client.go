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

// Package suanliclient suanli client
package suanliclient

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/prompb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
)

const (
	// LongDateTimeFormat long date time format
	LongDateTimeFormat = "2006-01-02 15:04:05"
	// LongDateTimeMminuteFormat long datetime minute format
	LongDateTimeMminuteFormat = "2006-01-02 15:04"
)

// Response response
type Response struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data *Data  `json:"data"`
}

// Data data
type Data struct {
	PageNum        int64             `json:"page_num"`
	MetricUnitDict map[string]string `json:"metric_unit_dict"`
	ChartInfoList  []*ChartInfo      `json:"chart_info"`
	TotalNum       int64             `json:"total_num"`
}

// ChartInfo chart info
type ChartInfo struct {
	Title          string        `json:"title"`
	DetailDataList []*DetailData `json:"detail_data_list"`
}

// TagFilter tag filter
type TagFilter struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}

// ToSeries to series
func (c *ChartInfo) ToSeries(labels []prompb.Label, loc *time.Location) (*prompb.TimeSeries, error) {
	series := &prompb.TimeSeries{Labels: labels, Samples: []prompb.Sample{}}

	for _, d := range c.DetailDataList {
		sample, err := d.ToSample(loc)
		if err != nil {
			continue
		}
		series.Samples = append(series.Samples, *sample)
	}
	return series, nil
}

// DetailData detail data
type DetailData struct {
	Current *float64 `json:"current"`
	Time    string   `json:"time"`
}

// ToSample to sample
func (d *DetailData) ToSample(loc *time.Location) (*prompb.Sample, error) {
	// 格式 "2021-12-12 00:00"
	var layout string

	if d.Current == nil {
		return nil, errors.New("current is null")
	}

	switch len(d.Time) {
	case 19: // LONG_DATETIME_FORMAT 长度
		layout = LongDateTimeFormat
	case 16: // LONG_DATETIME_MINUTE_FORMAT 长度
		layout = LongDateTimeMminuteFormat
	}

	if layout == "" {
		return nil, errors.New("time layout invalid")
	}

	t, err := time.ParseInLocation(layout, d.Time, loc)
	if err != nil {
		return nil, err
	}

	return &prompb.Sample{Value: *d.Current, Timestamp: t.Unix() * 1000}, nil
}

// SuanLiConfig suanli config
type SuanLiConfig struct {
	ProjectName string `yaml:"project_name"`
	Host        string `yaml:"host"`
	Token       string `yaml:"token"`
	AppMark     string `yaml:"app_mark"`
	Env         string `yaml:"env"`
}

// SuanLiClient xxx
type SuanLiClient struct {
	config *SuanLiConfig
	loc    *time.Location
}

// NewSuanLiClient xxx
func NewSuanLiClient(config *SuanLiConfig) (*SuanLiClient, error) {
	// 使用北京时间
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}

	c := &SuanLiClient{
		config: config,
		loc:    loc,
	}
	return c, nil
}

// Loc xxx
func (c *SuanLiClient) Loc() *time.Location {
	return c.loc
}

// QueryInfo 分钟级监控数据
func (c *SuanLiClient) QueryInfo(ctx context.Context, metricName string, tagFilters []*TagFilter, startTime,
	endTime int64) (*Response, error) {

	st := time.Unix(0, startTime*int64(time.Millisecond))
	et := time.Unix(0, endTime*int64(time.Millisecond))

	rawURL := c.config.Host + "/monitor/v2/api/chart/info/query"
	jsonData := map[string]interface{}{
		"app_mark":    c.config.AppMark, // 应用名称
		"env":         c.config.Env,     // 环境
		"is_english":  "yes",
		"is_together": true,
		"tag_set":     tagFilters,
		"metric_name": metricName,
		"begin_time":  st.In(c.loc).Format(LongDateTimeFormat), // 格式 "2021-12-12 00:00:00"
		"end_time":    et.In(c.loc).Format(LongDateTimeFormat),
		"gap":         1, // 时间粒度, 单位分钟
	}

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("projectname", c.config.ProjectName).
		SetHeader("token", c.config.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(jsonData).
		Post(rawURL)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	result := &Response{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// QuerySecondInfo 秒级监控数据
func (c *SuanLiClient) QuerySecondInfo(ctx context.Context, metricName string, tagFilters []*TagFilter, startTime,
	endTime int64) (*Response, error) {
	st := time.Unix(0, startTime*int64(time.Millisecond))
	et := time.Unix(0, endTime*int64(time.Millisecond))

	rawURL := c.config.Host + "/monitor/v1/api/chart/second/info/query"
	jsonData := map[string]interface{}{
		"app_mark":    c.config.AppMark, // 应用名称
		"env":         c.config.Env,     // 环境
		"is_english":  "yes",
		"is_together": true,
		"tag_set":     tagFilters,
		"metric_name": metricName,
		"begin_time":  st.In(c.loc).Format(LongDateTimeFormat), // 格式 "2021-12-12 00:00:00"
		"end_time":    et.In(c.loc).Format(LongDateTimeFormat),
	}

	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("projectname", c.config.ProjectName).
		SetHeader("token", c.config.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(jsonData).
		Post(rawURL)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	result := &Response{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
