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
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/scalingconfig"
)

// ========== Helper Functions ==========

// buildNodeGroupSetConfig creates a test NodeGroupSetList configuration
func buildNodeGroupSetConfig(sets []scalingconfig.NodeGroupSet,
	_ *scalingconfig.NodeGroupSet,
) *scalingconfig.NodeGroupSetList {
	return &scalingconfig.NodeGroupSetList{
		Sets: sets,
	}
}

// buildTestNodeWithResources creates a test node with specified resources
func buildTestNodeWithResources(name string, cpuMillis, memBytes int64) *apiv1.Node {
	node := BuildTestNode(name, cpuMillis, memBytes)
	SetNodeReadyState(node, true, time.Now())
	return node
}

// buildTestPodWithResources creates a test pod with specified resource requests
func buildTestPodWithResources(name string, cpuMillis, memBytes int64) *apiv1.Pod {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			SelfLink:  "/api/v1/namespaces/default/pods/" + name,
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:  "container",
					Image: "test-image",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuMillis, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(memBytes, resource.BinarySI),
						},
					},
				},
			},
		},
	}
	return pod
}

// setupTestAutoscalerContext creates a test autoscaling context with NodeGroupSetManager
// Returns the inner Context structure (from NewScaleTestAutoscalingContext)
func setupTestAutoscalerContext(t *testing.T, provider cloudprovider.CloudProvider,
	nodeGroupSetConfig *scalingconfig.NodeGroupSetList,
) *context.Context {
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUnneededTime:         time.Minute,
		},
		EstimatorName:    estimator.BinpackingEstimatorName,
		ScaleDownEnabled: true,
		MaxNodesTotal:    100,
		MaxCoresTotal:    1000,
		MaxMemoryTotal:   100000000,
	}

	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	ctx, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	assert.NoError(t, err)

	// Initialize NodeGroupSetManager if config provided
	if nodeGroupSetConfig != nil {
		nodeGroupSetManager := NewNodeGroupSetManager(nodeGroupSetConfig)
		ctx.NodeGroupSetManager = nodeGroupSetManager
	}

	return &ctx
}

// ========== Test Cases: Scale-Up Scenarios ==========

// TestNodeGroupSet_ScaleUpWithDifferentBufferRatios tests that different Sets trigger scale-up
// based on their configured buffered ratios
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_ScaleUpWithDifferentBufferRatios(t *testing.T) {
	// Setup NodeGroupSet configuration
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "high-buffer-set",
				NodeGroupIDs: []string{"ng-high"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.3, // 30% buffer required
					MemRatio: 0.3,
				},
			},
			{
				SetName:      "low-buffer-set",
				NodeGroupIDs: []string{"ng-low"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.1, // 10% buffer required
					MemRatio: 0.1,
				},
			},
		},
		nil, // no default set
	)

	// Create test nodes
	// ng-high: 2 nodes with 80% utilization (20% available, meets 30% requirement? No!)
	nodeHigh1 := buildTestNodeWithResources("node-high-1", 1000, 1000000)
	nodeHigh2 := buildTestNodeWithResources("node-high-2", 1000, 1000000)

	// ng-low: 2 nodes with 92% utilization (8% available, below 10% requirement)
	nodeLow1 := buildTestNodeWithResources("node-low-1", 1000, 1000000)
	nodeLow2 := buildTestNodeWithResources("node-low-2", 1000, 1000000)

	// Create pods to occupy resources
	// high-buffer-set: 800m * 2 = 1600m out of 2000m (80% utilization)
	podHigh1 := buildTestPodWithResources("pod-high-1", 800, 100000)
	podHigh1.Spec.NodeName = "node-high-1"

	podHigh2 := buildTestPodWithResources("pod-high-2", 800, 100000)
	podHigh2.Spec.NodeName = "node-high-2"

	// low-buffer-set: 920m * 2 = 1840m out of 2000m (92% utilization)
	podLow1 := buildTestPodWithResources("pod-low-1", 920, 100000)
	podLow1.Spec.NodeName = "node-low-1"

	podLow2 := buildTestPodWithResources("pod-low-2", 920, 100000)
	podLow2.Spec.NodeName = "node-low-2"

	// Setup cloud provider
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-high", 2, 10, 2)
	provider.AddNode("ng-high", nodeHigh1)
	provider.AddNode("ng-high", nodeHigh2)

	provider.AddNodeGroup("ng-low", 2, 10, 2)
	provider.AddNode("ng-low", nodeLow1)
	provider.AddNode("ng-low", nodeLow2)

	// Create NodeGroupSetManager
	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// Verify node group mapping
	assert.Equal(t, "high-buffer-set", manager.GetNodeGroupSet("ng-high"))
	assert.Equal(t, "low-buffer-set", manager.GetNodeGroupSet("ng-low"))

	// Build node infos for snapshot
	nodeInfos := make(map[string]*schedulerframework.NodeInfo)
	for _, node := range []*apiv1.Node{nodeHigh1, nodeHigh2, nodeLow1, nodeLow2} {
		nodeInfo := schedulerframework.NewNodeInfo()
		nodeInfo.SetNode(node)
		nodeInfos[node.Name] = nodeInfo
	}

	// Add pods to node infos
	for _, pod := range []*apiv1.Pod{podHigh1, podHigh2, podLow1, podLow2} {
		nodeInfo := nodeInfos[pod.Spec.NodeName]
		nodeInfo.AddPod(pod)
	}

	// Take resource snapshot
	snapshot, _, err := manager.TakeResourceSnapshot(nodeInfos, provider)
	assert.NoError(t, err)

	// Verify high-buffer-set resources
	highInfo := snapshot.Resources["high-buffer-set"]
	assert.NotNil(t, highInfo)
	assert.Equal(t, int64(2000), highInfo.TotalCPU) // 2 nodes * 1000m
	assert.InDelta(t, 0.80, highInfo.CPUUsageRatio, 0.01)
	// Available ratio = 1 - 0.80 = 0.20 (20%)

	// Verify low-buffer-set resources
	lowInfo := snapshot.Resources["low-buffer-set"]
	assert.NotNil(t, lowInfo)
	assert.Equal(t, int64(2000), lowInfo.TotalCPU)
	assert.InDelta(t, 0.92, lowInfo.CPUUsageRatio, 0.01)
	// Available ratio = 1 - 0.92 = 0.08 (8%)

	// Check scale-up triggers
	highNeedsScaleUp := manager.CheckScaleUpNeeded("high-buffer-set", highInfo)
	lowNeedsScaleUp := manager.CheckScaleUpNeeded("low-buffer-set", lowInfo)

	// high-buffer-set: 20% available < 30% required -> should scale up
	assert.True(t, highNeedsScaleUp, "high-buffer-set should trigger scale-up (20%% < 30%%)")

	// low-buffer-set: 8% available < 10% required -> should scale up
	assert.True(t, lowNeedsScaleUp, "low-buffer-set should trigger scale-up (8%% < 10%%)")

	// Verify priority ordering
	allSets := manager.GetAllSets()
	assert.Equal(t, []string{"high-buffer-set", "low-buffer-set"}, allSets)
}

