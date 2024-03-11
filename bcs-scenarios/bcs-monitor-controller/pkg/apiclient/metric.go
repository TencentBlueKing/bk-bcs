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

package apiclient

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// StatusErr error shows during lib call
	StatusErr = "err"
	// StatusTimeout timeout during lib call
	StatusTimeout = "timeout"
	// StatusOK lib call successfully
	StatusOK = "ok"
	// StatusExceedLimit lib call exceed limit
	StatusExceedLimit = "exceed_limit"

	// HandlerBKM bkm handler
	HandlerBKM = "bkm"
)

var (
	requestsTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_monitorctrl",
		Subsystem: "api",
		Name:      "request_total",
		Help:      "The total number of requests for bkbcs ingress controller api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_monitorctrl",
		Subsystem: "api",
		Name:      "request_latency_seconds",
		Help:      "api request latency statistic for bkbcs ingress controller api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "method", "status"})
)

func init() {
	metrics.Registry.MustRegister(requestsTotalAPI)
	metrics.Registry.MustRegister(requestLatencyAPI)
}

// ReportAPIRequestMetric report api request metrics
func ReportAPIRequestMetric(handler, method, status string, started time.Time) {
	requestsTotalAPI.WithLabelValues(handler, method, status).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}
