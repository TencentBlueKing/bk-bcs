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

package controller

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/mocks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

func TestSimpleBalancer_Ceil(t *testing.T) {
	// init balancer test data
	assertion := assert.New(t)
	groups := []*storage.GroupInfo{
		{NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", Weight: 10},
		{NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", Weight: 10},
		{NodeGroupID: "NodeGroup3", ClusterID: "ClusterID2", Weight: 10},
	}
	allo := newSimpleBalancer(groups)
	nodeGroups := allo.distribute(20)
	// expect result: 6, 6, 8
	expected := []int{6, 6, 8}
	for i, node := range nodeGroups {
		assertion.Equal(expected[i], node.partition)
	}
}

func TestSimpleBalancer_Floor(t *testing.T) {
	// init balancer test data
	assertion := assert.New(t)
	groups := []*storage.GroupInfo{
		{NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", Weight: 9},
		{NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", Weight: 1},
		{NodeGroupID: "NodeGroup3", ClusterID: "ClusterID2", Weight: 1},
	}
	allo := newSimpleBalancer(groups)
	nodeGroups := allo.distribute(20)
	// expect result: 6, 6, 8
	expected := []int{1, 1, 18}
	for i, node := range nodeGroups {
		assertion.Equal(expected[i], node.partition)
	}
}

func TestWeightBalancer_Limitation(t *testing.T) {
	assertion := assert.New(t)
	// init simple GroupInfo
	groups := []*storage.GroupInfo{
		{NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", Weight: 9},
		{NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", Weight: 1},
		{NodeGroupID: "NodeGroup3", ClusterID: "ClusterID2", Weight: 1},
	}
	// init simple NodeGroup info
	nodeGroups := map[string]*storage.NodeGroup{
		"NodeGroup1": {
			NodeGroupID: "NodeGroup1", MaxSize: 30, MinSize: 10, DesiredSize: 15, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
		},
		"NodeGroup2": {
			NodeGroupID: "NodeGroup2", MaxSize: 30, MinSize: 0, DesiredSize: 15, CmDesiredSize: 5,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP"},
		},
		"NodeGroup3": {
			NodeGroupID: "NodeGroup3", MaxSize: 30, MinSize: 0, DesiredSize: 10, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
		},
	}

	allo := newWeightBalancer(groups, nodeGroups)
	nodeInfos := allo.distribute(20)
	// nodegroup sorted: NodeGroup2, NodeGroup3, NodeGroup1
	// expect result: 5, 10, 0
	expected := []int{5, 10, 0}
	for i, node := range nodeInfos {
		assertion.Equal(expected[i], node.partition)
	}
}

func TestWeightBalancer(t *testing.T) {
	assertion := assert.New(t)
	// init simple GroupInfo
	groups := []*storage.GroupInfo{
		{NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", Weight: 2},
		{NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", Weight: 0},
	}
	// init simple NodeGroup info
	nodeGroups := map[string]*storage.NodeGroup{
		"NodeGroup1": {
			NodeGroupID: "NodeGroup1", MaxSize: 4, MinSize: 0, DesiredSize: 15, CmDesiredSize: 2,
			NodeIPs: []string{"IP", "IP"},
		},
		"NodeGroup2": {
			NodeGroupID: "NodeGroup2", MaxSize: 4, MinSize: 0, DesiredSize: 15, CmDesiredSize: 2,
			NodeIPs: []string{"IP", "IP"},
		},
	}

	allo := newWeightBalancer(groups, nodeGroups)
	nodeInfos := allo.distribute(2)
	// nodegroup sorted: NodeGroup2, NodeGroup3, NodeGroup1
	// expect result: 5, 10, 0
	expected := []int{1, 1}
	for i, node := range nodeInfos {
		assertion.Equal(expected[i], node.partition)
	}
}

func Test_getNodegroupLimitCount(t *testing.T) {
	allReadyNodes := []*cluster.Node{{
		Name: "test1",
		IP:   "test1",
	}, {
		Name: "test2",
		IP:   "test2",
	}, {
		Name: "test3",
		IP:   "test3",
	},
		{
			Name: "test4",
			IP:   "test4",
		}}
	test := []struct {
		name   string
		limit  *storage.NodegroupLimit
		ng     *storage.NodeGroup
		taskID string
		want   int
		on     func(f *MockFields)
	}{
		{
			name: "cluster-limit-0",
			limit: &storage.NodegroupLimit{
				NodegroupLimit:    false,
				NodegroupLimitNum: 0,
				ClusterLimit:      true,
				ClusterLimitNum:   0,
			},
			ng: &storage.NodeGroup{
				NodeGroupID:   "test-ng",
				ClusterID:     "test-cluster",
				MaxSize:       100,
				MinSize:       0,
				CmDesiredSize: 3,
				DesiredSize:   30,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "test-cluster").Return(allReadyNodes, nil)
			},
			want: 1996,
		},
		{
			name: "cluster-limit-100",
			limit: &storage.NodegroupLimit{
				NodegroupLimit:    false,
				NodegroupLimitNum: 0,
				ClusterLimit:      true,
				ClusterLimitNum:   100,
			},
			ng: &storage.NodeGroup{
				NodeGroupID:   "test-ng",
				ClusterID:     "test-cluster",
				MaxSize:       100,
				MinSize:       0,
				CmDesiredSize: 3,
				DesiredSize:   30,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "test-cluster").Return(allReadyNodes, nil)
			},
			want: 96,
		},
		{
			name: "ng-limit-max",
			limit: &storage.NodegroupLimit{
				NodegroupLimit:    true,
				NodegroupLimitNum: 0,
				ClusterLimit:      false,
				ClusterLimitNum:   0,
			},
			ng: &storage.NodeGroup{
				NodeGroupID:   "test-ng",
				ClusterID:     "test-cluster",
				MaxSize:       100,
				MinSize:       0,
				CmDesiredSize: 3,
				DesiredSize:   30,
			},
			on: func(f *MockFields) {
				//f.clusterClient.On("ListClusterNodes", "test-cluster").Return(allReadyNodes, nil)
			},
			want: 97,
		},
		{
			name: "ng-limit-200",
			limit: &storage.NodegroupLimit{
				NodegroupLimit:    true,
				NodegroupLimitNum: 200,
				ClusterLimit:      false,
				ClusterLimitNum:   0,
			},
			ng: &storage.NodeGroup{
				NodeGroupID:   "test-ng",
				ClusterID:     "test-cluster",
				MaxSize:       100,
				MinSize:       0,
				CmDesiredSize: 3,
				DesiredSize:   30,
			},
			on: func(f *MockFields) {
				//f.clusterClient.On("ListClusterNodes", "test-cluster").Return(allReadyNodes, nil)
			},
			want: 197,
		},
		{
			name: "cluster-100-ng-200",
			limit: &storage.NodegroupLimit{
				NodegroupLimit:    true,
				NodegroupLimitNum: 200,
				ClusterLimit:      true,
				ClusterLimitNum:   100,
			},
			ng: &storage.NodeGroup{
				NodeGroupID:   "test-ng",
				ClusterID:     "test-cluster",
				MaxSize:       100,
				MinSize:       0,
				CmDesiredSize: 3,
				DesiredSize:   30,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "test-cluster").Return(allReadyNodes, nil)
			},
			want: 96,
		},
		{
			name: "cluster-200-ng-100",
			limit: &storage.NodegroupLimit{
				NodegroupLimit:    true,
				NodegroupLimitNum: 100,
				ClusterLimit:      true,
				ClusterLimitNum:   200,
			},
			ng: &storage.NodeGroup{
				NodeGroupID:   "test-ng",
				ClusterID:     "test-cluster",
				MaxSize:       100,
				MinSize:       0,
				CmDesiredSize: 3,
				DesiredSize:   30,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "test-cluster").Return(allReadyNodes, nil)
			},
			want: 97,
		},
		{
			name: "cluster-200-ng-100",
			limit: &storage.NodegroupLimit{
				NodegroupLimit:    true,
				NodegroupLimitNum: 100,
				ClusterLimit:      true,
				ClusterLimitNum:   200,
			},
			ng: &storage.NodeGroup{
				NodeGroupID:   "test-ng",
				ClusterID:     "test-cluster",
				MaxSize:       100,
				MinSize:       0,
				CmDesiredSize: 3,
				DesiredSize:   30,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "test-cluster").Return(allReadyNodes, nil)
			},
			want: 97,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				clusterClient: mocks.NewClient(t),
			}
			tt.on(mockFields)
			num := getNodegroupLimitCount(tt.limit, mockFields.clusterClient, tt.ng)
			assert.Equal(t, tt.want, num)
		})
	}
}

func Test_getLimitWithBuffer(t *testing.T) {
	test := []struct {
		name        string
		buffer      *storage.NodegroupBuffer
		originLimit int
		deviceTotal int
		want        int
		desireSize  int
	}{
		{
			name: "bufferLessThanLimit",
			buffer: &storage.NodegroupBuffer{
				Percent: 10,
				Count:   50,
			},
			originLimit: 50,
			deviceTotal: 100,
			desireSize:  5,
			want:        45,
		},
		{
			name: "bufferLargerThanLimit",
			buffer: &storage.NodegroupBuffer{
				Percent: 10,
				Count:   50,
			},
			originLimit: 30,
			deviceTotal: 100,
			desireSize:  5,
			want:        30,
		},
		{
			name: "bufferLessThanDesire",
			buffer: &storage.NodegroupBuffer{
				Percent: 10,
				Count:   50,
			},
			originLimit: 30,
			deviceTotal: 100,
			desireSize:  60,
			want:        0,
		},
		{
			name: "percent",
			buffer: &storage.NodegroupBuffer{
				Percent: 60,
				Count:   10,
			},
			originLimit: 30,
			deviceTotal: 100,
			desireSize:  5,
			want:        30,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			num := getLimitWithBuffer(tt.buffer, tt.originLimit, tt.deviceTotal, tt.desireSize)
			assert.Equal(t, tt.want, num)
		})
	}
}

