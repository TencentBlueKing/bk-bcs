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

// Package simulator xxx
package simulator

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	simulatorinternal "k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/glogx"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	scheduler_util "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

var (
	skipNodesWithSystemPods   = true
	skipNodesWithLocalStorage = true
	minReplicaCount           = 0
)

// InitFlags extracts configuration from flags
func InitFlags() {
	var err error
	if skipNodesWithSystemPodsFlag := flag.Lookup("skip-nodes-with-system-pods"); skipNodesWithSystemPodsFlag != nil {
		skipNodesWithSystemPods, err = strconv.ParseBool(skipNodesWithSystemPodsFlag.Value.String())
		if err != nil {
			panic("Parse parameters error")
		}
		klog.Infof("skipNodesWithSystemPodsFlag: %+v", skipNodesWithSystemPods)
	}
	if skipNodesWithLocalStorageFlag := flag.Lookup(
		"skip-nodes-with-local-storage"); skipNodesWithLocalStorageFlag != nil {
		skipNodesWithLocalStorage, err = strconv.ParseBool(skipNodesWithLocalStorageFlag.Value.String())
		if err != nil {
			panic("Parse parameters error")
		}
		klog.Infof("skipNodesWithLocalStorageFlag: %+v", skipNodesWithLocalStorage)
	}
	if minReplicaCountFlag := flag.Lookup("min-replica-count"); minReplicaCountFlag != nil {
		minReplicaCount, err = strconv.Atoi(minReplicaCountFlag.Value.String())
		if err != nil {
			panic("Parse parameters error")
		}
	}
}

// NodeToBeRemoved contain information about a node that can be removed.
type NodeToBeRemoved struct {
	// Node to be removed.
	Node *apiv1.Node
	// PodsToReschedule contains pods on the node that should be rescheduled elsewhere.
	PodsToReschedule []*apiv1.Pod
}

// UnremovableNode represents a node that can't be removed by CA.
type UnremovableNode struct {
	Node        *apiv1.Node
	Reason      UnremovableReason
	BlockingPod *BlockingPod
}

// UnremovableReason represents a reason why a node can't be removed by CA.
type UnremovableReason string

const (
	// NoReason - sanity check, this should never be set explicitly. If this is found in the wild, it means that it was
	// implicitly initialized and might indicate a bug.
	NoReason UnremovableReason = "NoReason"
	// ScaleDownDisabledAnnotation - node can't be removed because it has a "scale down disabled" annotation.
	ScaleDownDisabledAnnotation UnremovableReason = "ScaleDownDisabledAnnotation"
	// NotAutoscaled - node can't be removed because it doesn't belong to an autoscaled node group.
	NotAutoscaled UnremovableReason = "NotAutoscaled"
	// NotUnneededLongEnough - node can't be removed because it wasn't unneeded for long enough.
	NotUnneededLongEnough UnremovableReason = "NotUnneededLongEnough"
	// NotUnreadyLongEnough - node can't be removed because it wasn't unready for long enough.
	NotUnreadyLongEnough UnremovableReason = "NotUnreadyLongEnough"
	// NodeGroupMinSizeReached - node can't be removed because its node group is at its minimal size already.
	NodeGroupMinSizeReached UnremovableReason = "NodeGroupMinSizeReached"
	// MinimalResourceLimitExceeded - node can't be removed because it would violate cluster-wide minimal resource limits.
	MinimalResourceLimitExceeded UnremovableReason = "MinimalResourceLimitExceeded"
	// CurrentlyBeingDeleted - node can't be removed because it's already in the process of being deleted.
	CurrentlyBeingDeleted UnremovableReason = "CurrentlyBeingDeleted"
	// NotUnderutilized - node can't be removed because it's not underutilized.
	NotUnderutilized UnremovableReason = "NotUnderutilized"
	// NotUnneededOtherReason - node can't be removed because it's not marked as unneeded for other reasons
	// (e.g. it wasn't inspected at all in a given autoscaler loop).
	NotUnneededOtherReason UnremovableReason = "NotUnneededOtherReason"
	// RecentlyUnremovable - node can't be removed because it was recently found to be unremovable.
	RecentlyUnremovable UnremovableReason = "RecentlyUnremovable"
	// NoPlaceToMovePods - node can't be removed because there's no place to move its pods to.
	NoPlaceToMovePods UnremovableReason = "NoPlaceToMovePods"
	// BlockedByPod - node can't be removed because a pod running on it can't be moved. The reason why
	// should be in BlockingPod.
	BlockedByPod UnremovableReason = "BlockedByPod"
	// BufferNotEnough - node can't be removed because of buffer ratio
	BufferNotEnough UnremovableReason = "BufferNotEnough"
	// UnexpectedError - node can't be removed because of an unexpected error.
	UnexpectedError UnremovableReason = "UnexpectedError"
)