// TestNodeGroupSet_MultipleSetsTriggerScaleUp tests multiple Sets triggering scale-up with priority ordering
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_MultipleSetsTriggerScaleUp(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "gpu-nodes",
				NodeGroupIDs: []string{"ng-gpu"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.15,
					MemRatio: 0.15,
				},
			},
			{
				SetName:      "general-nodes",
				NodeGroupIDs: []string{"ng-general"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.15,
					MemRatio: 0.15,
				},
			},
		},
		nil,
	)

	// Create nodes with low available ratios for both sets
	nodeGpu := buildTestNodeWithResources("node-gpu", 1000, 1000000)
	nodeGeneral := buildTestNodeWithResources("node-general", 1000, 1000000)

	// Pods consuming 90% of resources on each node (10% available < 15% buffer)
	podGpu := buildTestPodWithResources("pod-gpu", 900, 900000)
	podGpu.Spec.NodeName = "node-gpu"

	podGeneral := buildTestPodWithResources("pod-general", 900, 900000)
	podGeneral.Spec.NodeName = "node-general"

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-gpu", 1, 10, 1)
	provider.AddNode("ng-gpu", nodeGpu)

	provider.AddNodeGroup("ng-general", 1, 10, 1)
	provider.AddNode("ng-general", nodeGeneral)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// Build snapshot
	nodeInfos := make(map[string]*schedulerframework.NodeInfo)
	for _, node := range []*apiv1.Node{nodeGpu, nodeGeneral} {
		nodeInfo := schedulerframework.NewNodeInfo()
		nodeInfo.SetNode(node)
		nodeInfos[node.Name] = nodeInfo
	}
	nodeInfos["node-gpu"].AddPod(podGpu)
	nodeInfos["node-general"].AddPod(podGeneral)

	snapshot, _, err := manager.TakeResourceSnapshot(nodeInfos, provider)
	assert.NoError(t, err)

	// Both sets should trigger scale-up
	gpuNeedsScaleUp := manager.CheckScaleUpNeeded("gpu-nodes", snapshot.Resources["gpu-nodes"])
	generalNeedsScaleUp := manager.CheckScaleUpNeeded("general-nodes", snapshot.Resources["general-nodes"])

	assert.True(t, gpuNeedsScaleUp)
	assert.True(t, generalNeedsScaleUp)

	// Verify priority ordering (high priority first)
	allSets := manager.GetAllSets()
	assert.Equal(t, "gpu-nodes", allSets[0], "gpu-nodes should be first (priority 100)")
	assert.Equal(t, "general-nodes", allSets[1], "general-nodes should be second (priority 50)")
}

// TestNodeGroupSet_UpcomingNodesIncludedInSnapshot tests that upcoming nodes are included in resource calculations
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_UpcomingNodesIncludedInSnapshot(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "test-set",
				NodeGroupIDs: []string{"ng-test"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.1, // 10% buffer
					MemRatio: 0.1,
				},
			},
		},
		nil,
	)

	// Real node with high utilization
	realNode := buildTestNodeWithResources("node-real", 1000, 1000000)
	pod := buildTestPodWithResources("pod-1", 950, 950000)
	pod.Spec.NodeName = "node-real"

	// Upcoming node (simulates a node being provisioned)
	upcomingNode := buildTestNodeWithResources("node-upcoming", 1000, 1000000)
	if upcomingNode.Annotations == nil {
		upcomingNode.Annotations = make(map[string]string)
	}
	upcomingNode.Annotations[NodeUpcomingAnnotation] = valueTrue
	upcomingNode.Annotations[bcsNodeGroupAnnoKey] = "ng-test"

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-test", 1, 10, 2)
	provider.AddNode("ng-test", realNode)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// Build snapshot including upcoming node
	nodeInfos := make(map[string]*schedulerframework.NodeInfo)
	realNodeInfo := schedulerframework.NewNodeInfo()
	realNodeInfo.SetNode(realNode)
	realNodeInfo.AddPod(pod)
	nodeInfos["node-real"] = realNodeInfo

	upcomingNodeInfo := schedulerframework.NewNodeInfo()
	upcomingNodeInfo.SetNode(upcomingNode)
	nodeInfos["node-upcoming"] = upcomingNodeInfo

	scaleUpSnapshot, scaleDownSnapshot, err := manager.TakeResourceSnapshot(nodeInfos, provider)
	assert.NoError(t, err)

	// Verify scale-up snapshot includes upcoming node
	assert.Contains(t, scaleUpSnapshot.NodeCapacities, "node-upcoming")
	assert.Contains(t, scaleUpSnapshot.NodeCapacities, "node-real")

	// Total resources should include both nodes for scale-up
	testSetInfo := scaleUpSnapshot.Resources["test-set"]
	assert.NotNil(t, testSetInfo)
	assert.Equal(t, int64(2000), testSetInfo.TotalCPU, "Scale-up snapshot should include both real and upcoming nodes")

	// Available ratio should be high enough to not trigger scale-up
	// Total: 2000m, Used: 950m, Available: 1050m, Ratio: 52.5%
	assert.InDelta(t, 0.525, 1.0-testSetInfo.CPUUsageRatio, 0.01)

	needsScaleUp := manager.CheckScaleUpNeeded("test-set", testSetInfo)
	assert.False(t, needsScaleUp, "Should not scale up when upcoming node provides enough buffer")

	// Verify scale-down snapshot excludes upcoming node to avoid deleting real capacity based on virtual resources
	assert.NotContains(t, scaleDownSnapshot.NodeCapacities, "node-upcoming")
	assert.Contains(t, scaleDownSnapshot.NodeCapacities, "node-real")

	scaleDownSetInfo := scaleDownSnapshot.Resources["test-set"]
	assert.NotNil(t, scaleDownSetInfo)
	assert.Equal(t, int64(1000), scaleDownSetInfo.TotalCPU,
		"Scale-down snapshot should only include real nodes")
	assert.InDelta(t, 0.05, 1.0-scaleDownSetInfo.CPUUsageRatio, 0.01,
		"Scale-down snapshot should not count upcoming capacity as available buffer")
}

