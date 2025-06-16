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
	ctx "context"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	metricsinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/metrics"
)

const (
	// ScaleDownDisabledKey is the name of annotation marking node as not eligible for scale down.
	ScaleDownDisabledKey = "cluster-autoscaler.kubernetes.io/scale-down-disabled"
	// DelayDeletionAnnotationPrefix is the prefix of annotation marking node as it needs to wait
	// for other K8s components before deleting node.
	DelayDeletionAnnotationPrefix = "delay-deletion.cluster-autoscaler.kubernetes.io/"
	// NodeDeletionCost is the cost of node's deletion
	NodeDeletionCost = "io.tencent.bcs.dev/node-deletion-cost"
)

const (
	// MaxKubernetesEmptyNodeDeletionTime is the maximum time needed by Kubernetes to delete an empty node.
	MaxKubernetesEmptyNodeDeletionTime = 3 * time.Minute
	// MaxCloudProviderNodeDeletionTime is the maximum time needed by cloud provider to delete a node.
	MaxCloudProviderNodeDeletionTime = 5 * time.Minute
	// MaxPodEvictionTime is the maximum time CA tries to evict a pod before giving up.
	MaxPodEvictionTime = 2 * time.Minute
	// EvictionRetryTime is the time after CA retries failed pod eviction.
	EvictionRetryTime = 10 * time.Second
	// PodEvictionHeadroom is the extra time we wait to catch situations when the pod is ignoring SIGTERM and
	// is killed with SIGKILL after MaxGracefulTerminationTime
	PodEvictionHeadroom = 30 * time.Second
	// DaemonSetEvictionEmptyNodeTimeout is the time to evict all DaemonSet pods on empty node
	DaemonSetEvictionEmptyNodeTimeout = 10 * time.Second
	// DeamonSetTimeBetweenEvictionRetries is a time between retries to create eviction that uses
	// for DaemonSet eviction for empty nodes
	DeamonSetTimeBetweenEvictionRetries = 3 * time.Second
)

const (
	// NoReason - sanity check, this should never be set explicitly. If this is found in the wild, it means that it was
	// implicitly initialized and might indicate a bug.
	NoReason string = "NoReason"
	// ScaleDownDisabledAnnotation - node can't be removed because it has a "scale down disabled" annotation.
	ScaleDownDisabledAnnotation string = "ScaleDownDisabledAnnotation"
	// NotAutoscaled - node can't be removed because it doesn't belong to an autoscaled node group.
	NotAutoscaled string = "NotAutoscaled"
	// NotUnneededLongEnough - node can't be removed because it wasn't unneeded for long enough.
	NotUnneededLongEnough string = "NotUnneededLongEnough"
	// NotUnreadyLongEnough - node can't be removed because it wasn't unready for long enough.
	NotUnreadyLongEnough string = "NotUnreadyLongEnough"
	// NodeGroupMinSizeReached - node can't be removed because its node group is at its minimal size already.
	NodeGroupMinSizeReached string = "NodeGroupMinSizeReached"
	// MinimalResourceLimitExceeded - node can't be removed because it would violate cluster-wide minimal resource limits.
	MinimalResourceLimitExceeded string = "MinimalResourceLimitExceeded"
	// CurrentlyBeingDeleted - node can't be removed because it's already in the process of being deleted.
	CurrentlyBeingDeleted string = "CurrentlyBeingDeleted"
	// NotUnderutilized - node can't be removed because it's not underutilized.
	NotUnderutilized string = "NotUnderutilized"
	// NotUnneededOtherReason - node can't be removed because it's not marked as unneeded for other reasons
	// (e.g. it wasn't inspected at all in a given autoscaler loop).
	NotUnneededOtherReason string = "NotUnneededOtherReason"
	// RecentlyUnremovable - node can't be removed because it was recently found to be unremovable.
	RecentlyUnremovable string = "RecentlyUnremovable"
	// NoPlaceToMovePods - node can't be removed because there's no place to move its pods to.
	NoPlaceToMovePods string = "NoPlaceToMovePods"
	// BlockedByPod - node can't be removed because a pod running on it can't be moved. The reason why
	// should be in BlockingPod.
	BlockedByPod string = "BlockedByPod"
	// BufferNotEnough - node can't be removed because of buffer ratio
	BufferNotEnough string = "BufferNotEnough"
	// UnexpectedError - node can't be removed because of an unexpected error.
	UnexpectedError string = "UnexpectedError"

	// ControllerNotFound - pod is blocking scale down because its controller can't be found.
	ControllerNotFound string = "ControllerNotFound"
	// MinReplicasReached - pod is blocking scale down because its controller already has the minimum number of replicas.
	MinReplicasReached string = "MinReplicasReached"
	// NotReplicated - pod is blocking scale down because it's not replicated.
	NotReplicated string = "NotReplicated"
	// LocalStorageRequested - pod is blocking scale down because it requests local storage.
	LocalStorageRequested string = "LocalStorageRequested"
	// NotSafeToEvictAnnotation - pod is blocking scale down because it has a "not safe to evict" annotation.
	NotSafeToEvictAnnotation string = "NotSafeToEvictAnnotation"
	// UnmovableKubeSystemPod - pod is blocking scale down because it's a non-daemonset, non-mirrored,
	// non-pdb-assigned kube-system pod.
	UnmovableKubeSystemPod string = "UnmovableKubeSystemPod"
	// NotEnoughPdb - pod is blocking scale down because it doesn't have enough PDB left.
	NotEnoughPdb string = "NotEnoughPdb"

	// NodeDeleteErrorFailedToMarkToBeDeleted - node deletion failed because the node couldn't be marked to be deleted.
	NodeDeleteErrorFailedToMarkToBeDeleted = "FailedToMarkToBeDeleted"
	// NodeDeleteErrorFailedToEvictPods - node deletion failed because some of the pods couldn't be evicted from the node.
	NodeDeleteErrorFailedToEvictPods = "FailedToEvictPods"
	// NodeDeleteErrorFailedToDelete - failed to delete the node from the cloud provider.
	NodeDeleteErrorFailedToDelete = "FailedToDelete"
)

// NodeDeletionTracker keeps track of node deletions.
type NodeDeletionTracker struct {
	sync.Mutex
	nonEmptyNodeDeleteInProgress bool
	// A map of node delete results by node name. It's being constantly emptied into ScaleDownStatus
	// objects in order to notify the ScaleDownStatusProcessor that the node drain has ended or that
	// an error occurred during the deletion process.
	nodeDeleteResults map[string]status.NodeDeleteResult
	// A map which keeps track of deletions in progress for nodepools.
	// Key is a node group id and value is a number of node deletions in progress.
	deletionsInProgress map[string]int
}

// Get current time. Proxy for unit tests.
var now func() time.Time = time.Now // nolint

// NewNodeDeletionTracker creates new NodeDeletionTracker.
func NewNodeDeletionTracker() *NodeDeletionTracker {
	return &NodeDeletionTracker{
		nodeDeleteResults:   make(map[string]status.NodeDeleteResult),
		deletionsInProgress: make(map[string]int),
	}
}

// IsNonEmptyNodeDeleteInProgress returns true if a non empty node is being deleted.
func (n *NodeDeletionTracker) IsNonEmptyNodeDeleteInProgress() bool {
	n.Lock()
	defer n.Unlock()
	return n.nonEmptyNodeDeleteInProgress
}

// SetNonEmptyNodeDeleteInProgress sets non empty node deletion in progress status.
func (n *NodeDeletionTracker) SetNonEmptyNodeDeleteInProgress(status bool) {
	n.Lock()
	defer n.Unlock()
	n.nonEmptyNodeDeleteInProgress = status
}

// StartDeletion increments node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) StartDeletion(nodeGroupId string) {
	n.Lock()
	defer n.Unlock()
	n.deletionsInProgress[nodeGroupId]++
}

// EndDeletion decrements node deletion in progress counter for the given nodegroup.
func (n *NodeDeletionTracker) EndDeletion(nodeGroupId string) {
	n.Lock()
	defer n.Unlock()

	value, found := n.deletionsInProgress[nodeGroupId]
	if !found {
		klog.Errorf("This should never happen, counter for %s in DelayedNodeDeletionStatus wasn't found", nodeGroupId)
		return
	}
	if value <= 0 {
		klog.Errorf("This should never happen, counter for %s in DelayedNodeDeletionStatus isn't greater than 0,"+
			" counter value is %d", nodeGroupId, value)
	}
	n.deletionsInProgress[nodeGroupId]--
	if n.deletionsInProgress[nodeGroupId] <= 0 {
		delete(n.deletionsInProgress, nodeGroupId)
	}
}

// GetDeletionsInProgress returns the number of deletions in progress for the given node group.
func (n *NodeDeletionTracker) GetDeletionsInProgress(nodeGroupId string) int {
	n.Lock()
	defer n.Unlock()
	return n.deletionsInProgress[nodeGroupId]
}

// AddNodeDeleteResult adds a node delete result to the result map.
func (n *NodeDeletionTracker) AddNodeDeleteResult(nodeName string, result status.NodeDeleteResult) {
	n.Lock()
	defer n.Unlock()
	if result.ResultType != status.NodeDeleteOk {
		switch result.ResultType {
		case status.NodeDeleteErrorFailedToDelete:
			metrics.UpdateUnremovableNodes(nodeName, NodeDeleteErrorFailedToDelete, "", "")
			metricsinternal.RegisterFailedScaleDown(nodeName, NodeDeleteErrorFailedToDelete)
		case status.NodeDeleteErrorFailedToEvictPods:
			metrics.UpdateUnremovableNodes(nodeName, NodeDeleteErrorFailedToEvictPods, "", "")
			metricsinternal.RegisterFailedScaleDown(nodeName, NodeDeleteErrorFailedToEvictPods)
		case status.NodeDeleteErrorFailedToMarkToBeDeleted:
			metrics.UpdateUnremovableNodes(nodeName, NodeDeleteErrorFailedToMarkToBeDeleted, "", "")
			metricsinternal.RegisterFailedScaleDown(nodeName, NodeDeleteErrorFailedToMarkToBeDeleted)
		}
	}
	n.nodeDeleteResults[nodeName] = result
}

// GetAndClearNodeDeleteResults returns the whole result map and replaces it with a new empty one.
func (n *NodeDeletionTracker) GetAndClearNodeDeleteResults() map[string]status.NodeDeleteResult {
	n.Lock()
	defer n.Unlock()
	results := n.nodeDeleteResults
	n.nodeDeleteResults = make(map[string]status.NodeDeleteResult)
	return results
}

type scaleDownResourcesLimits map[string]int64
type scaleDownResourcesDelta map[string]int64

// used as a value in scaleDownResourcesLimits if actual limit could not be obtained
// due to errors talking to cloud provider
const scaleDownLimitUnknown = math.MinInt64

