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

// Package metrics DOTO
package metrics

import (
	"time"

	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	caNamespace = "cluster_autoscaler"
)

var (
	webhookParams = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "webhook_parameters",
			Help:      "Parameters of webhook mode of CA",
		},
		[]string{"mode", "config"},
	)

	webhookExecDuration = k8smetrics.NewGauge(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "webhook_exec_duration",
			Help:      "Exec duration of webhook mode of CA",
		},
	)

	webhookScaleUpResponse = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "webhook_scale_up_response",
			Help:      "Scale up response of webhook mode of CA",
		},
		[]string{"node_group"},
	)

	webhookScaleDownIPResponse = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "webhook_scale_down_ip_response",
			Help:      "Scale down response (type of NodeIPs) of webhook mode of CA",
		},
		[]string{"node_group", "node_IPs"},
	)

	webhookScaleDownNumResponse = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "webhook_scale_down_num_response",
			Help:      "Scale down response (type of NodeNum) of webhook mode of CA",
		},
		[]string{"node_group"},
	)

	scaleTask = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "scale_task",
			Help:      "Scale task status of CA",
		},
		[]string{"task_id", "node_group", "scale_type", "status"},
	)

	failedScaleDownCount = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_downs_total",
			Help:      "Number of times scale-down operation has failed.",
		}, []string{"node", "reason"},
	)

	resourceUsedRatio = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "resource_used_ratio",
			Help:      "Ratio of resource has been used",
		}, []string{"resource"})

	// ========== NodeGroupSet 相关指标 ==========

	// nodeGroupSetCPUUsageRatio 每个 Set 的 CPU 使用率
	nodeGroupSetCPUUsageRatio = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nodegroup_set_cpu_usage_ratio",
			Help:      "CPU usage ratio of each NodeGroupSet (0.0-1.0)",
		},
		[]string{"set_name"},
	)

	// nodeGroupSetMemoryUsageRatio 每个 Set 的内存使用率
	nodeGroupSetMemoryUsageRatio = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nodegroup_set_memory_usage_ratio",
			Help:      "Memory usage ratio of each NodeGroupSet (0.0-1.0)",
		},
		[]string{"set_name"},
	)

	// nodeGroupSetScaleUpTriggered 每个 Set 触发扩容的次数
	nodeGroupSetScaleUpTriggered = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "nodegroup_set_scale_up_triggered_total",
			Help:      "Total number of times scale-up was triggered for each NodeGroupSet",
		},
		[]string{"set_name"},
	)

	// nodeGroupSetScaleDownPrevented 每个 Set 被保护而阻止缩容的次数
	nodeGroupSetScaleDownPrevented = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace: caNamespace,
			Name:      "nodegroup_set_scale_down_prevented_total",
			Help:      "Total number of times scale-down was prevented due to buffer ratio for each NodeGroupSet",
		},
		[]string{"set_name"},
	)

	// nodeGroupSetTotalCPU 每个 Set 的总 CPU(毫核)
	nodeGroupSetTotalCPU = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nodegroup_set_total_cpu_millicores",
			Help:      "Total CPU millicores in each NodeGroupSet",
		},
		[]string{"set_name"},
	)

	// nodeGroupSetTotalMemory 每个 Set 的总内存(字节)
	nodeGroupSetTotalMemory = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace: caNamespace,
			Name:      "nodegroup_set_total_memory_bytes",
			Help:      "Total memory bytes in each NodeGroupSet",
		},
		[]string{"set_name"},
	)
)

