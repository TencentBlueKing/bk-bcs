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

package datajob

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// WorkloadDayPolicy workload day
type WorkloadDayPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// WorkloadHourPolicy workload hour
type WorkloadHourPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// WorkloadMinutePolicy workload minute
type WorkloadMinutePolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// WorkloadRequestPolicy workload request
type WorkloadRequestPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NewWorkloadDayPolicy init day policy
func NewWorkloadDayPolicy(getter metric.Server, store store.Server) *WorkloadDayPolicy {
	return &WorkloadDayPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewWorkloadHourPolicy init hour policy
func NewWorkloadHourPolicy(getter metric.Server, store store.Server) *WorkloadHourPolicy {
	return &WorkloadHourPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewWorkloadMinutePolicy init minute policy
func NewWorkloadMinutePolicy(getter metric.Server, store store.Server) *WorkloadMinutePolicy {
	return &WorkloadMinutePolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewWorkloadRequestPolicy init minute policy
func NewWorkloadRequestPolicy(getter metric.Server, store store.Server) *WorkloadRequestPolicy {
	return &WorkloadRequestPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// ImplementPolicy is a function that implements the day policy for a workload.
// It takes in the ctx, opts, and clients parameters.
// It retrieves the CPU, memory, and instance count metrics for the workload.
// It creates a new workload metric with the retrieved metrics and inserts it into the database.
func (p *WorkloadDayPolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts, clients *types.Clients) {
	// Retrieve the CPU, memory, and instance count metrics for the workload.
	cpuMetrics, err := p.MetricGetter.GetWorkloadCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do workload day policy error, opts: %v, err: %v", opts, err)
	}
	memoryMetrics, err := p.MetricGetter.GetWorkloadMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do workload day policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do workload day policy error, opts: %v, err: %v", opts, err)
	}
	hourOpts := &types.JobCommonOpts{
		ObjectType:   types.WorkloadType,
		ProjectID:    opts.ProjectID,
		ClusterID:    opts.ClusterID,
		Namespace:    opts.Namespace,
		Dimension:    types.DimensionHour,
		WorkloadType: opts.WorkloadType,
		WorkloadName: opts.WorkloadName,
	}
	bucket, _ := utils.GetBucketTime(opts.CurrentTime.AddDate(0, 0, -1), types.DimensionHour)
	hourMetrics, err := p.store.GetRawWorkloadInfo(ctx, hourOpts, bucket)
	if err != nil {
		blog.Errorf("do workload day policy failed, get workload metrics err:%v", err)
		return
	} else if len(hourMetrics) != 1 {
		blog.Errorf("do workload day policy failed, get workload metrics err, length not equal 1, metrics:%v", hourMetrics)
		return
	}
	hourMetric := hourMetrics[0]
	// Create a new workload metric with the retrieved metrics and insert it into the database.
	workloadMetric := &types.WorkloadMetrics{
		Index:              utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(opts.CurrentTime),
		CPURequest:         cpuMetrics.CPURequest,
		CPUUsage:           cpuMetrics.CPUUsage,
		CPUUsageAmount:     cpuMetrics.CPUUsed,
		MemoryRequest:      memoryMetrics.MemoryRequest,
		MemoryUsage:        memoryMetrics.MemoryUsage,
		MemoryUsageAmount:  memoryMetrics.MemoryUsed,
		MemoryLimit:        memoryMetrics.MemoryLimit,
		CPULimit:           cpuMetrics.CPULimit,
		InstanceCount:      instanceCount,
		MaxCPUUsageTime:    hourMetric.MaxCPUUsageTime,
		MinCPUUsageTime:    hourMetric.MinCPUUsageTime,
		MaxMemoryUsageTime: hourMetric.MaxMemoryUsageTime,
		MinMemoryUsageTime: hourMetric.MinMemoryUsageTime,
		MaxInstanceTime:    hourMetric.MaxInstanceTime,
		MinInstanceTime:    hourMetric.MinInstanceTime,
		MinMemoryTime:      hourMetric.MinMemoryTime,
		MaxMemoryTime:      hourMetric.MaxMemoryTime,
		MinCPUTime:         hourMetric.MinCPUTime,
		MaxCPUTime:         hourMetric.MaxCPUTime,
	}
	err = p.store.InsertWorkloadInfo(ctx, workloadMetric, opts)
	if err != nil {
		blog.Errorf("insert workload info err:%v", err)
	}
}

// ImplementPolicy hour policy implement
// It takes in the ctx, opts, and clients parameters.
// It retrieves the CPU, memory, and instance count metrics for the workload.
// It creates a new workload metric with the retrieved metrics and inserts it into the database.
func (p *WorkloadHourPolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts, clients *types.Clients) {
	cpuMetrics, err := p.MetricGetter.GetWorkloadCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do workload hour policy error, opts: %v, err: %v", opts, err)
	}
	memoryMetrics, err := p.MetricGetter.GetWorkloadMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do workload hour policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do workload hour policy error, opts: %v, err: %v", opts, err)
	}

	minuteOpts := &types.JobCommonOpts{
		ObjectType:   types.WorkloadType,
		ProjectID:    opts.ProjectID,
		ClusterID:    opts.ClusterID,
		Namespace:    opts.Namespace,
		Dimension:    types.DimensionMinute,
		WorkloadType: opts.WorkloadType,
		WorkloadName: opts.WorkloadName,
	}
	bucket, _ := utils.GetBucketTime(opts.CurrentTime.Add((-1)*time.Hour), types.DimensionMinute)
	minuteMetrics, err := p.store.GetRawWorkloadInfo(ctx, minuteOpts, bucket)
	if err != nil {
		blog.Errorf("do workload hour policy failed, get workload metrics err:%v", err)
		return
	} else if len(minuteMetrics) != 1 {
		blog.Errorf("do workload hour policy failed, get workload metrics err, length not equal 1, metrics:%v", minuteMetrics)
		return
	}
	minuteMetric := minuteMetrics[0]
	workloadMetric := &types.WorkloadMetrics{
		Index:              utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(opts.CurrentTime),
		CPURequest:         cpuMetrics.CPURequest,
		CPUUsage:           cpuMetrics.CPUUsage,
		CPUUsageAmount:     cpuMetrics.CPUUsed,
		MemoryRequest:      memoryMetrics.MemoryRequest,
		MemoryUsage:        memoryMetrics.MemoryUsage,
		MemoryUsageAmount:  memoryMetrics.MemoryUsed,
		MemoryLimit:        memoryMetrics.MemoryLimit,
		CPULimit:           cpuMetrics.CPULimit,
		InstanceCount:      instanceCount,
		MaxCPUUsageTime:    minuteMetric.MaxCPUUsageTime,
		MinCPUUsageTime:    minuteMetric.MinCPUUsageTime,
		MaxMemoryUsageTime: minuteMetric.MaxMemoryUsageTime,
		MinMemoryUsageTime: minuteMetric.MinMemoryUsageTime,
		MaxInstanceTime:    minuteMetric.MaxInstanceTime,
		MinInstanceTime:    minuteMetric.MinInstanceTime,
		MinMemoryTime:      minuteMetric.MinMemoryTime,
		MaxMemoryTime:      minuteMetric.MaxMemoryTime,
		MinCPUTime:         minuteMetric.MinCPUTime,
		MaxCPUTime:         minuteMetric.MaxCPUTime,
	}
	err = p.store.InsertWorkloadInfo(ctx, workloadMetric, opts)
	if err != nil {
		blog.Errorf("insert workload info err:%v", err)
	}
}