func Test_newLimitBalancer(t *testing.T) {
	allReadyNodes := []*cluster.Node{{
		Name: "test1",
		IP:   "test1",
	}, {
		Name: "test2",
		IP:   "test2",
	}, {
		Name: "test3",
		IP:   "test3",
	},
		{
			Name: "test4",
			IP:   "test4",
		}}
	test := []struct {
		name        string
		nodeGroups  map[string]*storage.NodeGroup
		buffer      map[string]*storage.NodegroupBuffer
		groupInfo   []*storage.GroupInfo
		deviceTotal int
		on          func(f *MockFields)
		want        *limitBalancer
	}{
		{
			name: "normal",
			nodeGroups: map[string]*storage.NodeGroup{
				"ng1": {
					NodeGroupID:   "ng1",
					ClusterID:     "cluster1",
					MaxSize:       100,
					MinSize:       0,
					CmDesiredSize: 5,
				},
				"ng2": {
					NodeGroupID:   "ng2",
					ClusterID:     "cluster2",
					MaxSize:       10,
					MinSize:       0,
					CmDesiredSize: 5,
				},
			},
			buffer: map[string]*storage.NodegroupBuffer{
				"ng1": {
					Percent: 0,
					Count:   50,
				},
				"ng2": {
					Percent: 20,
					Count:   0,
				},
			},
			groupInfo: []*storage.GroupInfo{
				{
					NodeGroupID: "ng1",
					ConsumerID:  "consumer1",
					ClusterID:   "cluster1",
					Weight:      1,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    false,
						NodegroupLimitNum: 0,
						ClusterLimit:      true,
						ClusterLimitNum:   100,
					},
				},
				{
					NodeGroupID: "ng2",
					ConsumerID:  "consumer2",
					ClusterID:   "cluster2",
					Weight:      1,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    true,
						NodegroupLimitNum: 0,
						ClusterLimit:      true,
						ClusterLimitNum:   100,
					},
				},
			},
			deviceTotal: 100,
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "cluster1").Return(allReadyNodes, nil)
				f.clusterClient.On("ListClusterNodes", "cluster2").Return(allReadyNodes, nil)
			},
			want: &limitBalancer{
				nodeGroups: []*nodeGroup{{
					GroupInfo: storage.GroupInfo{
						NodeGroupID: "ng1",
						ConsumerID:  "consumer1",
						ClusterID:   "cluster1",
						Weight:      1,
						Limit: &storage.NodegroupLimit{
							NodegroupLimit:    false,
							NodegroupLimitNum: 0,
							ClusterLimit:      true,
							ClusterLimitNum:   100,
						}},
					partition:  0,
					limitation: 45,
				},
					{
						GroupInfo: storage.GroupInfo{
							NodeGroupID: "ng2",
							ConsumerID:  "consumer2",
							ClusterID:   "cluster2",
							Weight:      1,
							Limit: &storage.NodegroupLimit{
								NodegroupLimit:    true,
								NodegroupLimitNum: 0,
								ClusterLimit:      true,
								ClusterLimitNum:   100,
							}},
						partition:  0,
						limitation: 5,
					},
				},
				totalWeight: 2,
			},
		},
		{
			name: "normal2",
			nodeGroups: map[string]*storage.NodeGroup{
				"ng1": {
					NodeGroupID:   "ng1",
					ClusterID:     "cluster1",
					MaxSize:       50,
					MinSize:       0,
					CmDesiredSize: 5,
				},
				"ng2": {
					NodeGroupID:   "ng2",
					ClusterID:     "cluster2",
					MaxSize:       10,
					MinSize:       0,
					CmDesiredSize: 5,
				},
			},
			buffer: map[string]*storage.NodegroupBuffer{
				"ng1": {
					Percent: 0,
					Count:   60,
				},
				"ng2": {
					Percent: 20,
					Count:   0,
				},
			},
			groupInfo: []*storage.GroupInfo{
				{
					NodeGroupID: "ng1",
					ConsumerID:  "consumer1",
					ClusterID:   "cluster1",
					Weight:      1,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    true,
						NodegroupLimitNum: 0,
						ClusterLimit:      true,
						ClusterLimitNum:   100,
					},
				},
				{
					NodeGroupID: "ng2",
					ConsumerID:  "consumer2",
					ClusterID:   "cluster2",
					Weight:      1,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    true,
						NodegroupLimitNum: 0,
						ClusterLimit:      true,
						ClusterLimitNum:   100,
					},
				},
			},
			deviceTotal: 100,
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "cluster1").Return(allReadyNodes, nil)
				f.clusterClient.On("ListClusterNodes", "cluster2").Return(allReadyNodes, nil)
			},
			want: &limitBalancer{
				nodeGroups: []*nodeGroup{{
					GroupInfo: storage.GroupInfo{
						NodeGroupID: "ng1",
						ConsumerID:  "consumer1",
						ClusterID:   "cluster1",
						Weight:      1,
						Limit: &storage.NodegroupLimit{
							NodegroupLimit:    true,
							NodegroupLimitNum: 0,
							ClusterLimit:      true,
							ClusterLimitNum:   100,
						}},
					partition:  0,
					limitation: 45,
				},
					{
						GroupInfo: storage.GroupInfo{
							NodeGroupID: "ng2",
							ConsumerID:  "consumer2",
							ClusterID:   "cluster2",
							Weight:      1,
							Limit: &storage.NodegroupLimit{
								NodegroupLimit:    true,
								NodegroupLimitNum: 0,
								ClusterLimit:      true,
								ClusterLimitNum:   100,
							}},
						partition:  0,
						limitation: 5,
					},
				},
				totalWeight: 2,
			},
		},
		{
			name: "normal3",
			nodeGroups: map[string]*storage.NodeGroup{
				"ng1": {
					NodeGroupID:   "ng1",
					ClusterID:     "cluster1",
					MaxSize:       50,
					MinSize:       0,
					CmDesiredSize: 5,
				},
				"ng2": {
					NodeGroupID:   "ng2",
					ClusterID:     "cluster2",
					MaxSize:       10,
					MinSize:       0,
					CmDesiredSize: 10,
				},
			},
			buffer: map[string]*storage.NodegroupBuffer{
				"ng1": {
					Percent: 0,
					Count:   60,
				},
				"ng2": {
					Percent: 20,
					Count:   0,
				},
			},
			groupInfo: []*storage.GroupInfo{
				{
					NodeGroupID: "ng1",
					ConsumerID:  "consumer1",
					ClusterID:   "cluster1",
					Weight:      1,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    true,
						NodegroupLimitNum: 0,
						ClusterLimit:      true,
						ClusterLimitNum:   4,
					},
				},
				{
					NodeGroupID: "ng2",
					ConsumerID:  "consumer2",
					ClusterID:   "cluster2",
					Weight:      1,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    true,
						NodegroupLimitNum: 0,
						ClusterLimit:      true,
						ClusterLimitNum:   100,
					},
				},
			},
			deviceTotal: 100,
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "cluster1").Return(allReadyNodes, nil)
				f.clusterClient.On("ListClusterNodes", "cluster2").Return(allReadyNodes, nil)
			},
			want: &limitBalancer{
				nodeGroups: []*nodeGroup{{
					GroupInfo: storage.GroupInfo{
						NodeGroupID: "ng1",
						ConsumerID:  "consumer1",
						ClusterID:   "cluster1",
						Weight:      1,
						Limit: &storage.NodegroupLimit{
							NodegroupLimit:    true,
							NodegroupLimitNum: 0,
							ClusterLimit:      true,
							ClusterLimitNum:   4,
						}},
					partition:  0,
					limitation: 0,
				},
					{
						GroupInfo: storage.GroupInfo{
							NodeGroupID: "ng2",
							ConsumerID:  "consumer2",
							ClusterID:   "cluster2",
							Weight:      1,
							Limit: &storage.NodegroupLimit{
								NodegroupLimit:    true,
								NodegroupLimitNum: 0,
								ClusterLimit:      true,
								ClusterLimitNum:   100,
							}},
						partition:  0,
						limitation: 0,
					},
				},
				totalWeight: 2,
			},
		},
		{
			name: "normal4",
			nodeGroups: map[string]*storage.NodeGroup{
				"ng1": {
					NodeGroupID:   "ng1",
					ClusterID:     "cluster1",
					MaxSize:       50,
					MinSize:       0,
					CmDesiredSize: 0,
				},
			},
			buffer: map[string]*storage.NodegroupBuffer{},
			groupInfo: []*storage.GroupInfo{
				{
					NodeGroupID: "ng1",
					ConsumerID:  "consumer1",
					ClusterID:   "cluster1",
					Weight:      0,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    true,
						NodegroupLimitNum: 10,
						ClusterLimit:      false,
						ClusterLimitNum:   0,
					},
				},
			},
			deviceTotal: 12,
			on:          func(f *MockFields) {},
			want: &limitBalancer{
				nodeGroups: []*nodeGroup{{
					GroupInfo: storage.GroupInfo{
						NodeGroupID: "ng1",
						ConsumerID:  "consumer1",
						ClusterID:   "cluster1",
						Weight:      1,
						Limit: &storage.NodegroupLimit{
							NodegroupLimit:    true,
							NodegroupLimitNum: 10,
							ClusterLimit:      false,
							ClusterLimitNum:   0,
						}},
					partition:  0,
					limitation: 10,
				}},
				totalWeight: 1,
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				clusterClient: mocks.NewClient(t),
			}
			tt.on(mockFields)
			balancer := newLimitBalancer(tt.nodeGroups, tt.buffer, tt.groupInfo, tt.deviceTotal, mockFields.clusterClient)
			assert.Equal(t, reflect.DeepEqual(tt.want, balancer), true)
		})
	}
}

