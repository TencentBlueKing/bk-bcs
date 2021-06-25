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

// RecordConfig wrap Api metrics
type RecordConfig struct {
	Version string
	Handler string
	Method  string
	Status  string
	Started time.Time
}

const (
	// BkBcsK8sCustomScheduler for bcs-k8s-customscheduler module metrics prefix
	BkBcsK8sCustomScheduler = "bkbcs_k8scustomscheduler"
)

var (
	// bcs-k8s-custom-scheduler request action metrics
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsK8sCustomScheduler,
		Name:      "api_request_total_num",
		Help:      "The total num of requests for bcs-k8s-custom-scheduler api",
	}, []string{"version", "handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsK8sCustomScheduler,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for bcs-k8s-custom-scheduler api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"version", "handler", "method", "status"})

	// bcs-k8s-custom-scheduler record canScheduler node metrics
	requestsSchedulerNodeNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsK8sCustomScheduler,
		Name:      "filter_node_total_num",
		Help:      "The total number of canSchedulerNode and canNotSchedulerNode",
	}, []string{"version", "scheduler"})
)

func init() {
	// bkbcs_k8s_customscheduler api
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)

	// bkbcs_k8s_customscheduler scheduler node
	prometheus.MustRegister(requestsSchedulerNodeNum)
}

// ReportK8sCustomSchedulerAPIMetrics report all api action metrics
func ReportK8sCustomSchedulerAPIMetrics(version, handler, method, status string, started time.Time) {
	requestTotalAPI.WithLabelValues(version, handler, method, status).Inc()
	requestLatencyAPI.WithLabelValues(version, handler, method, status).Observe(time.Since(started).Seconds())
}
// ReportK8sCustomSchedulerNodeNum report canScheduler/canNotScheduler/totalScheduler node num
func ReportK8sCustomSchedulerNodeNum(version, schedulerHandler string, nodeNum float64) {
	requestsSchedulerNodeNum.WithLabelValues(version, schedulerHandler).Set(nodeNum)
}