// ImplementPolicy minute policy implement
// It takes in the ctx, opts, and clients parameters.
// It retrieves the CPU, memory, and instance count metrics for the workload.
// It creates a new workload metric with the retrieved metrics and inserts it into the database.
func (p *WorkloadMinutePolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) {
	cpuMetrics, err := p.MetricGetter.GetWorkloadCPUMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do workload minute policy error, opts: %v, err: %v", opts, err)
	}
	memoryMetrics, err := p.MetricGetter.GetWorkloadMemoryMetrics(opts, clients)
	if err != nil {
		blog.Errorf("do workload minute policy error, opts: %v, err: %v", opts, err)
	}
	instanceCount, err := p.MetricGetter.GetInstanceCount(opts, clients)
	if err != nil {
		blog.Errorf("do workload minute policy error, opts: %v, err: %v", opts, err)
	}
	workloadMetric := &types.WorkloadMetrics{
		Index:             utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:              primitive.NewDateTimeFromTime(opts.CurrentTime),
		CPURequest:        cpuMetrics.CPURequest,
		CPUUsage:          cpuMetrics.CPUUsage,
		CPUUsageAmount:    cpuMetrics.CPUUsed,
		MemoryLimit:       memoryMetrics.MemoryLimit,
		CPULimit:          cpuMetrics.CPULimit,
		MemoryRequest:     memoryMetrics.MemoryRequest,
		MemoryUsage:       memoryMetrics.MemoryUsage,
		MemoryUsageAmount: memoryMetrics.MemoryUsed,
		InstanceCount:     instanceCount,
		MaxCPUUsageTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxCpuUsage",
			MetricName: "MaxCpuUsage",
			Value:      cpuMetrics.CPUUsage,
			Period:     opts.CurrentTime.String(),
		},
		MinCPUUsageTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MinCpuUsage",
			MetricName: "MinCpuUsage",
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
		MaxInstanceTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxInstance",
			MetricName: "MaxInstance",
			Value:      float64(instanceCount),
			Period:     opts.CurrentTime.String(),
		},
		MinInstanceTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MinInstance",
			MetricName: "MinInstance",
			Value:      float64(instanceCount),
			Period:     opts.CurrentTime.String(),
		},
		MaxCPUTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxCPU",
			MetricName: "MaxCPU",
			Value:      cpuMetrics.CPUUsed,
			Period:     opts.CurrentTime.String(),
		},
		MinCPUTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MinCPU",
			MetricName: "MinCPU",
			Value:      cpuMetrics.CPUUsed,
			Period:     opts.CurrentTime.String(),
		},
		MaxMemoryTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MaxMemory",
			MetricName: "MaxMemory",
			Value:      float64(memoryMetrics.MemoryUsed),
			Period:     opts.CurrentTime.String(),
		},
		MinMemoryTime: &bcsdatamanager.ExtremumRecord{
			Name:       "MinMemory",
			MetricName: "MinMemory",
			Value:      float64(memoryMetrics.MemoryUsed),
			Period:     opts.CurrentTime.String(),
		},
	}
	if err = p.store.InsertWorkloadInfo(ctx, workloadMetric, opts); err != nil {
		blog.Errorf("insert workload info err:%v", err)
	}
}

