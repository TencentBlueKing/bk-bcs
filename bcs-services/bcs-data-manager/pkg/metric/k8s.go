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
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

func (g *MetricGetter) getK8sClusterCPUMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	cpuMetrics := &types.CPUMetrics{}
	CPURequestMetric, err := clients.MonitorClient.QueryByPost(getK8sCPURequest(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	cpuMetrics.CPURequest = GetFloatData(CPURequestMetric)

	CPUUsedMetric, err := clients.MonitorClient.QueryByPost(getK8sCpuUsage(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	cpuMetrics.CPUUsed = GetFloatData(CPUUsedMetric)

	totalCPUMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterTotalCPU, opts.ClusterID), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	cpuMetrics.TotalCPU = GetFloatData(totalCPUMetric)
	if cpuMetrics.TotalCPU != 0 {
		cpuMetrics.CPUUsage = cpuMetrics.CPUUsed / cpuMetrics.TotalCPU
	}
	CPULimitMetric, err := clients.MonitorClient.QueryByPost(getK8sCPULimit(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	cpuMetrics.CPULimit = GetFloatData(CPULimitMetric)
	return cpuMetrics, nil
}

func (g *MetricGetter) getK8sClusterMemoryMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	memoryMetrics := &types.MemoryMetrics{}
	memoryRequestMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryRequest(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	memoryMetrics.MemoryRequest = GetInt64Data(memoryRequestMetric)
	memoryUsedMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryUsage(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	memoryMetrics.MemoryUsed = GetInt64Data(memoryUsedMetric)
	totalMemoryMetric, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(ClusterTotalMemory, opts.ClusterID),
		opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	memoryMetrics.TotalMemory = GetInt64Data(totalMemoryMetric)
	if memoryMetrics.TotalMemory != 0 {
		memoryMetrics.MemoryUsage = float64(memoryMetrics.MemoryUsed) / float64(memoryMetrics.TotalMemory)
	}
	memoryLimitMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryRequest(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get cluster metrics error: %v", err)
	}
	memoryMetrics.MemoryLimit = GetInt64Data(memoryLimitMetric)
	return memoryMetrics, nil
}

// getK8sNodeCount get k8s node count
func (g *MetricGetter) getK8sNodeCount(opts *types.JobCommonOpts,
	clients *types.Clients) (int64, int64, error) {
	var nodeCount, availableNode int64
	start := time.Now()
	nodes, err := clients.K8sStorageCli.QueryK8SNode(opts.ClusterID)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "QueryK8SNode",
			"GET", err, start)
		return nodeCount, availableNode, fmt.Errorf("get cluster metrics error:%v", err)
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "QueryK8SNode",
		"GET", err, start)
	// Note: k8s cluster use storage get nodes
	nodeCount = int64(len(nodes))
	for key := range nodes {
		if nodes[key].Data.Status.Phase == v1.NodeRunning {
			availableNode++
		}
	}
	return nodeCount, availableNode, nil
}

func (g *MetricGetter) getK8sNamespaceCPUMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	cpuMetrics := &types.CPUMetrics{}
	CPURequestMetric, err := clients.MonitorClient.QueryByPost(getK8sCPURequest(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	cpuMetrics.CPURequest = GetFloatData(CPURequestMetric)

	CPUUsedMetrics, err := clients.MonitorClient.QueryByPost(getK8sCpuUsage(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	cpuMetrics.CPUUsed = GetFloatData(CPUUsedMetrics)
	cpuMetrics.CPUUsage = GetFloatData(CPUUsedMetrics)
	// if CPURequest != 0 {
	// 	usage = CPUUsed / CPURequest
	// }
	CPULimitsMetric, err := clients.MonitorClient.QueryByPost(getK8sCPULimit(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	cpuMetrics.CPULimit = GetFloatData(CPULimitsMetric)
	return cpuMetrics, nil
}

func (g *MetricGetter) getK8SWorkloadCPU(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	cpuMetrics := &types.CPUMetrics{}
	podCPURequestMetric, err := clients.MonitorClient.QueryByPost(getK8sCPURequest(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	cpuMetrics.CPURequest = GetFloatData(podCPURequestMetric)

	podCPUUsedMetric, err := clients.MonitorClient.QueryByPost(getK8sCpuUsage(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	cpuMetrics.CPUUsed = GetFloatData(podCPUUsedMetric)
	cpuMetrics.CPUUsage = GetFloatData(podCPUUsedMetric)

	// workload limits
	podCPULimitsMetric, err := clients.MonitorClient.QueryByPost(getK8sCPULimit(opts), opts.CurrentTime)
	if err != nil {
		return cpuMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	cpuMetrics.CPULimit = GetFloatData(podCPULimitsMetric)
	return cpuMetrics, nil
}

func (g *MetricGetter) getK8sWorkloadMemory(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	memoryMetrics := &types.MemoryMetrics{}
	podMemoryRequestMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryRequest(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	memoryMetrics.MemoryRequest = GetInt64Data(podMemoryRequestMetric)

	podMemoryUsedMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryUsage(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	memoryMetrics.MemoryUsed = GetInt64Data(podMemoryUsedMetric)
	if memoryMetrics.MemoryRequest != 0 {
		memoryMetrics.MemoryUsage = float64(memoryMetrics.MemoryUsed) / float64(memoryMetrics.MemoryRequest)
	}

	podMemoryLimitsMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryLimit(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get pod metrics error: %v", err)
	}
	memoryMetrics.MemoryLimit = GetInt64Data(podMemoryLimitsMetric)
	return memoryMetrics, nil
}

func (g *MetricGetter) getK8sNamespaceMemoryMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	memoryMetrics := &types.MemoryMetrics{}
	memoryRequestMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryRequest(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	memoryMetrics.MemoryRequest = GetInt64Data(memoryRequestMetric)
	memoryUsedMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryUsage(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	memoryMetrics.MemoryUsed = GetInt64Data(memoryUsedMetric)
	if memoryMetrics.MemoryRequest != 0 {
		memoryMetrics.MemoryUsage = float64(memoryMetrics.MemoryUsed) / float64(memoryMetrics.MemoryRequest)
	}
	memoryLimitsMetric, err := clients.MonitorClient.QueryByPost(getK8sMemoryRequest(opts), opts.CurrentTime)
	if err != nil {
		return memoryMetrics, fmt.Errorf("get namespace metrics error: %v", err)
	}
	memoryMetrics.MemoryLimit = GetInt64Data(memoryLimitsMetric)
	return memoryMetrics, nil
}
