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

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// ErrStatus for call api failure
	ErrStatus = "failure"
	// SucStatus for success status
	SucStatus = "success"
)

const (
	// BkBcsAlertManager for bcs-alert-manager module metrics prefix
	BkBcsAlertManager = "bkbcs_alertmanager"
)

var (
	// alert_manager grpc request action metrics
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsAlertManager,
		Name:      "api_request_total_num",
		Help:      "The total num of requests for alertmanager api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsAlertManager,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for alertmanager api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "method", "status"})

	// external http requests action metrics
	requestsTotalAlert = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsAlertManager,
		Name:      "lib_request_total_num",
		Help:      "The total number of requests to alert-system api",
	}, []string{"handler", "method", "status"})

	requestLatencyAlert = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsAlertManager,
		Name:      "lib_request_latency_time",
		Help:      "request latency time for call alert-system api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "method", "status"})

	// handler monitor metrics
	requestsTotalHandlerQueue = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsAlertManager,
		Name:      "handler_queue_total_num",
		Help:      "The total number of handler queueLen",
	}, []string{"handler"})

	requestLatencyHandler = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsAlertManager,
		Name:      "handler_request_latency_time",
		Help:      "request latency time for queue parse data",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "name", "status"})
)

func init() {
	// requests alert
	prometheus.MustRegister(requestsTotalAlert)
	prometheus.MustRegister(requestLatencyAlert)
	// alert-manager api
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
	// handler monitor
	prometheus.MustRegister(requestsTotalHandlerQueue)
	prometheus.MustRegister(requestLatencyHandler)
}

//ReportAlertAPIMetrics report all api action metrics
func ReportAlertAPIMetrics(handler, method, status string, started time.Time) {
	requestsTotalAlert.WithLabelValues(handler, method, status).Inc()
	requestLatencyAlert.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}

//ReportAlertManagerAPIMetrics report all api action metrics
func ReportAlertManagerAPIMetrics(handler, method, status string, started time.Time) {
	requestTotalAPI.WithLabelValues(handler, method, status).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}

// ReportHandlerQueueLength report handler chanQueue length
func ReportHandlerQueueLength(handler string, queueLen float64) {
	requestsTotalHandlerQueue.WithLabelValues(handler).Set(queueLen)
}

// ReportHandlerFuncLatency report handler func latency
func ReportHandlerFuncLatency(handler, name, status string, started time.Time) {
	requestLatencyHandler.WithLabelValues(handler, name, status).Observe(time.Since(started).Seconds())
}
