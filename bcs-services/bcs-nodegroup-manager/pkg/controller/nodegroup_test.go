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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	basemock "github.com/stretchr/testify/mock"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	mockstorage "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mocks"
)

func TestUpComingElasticResources_Up(t *testing.T) {
	assertion := assert.New(t)
	// construct test data
	oldTime := time.Now().AddDate(0, 0, -1)
	actions := map[string]*storage.NodeGroupAction{
		"NodeGroup1": {NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", UpdatedTime: time.Now(),
			DeltaNum: 20, NewDesiredNum: 30, OriginalDesiredNum: 10, Event: storage.ScaleUpState},
		"NodeGroup2": {NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", UpdatedTime: oldTime,
			DeltaNum: 5, NewDesiredNum: 15, OriginalDesiredNum: 10, Event: storage.ScaleUpState},
		"NodeGroup3": {NodeGroupID: "NodeGroup3", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			DeltaNum: 5, NewDesiredNum: 5, OriginalDesiredNum: 0, Event: storage.ScaleUpState},
		"NodeGroup4": {NodeGroupID: "NodeGroup4", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			DeltaNum: 10, NewDesiredNum: 20, OriginalDesiredNum: 10, Event: storage.ScaleUpState},
	}
	nodeGroups := map[string]*storage.NodeGroup{
		"NodeGroup1": {NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 10, DesiredSize: 30, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
		"NodeGroup2": {NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 10, DesiredSize: 15, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
		"NodeGroup3": {NodeGroupID: "NodeGroup3", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 0, DesiredSize: 5, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
		"NodeGroup4": {NodeGroupID: "NodeGroup4", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 10, DesiredSize: 20, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
	}
	strategy := &storage.NodeGroupMgrStrategy{
		Strategy: &storage.Strategy{ScaleUpDelay: 10, ScaleDownDelay: 10},
	}
	coming := upComingElasticResources(actions, nodeGroups, storage.ScaleUpState, strategy)
	// only NodeGroup1 & NodeGroup4 can scaleUp, total 30
	assertion.Equal(30, coming)
}

func TestUpComingElasticResources_Down(t *testing.T) {
	assertion := assert.New(t)
	// construct test data
	oldTime := time.Now().AddDate(0, 0, -1)
	actions := map[string]*storage.NodeGroupAction{
		"NodeGroup1": {NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", UpdatedTime: time.Now(),
			DeltaNum: 20, NewDesiredNum: 5, OriginalDesiredNum: 10, Event: storage.ScaleDownState},
		"NodeGroup2": {NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", UpdatedTime: oldTime,
			DeltaNum: 5, NewDesiredNum: 5, OriginalDesiredNum: 10, Event: storage.ScaleDownState},
		"NodeGroup3": {NodeGroupID: "NodeGroup3", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			DeltaNum: 5, NewDesiredNum: 12, OriginalDesiredNum: 10, Event: storage.ScaleDownState},
		"NodeGroup4": {NodeGroupID: "NodeGroup4", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			DeltaNum: 10, NewDesiredNum: 5, OriginalDesiredNum: 10, Event: storage.ScaleDownState},
	}
	nodeGroups := map[string]*storage.NodeGroup{
		"NodeGroup1": {NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 0, DesiredSize: 5, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
		"NodeGroup2": {NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 0, DesiredSize: 10, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
		"NodeGroup3": {NodeGroupID: "NodeGroup3", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 0, DesiredSize: 12, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
		"NodeGroup4": {NodeGroupID: "NodeGroup4", ClusterID: "ClusterID2", UpdatedTime: time.Now(),
			MaxSize: 100, MinSize: 0, DesiredSize: 5, CmDesiredSize: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"}},
	}
	strategy := &storage.NodeGroupMgrStrategy{
		Strategy: &storage.Strategy{ScaleUpDelay: 10, ScaleDownDelay: 10},
	}
	coming := upComingElasticResources(actions, nodeGroups, storage.ScaleDownState, strategy)
	// only NodeGroup1 & NodeGroup4 can ScaleDown, total 10
	assertion.Equal(10, coming)
}

func TestScaleUp_listNodeGroupError(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleUpNum := 10
	// listElasticNodeGroups failure cases:
	// 1. Storage.GetNodeGroup failed
	// 2. get empty NodeGroup
	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).
		Return(nil, fmt.Errorf("database connection lost"))
	err := ctl.handleElasticNodeGroupScaleUp(strategy, scaleUpNum, 100)
	assertion.NotNil(err, "GetNodeGroup met any failure must return error")
	store.AssertExpectations(t)

	store.ExpectedCalls = nil
	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nil, nil)
	err = ctl.handleElasticNodeGroupScaleUp(strategy, scaleUpNum, 100)
	assertion.NotNil(err, "empty nodegroup response must error")
	store.AssertExpectations(t)
}

func TestScaleUp_updateNodeGroupErr(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleUpNum := 10
	// updateNodeGroup failure cases:
	// storage return failure
	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[0], nil)
	store.On("GetNodeGroup", nodeGroups[1].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[1], nil)
	store.On("UpdateNodeGroup",
		nodeGroups[0], &storage.UpdateOptions{}).
		Return(nodeGroups[0], fmt.Errorf("data storage failure"))

	err := ctl.handleElasticNodeGroupScaleUp(strategy, scaleUpNum, 100)
	assertion.NotNil(err, "UpdateNodeGroup met any failure must stop testcase")
	store.AssertExpectations(t)
}

func TestScaleUp_createActionErr(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleUpNum := 10
	// createNodeGroupAction failure cases:
	// storage return failure
	expectedActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaleup",
			DeltaNum: 5, NewDesiredNum: 20, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaleup",
			DeltaNum: 5, NewDesiredNum: 15, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
	}

	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[0], nil)
	store.On("GetNodeGroup", nodeGroups[1].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[1], nil)
	store.On("UpdateNodeGroup", nodeGroups[0], &storage.UpdateOptions{}).Return(nodeGroups[0], nil)
	// store.On("UpdateNodeGroup", nodeGroups[1], &storage.UpdateOptions{}).Return(nodeGroups[1], nil)
	store.On("CreateNodeGroupAction",
		basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
			return action.NodeGroupID == expectedActions[0].NodeGroupID &&
				action.DeltaNum == expectedActions[0].DeltaNum &&
				action.NewDesiredNum == expectedActions[0].NewDesiredNum
		}),
		&storage.CreateOptions{},
	).Return(fmt.Errorf("database connection lost"))

	err := ctl.handleElasticNodeGroupScaleUp(strategy, scaleUpNum, 100)
	assertion.NotNil(err, "CreateNodeGroupAction met any failure must stop testcase")
	store.AssertExpectations(t)
}

func TestScaleUp_createEvent(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleUpNum := 10
	// createNodeGroupEvent failure cases: storage return failure
	// but we can tolerate event failure
	expectedActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaleup",
			DeltaNum: 5, NewDesiredNum: 20, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaleup",
			DeltaNum: 5, NewDesiredNum: 15, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
	}

	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[0], nil)
	store.On("GetNodeGroup", nodeGroups[1].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[1], nil)
	// !first loop, create Event failure but stil go on
	store.On("UpdateNodeGroup", nodeGroups[0], &storage.UpdateOptions{}).Return(nodeGroups[0], nil)
	store.On("CreateNodeGroupAction",
		basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
			return action.NodeGroupID == expectedActions[0].NodeGroupID &&
				action.DeltaNum == expectedActions[0].DeltaNum &&
				action.NewDesiredNum == expectedActions[0].NewDesiredNum
		}),
		&storage.CreateOptions{},
	).Return(nil)
	store.On(
		"CreateNodeGroupEvent",
		basemock.AnythingOfType("*storage.NodeGroupEvent"),
		&storage.CreateOptions{},
	).Return(fmt.Errorf("database connection lost"))
	// !second loop
	store.On("UpdateNodeGroup", nodeGroups[1], &storage.UpdateOptions{}).Return(nodeGroups[1], nil)
	store.On("CreateNodeGroupAction",
		basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
			return action.NodeGroupID == expectedActions[1].NodeGroupID &&
				action.DeltaNum == expectedActions[1].DeltaNum &&
				action.NewDesiredNum == expectedActions[1].NewDesiredNum
		}),
		&storage.CreateOptions{},
	).Return(nil)
	store.On(
		"CreateNodeGroupEvent",
		basemock.AnythingOfType("*storage.NodeGroupEvent"),
		&storage.CreateOptions{},
	).Return(nil)

	err := ctl.handleElasticNodeGroupScaleUp(strategy, scaleUpNum, 100)
	assertion.Nil(err, "CreateNodeGroupEvent met any failure are acceptable")
	store.AssertExpectations(t)
}

