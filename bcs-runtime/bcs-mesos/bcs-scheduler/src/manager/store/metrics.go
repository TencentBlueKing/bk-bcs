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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ObjectResourceService     = "service"
	ObjectResourceDeployment  = "deployment"
	ObjectResourceApplication = "application"
	ObjectResourceConfigmap   = "configmap"
	ObjectResourceSecret      = "secret"

	ResourceStatusRunning   = "running"
	ResourceStatusFailed    = "failed"
	ResourceStatusFinish    = "finish"
	ResourceStatusOperating = "operating"
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
	}, []string{"resource", "namespace", "name", "status"})

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
	}, []string{"InnerIP", "clusterId"})

	AgentMemoryResourceTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_memory_resource_total",
		Help:      "Agent memory resource total",
	}, []string{"InnerIP", "clusterId"})

	AgentCpuResourceRemain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_cpu_resource_remain",
		Help:      "Agent cpu resource remain",
	}, []string{"InnerIP", "clusterId"})

	AgentMemoryResourceRemain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_memory_resource_remain",
		Help:      "Agent memory resource remain",
	}, []string{"InnerIP", "clusterId"})

	AgentIpResourceRemain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "agent_ip_resource_remain",
		Help:      "Agent ip resource remain",
	}, []string{"InnerIP", "clusterId"})

	ClusterCpuResourceRemain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "cluster_cpu_resource_remain",
		Help:      "Cluster cpu resource remain",
	}, []string{"clusterId"})

	ClusterCpuResourceAvailable = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "cluster_cpu_resource_available",
		Help:      "Cluster cpu resource available",
	}, []string{"cloudId"})

	ClusterMemoryResourceRemain = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "cluster_memory_resource_remain",
		Help:      "Cluster memory resource remain",
	}, []string{"clusterId"})

	ClusterMemoryResourceAvailable = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "cluster_memory_resource_available",
		Help:      "Cluster memory resource available",
	}, []string{"clusterId"})

	ClusterCpuResourceTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "cluster_cpu_resource_total",
		Help:      "Cluster cpu resource total",
	}, []string{"clusterId"})

	ClusterMemoryResourceTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "cluster_memory_resource_total",
		Help:      "Cluster memory resource total",
	}, []string{"clusterId"})

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
		StorageOperatorTotal, StorageOperatorLatencyMs, StorageOperatorFailedTotal, AgentCpuResourceRemain,
		AgentMemoryResourceRemain, AgentIpResourceRemain, ClusterCpuResourceRemain, ClusterMemoryResourceRemain,
		ClusterCpuResourceTotal, ClusterMemoryResourceTotal, ClusterCpuResourceAvailable,
		ClusterMemoryResourceAvailable)
}

func ReportObjectResourceInfoMetrics(resource, ns, name, status string) {
	var str string
	switch status {
	case types.APP_STATUS_STAGING, types.APP_STATUS_DEPLOYING, types.APP_STATUS_OPERATING, types.APP_STATUS_ROLLINGUPDATE:
		str = ResourceStatusOperating
	case types.APP_STATUS_RUNNING:
		str = ResourceStatusRunning
	case types.APP_STATUS_FINISH:
		str = ResourceStatusFinish
	case types.APP_STATUS_ERROR, types.APP_STATUS_ABNORMAL:
		str = ResourceStatusFailed
	default:
		str = ResourceStatusRunning
	}

	ObjectResourceInfo.WithLabelValues(resource, ns, name, str).Set(1)
}

func ReportTaskgroupInfoMetrics(ns, name, taskgroupId, status string) {
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

func ReportAgentInfoMetrics(ip, clusterId string, totalCpu, remainCpu, totalMem, remainMem, remainIp float64) {
	AgentCpuResourceTotal.WithLabelValues(ip, clusterId).Set(totalCpu)
	AgentCpuResourceRemain.WithLabelValues(ip, clusterId).Set(remainCpu)
	AgentMemoryResourceTotal.WithLabelValues(ip, clusterId).Set(totalMem)
	AgentMemoryResourceRemain.WithLabelValues(ip, clusterId).Set(remainMem)
	AgentIpResourceRemain.WithLabelValues(ip, clusterId).Set(remainIp)
}

func ReportClusterInfoMetrics(clusterId string, remainCpu, availableCpu, totalCpu, remainMem,
	availableMem, totalMem float64) {
	ClusterCpuResourceRemain.WithLabelValues(clusterId).Set(remainCpu)
	ClusterMemoryResourceRemain.WithLabelValues(clusterId).Set(remainMem)
	ClusterCpuResourceTotal.WithLabelValues(clusterId).Set(totalCpu)
	ClusterMemoryResourceTotal.WithLabelValues(clusterId).Set(totalMem)
	ClusterCpuResourceAvailable.WithLabelValues(clusterId).Set(availableCpu)
	ClusterMemoryResourceAvailable.WithLabelValues(clusterId).Set(availableMem)
}

func ReportStorageOperatorMetrics(operator string, started time.Time, failed bool) {
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
