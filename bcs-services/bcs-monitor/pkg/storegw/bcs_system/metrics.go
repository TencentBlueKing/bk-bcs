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
 *
 */

// Package bcs_system xxx
package bcs_system

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system/source/base"
	"github.com/prometheus/prometheus/prompb"
)

type metricsParams struct {
	client         base.MetricHandler
	ctx            context.Context
	projectID      string
	clusterID      string
	namespace      string
	podName        string
	podNames       []string
	containerNames []string
	ip             string
	startTime      time.Time
	endTime        time.Time
	stepDuration   time.Duration
}

type metricsFn func(metricsParams) ([]*prompb.TimeSeries, error)

var metricsMaps map[string]metricsFn = map[string]metricsFn{
	"bcs:cluster:cpu:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPUTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:cpu:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPUUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:cpu:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPUUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:cpu_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterCPURequestUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:memory:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:memory:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:memory:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:memory_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterMemoryRequestUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:disk:total": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskTotal(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:disk:used": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskUsed(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:disk:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:cluster:diskio:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetClusterDiskioUsage(mp.ctx, mp.projectID, mp.clusterID,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:info": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		nodeInfo, err := mp.client.GetNodeInfo(mp.ctx, mp.projectID, mp.clusterID, mp.ip, mp.endTime)
		return nodeInfo.PromSeries(mp.endTime), err
	},
	"bcs:node:cpu:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeCPUUsage(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:cpu_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeCPURequestUsage(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:memory:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeMemoryUsage(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:memory_request:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeMemoryRequestUsage(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:disk:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeDiskUsage(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:diskio:usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeDiskioUsage(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:pod_count": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodePodCount(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:container_count": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeContainerCount(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:network_transmit": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeNetworkTransmit(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:node:network_receive": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetNodeNetworkReceive(mp.ctx, mp.projectID, mp.clusterID, mp.ip,
			mp.startTime, mp.endTime, mp.stepDuration)
	},
	"bcs:pod:cpu_usage": func(mp metricsParams) ([]*prompb.TimeSeries, error) {
		return mp.client.GetPodCPUUsage(mp.ctx, mp.projectID, mp.clusterID, mp.namespace,
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
