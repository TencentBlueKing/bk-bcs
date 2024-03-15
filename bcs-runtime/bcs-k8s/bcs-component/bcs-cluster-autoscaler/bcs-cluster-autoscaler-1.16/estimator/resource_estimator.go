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
	"math"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	schedulerUtils "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
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
	predicateChecker *simulator.PredicateChecker
	readyNodes       map[string]*schedulernodeinfo.NodeInfo
	cpuRatio         float64
	memRatio         float64
	resourceRatio    float64
}

// NewClusterResourceEstimator builds a new BinpackingNodeEstimator.
func NewClusterResourceEstimator(predicateChecker *simulator.PredicateChecker,
	readyNodes map[string]*schedulernodeinfo.NodeInfo,
	cpuRatio, memRatio, resourceRatio float64) *ClusterResourceEstimator {
	return &ClusterResourceEstimator{
		predicateChecker: predicateChecker,
		readyNodes:       readyNodes,
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
func (estimator *ClusterResourceEstimator) Estimate(pods []*corev1.Pod, nodeTemplate *schedulernodeinfo.NodeInfo,
	upcomingNodes []*schedulernodeinfo.NodeInfo) int {
	podInfos := calculatePodScore(pods, nodeTemplate)
	sort.Slice(podInfos, func(i, j int) bool { return podInfos[i].score > podInfos[j].score })

	newNodes := make([]*schedulernodeinfo.NodeInfo, 0)
	newNodesMap := make(map[string]*schedulernodeinfo.NodeInfo)

	// build fake node name
	for _, node := range upcomingNodes {
		node := node.Node().DeepCopy()
		node.Name = string(uuid.NewUUID())
		node.Labels[corev1.LabelHostname] = node.Name
		nodeInfo := schedulernodeinfo.NewNodeInfo()
		err := nodeInfo.SetNode(node)
		if err != nil {
			klog.Warningf("SetNode failed. Error: %v", err)
		}
		newNodesMap[node.Name] = nodeInfo
		newNodes = append(newNodes, nodeInfo)
	}
	for _, podInfo := range podInfos {
		found := false
		meta := estimator.predicateChecker.GetPredicateMetadata(podInfo.pod, newNodesMap)
		// this may cause bad performance
		for i, nodeInfo := range newNodes {
			klog.Infof("Check pod %v on node %v labels %v, history pods: %v, affinity: %v",
				podInfo.pod.Name, nodeInfo.Node().Name, nodeInfo.Node().Labels,
				len(nodeInfo.Pods()), len(nodeInfo.PodsWithAffinity()))
			if err := estimator.predicateChecker.CheckPredicates(podInfo.pod, meta, nodeInfo); err == nil {
				found = true
				newNodes[i] = schedulerUtils.NodeWithPod(nodeInfo, podInfo.pod)
				newNodesMap[nodeInfo.Node().Name] = newNodes[i]
				break
			} else {
				klog.Infof("Pod %v try on node %v, failed reason: %v, detailed: %+v, errinfo: %v",
					podInfo.pod.Name, nodeInfo.Node().Name,
					err.Reasons(), err.OriginalReasons(), err.VerboseError())
			}
		}
		if !found {
			// do not support affinity/antiaffinity when topoKey is kubernetes.io/region
			// we did not set this in node pool
			if !supportTopoKey(podInfo.pod) {
				klog.Infof("Pod %v has not supported topokeys, do not scale up", podInfo.pod.Name)
				continue
			}
			node := nodeTemplate.Node().DeepCopy()
			node.Name = string(uuid.NewUUID())
			node.Labels[corev1.LabelHostname] = node.Name
			nodeInfo := schedulernodeinfo.NewNodeInfo()
			err := nodeInfo.SetNode(node)
			if err != nil {
				klog.Warningf("SetNode failed. Error: %v", err)
			}
			nodeInfo = schedulerUtils.NodeWithPod(nodeInfo, podInfo.pod)
			newNodesMap[node.Name] = nodeInfo
			newNodes = append(newNodes, nodeInfo)
			klog.Infof("Pod %v requires scaling up node %v", podInfo.pod.Name, node.Name)
		}
	}
	nodeNumPredict := len(newNodes) - len(upcomingNodes)
	nodeNumLoad := estimator.estimateAccordingToLoad(nodeTemplate, estimator.readyNodes, upcomingNodes)
	klog.Infof("nodeNumLoad: %v nodeNumPredict: %v", nodeNumLoad, nodeNumPredict)
	return int(math.Max(nodeNumLoad, float64(nodeNumPredict)))
}

func (estimator *ClusterResourceEstimator) estimateAccordingToLoad(nodeTemplate *schedulernodeinfo.NodeInfo,
	nodes map[string]*schedulernodeinfo.NodeInfo, upcomingNodes []*schedulernodeinfo.NodeInfo) float64 {
	var leftResourcesList, sumResourcesList schedulernodeinfo.Resource
	for _, nodeInfo := range nodes {
		node := nodeInfo.Node()
		if node == nil {
			continue
		}
		if node.Spec.Unschedulable {
			continue
		}
		if node.Labels["node.kubernetes.io/instance-type"] == "eklet" {
			continue
		}
		if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
			continue
		}
		if node.Labels["node-role.kubernetes.io/master"] == valueTrue {
			continue
		}
		allocatable := nodeInfo.AllocatableResource()
		sumResourcesList.Add(allocatable.ResourceList())
		leftResourcesList.Add(singleNodeResource(nodeInfo).ResourceList())
	}

	for _, nodeInfo := range upcomingNodes {
		node := nodeInfo.Node()
		if node == nil {
			continue
		}
		allocatable := nodeInfo.AllocatableResource()
		allocatableResourceList := allocatable.ResourceList()
		sumResourcesList.Add(allocatableResourceList)
		leftResourcesList.Add(allocatableResourceList)
	}

	var nodeNum float64
	leftResources := leftResourcesList.ResourceList()
	sumResources := sumResourcesList.ResourceList()
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
			tmpNode = computeNewNodeNumWithRatio(nodeTemplate, name, sum, left, estimator.cpuRatio)
		case corev1.ResourceMemory:
			tmpNode = computeNewNodeNumWithRatio(nodeTemplate, name, sum, left, estimator.memRatio)
		default:
			tmpNode = computeNewNodeNumWithRatio(nodeTemplate, name, sum, left, estimator.resourceRatio)
		}
		if tmpNode > nodeNum {
			nodeNum = tmpNode
		}
	}
	return nodeNum
}