func TestScaleDown_listNodeGroupError(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleDownNum := 10
	// listElasticNodeGroups failure cases:
	// 1. Storage.GetNodeGroup failed
	// 2. get empty NodeGroup
	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).
		Return(nil, fmt.Errorf("database connection lost"))
	err := ctl.handleElasticNodeGroupScaleDown(strategy, scaleDownNum)
	assertion.NotNil(err, "GetNodeGroup met any failure must return error")
	store.AssertExpectations(t)

	store.ExpectedCalls = nil
	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nil, nil)
	err = ctl.handleElasticNodeGroupScaleDown(strategy, scaleDownNum)
	assertion.NotNil(err, "empty nodegroup response must error")
	store.AssertExpectations(t)
}

func TestScaleDown_updateNodeGroupErr(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data,
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleDownNum := 10
	// updateNodeGroup failure cases:
	// storage return failure
	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[0], nil)
	store.On("GetNodeGroup", nodeGroups[1].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[1], nil)
	store.On("UpdateNodeGroup", nodeGroups[0], &storage.UpdateOptions{}).
		Return(nodeGroups[0], fmt.Errorf("data storage failure"))

	err := ctl.handleElasticNodeGroupScaleUp(strategy, scaleDownNum, 100)
	assertion.NotNil(err, "UpdateNodeGroup met any failure must stop testcase")
	store.AssertExpectations(t)
}