// TestNodeGroupSet_TakeResourceSnapshotSkipsToBeDeletedNodes tests that nodes already marked
// for deletion are excluded from snapshot resource accounting, even if they are still schedulable.
func TestNodeGroupSet_TakeResourceSnapshotSkipsToBeDeletedNodes(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "test-set",
				NodeGroupIDs: []string{"ng-test"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2,
					MemRatio: 0.2,
				},
			},
		},
		nil,
	)

	nodeActive := buildTestNodeWithResources("node-active", 1000, 1000000)
	nodeDeleting := buildTestNodeWithResources("node-deleting", 1000, 1000000)
	nodeDeleting.Spec.Taints = append(nodeDeleting.Spec.Taints, apiv1.Taint{
		Key:   deletetaint.ToBeDeletedTaint,
		Value: "1234567890",
	})

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-test", 2, 10, 2)
	provider.AddNode("ng-test", nodeActive)
	provider.AddNode("ng-test", nodeDeleting)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	nodeInfos := map[string]*schedulerframework.NodeInfo{
		"node-active":   schedulerframework.NewNodeInfo(),
		"node-deleting": schedulerframework.NewNodeInfo(),
	}
	nodeInfos["node-active"].SetNode(nodeActive)
	nodeInfos["node-deleting"].SetNode(nodeDeleting)

	snapshot, _, err := manager.TakeResourceSnapshot(nodeInfos, provider)
	assert.NoError(t, err)

	assert.Contains(t, snapshot.NodeCapacities, "node-active")
	assert.NotContains(t, snapshot.NodeCapacities, "node-deleting")
	assert.Contains(t, snapshot.NodeToGroup, "node-active")
	assert.NotContains(t, snapshot.NodeToGroup, "node-deleting")

	setInfo := snapshot.Resources["test-set"]
	assert.NotNil(t, setInfo)
	assert.Equal(t, int64(1000), setInfo.TotalCPU, "ToBeDeleted node should not count toward set total CPU")
	assert.Equal(t, int64(1000000), setInfo.TotalMemory,
		"ToBeDeleted node should not count toward set total memory")

	globalInfo := snapshot.Resources[GlobalSetName]
	assert.NotNil(t, globalInfo)
	assert.Equal(t, int64(1000), globalInfo.TotalCPU,
		"ToBeDeleted node should not count toward global total CPU")
	assert.Equal(t, int64(1000000), globalInfo.TotalMemory,
		"ToBeDeleted node should not count toward global total memory")
}

// TestNodeGroupSet_RefreshNodeGroups tests dynamic node group refresh
func TestNodeGroupSet_RefreshNodeGroups(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "set-a",
				NodeGroupIDs: []string{"ng-1", "ng-2"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2,
					MemRatio: 0.2,
				},
			},
		},
		&scalingconfig.NodeGroupSet{
			SetName: "default",
			BufferedRatios: scalingconfig.BufferedRatioConfig{
				CPURatio: 0.15,
				MemRatio: 0.15,
			},
		},
	)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-1", 0, 10, 0)
	provider.AddNodeGroup("ng-2", 0, 10, 0)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)

	// First refresh
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	assert.Equal(t, "set-a", manager.GetNodeGroupSet("ng-1"))
	assert.Equal(t, "set-a", manager.GetNodeGroupSet("ng-2"))

	// Add a new node group
	provider.AddNodeGroup("ng-3", 0, 10, 0)

	// Second refresh
	err = manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// New node group should use GlobalSetName (not in any explicit Set)
	assert.Equal(t, "set-a", manager.GetNodeGroupSet("ng-1"))
	assert.Equal(t, "set-a", manager.GetNodeGroupSet("ng-2"))
	assert.Equal(t, GlobalSetName, manager.GetNodeGroupSet("ng-3"))

	// Verify set node groups
	setANodeGroups := manager.GetNodeGroupsInSet("set-a")
	assert.Len(t, setANodeGroups, 2)

	// GlobalSetName is not a real Set, so GetNodeGroupsInSet should return empty
	globalNodeGroups := manager.GetNodeGroupsInSet(GlobalSetName)
	assert.Len(t, globalNodeGroups, 0, "GlobalSetName is not a real Set")
}

// ========== Test Cases: Scale-Down Scenarios ==========