func (sd *ScaleDown) computeScaleDownResourcesLeftLimits(nodes []*apiv1.Node,
	resourceLimiter *cloudprovider.ResourceLimiter,
	cp cloudprovider.CloudProvider, timestamp time.Time) scaleDownResourcesLimits {
	totalCores, totalMem := calculateScaleDownCoresMemoryTotal(nodes, timestamp)

	var totalResources map[string]int64
	var totalResourcesErr error
	if cloudprovider.ContainsCustomResources(resourceLimiter.GetResources()) {
		totalResources, totalResourcesErr = sd.calculateScaleDownCustomResourcesTotal(nodes, cp, timestamp)
	}

	resultScaleDownLimits := make(scaleDownResourcesLimits)
	for _, resource := range resourceLimiter.GetResources() {
		min := resourceLimiter.GetMin(resource)

		// we put only actual limits into final map. No entry means no limit.
		if min > 0 {
			switch {
			case resource == cloudprovider.ResourceNameCores:
				resultScaleDownLimits[resource] = computeAboveMin(totalCores, min)
			case resource == cloudprovider.ResourceNameMemory:
				resultScaleDownLimits[resource] = computeAboveMin(totalMem, min)
			case cloudprovider.IsCustomResource(resource):
				if totalResourcesErr != nil {
					resultScaleDownLimits[resource] = scaleDownLimitUnknown
				} else {
					resultScaleDownLimits[resource] = computeAboveMin(totalResources[resource], min)
				}
			default:
				klog.Errorf("Scale down limits defined for unsupported resource '%s'", resource)
			}
		}
	}
	return resultScaleDownLimits
}

func computeAboveMin(total int64, min int64) int64 {
	if total > min {
		return total - min
	}
	return 0

}

func calculateScaleDownCoresMemoryTotal(nodes []*apiv1.Node, timestamp time.Time) (int64, int64) {
	var coresTotal, memoryTotal int64
	for _, node := range nodes {
		if isNodeBeingDeleted(node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		cores, memory := core_utils.GetNodeCoresAndMemory(node)

		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal
}

// NOCC:tosa/fn_length(设计如此)
func (sd *ScaleDown) calculateScaleDownCustomResourcesTotal(nodes []*apiv1.Node, cp cloudprovider.CloudProvider,
	timestamp time.Time) (map[string]int64, error) {
	result := make(map[string]int64)
	ngCache := make(map[string][]customresources.CustomResourceTarget)
	for _, node := range nodes {
		if isNodeBeingDeleted(node, timestamp) {
			// Nodes being deleted do not count towards total cluster resources
			continue
		}
		nodeGroup, err := cp.NodeGroupForNode(node)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
				"can not get node group for node %v when calculating cluster gpu usage", node.Name)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			// We do not trust cloud providers to return properly constructed nil for interface type
			// - hence the reflection check.
			// See https://golang.org/doc/faq#nil_error
			// DOTO[lukaszos] consider creating cloud_provider sanitizer which will wrap cloud provider
			// and ensure sane behavior.
			nodeGroup = nil
		}

		var resourceTargets []customresources.CustomResourceTarget
		var cacheHit bool

		if nodeGroup != nil {
			resourceTargets, cacheHit = ngCache[nodeGroup.Id()]
		}
		if !cacheHit {
			resourceTargets, err = sd.processors.CustomResourcesProcessor.GetNodeResourceTargets(sd.context, node, nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).
					AddPrefix("can not get gpu count for node %v when calculating cluster gpu usage")
			}
			if nodeGroup != nil {
				ngCache[nodeGroup.Id()] = resourceTargets
			}
		}

		for _, resourceTarget := range resourceTargets {
			if resourceTarget.ResourceType == "" || resourceTarget.ResourceCount == 0 {
				continue
			}
			result[resourceTarget.ResourceType] += resourceTarget.ResourceCount
		}
	}

	return result, nil
}

func isNodeBeingDeleted(node *apiv1.Node, timestamp time.Time) bool {
	deleteTime, _ := deletetaint.GetToBeDeletedTime(node)
	return deleteTime != nil && (timestamp.Sub(*deleteTime) < MaxCloudProviderNodeDeletionTime ||
		timestamp.Sub(*deleteTime) < MaxKubernetesEmptyNodeDeletionTime)
}

func noScaleDownLimitsOnResources() scaleDownResourcesLimits {
	return nil
}

func copyScaleDownResourcesLimits(source scaleDownResourcesLimits) scaleDownResourcesLimits {
	copy := scaleDownResourcesLimits{}
	for k, v := range source {
		copy[k] = v
	}
	return copy
}

// nolint
func (sd *ScaleDown) computeScaleDownResourcesDelta(cp cloudprovider.CloudProvider,
	node *apiv1.Node,
	nodeGroup cloudprovider.NodeGroup,
	resourcesWithLimits []string) (scaleDownResourcesDelta, errors.AutoscalerError) {
	resultScaleDownDelta := make(scaleDownResourcesDelta)

	nodeCPU, nodeMemory := core_utils.GetNodeCoresAndMemory(node)
	resultScaleDownDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleDownDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	if cloudprovider.ContainsCustomResources(resourcesWithLimits) {
		resourceTargets, err := sd.processors.CustomResourcesProcessor.GetNodeResourceTargets(sd.context, node, nodeGroup)
		if err != nil {
			return scaleDownResourcesDelta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).
				AddPrefix("Failed to get node %v custom resources: %v", node.Name)
		}
		for _, resourceTarget := range resourceTargets {
			resultScaleDownDelta[resourceTarget.ResourceType] = resourceTarget.ResourceCount
		}
	}
	return resultScaleDownDelta, nil
}

type scaleDownLimitsCheckResult struct {
	exceeded          bool
	exceededResources []string
}

func scaleDownLimitsNotExceeded() scaleDownLimitsCheckResult {
	return scaleDownLimitsCheckResult{false, []string{}}
}

func (limits *scaleDownResourcesLimits) checkScaleDownDeltaWithinLimits(
	delta scaleDownResourcesDelta) scaleDownLimitsCheckResult {
	exceededResources := sets.NewString()
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*limits)[resource]
		if found {
			if (resourceDelta > 0) && (resourceLeft == scaleDownLimitUnknown || resourceDelta > resourceLeft) {
				exceededResources.Insert(resource)
			}
		}
	}
	if len(exceededResources) > 0 {
		return scaleDownLimitsCheckResult{true, exceededResources.List()}
	}

	return scaleDownLimitsNotExceeded()
}

func (limits *scaleDownResourcesLimits) tryDecrementLimitsByDelta(
	delta scaleDownResourcesDelta) scaleDownLimitsCheckResult {
	result := limits.checkScaleDownDeltaWithinLimits(delta)
	if result.exceeded {
		return result
	}
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*limits)[resource]
		if found {
			(*limits)[resource] = resourceLeft - resourceDelta
		}
	}
	return scaleDownLimitsNotExceeded()
}

// ScaleDown is responsible for maintaining the state needed to perform unneeded node removals.
type ScaleDown struct {
	context                      *context.AutoscalingContext
	processors                   *processors.AutoscalingProcessors
	clusterStateRegistry         *clusterstate.ClusterStateRegistry
	unneededNodes                map[string]time.Time
	unneededNodesList            []*apiv1.Node
	unremovableNodes             map[string]time.Time
	podLocationHints             map[string]string
	nodeUtilizationMap           map[string]simulator.UtilizationInfo
	usageTracker                 *simulator.UsageTracker
	nodeDeletionTracker          *NodeDeletionTracker
	unremovableNodeReasons       map[string]*simulator.UnremovableNode
	expendablePodsPriorityCutoff int
	cpuRatio                     float64
	memRatio                     float64
	ratio                        float64
	evictLatest                  bool
}

// NewScaleDown builds new ScaleDown object.
func NewScaleDown(context *context.AutoscalingContext,
	processors *processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	expendablePodsPriorityCutoff int, cpuRatio, memRatio, ratio float64,
	evictLatest bool) *ScaleDown {
	return &ScaleDown{
		context:                      context,
		processors:                   processors,
		clusterStateRegistry:         clusterStateRegistry,
		unneededNodes:                make(map[string]time.Time),
		unremovableNodes:             make(map[string]time.Time),
		podLocationHints:             make(map[string]string),
		nodeUtilizationMap:           make(map[string]simulator.UtilizationInfo),
		usageTracker:                 simulator.NewUsageTracker(),
		unneededNodesList:            make([]*apiv1.Node, 0),
		nodeDeletionTracker:          NewNodeDeletionTracker(),
		unremovableNodeReasons:       make(map[string]*simulator.UnremovableNode),
		expendablePodsPriorityCutoff: expendablePodsPriorityCutoff,
		cpuRatio:                     cpuRatio,
		memRatio:                     memRatio,
		ratio:                        ratio,
		evictLatest:                  evictLatest,
	}
}

// CleanUp cleans up the internal ScaleDown state.
func (sd *ScaleDown) CleanUp(timestamp time.Time) {
	// Use default ScaleDownUnneededTime as in this context the value
	// doesn't apply to any specific NodeGroup.
	sd.usageTracker.CleanUp(timestamp.Add(-sd.context.NodeGroupDefaults.ScaleDownUnneededTime))
	sd.clearUnremovableNodeReasons()
}

// CleanUpUnneededNodes clears the list of unneeded nodes.
func (sd *ScaleDown) CleanUpUnneededNodes() {
	sd.unneededNodesList = make([]*apiv1.Node, 0)
	sd.unneededNodes = make(map[string]time.Time)
}