// FindNodesToRemove finds nodes that can be removed. Returns also an information about good
// rescheduling location for each of the pods.
func FindNodesToRemove(candidates []*apiv1.Node, destinationNodes []*apiv1.Node, pods []*apiv1.Pod,
	listers kube_util.ListerRegistry, predicateChecker *simulatorinternal.PredicateChecker, maxCount int,
	fastCheck bool, oldHints map[string]string, usageTracker *simulatorinternal.UsageTracker,
	timestamp time.Time,
	podDisruptionBudgets []*policyv1.PodDisruptionBudget,
) (nodesToRemove []NodeToBeRemoved, unremovableNodes []*UnremovableNode, podReschedulingHints map[string]string,
	finalError errors.AutoscalerError) {
	klog.V(4).Infof("Will skip Nodes With Local Storage: %v", skipNodesWithLocalStorage)
	nodeNameToNodeInfo := scheduler_util.CreateNodeNameToInfoMap(pods, destinationNodes)
	result := make([]NodeToBeRemoved, 0)
	unremovable := make([]*UnremovableNode, 0)

	evaluationType := "Detailed evaluation"
	if fastCheck {
		evaluationType = "Fast evaluation"
	}
	newHints := make(map[string]string, len(oldHints))

candidateloop:
	for _, node := range candidates {
		klog.V(2).Infof("%s: %s for removal", evaluationType, node.Name)

		var podsToRemove []*apiv1.Pod
		var blockingPod *BlockingPod
		var err error

		nodeInfo, found := nodeNameToNodeInfo[node.Name]
		// nolint
		if found {
			if fastCheck {
				podsToRemove, _, blockingPod, err = FastGetPodsToMove(nodeInfo, skipNodesWithSystemPods,
					skipNodesWithLocalStorage, podDisruptionBudgets)
			} else {
				podsToRemove, _, blockingPod, err = DetailedGetPodsForMove(nodeInfo, skipNodesWithSystemPods,
					skipNodesWithLocalStorage, listers, int32(minReplicaCount), podDisruptionBudgets)
			}
			if err != nil {
				klog.V(2).Infof("%s: node %s cannot be removed: %v", evaluationType, node.Name, err)
				if blockingPod != nil {
					unremovable = append(unremovable, &UnremovableNode{Node: nodeInfo.Node(), Reason: BlockedByPod,
						BlockingPod: blockingPod})
				} else {
					unremovable = append(unremovable, &UnremovableNode{Node: nodeInfo.Node(), Reason: UnexpectedError})
				}
				continue candidateloop
			}
		} else {
			klog.V(2).Infof("%s: nodeInfo for %s not found", evaluationType, node.Name)
			unremovable = append(unremovable, &UnremovableNode{Node: node, Reason: UnexpectedError})
			continue candidateloop
		}
		findProblems := findPlaceFor(node.Name, podsToRemove, destinationNodes, nodeNameToNodeInfo,
			predicateChecker, oldHints, newHints, usageTracker, timestamp)

		if findProblems == nil {
			result = append(result, NodeToBeRemoved{
				Node:             node,
				PodsToReschedule: podsToRemove,
			})
			klog.V(2).Infof("%s: node %s may be removed", evaluationType, node.Name)
			if len(result) >= maxCount {
				break candidateloop
			}
		} else {
			klog.V(2).Infof("%s: node %s is not suitable for removal: %v", evaluationType, node.Name, findProblems)
			unremovable = append(unremovable, &UnremovableNode{Node: nodeInfo.Node(), Reason: NoPlaceToMovePods})
		}
	}
	return result, unremovable, newHints, nil
}