func TestScaleDown_createActionErr(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleDownNum := 10
	// createNodeGroupAction failure cases:
	// storage return failure
	expectedActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaledown",
			DeltaNum: 5, NewDesiredNum: 10, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaledown",
			DeltaNum: 5, NewDesiredNum: 5, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
	}

	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[0], nil)
	store.On("GetNodeGroup", nodeGroups[1].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[1], nil)
	store.On("UpdateNodeGroup", nodeGroups[0], &storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}).
		Return(nodeGroups[0], nil)
	// store.On("UpdateNodeGroup", nodeGroups[1], &storage.UpdateOptions{}).Return(nodeGroups[1], nil)
	store.On("CreateNodeGroupAction",
		basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
			t.Logf("ScaleDown number by weight, DesiredSize: %d, DeltaNum: %d", action.NewDesiredNum, action.DeltaNum)
			return action.NodeGroupID == expectedActions[0].NodeGroupID &&
				action.Event == expectedActions[0].Event && action.DeltaNum >= 4
		}),
		&storage.CreateOptions{},
	).Return(fmt.Errorf("database connection lost"))

	err := ctl.handleElasticNodeGroupScaleDown(strategy, scaleDownNum)
	assertion.NotNil(err, "CreateNodeGroupAction met any failure must stop testcase")
	store.AssertExpectations(t)
}

