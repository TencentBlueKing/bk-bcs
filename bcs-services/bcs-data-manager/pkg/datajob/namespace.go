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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// NamespaceDayPolicy day policy
type NamespaceDayPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NamespaceHourPolicy hour policy
type NamespaceHourPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NamespaceMinutePolicy minute policy
type NamespaceMinutePolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NewNamespaceDayPolicy init day policy
func NewNamespaceDayPolicy(getter metric.Server, store store.Server) *NamespaceDayPolicy {
	return &NamespaceDayPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewNamespaceHourPolicy init hour policy
func NewNamespaceHourPolicy(getter metric.Server, store store.Server) *NamespaceHourPolicy {
	return &NamespaceHourPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewNamespaceMinutePolicy init minute policy
func NewNamespaceMinutePolicy(getter metric.Server, store store.Server) *NamespaceMinutePolicy {
	return &NamespaceMinutePolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// ImplementPolicy day policy implement
func (p *NamespaceDayPolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts, clients *types.Clients) {
	cpuMetrics, err := p.MetricGetter.GetNamespaceCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do namespace day policy error, opts: %v, err: %v", opts, err)
	}
	memoryMetrics, err := p.MetricGetter.GetNamespaceMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do namespace day policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do namespace day policy error, opts: %v, err: %v", opts, err)
	}

	hourOpts := &types.JobCommonOpts{
		ObjectType: types.NamespaceType,
		ProjectID:  opts.ProjectID,
		ClusterID:  opts.ClusterID,
		Namespace:  opts.Namespace,
		Dimension:  types.DimensionHour,
	}
	bucket, _ := utils.GetBucketTime(opts.CurrentTime.AddDate(0, 0, -1), types.DimensionHour)
	hourMetrics, err := p.store.GetRawNamespaceInfo(ctx, hourOpts, bucket)
	if err != nil {
		blog.Errorf("do namespace day policy failed,  get namespace metrics err:%v", err)
		return
	} else if len(hourMetrics) != 1 {
		blog.Errorf("do namespace day policy failed, get namespace metrics err, length not equal 1, metrics:%v", hourMetrics)
		return
	}
	hourMetric := hourMetrics[0]
	// 当天出现过的workload，数量都统计上来
	hourBucket, _ := utils.GetBucketTime(opts.CurrentTime.AddDate(0, 0, -1), types.DimensionHour)
	workloadCount, err := p.store.GetWorkloadCount(ctx, opts, hourBucket, time.Time{})
	if err != nil {
		blog.Errorf("do namespace day policy failed, get namespace workload count err: %v", err)
	}
	namespaceMetric := &types.NamespaceMetrics{
		Index:              utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(utils.FormatTime(opts.CurrentTime, opts.Dimension)),
		CPUUsage:           cpuMetrics.CPUUsage,
		MemoryRequest:      memoryMetrics.MemoryRequest,
		MemoryUsage:        memoryMetrics.MemoryUsage,
		InstanceCount:      instanceCount,
		CPURequest:         cpuMetrics.CPURequest,
		MemoryLimit:        memoryMetrics.MemoryLimit,
		CPULimit:           cpuMetrics.CPULimit,
		WorkloadCount:      workloadCount,
		CPUUsageAmount:     cpuMetrics.CPUUsed,
		MemoryUsageAmount:  memoryMetrics.MemoryUsed,
		MaxCPUUsageTime:    hourMetric.MaxCPUUsageTime,
		MinCPUUsageTime:    hourMetric.MinCPUUsageTime,
		MaxMemoryUsageTime: hourMetric.MaxMemoryUsageTime,
		MinMemoryUsageTime: hourMetric.MinMemoryUsageTime,
		MinWorkloadUsage:   hourMetric.MinWorkloadUsage,
		MaxWorkloadUsage:   hourMetric.MaxWorkloadUsage,
	}
	err = p.store.InsertNamespaceInfo(ctx, namespaceMetric, opts)
	if err != nil {
		blog.Errorf("do namespace day policy error, opts: %v, err: %v", opts, err)
	}
}

// ImplementPolicy hour policy implement
func (p *NamespaceHourPolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) {
	cpuMetrics, err := p.MetricGetter.GetNamespaceCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do namespace hour policy error, opts: %v, err: %v", opts, err)
	}
	memoryMetrics, err := p.MetricGetter.GetNamespaceMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do namespace hour policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do namespace hour policy error, opts: %v, err: %v", opts, err)
	}

	minuteOpts := &types.JobCommonOpts{
		ObjectType: types.NamespaceType,
		ProjectID:  opts.ProjectID,
		ClusterID:  opts.ClusterID,
		Namespace:  opts.Namespace,
		Dimension:  types.DimensionMinute,
	}
	bucket, _ := utils.GetBucketTime(opts.CurrentTime.Add((-1)*time.Hour), types.DimensionMinute)
	minuteMetrics, err := p.store.GetRawNamespaceInfo(ctx, minuteOpts, bucket)
	if err != nil {
		blog.Errorf("do namespace hour policy failed, get namespace metrics err:%v", err)
		return
	} else if len(minuteMetrics) != 1 {
		blog.Errorf("do namespace hour policy failed, get namespace metrics err, length not equal 1, "+
			"metrics:%v", minuteMetrics)
		return
	}
	minuteMetric := minuteMetrics[0]
	// 统计上一个小时出现的workload数量
	hourBucket, _ := utils.GetBucketTime(opts.CurrentTime.Add(-1*time.Hour), types.DimensionHour)
	workloadCount, err := p.store.GetWorkloadCount(ctx, opts, hourBucket, opts.CurrentTime.Add(-10*time.Minute))
	if err != nil {
		blog.Errorf("do namespace hour policy failed, get namespace workload count err: %v", err)
	}
	namespaceMetric := &types.NamespaceMetrics{
		Index:              utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(utils.FormatTime(opts.CurrentTime, opts.Dimension)),
		CPUUsage:           cpuMetrics.CPUUsage,
		MemoryRequest:      memoryMetrics.MemoryRequest,
		MemoryUsage:        memoryMetrics.MemoryUsage,
		InstanceCount:      instanceCount,
		CPURequest:         cpuMetrics.CPURequest,
		MemoryLimit:        memoryMetrics.MemoryLimit,
		CPULimit:           cpuMetrics.CPULimit,
		WorkloadCount:      workloadCount,
		CPUUsageAmount:     cpuMetrics.CPUUsed,
		MemoryUsageAmount:  memoryMetrics.MemoryUsed,
		MaxCPUUsageTime:    minuteMetric.MaxCPUUsageTime,
		MinCPUUsageTime:    minuteMetric.MinCPUUsageTime,
		MaxMemoryUsageTime: minuteMetric.MaxMemoryUsageTime,
		MinMemoryUsageTime: minuteMetric.MinMemoryUsageTime,
		MinWorkloadUsage:   minuteMetric.MinWorkloadUsage,
		MaxWorkloadUsage:   minuteMetric.MaxWorkloadUsage,
	}
	err = p.store.InsertNamespaceInfo(ctx, namespaceMetric, opts)
	if err != nil {
		blog.Errorf("do namespace hour policy error, opts: %v, err: %v", opts, err)
	}
}