// TestNodeGroupSet_ScaleDownBlockedByBufferConstraint tests that scale-down is blocked when
// removing a node would violate the Set's buffer requirement
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_ScaleDownBlockedByBufferConstraint(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "test-set",
				NodeGroupIDs: []string{"ng-test"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2, // 20% buffer required
					MemRatio: 0.2,
				},
			},
		},
		nil, // No DefaultSet
	)

	// 3 nodes, each 1000m CPU
	node1 := buildTestNodeWithResources("node-1", 1000, 1000000)
	node2 := buildTestNodeWithResources("node-2", 1000, 1000000)
	node3 := buildTestNodeWithResources("node-3", 1000, 1000000)

	// Total: 3000m, pods use 2100m, available 900m (30% available ratio, meets 20% buffer)
	pod1 := buildTestPodWithResources("pod-1", 100, 100000)
	pod1.Spec.NodeName = "node-1"

	pod2 := buildTestPodWithResources("pod-2", 1000, 1000000)
	pod2.Spec.NodeName = "node-2"

	pod3 := buildTestPodWithResources("pod-3", 1000, 1000000)
	pod3.Spec.NodeName = "node-3"

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-test", 3, 10, 3)
	provider.AddNode("ng-test", node1)
	provider.AddNode("ng-test", node2)
	provider.AddNode("ng-test", node3)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// Build snapshot
	nodeInfos := make(map[string]*schedulerframework.NodeInfo)
	for _, node := range []*apiv1.Node{node1, node2, node3} {
		nodeInfo := schedulerframework.NewNodeInfo()
		nodeInfo.SetNode(node)
		nodeInfos[node.Name] = nodeInfo
	}
	nodeInfos["node-1"].AddPod(pod1)
	nodeInfos["node-2"].AddPod(pod2)
	nodeInfos["node-3"].AddPod(pod3)

	snapshot, _, err := manager.TakeResourceSnapshot(nodeInfos, provider)
	assert.NoError(t, err)

	// Verify initial state
	testSetInfo := snapshot.Resources["test-set"]
	assert.Equal(t, int64(3000), testSetInfo.TotalCPU)
	assert.Equal(t, int64(900), testSetInfo.AvailableCPU)
	assert.InDelta(t, 0.30, 1.0-testSetInfo.CPUUsageRatio, 0.01) // 30% available

	// Try to remove node-1 (which has pod-1 with 100m CPU)
	podsToReschedule := []*apiv1.Pod{pod1}

	newInfo, err := manager.SimulateNodeRemoval("test-set", snapshot, "node-1", podsToReschedule)
	assert.NoError(t, err)

	// After removal:
	// Total: 3000 - 1000 = 2000m
	// Original available: 900m
	// Minus node-1's available (1000 - 100 = 900m)
	// Minus pods to reschedule: 100m
	// Final available: 900 - 900 - 100 = -100m (negative means overcommitted)
	// Available ratio: -100/2000 = -5% (less than 0%, definitely violates 20% buffer)

	assert.True(t, newInfo.AvailableCPU < 0 || (float64(newInfo.AvailableCPU)/float64(newInfo.TotalCPU)) < 0.2,
		"Available ratio should be below 20%% buffer requirement")

	// Check if scale-up would be needed after removal (which means removal should be blocked)
	needsScaleUp := manager.CheckScaleUpNeeded("test-set", newInfo)
	assert.True(t, needsScaleUp, "Removing node-1 would require scale-up, so removal should be blocked")
}

// TestNodeGroupSet_ScaleDownAcrossSets tests scale-down decisions across different Sets
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_ScaleDownAcrossSets(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "set-a",
				NodeGroupIDs: []string{"ng-a"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2, // 20% buffer
					MemRatio: 0.2,
				},
			},
			{
				SetName:      "set-b",
				NodeGroupIDs: []string{"ng-b"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.3, // 30% buffer (higher requirement)
					MemRatio: 0.3,
				},
			},
		},
		nil,
	)

	// Set A: 3 nodes, 60% utilization (40% available, exceeds 20% buffer)
	// Total: 3 * 1000m = 3000m, Used: 1800m, Available: 1200m (40%)
	// After removing node-a1: Total: 2000m, Used: 1800m, Available: 200m (10%)
	// 10% < 20% buffer, so removal should be BLOCKED
	nodeA1 := buildTestNodeWithResources("node-a1", 1000, 1000000)
	nodeA2 := buildTestNodeWithResources("node-a2", 1000, 1000000)
	nodeA3 := buildTestNodeWithResources("node-a3", 1000, 1000000)

	podA1 := buildTestPodWithResources("pod-a1", 900, 900000)
	podA1.Spec.NodeName = "node-a2"
	podA2 := buildTestPodWithResources("pod-a2", 900, 900000)
	podA2.Spec.NodeName = "node-a3"
	// node-a1 is idle

	// Set B: 3 nodes, 40% utilization (60% available, exceeds 30% buffer)
	// Total: 3 * 1000m = 3000m, Used: 1200m, Available: 1800m (60%)
	// After removing node-b1: Total: 2000m, Used: 1200m, Available: 800m (40%)
	// 40% > 30% buffer, so removal should be ALLOWED
	nodeB1 := buildTestNodeWithResources("node-b1", 1000, 1000000)
	nodeB2 := buildTestNodeWithResources("node-b2", 1000, 1000000)
	nodeB3 := buildTestNodeWithResources("node-b3", 1000, 1000000)

	podB1 := buildTestPodWithResources("pod-b1", 600, 600000)
	podB1.Spec.NodeName = "node-b2"
	podB2 := buildTestPodWithResources("pod-b2", 600, 600000)
	podB2.Spec.NodeName = "node-b3"
	// node-b1 is idle

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-a", 3, 10, 3)
	provider.AddNode("ng-a", nodeA1)
	provider.AddNode("ng-a", nodeA2)
	provider.AddNode("ng-a", nodeA3)

	provider.AddNodeGroup("ng-b", 3, 10, 3)
	provider.AddNode("ng-b", nodeB1)
	provider.AddNode("ng-b", nodeB2)
	provider.AddNode("ng-b", nodeB3)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// Build snapshot
	nodeInfos := make(map[string]*schedulerframework.NodeInfo)
	for _, node := range []*apiv1.Node{nodeA1, nodeA2, nodeA3, nodeB1, nodeB2, nodeB3} {
		nodeInfo := schedulerframework.NewNodeInfo()
		nodeInfo.SetNode(node)
		nodeInfos[node.Name] = nodeInfo
	}
	nodeInfos["node-a2"].AddPod(podA1)
	nodeInfos["node-a3"].AddPod(podA2)
	nodeInfos["node-b2"].AddPod(podB1)
	nodeInfos["node-b3"].AddPod(podB2)

	snapshot, _, err := manager.TakeResourceSnapshot(nodeInfos, provider)
	assert.NoError(t, err)

	// Verify set-a state: 40% available >= 20% buffer (sufficient buffer)
	setAInfo := snapshot.Resources["set-a"]
	assert.InDelta(t, 0.40, 1.0-setAInfo.CPUUsageRatio, 0.02)
	setANeedsBuffer := manager.CheckScaleUpNeeded("set-a", setAInfo)
	assert.False(t, setANeedsBuffer, "set-a has sufficient buffer")

	// Verify set-b state: 50% available >= 30% buffer (sufficient buffer)
	setBInfo := snapshot.Resources["set-b"]
	assert.InDelta(t, 0.60, 1.0-setBInfo.CPUUsageRatio, 0.02)
	setBNeedsBuffer := manager.CheckScaleUpNeeded("set-b", setBInfo)
	assert.False(t, setBNeedsBuffer, "set-b has sufficient buffer")

	// Simulate removing node-a1 (idle, no pods to reschedule)
	newSetAInfo, err := manager.SimulateNodeRemoval("set-a", snapshot, "node-a1", nil)
	assert.NoError(t, err)
	// After removing node-a1: Total=2000m, Used=1800m, Available=200m, Ratio=10%
	assert.InDelta(t, 0.10, 1.0-newSetAInfo.CPUUsageRatio, 0.02)
	setACanRemove := !manager.CheckScaleUpNeeded("set-a", newSetAInfo)
	assert.False(t, setACanRemove, "node-a1 cannot be removed (10% < 20% buffer)")

	// Simulate removing node-b1 (idle, no pods to reschedule)
	newSetBInfo, err := manager.SimulateNodeRemoval("set-b", snapshot, "node-b1", nil)
	assert.NoError(t, err)
	// After removing node-b1: Total=2000m, Used=1200m, Available=800m, Ratio=40%
	assert.InDelta(t, 0.40, 1.0-newSetBInfo.CPUUsageRatio, 0.02)
	setBCanRemove := !manager.CheckScaleUpNeeded("set-b", newSetBInfo)
	assert.True(t, setBCanRemove, "node-b1 cannot be removed (40% > 30% buffer)")
}