func TestScaleDown_createNodeGroupEvent(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data, two
	strategy := getTestStrategy()
	nodeGroups := getStableNodeGroups()
	scaleDownNum := 10
	// createNodeGroupEvent failure cases: storage return failure
	// but we can tolerate event failure
	expectedActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaledown",
			DeltaNum: 5, NewDesiredNum: 10, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", CreatedTime: time.Now(), Event: "Scaledown",
			DeltaNum: 5, NewDesiredNum: 5, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: "InitState", UpdatedTime: time.Now(),
		},
	}

	store.On("GetNodeGroup", nodeGroups[0].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[0], nil)
	store.On("GetNodeGroup", nodeGroups[1].NodeGroupID, &storage.GetOptions{}).Return(nodeGroups[1], nil)
	// !first loop, create Event failure but still go on
	store.On("UpdateNodeGroup", nodeGroups[0], &storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}).
		Return(nodeGroups[0], nil)
	store.On("CreateNodeGroupAction",
		basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
			return action.NodeGroupID == expectedActions[0].NodeGroupID &&
				action.Event == expectedActions[0].Event && action.DeltaNum >= 4
		}),
		&storage.CreateOptions{},
	).Return(nil)
	store.On(
		"CreateNodeGroupEvent",
		basemock.AnythingOfType("*storage.NodeGroupEvent"),
		&storage.CreateOptions{},
	).Return(fmt.Errorf("database connection lost"))
	// !second loop
	store.On("UpdateNodeGroup", nodeGroups[1], &storage.UpdateOptions{OverwriteZeroOrEmptyStr: true}).
		Return(nodeGroups[1], nil)
	store.On("CreateNodeGroupAction",
		basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
			return action.NodeGroupID == expectedActions[1].NodeGroupID &&
				action.Event == expectedActions[1].Event && action.DeltaNum >= 4
		}),
		&storage.CreateOptions{},
	).Return(nil)
	store.On(
		"CreateNodeGroupEvent",
		basemock.AnythingOfType("*storage.NodeGroupEvent"),
		&storage.CreateOptions{},
	).Return(nil)

	err := ctl.handleElasticNodeGroupScaleDown(strategy, scaleDownNum)
	assertion.Nil(err, "CreateNodeGroupEvent met any failure are acceptable")
	store.AssertExpectations(t)
}

func TestTracingDown_cleanActionsErr(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data, two
	strategy := getTestStrategy()
	// nodeGroups := getScaleDownNodeGroups()
	scaleDownNum := 10
	scaleDownActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", Event: storage.ScaleUpState,
			DeltaNum: 5, NewDesiredNum: 20, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleUpState, UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", Event: storage.ScaleDownState,
			DeltaNum: 5, NewDesiredNum: 5, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleDownState, UpdatedTime: time.Now(),
		},
	}

	store.On("DeleteNodeGroupAction",
		scaleDownActions[0], &storage.DeleteOptions{},
	).Return(nil, fmt.Errorf("storage broken"))

	err := ctl.tracingScaleDownAction(strategy, scaleDownNum, scaleDownActions)
	assertion.Error(err, "testcase is setting error")
	store.AssertExpectations(t)
}

func TestTracingDown_releaseResEnough(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data, two
	strategy := getTestStrategy()
	nodeGroups := getScaleDownNodeGroups()
	scaleDownNum := 10
	scaleDownActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", Event: storage.ScaleDownState,
			DeltaNum: 5, NewDesiredNum: 10, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleDownState, UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", Event: storage.ScaleDownState,
			DeltaNum: 5, NewDesiredNum: 5, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleDownState, UpdatedTime: time.Now(),
		},
	}
	// get expected NodeGroup information
	for i := range nodeGroups {
		store.On("GetNodeGroup",
			nodeGroups[i].NodeGroupID, &storage.GetOptions{},
		).Return(nodeGroups[i], nil)
	}

	err := ctl.tracingScaleDownAction(strategy, scaleDownNum, scaleDownActions)
	assertion.Nil(err, fmt.Sprintf("expected scaledown %d nodes, but meet error", scaleDownNum))
	store.AssertExpectations(t)
}