func (sd *ScaleDown) checkNodeUtilization(timestamp time.Time, node *apiv1.Node,
	nodeInfo *schedulerframework.NodeInfo) (simulator.UnremovableReason, *simulator.UtilizationInfo) {
	// Skip nodes that were recently checked.
	if _, found := sd.unremovableNodes[node.Name]; found {
		return simulator.RecentlyUnremovable, nil
	}

	// Skip nodes marked to be deleted, if they were marked recently.
	// Old-time marked nodes are again eligible for deletion - something went wrong with them
	// and they have not been deleted.
	if isNodeBeingDeleted(node, timestamp) {
		klog.V(1).Infof("Skipping %s from delete consideration - the node is currently being deleted",
			node.Name)
		metrics.UpdateUnremovableNodes(node.Name, CurrentlyBeingDeleted, "", "")
		return simulator.CurrentlyBeingDeleted, nil
	}

	// Skip nodes marked with no scale down annotation
	if hasNoScaleDownAnnotation(node) {
		klog.V(1).Infof("Skipping %s from delete consideration - the node is marked as no scale down",
			node.Name)
		metrics.UpdateUnremovableNodes(node.Name, ScaleDownDisabledAnnotation, "", "")
		return simulator.ScaleDownDisabledAnnotation, nil
	}

	utilInfo, err := simulator.CalculateUtilization(node, nodeInfo, sd.context.IgnoreDaemonSetsUtilization,
		sd.context.IgnoreMirrorPodsUtilization, sd.context.CloudProvider.GPULabel(), timestamp)
	if err != nil {
		klog.Warningf("Failed to calculate utilization for %s: %v", node.Name, err)
	}

	nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		metrics.UpdateUnremovableNodes(node.Name, UnexpectedError, "", "")
		return simulator.UnexpectedError, nil
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		// We should never get here as non-autoscaled nodes should not be included in scaleDownCandidates list
		// (and the default PreFilteringScaleDownNodeProcessor would indeed filter them out).
		klog.Warningf("Skipped %s from delete consideration - the node is not autoscaled", node.Name)
		metrics.UpdateUnremovableNodes(node.Name, NotAutoscaled, "", "")
		return simulator.NotAutoscaled, nil
	}

	underutilized, err := sd.isNodeBelowUtilizationThreshold(node, nodeGroup, utilInfo)
	if err != nil {
		klog.Warningf("Failed to check utilization thresholds for %s: %v", node.Name, err)
		metrics.UpdateUnremovableNodes(node.Name, UnexpectedError, "", "")
		return simulator.UnexpectedError, nil
	}
	if !underutilized {
		klog.V(4).Infof("Node %s is not suitable for removal - %s utilization too big (%f)",
			node.Name, utilInfo.ResourceName, utilInfo.Utilization)
		metrics.UpdateUnremovableNodes(node.Name, NotUnderutilized, "", "")
		return simulator.NotUnderutilized, &utilInfo
	}

	klog.V(4).Infof("Node %s - %s utilization %f", node.Name, utilInfo.ResourceName, utilInfo.Utilization)

	return simulator.NoReason, &utilInfo
}

// UpdateUnneededNodes calculates which nodes are not needed, i.e. all pods can be scheduled somewhere else,
// and updates unneededNodes map accordingly. It also computes information where pods can be rescheduled and
// node utilization level. The computations are made only for the nodes managed by CA.
// * destinationNodes are the nodes that can potentially take in any pods that are evicted because of a scale down.
// * scaleDownCandidates are the nodes that are being considered for scale down.
// * timestamp is the current timestamp.
// * pdbs is a list of pod disruption budgets.
// NOCC:golint/fnsize(设计如此)
// nolint funlen
func (sd *ScaleDown) UpdateUnneededNodes(
	destinationNodes []*apiv1.Node,
	scaleDownCandidates []*apiv1.Node,
	timestamp time.Time,
	pdbs []*policyv1.PodDisruptionBudget,
) errors.AutoscalerError {

	// Only scheduled non expendable pods and pods waiting for lower priority pods preemption can prevent node delete.
	// Extract cluster state from snapshot for initial analysis
	allNodeInfos, err := sd.context.ClusterSnapshot.NodeInfos().List()
	if err != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return errors.ToAutoscalerError(errors.InternalError, err)
	}

	sd.updateUnremovableNodes(timestamp)

	skipped, utilizationMap, currentlyUnneededNodeNames := sd.checkScaleDownCandidates(scaleDownCandidates, timestamp)

	if skipped > 0 {
		klog.V(1).Infof("Scale-down calculation: ignoring %v nodes unremovable in the last %v",
			skipped, sd.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
	}

	emptyNodesList := sd.getEmptyNodesNoResourceLimits(currentlyUnneededNodeNames,
		len(currentlyUnneededNodeNames), timestamp)

	emptyNodes := make(map[string]bool)
	for _, node := range emptyNodesList {
		emptyNodes[node.Name] = true
	}

	currentlyUnneededNonEmptyNodes := make([]string, 0, len(currentlyUnneededNodeNames))
	for _, node := range currentlyUnneededNodeNames {
		if !emptyNodes[node] {
			currentlyUnneededNonEmptyNodes = append(currentlyUnneededNonEmptyNodes, node)
		}
	}

	// Phase2 - check which nodes can be probably removed using fast drain.
	currentCandidates, currentNonCandidates := sd.chooseCandidates(currentlyUnneededNonEmptyNodes)

	destinations := make([]string, 0, len(destinationNodes))
	for _, destinationNode := range destinationNodes {
		destinations = append(destinations, destinationNode.Name)
	}

	// Look for nodes to remove in the current candidates
	nodesToRemove, unremovable, newHints, simulatorErr := simulator.FindNodesToRemove(
		currentCandidates,
		destinations,
		nil,
		sd.context.ClusterSnapshot,
		sd.context.PredicateChecker,
		len(currentCandidates),
		true,
		sd.podLocationHints,
		sd.usageTracker,
		timestamp,
		pdbs)
	if simulatorErr != nil {
		return sd.markSimulationError(simulatorErr, timestamp)
	}

	nodesToRemove, unremovable, newHints, simulatorErr = sd.findAdditionalCandidates(allNodeInfos,
		currentCandidates, currentNonCandidates, destinations, timestamp, pdbs, nodesToRemove,
		unremovable, newHints)
	if simulatorErr != nil {
		return simulatorErr
	}

	for _, node := range emptyNodesList {
		nodesToRemove = append(nodesToRemove, simulator.NodeToBeRemoved{Node: node,
			PodsToReschedule: []*apiv1.Pod{}})
	}

	haveEmpty := len(emptyNodesList) > 0

	result, unneededNodesList := sd.updateUnneedTime(nodesToRemove, haveEmpty, timestamp)

	// Add nodes to unremovable map
	if len(unremovable) > 0 {
		unremovableTimeout := timestamp.Add(sd.context.AutoscalingOptions.UnremovableNodeRecheckTimeout)
		for _, unremovableNode := range unremovable {
			sd.unremovableNodes[unremovableNode.Node.Name] = unremovableTimeout
			sd.addUnremovableNode(unremovableNode)
			nodeReason := convertUnRemovableNodeReason(unremovableNode.Reason)
			if unremovableNode.BlockingPod != nil {
				podReason := convertBlockingPodReason(unremovableNode.BlockingPod.Reason)
				metrics.UpdateUnremovableNodes(unremovableNode.Node.Name, nodeReason,
					fmt.Sprintf("%s/%s", unremovableNode.BlockingPod.Pod.Namespace,
						unremovableNode.BlockingPod.Pod.Name), podReason)
			} else {
				metrics.UpdateUnremovableNodes(unremovableNode.Node.Name, nodeReason, "", "")
			}
		}
		klog.V(1).Infof("%v nodes found to be unremovable in simulation, will re-check them at %v",
			len(unremovable), unremovableTimeout)
	}

	// This method won't always check all nodes, so let's give a generic reason for all nodes that weren't checked.
	for _, node := range scaleDownCandidates {
		_, unremovableReasonProvided := sd.unremovableNodeReasons[node.Name]
		_, unneeded := result[node.Name]
		if !unneeded && !unremovableReasonProvided {
			sd.addUnremovableNodeReason(node, simulator.NotUnneededOtherReason)
			metrics.UpdateUnremovableNodes(node.Name, NotUnneededOtherReason, "", "")
		}
	}

	// Update state and metrics
	sd.unneededNodesList = unneededNodesList
	sd.unneededNodes = result
	sd.podLocationHints = newHints
	sd.nodeUtilizationMap = utilizationMap
	sd.clusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	metrics.UpdateUnneededNodesCount(len(sd.unneededNodesList))
	if klog.V(4).Enabled() {
		for key, val := range sd.unneededNodes {
			klog.Infof("%s is unneeded since %s duration %s",
				key, val.String(), timestamp.Sub(val).String())
		}
	}
	return nil
}

func (sd *ScaleDown) updateUnneedTime(nodesToRemove []simulator.NodeToBeRemoved, haveEmpty bool,
	timestamp time.Time) (map[string]time.Time, []*apiv1.Node) {
	// Update the timestamp map.
	result := make(map[string]time.Time)
	unneededNodesList := make([]*apiv1.Node, 0)
	for _, node := range nodesToRemove {
		name := node.Node.Name
		unneededNodesList = append(unneededNodesList, node.Node)
		if len(node.PodsToReschedule) > 0 && haveEmpty {
			// fix(bcs): 有空节点情况下，重置非空节点时间，保证优先缩容空节点
			result[name] = timestamp
		} else if val, found := sd.unneededNodes[name]; !found {
			result[name] = timestamp
		} else {
			result[name] = val
		}
	}
	return result, unneededNodesList
}

func (sd *ScaleDown) checkScaleDownCandidates(scaleDownCandidates []*apiv1.Node,
	timestamp time.Time) (int, map[string]simulator.UtilizationInfo, []string) {
	skipped := 0
	utilizationMap := make(map[string]simulator.UtilizationInfo)
	currentlyUnneededNodeNames := make([]string, 0, len(scaleDownCandidates))

	for _, node := range scaleDownCandidates {
		nodeInfo, err := sd.context.ClusterSnapshot.NodeInfos().Get(node.Name)
		if err != nil {
			klog.Errorf("Can't retrieve scale-down candidate %s from snapshot, err: %v", node.Name, err)
			sd.addUnremovableNodeReason(node, simulator.UnexpectedError)
			metrics.UpdateUnremovableNodes(node.Name, UnexpectedError, "", "")
			continue
		}

		reason, utilInfo := sd.checkNodeUtilization(timestamp, node, nodeInfo)
		if utilInfo != nil {
			utilizationMap[node.Name] = *utilInfo
		}
		if reason != simulator.NoReason {
			// For logging purposes.
			if reason == simulator.RecentlyUnremovable {
				skipped++
			}

			sd.addUnremovableNodeReason(node, reason)
			continue
		}

		currentlyUnneededNodeNames = append(currentlyUnneededNodeNames, node.Name)
	}

	return skipped, utilizationMap, currentlyUnneededNodeNames
}

