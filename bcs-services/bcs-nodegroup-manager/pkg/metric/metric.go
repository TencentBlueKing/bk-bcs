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

// Package metric xxx
package metric

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// StatusErr error shows call error
	StatusErr = "failure"
	// StatusOK call successfully
	StatusOK = "success"
)

const (
	// BkBcsNodegroupManager for prometheus namespace
	BkBcsNodegroupManager = "bcs_nodegroup_manager"
	// BkBcsResourceManager bcs resource manager
	BkBcsResourceManager = "bcs_resource_manager"
)

// InstanceIP xxx
var InstanceIP string

var (
	requestsTotalLib = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "lib_request_total_num",
		Help:      "The total number of requests for nodegroup manager to call other system api",
	}, []string{"system", "handler", "method", "status"})
	requestLatencyLib = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "lib_request_latency_time",
		Help:      "api request latency statistic for nodegroup manager to call other system",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"system", "handler", "method", "status"})

	requestsTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "api_request_total_num",
		Help:      "The total number of requests for nodegroup manager api",
	}, []string{"handler", "method", "status", "instance"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for nodegroup manager api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "method", "status", "instance"})

	clusterClientTotalNum = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "cluster_client_total_num",
		Help:      "The total number for nodegroup manager to call cluster client",
	}, []string{"instance", "cluster_id", "handler", "method", "status"})
	clusterClientLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "cluster_client_latency_time",
		Help:      "api request latency statistic for cluster client",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"instance", "cluster_id", "handler", "method", "status"})

	strategyNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "strategy_num",
		Help:      "the num of different type strategy",
	}, []string{"type", "instance"})

	taskNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "task_num",
		Help:      "the num of task",
	}, []string{"strategy", "instance"})

	terminatedTaskTotalNum = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "terminated_task_num",
		Help:      "The total number of terminated task",
	}, []string{"instance", "strategy", "drain"})

	finishedTaskTotalNum = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "finished_task_num",
		Help:      "The total number of finished task",
	}, []string{"instance", "strategy", "drain"})

	taskHandleLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "task_handle_time",
		Help:      "time for a task handle",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"instance", "strategy", "task", "status", "drain"})

	actionHandleLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "action_handle_time",
		Help:      "time for a action handle",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"instance", "strategy", "action", "status", "cluster_id", "nodegroup"})

	actionTotalNum = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsNodegroupManager,
		Name:      "action_num",
		Help:      "The total number of action",
	}, []string{"instance", "strategy", "cluster_id", "nodegroup", "action"})
)

func init() {
	prometheus.MustRegister(requestsTotalLib)
	prometheus.MustRegister(requestLatencyLib)
	prometheus.MustRegister(requestsTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
	prometheus.MustRegister(clusterClientTotalNum)
	prometheus.MustRegister(clusterClientLatency)
	prometheus.MustRegister(strategyNum)
	prometheus.MustRegister(taskNum)
	prometheus.MustRegister(terminatedTaskTotalNum)
	prometheus.MustRegister(finishedTaskTotalNum)
	prometheus.MustRegister(taskHandleLatency)
	prometheus.MustRegister(actionHandleLatency)
	prometheus.MustRegister(actionTotalNum)
	InstanceIP = os.Getenv("localIp")
}

// ReportLibRequestMetric report lib call metrics
func ReportLibRequestMetric(system, handler, method string, err error, started time.Time) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	requestsTotalLib.WithLabelValues(system, handler, method, status).Inc()
	requestLatencyLib.WithLabelValues(system, handler, method, status).Observe(time.Since(started).Seconds())
}

// ReportClusterClientRequestMetric report cluster client call metrics
func ReportClusterClientRequestMetric(clusterID, handler, method string, err error, started time.Time) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	clusterClientTotalNum.WithLabelValues(InstanceIP, clusterID, handler, method, status).Inc()
	clusterClientLatency.WithLabelValues(InstanceIP, clusterID, handler,
		method, status).Observe(time.Since(started).Seconds())
}

// ReportAPIRequestMetric report api request metrics
func ReportAPIRequestMetric(handler, method string, err error, started time.Time) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	requestsTotalAPI.WithLabelValues(handler, method, status, InstanceIP).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, status, InstanceIP).Observe(time.Since(started).Seconds())
}

// ReportStrategyNumMetric report strategy num metrics
func ReportStrategyNumMetric(strategyType string, num int) {
	strategyNum.WithLabelValues(InstanceIP, strategyType).Set(float64(num))
}

// ReportTaskNumMetric report task num metrics
func ReportTaskNumMetric(strategyName string, num int) {
	taskNum.WithLabelValues(InstanceIP, strategyName).Set(float64(num))
}

// ReportTerminatedTaskNumMetric report terminated task num metrics
func ReportTerminatedTaskNumMetric(strategyName, drain string) {
	terminatedTaskTotalNum.WithLabelValues(InstanceIP, strategyName, drain).Inc()
}

// ReportTaskFinishedMetric report finished task metrics
func ReportTaskFinishedMetric(strategyName, drain string) {
	finishedTaskTotalNum.WithLabelValues(InstanceIP, strategyName, drain).Inc()
}

// ReportTaskHandleLatencyMetric report task handle latency
func ReportTaskHandleLatencyMetric(strategyName, task, status, drain string, started time.Time) {
	taskHandleLatency.WithLabelValues(InstanceIP, strategyName, task, status, drain).Observe(time.Since(started).Minutes())
}

// ReportActionHandleLatencyMetric report action handle latency
func ReportActionHandleLatencyMetric(strategyName, action, status, clusterID, nodegroup string, started time.Time) {
	actionHandleLatency.WithLabelValues(InstanceIP, strategyName, action, status, clusterID, nodegroup).
		Observe(time.Since(started).Seconds())
}

// ReportActionNumMetric report action num metric
func ReportActionNumMetric(strategyName, clusterID, nodegroup, action string) {
	actionTotalNum.WithLabelValues(InstanceIP, strategyName, clusterID, nodegroup, action).Inc()
}
