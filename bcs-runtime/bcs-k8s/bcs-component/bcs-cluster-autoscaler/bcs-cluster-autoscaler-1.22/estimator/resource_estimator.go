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

// Package estimator provides the extend estimator, ClusterResourceEstimator
package estimator

import (
	"fmt"
	"math"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	klog "k8s.io/klog/v2"
	nodeLabel "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// filterNodeResourceAnnoKey filters nodes when calculating buffer and total resource
	filterNodeResourceAnnoKey = "io.tencent.bcs.dev/filter-node-resource"

	valueTrue = "true"
)

// podInfo contains Pod and score that corresponds to how important it is to handle the pod first.
type podInfo struct {
	score float64
	pod   *corev1.Pod
}

// ClusterResourceEstimator estimates the number of needed nodes to handle the given amount of pods.
type ClusterResourceEstimator struct {
	predicateChecker simulator.PredicateChecker
	clusterSnapshot  simulator.ClusterSnapshot
	cpuRatio         float64
	memRatio         float64
	resourceRatio    float64
}

// NewClusterResourceEstimator builds a new ClusterResourceEstimator.
func NewClusterResourceEstimator(predicateChecker simulator.PredicateChecker,
	clusterSnapshot simulator.ClusterSnapshot,
	cpuRatio, memRatio, resourceRatio float64) *ClusterResourceEstimator {
	return &ClusterResourceEstimator{
		predicateChecker: predicateChecker,
		clusterSnapshot:  clusterSnapshot,
		cpuRatio:         cpuRatio,
		memRatio:         memRatio,
		resourceRatio:    resourceRatio,
	}
}

// Estimate implements First Fit Decreasing bin-packing approximation algorithm.
// See https://en.wikipedia.org/wiki/Bin_packing_problem for more details.
// While it is a multi-dimensional bin packing (cpu, mem, ports) in most cases the main dimension
// will be cpu thus the estimated overprovisioning of 11/9 * optimal + 6/9 should be
// still be maintained.
// It is assumed that all pods from the given list can fit to nodeTemplate.
// Returns the number of nodes needed to accommodate all pods from the list.
func (e *ClusterResourceEstimator) Estimate(
	pods []*corev1.Pod,
	nodeTemplate *schedulerframework.NodeInfo) int {
	podInfos := calculatePodScore(pods, nodeTemplate)
	sort.Slice(podInfos, func(i, j int) bool { return podInfos[i].score > podInfos[j].score })

	newNodeNames := make(map[string]bool)

	if err := e.clusterSnapshot.Fork(); err != nil {
		klog.Errorf("Error while calling ClusterSnapshot.Fork; %v", err)
		return 0
	}
	defer func() {
		if err := e.clusterSnapshot.Revert(); err != nil {
			klog.Fatalf("Error while calling ClusterSnapshot.Revert; %v", err)
		}
	}()

	newNodeNameIndex := 0

	for _, podInfo := range podInfos {
		found := false

		nodeName, err := e.predicateChecker.FitsAnyNodeMatching(e.clusterSnapshot, podInfo.pod,
			func(nodeInfo *schedulerframework.NodeInfo) bool {
				return newNodeNames[nodeInfo.Node().Name]
			})
		if err == nil {
			found = true
			if err := e.clusterSnapshot.AddPod(podInfo.pod, nodeName); err != nil {
				klog.Errorf("Error adding pod %v.%v to node %v in ClusterSnapshot; %v",
					podInfo.pod.Namespace, podInfo.pod.Name, nodeName, err)
				return 0
			}
		}

		if !found {
			// Add new node
			newNodeName, err := e.addNewNodeToSnapshot(nodeTemplate, newNodeNameIndex)
			if err != nil {
				klog.Errorf("Error while adding new node for template to ClusterSnapshot; %v", err)
				return 0
			}
			newNodeNameIndex++
			// And schedule pod to it
			if err := e.clusterSnapshot.AddPod(podInfo.pod, newNodeName); err != nil {
				klog.Errorf("Error adding pod %v.%v to node %v in ClusterSnapshot; %v",
					podInfo.pod.Namespace, podInfo.pod.Name, newNodeName, err)
				return 0
			}
			newNodeNames[newNodeName] = true
		}
	}

	nodeNumPredict := len(newNodeNames)
	nodeNumLoad := e.estimateAccordingToLoad(nodeTemplate)
	klog.Infof("nodeNumLoad: %v nodeNumPredict: %v", nodeNumLoad, nodeNumPredict)
	return nodeNumPredict + int(nodeNumLoad)
}