// RegisterLocal registers local metrics
func RegisterLocal() {
	legacyregistry.MustRegister(webhookParams)
	legacyregistry.MustRegister(webhookExecDuration)
	legacyregistry.MustRegister(webhookScaleUpResponse)
	legacyregistry.MustRegister(webhookScaleDownIPResponse)
	legacyregistry.MustRegister(webhookScaleDownNumResponse)
	legacyregistry.MustRegister(failedScaleDownCount)
	legacyregistry.MustRegister(resourceUsedRatio)

	// 注册 NodeGroupSet 相关指标
	legacyregistry.MustRegister(nodeGroupSetCPUUsageRatio)
	legacyregistry.MustRegister(nodeGroupSetMemoryUsageRatio)
	legacyregistry.MustRegister(nodeGroupSetScaleUpTriggered)
	legacyregistry.MustRegister(nodeGroupSetScaleDownPrevented)
	legacyregistry.MustRegister(nodeGroupSetTotalCPU)
	legacyregistry.MustRegister(nodeGroupSetTotalMemory)
}

// RegisterWebhookParams collects parameters fo webhook mode
func RegisterWebhookParams(mode, config string) {
	webhookParams.WithLabelValues(mode, config).Set(1)
}

// UpdateWebhookExecDuration updates execute duration of webhook mode
func UpdateWebhookExecDuration(start time.Time) {
	duration := time.Since(start).Seconds()
	webhookExecDuration.Set(duration)
}

// UpdateWebhookScaleUpResponse updates scale up response of webhook mode
func UpdateWebhookScaleUpResponse(nodeGroup string, desired int) {
	webhookScaleUpResponse.WithLabelValues(nodeGroup).Set(float64(desired))
}

// UpdateWebhookScaleDownIPResponse updates scale down response(type of NodeIPs) of webhook mode
func UpdateWebhookScaleDownIPResponse(nodeGroup, ips string) {
	webhookScaleDownIPResponse.WithLabelValues(nodeGroup, ips).Set(1)
}

// UpdateWebhookScaleDownNumResponse updates scale down response(type of NodeNum) of webhook mode
func UpdateWebhookScaleDownNumResponse(nodeGroup string, num int) {
	webhookScaleDownNumResponse.WithLabelValues(nodeGroup).Set(float64(num))
}

// RegisterScaleTask registers scale task metrics
func RegisterScaleTask() {
	legacyregistry.MustRegister(scaleTask)
}

// UpdateScaleTask updates scale task status of CA
func UpdateScaleTask(id, nodeGroup, scaleType, status string) {
	scaleTask.WithLabelValues(id, nodeGroup, scaleType, status).Set(1)
}

// RegisterFailedScaleDown records a failed scale-down operation
func RegisterFailedScaleDown(node, reason string) {
	failedScaleDownCount.WithLabelValues(node, reason).Inc()
}

// UpdateResourceUsedRatio updates the ratio of resource used
func UpdateResourceUsedRatio(name string, ratio float64) {
	resourceUsedRatio.WithLabelValues(name).Set(ratio)
}

// ========== NodeGroupSet Metrics 记录函数 ==========

// RecordNodeGroupSetResourceUsage 记录 Set 的资源使用情况
func RecordNodeGroupSetResourceUsage(setName string, cpuRatio, memRatio float64, totalCPU, totalMemory int64) {
	nodeGroupSetCPUUsageRatio.WithLabelValues(setName).Set(cpuRatio)
	nodeGroupSetMemoryUsageRatio.WithLabelValues(setName).Set(memRatio)
	nodeGroupSetTotalCPU.WithLabelValues(setName).Set(float64(totalCPU))
	nodeGroupSetTotalMemory.WithLabelValues(setName).Set(float64(totalMemory))
}

// RecordNodeGroupSetScaleUpTriggered 记录 Set 触发扩容
func RecordNodeGroupSetScaleUpTriggered(setName string) {
	nodeGroupSetScaleUpTriggered.WithLabelValues(setName).Inc()
}

// RecordNodeGroupSetScaleDownPrevented 记录 Set 阻止缩容
// NOCC:tosa/fn_length(设计如此)
func RecordNodeGroupSetScaleDownPrevented(setName string) {
	nodeGroupSetScaleDownPrevented.WithLabelValues(setName).Inc()
}