// ImplementPolicy minute policy implement
func (p *NamespaceMinutePolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) {
	cpuMetrics, err := p.MetricGetter.GetNamespaceCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do namespace minute policy error, opts: %v, err: %v", opts, err)
	}
	memoryMetrics, err := p.MetricGetter.GetNamespaceMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do namespace minute policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do namespace minute policy error, opts: %v, err: %v", opts, err)
	}
	minuteBucket, _ := utils.GetBucketTime(opts.CurrentTime.Add(-10*time.Minute), types.DimensionMinute)
	workloadCount, err := p.store.GetWorkloadCount(ctx, opts, minuteBucket, opts.CurrentTime.Add(-10*time.Minute))
	if err != nil {
		blog.Errorf("do namespace minute policy failed, get namespace workload count err: %v", err)
	}
	namespaceMetric := &types.NamespaceMetrics{
		Index:             utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:              primitive.NewDateTimeFromTime(utils.FormatTime(opts.CurrentTime, opts.Dimension)),
		CPUUsage:          cpuMetrics.CPUUsage,
		MemoryRequest:     memoryMetrics.MemoryRequest,
		MemoryUsage:       memoryMetrics.MemoryUsage,
		InstanceCount:     instanceCount,
		CPURequest:        cpuMetrics.CPURequest,
		MemoryLimit:       memoryMetrics.MemoryLimit,
		CPULimit:          cpuMetrics.CPULimit,
		WorkloadCount:     workloadCount,
		CPUUsageAmount:    cpuMetrics.CPUUsed,
		MemoryUsageAmount: memoryMetrics.MemoryUsed,
		MaxCPUUsageTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxCPUUsage",
			MetricName: "MaxCPUUsage",
			Value:      cpuMetrics.CPUUsage,
			Period:     opts.CurrentTime.String(),
		},
		MinCPUUsageTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MinCPUUsage",
			MetricName: "MinCPUUsage",
			Value:      cpuMetrics.CPUUsage,
			Period:     opts.CurrentTime.String(),
		},
		MaxMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxMemoryUsage",
			MetricName: "MaxMemoryUsage",
			Value:      memoryMetrics.MemoryUsage,
			Period:     opts.CurrentTime.String(),
		},
		MinMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MinMemoryUsage",
			MetricName: "MinMemoryUsage",
			Value:      memoryMetrics.MemoryUsage,
			Period:     opts.CurrentTime.String(),
		},
		MinWorkloadUsage: nil,
		MaxWorkloadUsage: nil,
		MinInstanceTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MinInstance",
			MetricName: "MinInstance",
			Value:      float64(instanceCount),
			Period:     opts.CurrentTime.String(),
		},
		MaxInstanceTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxInstance",
			MetricName: "MaxInstance",
			Value:      float64(instanceCount),
			Period:     opts.CurrentTime.String(),
		},
	}
	err = p.store.InsertNamespaceInfo(ctx, namespaceMetric, opts)
	if err != nil {
		blog.Errorf("do namespace minute policy error, opts: %v, err: %v", opts, err)
	}
}
