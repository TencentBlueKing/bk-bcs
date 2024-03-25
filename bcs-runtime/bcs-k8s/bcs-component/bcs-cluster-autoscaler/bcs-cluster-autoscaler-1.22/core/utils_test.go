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
	"fmt"
	"testing"
	"time"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	kube_record "k8s.io/client-go/tools/record"

	"github.com/stretchr/testify/assert"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const MiB = 1024 * 1024

func TestPodSchedulableMap(t *testing.T) {
	rc1 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc1",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc1",
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}

	rc2 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc2",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc2",
			UID:       "12345678-1234-1234-1234-12345678901a",
		},
	}

	pMap := make(podSchedulableMap)

	podInRc1_1 := BuildTestPod("podInRc1_1", 500, 1000)
	podInRc1_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)

	podInRc2 := BuildTestPod("podInRc2", 500, 1000)
	podInRc2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)

	// Basic sanity checks
	_, found := pMap.get(podInRc1_1)
	assert.False(t, found)
	pMap.set(podInRc1_1, nil)
	err, found := pMap.get(podInRc1_1)
	assert.True(t, found)
	assert.Nil(t, err)

	cpuErr := &simulator.PredicateError{}

	// Pod in different RC
	_, found = pMap.get(podInRc2)
	assert.False(t, found)
	pMap.set(podInRc2, cpuErr)
	err, found = pMap.get(podInRc2)
	assert.True(t, found)
	assert.Equal(t, cpuErr, err)

	// Another replica in rc1
	podInRc1_2 := BuildTestPod("podInRc1_1", 500, 1000)
	podInRc1_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	err, found = pMap.get(podInRc1_2)
	assert.True(t, found)
	assert.Nil(t, err)

	// A pod in rc1, but with different requests
	differentPodInRc1 := BuildTestPod("differentPodInRc1", 1000, 1000)
	differentPodInRc1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController",
		"extensions/v1beta1", rc1.UID)
	_, found = pMap.get(differentPodInRc1)
	assert.False(t, found)
	pMap.set(differentPodInRc1, cpuErr)
	err, found = pMap.get(differentPodInRc1)
	assert.True(t, found)
	assert.Equal(t, cpuErr, err)

	// A non-replicated pod
	nonReplicatedPod := BuildTestPod("nonReplicatedPod", 1000, 1000)
	_, found = pMap.get(nonReplicatedPod)
	assert.False(t, found)
	pMap.set(nonReplicatedPod, err)
	_, found = pMap.get(nonReplicatedPod)
	assert.False(t, found)

	// Verify information about first pod has not been overwritten by adding
	// other pods
	err, found = pMap.get(podInRc1_1)
	assert.True(t, found)
	assert.Nil(t, err)
}