func Test_limitDistribute(t *testing.T) {
	test := []struct {
		name        string
		nodeGroups  map[string]*storage.NodeGroup
		buffer      map[string]*storage.NodegroupBuffer
		groupInfo   []*storage.GroupInfo
		deviceTotal int
		on          func(f *MockFields)
		want        *limitBalancer
	}{
		{
			name: "normal4",
			nodeGroups: map[string]*storage.NodeGroup{
				"ng1": {
					NodeGroupID:   "ng1",
					ClusterID:     "cluster1",
					MaxSize:       50,
					MinSize:       0,
					CmDesiredSize: 0,
				},
			},
			buffer: map[string]*storage.NodegroupBuffer{},
			groupInfo: []*storage.GroupInfo{
				{
					NodeGroupID: "ng1",
					ConsumerID:  "consumer1",
					ClusterID:   "cluster1",
					Weight:      1,
					Limit: &storage.NodegroupLimit{
						NodegroupLimit:    true,
						NodegroupLimitNum: 10,
						ClusterLimit:      false,
						ClusterLimitNum:   0,
					},
				},
			},
			deviceTotal: 12,
			on:          func(f *MockFields) {},
			want: &limitBalancer{
				nodeGroups: []*nodeGroup{{
					GroupInfo: storage.GroupInfo{
						NodeGroupID: "ng1",
						ConsumerID:  "consumer1",
						ClusterID:   "cluster1",
						Weight:      1,
						Limit: &storage.NodegroupLimit{
							NodegroupLimit:    true,
							NodegroupLimitNum: 10,
							ClusterLimit:      false,
							ClusterLimitNum:   0,
						}},
					partition:  10,
					limitation: 10,
				}},
				totalWeight: 1,
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				clusterClient: mocks.NewClient(t),
			}
			tt.on(mockFields)
			balancer := newLimitBalancer(tt.nodeGroups, tt.buffer, tt.groupInfo, tt.deviceTotal, mockFields.clusterClient)
			balancer.distribute(tt.deviceTotal)
			assert.Equal(t, reflect.DeepEqual(tt.want, balancer), true)
		})
	}
}
