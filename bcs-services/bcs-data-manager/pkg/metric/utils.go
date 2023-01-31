/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metric

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
)

const (
	// WorkloadCPUUsage TODO
	WorkloadCPUUsage = "sum(rate(container_cpu_usage_seconds_total{%s,image!=\"\"}[%s]))"
	// ClusterCPUUsage TODO
	ClusterCPUUsage = "sum(irate(node_cpu_seconds_total{cluster_id=\"%s\", job=\"%s\", " +
		"mode!=\"idle\"}[%s]))"
	// NamespaceCPUUsage TODO
	NamespaceCPUUsage = "sum(rate(container_cpu_usage_seconds_total{%s,image!=\"\"}[%s]))by(%s)"
	// K8sCPURequest TODO
	K8sCPURequest       = "sum(kube_pod_container_resource_requests_cpu_cores{%s})"
	K8sCPULimits        = "sum(kube_pod_container_resource_limits_cpu_cores{%s})"
	WorkloadMemoryUsed  = "sum(container_memory_working_set_bytes{%s,image!=\"\"})"
	NamespaceMemoryUsed = "sum(container_memory_working_set_bytes{%s,image!=\"\"})by(%s)"
	ClusterMemoryUsed   = "sum(node_memory_MemTotal_bytes{cluster_id=\"%s\",job=\"%s\"})" +
		"-sum(node_memory_MemFree_bytes{cluster_id=\"%s\",job=\"%s\"})" +
		"-sum(node_memory_Buffers_bytes{cluster_id=\"%s\",job=\"%s\"})" +
		"-sum(node_memory_Cached_bytes{cluster_id=\"%s\",job=\"%s\"})" +
		"+sum(node_memory_Shmem_bytes{cluster_id=\"%s\",job=\"%s\"})"
	K8sMemoryRequest         = "sum(kube_pod_container_resource_requests_memory_bytes{%s})"
	K8sMemoryLimit           = "sum(kube_pod_container_resource_limits_memory_bytes{%s})"
	WorkloadInstance         = "count(sum(container_memory_rss{%s}) by (%s))"
	NamespaceResourceQuota   = "kube_resourcequota{%s, type=\"hard\", %s}"
	PromMasterIP             = "sum(kube_node_role{cluster_id=\"%s\"})by(node)"
	PromNodeIP               = "sum(kube_node_created{cluster_id=\"%s\",%s})by(node)"
	NodeCount                = "count(sum(kube_node_created{cluster_id=\"%s\",%s})by(node))"
	ClusterTotalCPU          = "sum(kube_node_status_capacity_cpu_cores{cluster_id=\"%s\"})by(cluster_id)"
	ClusterTotalMemory       = "sum(kube_node_status_capacity_memory_bytes{cluster_id=\"%s\"})by(cluster_id)"
	NodeUsageQuantile        = "quantile(%s,%s)"
	MesosWorkloadMemoryLimit = "sum(sum(container_spec_memory_limit_bytes{%s})by(%s))"
	// MesosWorkloadCPULimit TODO
	MesosWorkloadCPULimit = "sum(sum(container_spec_cpu_quota{%s})by(%s)/100000)"
	// MesosMemoryLimit TODO
	MesosMemoryLimit = "sum(container_spec_memory_limit_bytes{%s})by(%s)"
	// MesosCPULimit TODO
	MesosCPULimit = "sum(container_spec_cpu_quota{%s})by(%s)/100000"
	// NodeCPUUsage TODO
	NodeCPUUsage = "sum(irate(node_cpu_seconds_total{cluster_id=\"%s\", job=\"node-exporter\", " +
		"mode!=\"idle\",%s}[1m]))by(instance)/" +
		"sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id=\"%s\", job=\"node-exporter\", " +
		"mode=\"idle\",%s}))by(instance)"
	// ClusterAutoscalerUpCount TODO
	ClusterAutoscalerUpCount = "sum(kube_event_unique_events_total{cluster_id=\"%s\", " +
		"source=\"/cluster-autoscaler\",reason=\"ScaledUpGroup\",involved_object_namespace=\"bcs-system\"})"
	// ClusterAutoscalerDownCount TODO
	ClusterAutoscalerDownCount = "sum(kube_event_unique_events_total{cluster_id=\"%s\", " +
		"source=\"/cluster-autoscaler\",reason=\"ScaleDown\", involved_object_namespace=\"bcs-system\"})"
	// GeneralPodAutoscalerCount TODO
	GeneralPodAutoscalerCount = "kube_event_unique_events_total{cluster_id=\"%s\", " +
		"involved_object_kind=\"GeneralPodAutoscaler\",involved_object_name=\"%s\",involved_object_namespace=\"%s\"," +
		"source=\"/pod-autoscaler\",reason=\"SuccessfulRescale\"}"
	// HorizontalPodAutoscalerCount TODO
	HorizontalPodAutoscalerCount = "kube_event_unique_events_total{cluster_id=\"%s\", " +
		"involved_object_kind=\"HorizontalPodAutoscaler\",involved_object_name=\"%s\",involved_object_namespace=\"%s\"," +
		"source=\"/horizontal-pod-autoscaler\",reason=\"SuccessfulRescale\"}"
	// MinOverTime TODO
	MinOverTime = "min_over_time(%s[%s])"
	// MaxOverTime TODO
	MaxOverTime = "max_over_time(%s[%s])"

	// BKMonitorNodeExporterLabelJob bkMonitor node-exporter label name
	BKMonitorNodeExporterLabelJob = "bkmonitor-operator-stack-prometheus-node-exporter"

	PrometheusNodeExporterLabelJob = "node-exporter"
)

