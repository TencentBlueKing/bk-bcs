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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPolicyFactory(t *testing.T) {
	storeServer := mock.NewMockStore()
	factory := NewPolicyFactory(storeServer)
	factory.Init()
	assert.NotNil(t, factory.GetPolicy(types.ClusterType, types.DimensionMinute))
	assert.Equal(t, 5, len(PolicyMap))
}

func Test_policyFactory_GetPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	factory := policyFactory{store: storeServer}
	factory.Init()
	assert.NotNil(t, factory.GetPolicy(types.ClusterType, types.DimensionMinute))
	assert.NotNil(t, factory.GetPolicy(types.ClusterType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.ClusterType, types.DimensionDay))
	assert.NotNil(t, factory.GetPolicy(types.NamespaceType, types.DimensionMinute))
	assert.NotNil(t, factory.GetPolicy(types.NamespaceType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.NamespaceType, types.DimensionDay))
	assert.Nil(t, factory.GetPolicy(types.ProjectType, types.DimensionMinute))
	assert.Nil(t, factory.GetPolicy(types.ProjectType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.ProjectType, types.DimensionDay))
	assert.Nil(t, factory.GetPolicy(types.PublicType, types.DimensionMinute))
	assert.Nil(t, factory.GetPolicy(types.PublicType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.PublicType, types.DimensionDay))
	assert.NotNil(t, factory.GetPolicy(types.WorkloadType, types.DimensionMinute))
	assert.NotNil(t, factory.GetPolicy(types.WorkloadType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.WorkloadType, types.DimensionDay))
}

func Test_policyFactory_initClusterMap(t *testing.T) {
	storeServer := mock.NewMockStore()
	factory := policyFactory{store: storeServer}
	factory.initClusterMap()
	PolicyMap[types.ClusterType] = ClusterMap
	assert.NotNil(t, factory.GetPolicy(types.ClusterType, types.DimensionMinute))
	assert.NotNil(t, factory.GetPolicy(types.ClusterType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.ClusterType, types.DimensionDay))
	assert.Equal(t, 3, len(PolicyMap[types.ClusterType]))
}

func Test_policyFactory_initNamespaceMap(t *testing.T) {
	storeServer := mock.NewMockStore()
	factory := policyFactory{store: storeServer}
	factory.initNamespaceMap()
	PolicyMap[types.NamespaceType] = NamespaceMap
	assert.NotNil(t, factory.GetPolicy(types.NamespaceType, types.DimensionMinute))
	assert.NotNil(t, factory.GetPolicy(types.NamespaceType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.NamespaceType, types.DimensionDay))
	assert.Equal(t, 3, len(PolicyMap[types.NamespaceType]))
}

func Test_policyFactory_initProjectMap(t *testing.T) {
	storeServer := mock.NewMockStore()
	factory := policyFactory{store: storeServer}
	factory.initProjectMap()
	PolicyMap[types.ProjectType] = ProjectMap
	assert.Nil(t, factory.GetPolicy(types.ProjectType, types.DimensionMinute))
	assert.Nil(t, factory.GetPolicy(types.ProjectType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.ProjectType, types.DimensionDay))
	assert.Equal(t, 1, len(PolicyMap[types.ProjectType]))
}

func Test_policyFactory_initPublicMap(t *testing.T) {
	storeServer := mock.NewMockStore()
	factory := policyFactory{store: storeServer}
	factory.initPublicMap()
	PolicyMap[types.PublicType] = PublicMap
	assert.Nil(t, factory.GetPolicy(types.PublicType, types.DimensionMinute))
	assert.Nil(t, factory.GetPolicy(types.PublicType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.PublicType, types.DimensionDay))
	assert.Equal(t, 1, len(PolicyMap[types.PublicType]))
}

func Test_policyFactory_initWorkloadMap(t *testing.T) {
	storeServer := mock.NewMockStore()
	factory := policyFactory{store: storeServer}
	factory.initWorkloadMap()
	PolicyMap[types.WorkloadType] = WorkloadMap
	assert.NotNil(t, factory.GetPolicy(types.WorkloadType, types.DimensionMinute))
	assert.NotNil(t, factory.GetPolicy(types.WorkloadType, types.DimensionHour))
	assert.NotNil(t, factory.GetPolicy(types.WorkloadType, types.DimensionDay))
	assert.Equal(t, 3, len(PolicyMap[types.WorkloadType]))
}
