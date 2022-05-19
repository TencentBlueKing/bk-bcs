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

package datajob

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ClusterDayPolicy cluster day policy
type ClusterDayPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// ClusterHourPolicy cluster hour policy
type ClusterHourPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// ClusterMinutePolicy cluster minute policy
type ClusterMinutePolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NewClusterDayPolicy init day policy
func NewClusterDayPolicy(getter metric.Server, store store.Server) *ClusterDayPolicy {
	return &ClusterDayPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewClusterHourPolicy init hour policy
func NewClusterHourPolicy(getter metric.Server, store store.Server) *ClusterHourPolicy {
	return &ClusterHourPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewClusterMinutePolicy init minute policy
func NewClusterMinutePolicy(getter metric.Server, store store.Server) *ClusterMinutePolicy {
	return &ClusterMinutePolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// ImplementPolicy day policy implement, calculate every day
func (p *ClusterDayPolicy) ImplementPolicy(ctx context.Context, opts *common.JobCommonOpts, clients *common.Clients) {
	totalCPU, CPURequest, CPUUsed, cpuUsage, err := p.MetricGetter.GetClusterCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster day policy error, opts: %v, err: %v", opts, err)
	}
	totalMemory, memoryRequest, memoryUsed, memoryUsage, err := p.MetricGetter.
		GetClusterMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster day policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do cluster day policy error, opts: %v, err: %v", opts, err)
	}
	minUsageNode, nodeQuantile, err := p.MetricGetter.GetClusterNodeMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster day policy error, opts: %v, err: %v", opts, err)
	}
	node, availableNode, err := p.MetricGetter.GetClusterNodeCount(opts, clients)
	if err != nil {
		blog.Errorf("do cluster day policy error, opts: %v, err: %v", opts, err)
	}
	var avgLoadCPU float64
	var avgLoadMemory int64
	if availableNode != 0 {
		avgLoadCPU = CPUUsed / float64(availableNode)
		avgLoadMemory = memoryUsed / availableNode
	}
	hourOpts := &common.JobCommonOpts{
		ObjectType: common.ClusterType,
		ProjectID:  opts.ProjectID,
		ClusterID:  opts.ClusterID,
		Dimension:  common.DimensionHour,
	}
	bucket, _ := common.GetBucketTime(opts.CurrentTime.AddDate(0, 0, -1), common.DimensionHour)
	hourMetrics, err := p.store.GetRawClusterInfo(ctx, hourOpts, bucket)
	if err != nil {
		blog.Errorf("do cluster day policy failed, get cluster metrics err:%v", err)
		return
	} else if len(hourMetrics) != 1 {
		blog.Errorf("do cluster day policy failed, get cluster metrics err, length not equal 1, metrics:%v", hourMetrics)
		return
	}
	hourMetric := hourMetrics[0]
	minuteBucket, _ := common.GetBucketTime(opts.CurrentTime.Add(-10*time.Minute), common.DimensionMinute)
	workloadCount, err := p.store.GetWorkloadCount(ctx, opts, minuteBucket, opts.CurrentTime.Add(-10*time.Minute))
	if err != nil {
		blog.Errorf("do cluster day policy failed, get cluster workload count err: %v", err)
	}
	clusterMetric := &common.ClusterMetrics{
		Index:              common.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(common.FormatTime(opts.CurrentTime, opts.Dimension)),
		TotalLoadCPU:       CPUUsed,
		CPUUsage:           cpuUsage,
		TotalLoadMemory:    memoryUsed,
		MemoryRequest:      memoryRequest,
		MemoryUsage:        memoryUsage,
		InstanceCount:      instanceCount,
		CpuRequest:         CPURequest,
		AvgLoadMemory:      avgLoadMemory,
		AvgLoadCPU:         avgLoadCPU,
		NodeCount:          node,
		AvailableNodeCount: availableNode,
		WorkloadCount:      workloadCount,
		MinNode:            hourMetric.MinNode,
		MaxNode:            hourMetric.MaxNode,
		MinInstance:        hourMetric.MinInstance,
		MaxInstance:        hourMetric.MaxInstance,
		MinUsageNode:       minUsageNode,
		NodeQuantile:       nodeQuantile,
		TotalCPU:           totalCPU,
		TotalMemory:        totalMemory,
	}
	err = p.store.InsertClusterInfo(ctx, clusterMetric, opts)
	if err != nil {
		blog.Errorf("do cluster day policy error, opts: %v, err: %v", opts, err)
	}
}