// ========== Test Cases: Compatibility Tests ==========

// TestNodeGroupSet_MixedModeWithGlobalSet tests that Sets and GlobalSet work together
// NodeGroups not in any Set should use GlobalSetName (not a real Set, uses global config)
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_MixedModeWithGlobalSet(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "gpu-set",
				NodeGroupIDs: []string{"ng-gpu"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.1, // 10% buffer
					MemRatio: 0.1,
				},
			},
		},
		nil, // No DefaultSet (feature removed)
	)

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-gpu", 1, 10, 1)
	provider.AddNodeGroup("ng-general", 1, 10, 1)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// Verify mapping
	assert.Equal(t, "gpu-set", manager.GetNodeGroupSet("ng-gpu"), "ng-gpu should be in gpu-set")
	assert.Equal(t, GlobalSetName, manager.GetNodeGroupSet("ng-general"), "ng-general should use GlobalSetName")

	// Verify GetAllSets only returns real Sets (not GlobalSetName)
	allSets := manager.GetAllSets()
	assert.Equal(t, 1, len(allSets), "Should only have 1 real Set")
	assert.Contains(t, allSets, "gpu-set")
	assert.NotContains(t, allSets, GlobalSetName, "GlobalSetName is not a real Set")

	// Verify GetSetConfig works for real Set
	gpuConfig, err := manager.GetSetConfig("gpu-set")
	assert.NoError(t, err)
	assert.Equal(t, "gpu-set", gpuConfig.SetName)
	assert.Equal(t, 0.1, gpuConfig.BufferedRatios.CPURatio)

	// Verify GetSetConfig returns error for GlobalSetName
	_, err = manager.GetSetConfig(GlobalSetName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "special marker for global configuration")
}

// ========== Test Cases: Snapshot and Caching ==========

// TestNodeGroupSet_SnapshotIsStale tests snapshot staleness detection
func TestNodeGroupSet_SnapshotIsStale(t *testing.T) {
	snapshot := &SetResourceSnapshot{
		Timestamp: time.Now(),
	}

	// Immediately after creation, should not be stale
	assert.False(t, snapshot.IsStale())

	// After 5 seconds, should still not be stale (TTL is 10s)
	snapshot.Timestamp = time.Now().Add(-5 * time.Second)
	assert.False(t, snapshot.IsStale())

	// After 11 seconds, should be stale
	snapshot.Timestamp = time.Now().Add(-11 * time.Second)
	assert.True(t, snapshot.IsStale())
}

// TestNodeGroupSet_SnapshotDeepCopy tests deep copy of snapshot
func TestNodeGroupSet_SnapshotDeepCopy(t *testing.T) {
	original := &SetResourceSnapshot{
		Timestamp: time.Now(),
		Resources: map[string]*NodeGroupSetResourceInfo{
			"set-a": {
				SetName:          "set-a",
				TotalCPU:         1000,
				AvailableCPU:     200,
				CPUUsageRatio:    0.8,
				TriggeredScaleUp: true,
			},
		},
		NodeCapacities: map[string]*NodeResourceCapacity{
			"node-1": {
				NodeName:          "node-1",
				AllocatableCPU:    1000,
				AllocatableMemory: 1000000,
				UsedCPU:           800,
				UsedMemory:        800000,
			},
		},
		NodeToGroup: map[string]string{
			"node-1": "ng-1",
		},
	}

	// Deep copy
	copied := original.DeepCopy()
	assert.NotNil(t, copied)

	// Verify independence: modify copied, original should not change
	copied.Resources["set-a"].TotalCPU = 2000
	assert.Equal(t, int64(1000), original.Resources["set-a"].TotalCPU)

	copied.NodeCapacities["node-1"].UsedCPU = 500
	assert.Equal(t, int64(800), original.NodeCapacities["node-1"].UsedCPU)

	copied.NodeToGroup["node-1"] = "ng-2"
	assert.Equal(t, "ng-1", original.NodeToGroup["node-1"])
}

