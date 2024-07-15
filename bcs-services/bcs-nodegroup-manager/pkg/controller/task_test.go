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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/mocks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	storagemock "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mocks"
)

func Test_checkScaleDownComplete(t *testing.T) {
	tests := []struct {
		name    string
		origin  []string
		compare []string
		want    bool
	}{
		{
			name:    "false",
			origin:  []string{"test1", "0.0.0.0"},
			compare: []string{"0.0.0.0"},
			want:    false,
		},
		{
			name:    "true",
			origin:  []string{"test1"},
			compare: []string{"0.0.0.0"},
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp := checkScaleDownComplete(tt.origin, tt.compare)
			assert.Equal(t, tt.want, rsp)
		})
	}
}

//
//func Test_filterAvailableNodes(t *testing.T) {
//	strategy := getTestStrategy()
//	nodes1 := map[string]*cluster.Node{"test1": {
//		Name:   "test1",
//		IP:     "test1",
//		Status: string(v1.ConditionTrue),
//		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup1", storage.NodeDrainTaskLabel: "task1", storage.NodeDrainDelayLabel: "72h"},
//	}}
//	nodegroup1 := &storage.NodeGroup{
//		NodeGroupID: "NodeGroup1",
//		ClusterID:   "Cluster1",
//		MaxSize:     0,
//		MinSize:     0,
//		NodeIPs:     []string{"test1"},
//	}
//	nodes2 := map[string]*cluster.Node{"test2": {
//		Name:   "test2",
//		IP:     "test2",
//		Status: string(v1.ConditionTrue),
//		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup2", storage.NodeDrainDelayLabel: "72h"},
//	}}
//	nodegroup2 := &storage.NodeGroup{
//		NodeGroupID: "NodeGroup2",
//		ClusterID:   "Cluster1",
//		MaxSize:     0,
//		MinSize:     0,
//		NodeIPs:     []string{"test2"},
//	}
//	nodes3 := map[string]*cluster.Node{"test3": {
//		Name:   "test3",
//		IP:     "test3",
//		Status: string(v1.ConditionTrue),
//		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup3", storage.NodeDrainDelayLabel: "72h"},
//	},
//		"test4": {
//			Name:   "test4",
//			IP:     "test4",
//			Status: string(v1.ConditionTrue),
//			Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup3", storage.NodeDrainDelayLabel: "96h"},
//		},
//		"test5": {
//			Name:   "test5",
//			IP:     "test5",
//			Status: string(v1.ConditionTrue),
//			Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup3", storage.NodeDrainDelayLabel: "48h"},
//		}}
//	nodes4 := map[string]*cluster.Node{"test6": {
//		Name:   "test6",
//		IP:     "test6",
//		Status: string(v1.ConditionTrue),
//		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup4", storage.NodeDrainDelayLabel: "72h"},
//	},
//		"test7": {
//			Name:   "test7",
//			IP:     "test7",
//			Status: string(v1.ConditionTrue),
//			Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup4", storage.NodeDrainDelayLabel: "96h"},
//		},
//		"test8": {
//			Name:   "test8",
//			IP:     "test8",
//			Status: string(v1.ConditionTrue),
//			Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup4", storage.NodeDrainDelayLabel: "48h"},
//		},
//		"test9": {
//			Name:   "test9",
//			IP:     "test9",
//			Status: string(v1.ConditionTrue),
//			Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup4", storage.NodeDrainDelayLabel: "24h"},
//		}}
//
//	strategy.ElasticNodeGroups = append(strategy.ElasticNodeGroups,
//		&storage.GroupInfo{ClusterID: "Cluster1", NodeGroupID: "NodeGroup3", Weight: 5, ConsumerID: "consumer3"},
//		&storage.GroupInfo{ClusterID: "Cluster1", NodeGroupID: "NodeGroup4", Weight: 5, ConsumerID: "consumer4"})
//	test := []struct {
//		name              string
//		strategy          *storage.NodeGroupMgrStrategy
//		taskID            string
//		drainDelay        string
//		wantScaleDownInfo []*storage.ScaleDownNodegroup
//		wantNum           int
//		wantErr           bool
//		on                func(f *MockFields)
//	}{
//		{
//			name:       "normal",
//			strategy:   strategy,
//			taskID:     "task1",
//			drainDelay: "72h",
//			wantErr:    false,
//			wantNum:    2,
//			wantScaleDownInfo: []*storage.ScaleDownNodegroup{{
//				DrainDelayHour: 72,
//				Total:          2,
//				GroupInfos:     strategy.ElasticNodeGroups[0:2],
//				NodeGroups:     map[string]*storage.NodeGroup{"NodeGroup1": nodegroup1, "NodeGroup2": nodegroup2},
//			}},
//			on: func(f *MockFields) {
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
//					Return(nodes1, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
//					Return(nodes2, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup3"}).
//					Return(nil, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup4"}).
//					Return(nil, nil)
//			},
//		},
//		{
//			name:       "notMatchTaskID",
//			strategy:   strategy,
//			taskID:     "task2",
//			drainDelay: "72h",
//			wantErr:    false,
//			wantNum:    1,
//			wantScaleDownInfo: []*storage.ScaleDownNodegroup{{
//				DrainDelayHour: 72,
//				Total:          1,
//				GroupInfos:     []*storage.GroupInfo{strategy.ElasticNodeGroups[1]},
//				NodeGroups:     map[string]*storage.NodeGroup{"NodeGroup2": nodegroup2},
//			}},
//			on: func(f *MockFields) {
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
//					Return(nodes1, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
//					Return(nodes2, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup3"}).
//					Return(nil, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup4"}).
//					Return(nil, nil)
//			},
//		},
//		{
//			name:       "testBackup",
//			strategy:   strategy,
//			taskID:     "task1",
//			drainDelay: "72h",
//			wantErr:    false,
//			wantNum:    1,
//			wantScaleDownInfo: []*storage.ScaleDownNodegroup{{
//				DrainDelayHour: 72,
//				Total:          3,
//				GroupInfos:     strategy.ElasticNodeGroups[0:3],
//				NodeGroups: map[string]*storage.NodeGroup{"NodeGroup1": nodegroup1, "NodeGroup2": nodegroup2, "NodeGroup3": {
//					NodeGroupID: "NodeGroup3",
//					ClusterID:   "Cluster1",
//					MaxSize:     0,
//					MinSize:     0,
//					NodeIPs:     []string{"test3"}}},
//			}, {
//				DrainDelayHour: 48,
//				Total:          1,
//				GroupInfos:     []*storage.GroupInfo{strategy.ElasticNodeGroups[2]},
//				NodeGroups: map[string]*storage.NodeGroup{"NodeGroup3": {
//					NodeGroupID: "NodeGroup3",
//					ClusterID:   "Cluster1",
//					MaxSize:     0,
//					MinSize:     0,
//					NodeIPs:     []string{"test5"}}},
//			}},
//			on: func(f *MockFields) {
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
//					Return(nodes1, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
//					Return(nodes2, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup3"}).
//					Return(nodes3, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup4"}).
//					Return(nil, nil)
//			},
//		},
//		{
//			name:       "testBackup-complicated",
//			strategy:   strategy,
//			taskID:     "task1",
//			drainDelay: "72h",
//			wantErr:    false,
//			wantNum:    1,
//			wantScaleDownInfo: []*storage.ScaleDownNodegroup{{
//				DrainDelayHour: 72,
//				Total:          4,
//				GroupInfos:     strategy.ElasticNodeGroups,
//				NodeGroups: map[string]*storage.NodeGroup{"NodeGroup1": nodegroup1, "NodeGroup2": nodegroup2, "NodeGroup3": {
//					NodeGroupID: "NodeGroup3",
//					ClusterID:   "Cluster1",
//					MaxSize:     0,
//					MinSize:     0,
//					NodeIPs:     []string{"test3"}},
//					"NodeGroup4": {
//						NodeGroupID: "NodeGroup4",
//						ClusterID:   "Cluster1",
//						MaxSize:     0,
//						MinSize:     0,
//						NodeIPs:     []string{"test6"}}},
//			}, {
//				DrainDelayHour: 48,
//				Total:          2,
//				GroupInfos:     []*storage.GroupInfo{strategy.ElasticNodeGroups[2], strategy.ElasticNodeGroups[3]},
//				NodeGroups: map[string]*storage.NodeGroup{"NodeGroup3": {
//					NodeGroupID: "NodeGroup3",
//					ClusterID:   "Cluster1",
//					MaxSize:     0,
//					MinSize:     0,
//					NodeIPs:     []string{"test5"}},
//					"NodeGroup4": {
//						NodeGroupID: "NodeGroup4",
//						ClusterID:   "Cluster1",
//						MaxSize:     0,
//						MinSize:     0,
//						NodeIPs:     []string{"test8"}},
//				},
//			}, {
//				DrainDelayHour: 24,
//				Total:          1,
//				GroupInfos:     []*storage.GroupInfo{strategy.ElasticNodeGroups[3]},
//				NodeGroups: map[string]*storage.NodeGroup{
//					"NodeGroup4": {
//						NodeGroupID: "NodeGroup4",
//						ClusterID:   "Cluster1",
//						MaxSize:     0,
//						MinSize:     0,
//						NodeIPs:     []string{"test9"}},
//				},
//			}},
//			on: func(f *MockFields) {
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
//					Return(nodes1, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
//					Return(nodes2, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup3"}).
//					Return(nodes3, nil)
//				f.clusterClient.On("ListNodesByLabel", "Cluster1",
//					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup4"}).
//					Return(nodes4, nil)
//			},
//		},
//	}
//	for _, tt := range test {
//		t.Run(tt.name, func(t *testing.T) {
//			mockFields := &MockFields{
//				clusterClient: mocks.NewClient(t),
//			}
//			tt.on(mockFields)
//			opts := &Options{
//				ClusterClient: mockFields.clusterClient,
//			}
//			controller := &taskController{opt: opts}
//			selectedGroups, err := controller.filterAvailableNodes(tt.taskID, tt.drainDelay, tt.strategy)
//			assert.Equal(t, tt.wantErr, err != nil)
//			assert.Equal(t, reflect.DeepEqual(tt.wantScaleDownInfo, selectedGroups), true)
//		})
//	}
//}