// ImplementPolicy hour policy implement, calculate every hour
func (p *ClusterHourPolicy) ImplementPolicy(ctx context.Context, opts *common.JobCommonOpts, clients *common.Clients) {
	totalCPU, CPURequest, CPUUsed, cpuUsage, err := p.MetricGetter.GetClusterCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster hour policy error, opts: %v, err: %v", opts, err)
	}
	totalMemory, memoryRequest, memoryUsed, memoryUsage, err := p.MetricGetter.
		GetClusterMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster hour policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do cluster hour policy error, opts: %v, err: %v", opts, err)
	}
	minUsageNode, nodeQuantile, err := p.MetricGetter.GetClusterNodeMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster hour policy error, opts: %v, err: %v", opts, err)
	}
	node, availableNode, err := p.MetricGetter.GetClusterNodeCount(opts, clients)
	if err != nil {
		blog.Errorf("do cluster hour policy error, opts: %v, err: %v", opts, err)
	}
	var avgLoadCPU float64
	var avgLoadMemory int64
	if availableNode != 0 {
		avgLoadCPU = CPUUsed / float64(availableNode)
		avgLoadMemory = memoryUsed / availableNode
	}
	minuteOpts := &common.JobCommonOpts{
		ObjectType: common.ClusterType,
		ProjectID:  opts.ProjectID,
		ClusterID:  opts.ClusterID,
		Dimension:  common.DimensionMinute,
	}
	bucket, _ := common.GetBucketTime(opts.CurrentTime.Add((-1)*time.Hour), common.DimensionMinute)
	minuteMetrics, err := p.store.GetRawClusterInfo(ctx, minuteOpts, bucket)
	if err != nil {
		blog.Errorf("do cluster hour policy failed, get cluster metrics err:%v", err)
		return
	} else if len(minuteMetrics) != 1 {
		blog.Errorf("do cluster hour policy failed, get cluster metrics err, length not equal 1, metrics:%v", minuteMetrics)
		return
	}
	minuteMetric := minuteMetrics[0]
	minuteBucket, _ := common.GetBucketTime(opts.CurrentTime.Add(-10*time.Minute), common.DimensionMinute)
	workloadCount, err := p.store.GetWorkloadCount(ctx, opts, minuteBucket, opts.CurrentTime.Add(-10*time.Minute))
	if err != nil {
		blog.Errorf("do cluster hour policy failed, get cluster workload count err: %v", err)
	}
	clusterMetric := &common.ClusterMetrics{
		Index:              common.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(common.FormatTime(opts.CurrentTime, opts.Dimension)),
		TotalLoadCPU:       CPUUsed,
		CPUUsage:           cpuUsage,
		TotalLoadMemory:    memoryUsed,
		MemoryRequest:      memoryRequest,
		MemoryUsage:        memoryUsage,
		InstanceCount:      instanceCount,
		CpuRequest:         CPURequest,
		AvgLoadMemory:      avgLoadMemory,
		AvgLoadCPU:         avgLoadCPU,
		NodeCount:          node,
		AvailableNodeCount: availableNode,
		WorkloadCount:      workloadCount,
		MinNode:            minuteMetric.MinNode,
		MaxNode:            minuteMetric.MaxNode,
		MinInstance:        minuteMetric.MinInstance,
		MaxInstance:        minuteMetric.MaxInstance,
		MinUsageNode:       minUsageNode,
		NodeQuantile:       nodeQuantile,
		TotalCPU:           totalCPU,
		TotalMemory:        totalMemory,
	}
	err = p.store.InsertClusterInfo(ctx, clusterMetric, opts)
	if err != nil {
		blog.Errorf("do cluster hour policy error, opts: %v, err: %v", opts, err)
	}
}

