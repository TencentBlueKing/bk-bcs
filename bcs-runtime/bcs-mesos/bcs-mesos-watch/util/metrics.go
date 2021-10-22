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

package util

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// StatusFailure for api failure
	StatusFailure = "failure"
	// StatusSuccess for api success
	StatusSuccess = "success"
)

const (
	// BkBcsMesosWatch for bcs-mesos-watch module metrics prefix
	BkBcsMesosWatch = "bkbcs_mesoswatch"
)

var (
	// storageTotal for request storage metrics counter
	storageTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsMesosWatch,
		Name:      "storage_request_total_num",
		Help:      "The total number of storage synchronization operation.",
	}, []string{"cluster_id", "datatype", "action", "handler", "status"})

	// storageLatency for request storage metrics latency
	storageLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsMesosWatch,
		Name:      "storage_request_latency_time",
		Help:      "BCS mesos datawatch storage operation latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"cluster_id", "datatype", "action", "handler", "status"})

	// requestsHandlerQueueLen for handler queue length metrics
	requestsHandlerQueueLen = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsMesosWatch,
		Name:      "queue_total_num",
		Help:      "The total number of handler queueLen",
	}, []string{"cluster_id", "handler"})

	// syncTotal report sync resource counter
	syncTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsMesosWatch,
		Name:      "sync_total",
		Help:      "The total number of data sync event.",
	}, []string{"cluster_id", "datatype", "action", "status"})

	// handlerDiscardEvents is a Counter that tracks the number of discarding events for the handler event.
	handlerDiscardEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: BkBcsMesosWatch,
			Name:      "queue_handler_discard_events",
			Help:      "The number of discard events in handler.",
		}, []string{"cluster_id", "handler"})
)

// ReportStorageMetrics report bcs-storage metrics
func ReportStorageMetrics(clusterID, datatype, action, handler, status string, started time.Time) {
	storageTotal.WithLabelValues(clusterID, datatype, action, handler, status).Inc()
	storageLatency.WithLabelValues(clusterID, datatype, action, handler, status).Observe(time.Since(started).Seconds())
}

// ReportHandlerQueueLength report handler queue len
func ReportHandlerQueueLength(clusterID, handler string, len float64) {
	requestsHandlerQueueLen.WithLabelValues(clusterID, handler).Set(len)
}

// ReportHandlerQueueLengthInc inc queue len
func ReportHandlerQueueLengthInc(clusterID, handler string) {
	requestsHandlerQueueLen.WithLabelValues(clusterID, handler).Inc()
}

// ReportHandlerQueueLengthDec dec queue len
func ReportHandlerQueueLengthDec(clusterID, handler string) {
	requestsHandlerQueueLen.WithLabelValues(clusterID, handler).Dec()
}

// ReportSyncTotal report sync event counter
func ReportSyncTotal(clusterID, datatype, action, status string) {
	syncTotal.WithLabelValues(clusterID, datatype, action, status).Inc()
}

// ReportHandlerDiscardEvents report handler discard events num
func ReportHandlerDiscardEvents(clusterID, handler string) {
	handlerDiscardEvents.WithLabelValues(clusterID, handler).Inc()
}

func init() {
	// bkbcs_datawatch storage metrics
	prometheus.MustRegister(storageTotal)
	prometheus.MustRegister(storageLatency)

	// bkbcs_datawatch handler queue length metrics
	prometheus.MustRegister(requestsHandlerQueueLen)

	// bkbcs_datawatch sync event total counter
	prometheus.MustRegister(syncTotal)

	// bkbcs_datawatch handler discard event
	prometheus.MustRegister(handlerDiscardEvents)
}
