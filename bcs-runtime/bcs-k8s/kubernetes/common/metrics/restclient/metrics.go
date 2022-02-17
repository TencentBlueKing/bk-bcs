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

package restclient

import (
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/metrics"
)

const (
	// BkBcsRestClient for bkbcs call rest client prefix
	BkBcsRestClient = "bkbcs_rest_client"
)

var (
	requestResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: BkBcsRestClient,
			Name:      "requests_total",
			Help:      "Number of HTTP requests, partitioned by status code, method, and host.",
		},
		[]string{"code", "method", "host"},
	)

	// requestLatency is a Prometheus Summary metric type partitioned by
	// "verb" and "url" labels. It is used for the rest client latency metrics.
	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: BkBcsRestClient,
			Name:      "request_duration_seconds",
			Help:      "Request latency in seconds. Broken down by verb and URL.",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
		},
		[]string{"verb", "url"},
	)

	rateLimiterLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: BkBcsRestClient,
			Name:      "rate_limiter_duration_seconds",
			Help:      "Client side rate limiter latency in seconds. Broken down by verb and URL.",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
		},
		[]string{"verb", "url"},
	)
)

func init() {

	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(rateLimiterLatency)
	prometheus.MustRegister(requestResult)
	opts := metrics.RegisterOpts{
		RequestLatency: &latencyAdapter{m: requestLatency},
		RequestResult:  &resultAdapter{m: requestResult},
	}
	metrics.Register(opts)
}

type latencyAdapter struct {
	m *prometheus.HistogramVec
}

func (l *latencyAdapter) Observe(verb string, u url.URL, latency time.Duration) {
	l.m.WithLabelValues(verb, u.String()).Observe(latency.Seconds())
}

type resultAdapter struct {
	m *prometheus.CounterVec
}

func (r *resultAdapter) Increment(code, method, host string) {
	r.m.WithLabelValues(code, method, host).Inc()
}
