/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcsmonitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/requester"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// ClientInterface the interface of bcs monitor client
type ClientInterface interface {
	LabelValues(labelName string, selectors []string, startTime, endTime time.Time) (*LabelResponse, error)
	Labels(selectors []string, startTime, endTime time.Time) (*LabelResponse, error)
	Query(promql string, time time.Time) (*QueryResponse, error)
	QueryByPost(promql string, time time.Time) (*QueryResponse, error)
	QueryRange(promql string, startTime, endTime time.Time, step time.Duration) (*QueryRangeResponse, error)
	QueryRangeByPost(promql string, startTime, endTime time.Time, step time.Duration) (*QueryRangeResponse, error)
	Series(selectors []string, startTime, endTime time.Time) (*SeriesResponse, error)
	SeriesByPost(selectors []string, startTime, endTime time.Time) (*SeriesResponse, error)
	GetBKMonitorGrayClusterList() (map[string]bool, error)
	CheckIfBKMonitor(clusterID string) (bool, error)
}

// BcsMonitorClient is the client for bcs monitor request
type BcsMonitorClient struct {
	opts                 BcsMonitorClientOpt
	defaultHeader        http.Header
	requestClient        requester.Requester
	grayClusterListCache *cache.Cache
}

// BcsMonitorClientOpt is the opts
type BcsMonitorClientOpt struct {
	Endpoint  string
	AppCode   string
	AppSecret string
}

// NewBcsMonitorClient new a BcsMonitorClient
func NewBcsMonitorClient(opts BcsMonitorClientOpt, r requester.Requester) *BcsMonitorClient {
	return &BcsMonitorClient{
		opts:                 opts,
		requestClient:        r,
		grayClusterListCache: cache.New(time.Minute*10, time.Minute*120),
	}
}

// SetDefaultHeader set default headers
func (c *BcsMonitorClient) SetDefaultHeader(h http.Header) {
	c.defaultHeader = h
}

