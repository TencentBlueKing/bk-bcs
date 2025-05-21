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

// Package metric xxx
package metric

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// Server metric interface
type Server interface {
	GetWorkloadCPUMetrics(opts *types.JobCommonOpts, clients *types.Clients) (*types.CPUMetrics, error)
	GetWorkloadMemoryMetrics(opts *types.JobCommonOpts, clients *types.Clients) (*types.MemoryMetrics, error)
	GetNamespaceCPUMetrics(opts *types.JobCommonOpts, clients *types.Clients) (*types.CPUMetrics, error)
	GetNamespaceMemoryMetrics(opts *types.JobCommonOpts, clients *types.Clients) (*types.MemoryMetrics, error)
	GetClusterCPUMetrics(opts *types.JobCommonOpts, clients *types.Clients) (*types.CPUMetrics, error)
	GetClusterMemoryMetrics(opts *types.JobCommonOpts, clients *types.Clients) (*types.MemoryMetrics, error)
	GetInstanceCount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error)
	GetClusterNodeMetrics(opts *types.JobCommonOpts,
		clients *types.Clients) (string, []*bcsdatamanager.NodeQuantile, error)
	GetClusterNodeCount(ctx context.Context, opts *types.JobCommonOpts, clients *types.Clients) (int64, int64, error)
	GetPodAutoscalerCount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error)
	GetCACount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error)
}

// MetricGetter metric getter
type MetricGetter struct{} // nolint