func TestTracingDown_releaseResNotEnough(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data, two
	strategy := getTestStrategy()
	// only scaledown 10 nodes
	nodeGroups := getScaleDownNodeGroups()
	scaleDownNum := 15
	scaleDownActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", Event: storage.ScaleDownState,
			DeltaNum: 5, NewDesiredNum: 10, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleDownState, UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", Event: storage.ScaleDownState,
			DeltaNum: 5, NewDesiredNum: 5, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleDownState, UpdatedTime: time.Now(),
		},
	}
	// get expected NodeGroup information
	for i := range nodeGroups {
		store.On("GetNodeGroup",
			nodeGroups[i].NodeGroupID, &storage.GetOptions{},
		).Return(nodeGroups[i], nil)
	}
	// update nodegroup with newDesired size
	for i := 0; i < 2; i++ {
		oldSize := scaleDownActions[i].NewDesiredNum
		comName := scaleDownActions[i].NodeGroupID
		t.Logf("outer info: %s, old DesiredSize: %d", comName, oldSize)
		// new DesiredSize must be less than before
		store.On("UpdateNodeGroup",
			basemock.MatchedBy(func(group *storage.NodeGroup) bool {
				// keep scaledown, new desired size must less than before
				t.Logf("%s", group.Message)
				return comName == group.NodeGroupID &&
					group.DesiredSize < oldSize
			}),
			&storage.UpdateOptions{OverwriteZeroOrEmptyStr: true},
		).Return(nil, nil)

		// expected create NodeGroupAction
		store.On("CreateNodeGroupAction",
			basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
				return action.NodeGroupID == comName
			}),
			&storage.CreateOptions{OverWriteIfExist: true},
		).Return(nil, nil)

		// expected create NodeGroupEvent, no matter what error
		store.On("CreateNodeGroupEvent",
			basemock.MatchedBy(func(event *storage.NodeGroupEvent) bool {
				t.Logf("create new NodeGroup Event, %s, Desired, %d, message: %s",
					event.NodeGroupID, event.DesiredNum, event.Message)
				return true
			}),
			&storage.CreateOptions{},
		).Return(nil, fmt.Errorf("ignore error"))
	}
	err := ctl.tracingScaleDownAction(strategy, scaleDownNum, scaleDownActions)
	assertion.Nil(err, fmt.Sprintf("expected scaledown %d nodes, but meet error", scaleDownNum))
	store.AssertExpectations(t)
}

func TestTracingUp_scaleUpResEnough(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data
	strategy := getTestStrategy()
	nodeGroups := getScaleUpNodeGroups()
	scaleUpNum := 10
	scaleUpActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", Event: storage.ScaleUpState,
			DeltaNum: 5, NewDesiredNum: 20, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleUpState, UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", Event: storage.ScaleUpState,
			DeltaNum: 5, NewDesiredNum: 15, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleUpState, UpdatedTime: time.Now(),
		},
	}
	// get expected NodeGroup information
	for i := range nodeGroups {
		store.On("GetNodeGroup",
			nodeGroups[i].NodeGroupID, &storage.GetOptions{},
		).Return(nodeGroups[i], nil)
	}
	// scaleup resource is nought, wait next tick
	err := ctl.tracingScaleUpAction(strategy, scaleUpNum, scaleUpActions)
	assertion.Nil(err, fmt.Sprintf("expected scaledown %d nodes, but meet error", scaleUpNum))
	store.AssertExpectations(t)
}