// TestNodeGroupSet_SnapshotUpdateSetResourceInfo tests updating set resource info in snapshot
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_SnapshotUpdateSetResourceInfo(t *testing.T) {
	snapshot := &SetResourceSnapshot{
		Resources: map[string]*NodeGroupSetResourceInfo{
			"set-a": {
				SetName:       "set-a",
				TotalCPU:      3000,
				AvailableCPU:  900,
				CPUUsageRatio: 0.7,
			},
		},
	}

	newInfo := &NodeGroupSetResourceInfo{
		SetName:       "set-a",
		TotalCPU:      2000,
		AvailableCPU:  400,
		CPUUsageRatio: 0.8,
	}

	snapshot.UpdateSetResourceInfo("set-a", newInfo)

	// Verify update
	updated := snapshot.Resources["set-a"]
	assert.Equal(t, int64(2000), updated.TotalCPU)
	assert.Equal(t, int64(400), updated.AvailableCPU)
	assert.InDelta(t, 0.8, updated.CPUUsageRatio, 0.01)

	// Verify it's a copy, not a reference
	newInfo.TotalCPU = 9999
	assert.Equal(t, int64(2000), snapshot.Resources["set-a"].TotalCPU)
}

// TestNodeGroupSet_NoConfigFallback tests that nil NodeGroupSetManager doesn't break existing logic
func TestNodeGroupSet_NoConfigFallback(t *testing.T) {
	// Create autoscaler context without NodeGroupSetManager
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-default", 1, 10, 1)

	ctx := setupTestAutoscalerContext(t, provider, nil)

	// NodeGroupSetManager should be nil
	assert.Nil(t, ctx.NodeGroupSetManager)

	// This should not panic and should use global buffered ratio instead
	// (tested implicitly through integration tests)
}

// ========== Test Cases: Edge Cases ==========

// TestNodeGroupSet_ColdStartWithZeroNodes tests that Sets with no nodes trigger scale up
// This is a critical edge case: when a Set has 0 nodes (cold start), the usage ratio is 0.0,
// which would be interpreted as 100% available if not handled specially.
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_ColdStartWithZeroNodes(t *testing.T) {
	// Create a Set configuration with buffer requirements
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "cold-start-set",
				NodeGroupIDs: []string{"ng-cold"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.3, // 30% buffer required
					MemRatio: 0.3,
				},
			},
		},
		nil,
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)

	// Create a NodeGroup but with NO nodes in the cluster
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-cold", 0, 10, 0) // minSize=0, maxSize=10, currentSize=0

	nodeGroups := provider.NodeGroups()
	err := manager.RefreshNodeGroups(nodeGroups)
	assert.NoError(t, err)

	// Create an EMPTY cluster snapshot (no nodes)
	allNodeInfos := make(map[string]*schedulerframework.NodeInfo)

	// Take resource snapshot (should handle empty cluster)
	snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider)
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)

	// Verify Set resource info shows zero resources
	setInfo := snapshot.Resources["cold-start-set"]
	assert.NotNil(t, setInfo, "Set should have resource info even with 0 nodes")
	assert.Equal(t, int64(0), setInfo.TotalCPU, "TotalCPU should be 0")
	assert.Equal(t, int64(0), setInfo.TotalMemory, "TotalMemory should be 0")
	assert.Equal(t, int64(0), setInfo.AvailableCPU, "AvailableCPU should be 0")
	assert.Equal(t, int64(0), setInfo.AvailableMemory, "AvailableMemory should be 0")
	assert.Equal(t, 0.0, setInfo.CPUUsageRatio, "CPUUsageRatio should be 0.0 (not calculated)")

	// Critical test: CheckScaleUpNeeded should return TRUE for cold start
	// Without the fix, this would return false (0.0 usage = 100% available)
	needsScaleUp := manager.CheckScaleUpNeeded("cold-start-set", setInfo)
	assert.True(t, needsScaleUp, "Set with 0 nodes (cold start) should trigger scale up to meet buffer requirement")
}

// TestNodeGroupSet_ColdStartVsNormalOperation compares behavior between cold start and normal operation
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_ColdStartVsNormalOperation(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "test-set",
				NodeGroupIDs: []string{"ng-test"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2, // 20% buffer
					MemRatio: 0.2,
				},
			},
		},
		nil,
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-test", 0, 10, 0)

	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// Scenario 1: Cold start (0 nodes)
	t.Run("ColdStart", func(t *testing.T) {
		allNodeInfos := make(map[string]*schedulerframework.NodeInfo)
		snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider)
		assert.NoError(t, err)

		setInfo := snapshot.Resources["test-set"]
		needsScaleUp := manager.CheckScaleUpNeeded("test-set", setInfo)
		assert.True(t, needsScaleUp, "Cold start should trigger scale up")
	})

	// Scenario 2: Normal operation with sufficient buffer (25% available > 20% required)
	t.Run("SufficientBuffer", func(t *testing.T) {
		// Add node to provider first (so NodeGroupForNode can find it)
		node1 := buildTestNodeWithResources("node-1", 1000, 2000000)
		provider.AddNode("ng-test", node1)

		pod1 := buildTestPodWithResources("pod-1", 750, 1500000) // 75% CPU usage

		allNodeInfos := map[string]*schedulerframework.NodeInfo{
			"node-1": schedulerframework.NewNodeInfo(pod1),
		}
		allNodeInfos["node-1"].SetNode(node1)

		snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider)
		assert.NoError(t, err)

		setInfo := snapshot.Resources["test-set"]
		// Available: 250m / 1000m = 25% > 20% buffer
		assert.InDelta(t, 0.25, 1.0-setInfo.CPUUsageRatio, 0.02)
		needsScaleUp := manager.CheckScaleUpNeeded("test-set", setInfo)
		assert.False(t, needsScaleUp, "Sufficient buffer should not trigger scale up")
	})

	// Scenario 3: Normal operation with insufficient buffer (15% available < 20% required)
	t.Run("InsufficientBuffer", func(t *testing.T) {
		// Create a fresh provider for this subtest
		provider2 := testprovider.NewTestCloudProvider(nil, nil)
		provider2.AddNodeGroup("ng-test", 0, 10, 0)

		// Add node to provider first (so NodeGroupForNode can find it)
		node1 := buildTestNodeWithResources("node-1", 1000, 2000000)
		provider2.AddNode("ng-test", node1)

		pod1 := buildTestPodWithResources("pod-1", 850, 1700000) // 85% CPU usage

		allNodeInfos := map[string]*schedulerframework.NodeInfo{
			"node-1": schedulerframework.NewNodeInfo(pod1),
		}
		allNodeInfos["node-1"].SetNode(node1)

		snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider2)
		assert.NoError(t, err)

		setInfo := snapshot.Resources["test-set"]
		// Available: 150m / 1000m = 15% < 20% buffer
		assert.InDelta(t, 0.15, 1.0-setInfo.CPUUsageRatio, 0.02)
		needsScaleUp := manager.CheckScaleUpNeeded("test-set", setInfo)
		assert.True(t, needsScaleUp, "Insufficient buffer should trigger scale up")
	})
}