func computeNewNodeNumWithRatio(nodeTemplate *schedulernodeinfo.NodeInfo, name corev1.ResourceName,
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
	nodeResource := nodeTemplate.AllocatableResource()
	nodeResourceList := nodeResource.ResourceList()
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

func singleNodeResource(info *schedulernodeinfo.NodeInfo) *schedulernodeinfo.Resource {
	leftResource := schedulernodeinfo.Resource{
		ScalarResources: make(map[corev1.ResourceName]int64),
	}

	allocatable := info.AllocatableResource()
	requested := info.RequestedResource()

	podCount := requested.AllowedPodNumber
	if podCount == 0 {
		podCount = len(info.Pods())
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
	return &leftResource
}

// calculatePodScore calculates score for all pods and returns podInfo structure.
// Score is defined as cpu_sum/node_capacity + mem_sum/node_capacity.
// Pods that have bigger requirements should be processed first, thus have higher scores.
func calculatePodScore(pods []*corev1.Pod, nodeTemplate *schedulernodeinfo.NodeInfo) []*podInfo {
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

func supportTopoKey(pod *corev1.Pod) bool {
	if pod.Spec.Affinity == nil {
		return true
	}
	if pod.Spec.Affinity.PodAntiAffinity == nil &&
		pod.Spec.Affinity.PodAffinity == nil {
		return true
	}
	if pod.Spec.Affinity.PodAntiAffinity != nil {
		for _, term := range pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			if term.TopologyKey == corev1.LabelZoneRegion {
				return false
			}
		}
	}
	if pod.Spec.Affinity.PodAffinity != nil {
		for _, term := range pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
			if term.TopologyKey == corev1.LabelZoneRegion {
				return false
			}
		}
	}
	return true
}