func TestTracingUp_scaleUpResNotEnough(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data, two
	strategy := getTestStrategy()
	// only scaledown 10 nodes
	nodeGroups := getScaleUpNodeGroups()
	scaleUpNum := 15
	scaleUpActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", Event: storage.ScaleUpState,
			DeltaNum: 5, NewDesiredNum: 20, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleUpState, UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", Event: storage.ScaleUpState,
			DeltaNum: 5, NewDesiredNum: 15, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleUpState, UpdatedTime: time.Now(),
		},
	}
	// get expected NodeGroup information
	for i := range nodeGroups {
		store.On("GetNodeGroup",
			nodeGroups[i].NodeGroupID, &storage.GetOptions{},
		).Return(nodeGroups[i], nil)
	}
	// update nodegroup with newDesired size
	for i := 0; i < 2; i++ {
		oldSize := scaleUpActions[i].NewDesiredNum
		comName := scaleUpActions[i].NodeGroupID
		t.Logf("outer info: %s, old DesiredSize: %d", comName, oldSize)
		// new DesiredSize must be less than before
		store.On("UpdateNodeGroup",
			basemock.MatchedBy(func(group *storage.NodeGroup) bool {
				// keep scaleup, new desired size must less than before
				t.Logf("%s", group.Message)
				return comName == group.NodeGroupID &&
					group.DesiredSize > oldSize
			}),
			&storage.UpdateOptions{OverwriteZeroOrEmptyStr: true},
		).Return(nil, nil)

		// expected create NodeGroupAction
		store.On("CreateNodeGroupAction",
			basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
				return action.NodeGroupID == comName
			}),
			&storage.CreateOptions{OverWriteIfExist: true},
		).Return(nil, nil)

		// expected create NodeGroupEvent, no matter what error
		store.On("CreateNodeGroupEvent",
			basemock.MatchedBy(func(event *storage.NodeGroupEvent) bool {
				t.Logf("%s", event.Message)
				return true
			}),
			&storage.CreateOptions{},
		).Return(nil, fmt.Errorf("ignore error"))
	}
	err := ctl.tracingScaleUpAction(strategy, scaleUpNum, scaleUpActions)
	assertion.Nil(err, fmt.Sprintf("expected scaleUp %d nodes, but meet error", scaleUpNum))
	store.AssertExpectations(t)
}

func TestTracingUp_resNotEnoughDiffAct(t *testing.T) {
	assertion := assert.New(t)
	store := mockstorage.NewStorage(t)
	ctl := control{
		opt: &Options{Storage: store},
	}
	// init test data, two
	strategy := getTestStrategy()
	// only scaledown 10 nodes
	nodeGroups := getDiffNodeGroups()
	scaleUpNum := 15
	scaleUpActions := []*storage.NodeGroupAction{
		{
			NodeGroupID: "NodeGroup1", ClusterID: "Cluster1", Event: storage.ScaleDownState,
			DeltaNum: 5, NewDesiredNum: 10, OriginalDesiredNum: 15, OriginalNodeNum: 15,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleDownState, UpdatedTime: time.Now(),
		},
		{
			NodeGroupID: "NodeGroup2", ClusterID: "Cluster1", Event: storage.ScaleUpState,
			DeltaNum: 5, NewDesiredNum: 15, OriginalDesiredNum: 10, OriginalNodeNum: 10,
			NodeIPs: []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			Process: 0, Status: storage.ScaleUpState, UpdatedTime: time.Now(),
		},
	}
	// first delete scaleup action
	store.On("DeleteNodeGroupAction", scaleUpActions[0], &storage.DeleteOptions{}).Return(nil, nil)

	// get expected NodeGroup information
	for i := range nodeGroups {
		store.On("GetNodeGroup",
			nodeGroups[i].NodeGroupID, &storage.GetOptions{},
		).Return(nodeGroups[i], nil)
	}
	// update nodegroup with newDesired size
	for i := 0; i < 2; i++ {
		oldSize := nodeGroups[i].DesiredSize
		comName := nodeGroups[i].NodeGroupID
		t.Logf("outer info: %s, old DesiredSize: %d", comName, oldSize)
		// new DesiredSize must be less than before
		store.On("UpdateNodeGroup",
			basemock.MatchedBy(func(group *storage.NodeGroup) bool {
				// keep scaleup, new desired size must less than before
				t.Logf("#####new dicision: %+v", group)
				return comName == group.NodeGroupID &&
					group.DesiredSize > oldSize
			}),
			&storage.UpdateOptions{OverwriteZeroOrEmptyStr: true},
		).Return(nil, nil)

		// expected create NodeGroupAction
		store.On("CreateNodeGroupAction",
			basemock.MatchedBy(func(action *storage.NodeGroupAction) bool {
				return action.NodeGroupID == comName
			}),
			&storage.CreateOptions{OverWriteIfExist: true},
		).Return(nil, nil)

		// expected create NodeGroupEvent, no matter what error
		store.On("CreateNodeGroupEvent",
			basemock.MatchedBy(func(event *storage.NodeGroupEvent) bool {
				t.Logf("%s", event.Message)
				return true
			}),
			&storage.CreateOptions{},
		).Return(nil, fmt.Errorf("ignore error"))
	}
	err := ctl.tracingScaleUpAction(strategy, scaleUpNum, scaleUpActions)
	assertion.Nil(err, fmt.Sprintf("expected scaleUp %d nodes, but meet error", scaleUpNum))
	store.AssertExpectations(t)
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
	return strategy
}