// TestNodeGroupSet_SimulateRemovalOfLastNode 测试删除 Set 中最后一个节点时的行为
// Bug fix: 当删除后 TotalCPU=0 时，应设置 CPUUsageRatio=1.0 以阻止继续缩容
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_SimulateRemovalOfLastNode(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "single-node-set",
				NodeGroupIDs: []string{"ng-single"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2, // 20% buffer required
					MemRatio: 0.2,
				},
			},
		},
		nil,
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-single", 1, 10, 1)

	// 创建一个只有1个节点的 Set
	node1 := buildTestNodeWithResources("node-1", 8000, 16000000000) // 8 core, 16GB
	provider.AddNode("ng-single", node1)

	// 节点上有一些 Pods（使用25%资源）
	pod1 := buildTestPodWithResources("pod-1", 2000, 4000000000)
	allNodeInfos := map[string]*schedulerframework.NodeInfo{
		"node-1": schedulerframework.NewNodeInfo(pod1),
	}
	allNodeInfos["node-1"].SetNode(node1)

	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// 创建资源快照
	snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider)
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)

	// 验证删除前的状态
	beforeInfo := snapshot.Resources["single-node-set"]
	assert.Equal(t, int64(8000), beforeInfo.TotalCPU)
	assert.Equal(t, int64(6000), beforeInfo.AvailableCPU)   // 8000 - 2000
	assert.InDelta(t, 0.25, beforeInfo.CPUUsageRatio, 0.01) // 25% usage

	// 模拟删除这个唯一的节点
	// Pod 需要重调度（但 Set 为空后无处可调度）
	afterInfo, err := manager.SimulateNodeRemoval(
		"single-node-set",
		snapshot,
		"node-1",
		[]*apiv1.Pod{pod1},
	)
	assert.NoError(t, err)
	assert.NotNil(t, afterInfo)

	// 验证删除后的状态
	assert.Equal(t, int64(0), afterInfo.TotalCPU, "After removing last node, TotalCPU should be 0")
	assert.Equal(t, int64(0), afterInfo.TotalMemory, "After removing last node, TotalMemory should be 0")

	// 关键测试：使用率应该是 1.0（100%），而不是 0.0
	// 这样 CheckScaleUpNeeded 会判断为资源不足，阻止删除这个节点
	assert.Equal(t, 1.0, afterInfo.CPUUsageRatio,
		"When Set becomes empty after node removal, CPUUsageRatio should be 1.0 to prevent scale-down")
	assert.Equal(t, 1.0, afterInfo.MemUsageRatio,
		"When Set becomes empty after node removal, MemUsageRatio should be 1.0 to prevent scale-down")

	// 验证 CheckScaleUpNeeded 会正确阻止缩容
	// 使用率 1.0 → 可用率 0.0 < 20% buffer → 需要扩容 → 不应缩容
	needsScaleUp := manager.CheckScaleUpNeeded("single-node-set", afterInfo)
	assert.True(t, needsScaleUp,
		"After removing last node, CheckScaleUpNeeded should return true to block scale-down")
}

// TestNodeGroupSet_SimulateRemovalOfNonLastNode 测试删除非最后一个节点的正常情况
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_SimulateRemovalOfNonLastNode(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "multi-node-set",
				NodeGroupIDs: []string{"ng-multi"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2, // 20% buffer required
					MemRatio: 0.2,
				},
			},
		},
		nil,
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-multi", 2, 10, 2)

	// 创建有2个节点的 Set
	node1 := buildTestNodeWithResources("node-1", 8000, 16000000000)
	node2 := buildTestNodeWithResources("node-2", 8000, 16000000000)
	provider.AddNode("ng-multi", node1)
	provider.AddNode("ng-multi", node2)

	// node1 上有 Pod（使用 2000m CPU）
	pod1 := buildTestPodWithResources("pod-1", 2000, 4000000000)
	allNodeInfos := map[string]*schedulerframework.NodeInfo{
		"node-1": schedulerframework.NewNodeInfo(pod1),
		"node-2": schedulerframework.NewNodeInfo(), // node2 空闲
	}
	allNodeInfos["node-1"].SetNode(node1)
	allNodeInfos["node-2"].SetNode(node2)

	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider)
	assert.NoError(t, err)

	beforeInfo := snapshot.Resources["multi-node-set"]
	assert.Equal(t, int64(16000), beforeInfo.TotalCPU)     // 2 nodes
	assert.Equal(t, int64(14000), beforeInfo.AvailableCPU) // 16000 - 2000

	// 模拟删除 node2（空闲节点）
	afterInfo, err := manager.SimulateNodeRemoval(
		"multi-node-set",
		snapshot,
		"node-2",
		[]*apiv1.Pod{}, // node2 没有 Pod
	)
	assert.NoError(t, err)

	// 删除后还有 node1 剩余
	assert.Equal(t, int64(8000), afterInfo.TotalCPU, "After removing one node, TotalCPU should be 8000")
	assert.Equal(t, int64(6000), afterInfo.AvailableCPU, "AvailableCPU should be 6000 (8000 - 2000)")

	// 使用率应该正常计算（不是特殊值 1.0）
	assert.InDelta(t, 0.25, afterInfo.CPUUsageRatio, 0.01, "CPUUsageRatio should be 0.25 (2000/8000)")

	// 可用率 75% > 20% buffer，不需要扩容，允许缩容
	needsScaleUp := manager.CheckScaleUpNeeded("multi-node-set", afterInfo)
	assert.False(t, needsScaleUp, "With 75% available buffer, should not need scale up")
}