// nolint
func (sd *ScaleDown) findAdditionalCandidates(allNodeInfos []*schedulerframework.NodeInfo,
	currentCandidates, currentNonCandidates []string, destinations []string,
	timestamp time.Time, pdbs []*policyv1.PodDisruptionBudget,
	nodesToRemove []simulator.NodeToBeRemoved, unremovable []*simulator.UnremovableNode,
	newHints map[string]string) ([]simulator.NodeToBeRemoved, []*simulator.UnremovableNode,
	map[string]string, errors.AutoscalerError) {
	additionalCandidatesCount := sd.context.ScaleDownNonEmptyCandidatesCount - len(nodesToRemove)
	if additionalCandidatesCount > len(currentNonCandidates) {
		additionalCandidatesCount = len(currentNonCandidates)
	}
	// Limit the additional candidates pool size for better performance.
	additionalCandidatesPoolSize := int(math.Ceil(float64(len(allNodeInfos)) * sd.context.ScaleDownCandidatesPoolRatio))
	if additionalCandidatesPoolSize < sd.context.ScaleDownCandidatesPoolMinCount {
		additionalCandidatesPoolSize = sd.context.ScaleDownCandidatesPoolMinCount
	}
	if additionalCandidatesPoolSize > len(currentNonCandidates) {
		additionalCandidatesPoolSize = len(currentNonCandidates)
	}
	if additionalCandidatesCount > 0 {
		// Look for additional nodes to remove among the rest of nodes.
		klog.V(3).Infof("Finding additional %v candidates for scale down.", additionalCandidatesCount)
		additionalNodesToRemove, additionalUnremovable, additionalNewHints, simulatorErr :=
			simulator.FindNodesToRemove(
				currentNonCandidates[:additionalCandidatesPoolSize],
				destinations,
				nil,
				sd.context.ClusterSnapshot,
				sd.context.PredicateChecker,
				additionalCandidatesCount,
				true,
				sd.podLocationHints,
				sd.usageTracker,
				timestamp,
				pdbs)
		if simulatorErr != nil {
			return nodesToRemove, unremovable, newHints, sd.markSimulationError(simulatorErr, timestamp)
		}
		if len(additionalNodesToRemove) > additionalCandidatesCount {
			additionalNodesToRemove = additionalNodesToRemove[:additionalCandidatesCount]
		}
		nodesToRemove = append(nodesToRemove, additionalNodesToRemove...)
		unremovable = append(unremovable, additionalUnremovable...)
		for key, value := range additionalNewHints {
			newHints[key] = value
		}
	}
	return nodesToRemove, unremovable, newHints, nil
}

// isNodeBelowUtilizationThreshold determines if a given node utilization is below threshold.
func (sd *ScaleDown) isNodeBelowUtilizationThreshold(node *apiv1.Node, nodeGroup cloudprovider.NodeGroup,
	utilInfo simulator.UtilizationInfo) (bool, error) {
	// 如果使用率为0，直接返回 true，避免 threshold 为 0 时返回 false
	if math.Abs(utilInfo.Utilization) < 1e-6 {
		return true, nil
	}
	var threshold float64
	var err error
	if gpu.NodeHasGpu(sd.context.CloudProvider.GPULabel(), node) {
		threshold, err = sd.processors.NodeGroupConfigProcessor.GetScaleDownGpuUtilizationThreshold(
			sd.context, nodeGroup)
		if err != nil {
			return false, err
		}
	} else {
		threshold, err = sd.processors.NodeGroupConfigProcessor.GetScaleDownUtilizationThreshold(
			sd.context, nodeGroup)
		if err != nil {
			return false, err
		}
	}
	if utilInfo.Utilization >= threshold {
		return false, nil
	}
	return true, nil
}

// updateUnremovableNodes updates unremovableNodes map according to current
// state of the cluster. Removes from the map nodes that are no longer in the
// nodes list.
func (sd *ScaleDown) updateUnremovableNodes(timestamp time.Time) {
	if len(sd.unremovableNodes) == 0 {
		return
	}
	newUnremovableNodes := make(map[string]time.Time, len(sd.unremovableNodes))
	for oldUnremovable, ttl := range sd.unremovableNodes {
		if _, err := sd.context.ClusterSnapshot.NodeInfos().Get(oldUnremovable); err != nil {
			// Not logging on error level as most likely cause is that node is no longer in the cluster.
			klog.Infof("Can't retrieve node %s from snapshot, removing from unremovable map, err: %v",
				oldUnremovable, err)
			continue
		}
		if ttl.After(timestamp) {
			// Keep nodes that are still in the cluster and haven't expired yet.
			newUnremovableNodes[oldUnremovable] = ttl
		}
	}
	sd.unremovableNodes = newUnremovableNodes
}

func (sd *ScaleDown) clearUnremovableNodeReasons() {
	sd.unremovableNodeReasons = make(map[string]*simulator.UnremovableNode)
}

func (sd *ScaleDown) addUnremovableNodeReason(node *apiv1.Node, reason simulator.UnremovableReason) {
	sd.unremovableNodeReasons[node.Name] = &simulator.UnremovableNode{Node: node, Reason: reason, BlockingPod: nil}
}

func (sd *ScaleDown) addUnremovableNode(unremovableNode *simulator.UnremovableNode) {
	sd.unremovableNodeReasons[unremovableNode.Node.Name] = unremovableNode
}

func (sd *ScaleDown) getUnremovableNodesCount() map[simulator.UnremovableReason]int {
	reasons := make(map[simulator.UnremovableReason]int)

	for _, node := range sd.unremovableNodeReasons {
		reasons[node.Reason]++
	}

	return reasons
}

// markSimulationError indicates a simulation error by clearing  relevant scale
// down state and returning an appropriate error.
func (sd *ScaleDown) markSimulationError(simulatorErr errors.AutoscalerError,
	timestamp time.Time) errors.AutoscalerError {
	klog.Errorf("Error while simulating node drains: %v", simulatorErr)
	sd.unneededNodesList = make([]*apiv1.Node, 0)
	sd.unneededNodes = make(map[string]time.Time)
	sd.nodeUtilizationMap = make(map[string]simulator.UtilizationInfo)
	sd.clusterStateRegistry.UpdateScaleDownCandidates(sd.unneededNodesList, timestamp)
	return simulatorErr.AddPrefix("error while simulating node drains: ")
}

// chooseCandidates splits nodes into current candidates for scale-down and the
// rest. Current candidates are unneeded nodes from the previous run that are
// still in the nodes list.
func (sd *ScaleDown) chooseCandidates(nodes []string) (candidates []string, nonCandidates []string) {
	// Number of candidates should not be capped. We will look for nodes to remove
	// from the whole set of nodes.
	if sd.context.ScaleDownNonEmptyCandidatesCount <= 0 {
		return nodes, nil
	}
	for _, node := range nodes {
		if _, found := sd.unneededNodes[node]; found {
			candidates = append(candidates, node)
		} else {
			nonCandidates = append(nonCandidates, node)
		}
	}
	return candidates, nonCandidates
}

func (sd *ScaleDown) mapNodesToStatusScaleDownNodes(nodes []*apiv1.Node,
	nodeGroups map[string]cloudprovider.NodeGroup,
	evictedPodLists map[string][]*apiv1.Pod) []*status.ScaleDownNode {
	var result []*status.ScaleDownNode
	for _, node := range nodes {
		result = append(result, &status.ScaleDownNode{
			Node:        node,
			NodeGroup:   nodeGroups[node.Name],
			UtilInfo:    sd.nodeUtilizationMap[node.Name],
			EvictedPods: evictedPodLists[node.Name],
		})
	}
	return result
}

// SoftTaintUnneededNodes manage soft taints of unneeded nodes.
func (sd *ScaleDown) SoftTaintUnneededNodes(allNodes []*apiv1.Node) (errors []error) {
	defer metrics.UpdateDurationFromStart(metrics.ScaleDownSoftTaintUnneeded, time.Now())
	apiCallBudget := sd.context.AutoscalingOptions.MaxBulkSoftTaintCount
	timeBudget := sd.context.AutoscalingOptions.MaxBulkSoftTaintTime
	skippedNodes := 0
	startTime := now()
	for _, node := range allNodes {
		if deletetaint.HasToBeDeletedTaint(node) {
			// Do not consider nodes that are scheduled to be deleted
			continue
		}
		alreadyTainted := deletetaint.HasDeletionCandidateTaint(node)
		_, unneeded := sd.unneededNodes[node.Name]

		// Check if expected taints match existing taints
		if unneeded != alreadyTainted {
			if apiCallBudget <= 0 || now().Sub(startTime) >= timeBudget {
				skippedNodes++
				continue
			}
			apiCallBudget--
			if unneeded && !alreadyTainted {
				err := deletetaint.MarkDeletionCandidate(node, sd.context.ClientSet)
				if err != nil {
					errors = append(errors, err)
					klog.Warningf("Soft taint on %s adding error %v", node.Name, err)
				}
			}
			if !unneeded && alreadyTainted {
				_, err := deletetaint.CleanDeletionCandidate(node, sd.context.ClientSet)
				if err != nil {
					errors = append(errors, err)
					klog.Warningf("Soft taint on %s removal error %v", node.Name, err)
				}
			}
		}
	}
	if skippedNodes > 0 {
		klog.V(4).Infof("Skipped adding/removing soft taints on %v nodes - API call limit exceeded", skippedNodes)
	}
	return errors
}