const (
	// DeploymentPodCondition TODO
	DeploymentPodCondition = "cluster_id=\"%s\", namespace=\"%s\",pod=~\"%s-[0-9a-z]*-[0-9a-z]*$\"," +
		"container_name!=\"POD\""
	// OtherPodCondition TODO
	OtherPodCondition = "cluster_id=\"%s\", namespace=\"%s\",pod=~\"%s-[0-9a-z]*$\"," +
		"container_name!=\"POD\""
	// MesosPodCondition TODO
	MesosPodCondition = "cluster_id=\"%s\", namespace=\"%s\",name=~\".*.%s.%s.%s.*\"," +
		"container_name!=\"POD\""
	// PodSumCondition TODO
	PodSumCondition = "pod"
	// MesosPodSumCondition TODO
	MesosPodSumCondition = "name"
	// NamespaceCondition TODO
	NamespaceCondition = "cluster_id=\"%s\", namespace=\"%s\""
	// NamespaceSumCondition TODO
	NamespaceSumCondition = "namespace"
	// ClusterCondition TODO
	ClusterCondition = "cluster_id=\"%s\""
	// ClusterSumCondition TODO
	ClusterSumCondition = "cluster_id"
)

var (
	queryMetricFromBKMonitor = false
	parseEnv                 = false
)

func IfQueryFromBKMonitor() bool {
	if parseEnv {
		return queryMetricFromBKMonitor
	}
	env := os.Getenv("queryMetricFromBKMonitor")
	result, err := strconv.ParseBool(env)
	if err != nil {
		blog.Errorf("parse env queryMetricFromBKMonitor error:%s", err.Error())
		return queryMetricFromBKMonitor
	}
	queryMetricFromBKMonitor = result
	parseEnv = true
	return queryMetricFromBKMonitor
}

// GetFloatData parse data to float64
func GetFloatData(response *bcsmonitor.QueryResponse) float64 {
	if len(response.Data.Result) == 0 {
		return 0
	}
	valueStr, ok := response.Data.Result[0].Value[1].(string)
	if !ok {
		return 0
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}
	return value
}

