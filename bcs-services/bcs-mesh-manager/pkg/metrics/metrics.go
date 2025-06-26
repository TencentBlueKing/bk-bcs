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

// Package metrics xxx
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// BkBcsHelmManager xxx
	BkBcsHelmManager = "bkbcs_helmmanager"
)

var (
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsHelmManager,
		Name:      "api_request_total_num",
		Help:      "The total number of requests for helm manager api",
	}, []string{"handler", "status"})

	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsHelmManager,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for helm manager api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "status"})

	operationTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsHelmManager,
		Name:      "operation_total_num",
		Help:      "The total number of helm manager release operation",
	}, []string{"action", "status"})

	operationLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsHelmManager,
		Name:      "operation_latency_time",
		Help:      "latency statistic for helm manager release operation",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0, 15.0, 20.0, 30.0, 600.0},
	}, []string{"action", "status"})

	operationCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: BkBcsHelmManager,
		Name:      "operation_current_count",
		Help:      "The current count of operation for helm manager release",
	})
)

func init() {
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
	prometheus.MustRegister(operationTotal)
	prometheus.MustRegister(operationLatency)
	prometheus.MustRegister(operationCount)
}

// ReportAPIRequestMetric report api request metrics
func ReportAPIRequestMetric(handler, status string, started time.Time) {
	requestTotalAPI.WithLabelValues(handler, status).Inc()
	requestLatencyAPI.WithLabelValues(handler, status).Observe(time.Since(started).Seconds())
}

// ReportOperationMetric report operation metrics
func ReportOperationMetric(action, status string, started time.Time) {
	operationTotal.WithLabelValues(action, status).Inc()
	operationLatency.WithLabelValues(action, status).Observe(time.Since(started).Seconds())
}

// ReportOperationCountMetric report operation count metrics
func ReportOperationCountMetric(count int32) {
	operationCount.Set(float64(count))
}