// TryToScaleDown tries to scale down the cluster. It returns a result inside a ScaleDownStatus indicating
// if any node was removed and error if such occurred.
// NOCC:golint/fnsize(设计如此)
// nolint funlen
func (sd *ScaleDown) TryToScaleDown(
	currentTime time.Time,
	pdbs []*policyv1.PodDisruptionBudget,
) (*status.ScaleDownStatus, errors.AutoscalerError) {

	scaleDownStatus := &status.ScaleDownStatus{NodeDeleteResults: sd.nodeDeletionTracker.GetAndClearNodeDeleteResults()}
	nodeDeletionDuration := time.Duration(0)
	findNodesToRemoveDuration := time.Duration(0)
	defer updateScaleDownMetrics(time.Now(), &findNodesToRemoveDuration, &nodeDeletionDuration)

	allNodeInfos, errSnapshot := sd.context.ClusterSnapshot.NodeInfos().List()
	if errSnapshot != nil {
		// This should never happen, List() returns err only because scheduler interface requires it.
		return scaleDownStatus, errors.ToAutoscalerError(errors.InternalError, errSnapshot)
	}

	nodesWithoutMaster := filterOutMasters(allNodeInfos)
	nodesWithoutMasterNames := make([]string, 0, len(nodesWithoutMaster))
	for _, node := range nodesWithoutMaster {
		nodesWithoutMasterNames = append(nodesWithoutMasterNames, node.Name)
	}

	nodesInfoWithoutMaster := filterOutMastersNodeInfo(allNodeInfos)

	candidateNames := make([]string, 0)
	readinessMap := make(map[string]bool)
	candidateNodeGroups := make(map[string]cloudprovider.NodeGroup)
	gpuLabel := sd.context.CloudProvider.GPULabel()
	availableGPUTypes := sd.context.CloudProvider.GetAvailableGPUTypes()

	resourceLimiter, errCP := sd.context.CloudProvider.GetResourceLimiter()
	if errCP != nil {
		scaleDownStatus.Result = status.ScaleDownError
		return scaleDownStatus, errors.ToAutoscalerError(errors.CloudProviderError, errCP)
	}

	scaleDownResourcesLeft := sd.computeScaleDownResourcesLeftLimits(nodesWithoutMaster,
		resourceLimiter, sd.context.CloudProvider, currentTime)

	nodeGroupSize := utils.GetNodeGroupSizeMap(sd.context.CloudProvider)
	resourcesWithLimits := resourceLimiter.GetResources()
	for nodeName, unneededSince := range sd.unneededNodes {
		klog.V(2).Infof("%s was unneeded for %s", nodeName, currentTime.Sub(unneededSince).String())

		var checkErr error
		var nodeGroup cloudprovider.NodeGroup
		readinessMap, nodeGroup, checkErr = sd.checkNodeRemovable(nodeName, unneededSince, currentTime,
			nodeGroupSize, resourceLimiter, resourcesWithLimits, scaleDownResourcesLeft, readinessMap)
		if checkErr != nil {
			klog.Warningf(checkErr.Error())
			continue
		}

		candidateNames = append(candidateNames, nodeName)
		candidateNodeGroups[nodeName] = nodeGroup
	}

	if len(candidateNames) == 0 {
		klog.V(1).Infof("No candidates for scale down")
		scaleDownStatus.Result = status.ScaleDownNoUnneeded
		return scaleDownStatus, nil
	}
	klog.Infof("%d candidates: %v", len(candidateNames), candidateNames)

	// Trying to delete empty nodes in bulk. If there are no empty nodes then CA will
	// try to delete not-so-empty nodes, possibly killing some pods and allowing them
	// to recreate on other nodes.
	emptyNodes := sd.getEmptyNodes(candidateNames, sd.context.MaxEmptyBulkDelete,
		scaleDownResourcesLeft, currentTime)
	if len(emptyNodes) > 0 {
		klog.Infof("%d empty node", len(emptyNodes))
		// filter empty nodes with ratio
		emptyNodesAfterFilter := sd.filterNode(emptyNodes, nodesInfoWithoutMaster)
		klog.Infof("%d empty node after filter", len(emptyNodesAfterFilter))
		emptyNodes = emptyNodesAfterFilter
		if len(emptyNodes) == 0 {
			klog.V(1).Info("No node can be deleted after filter")
			scaleDownStatus.Result = status.ScaleDownNoNodeDeleted
			return scaleDownStatus, nil
		}

		nodeDeletionStart := time.Now()
		deletedNodes, err := sd.scheduleDeleteEmptyNodes(emptyNodes,
			sd.context.ClientSet, sd.context.Recorder, readinessMap, candidateNodeGroups)
		nodeDeletionDuration = time.Since(nodeDeletionStart)

		// DOTO: Give the processor some information about the nodes that failed to be deleted.
		scaleDownStatus.ScaledDownNodes = sd.mapNodesToStatusScaleDownNodes(deletedNodes,
			candidateNodeGroups, make(map[string][]*apiv1.Pod))
		if len(deletedNodes) > 0 {
			scaleDownStatus.Result = status.ScaleDownNodeDeleteStarted
		} else {
			scaleDownStatus.Result = status.ScaleDownError
		}
		if err != nil {
			return scaleDownStatus, err.AddPrefix("failed to delete at least one empty node: ")
		}
		return scaleDownStatus, nil
	}

	findNodesToRemoveStart := time.Now()
	// 根据 NodeDeletionCost 排序候选缩容节点
	candidateNames = sortNodesByDeletionCost(candidateNames, sd.context.ClusterSnapshot)

	// We look for only 1 node so new hints may be incomplete.
	nodesToRemove, unremovable, _, err := simulator.FindNodesToRemove(
		candidateNames,
		nodesWithoutMasterNames,
		sd.context.ListerRegistry,
		sd.context.ClusterSnapshot,
		sd.context.PredicateChecker,
		1,
		false,
		sd.podLocationHints,
		sd.usageTracker,
		time.Now(),
		pdbs)
	findNodesToRemoveDuration = time.Since(findNodesToRemoveStart)

	for _, unremovableNode := range unremovable {
		sd.addUnremovableNode(unremovableNode)
	}
	if err != nil {
		scaleDownStatus.Result = status.ScaleDownError
		return scaleDownStatus, err.AddPrefix("Find node to remove failed: ")
	}
	if len(nodesToRemove) == 0 {
		klog.V(1).Infof("No node to remove")
		scaleDownStatus.Result = status.ScaleDownNoNodeDeleted
		return scaleDownStatus, nil
	}
	toRemove := nodesToRemove[0]
	nodesAfterFilter := sd.filterNode([]*apiv1.Node{toRemove.Node}, nodesInfoWithoutMaster)
	if len(nodesAfterFilter) == 0 {
		scaleDownStatus.Result = status.ScaleDownNoNodeDeleted
		return scaleDownStatus, nil
	}

	utilization := sd.nodeUtilizationMap[toRemove.Node.Name]
	podNames := make([]string, 0, len(toRemove.PodsToReschedule))
	for _, pod := range toRemove.PodsToReschedule {
		podNames = append(podNames, pod.Namespace+"/"+pod.Name)
	}
	klog.V(0).Infof("Scale-down: removing node %s, utilization: %v, pods to reschedule: %s",
		toRemove.Node.Name, utilization, strings.Join(podNames, ","))
	sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDown",
		"Scale-down: removing node %s, utilization: %v, pods to reschedule: %s",
		toRemove.Node.Name, utilization, strings.Join(podNames, ","))

	// Nothing super-bad should happen if the node is removed from tracker prematurely.
	simulator.RemoveNodeFromTracker(sd.usageTracker, toRemove.Node.Name, sd.unneededNodes)
	nodeDeletionStart := time.Now()

	// Starting deletion.
	nodeDeletionDuration = time.Since(nodeDeletionStart)
	sd.nodeDeletionTracker.SetNonEmptyNodeDeleteInProgress(true)

	go func() {
		// Finishing the delete process once this goroutine is over.
		var result status.NodeDeleteResult
		defer func() { sd.nodeDeletionTracker.AddNodeDeleteResult(toRemove.Node.Name, result) }()
		defer sd.nodeDeletionTracker.SetNonEmptyNodeDeleteInProgress(false)
		nodeGroup, found := candidateNodeGroups[toRemove.Node.Name]
		if !found {
			result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete,
				Err: errors.NewAutoscalerError(
					errors.InternalError, "failed to find node group for %s", toRemove.Node.Name)}
			return
		}
		result = sd.deleteNode(toRemove.Node, toRemove.PodsToReschedule, toRemove.DaemonSetPods, nodeGroup)
		if result.ResultType != status.NodeDeleteOk {
			klog.Errorf("Failed to delete %s: %v", toRemove.Node.Name, result.Err)
			return
		}
		if readinessMap[toRemove.Node.Name] {
			metrics.RegisterScaleDown(1,
				gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, toRemove.Node, nodeGroup), metrics.Underutilized)
		} else {
			metrics.RegisterScaleDown(1,
				gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, toRemove.Node, nodeGroup), metrics.Unready)
		}
	}()

	scaleDownStatus.ScaledDownNodes = sd.mapNodesToStatusScaleDownNodes([]*apiv1.Node{toRemove.Node},
		candidateNodeGroups, map[string][]*apiv1.Pod{toRemove.Node.Name: toRemove.PodsToReschedule})
	scaleDownStatus.Result = status.ScaleDownNodeDeleteStarted
	return scaleDownStatus, nil
}

func (sd *ScaleDown) filterNode(nodes []*apiv1.Node,
	nodesInfoWithoutMaster map[string]*schedulerframework.NodeInfo) []*apiv1.Node {
	nodesAfterFilter := make([]*apiv1.Node, 0, len(nodes))
	for _, node := range nodes {
		// fix(bcs): 被移除的节点需先有 candidateTaint 或 ToBeDeletedTaint 阻止新 Pod 调度，减少极端情况
		// 否则移到下一轮循环再判断
		if !deletetaint.HasDeletionCandidateTaint(node) && !deletetaint.HasToBeDeletedTaint(node) {
			klog.V(1).Infof("node %v should be deleted after DeletionCandidateTaint or ToBeDeletedTaint", node.Name)
			continue
		}
		// filter empty nodes with ratio
		delete(nodesInfoWithoutMaster, node.Name)
		if checkResourceNotEnough(nodesInfoWithoutMaster, nil, sd.cpuRatio, sd.memRatio, sd.ratio) {
			metrics.UpdateUnremovableNodes(node.Name, BufferNotEnough, "", "")
			klog.Infof("Skip node %v due to left resource ratio", node.Name)
			continue
		}
		nodesAfterFilter = append(nodesAfterFilter, node)
	}
	return nodesAfterFilter
}

