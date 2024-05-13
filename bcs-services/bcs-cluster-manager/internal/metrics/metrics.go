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

// Package metrics xxx
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// LibCallStatusErr error shows during lib call
	LibCallStatusErr = "failure"
	// LibCallStatusOK lib call successfully
	LibCallStatusOK = "success"
)

const (
	// BkBcsClusterManager xxx
	BkBcsClusterManager = "bkbcs_clustermanager"
)

var (
	requestsTotalLib = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsClusterManager,
		Name:      "lib_request_total_num",
		Help:      "The total number of requests for cluster manager to call other system api",
	}, []string{"system", "handler", "method", "status"})
	requestLatencyLib = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsClusterManager,
		Name:      "lib_request_latency_time",
		Help:      "api request latency statistic for cluster manager to call other system",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"system", "handler", "method", "status"})

	requestsTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsClusterManager,
		Name:      "api_request_total_num",
		Help:      "The total number of requests for cluster manager api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsClusterManager,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for cluster manager api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "method", "status"})

	reportVPCAvailableIPNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsClusterManager,
		Name:      "vpc_remaining_IPNum",
		Help:      "census cloud vpc available ip number",
	}, []string{"cloud", "vpc"})

	reportClusterHealthStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsClusterManager,
		Name:      "cluster_health",
		Help:      "cloud cluster health status",
	}, []string{"cloud", "cluster"})

	reportClusterGroupResourceNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsClusterManager,
		Name:      "cluster_group_nodeNum",
		Help:      "census cluster nodeGroup number",
	}, []string{"cluster", "group", "instance_type"})
	reportClusterGroupMaxResourceNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsClusterManager,
		Name:      "cluster_group_maxNodeNum",
		Help:      "census cluster nodeGroup max number",
	}, []string{"cluster", "group", "instance_type"})

	reportMasterTaskCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsClusterManager,
		Name:      "task_request_total_num",
		Help:      "The total number of task for running task",
	}, []string{"task_type", "status", "child_task"})
	reportMasterTaskLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsClusterManager,
		Name:      "task_request_latency_time",
		Help:      "task latency statistic for running task",
		Buckets: []float64{10.0, 25.0, 50.0, 100.0, 125.0, 150.0, 175.0, 200.0, 250.0, 300.0,
			350.0, 400.0, 500.0, 600.0, 700.0, 800.0},
	}, []string{"task_type", "status", "child_task"})

	reportMachineryTaskNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsClusterManager,
		Name:      "machinery_task",
		Help:      "cluster manager machinery task",
	}, []string{"task_name", "state"})
)

func init() {
	prometheus.MustRegister(requestsTotalLib)
	prometheus.MustRegister(requestLatencyLib)
	prometheus.MustRegister(requestsTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
	prometheus.MustRegister(reportVPCAvailableIPNum)
	prometheus.MustRegister(reportMasterTaskCount)
	prometheus.MustRegister(reportMasterTaskLatency)
	prometheus.MustRegister(reportClusterHealthStatus)
	prometheus.MustRegister(reportClusterGroupResourceNum)
	prometheus.MustRegister(reportClusterGroupMaxResourceNum)
	prometheus.MustRegister(reportMachineryTaskNum)
}

// ReportMasterTaskMetric report lib call metrics
func ReportMasterTaskMetric(taskType, status, childType string, started time.Time) {
	reportMasterTaskCount.WithLabelValues(taskType, status, childType).Inc()
	reportMasterTaskLatency.WithLabelValues(taskType, status, childType).Observe(time.Since(started).Seconds())
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

// ReportClusterGroupAvailableNodeNum report cluster group available nodeNum
func ReportClusterGroupAvailableNodeNum(cluster, group, instanceType string, num float64) {
	reportClusterGroupResourceNum.WithLabelValues(cluster, group, instanceType).Set(num)
}

// ReportClusterGroupMaxNodeNum report cluster group max nodeNum
func ReportClusterGroupMaxNodeNum(cluster, group, instanceType string, num float64) {
	reportClusterGroupMaxResourceNum.WithLabelValues(cluster, group, instanceType).Set(num)
}

// ReportCloudVpcAvailableIPNum report vpc available ipNum
func ReportCloudVpcAvailableIPNum(cloud, vpc string, num float64) {
	reportVPCAvailableIPNum.WithLabelValues(cloud, vpc).Set(num)
}

// ReportCloudClusterHealthStatus report cluster status
func ReportCloudClusterHealthStatus(cloud, cluster string, status float64) {
	reportClusterHealthStatus.WithLabelValues(cloud, cluster).Set(status)
}

// ReportMachineryTaskNum report cluster-manager machinery tasks
func ReportMachineryTaskNum(taskName, state string, num float64) {
	reportMachineryTaskNum.WithLabelValues(taskName, state).Set(num)
}
