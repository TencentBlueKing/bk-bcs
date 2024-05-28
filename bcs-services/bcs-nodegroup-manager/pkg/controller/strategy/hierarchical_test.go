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

package strategy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/mocks"
	resourcemock "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/resourcemgr/mocks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	storagemock "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mocks"
)

func TestHierarchical_IsAbleToScaleDown(t *testing.T) {

}

func TestHierarchical_IsAbleToScaleUp(t *testing.T) {
	tests := []struct {
		name       string
		strategy   *storage.NodeGroupMgrStrategy
		wantNum    int
		wantResult bool
		wantErr    bool
		on         func(f *MockFields)
	}{
		{
			name:       "taskExecuting",
			strategy:   getTestStrategy(),
			wantNum:    0,
			wantResult: false,
			wantErr:    false,
			on: func(f *MockFields) {
				f.storage.On("ListTasksByStrategy", "test-strategy", mock.Anything).
					Return([]*storage.ScaleDownTask{{
						TaskID:            "task1",
						TotalNum:          2,
						NodeGroupStrategy: "test-strategy",
						Deadline:          time.Now(),
						IsDeleted:         false,
						IsExecuted:        true,
					}}, nil)
			},
		},
		{
			name:       "scaleUp",
			strategy:   getTestStrategy(),
			wantNum:    5,
			wantResult: true,
			wantErr:    false,
			on: func(f *MockFields) {
				f.storage.On("ListTasksByStrategy", "test-strategy", mock.Anything).
					Return([]*storage.ScaleDownTask{{
						TaskID:            "task1",
						TotalNum:          2,
						NodeGroupStrategy: "test-strategy",
						Deadline:          time.Now().Add(24 * time.Hour),
						IsDeleted:         false,
						IsExecuted:        false,
					}}, nil)
				f.resourceCli.On("GetResourcePoolByCondition", "test-resource-pool", "consumer1", "", mock.Anything).
					Return(&storage.ResourcePool{
						ID:          "test-resource-pool",
						Name:        "test-resource-pool",
						CreatedTime: time.Now(),
						UpdatedTime: time.Now(),
						InitNum:     0,
						IdleNum:     5,
						ConsumedNum: 5,
						ReturnedNum: 0,
					}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				resourceCli:   resourcemock.NewClient(t),
				storage:       storagemock.NewStorage(t),
				clusterClient: mocks.NewClient(t),
			}
			tt.on(mockFields)
			opts := &Options{
				ResourceManager: mockFields.resourceCli,
				Storage:         mockFields.storage,
			}
			executor := NewHierarchicalStrategyExecutor(opts)
			num, result, _, err := executor.IsAbleToScaleUp(tt.strategy)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantNum, num)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

type MockFields struct {
	clusterClient *mocks.Client
	storage       *storagemock.Storage
	resourceCli   *resourcemock.Client
}

func getTestStrategy() *storage.NodeGroupMgrStrategy {
	strategy := &storage.NodeGroupMgrStrategy{
		Name:              "test-strategy",
		ResourcePool:      "test-resource-pool",
		ReservedNodeGroup: &storage.GroupInfo{ClusterID: "reserved-clusterID", NodeGroupID: "reserved-nodeGroupID"},
		ElasticNodeGroups: []*storage.GroupInfo{
			{ClusterID: "Cluster1", NodeGroupID: "NodeGroup1", Weight: 5, ConsumerID: "consumer1"},
			{ClusterID: "Cluster1", NodeGroupID: "NodeGroup2", Weight: 5, ConsumerID: "consumer2"},
		},
		Strategy: &storage.Strategy{
			Type:            storage.HierarchicalStrategyType,
			ScaleUpCoolDown: 0,
			ScaleUpDelay:    5,
			MinScaleUpSize:  2,
			ScaleDownDelay:  5,
			MaxIdleDelay:    0,
			Buffer: &storage.BufferStrategy{
				Low:  0,
				High: 0,
			},
		},
		Status: &storage.State{
			Status:      storage.InitState,
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	return strategy
}