// NOCC:golint/funlen(设计如此)
// nolint
func Test_removeNotReadyNodes(t *testing.T) {
	allReadyNodes := map[string]*cluster.Node{"test1": {
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: nil,
	}, "test2": {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionTrue),
		Labels: nil,
	}}
	notAllReadyNodes := map[string]*cluster.Node{"test1": {
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: nil,
	}, "test2": {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionFalse),
		Labels: nil,
	}}
	otherNodesWithLabel := map[string]*cluster.Node{"test1": {
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: nil,
	}, "test2": {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionTrue),
		Labels: nil,
	}, "test3": {
		Name:   "test3",
		IP:     "test3",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeDrainTaskLabel: "task1"},
	}}
	test := []struct {
		name            string
		scaleDownDetail *storage.ScaleDownDetail
		taskID          string
		want            *storage.ScaleDownDetail
		on              func(f *MockFields)
	}{
		{
			name: "allNodeReady",
			scaleDownDetail: &storage.ScaleDownDetail{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test2"},
				NodeNum:     2,
			},
			taskID: "task1",
			want: &storage.ScaleDownDetail{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test2"},
				NodeNum:     2,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).Return(allReadyNodes, nil)
			},
		},
		{
			name: "notAllReady",
			scaleDownDetail: &storage.ScaleDownDetail{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test2"},
				NodeNum:     2,
			},
			taskID: "task1",
			want: &storage.ScaleDownDetail{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1"},
				NodeNum:     1,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).Return(notAllReadyNodes, nil)
			},
		},
		{
			name: "removeNodeWithLabel",
			scaleDownDetail: &storage.ScaleDownDetail{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test2"},
				NodeNum:     2,
			},
			taskID: "task1",
			want: &storage.ScaleDownDetail{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test2"},
				NodeNum:     2,
			},
			on: func(f *MockFields) {
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).Return(otherNodesWithLabel, nil)
				f.clusterClient.On("UpdateNodeLabels", "Cluster1", "test3", mock.Anything).
					Return(nil)
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				clusterClient: mocks.NewClient(t),
			}
			tt.on(mockFields)
			opts := &Options{
				ClusterClient: mockFields.clusterClient,
			}
			controller := &taskController{opt: opts}
			controller.removeNotReadyNodes(tt.scaleDownDetail, tt.taskID)
			assert.Equal(t, reflect.DeepEqual(tt.want, tt.scaleDownDetail), true)
		})
	}
}

