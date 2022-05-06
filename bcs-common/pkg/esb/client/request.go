/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	// SchemeHTTP HTTP scheme prefix
	SchemeHTTP = "http://"
	// SchemeHTTPS HTTPS scheme prefix
	SchemeHTTPS = "https://"
)

// Request for http request
type Request struct {
	// http client with metric and ratelimiter
	client *RESTClient

	// method http method
	method string

	// params for url
	params url.Values

	// http header
	headers http.Header

	// http body, map or json
	mapBody  map[string]interface{}
	jsonBody interface{}

	// http scheme
	scheme string

	// endpoints
	endpoints []string

	// url prefix
	baseURL string

	// url path
	subPath string

	// path parameters
	subPathArgs []interface{}

	// http timeout
	timeout time.Duration

	err error
}

// WithParams add params
func (r *Request) WithParams(params map[string]string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	for paramName, value := range params {
		r.params[paramName] = append(r.params[paramName], value)
	}
	return r
}

// WithParamsFromURL set params from url
func (r *Request) WithParamsFromURL(u *url.URL) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	params := u.Query()
	for paramName, value := range params {
		r.params[paramName] = append(r.params[paramName], value...)
	}
	return r
}

// WithParam add param
func (r *Request) WithParam(paramName, value string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params[paramName] = append(r.params[paramName], value)
	return r
}

// WithEndpoints set endpoints
func (r *Request) WithEndpoints(endpoints []string) *Request {
	r.endpoints = endpoints
	return r
}

// WithHeaders with http headers
func (r *Request) WithHeaders(header http.Header) *Request {
	if r.headers == nil {
		r.headers = header
		return r
	}

	for key, values := range header {
		for _, v := range values {
			r.headers.Add(key, v)
		}
	}
	return r
}

// WithBasePath set base path
func (r *Request) WithBasePath(basePath string) *Request {
	r.baseURL = basePath
	return r
}

// WithTimeout with http timeout
func (r *Request) WithTimeout(d time.Duration) *Request {
	r.timeout = d
	return r
}

// SubPathf set sub path
func (r *Request) SubPathf(subPath string, args ...interface{}) *Request {
	r.subPathArgs = args
	return r.subResource(subPath)
}

func (r *Request) subResource(subPath string) *Request {
	subPath = strings.TrimLeft(subPath, "/")
	r.subPath = subPath
	return r
}

// Body set http body
func (r *Request) Body(body map[string]interface{}) *Request {
	r.mapBody = body
	return r
}

// WithJSON set http json body
func (r *Request) WithJSON(jsonBody interface{}) *Request {
	r.jsonBody = jsonBody
	return r
}

func (r *Request) getBody() interface{} {
	if r.mapBody != nil {
		if len(r.client.credential) > 0 {
			for key, obj := range r.client.credential {
				r.mapBody[key] = obj
			}
		}
		return r.mapBody
	}
	if r.jsonBody != nil {
		return r.jsonBody
	}
	return make(map[string]string)
}

// WrapURL use subPath args and params to fill url
func (r *Request) WrapURL() *url.URL {
	finalURL := &url.URL{}
	if len(r.baseURL) != 0 {
		u, err := url.Parse(r.baseURL)
		if err != nil {
			r.err = err
			return new(url.URL)
		}
		*finalURL = *u
	}

	if len(r.subPathArgs) > 0 {
		finalURL.Path = finalURL.Path + fmt.Sprintf(r.subPath, r.subPathArgs...)
	} else {
		finalURL.Path = finalURL.Path + r.subPath
	}

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	finalURL.RawQuery = query.Encode()
	return finalURL
}

const maxLatency = 100 * time.Millisecond

func (r *Request) tryThrottle(url string) {
	now := time.Now()
	if r.client.throttle != nil {
		r.client.throttle.Accept()
	}

	if latency := time.Since(now); latency > maxLatency {
		blog.Infof("Throttling request took %d ms, method: %s, request: %s", latency, r.method, url)
	}
}

// Do do http request
func (r *Request) Do() *Result {
	result := new(Result)

	if requestInflight != nil {
		requestInflight.Inc()
		defer requestInflight.Dec()
	}
	if requestDuration != nil {
		startTime := time.Now()
		defer func() {
			requestDuration.WithLabelValues(r.method, strconv.Itoa(result.StatusCode)).Observe(
				float64(time.Since(startTime).Milliseconds()))
		}()
	}

	if len(r.endpoints) == 0 {
		result.Err = fmt.Errorf("lack endpoints for http client")
		return result
	}

	if r.client.tlsConf != nil {
		r.scheme = SchemeHTTPS
	} else {
		r.scheme = SchemeHTTP
	}

	maxRetryCycle := 3
	inds := generateRandomList(0, len(r.endpoints), len(r.endpoints))
	for try := 0; try < maxRetryCycle; try++ {
		for i, ind := range inds {
			var host string
			if r.client.randomAccess {
				host = r.endpoints[ind]
			} else {
				host = r.endpoints[i]
			}
			url := r.scheme + host + r.WrapURL().Path

			r.tryThrottle(url)

			body := r.getBody()
			bodyData, err := json.Marshal(body)
			if err != nil {
				result.Err = fmt.Errorf("invalid body")
				return result
			}

			blog.V(2).Infof("do request to url %s\n", url)
			resp, err := r.client.httpCli.RequestEx(url, r.method, r.headers, bodyData)
			if err != nil {
				blog.Errorf("RESTClient method:%s url:%s err %s", r.method, url, err.Error())
				// retry now
				time.Sleep(200 * time.Millisecond)
				continue
			}

			result.Body = resp.Reply
			result.StatusCode = resp.StatusCode
			result.Status = resp.Status
			return result
		}
	}
	result.Err = fmt.Errorf("RESTClient unexpected error")
	return result
}

// Result http result
type Result struct {
	Body       []byte
	Err        error
	StatusCode int
	Status     string
}

// Into decode result
func (r *Result) Into(obj interface{}) error {
	if nil != r.Err {
		return r.Err
	}

	if 0 != len(r.Body) {
		err := json.Unmarshal(r.Body, obj)
		if nil != err {
			if r.StatusCode >= 300 {
				return fmt.Errorf("http request err: %s", string(r.Body))
			}
			blog.Errorf("invalid response body, unmarshal json failed, reply:%s, error:%s", r.Body, err.Error())
			return fmt.Errorf("http response err: %v, raw data: %s", err, r.Body)
		}
	} else if r.StatusCode >= 300 {
		return fmt.Errorf("http request failed: %s", r.Status)
	}
	return nil
}

func generateRandomList(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}

	nums := make([]int, 0)
	exists := make(map[int]struct{})
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn((end - start)) + start
		if _, ok := exists[num]; !ok {
			exists[num] = struct{}{}
			nums = append(nums, num)
		}
	}

	return nums
}
