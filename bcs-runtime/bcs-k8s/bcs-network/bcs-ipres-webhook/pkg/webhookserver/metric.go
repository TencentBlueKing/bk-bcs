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

package webhookserver

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// ResultSuccess webhook result is successfully
	ResultSuccess = "success"
	// ResultFail webhook result is failed
	ResultFail = "fail"
)

var (
	// WebhookRequestTotal the total number of requests for webhook server
	WebhookRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs",
		Subsystem: "webhook",
		Name:      "request_total",
		Help:      "the total number of requests for webhook server",
	}, []string{"handler", "result"})

	// WebhookRequestLatency the latency of request for webhook server
	WebhookRequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs",
		Subsystem: "webhook",
		Name:      "request_latency",
		Help:      "the latency of request for webhook server",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "result"})
)

// ReportMetric report metric for webhook
func ReportMetric(handler, result string, started time.Time) {
	WebhookRequestTotal.WithLabelValues(handler, result).Inc()
	WebhookRequestLatency.WithLabelValues(handler, result).Observe(time.Since(started).Seconds())
}
