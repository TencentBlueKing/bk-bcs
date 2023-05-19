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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

// NodeInfo 节点信息
type NodeInfo struct {
	CPUCount                string `json:"cpu_count"`                 // CPU
	Memory                  string `json:"memory"`                    // 内存, 单位 Byte
	Disk                    string `json:"disk"`                      // 存储, 单位 Byte
	Provider                string `json:"provider"`                  // IP来源, BKMonitor / Prometheus
	Release                 string `json:"release"`                   // 内核, 3.10.107-1-tlinux2_kvm_guest-0052
	DockerVersion           string `json:"dockerVersion"`             // Docker, 18.6.3-ce-tke.1
	Sysname                 string `json:"sysname"`                   // 操作系统, linux
	IP                      string `json:"ip"`                        // ip，多个使用 , 分隔
	ContainerRuntimeVersion string `json:"container_runtime_version"` // 容器运行时版本
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
		{Name: "ip", Value: n.IP},
		{Name: "container_runtime_version", Value: n.ContainerRuntimeVersion},
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
	GetClusterCPURequest(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterCPUUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterCPURequestUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryRequest(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterMemoryRequestUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskioUsage(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskioUsed(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetClusterDiskioTotal(ctx context.Context, projectId, clusterId string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeInfo(ctx context.Context, projectId, clusterId, nodeName string, t time.Time) (*NodeInfo, error)
	GetNodeCPUUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeCPURequestUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeMemoryUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeMemoryRequestUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeDiskUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeDiskioUsage(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeNetworkTransmit(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeNetworkReceive(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodePodCount(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
		step time.Duration) ([]*prompb.TimeSeries, error)
	GetNodeContainerCount(ctx context.Context, projectId, clusterId, nodeName string, start, end time.Time,
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
	GetContainerGPUMemoryUsage(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerGPUUsed(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerGPUUsage(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerDiskReadTotal(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
	GetContainerDiskWriteTotal(ctx context.Context, projectId, clusterId, namespace, podname string,
		containerNameList []string, start, end time.Time, step time.Duration) ([]*prompb.TimeSeries, error)
}

// GetNodeMatch 按集群node节点正则匹配
func GetNodeMatch(ctx context.Context, clusterId string) (string, string, error) {
	nodeList, nodeNameList, err := k8sclient.GetNodeList(ctx, clusterId, true)
	if err != nil {
		return "", "", err
	}
	return utils.StringJoinWithRegex(nodeList, "|", ".*"), utils.StringJoinWithRegex(nodeNameList, "|", "$"), nil
}

// GetNodeMatchByName 按集群node节点正则匹配
func GetNodeMatchByName(ctx context.Context, clusterId, nodeName string) (string, string, error) {
	nodeIPList, err := k8sclient.GetNodeByName(ctx, clusterId, nodeName)
	if err != nil {
		return "", "", err
	}
	return utils.StringJoinIPWithRegex(nodeIPList, "|", ".*"), strings.Join(nodeIPList, ","), nil
}

// GetNodeContainerRuntimeVersionByName 通过节点名称获取容器运行时版本
func GetNodeContainerRuntimeVersionByName(ctx context.Context, clusterId, nodeName string) (string, error) {
	version, err := k8sclient.GetNodeContainerRuntimeVersionByName(ctx, clusterId, nodeName)
	if err != nil {
		return "", err
	}
	return version, nil
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

// GetNodeMatchWithScale 处理集群的节点列表，按照给定的粒度划分
func GetNodeMatchWithScale(ctx context.Context, clusterId string, scale int) ([]*ResultTuple, error) {
	nodeList, nodeNameList, err := k8sclient.GetNodeList(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}
	resslice := chunkSlice(nodeList, nodeNameList, scale)
	return resslice, nil
}

func chunkSlice(nodeList []string, nodeNameList []string, chunkSize int) []*ResultTuple {
	var res []*ResultTuple
	res = make([]*ResultTuple, 0)
	for i := 0; i < len(nodeList); i += chunkSize {
		end := i + chunkSize
		if end > len(nodeList) {
			end = len(nodeList)
		}
		res = append(res, &ResultTuple{
			utils.StringJoinWithRegex(nodeList[i:end], "|", ":9101$"),
			utils.StringJoinWithRegex(nodeNameList[i:end], "|", "$"),
			nil,
		})
	}

	return res
}

type ResultTuple struct {
	NodeMatch     string
	NodeNameMatch string
	Err           error
}

// MatrixsToSeries prom返回转换为时序对象
func MatrixsToSeries(matrixs []model.Matrix) []*prompb.TimeSeries {
	series := make([]*prompb.TimeSeries, 0)
	for _, matrix := range matrixs {
		for _, m := range matrix {
			series = append(series, sampleStreamToSeries(m))
		}
	}
	return series
}

// MergeSameSeries merge same metrics series
func MergeSameSeries(series []*prompb.TimeSeries) []*prompb.TimeSeries {
	if len(series) == 0 {
		return nil
	}
	result := &prompb.TimeSeries{
		Labels:  make([]prompb.Label, 0),
		Samples: make([]prompb.Sample, 0),
	}
	for _, s := range series {
		result.Labels = s.Labels
		result.Samples = MergeSameSamples(result.Samples, s.Samples)
	}
	return []*prompb.TimeSeries{result}
}

// MergeSameSamples merge same samples
func MergeSameSamples(samples1, samples2 []prompb.Sample) []prompb.Sample {
	if len(samples1) == 0 {
		return samples2
	}
	for i := range samples1 {
		for j := range samples2 {
			if samples1[i].Timestamp == samples2[j].Timestamp {
				samples1[i].Value += samples2[j].Value
				break
			}
		}
	}
	return samples1
}

// DivideSeries divide same metrics series, series1 divide series2, series must only have one element
func DivideSeries(series1, series2 []*prompb.TimeSeries) []*prompb.TimeSeries {
	if len(series1) == 0 || len(series2) == 0 {
		return nil
	}
	result := &prompb.TimeSeries{
		Labels:  series1[0].Labels,
		Samples: make([]prompb.Sample, 0),
	}
	result.Samples = DivideSamples(series1[0].Samples, series2[0].Samples)
	return []*prompb.TimeSeries{result}
}

// DivideSamples samples1 divide samples2
func DivideSamples(samples1, samples2 []prompb.Sample) []prompb.Sample {
	if len(samples1) == 0 || len(samples2) == 0 {
		return nil
	}
	for i := range samples1 {
		for j := range samples2 {
			if samples1[i].Timestamp == samples2[j].Timestamp {
				if samples2[j].Value == 0 {
					samples1[i].Value = 0
				} else {
					samples1[i].Value = samples1[i].Value / samples2[j].Value * 100
				}
				break
			}
		}
	}
	return samples1
}
