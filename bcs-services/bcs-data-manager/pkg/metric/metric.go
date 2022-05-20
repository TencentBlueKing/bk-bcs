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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"strconv"
	"time"

	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// Server metric interface
type Server interface {
	GetWorkloadCPUMetrics(opts *common.JobCommonOpts, clients *common.Clients) (float64, float64, float64, error)
	GetWorkloadMemoryMetrics(opts *common.JobCommonOpts, clients *common.Clients) (int64, int64, float64, error)
	GetNamespaceCPUMetrics(opts *common.JobCommonOpts, clients *common.Clients) (float64, float64, float64, error)
	GetNamespaceMemoryMetrics(opts *common.JobCommonOpts, clients *common.Clients) (int64, int64, float64, error)
	GetClusterCPUMetrics(opts *common.JobCommonOpts, clients *common.Clients) (float64,
		float64, float64, float64, error)
	GetClusterMemoryMetrics(opts *common.JobCommonOpts, clients *common.Clients) (int64,
		int64, int64, float64, error)
	GetInstanceCount(opts *common.JobCommonOpts, clients *common.Clients) (int64, error)
	GetClusterNodeMetrics(opts *common.JobCommonOpts,
		clients *common.Clients) (string, []*bcsdatamanager.NodeQuantile, error)
	GetClusterNodeCount(opts *common.JobCommonOpts, clients *common.Clients) (int64, int64, error)
}

// MetricGetter metric getter
type MetricGetter struct{}

// GetClusterCPUMetrics get cluster cpu metrics
func (g *MetricGetter) GetClusterCPUMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (float64, float64, float64, float64, error) {
	switch opts.ClusterType {
	case common.Kubernetes:
		return g.getK8sClusterCPUMetrics(opts, clients)
	case common.Mesos:
		return g.getMesosClusterCPUMetrics(opts, clients)
	default:
		return 0, 0, 0, 0, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetClusterMemoryMetrics get cluster memory metrics
func (g *MetricGetter) GetClusterMemoryMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, int64, int64, float64, error) {
	switch opts.ClusterType {
	case common.Kubernetes:
		return g.getK8sClusterMemoryMetrics(opts, clients)
	case common.Mesos:
		return g.getMesosClusterMemoryMetrics(opts, clients)
	default:
		return 0, 0, 0, 0, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetClusterNodeMetrics get cluster node metrics
func (g *MetricGetter) GetClusterNodeMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (string, []*bcsdatamanager.NodeQuantile, error) {
	var minUsageNode string
	nodeQuantie := make([]*bcsdatamanager.NodeQuantile, 0)
	start := time.Now()
	cluster, err := clients.CmCli.Cli.GetCluster(clients.CmCli.Ctx, &cm.GetClusterReq{
		ClusterID: opts.ClusterID,
	})
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "GetCluster",
			"GET", err, start)
		return minUsageNode, nodeQuantie, err
	}
	prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListNodesInCluster",
		"GET", err, start)
	queryCond := fmt.Sprintf(ClusterCondition, opts.ClusterID)

	for key := range cluster.Data.Master {
		queryCond = fmt.Sprintf("%s,instance!=\"%s:9100\"", queryCond, cluster.Data.Master[key].InnerIP)
	}
	nodeCPUQuery := fmt.Sprintf(NodeCPUUsage, opts.ClusterID, queryCond, opts.ClusterID, queryCond)
	nodeCpuUsageList, err := clients.MonitorClient.QueryByPost(nodeCPUQuery, opts.CurrentTime)
	if err != nil {
		return minUsageNode, nodeQuantie, err
	}
	var minUsage float64
	for key := range nodeCpuUsageList.Data.Result {
		usage, ok := nodeCpuUsageList.Data.Result[key].Value[1].(string)
		if ok {
			nodeUsage, err := strconv.ParseFloat(usage, 64)
			if err != nil {
				continue
			}
			if nodeUsage < minUsage {
				minUsage = nodeUsage
				minUsageNode = nodeCpuUsageList.Data.Result[key].Metric["node"]
			}
		}
	}

	nodeQuantieResponse, err := clients.MonitorClient.QueryByPost(fmt.Sprintf(NodeUsageQuantile,
		"0.5", nodeCPUQuery), opts.CurrentTime)
	if err != nil {
		return minUsageNode, nodeQuantie, err
	}
	if len(nodeQuantieResponse.Data.Result) != 0 {
		nodeQuantileCPU, ok := nodeQuantieResponse.Data.Result[0].Value[1].(string)
		if ok {
			quantile := &bcsdatamanager.NodeQuantile{
				Percentage:   "50",
				NodeCPUUsage: nodeQuantileCPU,
			}
			nodeQuantie = append(nodeQuantie, quantile)
		}
	}
	return minUsageNode, nodeQuantie, nil
}

// GetClusterNodeCount get cluster node count
func (g *MetricGetter) GetClusterNodeCount(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, int64, error) {
	switch opts.ClusterType {
	case common.Kubernetes:
		return g.getK8sNodeCount(opts, clients)
	case common.Mesos:
		return g.getMesosNodeCount(opts, clients)
	default:
		return 0, 0, fmt.Errorf("wrong clusterType:%s", opts.ClusterType)
	}
}

