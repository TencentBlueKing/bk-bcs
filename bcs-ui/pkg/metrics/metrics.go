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

// Package metrics xxx
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// http 请求总量
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "http_requests_total",
		Help:      "Counter of HTTP requests to bcs-ui.",
	}, []string{"handler", "method", "code"})

	// http 请求耗时, 包含页面返回, API请求
	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "http_request_duration_seconds",
		Help:      "Histogram of latencies for HTTP requests bcs-ui.",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 30, 60},
	}, []string{"handler", "method", "code"})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// collectHTTPRequestMetric http metrics 处理
func collectHTTPRequestMetric(handler, method, code string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(handler, method, code).Inc()
	httpRequestDuration.WithLabelValues(handler, method, code).Observe(duration.Seconds())
}
