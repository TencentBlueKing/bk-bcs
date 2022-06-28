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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
)

const (
	WorkloadCPUUsage = "sum(rate(container_cpu_usage_seconds_total{%s}[%s]))by(%s)"
	ClusterCPUUsage  = "sum(irate(node_cpu_seconds_total{cluster_id=\"%s\", job=\"node-exporter\", " +
		"mode!=\"idle\"}[%s]))"
	NamespaceCPUUsage        = "sum(rate(container_cpu_usage_seconds_total{%s}[%s]))by(%s)"
	InstanceCount            = "count(sum(rate(container_cpu_usage_seconds_total{%s}[%s]))by(%s))"
	K8sWorkloadCPURequest    = "sum(sum(kube_pod_container_resource_requests_cpu_cores{%s})by(%s))"
	K8sCPURequest            = "sum(kube_pod_container_resource_requests_cpu_cores{%s})by(%s)"
	WorkloadMemoryUsed       = "sum(sum(container_memory_rss{%s})by(%s))"
	NamespaceMemoryUsed      = "sum(container_memory_rss{%s})by(%s)"
	ClusterMemoryUsed        = "sum(container_memory_rss{%s})by(%s)"
	K8sWorkloadMemoryRequest = "sum(sum(kube_pod_container_resource_requests_memory_bytes{%s})by(%s))"
	K8sMemoryRequest         = "sum(kube_pod_container_resource_requests_memory_bytes{%s})by(%s)"
	WorkloadInstance         = "count(sum(container_memory_rss{%s}) by (%s))"
	NamespaceResourceQuota   = "kube_resourcequota{%s, type=\"hard\", %s}"
	PromMasterIP             = "sum(kube_node_role{cluster_id=\"%s\"})by(node)"
	PromNodeIP               = "sum(kube_node_created{cluster_id=\"%s\",%s})by(node)"
	NodeCount                = "count(sum(kube_node_created{cluster_id=\"%s\",%s})by(node))"
	ClusterTotalCPU          = "sum(machine_cpu_cores{cluster_id=\"%s\"})by(cluster_id)"
	ClusterTotalMemory       = "sum(machine_memory_bytes{cluster_id=\"%s\"})by(cluster_id)"
	NodeUsageQuantile        = "quantile(%s,%s)"
	MesosWorkloadMemoryLimit = "sum(sum(container_spec_memory_limit_bytes{%s})by(%s))"
	MesosWorkloadCPULimit    = "sum(sum(container_spec_cpu_quota{%s})by(%s)/100000)"
	MesosMemoryLimit         = "sum(container_spec_memory_limit_bytes{%s})by(%s)"
	MesosCPULimit            = "sum(container_spec_cpu_quota{%s})by(%s)/100000"
	NodeCPUUsage             = "sum(irate(node_cpu_seconds_total{cluster_id=\"%s\", job=\"node-exporter\", " +
		"mode!=\"idle\",%s}[1m]))by(instance)/" +
		"sum(count without(cpu, mode) (node_cpu_seconds_total{cluster_id=\"%s\", job=\"node-exporter\", " +
		"mode=\"idle\",%s}))by(instance)"
	ClusterAutoscalerUpCount = "sum(kube_event_unique_events_total{cluster_id=\"%s\", " +
		"source=\"/cluster-autoscaler\",reason=\"ScaledUpGroup\",involved_object_namespace=\"bcs-system\"})"
	ClusterAutoscalerDownCount = "sum(kube_event_unique_events_total{cluster_id=\"%s\", " +
		"source=\"/cluster-autoscaler\",reason=\"ScaleDown\", involved_object_namespace=\"bcs-system\"})"
	GeneralPodAutoscalerCount = "kube_event_unique_events_total{cluster_id=\"%s\", " +
		"involved_object_kind=\"GeneralPodAutoscaler\",involved_object_name=\"%s\",namespace=\"%s\"," +
		"source=\"/pod-autoscaler\",reason=\"SuccessfulRescale\"}"
	HorizontalPodAutoscalerCount = "kube_event_unique_events_total{cluster_id=\"%s\", " +
		"involved_object_kind=\"HorizontalPodAutoscaler\",involved_object_name=\"%s\",namespace=\"%s\"," +
		"source=\"/horizontal-pod-autoscaler\",reason=\"SuccessfulRescale\"}"
	MinOverTime = "min_over_time(%s[%s])"
	MaxOverTime = "max_over_time(%s[%s])"
)

const (
	DeploymentPodCondition = "cluster_id=\"%s\", namespace=\"%s\",pod=~\"%s-[0-9a-z]*-[0-9a-z]*$\"," +
		"container_name!=\"POD\""
	OtherPodCondition = "cluster_id=\"%s\", namespace=\"%s\",pod=~\"%s-[0-9a-z]*$\"," +
		"container_name!=\"POD\""
	MesosPodCondition = "cluster_id=\"%s\", namespace=\"%s\",name=~\".*.%s.%s.%s.*\"," +
		"container_name!=\"POD\""
	PodSumCondition       = "pod"
	MesosPodSumCondition  = "name"
	NamespaceCondition    = "cluster_id=\"%s\", namespace=\"%s\""
	NamespaceSumCondition = "namespace"
	ClusterCondition      = "cluster_id=\"%s\""
	ClusterSumCondition   = "cluster_id"
)

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

func getDimensionPromql(dimension string) string {
	switch dimension {
	case types.DimensionDay:
		return "1d"
	case types.DimensionHour:
		return "1h"
	case types.DimensionMinute:
		return "1m"
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

// 获取数组内所有递增区间的差值，比如[0,1,2,1,2],返回结果为3
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

// 时间不连续的地方用0填充
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