func TestFilterOutExpendableAndSplit(t *testing.T) {
	var priority1 int32 = 1
	var priority100 int32 = 100

	p1 := BuildTestPod("p1", 1000, 200000)
	p1.Spec.Priority = &priority1
	p2 := BuildTestPod("p2", 1000, 200000)
	p2.Spec.Priority = &priority100

	podWaitingForPreemption1 := BuildTestPod("w1", 1000, 200000)
	podWaitingForPreemption1.Spec.Priority = &priority1
	podWaitingForPreemption1.Status.NominatedNodeName = "node1"
	podWaitingForPreemption2 := BuildTestPod("w2", 1000, 200000)
	podWaitingForPreemption2.Spec.Priority = &priority100
	podWaitingForPreemption2.Status.NominatedNodeName = "node1"

	res1, res2 := filterOutExpendableAndSplit([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, 0)
	assert.Equal(t, 2, len(res1))
	assert.Equal(t, p1, res1[0])
	assert.Equal(t, p2, res1[1])
	assert.Equal(t, 2, len(res2))
	assert.Equal(t, podWaitingForPreemption1, res2[0])
	assert.Equal(t, podWaitingForPreemption2, res2[1])

	res1, res2 = filterOutExpendableAndSplit([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, 10)
	assert.Equal(t, 1, len(res1))
	assert.Equal(t, p2, res1[0])
	assert.Equal(t, 1, len(res2))
	assert.Equal(t, podWaitingForPreemption2, res2[0])
}

func TestFilterOutExpendablePods(t *testing.T) {
	p1 := BuildTestPod("p1", 1500, 200000)
	p2 := BuildTestPod("p2", 3000, 200000)

	podWaitingForPreemption1 := BuildTestPod("w1", 1500, 200000)
	var priority1 int32 = -10
	podWaitingForPreemption1.Spec.Priority = &priority1
	podWaitingForPreemption1.Status.NominatedNodeName = "node1"

	podWaitingForPreemption2 := BuildTestPod("w1", 1500, 200000)
	var priority2 int32 = 10
	podWaitingForPreemption2.Spec.Priority = &priority2
	podWaitingForPreemption2.Status.NominatedNodeName = "node1"

	res := filterOutExpendablePods([]*apiv1.Pod{p1, p2, podWaitingForPreemption1, podWaitingForPreemption2}, 0)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, p1, res[0])
	assert.Equal(t, p2, res[1])
	assert.Equal(t, podWaitingForPreemption2, res[2])
}

func TestGetNodeInfosForGroups(t *testing.T) {
	ready1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(ready1, true, time.Now())
	ready2 := BuildTestNode("n2", 2000, 2000)
	SetNodeReadyState(ready2, true, time.Now())
	unready3 := BuildTestNode("n3", 3000, 3000)
	SetNodeReadyState(unready3, false, time.Now())
	unready4 := BuildTestNode("n4", 4000, 4000)
	SetNodeReadyState(unready4, false, time.Now())

	tn := BuildTestNode("tn", 5000, 5000)
	tni := schedulerframework.NewNodeInfo()
	tni.SetNode(tn)
	tn1 := schedulerframework.NewNodeInfo()
	tn1.SetNode(ready1)
	tn2 := schedulerframework.NewNodeInfo()
	tn2.SetNode(ready2)

	// Cloud provider with TemplateNodeInfo implemented.
	provider1 := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil, nil, nil, nil,
		map[string]*schedulerframework.NodeInfo{"ng3": tni, "ng4": tni, "ng2": tn2, "ng1": tn1})
	provider1.AddNodeGroup("ng1", 1, 10, 1) // Nodegroup with ready node.
	provider1.AddNode("ng1", ready1)
	provider1.AddNodeGroup("ng2", 1, 10, 1) // Nodegroup with ready and unready node.
	provider1.AddNode("ng2", ready2)
	provider1.AddNode("ng2", unready3)
	provider1.AddNodeGroup("ng3", 1, 10, 1) // Nodegroup with unready node.
	provider1.AddNode("ng3", unready4)
	provider1.AddNodeGroup("ng4", 0, 1000, 0) // Nodegroup without nodes.

	// Cloud provider with TemplateNodeInfo not implemented.
	provider2 := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil, nil, nil, nil, nil)
	provider2.AddNodeGroup("ng5", 1, 10, 1) // Nodegroup without nodes.

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	predicateChecker, err := simulator.NewTestPredicateChecker()
	assert.NoError(t, err)

	res, err := getNodeInfosForGroups([]*apiv1.Node{unready4, unready3, ready2, ready1}, nil,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(res))
	info, found := res["ng1"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready1, info.Node())
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	info, found = res["ng3"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())

	// Test for a nodegroup without nodes and TemplateNodeInfo not implemented by cloud provider
	res, err = getNodeInfosForGroups([]*apiv1.Node{}, nil, provider2, registry,
		[]*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))
}