// GetClusterCPUMetrics get cluster cpu metrics
func (g *MetricGetter) GetClusterCPUMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	switch opts.ClusterType {
	case types.Kubernetes:
		return g.getK8sClusterCPUMetrics(opts, clients)
	case types.Mesos:
		return g.getMesosClusterCPUMetrics(opts, clients)
	default:
		return nil, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetClusterMemoryMetrics get cluster memory metrics
func (g *MetricGetter) GetClusterMemoryMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	switch opts.ClusterType {
	case types.Kubernetes:
		return g.getK8sClusterMemoryMetrics(opts, clients)
	case types.Mesos:
		return g.getMesosClusterMemoryMetrics(opts, clients)
	default:
		return nil, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetClusterNodeMetrics get cluster node metrics
func (g *MetricGetter) GetClusterNodeMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (string, []*bcsdatamanager.NodeQuantile, error) {
	var minUsageNode string
	nodeQuantie := make([]*bcsdatamanager.NodeQuantile, 0)
	// start := time.Now()
	// cluster, err := clients.CmCli.Cli.GetCluster(clients.CmCli.Ctx, &cm.GetClusterReq{
	//	ClusterID: opts.ClusterID,
	// })
	// if err != nil {
	//	prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "GetCluster",
	//		"GET", err, start)
	//	return minUsageNode, nodeQuantie, err
	// }
	// prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListNodesInCluster",
	//	"GET", err, start)
	// queryCond := fmt.Sprintf(ClusterCondition, opts.ClusterID)
	//
	// for key := range cluster.Data.Master {
	//	queryCond = fmt.Sprintf("%s,instance!=\"%s:9100\"", queryCond, cluster.Data.Master[key].InnerIP)
	// }
	// nodeCPUQuery := fmt.Sprintf(NodeCPUUsage, opts.ClusterID, queryCond, opts.ClusterID, queryCond)
	// nodeCpuUsageList, err := clients.MonitorClient.QueryByPost(nodeCPUQuery, opts.CurrentTime)
	// if err != nil {
	//	return minUsageNode, nodeQuantie, err
	// }
	// var minUsage float64
	// for key := range nodeCpuUsageList.Data.Result {
	//	usage, ok := nodeCpuUsageList.Data.Result[key].Value[1].(string)
	//	if ok {
	//		nodeUsage, err := strconv.ParseFloat(usage, 64)
	//		if err != nil {
	//			continue
	//		}
	//		if nodeUsage < minUsage {
	//			minUsage = nodeUsage
	//			minUsageNode = nodeCpuUsageList.Data.Result[key].Metric["node"]
	//		}
	//	}
	// }

	// nodeQuantieResponse, err := clients.MonitorClient.QueryByPost(fmt.Sprintf(NodeUsageQuantile,
	//	"0.5", nodeCPUQuery), opts.CurrentTime)
	// if err != nil {
	//	return minUsageNode, nodeQuantie, err
	// }
	// if len(nodeQuantieResponse.Data.Result) != 0 {
	//	nodeQuantileCPU, ok := nodeQuantieResponse.Data.Result[0].Value[1].(string)
	//	if ok {
	//		quantile := &bcsdatamanager.NodeQuantile{
	//			Percentage:   "50",
	//			NodeCPUUsage: nodeQuantileCPU,
	//		}
	//		nodeQuantie = append(nodeQuantie, quantile)
	//	}
	// }
	return minUsageNode, nodeQuantie, nil
}

// GetClusterNodeCount get cluster node count
func (g *MetricGetter) GetClusterNodeCount(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) (int64, int64, error) {
	switch opts.ClusterType {
	case types.Kubernetes:
		return g.getK8sNodeCount(opts, clients)
	case types.Mesos:
		return g.getMesosNodeCount(ctx, opts)
	default:
		return 0, 0, fmt.Errorf("wrong clusterType:%s", opts.ClusterType)
	}
}

// GetNamespaceCPUMetrics get namespace cpu metrics
func (g *MetricGetter) GetNamespaceCPUMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.CPUMetrics, error) {
	switch opts.ClusterType {
	case types.Kubernetes:
		return g.getK8sNamespaceCPUMetrics(opts, clients)
	case types.Mesos:
		return g.getMesosNamespaceCPUMetrics(opts, clients)
	default:
		return nil, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetNamespaceMemoryMetrics get namespace memory metric
func (g *MetricGetter) GetNamespaceMemoryMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (*types.MemoryMetrics, error) {
	switch opts.ClusterType {
	case types.Kubernetes:
		return g.getK8sNamespaceMemoryMetrics(opts, clients)
	case types.Mesos:
		return g.getMesosNamespaceMemoryMetrics(opts, clients)
	default:
		return nil, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetNamespaceResourceLimit get namespace resource limit
func (g *MetricGetter) GetNamespaceResourceLimit(opts *types.JobCommonOpts,
	clients *types.Clients) (float64, int64, float64, int64, error) {
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
func (g *MetricGetter) GetNamespaceInstanceCount(opts *types.JobCommonOpts,
	clients *types.Clients) (int64, error) {
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
func (g *MetricGetter) GetWorkloadCPUMetrics(opts *types.JobCommonOpts, clients *types.Clients) (*types.CPUMetrics,
	error) {
	switch opts.ClusterType {
	case types.Kubernetes:
		return g.getK8SWorkloadCPU(opts, clients)
	case types.Mesos:
		return g.getMesosWorkloadCPU(opts, clients)
	default:
		return nil, fmt.Errorf("wrong clusterType :%s", opts.ClusterType)
	}
}

// GetWorkloadMemoryMetrics get workload memory metrics
func (g *MetricGetter) GetWorkloadMemoryMetrics(opt *types.JobCommonOpts, clients *types.Clients) (*types.MemoryMetrics,
	error) {
	switch opt.ClusterType {
	case types.Kubernetes:
		return g.getK8sWorkloadMemory(opt, clients)
	case types.Mesos:
		return g.getMesosWorkloadMemory(opt, clients)
	default:
		return nil, fmt.Errorf("wrong clusterType :%s", opt.ClusterType)
	}
}

// GetInstanceCount get instance count
func (g *MetricGetter) GetInstanceCount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error) {
	var count int64
	var query string
	podSumCondition := PodSumCondition
	if opts.ClusterType == types.Mesos {
		podSumCondition = MesosPodSumCondition
	}
	switch opts.ObjectType {
	case types.ClusterType:
		query = fmt.Sprintf(WorkloadInstance, fmt.Sprintf(ClusterCondition, opts.ClusterID), podSumCondition)
	case types.NamespaceType:
		query = fmt.Sprintf(WorkloadInstance,
			fmt.Sprintf(NamespaceCondition, opts.ClusterID, opts.Namespace), podSumCondition)
	case types.WorkloadType:
		podCondition := generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
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

// GetPodAutoscalerCount get pod autoscaler count
func (g *MetricGetter) GetPodAutoscalerCount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error) {
	var countQuery string
	switch opts.PodAutoscalerType {
	case types.HPAType:
		countQuery = fmt.Sprintf(HorizontalPodAutoscalerCount, opts.ClusterID, opts.PodAutoscalerName, opts.Namespace)
	case types.GPAType:
		countQuery = fmt.Sprintf(GeneralPodAutoscalerCount, opts.ClusterID, opts.PodAutoscalerName, opts.Namespace)
	}
	startMetric, err := clients.MonitorClient.QueryByPost(countQuery, opts.CurrentTime.Add(-30*time.Minute))
	if err != nil {
		return 0, fmt.Errorf("query former count error:%v", err)
	}
	minQuery := fmt.Sprintf(MinOverTime, countQuery, "30m")
	maxQuery := fmt.Sprintf(MaxOverTime, countQuery, "30m")
	minMetric, err := clients.MonitorClient.QueryByPost(minQuery, opts.CurrentTime)
	if err != nil {
		return 0, fmt.Errorf("query min count error:%v", err)
	}
	minCount := GetInt64Data(minMetric)
	currentMetric, err := clients.MonitorClient.QueryByPost(countQuery, opts.CurrentTime)
	if err != nil {
		return 0, fmt.Errorf("query current count error:%v", err)
	}
	// 区间无数据，直接返回0
	if (len(startMetric.Data.Result) == 0 && len(currentMetric.Data.Result) == 0) || len(minMetric.Data.Result) == 0 {
		return 0, nil
	}
	// 不连续区间，起点无数据时，取终点值；终点无数据时，用区间内最大值-起点值，得到变化值
	if len(startMetric.Data.Result) == 0 {
		return GetInt64Data(currentMetric), nil
	} else if len(currentMetric.Data.Result) == 0 {
		maxMetric, queryErr := clients.MonitorClient.QueryByPost(maxQuery, opts.CurrentTime)
		if err != nil {
			return 0, fmt.Errorf("query max count error:%v", queryErr)
		}
		return GetInt64Data(maxMetric) - GetInt64Data(startMetric), nil
	}

	// 最小值大于1，区间连续
	if minCount > 1 {
		return GetInt64Data(currentMetric) - GetInt64Data(startMetric), nil
	}
	// 最小值等于1，判断是否连续，如果不连续，metric被delete过或exporter重启过；如果连续，终点-1
	rangeMetrics, err := clients.MonitorClient.QueryRangeByPost(countQuery, opts.CurrentTime.Add(-30*time.Minute),
		opts.CurrentTime, 30*time.Second)
	if err != nil {
		return 0, fmt.Errorf(err.Error())
	}
	var rangeLength int
	for _, result := range rangeMetrics.Data.Result {
		rangeLength += len(result.Values)
	}
	if rangeLength != 61 {
		maxMetric, queryErr := clients.MonitorClient.QueryByPost(maxQuery, opts.CurrentTime)
		if err != nil {
			return 0, fmt.Errorf("query max count error:%v", queryErr)
		}
		if GetInt64Data(maxMetric) > GetInt64Data(startMetric) {
			return GetInt64Data(maxMetric) - GetInt64Data(startMetric), nil
		}
		return GetInt64Data(currentMetric), nil
	}
	return GetInt64Data(currentMetric) - 1, nil
}

// GetCACount get ca trigger times
func (g *MetricGetter) GetCACount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error) {
	var total int
	upQuery := fmt.Sprintf(ClusterAutoscalerUpCount, opts.ClusterID)
	upMetrics, err := clients.MonitorClient.QueryRangeByPost(upQuery, opts.CurrentTime.Add(-10*time.Minute),
		opts.CurrentTime, 30*time.Second)
	firstTimestamp := float64(opts.CurrentTime.Add(-10 * time.Minute).Unix())
	if err != nil {
		return 0, fmt.Errorf("query cluster autoscaler up count error:%v", err)
	}
	if len(upMetrics.Data.Result) > 0 {
		fillUpMetrics := fillMetrics(firstTimestamp, upMetrics.Data.Result[0].Values, 30)
		total += getIncreasingIntervalDifference(fillUpMetrics)
	}

	downQuery := fmt.Sprintf(ClusterAutoscalerDownCount, opts.ClusterID)
	downMetrics, err := clients.MonitorClient.QueryRangeByPost(downQuery, opts.CurrentTime.Add(-10*time.Minute),
		opts.CurrentTime, 30*time.Second)
	if err != nil {
		return 0, fmt.Errorf("query cluster autoscaler down count error:%v", err)
	}
	if len(downMetrics.Data.Result) > 0 {
		fillDownMetrics := fillMetrics(firstTimestamp, downMetrics.Data.Result[0].Values, 30)
		total += getIncreasingIntervalDifference(fillDownMetrics)
	}
	return int64(total), nil
}