// TestNodeGroupSet_SimulateRemovalWithZeroBufferRatio 测试 buffer ratio 为 0 时的行为
// 当所有 buffer ratio 都为 0 时，应该允许将 Set 缩容到空
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_SimulateRemovalWithZeroBufferRatio(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "zero-buffer-set",
				NodeGroupIDs: []string{"ng-zero"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0, // 不要求 CPU buffer
					MemRatio: 0, // 不要求 Memory buffer
				},
			},
		},
		nil,
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-zero", 1, 10, 1)

	// 创建只有1个节点的 Set
	node1 := buildTestNodeWithResources("node-1", 8000, 16000000000)
	provider.AddNode("ng-zero", node1)

	// 节点上有 Pod（使用 90% 资源，但 buffer ratio 为 0 所以不关心）
	pod1 := buildTestPodWithResources("pod-1", 7200, 14000000000)
	allNodeInfos := map[string]*schedulerframework.NodeInfo{
		"node-1": schedulerframework.NewNodeInfo(pod1),
	}
	allNodeInfos["node-1"].SetNode(node1)

	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider)
	assert.NoError(t, err)

	// 模拟删除这个唯一的节点
	afterInfo, err := manager.SimulateNodeRemoval(
		"zero-buffer-set",
		snapshot,
		"node-1",
		[]*apiv1.Pod{pod1},
	)
	assert.NoError(t, err)

	// 删除后 Set 为空
	assert.Equal(t, int64(0), afterInfo.TotalCPU)
	assert.Equal(t, int64(0), afterInfo.TotalMemory)
	assert.Equal(t, 1.0, afterInfo.CPUUsageRatio, "Empty set should have usage ratio 1.0")
	assert.Equal(t, 1.0, afterInfo.MemUsageRatio, "Empty set should have usage ratio 1.0")

	// 关键测试：当所有 buffer ratio 都为 0 时，CheckScaleUpNeeded 应该返回 false
	// 允许将 Set 缩容到空
	needsScaleUp := manager.CheckScaleUpNeeded("zero-buffer-set", afterInfo)
	assert.False(t, needsScaleUp,
		"With all buffer ratios at 0, should allow scaling set to empty")
}

// TestNodeGroupSet_SimulateRemovalWithNonZeroBufferRatio 测试有非零 buffer ratio 时的行为
// 当配置了非零 buffer ratio 时，不应该允许将 Set 缩容到空
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_SimulateRemovalWithNonZeroBufferRatio(t *testing.T) {
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "nonzero-buffer-set",
				NodeGroupIDs: []string{"ng-nonzero"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.2, // 要求 20% CPU buffer
					MemRatio: 0,   // 不要求 Memory buffer
				},
			},
		},
		nil,
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-nonzero", 1, 10, 1)

	// 创建只有1个节点的 Set
	node1 := buildTestNodeWithResources("node-1", 8000, 16000000000)
	provider.AddNode("ng-nonzero", node1)

	// 节点上几乎没有负载
	pod1 := buildTestPodWithResources("pod-1", 100, 100000000)
	allNodeInfos := map[string]*schedulerframework.NodeInfo{
		"node-1": schedulerframework.NewNodeInfo(pod1),
	}
	allNodeInfos["node-1"].SetNode(node1)

	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	snapshot, _, err := manager.TakeResourceSnapshot(allNodeInfos, provider)
	assert.NoError(t, err)

	// 模拟删除这个唯一的节点
	afterInfo, err := manager.SimulateNodeRemoval(
		"nonzero-buffer-set",
		snapshot,
		"node-1",
		[]*apiv1.Pod{pod1},
	)
	assert.NoError(t, err)

	// 删除后 Set 为空
	assert.Equal(t, int64(0), afterInfo.TotalCPU)
	assert.Equal(t, int64(0), afterInfo.TotalMemory)

	// 关键测试：配置了非零 CPU ratio，所以不允许缩容到空
	needsScaleUp := manager.CheckScaleUpNeeded("nonzero-buffer-set", afterInfo)
	assert.True(t, needsScaleUp,
		"With non-zero CPU ratio, should not allow scaling set to empty")
}

// TestNodeGroupSet_GlobalSetNameWithGlobalBufferRatio 测试 GlobalSetName 节点使用全局 buffer ratio
// NOCC:tosa/fn_length(设计如此)
func TestNodeGroupSet_GlobalSetNameWithGlobalBufferRatio(t *testing.T) {
	// 只配置一个显式的 Set
	nodeGroupSetConfig := buildNodeGroupSetConfig(
		[]scalingconfig.NodeGroupSet{
			{
				SetName:      "explicit-set",
				NodeGroupIDs: []string{"ng-explicit"},
				BufferedRatios: scalingconfig.BufferedRatioConfig{
					CPURatio: 0.3,
					MemRatio: 0.3,
				},
			},
		},
		nil,
	)

	manager := NewNodeGroupSetManager(nodeGroupSetConfig)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng-explicit", 1, 10, 1)
	provider.AddNodeGroup("ng-global", 1, 10, 1) // 不在任何显式 Set 中

	err := manager.RefreshNodeGroups(provider.NodeGroups())
	assert.NoError(t, err)

	// 验证 Set 分配
	assert.Equal(t, "explicit-set", manager.GetNodeGroupSet("ng-explicit"))
	assert.Equal(t, GlobalSetName, manager.GetNodeGroupSet("ng-global"))

	// 验证 GlobalSetName 不能获取 Set 配置
	_, err = manager.GetSetConfig(GlobalSetName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "special marker for global configuration")

	// 验证 GlobalSetName 节点不会被 GetNodeGroupsInSet 返回
	globalNodes := manager.GetNodeGroupsInSet(GlobalSetName)
	assert.Empty(t, globalNodes, "GlobalSetName is not a real Set, should return empty")

	// 验证 GlobalSetName 不在 GetAllSets 结果中
	allSets := manager.GetAllSets()
	assert.NotContains(t, allSets, GlobalSetName)
}

// mockNodeGroupSetManager implements NodeGroupSetManagerInterface for testing
type mockNodeGroupSetManager struct {
	nodeGroupToSet map[string]string
}

func (m *mockNodeGroupSetManager) GetNodeGroupSet(nodeGroupID string) string {
	if setName, exists := m.nodeGroupToSet[nodeGroupID]; exists {
		return setName
	}
	return GlobalSetName
}

// TestDeduplicatePodAssignments tests the shared Pod deduplication function
