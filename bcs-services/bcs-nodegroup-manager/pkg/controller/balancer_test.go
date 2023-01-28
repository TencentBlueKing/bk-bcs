/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
		{NodeGroupID: "NodeGroup1", ClusterID: "ClusterID1", Weight: 1},
		{NodeGroupID: "NodeGroup2", ClusterID: "ClusterID1", Weight: 1},
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
