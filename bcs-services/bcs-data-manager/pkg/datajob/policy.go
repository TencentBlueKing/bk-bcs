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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
)

var (
	// PolicyMap : type:map
	PolicyMap    = map[string]map[string]Policy{}
	ClusterMap   = map[string]Policy{}
	NamespaceMap = map[string]Policy{}
	WorkloadMap  = map[string]Policy{}
	PublicMap    = map[string]Policy{}
	ProjectMap   = map[string]Policy{}
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
	PolicyMap[common.ClusterType] = ClusterMap
	PolicyMap[common.NamespaceType] = NamespaceMap
	PolicyMap[common.WorkloadType] = WorkloadMap
	PolicyMap[common.ProjectType] = ProjectMap
	PolicyMap[common.PublicType] = PublicMap
}
func (f *policyFactory) initClusterMap() {
	ClusterMap[common.DimensionMinute] = NewClusterMinutePolicy(&metric.MetricGetter{}, f.store)
	ClusterMap[common.DimensionHour] = NewClusterHourPolicy(&metric.MetricGetter{}, f.store)
	ClusterMap[common.DimensionDay] = NewClusterDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initNamespaceMap() {
	NamespaceMap[common.DimensionMinute] = NewNamespaceMinutePolicy(&metric.MetricGetter{}, f.store)
	NamespaceMap[common.DimensionHour] = NewNamespaceHourPolicy(&metric.MetricGetter{}, f.store)
	NamespaceMap[common.DimensionDay] = NewNamespaceDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initWorkloadMap() {
	WorkloadMap[common.DimensionMinute] = NewWorkloadMinutePolicy(&metric.MetricGetter{}, f.store)
	WorkloadMap[common.DimensionHour] = NewWorkloadHourPolicy(&metric.MetricGetter{}, f.store)
	WorkloadMap[common.DimensionDay] = NewWorkloadDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initPublicMap() {
	PublicMap[common.DimensionDay] = NewPublicDayPolicy(&metric.MetricGetter{}, f.store)
}

func (f *policyFactory) initProjectMap() {
	ProjectMap[common.DimensionDay] = NewProjectDayPolicy(&metric.MetricGetter{}, f.store)
}