// GetInt64Data parse data to int64
func GetInt64Data(response *bcsmonitor.QueryResponse) int64 {
	if len(response.Data.Result) == 0 {
		return 0
	}
	valueStr, ok := response.Data.Result[0].Value[1].(string)
	if !ok {
		return 0
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

// GetIntData parse data to int
func GetIntData(response *bcsmonitor.QueryResponse) int {
	if len(response.Data.Result) == 0 {
		return 0
	}
	valueStr, ok := response.Data.Result[0].Value[1].(string)
	if !ok {
		return 0
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0
	}
	return value
}

func getDimensionPromql(dimension, objectType string) string {
	switch dimension {
	case types.DimensionDay:
		return "1d"
	case types.DimensionHour:
		return "1h"
	case types.DimensionMinute:
		if objectType == types.WorkloadType {
			return "2m"
		}
		return "10m"
	default:
		return ""
	}
}

func generatePodCondition(clusterID, namespace, workloadType, workloadName string) string {
	switch workloadType {
	case types.DeploymentType:
		return fmt.Sprintf(DeploymentPodCondition, clusterID, namespace, workloadName)
	default:
		return fmt.Sprintf(OtherPodCondition, clusterID, namespace, workloadName)
	}
}

func generateMesosPodCondition(clusterID, namespace, workloadName string) string {
	s := strings.Split(clusterID, "-")
	return fmt.Sprintf(MesosPodCondition, clusterID, namespace, workloadName, namespace, s[len(s)-1])
}

// getIncreasingIntervalDifference 获取数组内所有递增区间的差值，比如[0,1,2,1,2],返回结果为3
func getIncreasingIntervalDifference(initialNums []int) int {
	if len(initialNums) == 0 {
		return 0
	}
	var start, last, total int
	start = initialNums[0]
	last = initialNums[0]
	increasing := true
	for index := 1; index < len(initialNums); index++ {
		if initialNums[index] >= initialNums[last] {
			increasing = true
			last = index
			continue
		}
		if increasing {
			total += initialNums[last] - initialNums[start]
		}
		start = index
		last = index
		increasing = false
	}
	if increasing {
		total += initialNums[last] - initialNums[start]
	}
	return total
}

// fillMetrics 时间不连续的地方用0填充
func fillMetrics(firstTime float64, initialSlice [][]interface{}, step float64) []int {
	fillSlice := make([]int, 0)
	lastTime := initialSlice[0][0].(float64)
	if lastTime != firstTime {
		fillSlice = append(fillSlice, 0)
	} else {
		firstValue, err := strconv.Atoi(initialSlice[0][1].(string))
		if err != nil {
			return fillSlice
		}
		fillSlice = append(fillSlice, firstValue)
	}
	for index := 1; index < len(initialSlice); index++ {
		time := initialSlice[index][0].(float64)
		if (time - lastTime) != step {
			fillSlice = append(fillSlice, 0)
		}
		value, err := strconv.Atoi(initialSlice[index][1].(string))
		if err != nil {
			return fillSlice
		}
		fillSlice = append(fillSlice, value)
		lastTime = time
	}
	return fillSlice
}

func getK8sCPURequest(opts *types.JobCommonOpts) string {
	var condition string
	switch opts.ObjectType {
	case types.ClusterType:
		condition = fmt.Sprintf(ClusterCondition, opts.ClusterID)
	case types.NamespaceType:
		condition = fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace)
	case types.WorkloadType:
		condition = generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
	}
	return fmt.Sprintf(K8sCPURequest, condition)
}

func getK8sCPULimit(opts *types.JobCommonOpts) string {
	var condition string
	switch opts.ObjectType {
	case types.ClusterType:
		condition = fmt.Sprintf(ClusterCondition, opts.ClusterID)
	case types.NamespaceType:
		condition = fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace)
	case types.WorkloadType:
		condition = generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
	}
	return fmt.Sprintf(K8sCPULimits, condition)
}

func getK8sCpuUsage(opts *types.JobCommonOpts) string {
	switch opts.ObjectType {
	case types.ClusterType:
		jobName := PrometheusNodeExporterLabelJob

		if opts.IsBKMonitor && IfQueryFromBKMonitor() {
			jobName = BKMonitorNodeExporterLabelJob
		}
		return fmt.Sprintf(ClusterCPUUsage, opts.ClusterID, jobName, getDimensionPromql(opts.Dimension, opts.ObjectType))
	case types.NamespaceType:
		return fmt.Sprintf(NamespaceCPUUsage, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace),
			getDimensionPromql(opts.Dimension, opts.ObjectType), NamespaceSumCondition)
	case types.WorkloadType:
		condition := generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
		return fmt.Sprintf(WorkloadCPUUsage, condition, getDimensionPromql(opts.Dimension, opts.ObjectType))
	}
	return ""
}

func getK8sMemoryRequest(opts *types.JobCommonOpts) string {
	var condition string
	switch opts.ObjectType {
	case types.ClusterType:
		condition = fmt.Sprintf(ClusterCondition, opts.ClusterID)
	case types.NamespaceType:
		condition = fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace)
	case types.WorkloadType:
		condition = generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
	}
	return fmt.Sprintf(K8sMemoryRequest, condition)
}

func getK8sMemoryLimit(opts *types.JobCommonOpts) string {
	var condition string
	switch opts.ObjectType {
	case types.ClusterType:
		condition = fmt.Sprintf(ClusterCondition, opts.ClusterID)
	case types.NamespaceType:
		condition = fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace)
	case types.WorkloadType:
		condition = generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
	}
	return fmt.Sprintf(K8sMemoryLimit, condition)
}

func getK8sMemoryUsage(opts *types.JobCommonOpts) string {
	switch opts.ObjectType {
	case types.ClusterType:
		jobName := PrometheusNodeExporterLabelJob
		if opts.IsBKMonitor && IfQueryFromBKMonitor() {
			jobName = BKMonitorNodeExporterLabelJob
		}
		return fmt.Sprintf(ClusterMemoryUsed, opts.ClusterID, jobName, opts.ClusterID, jobName, opts.ClusterID, jobName,
			opts.ClusterID, jobName, opts.ClusterID, jobName)
	case types.NamespaceType:
		return fmt.Sprintf(NamespaceMemoryUsed, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace),
			NamespaceSumCondition)
	case types.WorkloadType:
		condition := generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
		return fmt.Sprintf(WorkloadMemoryUsed, condition)
	}
	return ""
}