func (e *ClusterResourceEstimator) estimateAccordingToLoad(nodeTemplate *schedulerframework.NodeInfo) float64 {
	sumResourcesList := &schedulerframework.Resource{}
	leftResourcesList := &schedulerframework.Resource{}
	nodes, err := e.clusterSnapshot.NodeInfos().List()
	if err != nil {
		klog.Errorf("Error list node infos from ClusterSnapshot; %v", err)
		return 0
	}
	for _, nodeInfo := range nodes {
		node := nodeInfo.Node()
		if node == nil {
			continue
		}
		if node.Spec.Unschedulable {
			continue
		}
		if node.Labels[corev1.LabelInstanceTypeStable] == "eklet" {
			continue
		}
		if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
			continue
		}
		if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
			continue
		}
		if node.Labels[nodeLabel.LabelNodeRoleControlPlane] == valueTrue {
			continue
		}
		sumResourcesList.Add(scheduler.ResourceToResourceList(nodeInfo.Allocatable))
		leftResourcesList.Add(singleNodeResource(nodeInfo))
	}

	var nodeNum float64
	leftResources := scheduler.ResourceToResourceList(leftResourcesList)
	sumResources := scheduler.ResourceToResourceList(sumResourcesList)
	for name, sum := range sumResources {
		left, ok := leftResources[name]
		if !ok {
			continue
		}
		if sum.IsZero() {
			continue
		}
		var tmpNode float64
		switch name {
		case corev1.ResourceCPU:
			tmpNode = computeNewNodeNumWithRatio(nodeTemplate, name, sum, left, e.cpuRatio)
		case corev1.ResourceMemory:
			tmpNode = computeNewNodeNumWithRatio(nodeTemplate, name, sum, left, e.memRatio)
		default:
			tmpNode = computeNewNodeNumWithRatio(nodeTemplate, name, sum, left, e.resourceRatio)
		}

		if tmpNode > nodeNum {
			nodeNum = tmpNode
		}

	}
	return nodeNum
}

func computeNewNodeNumWithRatio(nodeTemplate *schedulerframework.NodeInfo, name corev1.ResourceName,
	sum, left resource.Quantity, ratio float64) float64 {
	if math.Abs(ratio) < 1e-9 {
		klog.V(4).Infof("the ratio of %s is 0, cannot calculate desired node num", name)
		return 0
	}
	var num float64
	r := float64(left.MilliValue()) / float64(sum.MilliValue())
	// (left + num)/ (sum + num) >= ratio
	resourceRequest := (float64(sum.MilliValue())*ratio - float64(left.MilliValue())) / (1 - ratio)
	// num >= 0
	resourceRequest = math.Max(0, resourceRequest)
	nodeResource := nodeTemplate.Allocatable
	nodeResourceList := scheduler.ResourceToResourceList(nodeResource)
	nodeCapacity, ok := nodeResourceList[name]
	if !ok {
		klog.V(4).Infof("resource %v is not in nodeTemplate's capacity, don't scale up nodes", name)
		num = 0
	} else {
		num = math.Ceil(resourceRequest / float64(nodeCapacity.MilliValue()))
	}

	klog.V(4).Infof("resource: %v, sum: %v, left: %v, desired-ratio: %v, current-ratio: %v, desired-node: %v",
		name, sum.MilliValue(), left.MilliValue(), ratio, r, num)
	return math.Max(0, num)
}

func (e *ClusterResourceEstimator) addNewNodeToSnapshot(
	template *schedulerframework.NodeInfo,
	nameIndex int) (string, error) {

	newNodeInfo := scheduler.DeepCopyTemplateNode(template, fmt.Sprintf("estimator-%d", nameIndex))
	var pods []*corev1.Pod
	for _, podInfo := range newNodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}
	if err := e.clusterSnapshot.AddNodeWithPods(newNodeInfo.Node(), pods); err != nil {
		return "", err
	}
	return newNodeInfo.Node().Name, nil
}

func singleNodeResource(info *schedulerframework.NodeInfo) corev1.ResourceList {
	leftResource := &schedulerframework.Resource{
		ScalarResources: make(map[corev1.ResourceName]int64),
	}

	allocatable := info.Allocatable
	requested := info.Requested

	podCount := requested.AllowedPodNumber
	if podCount == 0 {
		podCount = len(info.Pods)
	}

	leftResource.AllowedPodNumber = allocatable.AllowedPodNumber - podCount
	leftResource.MilliCPU = allocatable.MilliCPU - requested.MilliCPU
	leftResource.Memory = allocatable.Memory - requested.Memory
	leftResource.EphemeralStorage = allocatable.EphemeralStorage - requested.EphemeralStorage

	// calculate extend resources
	for k, allocatableEx := range allocatable.ScalarResources {
		requestEx, ok := requested.ScalarResources[k]
		if !ok {
			leftResource.ScalarResources[k] = allocatableEx
		} else {
			leftResource.ScalarResources[k] = allocatableEx - requestEx
		}
	}
	klog.Infof("Node %v left resource %+v", info.Node().Name, leftResource)
	return scheduler.ResourceToResourceList(leftResource)
}

// Calculates score for all pods and returns podInfo structure.
// Score is defined as cpu_sum/node_capacity + mem_sum/node_capacity.
// Pods that have bigger requirements should be processed first, thus have higher scores.
func calculatePodScore(pods []*corev1.Pod, nodeTemplate *schedulerframework.NodeInfo) []*podInfo {
	podInfos := make([]*podInfo, 0, len(pods))

	for _, pod := range pods {
		cpuSum := resource.Quantity{}
		memorySum := resource.Quantity{}

		for _, container := range pod.Spec.Containers {
			if request, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				cpuSum.Add(request)
			}
			if request, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				memorySum.Add(request)
			}
		}
		score := float64(0)
		if cpuAllocatable, ok := nodeTemplate.Node().Status.Allocatable[corev1.ResourceCPU]; ok &&
			cpuAllocatable.MilliValue() > 0 {
			score += float64(cpuSum.MilliValue()) / float64(cpuAllocatable.MilliValue())
		}
		if memAllocatable, ok := nodeTemplate.Node().Status.Allocatable[corev1.ResourceMemory]; ok &&
			memAllocatable.Value() > 0 {
			score += float64(memorySum.Value()) / float64(memAllocatable.Value())
		}

		podInfos = append(podInfos, &podInfo{
			score: score,
			pod:   pod,
		})
	}
	return podInfos
}
