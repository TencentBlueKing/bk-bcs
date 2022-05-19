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
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
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
)

const (
	DeploymentPodCondition = "cluster_id=\"%s\", namespace=\"%s\",image!=\"\",pod=~\"%s-[0-9a-z]*-[0-9a-z]*$\"," +
		"container_name!=\"POD\""
	OtherPodCondition = "cluster_id=\"%s\", namespace=\"%s\",image!=\"\",pod=~\"%s-[0-9a-z]*$\"," +
		"container_name!=\"POD\""
	MesosPodCondition = "cluster_id=\"%s\", namespace=\"%s\",image!=\"\",name=~\".*.%s.%s.%s.*\"," +
		"container_name!=\"POD\""
	PodSumCondition       = "pod"
	MesosPodSumCondition  = "name"
	NamespaceCondition    = "cluster_id=\"%s\", namespace=\"%s\",image!=\"\""
	NamespaceSumCondition = "namespace"
	ClusterCondition      = "cluster_id=\"%s\",image!=\"\""
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

func getDimensionPromql(dimension string) string {
	switch dimension {
	case common.DimensionDay:
		return "1d"
	case common.DimensionHour:
		return "1h"
	case common.DimensionMinute:
		return "1m"
	default:
		return ""
	}
}

func generatePodCondition(clusterID, namespace, workloadType, workloadName string) string {
	switch workloadType {
	case common.DeploymentType:
		return fmt.Sprintf(DeploymentPodCondition, clusterID, namespace, workloadName)
	default:
		return fmt.Sprintf(OtherPodCondition, clusterID, namespace, workloadName)
	}
}

func generateMesosPodCondition(clusterID, namespace, workloadName string) string {
	s := strings.Split(clusterID, "-")
	return fmt.Sprintf(MesosPodCondition, clusterID, namespace, workloadName, namespace, s[len(s)-1])
}
