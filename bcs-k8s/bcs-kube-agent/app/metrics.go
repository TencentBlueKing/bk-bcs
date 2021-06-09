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

package app

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	// ErrStatus for call api failure
	ErrStatus = "failure"
	// SucStatus for call api success
	SucStatus = "success"
	// FailConnect for server connect failure
	FailConnect = "999"

	// BkBcsKubeAgent for bcs-kube-agent module metrics prefix
	BkBcsKubeAgent = "bkbcs_kubeagent"
)

var (
	// bcs-kube-agent request action metrics
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsKubeAgent,
		Name:      "clustermanager_request_total_num",
		Help:      "The total num of requests for bcs-cluster-manager api",
	}, []string{"handler", "method", "code"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsKubeAgent,
		Name:      "clustermanager_request_latency_time",
		Help:      "api request latency statistic for bcs-cluster-manager api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "method", "code"})

	// bcs-kube-agent record websocket connection failure num
	requestsClusterManagerWsFailure = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsKubeAgent,
		Name:      "clustermanager_ws_connection_num",
		Help:      "The total number of websocket connection failure",
	}, []string{"handler"})
)

func init() {
	// bcs-kube-agent call bcs-cluster-manager api
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)

	// bcs-kube-agent ws failure num
	prometheus.MustRegister(requestsClusterManagerWsFailure)
}

// reportBcsKubeAgentAPIMetrics report all api action metrics
func reportBcsKubeAgentAPIMetrics(handler, method, code string, started time.Time) {
	requestTotalAPI.WithLabelValues(handler, method, code).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, code).Observe(time.Since(started).Seconds())
}

// reportBcsKubeAgentClusterManagerWsFail report websocket connection num when failure
func reportBcsKubeAgentClusterManagerWsFail(handler string) {
	requestsClusterManagerWsFailure.WithLabelValues(handler).Inc()
}