// GetNamespaceCPUMetrics get namespace cpu metrics
func (g *MetricGetter) GetNamespaceCPUMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (float64, float64, float64, error) {
	switch opts.ClusterType {
	case common.Kubernetes:
		return g.getK8sNamespaceCPUMetrics(opts, clients)
	case common.Mesos:
		return g.getMesosNamespaceCPUMetrics(opts, clients)
	default:
		return 0, 0, 0, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetNamespaceMemoryMetrics get namespace memory metric
func (g *MetricGetter) GetNamespaceMemoryMetrics(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, int64, float64, error) {
	switch opts.ClusterType {
	case common.Kubernetes:
		return g.getK8sNamespaceMemoryMetrics(opts, clients)
	case common.Mesos:
		return g.getMesosNamespaceMemoryMetrics(opts, clients)
	default:
		return 0, 0, 0, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetNamespaceResourceLimit get namespace resource limit
func (g *MetricGetter) GetNamespaceResourceLimit(opts *common.JobCommonOpts,
	clients *common.Clients) (float64, int64, float64, int64, error) {
	var requestCPU, limitCPU float64
	var requestMemory, limitMemory int64
	requestCPUResponse, err := clients.MonitorClient.QueryByPost(fmt.Sprintf(NamespaceResourceQuota,
		fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), "requests.cpu"),
		opts.CurrentTime)
	if err != nil {
		return requestCPU, requestMemory, limitCPU, limitMemory, fmt.Errorf("get namespace metrics error: %v", err)
	}
	requestCPU = GetFloatData(requestCPUResponse)
	limitCPUResponse, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(NamespaceResourceQuota, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), "limit.cpu"),
		opts.CurrentTime)
	if err != nil {
		return requestCPU, requestMemory, limitCPU, limitMemory, fmt.Errorf("get namespace metrics error: %v", err)
	}
	limitCPU = GetFloatData(limitCPUResponse)
	requestMemoryResponse, err := clients.MonitorClient.QueryByPost(fmt.Sprintf(NamespaceResourceQuota,
		fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), "requests.memory"),
		opts.CurrentTime)
	if err != nil {
		return requestCPU, requestMemory, limitCPU, limitMemory, fmt.Errorf("get namespace metrics error: %v", err)
	}
	requestMemory = GetInt64Data(requestMemoryResponse)
	limitMemoryResponse, err := clients.MonitorClient.QueryByPost(fmt.Sprintf(NamespaceResourceQuota,
		fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), "limit.memory"),
		opts.CurrentTime)
	if err != nil {
		return requestCPU, requestMemory, limitCPU, limitMemory, fmt.Errorf("get namespace metrics error: %v", err)
	}
	limitMemory = GetInt64Data(limitMemoryResponse)
	return requestCPU, requestMemory, limitCPU, limitMemory, nil
}

// GetNamespaceInstanceCount get namespace instance count
func (g *MetricGetter) GetNamespaceInstanceCount(opts *common.JobCommonOpts,
	clients *common.Clients) (int64, error) {
	var count int64
	response, err := clients.MonitorClient.QueryByPost(
		fmt.Sprintf(WorkloadInstance, fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), PodSumCondition),
		opts.CurrentTime)
	if err != nil {
		return count, fmt.Errorf("get pod metrics error: %v", err)
	}
	if len(response.Data.Result) == 0 {
		return 0, nil
	}
	value, ok := response.Data.Result[0].Value[1].(string)
	if !ok {
		return count, fmt.Errorf("get count error, wrong result type: %t", response.Data.Result[0].Value[1])
	}
	count, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		return count, fmt.Errorf("parse result to int64 error: %v", err)
	}
	return count, nil
}

// GetWorkloadCPUMetrics get workload cpu metrics
func (g *MetricGetter) GetWorkloadCPUMetrics(opts *common.JobCommonOpts, clients *common.Clients) (float64,
	float64, float64, error) {
	switch opts.ClusterType {
	case common.Kubernetes:
		return g.getK8SWorkloadCPU(opts, clients)
	case common.Mesos:
		return g.getMesosWorkloadCPU(opts, clients)
	default:
		return 0, 0, 0, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetWorkloadMemoryMetrics get workload memory metrics
func (g *MetricGetter) GetWorkloadMemoryMetrics(opts *common.JobCommonOpts, clients *common.Clients) (int64,
	int64, float64, error) {
	switch opts.ClusterType {
	case common.Kubernetes:
		return g.getK8sWorkloadMemory(opts, clients)
	case common.Mesos:
		return g.getMesosWorkloadMemory(opts, clients)
	default:
		return 0, 0, 0, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetInstanceCount get instance count
func (g *MetricGetter) GetInstanceCount(opts *common.JobCommonOpts, clients *common.Clients) (int64, error) {
	var count int64
	var query string
	podSumCondition := PodSumCondition
	if opts.ClusterType == common.Mesos {
		podSumCondition = MesosPodSumCondition
	}
	switch opts.ObjectType {
	case common.ClusterType:
		query = fmt.Sprintf(WorkloadInstance, fmt.Sprintf(ClusterCondition, opts.ClusterID), podSumCondition)
	case common.NamespaceType:
		query = fmt.Sprintf(WorkloadInstance,
			fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), podSumCondition)
	case common.WorkloadType:
		podCondition := generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.Name)
		query = fmt.Sprintf(WorkloadInstance, podCondition, podSumCondition)
	default:
		return count, fmt.Errorf("wrong object type: %s", opts.ObjectType)
	}

	response, err := clients.MonitorClient.QueryByPost(query, opts.CurrentTime)
	if err != nil {
		return count, fmt.Errorf("get pod metrics error: %v", err)
	}
	if len(response.Data.Result) == 0 {
		return 0, nil
	}
	value, ok := response.Data.Result[0].Value[1].(string)
	if !ok {
		return count, fmt.Errorf("get count error, wrong result type: %t", response.Data.Result[0].Value[1])
	}
	count, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		return count, fmt.Errorf("parse result to int64 error: %v", err)
	}
	return count, nil
}
