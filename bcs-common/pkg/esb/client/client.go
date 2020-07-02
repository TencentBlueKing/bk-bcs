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
	"crypto/tls"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"

	"github.com/prometheus/client_golang/prometheus"
)

type Credential map[string]interface{}

// RESTClient client with metrics, ratelimit and
type RESTClient struct {
	// TODO: change to golang http client, because HttpClient does not have context
	httpCli    *httpclient.HttpClient
	isTLS      bool
	throttle   throttle.RateLimiter
	credential Credential

	requestDuration *prometheus.HistogramVec
	requestInflight prometheus.Gauge
}

// NewRESTClient create rest client
func NewRESTClient() *RESTClient {
	client := &RESTClient{
		httpCli: httpclient.NewHttpClient(),
	}
	client.initMetrics()
	return client
}

// NewRESTClientWithTLS create rest client with tls
func NewRESTClientWithTLS(conf *tls.Config) *RESTClient {
	client := &RESTClient{
		httpCli: httpclient.NewHttpClient(),
	}
	client.httpCli.SetTlsVerityConfig(conf)
	client.isTLS = true
	return client
}

// WithRateLimiter set rate limiter
func (r *RESTClient) WithRateLimiter(th throttle.RateLimiter) *RESTClient {
	if th != nil {
		r.throttle = th
	}
	return r
}

// WithCredential set credential
func (r *RESTClient) WithCredential(c Credential) *RESTClient {
	if c != nil {
		r.credential = c
	}
	return r
}

func (r *RESTClient) initMetrics() {
	r.requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "bcs_paas_requests_duration_millisecond",
		Help: "third party api request duration millisecond",
	}, []string{"handler", "status_code"})
	prometheus.Register(r.requestDuration)

	r.requestInflight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bcs_requests_in_flight",
		Help: "third party api request number in flight",
	})
	prometheus.Register(r.requestInflight)
}

// Post create post request
func (r *RESTClient) Post() *Request {
	return &Request{
		client: r,
		method: http.MethodPost,
	}
}

// Put create put request
func (r *RESTClient) Put() *Request {
	return &Request{
		client: r,
		method: http.MethodPut,
	}
}

// Get create get request
func (r *RESTClient) Get() *Request {
	return &Request{
		client: r,
		method: http.MethodGet,
	}
}

// Delete create delete request
func (r *RESTClient) Delete() *Request {
	return &Request{
		client: r,
		method: http.MethodDelete,
	}
}

// Patch create patch request
func (r *RESTClient) Patch() *Request {
	return &Request{
		client: r,
		method: http.MethodPatch,
	}
}

// Head create head request
func (r *RESTClient) Head() *Request {
	return &Request{
		client: r,
		method: http.MethodHead,
	}
}
