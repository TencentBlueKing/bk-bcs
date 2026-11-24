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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/scalingconfig"
)

// TestNodeGroupSet_GlobalSetNameWithoutDefaultSet tests the behavior when:
// - NodeGroupSetManager is enabled
// - A NodeGroup doesn't belong to any explicit Set
// Expected: NodeGroup should use GlobalSetName and global cluster-level ratios
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_GlobalSetNameWithoutDefaultSet(t *testing.T) {
	// Create config without DefaultSet (DefaultSet is removed from design)
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "gpu-set",
				NodeGroupIDs: []string{"ng-gpu"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.3, // 30% buffer for GPU
					MemRatio: 0.3,
				},
			},
		},
		nil, // No DefaultSet (feature removed)
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)

	// Test 1: GPU NodeGroup should be in gpu-set
	setName := manager.findMatchingSet(&testNodeGroup{id: "ng-gpu"})
	assert.Equal(t, "gpu-set", setName, "ng-gpu should be in gpu-set")

	// Test 2: General NodeGroup should get GlobalSetName
	setName = manager.findMatchingSet(&testNodeGroup{id: "ng-general"})
	assert.Equal(t, GlobalSetName, setName, "ng-general should get GlobalSetName (not in any Set)")

	// Test 3: Memory NodeGroup should also get GlobalSetName
	setName = manager.findMatchingSet(&testNodeGroup{id: "ng-memory"})
	assert.Equal(t, GlobalSetName, setName, "ng-memory should get GlobalSetName")

	// Test 4: GetSetConfig should return error for GlobalSetName
	_, err := manager.GetSetConfig(GlobalSetName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "special marker for global configuration")
}

// TestNodeGroupSet_GetAllSetsExcludesGlobalSetName verifies that GlobalSetName
// is NOT included in GetAllSets() results, as it's not a real Set
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_GetAllSetsExcludesGlobalSetName(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "gpu-set",
				NodeGroupIDs: []string{"ng-gpu"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.3,
					MemRatio: 0.3,
				},
			},
			{
				SetName:      "memory-set",
				NodeGroupIDs: []string{"ng-memory"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.25,
					MemRatio: 0.25,
				},
			},
		},
		nil, // No DefaultSet
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)

	allSets := manager.GetAllSets()

	// Should only contain real Sets (gpu-set, memory-set), not GlobalSetName
	assert.Equal(t, 2, len(allSets))
	assert.Contains(t, allSets, "gpu-set")
	assert.Contains(t, allSets, "memory-set")
	assert.NotContains(t, allSets, GlobalSetName, "GlobalSetName should NOT be in GetAllSets()")
}

// testNodeGroup is a simple mock for cloudprovider.NodeGroup
type testNodeGroup struct {
	id string
}

func (ng *testNodeGroup) Id() string {
	return ng.id
}

func (ng *testNodeGroup) MaxSize() int {
	return 10
}

func (ng *testNodeGroup) MinSize() int {
	return 0
}

func (ng *testNodeGroup) TargetSize() (int, error) {
	return 0, nil
}

func (ng *testNodeGroup) IncreaseSize(delta int) error {
	return nil
}

func (ng *testNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	return nil
}

func (ng *testNodeGroup) DecreaseTargetSize(delta int) error {
	return nil
}

func (ng *testNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	return []cloudprovider.Instance{}, nil
}

func (ng *testNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, nil
}

func (ng *testNodeGroup) Exist() bool {
	return true
}

func (ng *testNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return ng, nil
}

func (ng *testNodeGroup) Delete() error {
	return nil
}

func (ng *testNodeGroup) Autoprovisioned() bool {
	return false
}

func (ng *testNodeGroup) Debug() string {
	return ng.id
}

func (ng *testNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return &defaults, nil
}
