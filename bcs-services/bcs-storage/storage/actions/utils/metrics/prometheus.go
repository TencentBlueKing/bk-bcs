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

// file prometheus.go for http request data
package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Config has the dependencies and values of the recorder.
type Config struct {
	// Prefix is the prefix that will be set on the metrics, by default it will be empty.
	Prefix string
	// DurationBuckets are the buckets used by Prometheus for the HTTP request duration metrics,
	// by default uses Prometheus default buckets (from 5ms to 10s).
	DurationBuckets []float64
	// SizeBuckets are the buckets used by Prometheus for the HTTP response size metrics,
	// by default uses a exponential buckets from 100B to 1GB.
	SizeBuckets []float64
	// Registry is the registry that will be used by the recorder to store the metrics,
	// if the default registry is not used then it will use the default one.
	Registry prometheus.Registerer
	// HandlerIDLabel is the name that will be set to the handler ID label, by default is `handler`.
	HandlerIDLabel string
	// StatusCodeLabel is the name that will be set to the status code label, by default is `code`.
	StatusCodeLabel string
	// MethodLabel is the name that will be set to the method label, by default is `method`.
	MethodLabel string
	// ClusterIDLabel is the name that will be set to the cluster_id label, default is `cluster_id`
	ClusterIDLabel string
	// ResourceTypeLabel is the name that will be set to the resource_type label, default is `resource_type`
	ResourceTypeLabel string
}

func (c *Config) defaults() {
	if len(c.DurationBuckets) == 0 {
		c.DurationBuckets = prometheus.DefBuckets
	}

	if len(c.SizeBuckets) == 0 {
		c.SizeBuckets = prometheus.ExponentialBuckets(100, 10, 8)
	}

	if c.HandlerIDLabel == "" {
		c.HandlerIDLabel = "handler"
	}

	if c.StatusCodeLabel == "" {
		c.StatusCodeLabel = "status"
	}

	if c.MethodLabel == "" {
		c.MethodLabel = "method"
	}

	if c.ClusterIDLabel == "" {
		c.ClusterIDLabel = "cluster_id"
	}

	if c.ResourceTypeLabel == "" {
		c.ResourceTypeLabel = "resource_type"
	}
}

type recorder struct {
	// Registry is the registry that will be used by the recorder to store the metrics,
	// if the default registry is not used then it will use the default one.
	registry prometheus.Registerer

	httpRequestDurHistogram   *prometheus.HistogramVec
	httpResponseSizeHistogram *prometheus.HistogramVec
	httpRequestsInflight      *prometheus.GaugeVec
	httpRequestCounter        *prometheus.CounterVec
}

// NewRecorder returns a new metrics recorder that implements the recorder
// using Prometheus as the backend.
func NewRecorder(cfg Config) Recorder {
	cfg.defaults()

	r := &recorder{
		httpRequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: cfg.Prefix,
			Name:      "api_request_total_num",
			Help:      "The total number of requests to bcs-storage api",
		}, []string{cfg.HandlerIDLabel, cfg.MethodLabel, cfg.StatusCodeLabel}),

		httpRequestDurHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Prefix,
			Name:      "api_request_duration_time",
			Help:      "The latency of the HTTP requests.",
			Buckets:   cfg.DurationBuckets,
		}, []string{cfg.HandlerIDLabel, cfg.MethodLabel, cfg.StatusCodeLabel}),

		httpResponseSizeHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Prefix,
			Name:      "api_response_size_bytes",
			Help:      "The size of the HTTP responses.",
			Buckets:   cfg.SizeBuckets,
		}, []string{cfg.HandlerIDLabel, cfg.MethodLabel, cfg.StatusCodeLabel}),

		httpRequestsInflight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: cfg.Prefix,
			Name:      "api_requests_inflight",
			Help:      "The number of inflight requests being handled at the same time.",
		}, []string{cfg.HandlerIDLabel, cfg.MethodLabel}),
	}

	if cfg.Registry == nil {
		r.registry = prometheus.DefaultRegisterer
	} else {
		r.registry = cfg.Registry
	}

	r.registry.MustRegister(
		r.httpRequestCounter,
		r.httpRequestDurHistogram,
		r.httpResponseSizeHistogram,
		r.httpRequestsInflight,
	)

	return r
}

// ObserveHTTPRequestCounterDuration report counter & latency metrics
func (r recorder) ObserveHTTPRequestCounterDuration(_ context.Context, p HTTPReqProperties, duration time.Duration) {
	r.httpRequestCounter.WithLabelValues(p.Handler, p.Method, p.Code).Inc()
	r.httpRequestDurHistogram.WithLabelValues(p.Handler, p.Method, p.Code).Observe(duration.Seconds())
}

// ObserveHTTPResponseSize report responseSize metrics
func (r recorder) ObserveHTTPResponseSize(_ context.Context, p HTTPReqProperties, sizeBytes int64) {
	r.httpResponseSizeHistogram.WithLabelValues(p.Handler, p.Method, p.Code).Observe(float64(sizeBytes))
}

// AddInflightRequests report inflight request being processed
func (r recorder) AddInflightRequests(_ context.Context, p HTTPProperties, quantity int) {
	r.httpRequestsInflight.WithLabelValues(p.Handler, p.Method).Add(float64(quantity))
}
