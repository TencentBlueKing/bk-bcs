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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	hostnetSegmentAllocated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "hostnetportpool",
		Name:      "segment_allocated",
		Help:      "number of allocated segments per pool per node",
	}, []string{"pool_name", "pool_namespace", "node_name"})

	hostnetSegmentTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "hostnetportpool",
		Name:      "segment_total",
		Help:      "total number of segments per pool per node",
	}, []string{"pool_name", "pool_namespace", "node_name"})

	hostnetAllocateFailedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "hostnetportpool",
		Name:      "allocate_failed",
		Help:      "gauge marking pods with failed allocation (1=failed, no data/0=ok)",
	}, []string{"pod_name", "pod_namespace"})

	hostnetAllocateFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "hostnetportpool",
		Name:      "allocate_failed_total",
		Help:      "total number of allocation failures per pool per node",
	}, []string{"pool_name", "node_name"})

	hostnetCacheRebuildPodsRecovered = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "hostnetportpool",
		Name:      "cache_rebuild_pods_recovered",
		Help:      "number of pods recovered during cache rebuild",
	}, []string{})

	hostnetPoolShrinkConflictTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "hostnetportpool",
		Name:      "pool_shrink_conflict_total",
		Help:      "total number of pool shrink conflicts per pool per node",
	}, []string{"pool_name", "pool_namespace", "node_name"})

	hostnetSegmentLeakReleasedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "hostnetportpool",
		Name:      "segment_leak_released_total",
		Help:      "total number of leaked segments released by periodic checker",
	}, []string{"pool_key", "node_name", "reason"})
)

func init() {
	metrics.Registry.MustRegister(hostnetSegmentAllocated)
	metrics.Registry.MustRegister(hostnetSegmentTotal)
	metrics.Registry.MustRegister(hostnetAllocateFailedGauge)
	metrics.Registry.MustRegister(hostnetAllocateFailedTotal)
	metrics.Registry.MustRegister(hostnetCacheRebuildPodsRecovered)
	metrics.Registry.MustRegister(hostnetPoolShrinkConflictTotal)
	metrics.Registry.MustRegister(hostnetSegmentLeakReleasedTotal)
}

// ReportHostNetSegmentMetrics reports allocated/total segment gauge for a pool+node.
func ReportHostNetSegmentMetrics(poolName, poolNamespace, nodeName string, allocated, total int) {
	if hostnetSegmentAllocated != nil {
		hostnetSegmentAllocated.WithLabelValues(poolName, poolNamespace, nodeName).Set(float64(allocated))
	}
	if hostnetSegmentTotal != nil {
		hostnetSegmentTotal.WithLabelValues(poolName, poolNamespace, nodeName).Set(float64(total))
	}
}

// ReportHostNetAllocateFailed sets the per-pod allocation failure gauge.
func ReportHostNetAllocateFailed(podName, podNamespace string, failed bool) {
	if hostnetAllocateFailedGauge == nil {
		return
	}
	if failed {
		hostnetAllocateFailedGauge.WithLabelValues(podName, podNamespace).Set(1)
	} else {
		hostnetAllocateFailedGauge.WithLabelValues(podName, podNamespace).Set(0)
	}
}

// CleanHostNetAllocateFailedMetric removes the per-pod allocation failure gauge entry.
func CleanHostNetAllocateFailedMetric(podName, podNamespace string) {
	if hostnetAllocateFailedGauge == nil {
		return
	}
	hostnetAllocateFailedGauge.Delete(prometheus.Labels{"pod_name": podName, "pod_namespace": podNamespace})
}

// IncreaseHostNetAllocateFailedTotal increments the per-pool-per-node failure counter.
func IncreaseHostNetAllocateFailedTotal(poolName, nodeName string) {
	if hostnetAllocateFailedTotal == nil {
		return
	}
	hostnetAllocateFailedTotal.WithLabelValues(poolName, nodeName).Inc()
}

// CleanHostNetSegmentMetrics removes the segment_allocated and segment_total
// gauge entries for a specific pool+node combination. Called when a node is
// removed so stale label sets don't linger in Prometheus.
func CleanHostNetSegmentMetrics(poolName, poolNamespace, nodeName string) {
	labels := prometheus.Labels{
		"pool_name":      poolName,
		"pool_namespace": poolNamespace,
		"node_name":      nodeName,
	}
	if hostnetSegmentAllocated != nil {
		hostnetSegmentAllocated.Delete(labels)
	}
	if hostnetSegmentTotal != nil {
		hostnetSegmentTotal.Delete(labels)
	}
}

// IncreaseHostNetPoolShrinkConflict increments the pool shrink conflict counter.
func IncreaseHostNetPoolShrinkConflict(poolName, poolNamespace, nodeName string) {
	if hostnetPoolShrinkConflictTotal == nil {
		return
	}
	hostnetPoolShrinkConflictTotal.WithLabelValues(poolName, poolNamespace, nodeName).Inc()
}

// IncreaseHostNetSegmentLeakReleased increments the counter when the periodic checker
// releases a leaked segment. reason should be "pod_not_found" or "pod_terminated".
func IncreaseHostNetSegmentLeakReleased(poolKey, nodeName, reason string) {
	if hostnetSegmentLeakReleasedTotal == nil {
		return
	}
	hostnetSegmentLeakReleasedTotal.WithLabelValues(poolKey, nodeName, reason).Inc()
}

// ReportHostNetCacheRebuildRecovered sets the number of pods recovered during cache rebuild.
func ReportHostNetCacheRebuildRecovered(count int) {
	if hostnetCacheRebuildPodsRecovered == nil {
		return
	}
	hostnetCacheRebuildPodsRecovered.WithLabelValues().Set(float64(count))
}
