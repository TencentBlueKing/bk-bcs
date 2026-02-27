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

package metric

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

// getMesosWorkloadMemory get mesos memory
func (g *MetricGetter) getMesosWorkloadMemory(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	memoryMetrics := &types.MemoryMetrics{}
	podCondition := generateMesosPodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadName)
	podMemoryLimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosWorkloadMemoryLimit, podCondition, MesosPodSumCondition),
		opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	memoryMetrics.MemoryLimit = GetInt64Data(podMemoryLimitMetric)
	memoryMetrics.MemoryRequest = GetInt64Data(podMemoryLimitMetric)
	podMemoryUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(WorkloadMemoryUsed, podCondition),
		opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	memoryMetrics.MemoryUsed = GetInt64Data(podMemoryUsedMetric)
	if memoryMetrics.MemoryLimit != 0 {
		memoryMetrics.MemoryUsage = float64(memoryMetrics.MemoryLimit) / float64(memoryMetrics.MemoryLimit)
	}
	return memoryMetrics, nil
}

// getMesosWorkloadCPU get mesos cpu
func (g *MetricGetter) getMesosWorkloadCPU(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	cpuMetrics := &types.CPUMetrics{}
	podCondition := generateMesosPodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadName)
	podCPULimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosWorkloadCPULimit, podCondition, MesosPodSumCondition),
		opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	cpuMetrics.CPULimit = GetFloatData(podCPULimitMetric)
	cpuMetrics.CPURequest = GetFloatData(podCPULimitMetric)
	podCPUUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(WorkloadCPUUsage, podCondition, getDimensionPromql(opts.Dimension, opts.ObjectType)),
		opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	cpuMetrics.CPUUsed = GetFloatData(podCPUUsedMetric)
	// if workloadCPURequest != 0 {
	// 	usage = workloadCPUUsed / workloadCPURequest
	// }
	cpuMetrics.CPUUsage = cpuMetrics.CPUUsed
	return cpuMetrics, nil
}

// getMesosNamespaceMemoryMetrics get namespace memory
func (g *MetricGetter) getMesosNamespaceMemoryMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	memoryMetrics := &types.MemoryMetrics{}
	memoryLimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosMemoryLimit,
			fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), NamespaceSumCondition),
		opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	memoryMetrics.MemoryLimit = GetInt64Data(memoryLimitMetric)
	memoryMetrics.MemoryRequest = GetInt64Data(memoryLimitMetric)
	memoryUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(NamespaceMemoryUsed, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace),
			NamespaceSumCondition),
		opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	memoryMetrics.MemoryRequest = GetInt64Data(memoryLimitMetric)
	memoryMetrics.MemoryUsed = GetInt64Data(memoryUsedMetric)
	if memoryMetrics.MemoryRequest != 0 {
		memoryMetrics.MemoryUsage = float64(memoryMetrics.MemoryUsed) / float64(memoryMetrics.MemoryRequest)
	}
	return memoryMetrics, nil
}

// getMesosNamespaceCPUMetrics get namespace cpu
func (g *MetricGetter) getMesosNamespaceCPUMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	cpuMetrics := &types.CPUMetrics{}
	query := fmt.Sprintf(MesosCPULimit,
		fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), NamespaceSumCondition)
	CPULimitMetric, err := clients.MonitorClient.QueryByPost(
		query,
		opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}

	CPUUsedMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(NamespaceCPUUsage, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace),
			getDimensionPromql(opts.Dimension, opts.ObjectType), NamespaceSumCondition),
		opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	cpuMetrics.CPUUsed = GetFloatData(CPUUsedMetric)
	cpuMetrics.CPURequest = GetFloatData(CPULimitMetric)
	cpuMetrics.CPULimit = GetFloatData(CPULimitMetric)
	cpuMetrics.CPUUsage = GetFloatData(CPUUsedMetric)
	return cpuMetrics, nil
}

// getMesosNodeCount get mesos node count
func (g *MetricGetter) getMesosNodeCount(ctx context.Context, opts *types.JobCommonOpts) (int64, int64, error) {
	var nodeCount, availableNode int64
	start := time.Now()
	cmCli, close, err := clustermanager.GetClient(common.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nodeCount, availableNode, err
	}
	nodes, err := cmCli.ListNodesInCluster(ctx, &cm.ListNodesInClusterRequest{
		ClusterID: opts.ClusterID,
	})
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListNodesInCluster",
			"GET", err, start)
		return nodeCount, availableNode, fmt.Errorf("get cluster metrics error:%v", err)
	}
	prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListNodesInCluster",
		"GET", err, start)
	nodeCount = int64(len(nodes.Data))
	for key := range nodes.Data {
		if nodes.Data[key].Status == "RUNNING" {
			availableNode++
		}
	}
	return nodeCount, availableNode, nil
}

// getMesosClusterMemoryMetrics get mesos cluster memory
func (g *MetricGetter) getMesosClusterMemoryMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	memoryMetrics := &types.MemoryMetrics{}
	memoryLimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosMemoryLimit,
			fmt.Sprintf(ClusterCondition, opts.ClusterID), ClusterSumCondition), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	memoryUsedMetric, err := clients.MonitorClient.QueryByPost(
		getK8sMemoryUsage(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	totalMemoryMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterTotalMemory, opts.ClusterID),
		opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	memoryMetrics.TotalMemory = GetInt64Data(totalMemoryMetric)
	memoryMetrics.MemoryLimit = GetInt64Data(memoryLimitMetric)
	memoryMetrics.MemoryRequest = GetInt64Data(memoryLimitMetric)
	memoryMetrics.MemoryUsed = GetInt64Data(memoryUsedMetric)
	if memoryMetrics.TotalMemory != 0 {
		memoryMetrics.MemoryUsage = float64(memoryMetrics.MemoryUsed) / float64(memoryMetrics.TotalMemory)
	}
	return memoryMetrics, nil
}

// getMesosClusterCPUMetrics get cluster cpu
func (g *MetricGetter) getMesosClusterCPUMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	cpuMetrics := &types.CPUMetrics{}
	CPULimitMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(MesosCPULimit,
			fmt.Sprintf(ClusterCondition, opts.ClusterID), ClusterSumCondition),
		opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}

	CPUUsageMetric, err := clients.MonitorClient.QueryByPost(getK8sCpuUsage(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	cpuMetrics.CPUUsed = GetFloatData(CPUUsageMetric)
	totalCPUMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterTotalCPU, opts.ClusterID),
		opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}

	cpuMetrics.CPURequest = GetFloatData(CPULimitMetric)
	cpuMetrics.CPULimit = GetFloatData(CPULimitMetric)
	cpuMetrics.TotalCPU = GetFloatData(totalCPUMetric)
	if cpuMetrics.TotalCPU != 0 {
		cpuMetrics.CPUUsage = cpuMetrics.CPUUsed / cpuMetrics.TotalCPU
	}
	return cpuMetrics, nil
}