// NOCC:golint/fnsize(设计如此)
// nolint
func (sd *ScaleDown) checkNodeRemovable(nodeName string,
	unneededSince time.Time, currentTime time.Time,
	nodeGroupSize map[string]int, resourceLimiter *cloudprovider.ResourceLimiter,
	resourcesWithLimits []string,
	scaleDownResourcesLeft scaleDownResourcesLimits,
	readinessMap map[string]bool) (map[string]bool, cloudprovider.NodeGroup, error) {
	var checkErr error
	var nodeGroup cloudprovider.NodeGroup

	nodeInfo, err := sd.context.ClusterSnapshot.NodeInfos().Get(nodeName)
	if err != nil {
		checkErr = fmt.Errorf("Can't retrieve unneeded node %s from snapshot, err: %v", nodeName, err)
		return readinessMap, nodeGroup, checkErr
	}

	node := nodeInfo.Node()
	// Check if node is marked with no scale down annotation.
	if hasNoScaleDownAnnotation(node) {
		checkErr = fmt.Errorf("Skipping %s - scale down disabled annotation found", node.Name)
		sd.addUnremovableNodeReason(node, simulator.ScaleDownDisabledAnnotation)
		metrics.UpdateUnremovableNodes(node.Name, ScaleDownDisabledAnnotation, "", "")
		return readinessMap, nodeGroup, checkErr
	}

	ready, _, _ := kube_util.GetReadinessState(node)
	readinessMap[node.Name] = ready

	nodeGroup, err = sd.context.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		checkErr = fmt.Errorf("Error while checking node group for %s: %v", node.Name, err)
		sd.addUnremovableNodeReason(node, simulator.UnexpectedError)
		metrics.UpdateUnremovableNodes(node.Name, UnexpectedError, "", "")
		return readinessMap, nodeGroup, checkErr
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		checkErr = fmt.Errorf("Skipping %s - no node group config", node.Name)
		sd.addUnremovableNodeReason(node, simulator.NotAutoscaled)
		metrics.UpdateUnremovableNodes(node.Name, NotAutoscaled, "", "")
		return readinessMap, nodeGroup, checkErr
	}

	if ready {
		// Check how long a ready node was underutilized.
		unneededTime, getErr := sd.processors.NodeGroupConfigProcessor.GetScaleDownUnneededTime(
			sd.context, nodeGroup)
		if getErr != nil {
			checkErr = fmt.Errorf("Error trying to get ScaleDownUnneededTime for node %s (in group: %s)",
				node.Name, nodeGroup.Id())
			return readinessMap, nodeGroup, checkErr
		}
		if !unneededSince.Add(unneededTime).Before(currentTime) {
			checkErr = fmt.Errorf("Skipping %s - unneeded not long enough", node.Name)
			sd.addUnremovableNodeReason(node, simulator.NotUnneededLongEnough)
			metrics.UpdateUnremovableNodes(node.Name, NotUnneededLongEnough, "", "")
			return readinessMap, nodeGroup, checkErr
		}
	} else {
		// Unready nodes may be deleted after a different time than underutilized nodes.
		unreadyTime, getErr := sd.processors.NodeGroupConfigProcessor.GetScaleDownUnreadyTime(
			sd.context, nodeGroup)
		if getErr != nil {
			checkErr = fmt.Errorf("Error trying to get ScaleDownUnreadyTime for node %s (in group: %s)",
				node.Name, nodeGroup.Id())
			return readinessMap, nodeGroup, checkErr
		}
		if !unneededSince.Add(unreadyTime).Before(currentTime) {
			checkErr = fmt.Errorf("Skipping %s - unready not long enough", node.Name)
			sd.addUnremovableNodeReason(node, simulator.NotUnreadyLongEnough)
			metrics.UpdateUnremovableNodes(node.Name, NotUnreadyLongEnough, "", "")
			return readinessMap, nodeGroup, checkErr
		}
	}

	size, found := nodeGroupSize[nodeGroup.Id()]
	if !found {
		checkErr = fmt.Errorf("Error while checking node group size %s: group size not found in cache", nodeGroup.Id())
		sd.addUnremovableNodeReason(node, simulator.UnexpectedError)
		metrics.UpdateUnremovableNodes(node.Name, UnexpectedError, "", "")
		return readinessMap, nodeGroup, checkErr
	}

	deletionsInProgress := sd.nodeDeletionTracker.GetDeletionsInProgress(nodeGroup.Id())
	if size-deletionsInProgress <= nodeGroup.MinSize() {
		checkErr = fmt.Errorf("Skipping %s - node group min size reached", node.Name)
		sd.addUnremovableNodeReason(node, simulator.NodeGroupMinSizeReached)
		metrics.UpdateUnremovableNodes(node.Name, NodeGroupMinSizeReached, "", "")
		return readinessMap, nodeGroup, checkErr
	}

	scaleDownResourceDelta, err := sd.computeScaleDownResourcesDelta(
		sd.context.CloudProvider, node, nodeGroup, resourcesWithLimits)
	if err != nil {
		checkErr = fmt.Errorf("Error getting node %s resources: %v", node.Name, err)
		sd.addUnremovableNodeReason(node, simulator.UnexpectedError)
		metrics.UpdateUnremovableNodes(node.Name, UnexpectedError, "", "")
		return readinessMap, nodeGroup, checkErr
	}

	checkResult := scaleDownResourcesLeft.checkScaleDownDeltaWithinLimits(scaleDownResourceDelta)
	if checkResult.exceeded {
		checkErr = fmt.Errorf("Skipping %s - minimal limit exceeded for %v", node.Name, checkResult.exceededResources)
		metrics.UpdateUnremovableNodes(node.Name, MinimalResourceLimitExceeded, "", "")
		for _, resource := range checkResult.exceededResources {
			switch resource {
			case cloudprovider.ResourceNameCores:
				metrics.RegisterSkippedScaleDownCPU()
			case cloudprovider.ResourceNameMemory:
				metrics.RegisterSkippedScaleDownMemory()
			default:
				continue
			}
		}
		return readinessMap, nodeGroup, checkErr
	}
	return readinessMap, nodeGroup, checkErr
}

// updateScaleDownMetrics registers duration of different parts of scale down.
// Separates time spent on finding nodes to remove, deleting nodes and other operations.
func updateScaleDownMetrics(scaleDownStart time.Time, findNodesToRemoveDuration *time.Duration,
	nodeDeletionDuration *time.Duration) {
	stop := time.Now()
	miscDuration := stop.Sub(scaleDownStart) - *nodeDeletionDuration - *findNodesToRemoveDuration
	metrics.UpdateDuration(metrics.ScaleDownNodeDeletion, *nodeDeletionDuration)
	metrics.UpdateDuration(metrics.ScaleDownFindNodesToRemove, *findNodesToRemoveDuration)
	metrics.UpdateDuration(metrics.ScaleDownMiscOperations, miscDuration)
}

func (sd *ScaleDown) getEmptyNodesNoResourceLimits(candidates []string,
	maxEmptyBulkDelete int, timestamp time.Time) []*apiv1.Node {
	return sd.getEmptyNodes(candidates, maxEmptyBulkDelete,
		noScaleDownLimitsOnResources(), timestamp)
}

// This functions finds empty nodes among passed candidates and returns a list of empty nodes
// that can be deleted at the same time.
// This functions finds empty nodes among passed candidates and returns a list of empty nodes
// that can be deleted at the same time.
func (sd *ScaleDown) getEmptyNodes(candidates []string, maxEmptyBulkDelete int,
	resourcesLimits scaleDownResourcesLimits, timestamp time.Time) []*apiv1.Node {

	emptyNodes := simulator.FindEmptyNodesToRemove(sd.context.ClusterSnapshot, candidates, timestamp)
	// 空节点列表排序，避免空节点过多时循环等待
	sort.SliceStable(emptyNodes, func(i, j int) bool {
		return emptyNodes[i] < emptyNodes[j]
	})
	availabilityMap := make(map[string]int)
	result := make([]*apiv1.Node, 0)
	resourcesLimitsCopy := copyScaleDownResourcesLimits(resourcesLimits) // we do not want to modify input parameter
	resourcesNames := sets.StringKeySet(resourcesLimits).List()
	for _, nodeName := range emptyNodes {
		nodeInfo, err := sd.context.ClusterSnapshot.NodeInfos().Get(nodeName)
		if err != nil {
			klog.Errorf("Can't retrieve node %s from snapshot, err: %v", nodeName, err)
			continue
		}
		node := nodeInfo.Node()
		nodeGroup, err := sd.context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			klog.Errorf("Failed to get group for %s", nodeName)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			continue
		}
		var available int
		var found bool
		if available, found = availabilityMap[nodeGroup.Id()]; !found {
			// Will be cached.
			size, err := nodeGroup.TargetSize()
			if err != nil {
				klog.Errorf("Failed to get size for %s: %v ", nodeGroup.Id(), err)
				continue
			}
			deletionsInProgress := sd.nodeDeletionTracker.GetDeletionsInProgress(nodeGroup.Id())
			available = size - nodeGroup.MinSize() - deletionsInProgress
			if available < 0 {
				available = 0
			}
			availabilityMap[nodeGroup.Id()] = available
		}
		if available > 0 {
			resourcesDelta, err := sd.computeScaleDownResourcesDelta(sd.context.CloudProvider,
				node, nodeGroup, resourcesNames)
			if err != nil {
				klog.Errorf("Error: %v", err)
				continue
			}
			checkResult := resourcesLimitsCopy.tryDecrementLimitsByDelta(resourcesDelta)
			if checkResult.exceeded {
				continue
			}
			available--
			availabilityMap[nodeGroup.Id()] = available
			result = append(result, node)
		}
	}
	limit := maxEmptyBulkDelete
	if len(result) < limit {
		limit = len(result)
	}
	return result[:limit]
}

func (sd *ScaleDown) scheduleDeleteEmptyNodes(emptyNodes []*apiv1.Node, client kube_client.Interface,
	recorder kube_record.EventRecorder, readinessMap map[string]bool,
	candidateNodeGroups map[string]cloudprovider.NodeGroup) ([]*apiv1.Node, errors.AutoscalerError) {
	deletedNodes := []*apiv1.Node{}
	for _, node := range emptyNodes {
		klog.V(0).Infof("Scale-down: removing empty node %s", node.Name)
		sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty",
			"Scale-down: removing empty node %s", node.Name)
		simulator.RemoveNodeFromTracker(sd.usageTracker, node.Name, sd.unneededNodes)
		nodeGroup, found := candidateNodeGroups[node.Name]
		if !found {
			return deletedNodes, errors.NewAutoscalerError(
				errors.CloudProviderError, "failed to find node group for %s", node.Name)
		}
		taintErr := deletetaint.MarkToBeDeleted(node, client, sd.context.CordonNodeBeforeTerminate)
		if taintErr != nil {
			recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed",
				"failed to mark the node as toBeDeleted/unschedulable: %v", taintErr)
			return deletedNodes, errors.ToAutoscalerError(errors.ApiCallError, taintErr)
		}
		deletedNodes = append(deletedNodes, node)
		go func(nodeToDelete *apiv1.Node, nodeGroupForDeletedNode cloudprovider.NodeGroup, evictByDefault bool) {
			sd.nodeDeletionTracker.StartDeletion(nodeGroupForDeletedNode.Id())
			defer sd.nodeDeletionTracker.EndDeletion(nodeGroupForDeletedNode.Id())
			var result status.NodeDeleteResult
			defer func() { sd.nodeDeletionTracker.AddNodeDeleteResult(nodeToDelete.Name, result) }()

			var deleteErr errors.AutoscalerError
			// If we fail to delete the node we want to remove delete taint
			defer func() {
				if deleteErr != nil {
					_, cleanErr := deletetaint.CleanToBeDeleted(nodeToDelete, client, sd.context.CordonNodeBeforeTerminate)
					if cleanErr != nil {
						klog.Errorf("CleanToBeDeleted failed. Error: %v", cleanErr)
					}
					recorder.Eventf(nodeToDelete, apiv1.EventTypeWarning, "ScaleDownFailed",
						"failed to delete empty node: %v", deleteErr)
				} else {
					sd.context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaleDownEmpty",
						"Scale-down: empty node %s removed", nodeToDelete.Name)
				}
			}()
			if err := evictDaemonSetPods(sd.context.ClusterSnapshot, nodeToDelete, client,
				sd.context.MaxGracefulTerminationSec, time.Now(), DaemonSetEvictionEmptyNodeTimeout,
				DeamonSetTimeBetweenEvictionRetries, recorder, evictByDefault); err != nil {
				klog.Warningf("error while evicting DS pods from an empty node: %v", err)
			}
			deleteErr = waitForDelayDeletion(nodeToDelete, sd.context.ListerRegistry.AllNodeLister(),
				sd.context.AutoscalingOptions.NodeDeletionDelayTimeout)
			if deleteErr != nil {
				klog.Errorf("Problem with empty node deletion: %v", deleteErr)
				result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: deleteErr}
				return
			}
			deleteErr = deleteNodeFromCloudProvider(nodeToDelete, sd.context.CloudProvider,
				sd.context.Recorder, sd.clusterStateRegistry)
			if deleteErr != nil {
				klog.Errorf("Problem with empty node deletion: %v", deleteErr)
				result = status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete, Err: deleteErr}
				return
			}
			if readinessMap[nodeToDelete.Name] {
				metrics.RegisterScaleDown(1, gpu.GetGpuTypeForMetrics(sd.context.CloudProvider.GPULabel(),
					sd.context.CloudProvider.GetAvailableGPUTypes(), nodeToDelete, nodeGroupForDeletedNode), metrics.Empty)
			} else {
				metrics.RegisterScaleDown(1, gpu.GetGpuTypeForMetrics(sd.context.CloudProvider.GPULabel(),
					sd.context.CloudProvider.GetAvailableGPUTypes(), nodeToDelete, nodeGroupForDeletedNode), metrics.Unready)
			}
			result = status.NodeDeleteResult{ResultType: status.NodeDeleteOk}
		}(node, nodeGroup, sd.context.DaemonSetEvictionForEmptyNodes)
	}
	return deletedNodes, nil
}

