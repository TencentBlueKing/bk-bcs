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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

var (
	// PolicyMap : type:map
	PolicyMap = map[string]map[string]Policy{}
	// ClusterMap cluster map
	ClusterMap = map[string]Policy{}
	// NamespaceMap namespace map
	NamespaceMap = map[string]Policy{}
	// WorkloadMap workload map
	WorkloadMap = map[string]Policy{}
	// PublicMap policy map
	PublicMap = map[string]Policy{}
	// ProjectMap policy map
	ProjectMap = map[string]Policy{}
	// PodAutoscalerMap hpa/gpa map
	PodAutoscalerMap = map[string]Policy{}
)

// PolicyFactoryInterface PolicyMap interface
type PolicyFactoryInterface interface {
	GetPolicy(objectType, dimension string) Policy
	Init()
}

type policyFactory struct {
	store store.Server
}

// NewPolicyFactory init policy factory
func NewPolicyFactory(store store.Server) PolicyFactoryInterface {
	return &policyFactory{store: store}
}

// GetPolicy get policy by type and dimension
func (f *policyFactory) GetPolicy(objectType, dimension string) Policy {
	return PolicyMap[objectType][dimension]
}

// Init factory init
func (f *policyFactory) Init() {
	f.initClusterMap()
	f.initNamespaceMap()
	f.initWorkloadMap()
	f.initPublicMap()
	f.initProjectMap()
	f.initPodAutoscalerMap()
	PolicyMap[types.ClusterType] = ClusterMap
	PolicyMap[types.NamespaceType] = NamespaceMap
	PolicyMap[types.WorkloadType] = WorkloadMap
	PolicyMap[types.ProjectType] = ProjectMap
	PolicyMap[types.PublicType] = PublicMap
	PolicyMap[types.PodAutoscalerType] = PodAutoscalerMap
}
func (f *policyFactory) initClusterMap() {
	ClusterMap[types.DimensionMinute] = NewClusterMinutePolicy(&metric.MetricGetter{}, f.store)
	ClusterMap[types.DimensionHour] = NewClusterHourPolicy(&metric.MetricGetter{}, f.store)
	ClusterMap[types.DimensionDay] = NewClusterDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initNamespaceMap() {
	NamespaceMap[types.DimensionMinute] = NewNamespaceMinutePolicy(&metric.MetricGetter{}, f.store)
	NamespaceMap[types.DimensionHour] = NewNamespaceHourPolicy(&metric.MetricGetter{}, f.store)
	NamespaceMap[types.DimensionDay] = NewNamespaceDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initWorkloadMap() {
	WorkloadMap[types.DimensionMinute] = NewWorkloadMinutePolicy(&metric.MetricGetter{}, f.store)
	WorkloadMap[types.DimensionHour] = NewWorkloadHourPolicy(&metric.MetricGetter{}, f.store)
	WorkloadMap[types.DimensionDay] = NewWorkloadDayPolicy(&metric.MetricGetter{}, f.store)
	WorkloadMap[types.GetWorkloadRequestType] = NewWorkloadRequestPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initPublicMap() {
	PublicMap[types.DimensionDay] = NewPublicDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initProjectMap() {
	ProjectMap[types.DimensionDay] = NewProjectDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initPodAutoscalerMap() {
	PodAutoscalerMap[types.DimensionDay] = NewPodAutoscalerDayPolicy(&metric.MetricGetter{}, f.store)
	PodAutoscalerMap[types.DimensionHour] = NewPodAutoscalerHourPolicy(&metric.MetricGetter{}, f.store)
	PodAutoscalerMap[types.DimensionMinute] = NewPodAutoscalerMinutePolicy(&metric.MetricGetter{}, f.store)
}
