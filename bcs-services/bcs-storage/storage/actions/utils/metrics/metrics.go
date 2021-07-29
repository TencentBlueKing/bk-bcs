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

// file metrics.go for common metrics data
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	pushStatusSuccess = "succeed"
	pushStatusFail    = "failed"
)

const (
	// BkBcsStorage for storage module metrics prefix
	BkBcsStorage = "bkbcs_storage"
)

var (
	// http watch action metrics
	watchRequestTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsStorage,
		Name:      "watch_request_total_num",
		Help:      "The total number of requests to bcs-storage watch connection",
	}, []string{"handler", "table"})

	watchHTTPResponseSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsStorage,
		Name:      "watch_response_size_bytes",
		Help:      "The size of http response.",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
	}, []string{"handler", "table"})

	// queue push data metrics
	queuePushTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsStorage,
		Name:      "queue_push_total",
		Help:      "the total number of queue push data",
	}, []string{"name", "status"})
	queuePushLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsStorage,
		Name:      "queue_latency_seconds",
		Help:      "BCS storage queue push operation latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"name", "status"})
)

func init() {
	// watch requests
	prometheus.MustRegister(watchRequestTotal)
	prometheus.MustRegister(watchHTTPResponseSize)
	// queue
	prometheus.MustRegister(queuePushTotal)
	prometheus.MustRegister(queuePushLatency)
}

// ReportWatchRequestInc report watch connection inc
func ReportWatchRequestInc(handler, table string) {
	watchRequestTotal.WithLabelValues(handler, table).Inc()
}

// ReportWatchRequestDec report watch connection des
func ReportWatchRequestDec(handler, table string) {
	watchRequestTotal.WithLabelValues(handler, table).Dec()
}

// ObserveHTTPResponseSize report responseSize metrics when break connection
func ReportWatchHTTPResponseSize(handler, table string, sizeBytes int64) {
	watchHTTPResponseSize.WithLabelValues(handler, table).Observe(float64(sizeBytes))
}

// ReportQueuePushMetrics report all queue push metrics
func ReportQueuePushMetrics(name string, err error, started time.Time) {
	status := pushStatusSuccess
	if err != nil {
		status = pushStatusFail
	}
	queuePushTotal.WithLabelValues(name, status).Inc()
	queuePushLatency.WithLabelValues(name, status).Observe(time.Since(started).Seconds())
}
