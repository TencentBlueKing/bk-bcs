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

func TestIsAbleToScaleDown(t *testing.T) {
	// construct test data
	testcases := []struct {
		name              string
		strategy          *storage.NodeGroupMgrStrategy
		expectedNum       int
		expectedScaleDown bool
		message           string
		on                func(f *MockFields)
	}{
		{ //resource is not idle
			name: "scaleDown",
			strategy: &storage.NodeGroupMgrStrategy{
				Name:              "test",
				ResourcePool:      "testpool",
				ReservedNodeGroup: &storage.GroupInfo{ConsumerID: "consumer1"},
				Strategy: &storage.Strategy{
					MaxIdleDelay:   3,
					MinScaleUpSize: 3,
					Buffer: &storage.BufferStrategy{
						High: 15,
						Low:  10,
					},
				},
			},
			expectedNum:       5,
			expectedScaleDown: true,
			message:           "resource is not idle",
			on: func(f *MockFields) {
				f.resourceCli.On("GetResourcePoolByCondition", "testpool", "consumer1", "", mock.Anything).
					Return(&storage.ResourcePool{
						// pool max size 100
						UpdatedTime: time.Now(),
						InitNum:     2,
						IdleNum:     3,
						ReturnedNum: 0,
						ConsumedNum: 95,
					}, nil)
			},
		},
		{ //resource is idle
			name: "notScaleDown",
			strategy: &storage.NodeGroupMgrStrategy{
				ResourcePool: "testpool",
				ReservedNodeGroup: &storage.GroupInfo{
					ConsumerID: "consumer1",
				},
				Strategy: &storage.Strategy{
					MaxIdleDelay:   3,
					MinScaleUpSize: 3,
					Buffer: &storage.BufferStrategy{
						High: 15,
						Low:  10,
					},
				},
			},
			on: func(f *MockFields) {
				f.resourceCli.On("GetResourcePoolByCondition", "testpool", "consumer1", "", mock.Anything).
					Return(&storage.ResourcePool{ // pool max size 100
						UpdatedTime: time.Now(),
						InitNum:     0,
						IdleNum:     15,
						ReturnedNum: 0,
						ConsumedNum: 85,
					}, nil)
			},
			expectedNum:       0,
			expectedScaleDown: false,
			message:           "resource is idle",
		},
	}
	for _, tt := range testcases {
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
			executor := NewBufferStrategyExecutor(opts)
			num, result, _ := executor.IsAbleToScaleDown(tt.strategy)
			assert.Equal(t, tt.expectedScaleDown, result)
			assert.Equal(t, tt.expectedNum, num)
		})
	}
}

func TestIsAbleToScaleUp(t *testing.T) {
	// construct test data
	testcases := []struct {
		name            string
		strategy        *storage.NodeGroupMgrStrategy
		expectedNum     int
		expectedScaleUp bool
		message         string
		on              func(f *MockFields)
	}{
		{ //resource is not idle enough
			name: "notScaleUp",
			strategy: &storage.NodeGroupMgrStrategy{
				ResourcePool: "testpool",
				ReservedNodeGroup: &storage.GroupInfo{
					ConsumerID: "consumer1",
				},
				Strategy: &storage.Strategy{
					MaxIdleDelay:   3,
					MinScaleUpSize: 3,
					Buffer: &storage.BufferStrategy{
						High: 15,
						Low:  10,
					},
				},
			},
			expectedNum:     0,
			expectedScaleUp: false,
			message:         "resource is not idle enough",
			on: func(f *MockFields) {
				f.resourceCli.On("GetResourcePoolByCondition", "testpool", "consumer1", "", mock.Anything).
					Return(&storage.ResourcePool{
						// pool max size 100
						UpdatedTime: time.Now(),
						InitNum:     0,
						IdleNum:     10,
						ReturnedNum: 0,
						ConsumedNum: 90,
					}, nil)
			},
		},
		{ //resource is not idle enough time
			name: "notEnoughTime",
			strategy: &storage.NodeGroupMgrStrategy{
				ResourcePool: "testpool",
				ReservedNodeGroup: &storage.GroupInfo{
					ConsumerID: "consumer1",
				},
				Strategy: &storage.Strategy{
					MaxIdleDelay:   3,
					MinScaleUpSize: 3,
					Buffer: &storage.BufferStrategy{
						High: 15,
						Low:  10,
					},
				},
			},
			expectedNum:     0,
			expectedScaleUp: false,
			message:         "resource is not idle enough time",
			on: func(f *MockFields) {
				f.resourceCli.On("GetResourcePoolByCondition", "testpool", "consumer1", "", mock.Anything).
					Return(&storage.ResourcePool{ // pool max size 100
						UpdatedTime: time.Now(),
						InitNum:     0,
						IdleNum:     20,
						ReturnedNum: 0,
						ConsumedNum: 80,
					}, nil)
			},
		},
		{ // MinScaleUpSize limitation
			name: "MinScaleUpSizeLimitation",
			strategy: &storage.NodeGroupMgrStrategy{
				ResourcePool: "testpool",
				ReservedNodeGroup: &storage.GroupInfo{
					ConsumerID: "consumer1",
				},
				Strategy: &storage.Strategy{
					MaxIdleDelay:   3,
					MinScaleUpSize: 3,
					Buffer: &storage.BufferStrategy{
						High: 15,
						Low:  10,
					},
				},
			},
			expectedNum:     0,
			expectedScaleUp: false,
			message:         "limit by MinScaleUpSize",
			on: func(f *MockFields) {
				f.resourceCli.On("GetResourcePoolByCondition", "testpool", "consumer1", "", mock.Anything).
					Return(&storage.ResourcePool{ // pool max size 100
						UpdatedTime: time.Now().AddDate(0, 0, -1),
						InitNum:     10,
						IdleNum:     6,
						ReturnedNum: 0,
						ConsumedNum: 84,
					}, nil)
			},
		},
		{
			name: "scaleUp",
			strategy: &storage.NodeGroupMgrStrategy{
				ResourcePool: "testpool",
				ReservedNodeGroup: &storage.GroupInfo{
					ConsumerID: "consumer1",
				},
				Strategy: &storage.Strategy{
					MaxIdleDelay:   3,
					MinScaleUpSize: 3,
					Buffer: &storage.BufferStrategy{
						High: 15,
						Low:  10,
					},
				},
			},
			expectedNum:     25,
			expectedScaleUp: true,
			message:         "normal situation failure",
			on: func(f *MockFields) {
				f.resourceCli.On("GetResourcePoolByCondition", "testpool", "consumer1", "", mock.Anything).
					Return(&storage.ResourcePool{ // pool max size 100
						UpdatedTime: time.Now().AddDate(0, 0, -1),
						InitNum:     10,
						IdleNum:     20,
						ReturnedNum: 10,
						ConsumedNum: 60,
					}, nil)
			},
		},
	}
	for _, tt := range testcases {
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
			executor := NewBufferStrategyExecutor(opts)
			num, result, _ := executor.IsAbleToScaleUp(tt.strategy)
			assert.Equal(t, tt.expectedScaleUp, result)
			assert.Equal(t, tt.expectedNum, num)
		})
	}
}