// ImplementPolicy implement policy
func (p *WorkloadRequestPolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) {
	if opts.ClusterType != types.Kubernetes {
		return
	}
	deployments, err := clients.K8sStorageCli.QueryK8SDeployment(opts.ClusterID, opts.Namespace)
	if err != nil {
		blog.Errorf("[handler] query %s %s deployment list error:%s", opts.ClusterID, opts.Namespace, err.Error())
		return
	}
	for key := range deployments {
		deploy := deployments[key]
		go func(workload *storage.Deployment) {
			cpu, memory := getWorkloadContainerRequest(workload.Data.Spec.Template.Spec)
			workloadResult := &types.WorkloadOriginRequestResult{
				ProjectID:    opts.ProjectID,
				ClusterID:    opts.ClusterID,
				Namespace:    opts.Namespace,
				WorkloadType: workload.ResourceType,
				WorkloadName: workload.ResourceName,
				Cpu:          cpu,
				Memory:       memory,
				CreateTime:   time.Now(),
			}
			if err = p.store.CreateWorkloadOriginRequest(ctx, workloadResult); err != nil {
				blog.Errorf("CreateWorkloadOriginRequest error:%s, clusterID:%s, namespace:%s, deploy:%s",
					err.Error(), opts.ClusterID, opts.Namespace, workload.ResourceName)
			}
		}(deploy)
	}
	statefulsets, err := clients.K8sStorageCli.QueryK8SStatefulSet(opts.ClusterID, opts.Namespace)
	if err != nil {
		blog.Errorf("[handler] query %s %s statefulset list error:%s", opts.ClusterID, opts.Namespace, err.Error())
		return
	}
	for key := range statefulsets {
		statefulSet := statefulsets[key]
		go func(workload *storage.StatefulSet) {
			cpu, memory := getWorkloadContainerRequest(workload.Data.Spec.Template.Spec)
			workloadResult := &types.WorkloadOriginRequestResult{
				ClusterID:    opts.ClusterID,
				Namespace:    opts.Namespace,
				WorkloadType: workload.ResourceType,
				WorkloadName: workload.ResourceName,
				Cpu:          cpu,
				Memory:       memory,
				CreateTime:   time.Now(),
			}
			if err = p.store.CreateWorkloadOriginRequest(ctx, workloadResult); err != nil {
				blog.Errorf("CreateWorkloadOriginRequest error:%s, clusterID:%s, namespace:%s, statefulset:%s",
					err.Error(), opts.ClusterID, opts.Namespace, workload.ResourceName)
			}
		}(statefulSet)
	}
	daemonSets, err := clients.K8sStorageCli.QueryK8SDaemonSet(opts.ClusterID, opts.Namespace)
	if err != nil {
		blog.Errorf("[handler] query %s %s daemonSet list error:%s", opts.ClusterID, opts.Namespace, err.Error())
		return
	}
	for key := range daemonSets {
		daemonSet := daemonSets[key]
		go func(workload *storage.DaemonSet) {
			cpu, memory := getWorkloadContainerRequest(workload.Data.Spec.Template.Spec)
			workloadResult := &types.WorkloadOriginRequestResult{
				ClusterID:    opts.ClusterID,
				Namespace:    opts.Namespace,
				WorkloadType: workload.ResourceType,
				WorkloadName: workload.ResourceName,
				Cpu:          cpu,
				Memory:       memory,
				CreateTime:   time.Now(),
			}
			if err = p.store.CreateWorkloadOriginRequest(ctx, workloadResult); err != nil {
				blog.Errorf("CreateWorkloadOriginRequest error:%s, clusterID:%s, namespace:%s, daemonset:%s",
					err.Error(), opts.ClusterID, opts.Namespace, workload.ResourceName)
			}
		}(daemonSet)
	}
	gameDeploys, err := clients.K8sStorageCli.QueryK8SGameDeployment(opts.ClusterID, opts.Namespace)
	if err != nil {
		blog.Errorf("[handler] query %s %s gameDeploy list error:%s", opts.ClusterID, opts.Namespace, err.Error())
		return
	}
	for key := range gameDeploys {
		gameDeploy := gameDeploys[key]
		go func(workload *storage.GameDeployment) {
			cpu, memory := getWorkloadContainerRequest(workload.Data.Spec.Template.Spec)
			workloadResult := &types.WorkloadOriginRequestResult{
				ClusterID:    opts.ClusterID,
				Namespace:    opts.Namespace,
				WorkloadType: workload.ResourceType,
				WorkloadName: workload.ResourceName,
				Cpu:          cpu,
				Memory:       memory,
				CreateTime:   time.Now(),
			}
			if err = p.store.CreateWorkloadOriginRequest(ctx, workloadResult); err != nil {
				blog.Errorf("CreateWorkloadOriginRequest error:%s, clusterID:%s, namespace:%s, gameDeploy:%s",
					err.Error(), opts.ClusterID, opts.Namespace, workload.ResourceName)
			}
		}(gameDeploy)
	}
	gameStatefuls, err := clients.K8sStorageCli.QueryK8SGameStatefulSet(opts.ClusterID, opts.Namespace)
	if err != nil {
		blog.Errorf("[handler] query %s %s gameStatefuls list error:%s", opts.ClusterID, opts.Namespace, err.Error())
		return
	}
	for key := range gameStatefuls {
		gameStateful := gameStatefuls[key]
		go func(workload *storage.GameStatefulSet) {
			cpu, memory := getWorkloadContainerRequest(workload.Data.Spec.Template.Spec)
			workloadResult := &types.WorkloadOriginRequestResult{
				ClusterID:    opts.ClusterID,
				Namespace:    opts.Namespace,
				WorkloadType: workload.ResourceType,
				WorkloadName: workload.ResourceName,
				Cpu:          cpu,
				Memory:       memory,
				CreateTime:   time.Now(),
			}
			if err = p.store.CreateWorkloadOriginRequest(ctx, workloadResult); err != nil {
				blog.Errorf("CreateWorkloadOriginRequest error:%s, clusterID:%s, namespace:%s, gameStateful:%s",
					err.Error(), opts.ClusterID, opts.Namespace, workload.ResourceName)
			}
		}(gameStateful)
	}
}

