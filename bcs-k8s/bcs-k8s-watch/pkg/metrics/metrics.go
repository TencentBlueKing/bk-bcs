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
	// BkBcsK8sWatch for bcs-k8s-watch module metrics prefix
	BkBcsK8sWatch = "bkbcs_k8swatch"
)

var (
	// bcs-k8s-watch request action metrics
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsK8sWatch,
		Name:      "storage_request_total_num",
		Help:      "The total num of requests for bcs-storage api",
	}, []string{"cluster_id", "handler", "namespace", "resource_type", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsK8sWatch,
		Name:      "storage_request_latency_time",
		Help:      "api request latency statistic for bcs-storage api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"cluster_id", "handler", "namespace", "resource_type", "method", "status"})

	// bcs-k8s-watch record queueData metrics
	requestsTotalHandlerQueue = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsK8sWatch,
		Name:      "queue_handler_total_num",
		Help:      "The total number of handler queueLen",
	}, []string{"cluster_id", "handler"})

	// handlerDiscardEvents is a Counter that tracks the number of discarding events for the handler event.
	handlerDiscardEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: BkBcsK8sWatch,
			Name:      "queue_handler_discard_events",
			Help:      "The number of discard events in handler.",
		}, []string{"cluster_id", "handler"},
	)

	requestLatencyHandler = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsK8sWatch,
		Name:      "queue_latency_time",
		Help:      "request latency time for queue parse data",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"cluster_id", "handler", "name", "status"})
)

func init() {
	// bcs-k8s-watch api
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)

	// handler queue data
	prometheus.MustRegister(requestsTotalHandlerQueue)
	prometheus.MustRegister(requestLatencyHandler)

	// handler discard events
	prometheus.MustRegister(handlerDiscardEvents)
}

//ReportK8sWatchAPIMetrics report all api action metrics
func ReportK8sWatchAPIMetrics(clusterID, handler, namespace, resourceType, method, status string, started time.Time) {
	requestTotalAPI.WithLabelValues(clusterID, handler, namespace, resourceType, method, status).Inc()
	requestLatencyAPI.WithLabelValues(clusterID, handler, namespace, resourceType, method, status).Observe(time.Since(started).Seconds())
}

// ReportK8sWatchHandlerQueueLength report handler chanQueue length
func ReportK8sWatchHandlerQueueLength(clusterID, handler string, queueLen float64) {
	requestsTotalHandlerQueue.WithLabelValues(clusterID, handler).Set(queueLen)
}

// ReportK8sWatchHandlerQueueLengthInc inc queue len
func ReportK8sWatchHandlerQueueLengthInc(clusterID, handler string) {
	requestsTotalHandlerQueue.WithLabelValues(clusterID, handler).Inc()
}

// ReportK8sWatchHandlerQueueLengthDec dec queue len
func ReportK8sWatchHandlerQueueLengthDec(clusterID, handler string) {
	requestsTotalHandlerQueue.WithLabelValues(clusterID, handler).Dec()
}

// ReportK8sWatchHandlerDiscardEvents report handler discard events num
func ReportK8sWatchHandlerDiscardEvents(clusterID, handler string) {
	handlerDiscardEvents.WithLabelValues(clusterID, handler).Inc()
}

// ReportK8sWatchHandlerFuncLatency report handler func latency
func ReportK8sWatchHandlerFuncLatency(clusterID, handler, name, status string, started time.Time) {
	requestLatencyHandler.WithLabelValues(clusterID, handler, name, status).Observe(time.Since(started).Seconds())
}
