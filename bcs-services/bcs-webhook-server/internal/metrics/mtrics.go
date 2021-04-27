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
	// StatusFailure for api failure
	StatusFailure = "failure"
	// StatusSuccess for api success
	StatusSuccess = "success"
)

const (
	// BkBcsWebhookServer for module bcs-webhook-server metrics prefix
	BkBcsWebhookServer = "bkbcs_webhookserver"
)

var (
	// bcs-webhook-server request action metrics
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsWebhookServer,
		Name:      "api_request_total_num",
		Help:      "The total num of requests for bcs-webhook-server api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsWebhookServer,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for bcs-webhook-server api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "method", "status"})

	pluginLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsWebhookServer,
		Name:      "plugin_request_latency_time",
		Help:      "plugin request latency statistic for bcs-webhook-server",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"pluginName", "status"})
)

func init() {
	// bcs-webhook-server request api
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)

	prometheus.MustRegister(pluginLatency)
}

//ReportBcsWebhookServerAPIMetrics report all api action metrics
func ReportBcsWebhookServerAPIMetrics(handler, method, status string, started time.Time) {
	requestTotalAPI.WithLabelValues(handler, method, status).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}

// ReportBcsWebhookServerPluginLantency report call plugin lantency
func ReportBcsWebhookServerPluginLantency(pluginName, status string, started time.Time) {
	pluginLatency.WithLabelValues(pluginName, status).Observe(time.Since(started).Seconds())
}