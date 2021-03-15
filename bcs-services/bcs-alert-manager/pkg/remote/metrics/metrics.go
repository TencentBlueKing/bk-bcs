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
	ErrStatus = "999"
	// SucStatus for success status
	SucStatus = "000"
)

var (
	// alert_manager grpc request action metrics
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_alertmanager",
		Subsystem: "api",
		Name:      "request_total_num",
		Help:      "The total num of requests for alertmanager api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_alertmanager",
		Subsystem: "api",
		Name:      "request_latency_time",
		Help:      "api request latency statistic for alertmanager api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "method", "status"})

	// external http requests action metrics
	requestsTotalAlert = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_alertmanager",
		Subsystem: "alert",
		Name:      "request_total_num",
		Help:      "The total number of requests to alert-system api",
	}, []string{"handler", "path", "method", "status"})

	requestLatencyAlert = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_alertmanager",
		Subsystem: "alert",
		Name:      "request_latency_time",
		Help:      "request latency time for call alert-system api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "path", "method", "status"})

	// handler monitor metrics
	requestsTotalHandlerQueue = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_alertmanager",
		Subsystem: "handler",
		Name:      "queue_total_num",
		Help:      "The total number of handler queueLen",
	}, []string{"handler"})

	requestLatencyHandler = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_alertmanager",
		Subsystem: "handler",
		Name:      "request_latency_time",
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
func ReportAlertAPIMetrics(handler, path, method, status string, started time.Time) {
	requestsTotalAlert.WithLabelValues(handler, path, method, status).Inc()
	requestLatencyAlert.WithLabelValues(handler, path, method, status).Observe(time.Since(started).Seconds())
}

//ReportAlertManagerAPIMetrics report all api action metrics
func ReportAlertManagerAPIMetrics(handler, method, status string, started time.Time) {
	requestTotalAPI.WithLabelValues(handler, method, status).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}

// ReportAlertManagerHandlerQueueLength report handler chanQueue length
func ReportAlertManagerHandlerQueueLength(handler string, queueLen float64) {
	requestsTotalHandlerQueue.WithLabelValues(handler).Set(queueLen)
}

// ReportAlertManagerHandlerFuncLatency report handler func latency
func ReportAlertManagerHandlerFuncLatency(handler, name, status string, started time.Time) {
	requestLatencyHandler.WithLabelValues(handler, name, status).Observe(time.Since(started).Seconds())
}
