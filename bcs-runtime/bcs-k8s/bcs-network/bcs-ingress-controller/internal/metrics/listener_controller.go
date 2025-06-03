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

package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// ListenerMethodEnsureListener ensure listener
	ListenerMethodEnsureListener = "ensureListener"
	// ListenerMethodDeleteListener delete listener
	ListenerMethodDeleteListener = "deleteListener"
)

var (
	WorkerTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "controller",
		Name:      "lbworker",
		Help:      "The total lb worker managed by controller",
	}, []string{"lbid"})

	handleTotalListener = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "listener",
		Name:      "handle_total",
		Help:      "The total number of requests for bkbcs ingress controller to reconcile listener batch",
	}, []string{"batch_size", "is_bulk_mode", "method", "status"})
	handleLatencyListener = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "listener",
		Name:      "handle_latency_seconds",
		Help:      "handle latency statistic for bkbcs ingress controller to reconcile listener batch",
		Buckets:   []float64{10.0, 20.0, 30.0, 60.0, 90.0, 120.0, 150.0, 180.0},
	}, []string{"batch_size", "is_bulk_mode", "method", "status"})
)

func init() {
	metrics.Registry.MustRegister(WorkerTotal)
	metrics.Registry.MustRegister(handleTotalListener)
	metrics.Registry.MustRegister(handleLatencyListener)
}

// ReportHandleListenerMetric report listener handle metric
func ReportHandleListenerMetric(batchSize int, isBulkMode bool, method string, err error, startTime time.Time) {
	batchSizeStr := strconv.Itoa(batchSize)

	isBulk := "true"
	if !isBulkMode {
		isBulk = "false"
	}

	status := "success"
	if err != nil {
		status = "fail"
	}

	handleLatencyListener.WithLabelValues(batchSizeStr, isBulk, method, status).Observe(time.Since(startTime).Seconds())
	handleTotalListener.WithLabelValues(batchSizeStr, isBulk, method, status).Inc()
}