// ImplementPolicy minute policy implement, calculate every 10 min
func (p *ClusterMinutePolicy) ImplementPolicy(ctx context.Context, opts *common.JobCommonOpts,
	clients *common.Clients) {
	totalCPU, CPURequest, CPUUsed, cpuUsage, err := p.MetricGetter.GetClusterCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster minute policy error, opts: %v, err: %v", opts, err)
	}
	totalMemory, memoryRequest, memoryUsed, memoryUsage, err := p.MetricGetter.GetClusterMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster minute policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do cluster minute policy error, opts: %v, err: %v", opts, err)
	}
	minUsageNode, nodeQuantile, err := p.MetricGetter.GetClusterNodeMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do cluster minute policy error, opts: %v, err: %v", opts, err)
	}
	node, availableNode, err := p.MetricGetter.GetClusterNodeCount(opts, clients)
	if err != nil {
		blog.Errorf("do cluster minute policy error, opts: %v, err: %v", opts, err)
	}
	var avgLoadCPU float64
	var avgLoadMemory int64
	if availableNode != 0 {
		avgLoadCPU = CPUUsed / float64(availableNode)
		avgLoadMemory = memoryUsed / availableNode
	}
	minuteBucket, _ := common.GetBucketTime(opts.CurrentTime.Add(-10*time.Minute), common.DimensionMinute)
	workloadCount, err := p.store.GetWorkloadCount(ctx, opts, minuteBucket, opts.CurrentTime.Add(-10*time.Minute))
	if err != nil {
		blog.Errorf("do cluster minute policy failed, get cluster workload count err: %v", err)
	}
	clusterMetric := &common.ClusterMetrics{
		Index:              common.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(common.FormatTime(opts.CurrentTime, opts.Dimension)),
		TotalLoadCPU:       CPUUsed,
		CPUUsage:           cpuUsage,
		TotalLoadMemory:    memoryUsed,
		MemoryRequest:      memoryRequest,
		MemoryUsage:        memoryUsage,
		InstanceCount:      instanceCount,
		CpuRequest:         CPURequest,
		AvgLoadMemory:      avgLoadMemory,
		AvgLoadCPU:         avgLoadCPU,
		NodeCount:          node,
		AvailableNodeCount: availableNode,
		WorkloadCount:      workloadCount,
		MinNode: &bcsdatamanager.ExtremumRecord{
			Name:       "MinNode",
			MetricName: "MinNode",
			Value:      float64(node),
			Period:     common.FormatTime(opts.CurrentTime, opts.Dimension).String(),
		},
		MaxNode: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxNode",
			MetricName: "MaxNode",
			Value:      float64(node),
			Period:     common.FormatTime(opts.CurrentTime, opts.Dimension).String(),
		},
		MinInstance: &bcsdatamanager.ExtremumRecord{
			Name:       "MinInstance",
			MetricName: "MinInstance",
			Value:      float64(instanceCount),
			Period:     common.FormatTime(opts.CurrentTime, opts.Dimension).String(),
		},
		MaxInstance: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxInstance",
			MetricName: "MaxInstance",
			Value:      float64(instanceCount),
			Period:     common.FormatTime(opts.CurrentTime, opts.Dimension).String(),
		},
		MinUsageNode: minUsageNode,
		NodeQuantile: nodeQuantile,
		TotalCPU:     totalCPU,
		TotalMemory:  totalMemory,
	}
	if err = p.store.InsertClusterInfo(ctx, clusterMetric, opts); err != nil {
		blog.Errorf("do cluster minute policy error, opts: %v, err: %v", opts, err)
	}
}
