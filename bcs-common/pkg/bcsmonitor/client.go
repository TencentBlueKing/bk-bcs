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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
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
}

// BcsMonitorClient is the client for bcs monitor request
type BcsMonitorClient struct {
	opts             BcsMonitorClientOpt
	defaultHeader    http.Header
	completeEndpoint string
	requestClient    Requester
}

// BcsMonitorClientOpt is the opts
type BcsMonitorClientOpt struct {
	Schema   string
	Endpoint string
	UserName string //basic auth username
	Password string //basic auth password
}

//Requester is the interface to do request
type Requester interface {
	DoRequest(url, method string, header http.Header, data []byte) ([]byte, error)
}

type requester struct {
	httpCli *httpclient.HttpClient
}

func (r *requester) DoRequest(url, method string, header http.Header, data []byte) ([]byte, error) {
	rsp, err := r.httpCli.Request(url, method, header, data)
	if err != nil {
		blog.Errorf("do request error, url: %s, error:%v", url, err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return rsp, nil
}

func NewRequester() Requester {
	return &requester{
		httpCli: httpclient.NewHttpClient(),
	}
}

// NewBcsMonitorClient new a BcsMonitorClient
func NewBcsMonitorClient(opts BcsMonitorClientOpt, r Requester) *BcsMonitorClient {
	return &BcsMonitorClient{
		opts:          opts,
		requestClient: r,
	}
}

// SetDefaultHeader set default headers
func (c *BcsMonitorClient) SetDefaultHeader(h http.Header) {
	c.defaultHeader = h
}

// SetCompleteEndpoint set complete endpoint
func (c *BcsMonitorClient) SetCompleteEndpoint() {
	if c.opts.UserName != "" && c.opts.Password != "" {
		c.completeEndpoint = fmt.Sprintf("%s://%s:%s@%s", c.opts.Schema,
			c.opts.UserName, c.opts.Password, c.opts.Endpoint)
	} else {
		c.completeEndpoint = fmt.Sprintf("%s://%s", c.opts.Schema, c.opts.Endpoint)
	}
}

// LabelValues get label values
// labelName is essential
// selectors, startTime, endTime optional
func (c *BcsMonitorClient) LabelValues(labelName string, selectors []string,
	startTime, endTime time.Time) (*LabelResponse, error) {
	var queryString string
	if selectors != nil && len(selectors) != 0 {
		queryString = c.setSelectors(queryString, selectors)
	}
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	url := fmt.Sprintf("%s%s", c.completeEndpoint, fmt.Sprintf(LabelValuesPath, labelName))
	if queryString != "" {
		url = fmt.Sprintf("%s?%s", url, queryString)
	}
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &LabelResponse{}
	err = json.Unmarshal(rsp, &result)
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
	if selectors != nil && len(selectors) != 0 {
		queryString = c.setSelectors(queryString, selectors)
	}
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	url := fmt.Sprintf("%s%s", c.completeEndpoint, LabelsPath)
	if queryString != "" {
		url = fmt.Sprintf("%s?%s", url, queryString)
	}
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &LabelResponse{}
	err = json.Unmarshal(rsp, &result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// Query get instant vectors
// promql essential, time optional
func (c *BcsMonitorClient) Query(promql string, time time.Time) (*QueryResponse, error) {
	var queryString string
	queryString = c.setQuery(queryString, "query", promql)
	if !time.IsZero() {
		queryString = c.setQuery(queryString, "time", fmt.Sprintf("%d", time.Unix()))
	}
	url := fmt.Sprintf("%s%s?%s", c.completeEndpoint, QueryPath, queryString)
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &QueryResponse{}
	err = json.Unmarshal(rsp, &result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// QueryByPost get instant vectors
// promql essential, time optional
// You can use QueryByPost when the promql is very long that may breach server-side URL character limits.
func (c *BcsMonitorClient) QueryByPost(promql string, time time.Time) (*QueryResponse, error) {
	var queryString string
	queryString = c.setQuery(queryString, "query", promql)
	if !time.IsZero() {
		queryString = c.setQuery(queryString, "time", fmt.Sprintf("%d", time.Unix()))
	}
	url := fmt.Sprintf("%s%s", c.completeEndpoint, QueryPath)
	header := c.defaultHeader
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	rsp, err := c.requestClient.DoRequest(url, "POST", header, []byte(queryString))
	if err != nil {
		return nil, err
	}
	result := &QueryResponse{}
	err = json.Unmarshal(rsp, &result)
	if err != nil {
		blog.Errorf("json unmarshal error:%v", err)
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return result, nil
}

// QueryRange get range vectors
// promql, startTime, endTime, step essential
func (c *BcsMonitorClient) QueryRange(promql string, startTime, endTime time.Time,
	step time.Duration) (*QueryRangeResponse, error) {
	var queryString string
	queryString = c.setQuery(queryString, "query", promql)
	queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	queryString = c.setQuery(queryString, "step", step.String())
	url := fmt.Sprintf("%s%s?%s", c.completeEndpoint, QueryRangePath, queryString)
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &QueryRangeResponse{}
	err = json.Unmarshal(rsp, &result)
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
	queryString = c.setQuery(queryString, "query", promql)
	queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	queryString = c.setQuery(queryString, "step", step.String())
	url := fmt.Sprintf("%s%s", c.completeEndpoint, QueryRangePath)
	header := c.defaultHeader
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	rsp, err := c.requestClient.DoRequest(url, "POST", c.defaultHeader, []byte(queryString))
	if err != nil {
		return nil, err
	}
	result := &QueryRangeResponse{}
	err = json.Unmarshal(rsp, &result)
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
	queryString = c.setSelectors(queryString, selectors)
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	url := fmt.Sprintf("%s%s?%s", c.completeEndpoint, SeriesPath, queryString)
	rsp, err := c.requestClient.DoRequest(url, "GET", c.defaultHeader, nil)
	if err != nil {
		return nil, err
	}
	result := &SeriesResponse{}
	err = json.Unmarshal(rsp, &result)
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
	queryString = c.setSelectors(queryString, selectors)
	if !startTime.IsZero() {
		queryString = c.setQuery(queryString, "start", fmt.Sprintf("%d", startTime.Unix()))
	}
	if !endTime.IsZero() {
		queryString = c.setQuery(queryString, "end", fmt.Sprintf("%d", endTime.Unix()))
	}
	url := fmt.Sprintf("%s%s", c.completeEndpoint, SeriesPath)
	header := c.defaultHeader
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	rsp, err := c.requestClient.DoRequest(url, "POST", header, []byte(queryString))
	if err != nil {
		return nil, err
	}
	result := &SeriesResponse{}
	err = json.Unmarshal(rsp, &result)
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
