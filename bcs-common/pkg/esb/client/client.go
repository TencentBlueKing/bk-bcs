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

// Package client xxx
package client

import (
	"crypto/tls"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

var (
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "bcs_esb_requests_duration_millisecond",
		Help:    "esb api request duration millisecond",
		Buckets: []float64{10, 20, 40, 100, 150, 200, 400, 1000, 2000, 5000, 10000},
	}, []string{"method", "status_code"})

	requestInflight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bcs_esb_requests_in_flight",
		Help: "esb api request number in flight",
	})
)

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestInflight)
}

// Credential credential to be filled in post body
type Credential struct {
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
}

// RESTClient client with metrics, ratelimit and
type RESTClient struct {
	// Note: change to golang http client, because HttpClient does not have context
	httpCli    *httpclient.HttpClient
	tlsConf    *tls.Config
	throttle   throttle.RateLimiter
	credential Credential

	requestDuration *prometheus.HistogramVec // nolint
	requestInflight prometheus.Gauge         // nolint
	randomAccess    bool
}

// NewRESTClient create rest client
func NewRESTClient() *RESTClient {
	client := &RESTClient{
		httpCli: httpclient.NewHttpClient(),
	}
	return client
}

// NewRESTClientWithTLS create rest client with tls
func NewRESTClientWithTLS(conf *tls.Config) *RESTClient {
	client := &RESTClient{
		httpCli: httpclient.NewHttpClient(),
	}
	client.httpCli.SetTlsVerityConfig(conf)
	client.tlsConf = conf
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
func (r *RESTClient) WithCredential(appCode, appSecret string) *RESTClient {
	r.credential = Credential{
		AppCode:   appCode,
		AppSecret: appSecret,
	}
	return r
}

// WithTransport set transport
// Attention: transport should have non-nil TLSClientConfig if https is used
func (r *RESTClient) WithTransport(t *http.Transport) *RESTClient {
	if t == nil {
		return r
	}
	r.tlsConf = t.TLSClientConfig
	r.httpCli.SetTransPort(t)
	return r
}

// WithRandomAccess set random access endpoints
// Attention: transport should have non-nil TLSClientConfig if https is used
func (r *RESTClient) WithRandomAccess(set bool) *RESTClient {
	r.randomAccess = set
	return r
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