// Create eviction object for all DaemonSet pods on the node
func evictDaemonSetPods(clusterSnapshot simulator.ClusterSnapshot, nodeToDelete *apiv1.Node,
	client kube_client.Interface, maxGracefulTerminationSec int, timeNow time.Time,
	dsEvictionTimeout time.Duration, waitBetweenRetries time.Duration,
	recorder kube_record.EventRecorder, evictByDefault bool) error {
	nodeInfo, err := clusterSnapshot.NodeInfos().Get(nodeToDelete.Name)
	if err != nil {
		return fmt.Errorf("failed to get node info for %s", nodeToDelete.Name)
	}
	_, daemonSetPods, _, err := simulator.FastGetPodsToMove(nodeInfo, true, true,
		[]*policyv1.PodDisruptionBudget{}, timeNow)
	if err != nil {
		return fmt.Errorf("failed to get DaemonSet pods for %s (error: %v)", nodeToDelete.Name, err)
	}

	daemonSetPods = daemonset.PodsToEvict(daemonSetPods, evictByDefault)

	dsEviction := make(chan status.PodEvictionResult, len(daemonSetPods))

	// Perform eviction of DaemonSet pods
	for _, daemonSetPod := range daemonSetPods {
		go func(podToEvict *apiv1.Pod) {
			dsEviction <- evictPod(podToEvict, true, client, recorder, maxGracefulTerminationSec,
				timeNow.Add(dsEvictionTimeout), waitBetweenRetries)
		}(daemonSetPod)
	}
	// Wait for creating eviction of DaemonSet pods
	var failedPodErrors []string
	for range daemonSetPods {
		select {
		case res := <-dsEviction:
			if res.Err != nil {
				failedPodErrors = append(failedPodErrors, res.Err.Error())
			}
		// adding waitBetweenRetries in order to have a bigger time interval than evictPod()
		case <-time.After(dsEvictionTimeout):
			return fmt.Errorf("failed to create DaemonSet eviction for %v seconds on the %s",
				dsEvictionTimeout, nodeToDelete.Name)
		}
	}
	if len(failedPodErrors) > 0 {
		return fmt.Errorf("following DaemonSet pod failed to evict on the %s:\n%s",
			nodeToDelete.Name, fmt.Errorf(strings.Join(failedPodErrors, "\n")))
	}
	return nil
}

func (sd *ScaleDown) deleteNode(node *apiv1.Node, pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod,
	nodeGroup cloudprovider.NodeGroup) status.NodeDeleteResult {
	deleteSuccessful := false
	drainSuccessful := false

	if err := deletetaint.MarkToBeDeleted(node, sd.context.ClientSet, sd.context.CordonNodeBeforeTerminate); err != nil {
		sd.context.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed",
			"failed to mark the node as toBeDeleted/unschedulable: %v", err)
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToMarkToBeDeleted,
			Err: errors.ToAutoscalerError(errors.ApiCallError, err)}
	}

	var err error

	sd.nodeDeletionTracker.StartDeletion(nodeGroup.Id())
	defer sd.nodeDeletionTracker.EndDeletion(nodeGroup.Id())

	// If we fail to evict all the pods from the node we want to remove delete taint
	defer func() {
		if !deleteSuccessful {
			_, _ = deletetaint.CleanToBeDeleted(node, sd.context.ClientSet,
				sd.context.CordonNodeBeforeTerminate)
			if !drainSuccessful {
				sd.context.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed",
					"failed to drain the node, aborting ScaleDown")
			} else {
				sd.context.Recorder.Eventf(node, apiv1.EventTypeWarning, "ScaleDownFailed",
					"failed to delete the node")
			}
		}
	}()

	var podsToDrain []*apiv1.Pod
	if sd.evictLatest {
		podsToDrain, err = getLatestPodsToDrain(sd, node)
		if err != nil {
			return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToEvictPods, Err: err}
		}
	} else {
		podsToDrain = pods
	}

	sd.context.Recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown",
		"marked the node as toBeDeleted/unschedulable")

	daemonSetPods = daemonset.PodsToEvict(daemonSetPods, sd.context.DaemonSetEvictionForOccupiedNodes)

	// attempt drain
	evictionResults, err := drainNode(node, podsToDrain, daemonSetPods, sd.context.ClientSet,
		sd.context.Recorder, sd.context.MaxGracefulTerminationSec, MaxPodEvictionTime,
		EvictionRetryTime, PodEvictionHeadroom)
	if err != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToEvictPods,
			Err: err, PodEvictionResults: evictionResults}
	}
	drainSuccessful = true

	if typedErr := waitForDelayDeletion(node, sd.context.ListerRegistry.AllNodeLister(),
		sd.context.AutoscalingOptions.NodeDeletionDelayTimeout); typedErr != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete,
			Err: typedErr}
	}

	// attempt delete from cloud provider

	if typedErr := deleteNodeFromCloudProvider(node, sd.context.CloudProvider,
		sd.context.Recorder, sd.clusterStateRegistry); typedErr != nil {
		return status.NodeDeleteResult{ResultType: status.NodeDeleteErrorFailedToDelete,
			Err: typedErr}
	}

	deleteSuccessful = true // Let the deferred function know there is no need to cleanup
	return status.NodeDeleteResult{ResultType: status.NodeDeleteOk}
}

func getLatestPodsToDrain(sd *ScaleDown, node *apiv1.Node) ([]*apiv1.Pod, error) {
	// fix(bcs): 获取最新的 drain pod 列表
	fieldSelector := fmt.Sprintf("spec.nodeName=%s", node.Name)
	podList, listErr := sd.context.ClientSet.CoreV1().Pods(apiv1.NamespaceAll).List(
		ctx.TODO(), metav1.ListOptions{FieldSelector: fieldSelector})
	if listErr != nil {
		return nil, listErr
	}
	podsOfNode := make([]*apiv1.Pod, 0)
	for i := range podList.Items {
		podsOfNode = append(podsOfNode, &podList.Items[i])
	}
	// fix(bcs): 过滤低优先级 pod
	unexpendablePods := filterOutExpendablePods(podsOfNode, sd.expendablePodsPriorityCutoff)
	// fix(bcs): 过滤 dpm pod
	unexpendablePods = filterOutDpmPods(unexpendablePods)
	podsToDrain, _, _, getErr := drain.GetPodsForDeletionOnNodeDrain(
		unexpendablePods, []*policyv1.PodDisruptionBudget{}, false, false, false,
		nil, 0, time.Now())
	if getErr != nil {
		return nil, getErr
	}
	return podsToDrain, nil
}

func evictPod(podToEvict *apiv1.Pod, isDaemonSetPod bool, client kube_client.Interface,
	recorder kube_record.EventRecorder, maxGracefulTerminationSec int, retryUntil time.Time,
	waitBetweenRetries time.Duration) status.PodEvictionResult {
	recorder.Eventf(podToEvict, apiv1.EventTypeNormal, "ScaleDown",
		"deleting pod for node scale down")

	maxTermination := int64(apiv1.DefaultTerminationGracePeriodSeconds)
	if podToEvict.Spec.TerminationGracePeriodSeconds != nil {
		if *podToEvict.Spec.TerminationGracePeriodSeconds < int64(maxGracefulTerminationSec) {
			maxTermination = *podToEvict.Spec.TerminationGracePeriodSeconds
		} else {
			maxTermination = int64(maxGracefulTerminationSec)
		}
	}

	var lastError error
	for first := true; first || time.Now().Before(retryUntil); time.Sleep(waitBetweenRetries) {
		first = false
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: podToEvict.Namespace,
				Name:      podToEvict.Name,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: &maxTermination,
			},
		}
		lastError = client.CoreV1().Pods(podToEvict.Namespace).Evict(ctx.TODO(), eviction)
		if lastError == nil || kube_errors.IsNotFound(lastError) {
			return status.PodEvictionResult{Pod: podToEvict, TimedOut: false, Err: nil}
		}
	}
	if !isDaemonSetPod {
		klog.Errorf("Failed to evict pod %s, error: %v", podToEvict.Name, lastError)
		recorder.Eventf(podToEvict, apiv1.EventTypeWarning, "ScaleDownFailed",
			"failed to delete pod for ScaleDown")
	}
	return status.PodEvictionResult{Pod: podToEvict, TimedOut: true,
		Err: fmt.Errorf("failed to evict pod %s/%s within allowed timeout (last error: %v)",
			podToEvict.Namespace, podToEvict.Name, lastError)}
}

