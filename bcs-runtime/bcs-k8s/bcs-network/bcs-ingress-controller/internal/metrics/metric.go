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

// Package metrics include sdk and crd metrics
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// LibCallStatusErr error shows during lib call
	LibCallStatusErr = "err"
	// LibCallStatusTimeout timeout during lib call
	LibCallStatusTimeout = "timeout"
	// LibCallStatusOK lib call successfully
	LibCallStatusOK = "ok"
	// LibCallStatusExceedLimit lib call exceed limit
	LibCallStatusExceedLimit = "exceed_limit"
	// LibCallStatusLBLock operate lb is locked
	LibCallStatusLBLock = "lb_lock"

	// EventTypeAdd type for add event
	EventTypeAdd = "add"
	// EventTypeUpdate type for update event
	EventTypeUpdate = "update"
	// EventTypeDelete type for delete event
	EventTypeDelete = "delete"
	// EventTypeUnknown unknown event type
	EventTypeUnknown = "other"

	// ObjectPortbinding object for portbinding
	ObjectPortbinding = "portbinding"
	// ObjectIngress object for Ingress
	ObjectIngress = "ingress"
	// ObjectPortPool object for port pool
	ObjectPortPool = "portpool"

	FailTypeConfigError    = "config_error"
	FailTypeDeleteFailed   = "delete_failed"
	FailTypeReconcileError = "reconcile_error"
)

var (
	ControllerInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "controller",
		Name:      "info",
		Help:      "Controller info",
	}, []string{"version", "cloud"})
	requestsTotalLib = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "lib",
		Name:      "request_total",
		Help:      "The total number of requests for bkbcs ingress controller to call other system api",
	}, []string{"system", "handler", "method", "status"})
	requestLatencyLib = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "lib",
		Name:      "request_latency_seconds",
		Help:      "api request latency statistic for bkbcs ingress controller to call other system",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"system", "handler", "method", "status"})

	requestsTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "api",
		Name:      "request_total",
		Help:      "The total number of requests for bkbcs ingress controller api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "api",
		Name:      "request_latency_seconds",
		Help:      "api request latency statistic for bkbcs ingress controller api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "method", "status"})

	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "event",
		Name:      "counter",
		Help:      "The total event counter for different object",
	}, []string{"object", "type"})

	failCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "failed",
		Name:      "counter",
		Help:      "The total failed counter for different object",
	}, []string{"object", "type", "namespace", "name"})
)

func init() {
	metrics.Registry.MustRegister(requestsTotalLib)
	metrics.Registry.MustRegister(requestLatencyLib)
	metrics.Registry.MustRegister(requestsTotalAPI)
	metrics.Registry.MustRegister(requestLatencyAPI)
	metrics.Registry.MustRegister(eventCounter)
	metrics.Registry.MustRegister(failCounter)
	metrics.Registry.MustRegister(ControllerInfo)
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

// IncreaseEventCounter increase event counter
func IncreaseEventCounter(object, eventType string) {
	eventCounter.WithLabelValues(object, eventType).Inc()
}

// IncreaseFailMetric increase fail counter
func IncreaseFailMetric(object string, failedType string, namespace, name string) {
	failCounter.WithLabelValues(object, failedType, namespace, name).Inc()
}
