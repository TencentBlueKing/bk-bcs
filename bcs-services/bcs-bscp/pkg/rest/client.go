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

// Package rest provides http rest client.
package rest

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/client"
)

// ClientInterface http client interface.
type ClientInterface interface {
	Verb(verb VerbType) *Request
	Post() *Request
	Put() *Request
	Get() *Request
	Delete() *Request
	Patch() *Request
	Head() *Request
}

// NewClient get rest client.
func NewClient(c *client.Capability, baseURL string) ClientInterface {
	if baseURL != "/" {
		baseURL = strings.Trim(baseURL, "/")
		baseURL = "/" + baseURL + "/"
	}

	if c.ToleranceLatencyTime <= 0 {
		// set default tolerance latency time
		c.ToleranceLatencyTime = 2 * time.Second
	}

	client := &Client{
		baseURL:    baseURL,
		capability: c,
	}

	if c.MetricOpts.Register != nil {

		var buckets []float64
		if len(c.MetricOpts.DurationBuckets) == 0 {
			// set default buckets
			buckets = []float64{10, 30, 50, 70, 100, 200, 300, 400, 500, 1000, 2000, 5000}
		} else {
			// use user defined buckets
			buckets = c.MetricOpts.DurationBuckets
		}

		client.requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "bscp_http_requests_duration_millisecond",
			Help:    "third party api request duration millisecond.",
			Buckets: buckets,
		}, []string{"handler", "status_code", "dimension"})

		if err := c.MetricOpts.Register.Register(client.requestDuration); err != nil {
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok { //nolint:revive
				client.requestDuration = are.ExistingCollector.(*prometheus.HistogramVec)
			} else {
				panic(err)
			}
		}
	}

	return client
}

// Client http client.
type Client struct {
	// base url.
	baseURL string

	// client capability.
	capability *client.Capability

	// client detection.
	requestDuration *prometheus.HistogramVec
}

// Verb get request.
func (r *Client) Verb(verb VerbType) *Request {
	return &Request{
		client:     r,
		verb:       verb,
		baseURL:    r.baseURL,
		capability: r.capability,
	}
}

// Post post method.
func (r *Client) Post() *Request {
	return r.Verb(POST)
}

// Put put method.
func (r *Client) Put() *Request {
	return r.Verb(PUT)
}

// Get get method.
func (r *Client) Get() *Request {
	return r.Verb(GET)
}

// Delete delete method.
func (r *Client) Delete() *Request {
	return r.Verb(DELETE)
}

// Patch patch method.
func (r *Client) Patch() *Request {
	return r.Verb(PATCH)
}

// Head patch method.
func (r *Client) Head() *Request {
	return r.Verb(HEAD)
}