func getStableNodeGroups() []*storage.NodeGroup {
	nodeGroups := []*storage.NodeGroup{
		{
			NodeGroupID:   "NodeGroup1",
			ClusterID:     "Cluster1",
			MaxSize:       100,
			MinSize:       0,
			DesiredSize:   15,
			UpcomingSize:  0,
			NodeIPs:       []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			CmDesiredSize: 15,
			UpdatedTime:   time.Now(),
			Status:        storage.StableState,
		},
		{
			NodeGroupID:   "NodeGroup2",
			ClusterID:     "Cluster1",
			MaxSize:       100,
			MinSize:       0,
			DesiredSize:   10,
			UpcomingSize:  0,
			NodeIPs:       []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			CmDesiredSize: 10,
			UpdatedTime:   time.Now(),
			Status:        storage.StableState,
		},
	}
	return nodeGroups
}

func getScaleUpNodeGroups() []*storage.NodeGroup {
	nodeGroups := []*storage.NodeGroup{
		{
			NodeGroupID:  "NodeGroup1",
			ClusterID:    "Cluster1",
			MaxSize:      100,
			MinSize:      0,
			DesiredSize:  20,
			UpcomingSize: 5,
			NodeIPs:      []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			UpdatedTime:  time.Now(),
			Status:       storage.ScaleUpState,
		},
		{
			NodeGroupID:  "NodeGroup2",
			ClusterID:    "Cluster1",
			MaxSize:      100,
			MinSize:      0,
			DesiredSize:  15,
			UpcomingSize: 5,
			NodeIPs:      []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			UpdatedTime:  time.Now(),
			Status:       storage.ScaleUpState,
		},
	}
	return nodeGroups
}

func getScaleDownNodeGroups() []*storage.NodeGroup {
	nodeGroups := []*storage.NodeGroup{
		{
			NodeGroupID:   "NodeGroup1",
			ClusterID:     "Cluster1",
			MaxSize:       100,
			MinSize:       0,
			DesiredSize:   10,
			UpcomingSize:  0,
			NodeIPs:       []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			CmDesiredSize: 15,
			UpdatedTime:   time.Now(),
			Status:        storage.ScaleDownState,
		},
		{
			NodeGroupID:   "NodeGroup2",
			ClusterID:     "Cluster1",
			MaxSize:       100,
			MinSize:       0,
			DesiredSize:   5,
			UpcomingSize:  0,
			NodeIPs:       []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			CmDesiredSize: 10,
			UpdatedTime:   time.Now(),
			Status:        storage.ScaleDownState,
		},
	}
	return nodeGroups
}

func getDiffNodeGroups() []*storage.NodeGroup {
	nodeGroups := []*storage.NodeGroup{
		{
			NodeGroupID:  "NodeGroup1",
			ClusterID:    "Cluster1",
			MaxSize:      100,
			MinSize:      0,
			DesiredSize:  10,
			UpcomingSize: 0,
			NodeIPs:      []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			UpdatedTime:  time.Now(),
			Status:       storage.ScaleDownState,
		},
		{
			NodeGroupID:  "NodeGroup2",
			ClusterID:    "Cluster1",
			MaxSize:      100,
			MinSize:      0,
			DesiredSize:  15,
			UpcomingSize: 5,
			NodeIPs:      []string{"IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP", "IP"},
			UpdatedTime:  time.Now(),
			Status:       storage.ScaleUpState,
		},
	}
	return nodeGroups
}
