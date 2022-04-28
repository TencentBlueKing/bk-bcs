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
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
)

func (g *MetricGetter) getMesosWorkloadMemory(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, int64, float64, error) {
	var workloadMemoryLimit, workloadMemoryUsed int64
	var usage float64
	podCondition := generateMesosPodCondition(opts.ClusterID, opts.Namespace, opts.Name)
	podMemoryLimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosWorkloadMemoryLimit, podCondition, MesosPodSumCondition),
		opts.CurrentTime)
	if err != nil {
		return workloadMemoryLimit, workloadMemoryUsed, usage, fmt.Errorf("get pod metrics error: %v", err)
	}
	podMemoryUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(WorkloadMemoryUsed, podCondition, MesosPodSumCondition),
		opts.CurrentTime)
	if err != nil {
		return workloadMemoryLimit, workloadMemoryUsed, usage, fmt.Errorf("get pod metrics error: %v", err)
	}
	workloadMemoryLimit = GetInt64Data(podMemoryLimitMetric)
	workloadMemoryUsed = GetInt64Data(podMemoryUsedMetric)
	if workloadMemoryLimit != 0 {
		usage = float64(workloadMemoryUsed) / float64(workloadMemoryLimit)
	}
	return workloadMemoryLimit, workloadMemoryUsed, usage, nil
}

func (g *MetricGetter) getMesosWorkloadCPU(opts *common.JobCommonOpts,
	clients *common.Clients) (float64, float64, float64, error) {
	var workloadCPURequest, workloadCPUUsed float64
	var usage float64
	podCondition := generateMesosPodCondition(opts.ClusterID, opts.Namespace, opts.Name)
	podCPULimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosWorkloadCPULimit, podCondition, MesosPodSumCondition),
		opts.CurrentTime)
	if err != nil {
		return workloadCPURequest, workloadCPUUsed, usage, fmt.Errorf("get pod metrics error: %v", err)
	}

	podCPUUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(WorkloadCPUUsage, podCondition, getDimensionPromql(opts.Dimension), MesosPodSumCondition),
		opts.CurrentTime)
	if err != nil {
		return workloadCPURequest, workloadCPUUsed, usage, fmt.Errorf("get pod metrics error: %v", err)
	}
	workloadCPUUsed = GetFloatData(podCPUUsedMetric)
	if err != nil {
		return workloadCPURequest, workloadCPUUsed, usage, fmt.Errorf("get pod metrics error: %v", err)
	}
	workloadCPURequest = GetFloatData(podCPULimitMetric)
	// if workloadCPURequest != 0 {
	// 	usage = workloadCPUUsed / workloadCPURequest
	// }
	usage = workloadCPUUsed
	return workloadCPURequest, workloadCPUUsed, usage, nil
}

func (g *MetricGetter) getMesosNamespaceMemoryMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, int64, float64, error) {
	var memoryRequest, memoryUsed int64
	var usage float64
	memoryLimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosMemoryLimit,
			fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), NamespaceSumCondition),
		opts.CurrentTime)
	if err != nil {
		return memoryRequest, memoryUsed, usage, fmt.Errorf("get namespace metrics error: %v", err)
	}
	memoryUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(NamespaceMemoryUsed, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace),
			NamespaceSumCondition),
		opts.CurrentTime)
	if err != nil {
		return memoryRequest, memoryUsed, usage, fmt.Errorf("get namespace metrics error: %v", err)
	}
	memoryRequest = GetInt64Data(memoryLimitMetric)
	memoryUsed = GetInt64Data(memoryUsedMetric)
	if memoryRequest != 0 {
		usage = float64(memoryUsed) / float64(memoryRequest)
	}
	return memoryRequest, memoryUsed, usage, nil
}