// LabelValues get label values
// labelName is essential
// selectors, startTime, endTime optional
func (c *BcsMonitorClient) LabelValues(labelName string, selectors []string,
	startTime, endTime time.Time) (*LabelResponse, error) {
	var queryString string
	var err error
	if len(selectors) != 0 {
		queryString = c.setSelectors(queryString, selectors)
	}
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	url := fmt.Sprintf("%s%s", c.opts.Endpoint, fmt.Sprintf(LabelValuesPath, labelName))
	if queryString != "" {
		url = fmt.Sprintf("%s?%s", url, queryString)
	}
	url = c.addAppMessage(url)
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "LabelValues", "GET", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &LabelResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// Labels get labels
// selectors, startTime, endTime optional
func (c *BcsMonitorClient) Labels(selectors []string, startTime, endTime time.Time) (*LabelResponse, error) {
	var queryString string
	var err error
	if len(selectors) != 0 {
		queryString = c.setSelectors(queryString, selectors)
	}
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	url := fmt.Sprintf("%s%s", c.opts.Endpoint, LabelsPath)
	if queryString != "" {
		url = fmt.Sprintf("%s?%s", url, queryString)
	}
	url = c.addAppMessage(url)
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "Labels", "GET", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &LabelResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// Query get instant vectors
// promql essential, time optional
func (c *BcsMonitorClient) Query(promql string, requestTime time.Time) (*QueryResponse, error) {
	var queryString string
	var err error
	queryString = c.setQuery(queryString, "query", promql)
	if !requestTime.IsZero() {
		queryString = c.setQuery(queryString, "time", fmt.Sprintf("%d", requestTime.Unix()))
	}
	url := fmt.Sprintf("%s%s?%s", c.opts.Endpoint, QueryPath, queryString)
	url = c.addAppMessage(url)
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "Query", "GET", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &QueryResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// QueryByPost get instant vectors
// promql essential, time optional
// You can use QueryByPost when the promql is very long that may breach server-side URL character limits.
func (c *BcsMonitorClient) QueryByPost(promql string, requestTime time.Time) (*QueryResponse, error) {
	var queryString string
	var err error
	encodePromQl := url.QueryEscape(promql)
	queryString = c.setQuery(queryString, "query", encodePromQl)
	if !requestTime.IsZero() {
		queryString = c.setQuery(queryString, "time", fmt.Sprintf("%d", requestTime.Unix()))
	}
	requestUrl := fmt.Sprintf("%s%s", c.opts.Endpoint, QueryPath)
	header := c.defaultHeader.Clone()
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	requestUrl = c.addAppMessage(requestUrl)
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "QueryByPost", "POST", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(requestUrl, "POST", header, []byte(queryString))
	if err != nil {
		return nil, err
	}
	result := &QueryResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", requestUrl, err)
	}
	return result, nil
}

// QueryRange get range vectors
// promql, startTime, endTime, step essential
func (c *BcsMonitorClient) QueryRange(promql string, startTime, endTime time.Time,
	step time.Duration) (*QueryRangeResponse, error) {
	var queryString string
	var err error
	queryString = c.setQuery(queryString, "query", promql)
	queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	queryString = c.setQuery(queryString, "step", step.String())
	url := fmt.Sprintf("%s%s?%s", c.opts.Endpoint, QueryRangePath, queryString)
	url = c.addAppMessage(url)
	header := c.defaultHeader.Clone()
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "QueryRange", "GET", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(url, "GET", header, nil)
	if err != nil {
		return nil, err
	}
	result := &QueryRangeResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// QueryRangeByPost get range vectors
// promql, startTime, endTime, step essential
// You can use QueryRangeByPost when the promql is very long that may breach server-side URL character limits.
func (c *BcsMonitorClient) QueryRangeByPost(promql string, startTime, endTime time.Time,
	step time.Duration) (*QueryRangeResponse, error) {
	var queryString string
	var err error
	queryString = c.setQuery(queryString, "query", promql)
	queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	queryString = c.setQuery(queryString, "step", step.String())
	url := fmt.Sprintf("%s%s", c.opts.Endpoint, QueryRangePath)
	header := c.defaultHeader.Clone()
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	url = c.addAppMessage(url)
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "QueryRangeByPost", "POST", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(url, "POST", header, []byte(queryString))
	if err != nil {
		return nil, err
	}
	result := &QueryRangeResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// Series finding series by selectors
// selectors essential
// startTime, endTime optional
func (c *BcsMonitorClient) Series(selectors []string, startTime, endTime time.Time) (*SeriesResponse, error) {
	var queryString string
	var err error
	queryString = c.setSelectors(queryString, selectors)
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	header := c.defaultHeader.Clone()
	url := fmt.Sprintf("%s%s?%s", c.opts.Endpoint, SeriesPath, queryString)
	url = c.addAppMessage(url)
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "Series", "GET", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(url, "GET", header, nil)
	if err != nil {
		return nil, err
	}
	result := &SeriesResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// SeriesByPost finding series by selectors
// selectors essential
// startTime, endTime optional
// You can use SeriesByPost when the selectors is very large that may breach server-side URL character limits.
func (c *BcsMonitorClient) SeriesByPost(selectors []string, startTime, endTime time.Time) (*SeriesResponse, error) {
	var queryString string
	var err error
	queryString = c.setSelectors(queryString, selectors)
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	url := fmt.Sprintf("%s%s", c.opts.Endpoint, SeriesPath)
	header := c.defaultHeader.Clone()
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	url = c.addAppMessage(url)
	start := time.Now()
	defer func() {
		prom.ReportLibRequestMetric(prom.BkBcsMonitor, "SeriesByPost", "POST", err, start)
	}()
	rsp, err := c.requestClient.DoRequest(url, "POST", header, []byte(queryString))
	if err != nil {
		return nil, err
	}
	result := &SeriesResponse{}
	err = json.Unmarshal(rsp, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

func (c *BcsMonitorClient) setQuery(queryString, key, value string) string {
	if queryString != "" {
		return fmt.Sprintf("%s&%s=%s", queryString, key, value)
	}
	return fmt.Sprintf("%s=%s", key, value)
}

func (c *BcsMonitorClient) setSelectors(queryString string, selectors []string) string {
	for _, selector := range selectors {
		queryString = c.setQuery(queryString, "match[]", selector)
	}
	return queryString
}

func (c *BcsMonitorClient) addAppMessage(url string) string {
	if c.opts.AppCode != "" && c.opts.AppSecret != "" {
		addStr := fmt.Sprintf("app_code=%s&app_secret=%s", c.opts.AppCode, c.opts.AppSecret)
		url = fmt.Sprintf("%s?%s", url, addStr)
	}
	return url
}

// GetBKMonitorGrayClusterList get bk monitor gray cluster list
func (c *BcsMonitorClient) GetBKMonitorGrayClusterList() (map[string]bool, error) {
	if clusterMap, ok := c.grayClusterListCache.Get("grayClusterMap"); ok {
		grayClusterMap := clusterMap.(map[string]bool)
		return grayClusterMap, nil
	}
	url := fmt.Sprintf("%s/%s", c.opts.Endpoint, StorePath)
	response, err := c.requestClient.DoRequest(url, http.MethodGet, c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &StoreGWResponse{}
	err = json.Unmarshal(response, result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	grayClusterMap := make(map[string]bool)
	for _, query := range result.Data.Query {
		for _, labelSet := range query.LabelSets {
			if provider, ok := labelSet["provider"]; ok && provider == "BK_MONITOR" {
				clusterID := labelSet["cluster_id"]
				grayClusterMap[clusterID] = true
			}
		}
	}
	c.grayClusterListCache.Set("grayClusterMap", grayClusterMap, 1*time.Hour)
	return grayClusterMap, nil
}

// CheckIfBKMonitor check if bk monitor gray cluster
func (c *BcsMonitorClient) CheckIfBKMonitor(clusterID string) (bool, error) {
	if clusterMap, ok := c.grayClusterListCache.Get("grayClusterMap"); ok {
		grayClusterMap := clusterMap.(map[string]bool)
		if grayClusterMap[clusterID] == true {
			return true, nil
		}
		return false, nil
	}
	grayClusterMap, err := c.GetBKMonitorGrayClusterList()
	if err != nil {
		return false, fmt.Errorf("get bcs gray cluster list from bcs storegw err:%s", err.Error())
	}
	if grayClusterMap[clusterID] == true {
		return true, nil
	}
	return false, nil
}
