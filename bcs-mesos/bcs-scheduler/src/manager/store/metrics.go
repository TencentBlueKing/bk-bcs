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

package store

import (
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	ObjectResourceService     = "service"
	ObjectResourceDeployment  = "deployment"
	ObjectResourceApplication = "application"
	ObjectResourceConfigmap   = "configmap"
	ObjectResourceSecret      = "secret"
)

const (
	StoreOperatorCreate = "create"
	StoreOperatorDelete = "delete"
	StoreOperatorUpdate = "update"
	StoreOperatorFetch  = "fetch"
)

// Metrics the store info
var (
	//metric value is object status
	//service、configmap、secret、deployment status only 0 show success
	//application status 0 show Staging、Deploying、Operating、RollingUpdate; 1 show Running; 2 show Finish; 3 show Abnormal,Error
	ObjectResourceInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "object_resource_info",
		Help:      "Object resource info",
	}, []string{"resource", "namespace", "name"})

	//metric value is taskgroup status
	//0 show Staging、Starting; 1 show Running; 2 show Finish、Killing、Killed; 3 show Error、Failed; 4 show Lost
	TaskgroupInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "taskgroup_info",
		Help:      "Taskgroup info",
	}, []string{"namespace", "application", "taskgroup"})

	AgentCpuResourceTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_cpu_resource_total",
		Help:      "Agent cpu resource total",
	}, []string{"InnerIP"})

	AgentMemoryResourceTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_memory_resource_total",
		Help:      "Agent memory resource total",
	}, []string{"InnerIP"})

	AgentCpuResourceRemain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_cpu_resource_remain",
		Help:      "Agent cpu resource remain",
	}, []string{"InnerIP"})

	AgentMemoryResourceRemain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_memory_resource_remain",
		Help:      "Agent memory resource remain",
	}, []string{"InnerIP"})

	StorageOperatorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "storage_operator_total",
		Help:      "Storage operator total",
	}, []string{"operator"})

	StorageOperatorLatencyMs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "storage_operator_latency_ms",
		Help:      "Storage operator latency Millisecond",
	}, []string{"operator"})

	StorageOperatorFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "storage_operator_failed_total",
		Help:      "Storage operator failed total",
	}, []string{"operator"})
)

func init() {
	prometheus.MustRegister(ObjectResourceInfo, TaskgroupInfo, AgentCpuResourceTotal, AgentMemoryResourceTotal,
		StorageOperatorTotal, StorageOperatorLatencyMs, StorageOperatorFailedTotal, AgentCpuResourceRemain, AgentMemoryResourceRemain)
}

func reportObjectResourceInfoMetrics(resource, ns, name, status string) {
	var val float64
	switch status {
	case types.APP_STATUS_STAGING, types.APP_STATUS_DEPLOYING, types.APP_STATUS_OPERATING, types.APP_STATUS_ROLLINGUPDATE:
		val = 0
	case types.APP_STATUS_RUNNING:
		val = 1
	case types.APP_STATUS_FINISH:
		val = 2
	case types.APP_STATUS_ERROR, types.APP_STATUS_ABNORMAL:
		val = 3
	default:
		val = 0
	}

	ObjectResourceInfo.WithLabelValues(resource, ns, name).Set(val)
}

func reportTaskgroupInfoMetrics(ns, name, taskgroupId, status string) {
	var val float64
	switch status {
	case types.TASKGROUP_STATUS_STAGING, types.TASKGROUP_STATUS_STARTING:
		val = 0
	case types.TASKGROUP_STATUS_RUNNING:
		val = 1
	case types.TASKGROUP_STATUS_FINISH, types.TASKGROUP_STATUS_KILLED, types.TASKGROUP_STATUS_KILLING:
		val = 2
	case types.TASKGROUP_STATUS_ERROR, types.TASKGROUP_STATUS_FAIL:
		val = 3
	case types.TASKGROUP_STATUS_LOST:
		val = 4
	default:
		val = 5
	}

	TaskgroupInfo.WithLabelValues(ns, name, taskgroupId).Set(val)
}

func reportAgentInfoMetrics(ip string, totalCpu, remainCpu, totalMem, remainMem float64) {
	AgentCpuResourceTotal.WithLabelValues(ip).Set(totalCpu)
	AgentCpuResourceRemain.WithLabelValues(ip).Set(remainCpu)
	AgentMemoryResourceTotal.WithLabelValues(ip).Set(totalMem)
	AgentMemoryResourceRemain.WithLabelValues(ip).Set(remainMem)
}

func reportStorageOperatorMetrics(operator string, started time.Time, failed bool) {
	StorageOperatorTotal.WithLabelValues(operator).Inc()
	d := time.Duration(time.Since(started).Nanoseconds())
	sec := d / time.Millisecond
	nsec := d % time.Millisecond
	ms := float64(sec) + float64(nsec)/1e6
	StorageOperatorLatencyMs.WithLabelValues(operator).Observe(ms)
	if failed {
		StorageOperatorFailedTotal.WithLabelValues(operator).Inc()
	}
}