func Test_removeLabel(t *testing.T) {
	nodes := []*cluster.Node{{
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeDrainTaskLabel: "task1"},
	}, {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeDrainTaskLabel: "task2"},
	}}
	test := []struct {
		name      string
		clusterID string
		taskID    string
		wantErr   bool
		on        func(f *MockFields)
	}{
		{
			name:      "normal",
			clusterID: "Cluster1",
			taskID:    "task1",
			wantErr:   false,
			on: func(f *MockFields) {
				f.clusterClient.On("ListClusterNodes", "Cluster1").Return(nodes, nil)
				f.clusterClient.On("UpdateNodeLabels", "Cluster1", "test1", mock.Anything).
					Return(nil)
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				clusterClient: mocks.NewClient(t),
			}
			tt.on(mockFields)
			opts := &Options{
				ClusterClient: mockFields.clusterClient,
			}
			controller := &taskController{opt: opts}
			err := controller.removeLabel(tt.clusterID, tt.taskID)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

// NOCC:golint/funlen(设计如此)
// nolint
func Test_nodeSelector(t *testing.T) {
	nodes1 := map[string]*cluster.Node{"test1": {
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup1", storage.NodeDrainTaskLabel: "task1", storage.NodeDrainDelayLabel: "72h"},
	}, "test3": {
		Name:   "test3",
		IP:     "test3",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup1", storage.NodeDrainTaskLabel: "task1", storage.NodeDrainDelayLabel: "72h"},
	}}
	nodes2 := map[string]*cluster.Node{"test2": {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup2", storage.NodeDrainDelayLabel: "72h"},
	}, "test4": {
		Name:   "test4",
		IP:     "test4",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup2", storage.NodeDrainDelayLabel: "72h"},
	}}

	strategy2 := &storage.NodeGroupMgrStrategy{
		Name:              "test-strategy",
		ResourcePool:      "test-resource-pool",
		ReservedNodeGroup: &storage.GroupInfo{ClusterID: "reserved-clusterID", NodeGroupID: "reserved-nodeGroupID"},
		ElasticNodeGroups: []*storage.GroupInfo{
			{ClusterID: "Cluster1", NodeGroupID: "NodeGroup1", Weight: 5, ConsumerID: "consumer1"},
			{ClusterID: "Cluster1", NodeGroupID: "NodeGroup2", Weight: 5, ConsumerID: "consumer2"},
			{ClusterID: "Cluster1", NodeGroupID: "NodeGroup3", Weight: 5, ConsumerID: "consumer3"},
			{ClusterID: "Cluster1", NodeGroupID: "NodeGroup4", Weight: 5, ConsumerID: "consumer4"},
		},
		Strategy: &storage.Strategy{
			ScaleUpCoolDown: 0,
			ScaleUpDelay:    5,
			MinScaleUpSize:  2,
			ScaleDownDelay:  5,
			MaxIdleDelay:    1,
			Buffer:          &storage.BufferStrategy{Low: 10, High: 15},
		},
		Status: &storage.State{
			Status:      storage.InitState,
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	nodes3 := map[string]*cluster.Node{"test1": {
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup3", storage.NodeDrainDelayLabel: "72h"},
	}, "test2": {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup3", storage.NodeDrainDelayLabel: "48h"},
	}}
	nodes4 := map[string]*cluster.Node{"test1": {
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup4", storage.NodeDrainDelayLabel: "72h"},
	}, "test2": {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup4", storage.NodeDrainDelayLabel: "48h"},
	}, "test3": {
		Name:   "test3",
		IP:     "test3",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup4", storage.NodeDrainDelayLabel: "96h"},
	}}
	test := []struct {
		name    string
		task    *storage.ScaleDownTask
		wantErr bool
		want    []*storage.ScaleDownDetail
		on      func(f *MockFields)
	}{
		{
			name: "normal",
			task: &storage.ScaleDownTask{
				TaskID:            "task1",
				TotalNum:          2,
				NodeGroupStrategy: "test-strategy",
				ScaleDownGroups:   nil,
				DrainDelay:        "72h",
				Deadline:          time.Now().Add(72 * time.Hour),
				CreatedTime:       time.Now(),
				UpdatedTime:       time.Now(),
				IsDeleted:         false,
				IsExecuted:        false,
				Status:            "",
			},
			wantErr: false,
			want: []*storage.ScaleDownDetail{{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1"},
				NodeNum:     1,
			}, {
				ConsumerID:  "consumer2",
				NodeGroupID: "NodeGroup2",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test2"},
				NodeNum:     1,
			}},
			on: func(f *MockFields) {
				f.storage.On("GetNodeGroupStrategy", "test-strategy", &storage.GetOptions{}).
					Return(getTestStrategy(), nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
					Return(nodes1, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
					Return(nodes2, nil)
				f.storage.On("UpdateTask", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "notEqual",
			task: &storage.ScaleDownTask{
				TaskID:            "task1",
				TotalNum:          3,
				NodeGroupStrategy: "test-strategy",
				ScaleDownGroups:   nil,
				DrainDelay:        "72h",
				Deadline:          time.Now().Add(72 * time.Hour),
				CreatedTime:       time.Now(),
				UpdatedTime:       time.Now(),
				IsDeleted:         false,
				IsExecuted:        false,
				Status:            "",
			},
			wantErr: false,
			want: []*storage.ScaleDownDetail{{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test3"},
				NodeNum:     2,
			}, {
				ConsumerID:  "consumer2",
				NodeGroupID: "NodeGroup2",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test2"},
				NodeNum:     1,
			}},
			on: func(f *MockFields) {
				f.storage.On("GetNodeGroupStrategy", "test-strategy", &storage.GetOptions{}).
					Return(getTestStrategy(), nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
					Return(nodes1, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
					Return(nodes2, nil)
				f.storage.On("UpdateTask", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "complicated",
			task: &storage.ScaleDownTask{
				TaskID:            "task1",
				TotalNum:          6,
				NodeGroupStrategy: "test-strategy",
				ScaleDownGroups:   nil,
				DrainDelay:        "72h",
				Deadline:          time.Now().Add(72 * time.Hour),
				CreatedTime:       time.Now(),
				UpdatedTime:       time.Now(),
				IsDeleted:         false,
				IsExecuted:        false,
				Status:            "",
			},
			wantErr: false,
			want: []*storage.ScaleDownDetail{{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test3"},
				NodeNum:     2,
			}, {
				ConsumerID:  "consumer2",
				NodeGroupID: "NodeGroup2",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test2", "test4"},
				NodeNum:     2,
			}, {
				ConsumerID:  "consumer3",
				NodeGroupID: "NodeGroup3",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1"},
				NodeNum:     1,
			}, {
				ConsumerID:  "consumer4",
				NodeGroupID: "NodeGroup4",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1"},
				NodeNum:     1,
			}},
			on: func(f *MockFields) {
				f.storage.On("GetNodeGroupStrategy", "test-strategy", &storage.GetOptions{}).
					Return(strategy2, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
					Return(nodes1, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
					Return(nodes2, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup3"}).
					Return(nodes3, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup4"}).
					Return(nodes4, nil)
				f.storage.On("UpdateTask", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "complicated2",
			task: &storage.ScaleDownTask{
				TaskID:            "task1",
				TotalNum:          8,
				NodeGroupStrategy: "test-strategy",
				ScaleDownGroups:   nil,
				DrainDelay:        "72h",
				Deadline:          time.Now().Add(72 * time.Hour),
				CreatedTime:       time.Now(),
				UpdatedTime:       time.Now(),
				IsDeleted:         false,
				IsExecuted:        false,
				Status:            "",
			},
			wantErr: false,
			want: []*storage.ScaleDownDetail{{
				ConsumerID:  "consumer1",
				NodeGroupID: "NodeGroup1",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test3"},
				NodeNum:     2,
			}, {
				ConsumerID:  "consumer2",
				NodeGroupID: "NodeGroup2",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test2", "test4"},
				NodeNum:     2,
			}, {
				ConsumerID:  "consumer3",
				NodeGroupID: "NodeGroup3",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test2"},
				NodeNum:     2,
			}, {
				ConsumerID:  "consumer4",
				NodeGroupID: "NodeGroup4",
				ClusterID:   "Cluster1",
				NodeIPs:     []string{"test1", "test2"},
				NodeNum:     2,
			}},
			on: func(f *MockFields) {
				f.storage.On("GetNodeGroupStrategy", "test-strategy", &storage.GetOptions{}).
					Return(strategy2, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
					Return(nodes1, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
					Return(nodes2, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup3"}).
					Return(nodes3, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup4"}).
					Return(nodes4, nil)
				f.storage.On("UpdateTask", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				clusterClient: mocks.NewClient(t),
				storage:       storagemock.NewStorage(t),
			}
			tt.on(mockFields)
			opts := &Options{
				ClusterClient: mockFields.clusterClient,
				Storage:       mockFields.storage,
			}
			controller := &taskController{opt: opts}
			err := controller.nodeSelector(tt.task)
			assert.Equal(t, tt.wantErr, err != nil)
			total := 0
			for _, group := range tt.task.ScaleDownGroups {
				total += group.NodeNum
			}
			assert.Equal(t, reflect.DeepEqual(tt.want, tt.task.ScaleDownGroups) || total == tt.task.TotalNum, true)
		})
	}
}

func Test_handleOneNormalTask(t *testing.T) {
	nodes1 := map[string]*cluster.Node{"test1": {
		Name:   "test1",
		IP:     "test1",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup1", storage.NodeDrainTaskLabel: "task1"},
	}, "test3": {
		Name:   "test3",
		IP:     "test3",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup1", storage.NodeDrainTaskLabel: "task1"},
	}}
	nodes2 := map[string]*cluster.Node{"test2": {
		Name:   "test2",
		IP:     "test2",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup2"},
	}, "test4": {
		Name:   "test4",
		IP:     "test4",
		Status: string(v1.ConditionTrue),
		Labels: map[string]string{storage.NodeGroupLabel: "NodeGroup2"},
	}}
	task := getTestTask()
	labelMap := map[string]interface{}{
		storage.NodeDrainTaskLabel:  "task1",
		storage.NodeDrainDelayLabel: "48h",
	}
	annotationMap := map[string]interface{}{
		storage.NodeDeadlineLabel: task.Deadline.Format(time.RFC3339),
	}
	tests := []struct {
		name string
		on   func(f *MockFields)
		task *storage.ScaleDownTask
	}{
		{
			name: "normalSelect",
			task: task,
			on: func(f *MockFields) {
				f.storage.On("UpdateTask", mock.Anything, mock.Anything).Return(getTestTask(), nil)
				f.storage.On("GetNodeGroupStrategy", "test-strategy", &storage.GetOptions{}).
					Return(getTestStrategy(), nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup1"}).
					Return(nodes1, nil)
				f.clusterClient.On("ListNodesByLabel", "Cluster1",
					map[string]interface{}{storage.NodeGroupLabel: "NodeGroup2"}).
					Return(nodes2, nil)
				f.clusterClient.On("UpdateNodeMetadata", "Cluster1", mock.Anything, labelMap, annotationMap).
					Return(nil)
				f.clusterClient.On("UpdateNodeMetadata", "Cluster1", mock.Anything, labelMap, annotationMap).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFields := &MockFields{
				clusterClient: mocks.NewClient(t),
				storage:       storagemock.NewStorage(t),
			}
			tt.on(mockFields)
			opts := &Options{
				ClusterClient: mockFields.clusterClient,
				Storage:       mockFields.storage,
			}
			controller := &taskController{opt: opts}
			controller.handleOneNormalTask(tt.task)
		})
	}
}

func Test_checkExpiredTask(t *testing.T) {
	expiredTask := &storage.ScaleDownTask{
		TaskID:            "task2",
		TotalNum:          2,
		NodeGroupStrategy: "test-strategy",
		ScaleDownGroups:   nil,
		DrainDelay:        "48",
		Deadline:          time.Now().Add(-1 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDeleted:         false,
		IsExecuted:        false,
		Status:            "",
	}
	tasks := []*storage.ScaleDownTask{{
		TaskID:            "task1",
		TotalNum:          2,
		NodeGroupStrategy: "test-strategy",
		ScaleDownGroups:   nil,
		DrainDelay:        "48h",
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDeleted:         false,
		IsExecuted:        false,
		Status:            "",
	}, expiredTask}

	result := checkExpiredTask(tasks)
	assert.Equal(t, reflect.DeepEqual(result, []*storage.ScaleDownTask{expiredTask}), true)
}

type MockFields struct {
	clusterClient *mocks.Client
	storage       *storagemock.Storage
}

func getTestTask() *storage.ScaleDownTask {
	return &storage.ScaleDownTask{
		TaskID:            "task1",
		TotalNum:          2,
		NodeGroupStrategy: "test-strategy",
		ScaleDownGroups:   nil,
		DrainDelay:        "48",
		Deadline:          time.Now().Add(48 * time.Hour),
		CreatedTime:       time.Now(),
		UpdatedTime:       time.Now(),
		IsDeleted:         false,
		IsExecuted:        false,
		Status:            "",
	}
}