// FindEmptyNodesToRemove finds empty nodes that can be removed.
func FindEmptyNodesToRemove(candidates []*apiv1.Node, pods []*apiv1.Pod) []*apiv1.Node {
	nodeNameToNodeInfo := scheduler_util.CreateNodeNameToInfoMap(pods, candidates)
	result := make([]*apiv1.Node, 0)
	for _, node := range candidates {
		if nodeInfo, found := nodeNameToNodeInfo[node.Name]; found {
			// Should block on all pods.
			podsToRemove, _, _, err := FastGetPodsToMove(nodeInfo, true, true, nil)
			if err == nil && len(podsToRemove) == 0 {
				result = append(result, node)
			}
		} else {
			// Node without pods.
			result = append(result, node)
		}
	}
	return result
}

// findPlaceFor xxx
// DOTO: We don't need to pass list of nodes here as they are already available in nodeInfos.
// nolint funlen
func findPlaceFor(removedNode string, pods []*apiv1.Pod, nodes []*apiv1.Node,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	predicateChecker *simulatorinternal.PredicateChecker, oldHints map[string]string, newHints map[string]string,
	usageTracker *simulatorinternal.UsageTracker, timestamp time.Time) error {

	newNodeInfos := make(map[string]*schedulernodeinfo.NodeInfo)
	for k, v := range nodeInfos {
		newNodeInfos[k] = v
	}

	podKey := func(pod *apiv1.Pod) string {
		return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	}

	loggingQuota := glogx.PodsLoggingQuota()

	tryNodeForPod := func(nodename string, pod *apiv1.Pod, predicateMeta predicates.PredicateMetadata) bool {
		nodeInfo, found := newNodeInfos[nodename]
		if found {
			if nodeInfo.Node() == nil {
				// NodeInfo is generated based on pods. It is possible that node is removed from
				// an api server faster than the pod that were running on them. In such a case
				// we have to skip this nodeInfo. It should go away pretty soon.
				klog.Warningf("No node in nodeInfo %s -> %v", nodename, nodeInfo)
				return false
			}
			err := predicateChecker.CheckPredicates(pod, predicateMeta, nodeInfo)
			// nolint
			if err != nil {
				glogx.V(4).UpTo(loggingQuota).Infof("Evaluation %s for %s/%s -> %v",
					nodename, pod.Namespace, pod.Name, err.VerboseError())
			} else {
				// DOTO(mwielgus): Optimize it.
				klog.V(4).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, nodename)
				podsOnNode := nodeInfo.Pods()
				podsOnNode = append(podsOnNode, pod)
				newNodeInfo := schedulernodeinfo.NewNodeInfo(podsOnNode...)
				err := newNodeInfo.SetNode(nodeInfo.Node())
				if err != nil {
					klog.Warningf("SetNode falied. Error: %v", err)
				}
				newNodeInfos[nodename] = newNodeInfo
				newHints[podKey(pod)] = nodename
				return true
			}
		}
		return false
	}

	// DOTO: come up with a better semi-random semi-utilization sorted
	// layout.
	shuffledNodes := shuffleNodes(nodes)

	pods = tpu.ClearTPURequests(pods)
	for _, podptr := range pods {
		newpod := *podptr
		newpod.Spec.NodeName = ""
		pod := &newpod

		foundPlace := false
		targetNode := ""
		predicateMeta := predicateChecker.GetPredicateMetadata(pod, newNodeInfos)
		loggingQuota.Reset()

		klog.V(5).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)

		hintedNode, hasHint := oldHints[podKey(pod)]
		if hasHint {
			if hintedNode != removedNode && tryNodeForPod(hintedNode, pod, predicateMeta) {
				foundPlace = true
				targetNode = hintedNode
			}
		}
		if !foundPlace {
			for _, node := range shuffledNodes {
				if node.Name == removedNode {
					continue
				}
				if tryNodeForPod(node.Name, pod, predicateMeta) {
					foundPlace = true
					targetNode = node.Name
					break
				}
			}
			if !foundPlace {
				glogx.V(4).Over(loggingQuota).Infof("%v other nodes evaluated for %s/%s",
					-loggingQuota.Left(), pod.Namespace, pod.Name)
				return fmt.Errorf("failed to find place for %s", podKey(pod))
			}
		}

		usageTracker.RegisterUsage(removedNode, targetNode, timestamp)
	}
	return nil
}

func shuffleNodes(nodes []*apiv1.Node) []*apiv1.Node {
	result := make([]*apiv1.Node, len(nodes))
	copy(result, nodes)
	for i := range result {
		j := rand.Intn(len(result)) // nolint instead of crypto/rand
		result[i], result[j] = result[j], result[i]
	}
	return result
}