func (g *MetricGetter) getMesosNamespaceCPUMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (float64, float64, float64, error) {
	var CPURequest, CPUUsed float64
	var usage float64
	query := fmt.Sprintf(MesosCPULimit,
		fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), NamespaceSumCondition)
	CPULimitMetric, err := clients.MonitorClient.QueryByPost(
		query,
		opts.CurrentTime)
	if err != nil {
		return CPURequest, CPUUsed, usage, fmt.Errorf("get namespace metrics error: %v", err)
	}

	CPUUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(NamespaceCPUUsage, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace),
			getDimensionPromql(opts.Dimension), NamespaceSumCondition),
		opts.CurrentTime)
	if err != nil {
		return CPURequest, CPUUsed, usage, fmt.Errorf("get namespace metrics error: %v", err)
	}
	CPUUsed = GetFloatData(CPUUsedMetric)
	CPURequest = GetFloatData(CPULimitMetric)
	usage = CPUUsed
	return CPURequest, CPUUsed, usage, nil
}

func (g *MetricGetter) getMesosNodeCount(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, int64, error) {
	var nodeCount, availableNode int64

	nodes, err := clients.CmCli.Cli.ListNodesInCluster(clients.CmCli.Ctx, &cm.ListNodesInClusterRequest{
		ClusterID: opts.ClusterID,
	})
	if err != nil {
		return nodeCount, availableNode, fmt.Errorf("get cluster metrics error:%v", err)
	}
	nodeCount = int64(len(nodes.Data))
	for key := range nodes.Data {
		if nodes.Data[key].Status == "RUNNING" {
			availableNode++
		}
	}
	return nodeCount, availableNode, nil
}

// TODO
func (g *MetricGetter) getMesosClusterMemoryMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, int64, int64, float64, error) {
	var totalMemory, memoryRequest, memoryUsed int64
	var usage float64
	memoryLimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosMemoryLimit,
			fmt.Sprintf(ClusterCondition, opts.ClusterID), ClusterSumCondition), opts.CurrentTime)
	if err != nil {
		return totalMemory, memoryRequest, memoryUsed, usage, fmt.Errorf("get cluster metrics error: %v", err)
	}
	memoryUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterMemoryUsed, fmt.Sprintf(ClusterCondition, opts.ClusterID), ClusterSumCondition),
		opts.CurrentTime)
	if err != nil {
		return totalMemory, memoryRequest, memoryUsed, usage, fmt.Errorf("get cluster metrics error: %v", err)
	}
	totalMemoryMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterTotalMemory, opts.ClusterID),
		opts.CurrentTime)
	if err != nil {
		return totalMemory, memoryRequest, memoryUsed, usage, fmt.Errorf("get cluster metrics error: %v", err)
	}
	totalMemory = GetInt64Data(totalMemoryMetric)
	memoryRequest = GetInt64Data(memoryLimitMetric)
	memoryUsed = GetInt64Data(memoryUsedMetric)
	if totalMemory != 0 {
		usage = float64(memoryUsed) / float64(totalMemory)
	}
	return totalMemory, memoryRequest, memoryUsed, usage, nil
}

func (g *MetricGetter) getMesosClusterCPUMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (float64, float64, float64, float64, error) {
	var totalCPU, CPURequest, CPUUsed float64
	var usage float64
	CPULimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosCPULimit,
			fmt.Sprintf(ClusterCondition, opts.ClusterID), ClusterSumCondition),
		opts.CurrentTime)
	if err != nil {
		return totalCPU, CPURequest, CPUUsed, usage, fmt.Errorf("get cluster metrics error: %v", err)
	}

	CPUUsageMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterCPUUsage, opts.ClusterID, getDimensionPromql(opts.Dimension)),
		opts.CurrentTime)
	if err != nil {
		return totalCPU, CPURequest, CPUUsed, usage, fmt.Errorf("get cluster metrics error: %v", err)
	}
	CPUUsed = GetFloatData(CPUUsageMetric)
	totalCPUMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterTotalCPU, opts.ClusterID),
		opts.CurrentTime)
	if err != nil {
		return totalCPU, CPURequest, CPUUsed, usage, fmt.Errorf("get cluster metrics error: %v", err)
	}

	CPURequest = GetFloatData(CPULimitMetric)
	totalCPU = GetFloatData(totalCPUMetric)
	if totalCPU != 0 {
		usage = CPUUsed / totalCPU
	}
	return totalCPU, CPURequest, CPUUsed, usage, nil
}
