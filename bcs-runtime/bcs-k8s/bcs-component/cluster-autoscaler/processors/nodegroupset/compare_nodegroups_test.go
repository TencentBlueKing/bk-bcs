/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodegroupset

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/stretchr/testify/assert"
)

func checkNodesSimilar(t *testing.T, n1, n2 *apiv1.Node, comparator NodeInfoComparator, shouldEqual bool) {
	checkNodesSimilarWithPods(t, n1, n2, []*apiv1.Pod{}, []*apiv1.Pod{}, comparator, shouldEqual)
}

func checkNodesSimilarWithPods(t *testing.T, n1, n2 *apiv1.Node, pods1, pods2 []*apiv1.Pod, comparator NodeInfoComparator, shouldEqual bool) {
	ni1 := schedulernodeinfo.NewNodeInfo(pods1...)
	ni1.SetNode(n1)
	ni2 := schedulernodeinfo.NewNodeInfo(pods2...)
	ni2.SetNode(n2)
	assert.Equal(t, shouldEqual, comparator(ni1, ni2))
}

func TestIdenticalNodesSimilar(t *testing.T) {
	n1 := BuildTestNode("node1", 1000, 2000)
	n2 := BuildTestNode("node2", 1000, 2000)
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, true)
}

func TestNodesSimilarVariousRequirements(t *testing.T) {
	n1 := BuildTestNode("node1", 1000, 2000)

	// Different CPU capacity
	n2 := BuildTestNode("node2", 1000, 2000)
	n2.Status.Capacity[apiv1.ResourceCPU] = *resource.NewMilliQuantity(1001, resource.DecimalSI)
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, false)

	// Same CPU capacity, but slightly different allocatable
	n3 := BuildTestNode("node3", 1000, 2000)
	n3.Status.Allocatable[apiv1.ResourceCPU] = *resource.NewMilliQuantity(999, resource.DecimalSI)
	checkNodesSimilar(t, n1, n3, IsNodeInfoSimilar, true)

	// Same CPU capacity, significantly different allocatable
	n4 := BuildTestNode("node4", 1000, 2000)
	n4.Status.Allocatable[apiv1.ResourceCPU] = *resource.NewMilliQuantity(500, resource.DecimalSI)
	checkNodesSimilar(t, n1, n4, IsNodeInfoSimilar, false)

	// One with GPU, one without
	n5 := BuildTestNode("node5", 1000, 2000)
	n5.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	n5.Status.Allocatable[gpu.ResourceNvidiaGPU] = n5.Status.Capacity[gpu.ResourceNvidiaGPU]
	checkNodesSimilar(t, n1, n5, IsNodeInfoSimilar, false)
}

func TestNodesSimilarVariousRequirementsAndPods(t *testing.T) {
	n1 := BuildTestNode("node1", 1000, 2000)
	p1 := BuildTestPod("pod1", 500, 1000)
	p1.Spec.NodeName = "node1"

	// Different allocatable, but same free
	n2 := BuildTestNode("node2", 1000, 2000)
	n2.Status.Allocatable[apiv1.ResourceCPU] = *resource.NewMilliQuantity(500, resource.DecimalSI)
	n2.Status.Allocatable[apiv1.ResourceMemory] = *resource.NewQuantity(1000, resource.DecimalSI)
	checkNodesSimilarWithPods(t, n1, n2, []*apiv1.Pod{p1}, []*apiv1.Pod{}, IsNodeInfoSimilar, false)

	// Same requests of pods
	n3 := BuildTestNode("node3", 1000, 2000)
	p3 := BuildTestPod("pod3", 500, 1000)
	p3.Spec.NodeName = "node3"
	checkNodesSimilarWithPods(t, n1, n3, []*apiv1.Pod{p1}, []*apiv1.Pod{p3}, IsNodeInfoSimilar, true)

	// Similar allocatable, similar pods
	n4 := BuildTestNode("node4", 1000, 2000)
	n4.Status.Allocatable[apiv1.ResourceCPU] = *resource.NewMilliQuantity(999, resource.DecimalSI)
	p4 := BuildTestPod("pod4", 501, 1001)
	p4.Spec.NodeName = "node4"
	checkNodesSimilarWithPods(t, n1, n4, []*apiv1.Pod{p1}, []*apiv1.Pod{p4}, IsNodeInfoSimilar, true)
}

func TestNodesSimilarVariousMemoryRequirements(t *testing.T) {
	n1 := BuildTestNode("node1", 1000, MaxMemoryDifferenceInKiloBytes)

	// Different memory capacity within tolerance
	n2 := BuildTestNode("node2", 1000, MaxMemoryDifferenceInKiloBytes)
	n2.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(2*MaxMemoryDifferenceInKiloBytes, resource.DecimalSI)
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, true)

	// Different memory capacity exceeds tolerance
	n3 := BuildTestNode("node3", 1000, MaxMemoryDifferenceInKiloBytes)
	n3.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(2*MaxMemoryDifferenceInKiloBytes+1, resource.DecimalSI)
	checkNodesSimilar(t, n1, n3, IsNodeInfoSimilar, false)
}

func TestNodesSimilarVariousLabels(t *testing.T) {
	n1 := BuildTestNode("node1", 1000, 2000)
	n1.ObjectMeta.Labels["test-label"] = "test-value"
	n1.ObjectMeta.Labels["character"] = "winnie the pooh"

	n2 := BuildTestNode("node2", 1000, 2000)
	n2.ObjectMeta.Labels["test-label"] = "test-value"

	// Missing character label
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, false)

	n2.ObjectMeta.Labels["character"] = "winnie the pooh"
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, true)

	// Different hostname labels shouldn't matter
	n1.ObjectMeta.Labels[apiv1.LabelHostname] = "node1"
	n2.ObjectMeta.Labels[apiv1.LabelHostname] = "node2"
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, true)

	// Different zone shouldn't matter either
	n1.ObjectMeta.Labels[apiv1.LabelZoneFailureDomain] = "mars-olympus-mons1-b"
	n2.ObjectMeta.Labels[apiv1.LabelZoneFailureDomain] = "us-houston1-a"
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, true)

	// Different beta.kubernetes.io/fluentd-ds-ready should not matter
	n1.ObjectMeta.Labels["beta.kubernetes.io/fluentd-ds-ready"] = "true"
	n2.ObjectMeta.Labels["beta.kubernetes.io/fluentd-ds-ready"] = "false"
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, true)

	n1.ObjectMeta.Labels["beta.kubernetes.io/fluentd-ds-ready"] = "true"
	delete(n2.ObjectMeta.Labels, "beta.kubernetes.io/fluentd-ds-ready")
	checkNodesSimilar(t, n1, n2, IsNodeInfoSimilar, true)
}
