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

package bcssystem

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/prompb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
)

// metricsParams
type metricsParams struct {
	// metrics handler client
	client base.MetricHandler
	ctx    context.Context
	// metrics handler params
	projectID string
	clusterID string
	namespace string
	// group params
	group string
	// pod params
	podName        string
	podNames       []string
	containerNames []string
	node           string
	// query params
	startTime    time.Time
	endTime      time.Time
	stepDuration time.Duration
}

// metricsFn
type metricsFn func(metricsParams) ([]*prompb.TimeSeries, error)

// metrics maps
// map metrics to metrics func
var metricsMaps = map[string]metricsFn{
	// GetClusterCPUTotal
	"bcs:cluster:cpu:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPUTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterCPUUsed
	"bcs:cluster:cpu:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPUUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterPodUsage
	"bcs:cluster:pod:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterPodUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterPodTotal
	"bcs:cluster:pod:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterPodTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterPodUsed
	"bcs:cluster:pod:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterPodUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterCPURequest
	"bcs:cluster:cpu:request": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPURequest(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterCPUUsage
	"bcs:cluster:cpu:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPUUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterCPURequestUsage
	"bcs:cluster:cpu_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPURequestUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterMemoryTotal
	"bcs:cluster:memory:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterMemoryUsed
	"bcs:cluster:memory:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterMemoryRequest
	"bcs:cluster:memory:request": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryRequest(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterMemoryUsage
	"bcs:cluster:memory:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterMemoryRequestUsage
	"bcs:cluster:memory_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryRequestUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterDiskTotal
	"bcs:cluster:disk:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterDiskUsed
	"bcs:cluster:disk:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterDiskUsage
	"bcs:cluster:disk:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterDiskioUsage
	"bcs:cluster:diskio:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskioUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterDiskioUsed
	"bcs:cluster:diskio:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskioUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterDiskioTotal
	"bcs:cluster:diskio:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskioTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterGroupNodeNum
	"bcs:cluster:group:node_num": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterGroupNodeNum(mp.ctx, mp.projectID, mp.clusterID, mp.group,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetClusterGroupMaxNodeNum
	"bcs:cluster:group:max_node_num": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterGroupMaxNodeNum(mp.ctx, mp.projectID, mp.clusterID, mp.group,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// node metrics
	"bcs:node:info": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		// get nodeinfo from k8s api
		nodeInfo, err := mp.client.GetNodeInfo(mp.ctx, mp.projectID, mp.clusterID, mp.node, mp.endTime)
		return nodeInfo.PromSeries(mp.endTime), err
	},
	// GetNodeCPUUsage
	"bcs:node:cpu:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeCPUUsage(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeCPUTotal
	"bcs:node:cpu:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeCPUTotal(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeCPURequest
	"bcs:node:cpu:request": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeCPURequest(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeCPUUsed
	"bcs:node:cpu:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeCPUUsed(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeCPURequestUsage
	"bcs:node:cpu_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeCPURequestUsage(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeMemoryTotal
	"bcs:node:memory:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeMemoryTotal(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeMemoryRequest
	"bcs:node:memory:request": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeMemoryRequest(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeMemoryUsed
	"bcs:node:memory:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeMemoryUsed(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeMemoryUsage
	"bcs:node:memory:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeMemoryUsage(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// GetNodeMemoryRequestUsage
	"bcs:node:memory_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeMemoryRequestUsage(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:disk:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeDiskUsage(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:disk:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeDiskUsed(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:disk:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeDiskTotal(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:diskio:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeDiskioUsage(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:pod_count": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodePodCount(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:pod_total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodePodTotal(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:container_count": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeContainerCount(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:network_transmit": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeNetworkTransmit(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:network_receive": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeNetworkReceive(mp.ctx, mp.projectID, mp.clusterID, mp.node,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	// pod metrics
	"bcs:pod:cpu_usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetPodCPUUsage(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:pod:cpu_limit_usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetPodCPULimitUsage(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:pod:cpu_request_usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetPodCPURequestUsage(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:pod:memory_used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetPodMemoryUsed(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:pod:network_receive": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetPodNetworkReceive(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:pod:network_transmit": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetPodNetworkTransmit(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	// container metrics
	"bcs:container:cpu_usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerCPUUsage(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:memory_used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerMemoryUsed(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:cpu_limit": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerCPULimit(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:memory_limit": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerMemoryLimit(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:gpu_memory_usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerGPUMemoryUsage(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:gpu_used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerGPUUsed(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:gpu_usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerGPUUsage(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:disk_read_total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerDiskReadTotal(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:container:disk_write_total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetContainerDiskWriteTotal(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
			mp.podName, mp.containerNames, mp.startTime, mp.endTime, mp.stepDuration)
	},
}
