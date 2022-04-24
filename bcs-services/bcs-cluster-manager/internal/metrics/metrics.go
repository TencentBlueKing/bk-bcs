/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// LibCallStatusErr error shows during lib call
	LibCallStatusErr = "failure"
	// LibCallStatusOK lib call successfully
	LibCallStatusOK = "success"
)

const (
	// BkBcsClusterManager for prometheus namespace
	BkBcsClusterManager = "bkbcs_clustermanager"
)

var (
	requestsTotalLib = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsClusterManager,
		Name:      "lib_request_total_num",
		Help:      "The total number of requests for cluster manager to call other system api",
	}, []string{"system", "handler", "method", "status"})
	requestLatencyLib = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsClusterManager,
		Name:      "lib_request_latency_time",
		Help:      "api request latency statistic for cluster manager to call other system",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"system", "handler", "method", "status"})

	requestsTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsClusterManager,
		Name:      "api_request_total_num",
		Help:      "The total number of requests for cluster manager api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsClusterManager,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for cluster manager api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "method", "status"})
)

func init() {
	prometheus.MustRegister(requestsTotalLib)
	prometheus.MustRegister(requestLatencyLib)
	prometheus.MustRegister(requestsTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
}

// ReportLibRequestMetric report lib call metrics
func ReportLibRequestMetric(system, handler, method, status string, started time.Time) {
	requestsTotalLib.WithLabelValues(system, handler, method, status).Inc()
	requestLatencyLib.WithLabelValues(system, handler, method, status).Observe(time.Since(started).Seconds())
}

// ReportAPIRequestMetric report api request metrics
func ReportAPIRequestMetric(handler, method, status string, started time.Time) {
	requestsTotalAPI.WithLabelValues(handler, method, status).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}