func getWorkloadContainerRequest(podSpec v1.PodSpec) ([]*bcsdatamanager.WorkloadOriginRequestContainer,
	[]*bcsdatamanager.WorkloadOriginRequestContainer) {
	cpu := make([]*bcsdatamanager.WorkloadOriginRequestContainer, 0)
	memory := make([]*bcsdatamanager.WorkloadOriginRequestContainer, 0)
	for _, container := range podSpec.Containers {
		var cpuLimitStr, cpuRequestStr, memoryLimitStr, memoryRequestStr string
		if cpuRequest, ok := container.Resources.Requests[v1.ResourceCPU]; ok {
			cpuRequestStr = cpuRequest.String()
		}
		if cpuLimit, ok := container.Resources.Limits[v1.ResourceCPU]; ok {
			cpuLimitStr = cpuLimit.String()
		}
		if memoryRequest, ok := container.Resources.Requests[v1.ResourceMemory]; ok {
			memoryRequestStr = memoryRequest.String()
		}
		if memoryLimit, ok := container.Resources.Limits[v1.ResourceMemory]; ok {
			memoryLimitStr = memoryLimit.String()
		}
		cpuContainer := &bcsdatamanager.WorkloadOriginRequestContainer{
			Container: container.Name,
			Request:   cpuRequestStr,
			Limit:     cpuLimitStr,
		}
		memoryContainer := &bcsdatamanager.WorkloadOriginRequestContainer{
			Container: container.Name,
			Request:   memoryRequestStr,
			Limit:     memoryLimitStr,
		}
		cpu = append(cpu, cpuContainer)
		memory = append(memory, memoryContainer)
	}
	return cpu, memory
}
