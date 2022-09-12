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

// Package base xxx
package base

import (
	"context"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

// NodeInfo 节点信息
type NodeInfo struct {
	CPUCount      string `json:"cpu_count"`     // CPU
	Memory        string `json:"memory"`        // 内存, 单位 Byte
	Disk          string `json:"disk"`          // 存储, 单位 Byte
	Provider      string `json:"provider"`      // IP来源, BKMonitor / Prometheus
	Release       string `json:"release"`       // 内核, 3.10.107-1-tlinux2_kvm_guest-0052
	DockerVersion string `json:"dockerVersion"` // Docker, 18.6.3-ce-tke.1
	Sysname       string `json:"sysname"`       // 操作系统, linux
}

// PromSeries 给 series
func (n *NodeInfo) PromSeries(t time.Time) []*prompb.TimeSeries {
	labelSet := []prompb.Label{
		{Name: "cpu_count", Value: n.CPUCount},
		{Name: "memory", Value: n.Memory},
		{Name: "disk", Value: n.Disk},
		{Name: "provider", Value: n.Provider},
		{Name: "release", Value: n.Release},
		{Name: "dockerVersion", Value: n.DockerVersion},
		{Name: "sysname", Value: n.Sysname},
	}

	sample := []prompb.Sample{
		{Value: float64(1), Timestamp: t.UnixMilli()},
	}
	series := []*prompb.TimeSeries{
		{Labels: labelSet, Samples: sample},
	}
	return series
}

// MetricHandler xxx
type MetricHandler interface {
	GetClusterCPUTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterCPUUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterCPUUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeInfo(ctx context.Context, projectId, clusterId, ip string, t time.Time) (*NodeInfo, error)
	GetNodeCPUUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeMemoryUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeDiskUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeDiskioUsage(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeNetworkTransmit(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeNetworkReceive(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodePodCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeContainerCount(ctx context.Context, projectId, clusterId, ip string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetPodCPUUsage(ctx context.Context, projectId, clusterId, namespace string, podNameList []string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetPodMemoryUsed(ctx context.Context, projectId, clusterId, namespace string, podNameList []string, start,
		end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetPodNetworkReceive(ctx context.Context, projectId, clusterId, namespace string, podNameList []string, start,
		end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetPodNetworkTransmit(ctx context.Context, projectId, clusterId, namespace string, podNameList []string, start,
		end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerCPUUsage(ctx context.Context, projectId, clusterId, namespace, podname string, containerNameList []string,
		start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerMemoryUsed(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerCPULimit(ctx context.Context, projectId, clusterId, namespace, podname string, containerNameList []string,
		start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerMemoryLimit(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerDiskReadTotal(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerDiskWriteTotal(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
}

// GetNodeMatch 按集群node节点正则匹配
func GetNodeMatch(ctx context.Context, clusterId string, withRegex bool) (string, error) {
	nodeList, err := k8sclient.GetNodeList(ctx, clusterId, true)
	if err != nil {
		return "", err
	}

	instanceList := make([]string, 0, len(nodeList))
	for _, node := range nodeList {
		if withRegex {
			instanceList = append(instanceList, node+`:.*`)
		} else {
			instanceList = append(instanceList, node)
		}
	}
	return strings.Join(instanceList, "|"), nil
}

func sampleStreamToSeries(m *model.SampleStream) *prompb.TimeSeries {
	series := &prompb.TimeSeries{}
	for k, v := range m.Metric {
		series.Labels = append(series.Labels, prompb.Label{
			Name:  string(k),
			Value: string(v),
		})
	}
	for _, v := range m.Values {
		series.Samples = append(series.Samples, prompb.Sample{
			Timestamp: v.Timestamp.Time().UnixMilli(),
			Value:     float64(v.Value),
		})
	}
	return series
}

// MatrixToSeries prom返回转换为时序对象
func MatrixToSeries(matrix model.Matrix) []*prompb.TimeSeries {
	series := make([]*prompb.TimeSeries, 0, len(matrix))
	for _, m := range matrix {
		series = append(series, sampleStreamToSeries(m))
	}
	return series
}

// GetFirstValue 获取第一个值
func GetFirstValue(vector model.Vector) string {
	if len(vector) == 0 {
		return "0"
	}
	return vector[0].Value.String()
}

// GetLabelSet 获取第一个值的labels
func GetLabelSet(vector model.Vector) map[string]string {
	labelSet := map[string]string{}
	if len(vector) == 0 {
		return labelSet
	}
	for k, v := range vector[0].Metric {
		labelSet[string(k)] = string(v)
	}
	return labelSet
}