// Performs drain logic on the node. Marks the node as unschedulable and later removes all pods, giving
// them up to MaxGracefulTerminationTime to finish.
// NOCC:golint/fnsize(设计如此)
// nolint funlen
func drainNode(node *apiv1.Node, pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod,
	client kube_client.Interface, recorder kube_record.EventRecorder,
	maxGracefulTerminationSec int, maxPodEvictionTime time.Duration, waitBetweenRetries time.Duration,
	podEvictionHeadroom time.Duration) (evictionResults map[string]status.PodEvictionResult, err error) {

	evictionResults = make(map[string]status.PodEvictionResult)
	retryUntil := time.Now().Add(maxPodEvictionTime)
	confirmations := make(chan status.PodEvictionResult, len(pods))
	daemonSetConfirmations := make(chan status.PodEvictionResult, len(daemonSetPods))
	for _, pod := range pods {
		evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: nil}
		go func(podToEvict *apiv1.Pod) {
			confirmations <- evictPod(podToEvict, false, client, recorder,
				maxGracefulTerminationSec, retryUntil, waitBetweenRetries)
		}(pod)
	}

	// Perform eviction of daemonset. We don't want to raise an error if daemonsetPod wasn't evict properly
	for _, daemonSetPod := range daemonSetPods {
		go func(podToEvict *apiv1.Pod) {
			daemonSetConfirmations <- evictPod(podToEvict, true, client, recorder,
				maxGracefulTerminationSec, retryUntil, waitBetweenRetries)
		}(daemonSetPod)

	}

	podsEvictionCounter := 0
	for i := 0; i < len(pods)+len(daemonSetPods); i++ {
		select {
		case evictionResult := <-confirmations:
			podsEvictionCounter++
			evictionResults[evictionResult.Pod.Name] = evictionResult
			if evictionResult.WasEvictionSuccessful() {
				metrics.RegisterEvictions(1)
			}
		case <-daemonSetConfirmations:
		case <-time.After(retryUntil.Sub(time.Now()) + 5*time.Second): // nolint
			if podsEvictionCounter < len(pods) {
				// All pods initially had results with TimedOut set to true, so the ones that
				// didn't receive an actual result are correctly marked as timed out.
				return evictionResults, errors.NewAutoscalerError(errors.ApiCallError,
					"Failed to drain node %s/%s: timeout when waiting for creating evictions",
					node.Namespace, node.Name)
			}
			klog.Infof("Timeout when waiting for creating daemonSetPods eviction")
		}
	}

	evictionErrs := make([]error, 0)
	for _, result := range evictionResults {
		if !result.WasEvictionSuccessful() {
			evictionErrs = append(evictionErrs, result.Err)
		}
	}
	if len(evictionErrs) != 0 {
		return evictionResults, errors.NewAutoscalerError(errors.ApiCallError,
			"Failed to drain node %s/%s, due to following errors: %v",
			node.Namespace, node.Name, evictionErrs)
	}

	// Evictions created successfully, wait maxGracefulTerminationSec + podEvictionHeadroom to
	// see if pods really disappeared.
	var allGone bool
	for start := time.Now(); time.Since(start) < time.Duration(
		maxGracefulTerminationSec)*time.Second+podEvictionHeadroom; time.Sleep(5 * time.Second) {
		allGone = true
		for _, pod := range pods {
			podreturned, err := client.CoreV1().Pods(pod.Namespace).Get(ctx.TODO(), pod.Name, metav1.GetOptions{})
			if err == nil && (podreturned == nil || podreturned.Spec.NodeName == node.Name) {
				klog.V(1).Infof("Not deleted yet %s/%s", pod.Namespace, pod.Name)
				allGone = false
				break
			}
			if err != nil && !kube_errors.IsNotFound(err) {
				klog.Errorf("Failed to check pod %s/%s: %v", pod.Namespace, pod.Name, err)
				allGone = false
				break
			}
		}
		if allGone {
			klog.V(1).Infof("All pods removed from %s", node.Name)
			// Let the deferred function know there is no need for cleanup
			return evictionResults, nil
		}
	}

	for _, pod := range pods {
		podReturned, err := client.CoreV1().Pods(pod.Namespace).Get(ctx.TODO(), pod.Name, metav1.GetOptions{})
		if err == nil && (podReturned == nil || podReturned.Spec.NodeName == node.Name) {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: nil}
		} else if err != nil && !kube_errors.IsNotFound(err) {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: true, Err: err}
		} else {
			evictionResults[pod.Name] = status.PodEvictionResult{Pod: pod, TimedOut: false, Err: nil}
		}
	}

	return evictionResults, errors.NewAutoscalerError(errors.TransientError,
		"Failed to drain node %s/%s: pods remaining after timeout", node.Namespace, node.Name)
}

// Removes the given node from cloud provider. No extra pre-deletion actions are executed on
// the Kubernetes side.
func deleteNodeFromCloudProvider(node *apiv1.Node, cloudProvider cloudprovider.CloudProvider,
	recorder kube_record.EventRecorder, registry *clusterstate.ClusterStateRegistry) errors.AutoscalerError {
	nodeGroup, err := cloudProvider.NodeGroupForNode(node)
	if err != nil {
		return errors.NewAutoscalerError(
			errors.CloudProviderError, "failed to find node group for %s: %v", node.Name, err)
	}
	if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
		return errors.NewAutoscalerError(errors.InternalError,
			"picked node that doesn't belong to a node group: %s", node.Name)
	}
	if err = nodeGroup.DeleteNodes([]*apiv1.Node{node}); err != nil {
		return errors.NewAutoscalerError(errors.CloudProviderError, "failed to delete %s: %v", node.Name, err)
	}
	recorder.Eventf(node, apiv1.EventTypeNormal, "ScaleDown", "node removed by cluster autoscaler")
	registry.RegisterScaleDown(&clusterstate.ScaleDownRequest{
		NodeGroup:          nodeGroup,
		NodeName:           node.Name,
		Time:               time.Now(),
		ExpectedDeleteTime: time.Now().Add(MaxCloudProviderNodeDeletionTime),
	})
	return nil
}

func waitForDelayDeletion(node *apiv1.Node, nodeLister kubernetes.NodeLister,
	timeout time.Duration) errors.AutoscalerError {
	if timeout != 0 && hasDelayDeletionAnnotation(node) {
		klog.V(1).Infof("Wait for removing %s annotations on node %v",
			DelayDeletionAnnotationPrefix, node.Name)
		err := wait.Poll(5*time.Second, timeout, func() (bool, error) {
			klog.V(5).Infof("Waiting for removing %s annotations on node %v",
				DelayDeletionAnnotationPrefix, node.Name)
			freshNode, err := nodeLister.Get(node.Name)
			if err != nil || freshNode == nil {
				return false, fmt.Errorf("failed to get node %v: %v", node.Name, err)
			}
			return !hasDelayDeletionAnnotation(freshNode), nil
		})
		if err != nil && err != wait.ErrWaitTimeout {
			return errors.ToAutoscalerError(errors.ApiCallError, err)
		}
		if err == wait.ErrWaitTimeout {
			klog.Warningf("Delay node deletion timed out for node %v, delay deletion annotation wasn't removed within %v,"+
				"this might slow down scale down.", node.Name, timeout)
		} else {
			klog.V(2).Infof("Annotation %s removed from node %v", DelayDeletionAnnotationPrefix, node.Name)
		}
	}
	return nil
}

func hasDelayDeletionAnnotation(node *apiv1.Node) bool {
	for annotation := range node.Annotations {
		if strings.HasPrefix(annotation, DelayDeletionAnnotationPrefix) {
			return true
		}
	}
	return false
}

func hasNoScaleDownAnnotation(node *apiv1.Node) bool {
	return node.Annotations[ScaleDownDisabledKey] == "true"
}

const (
	apiServerLabelKey   = "component"
	apiServerLabelValue = "kube-apiserver"
)

func isMasterNode(nodeInfo *schedulerframework.NodeInfo) bool {
	for _, podInfo := range nodeInfo.Pods {
		if podInfo.Pod.Namespace == metav1.NamespaceSystem && podInfo.Pod.Labels[apiServerLabelKey] == apiServerLabelValue {
			return true
		}
	}
	return false
}

func filterOutMasters(nodeInfos []*schedulerframework.NodeInfo) []*apiv1.Node {
	result := make([]*apiv1.Node, 0, len(nodeInfos))
	for _, nodeInfo := range nodeInfos {
		if !isMasterNode(nodeInfo) {
			result = append(result, nodeInfo.Node())
		}
	}
	return result
}

func filterOutMastersNodeInfo(nodeInfos []*schedulerframework.NodeInfo) map[string]*schedulerframework.NodeInfo {
	result := make(map[string]*schedulerframework.NodeInfo, 0)
	for _, nodeInfo := range nodeInfos {
		if !isMasterNode(nodeInfo) {
			result[nodeInfo.Node().Name] = nodeInfo
		}
	}
	return result
}

func sortNodesByDeletionCost(candidates []string,
	clusterSnapshot simulator.ClusterSnapshot) []string {
	if len(candidates) <= 1 {
		return candidates
	}
	nodes := make([]*apiv1.Node, 0)
	for _, name := range candidates {
		nodeInfo, err := clusterSnapshot.NodeInfos().Get(name)
		if err != nil {
			klog.Errorf("sortNodesByDeletionCost: cannot get node info, err: %v", err)
			return candidates
		}
		nodes = append(nodes, nodeInfo.Node())
	}
	sort.Slice(nodes, func(i, j int) bool {
		costI := getCostFromNode(nodes[i])
		costJ := getCostFromNode(nodes[j])
		return costI < costJ
	})
	nodeNames := make([]string, 0, len(nodes))
	for i := range nodes {
		nodeNames = append(nodeNames, nodes[i].Name)
	}
	return nodeNames
}

func getCostFromNode(node *apiv1.Node) float64 {
	costAnnotation := node.Annotations[NodeDeletionCost]
	if len(costAnnotation) == 0 {
		return 0
	}
	cost, err := strconv.ParseFloat(costAnnotation, 64)
	if err != nil {
		return 0
	}
	return cost
}

func convertUnRemovableNodeReason(reason simulator.UnremovableReason) string {
	switch reason {
	case simulator.NoReason:
		return NoReason
	case simulator.ScaleDownDisabledAnnotation:
		return ScaleDownDisabledAnnotation
	case simulator.NotAutoscaled:
		return NotAutoscaled
	case simulator.NotUnneededLongEnough:
		return NotUnneededLongEnough
	case simulator.NotUnreadyLongEnough:
		return NotUnreadyLongEnough
	case simulator.NodeGroupMinSizeReached:
		return NodeGroupMinSizeReached
	case simulator.MinimalResourceLimitExceeded:
		return MinimalResourceLimitExceeded
	case simulator.CurrentlyBeingDeleted:
		return CurrentlyBeingDeleted
	case simulator.NotUnderutilized:
		return NotUnderutilized
	case simulator.NotUnneededOtherReason:
		return NotUnneededOtherReason
	case simulator.RecentlyUnremovable:
		return RecentlyUnremovable
	case simulator.NoPlaceToMovePods:
		return NoPlaceToMovePods
	case simulator.BlockedByPod:
		return BlockedByPod
	case simulator.UnexpectedError:
		return UnexpectedError
	}
	return ""
}

func convertBlockingPodReason(reason drain.BlockingPodReason) string {
	switch reason {
	case drain.NoReason:
		return NoReason
	case drain.ControllerNotFound:
		return ControllerNotFound
	case drain.MinReplicasReached:
		return MinReplicasReached
	case drain.NotReplicated:
		return NotReplicated
	case drain.LocalStorageRequested:
		return LocalStorageRequested
	case drain.NotSafeToEvictAnnotation:
		return NotSafeToEvictAnnotation
	case drain.UnmovableKubeSystemPod:
		return UnmovableKubeSystemPod
	case drain.NotEnoughPdb:
		return NotEnoughPdb
	case drain.UnexpectedError:
		return UnexpectedError
	}
	return ""
}