func TestGetNodeInfosForGroupsCache(t *testing.T) {
	ready1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(ready1, true, time.Now())
	ready2 := BuildTestNode("n2", 2000, 2000)
	SetNodeReadyState(ready2, true, time.Now())
	unready3 := BuildTestNode("n3", 3000, 3000)
	SetNodeReadyState(unready3, false, time.Now())
	unready4 := BuildTestNode("n4", 4000, 4000)
	SetNodeReadyState(unready4, false, time.Now())
	ready5 := BuildTestNode("n5", 5000, 5000)
	SetNodeReadyState(ready5, true, time.Now())
	ready6 := BuildTestNode("n6", 6000, 6000)
	SetNodeReadyState(ready6, true, time.Now())

	tn := BuildTestNode("tn", 10000, 10000)
	tni := schedulerframework.NewNodeInfo()
	tni.SetNode(tn)
	tn1 := schedulerframework.NewNodeInfo()
	tn1.SetNode(ready1)
	tn2 := schedulerframework.NewNodeInfo()
	tn2.SetNode(ready2)

	lastDeletedGroup := ""
	onDeleteGroup := func(id string) error {
		lastDeletedGroup = id
		return nil
	}

	// Cloud provider with TemplateNodeInfo implemented.
	provider1 := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil, nil, onDeleteGroup, nil,
		map[string]*schedulerframework.NodeInfo{"ng3": tni, "ng4": tni, "ng2": tn2, "ng1": tn1})
	provider1.AddNodeGroup("ng1", 1, 10, 1) // Nodegroup with ready node.
	provider1.AddNode("ng1", ready1)
	provider1.AddNodeGroup("ng2", 1, 10, 1) // Nodegroup with ready and unready node.
	provider1.AddNode("ng2", ready2)
	provider1.AddNode("ng2", unready3)
	provider1.AddNodeGroup("ng3", 1, 10, 1) // Nodegroup with unready node (and 1 previously ready node).
	provider1.AddNode("ng3", unready4)
	provider1.AddNode("ng3", ready5)
	provider1.AddNodeGroup("ng4", 0, 1000, 0) // Nodegroup without nodes (and 1 previously ready node).
	provider1.AddNode("ng4", ready6)

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	predicateChecker, err := simulator.NewTestPredicateChecker()
	assert.NoError(t, err)

	nodeInfoCache := make(map[string]cacheItem)

	// Fill cache
	res, err := getNodeInfosForGroups([]*apiv1.Node{unready4, unready3, ready2, ready1}, nodeInfoCache,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	// Check results
	assert.Equal(t, 4, len(res))
	info, found := res["ng1"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready1, info.Node())
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	info, found = res["ng3"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	// Check cache
	cachedInfo, found := nodeInfoCache["ng1"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready1, cachedInfo.Node())
	cachedInfo, found = nodeInfoCache["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, cachedInfo.Node())
	_, found = nodeInfoCache["ng3"]
	assert.False(t, found)
	_, found = nodeInfoCache["ng4"]
	assert.False(t, found)

	// Invalidate part of cache in two different ways
	provider1.DeleteNodeGroup("ng1")
	err = provider1.GetNodeGroup("ng3").Delete()
	if err != nil {
		t.Logf("Delete NodeGroup failed. Error: %v", err)
	}
	assert.Equal(t, "ng3", lastDeletedGroup)

	// Check cache with all nodes removed
	res, err = getNodeInfosForGroups([]*apiv1.Node{}, nodeInfoCache,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	// Check results
	assert.Equal(t, 2, len(res))
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	// Check ng4 result and cache
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	// Check cache
	cachedInfo, found = nodeInfoCache["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, cachedInfo.Node())
	_, found = nodeInfoCache["ng4"]
	assert.False(t, found)

	// Fill cache manually
	infoNg4Node6 := schedulerframework.NewNodeInfo()
	infoNg4Node6.SetNode(ready6.DeepCopy())
	nodeInfoCache = map[string]cacheItem{"ng4": {NodeInfo: infoNg4Node6, added: time.Now()}}
	// Check if cache was used
	res, err = getNodeInfosForGroups([]*apiv1.Node{ready1, ready2}, nodeInfoCache,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready6, info.Node())
}

func assertEqualNodeCapacities(t *testing.T, expected, actual *apiv1.Node) {
	t.Helper()
	assert.Equal(t, getNodeResource(expected, apiv1.ResourceCPU), getNodeResource(actual, apiv1.ResourceCPU),
		"CPU should be the same")
	assert.Equal(t, getNodeResource(expected, apiv1.ResourceMemory), getNodeResource(actual, apiv1.ResourceMemory),
		"Memory should be the same")
}

func TestRemoveOldUnregisteredNodes(t *testing.T) {
	deletedNodes := make(chan string, 10)

	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_1.Spec.ProviderID = "ng1-1"
	ng1_1.Status.Addresses = []apiv1.NodeAddress{
		{
			Type:    apiv1.NodeInternalIP,
			Address: "ng1-1",
		},
	}
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	ng1_2.Spec.ProviderID = "ng1-2"
	ng1_2.Status.Addresses = []apiv1.NodeAddress{
		{
			Type:    apiv1.NodeInternalIP,
			Address: "ng1-2",
		},
	}
	provider := testprovider.NewTestCloudProvider(nil, func(nodegroup string, node string) error {
		deletedNodes <- fmt.Sprintf("%s/%s", nodegroup, node)
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff())
	err := clusterState.UpdateNodes([]*apiv1.Node{ng1_1}, nil, now.Add(-time.Hour))
	assert.NoError(t, err)

	context := &context.AutoscalingContext{
		AutoscalingOptions: config.AutoscalingOptions{
			MaxNodeProvisionTime: 45 * time.Minute,
		},
		CloudProvider: provider,
	}
	unregisteredNodes := clusterState.GetUnregisteredNodes()
	assert.Equal(t, 1, len(unregisteredNodes))

	// Nothing should be removed. The unregistered node is not old enough.
	removed, err := removeOldUnregisteredNodes(unregisteredNodes, context, clusterState, now.Add(-50*time.Minute), fakeLogRecorder)
	assert.NoError(t, err)
	assert.False(t, removed)

	// ng1_2 should be removed.
	removed, err = removeOldUnregisteredNodes(unregisteredNodes, context, clusterState, now, fakeLogRecorder)
	assert.NoError(t, err)
	assert.True(t, removed)
	deletedNode := core_utils.GetStringFromChan(deletedNodes)
	assert.Equal(t, "ng1/ng1-2", deletedNode)
}

func TestSanitizeNodeInfo(t *testing.T) {
	pod := BuildTestPod("p1", 80, 0)
	pod.Spec.NodeName = "n1"

	node := BuildTestNode("node", 1000, 1000)

	nodeInfo := schedulerframework.NewNodeInfo(pod)
	nodeInfo.SetNode(node)

	res, err := sanitizeNodeInfo(nodeInfo, "test-group", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Pods))
}

func TestSanitizeLabels(t *testing.T) {
	oldNode := BuildTestNode("ng1-1", 1000, 1000)
	oldNode.Labels = map[string]string{
		apiv1.LabelHostname: "abc",
		"x":                 "y",
	}
	node, err := sanitizeTemplateNode(oldNode, "bzium", nil)
	assert.NoError(t, err)
	assert.NotEqual(t, node.Labels[apiv1.LabelHostname], "abc", nil)
	assert.Equal(t, node.Labels["x"], "y")
	assert.NotEqual(t, node.Name, oldNode.Name)
	assert.Equal(t, node.Labels[apiv1.LabelHostname], node.Name)
}

func TestSanitizeTaints(t *testing.T) {
	oldNode := BuildTestNode("ng1-1", 1000, 1000)
	taints := make([]apiv1.Taint, 0)
	taints = append(taints, apiv1.Taint{
		Key:    ReschedulerTaintKey,
		Value:  "test1",
		Effect: apiv1.TaintEffectNoSchedule,
	})
	taints = append(taints, apiv1.Taint{
		Key:    "test-taint",
		Value:  "test2",
		Effect: apiv1.TaintEffectNoSchedule,
	})
	taints = append(taints, apiv1.Taint{
		Key:    deletetaint.ToBeDeletedTaint,
		Value:  "1",
		Effect: apiv1.TaintEffectNoSchedule,
	})
	taints = append(taints, apiv1.Taint{
		Key:    "ignore-me",
		Value:  "1",
		Effect: apiv1.TaintEffectNoSchedule,
	})
	taints = append(taints, apiv1.Taint{
		Key:    "node.kubernetes.io/memory-pressure",
		Value:  "1",
		Effect: apiv1.TaintEffectNoSchedule,
	})

	ignoredTaints := map[string]bool{"ignore-me": true}

	oldNode.Spec.Taints = taints
	node, err := sanitizeTemplateNode(oldNode, "bzium", ignoredTaints)
	assert.NoError(t, err)
	assert.Equal(t, len(node.Spec.Taints), 1)
	assert.Equal(t, node.Spec.Taints[0].Key, "test-taint")
}

func TestRemoveFixNodeTargetSize(t *testing.T) {
	sizeChanges := make(chan string, 10)
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_1.Spec.ProviderID = "ng1-1"
	provider := testprovider.NewTestCloudProvider(func(nodegroup string, delta int) error {
		sizeChanges <- fmt.Sprintf("%s/%d", nodegroup, delta)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 3)
	provider.AddNode("ng1", ng1_1)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false, "my-cool-configmap")
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder, newBackoff())
	err := clusterState.UpdateNodes([]*apiv1.Node{ng1_1}, nil, now.Add(-time.Hour))
	assert.NoError(t, err)

	context := &context.AutoscalingContext{
		AutoscalingOptions: config.AutoscalingOptions{
			MaxNodeProvisionTime: 45 * time.Minute,
		},
		CloudProvider: provider,
	}

	// Nothing should be fixed. The incorrect size state is not old enough.
	removed, err := fixNodeGroupSize(context, clusterState, now.Add(-50*time.Minute))
	assert.NoError(t, err)
	assert.False(t, removed)

	// Node group should be decreased.
	removed, err = fixNodeGroupSize(context, clusterState, now)
	assert.NoError(t, err)
	assert.True(t, removed)
	change := core_utils.GetStringFromChan(sizeChanges)
	assert.Equal(t, "ng1/-2", change)
}

func TestGetNodeResource(t *testing.T) {
	node := BuildTestNode("n1", 1000, 2*MiB)

	cores := getNodeResource(node, apiv1.ResourceCPU)
	assert.Equal(t, int64(1), cores)

	memory := getNodeResource(node, apiv1.ResourceMemory)
	assert.Equal(t, int64(2*MiB), memory)

	unknownResourceValue := getNodeResource(node, "unknown resource")
	assert.Equal(t, int64(0), unknownResourceValue)

	// if we have no resources in capacity we expect getNodeResource to return 0
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

	// if we have negative values in resources we expect getNodeResource to return 0
	nodeWithNegativeCapacity := BuildTestNode("n1", -1000, -2*MiB)
	nodeWithNegativeCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

}

func TestGetNodeCoresAndMemory(t *testing.T) {
	node := BuildTestNode("n1", 2000, 2048*MiB)

	cores, memory := getNodeCoresAndMemory(node)
	assert.Equal(t, int64(2), cores)
	assert.Equal(t, int64(2048*MiB), memory)

	// if we have no cpu/memory defined in capacity we expect getNodeCoresAndMemory to return 0s
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores, memory = getNodeCoresAndMemory(nodeWithMissingCapacity)
	assert.Equal(t, int64(0), cores)
	assert.Equal(t, int64(0), memory)

	// filter eklet node
	ekletNode := BuildTestNode("n1", 2000, 2048*MiB)
	ekletNode.Labels[nodeInstanceTypeLabelKey] = nodeInstanceTypeEklet
	cores, memory = getNodeCoresAndMemory(ekletNode)
	assert.Equal(t, int64(0), cores)
	assert.Equal(t, int64(0), memory)

	// filer specific node
	nodeWithAnno := BuildTestNode("n1", 2000, 2048*MiB)
	nodeWithAnno.Annotations = map[string]string{}
	nodeWithAnno.Annotations[filterNodeResourceAnnoKey] = "true"
	cores, memory = getNodeCoresAndMemory(nodeWithAnno)
	assert.Equal(t, int64(0), cores)
	assert.Equal(t, int64(0), memory)
}

func TestGetOldestPod(t *testing.T) {
	p1 := BuildTestPod("p1", 500, 1000)
	p1.CreationTimestamp = metav1.NewTime(time.Now().Add(-1 * time.Minute))
	p2 := BuildTestPod("p2", 500, 1000)
	p2.CreationTimestamp = metav1.NewTime(time.Now().Add(+1 * time.Minute))
	p3 := BuildTestPod("p3", 500, 1000)
	p3.CreationTimestamp = metav1.NewTime(time.Now())

	assert.Equal(t, p1.CreationTimestamp.Time, getOldestCreateTime([]*apiv1.Pod{p1, p2, p3}))
	assert.Equal(t, p1.CreationTimestamp.Time, getOldestCreateTime([]*apiv1.Pod{p3, p2, p1}))
}
