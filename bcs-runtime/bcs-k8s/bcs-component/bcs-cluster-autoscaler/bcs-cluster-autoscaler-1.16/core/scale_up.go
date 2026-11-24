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
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/glogx"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
	estimatorinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/estimator"
)

type (
	scaleUpResourcesLimits map[string]int64
	scaleUpResourcesDelta  map[string]int64
)

// used as a value in scaleUpResourcesLimits if actual limit could not be obtained
// due to errors talking to cloud provider
const scaleUpLimitUnknown = math.MaxInt64

// computeScaleUpResourcesLeftLimits computes the resource limits of scale up
func computeScaleUpResourcesLeftLimits(
	cp cloudprovider.CloudProvider,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node,
	resourceLimiter *cloudprovider.ResourceLimiter,
) (scaleUpResourcesLimits, errors.AutoscalerError) {
	totalCores, totalMem, errCoresMem := calculateScaleUpCoresMemoryTotal(nodeGroups, nodeInfos,
		nodesFromNotAutoscaledGroups)

	var totalGpus map[string]int64
	var totalGpusErr error
	if cloudprovider.ContainsGpuResources(resourceLimiter.GetResources()) {
		totalGpus, totalGpusErr = calculateScaleUpGpusTotal(cp.GPULabel(), nodeGroups, nodeInfos,
			nodesFromNotAutoscaledGroups)
	}

	resultScaleUpLimits := make(scaleUpResourcesLimits)
	for _, resource := range resourceLimiter.GetResources() {
		max := resourceLimiter.GetMax(resource)

		// we put only actual limits into final map. No entry means no limit.
		if max > 0 {
			if (resource == cloudprovider.ResourceNameCores || resource == cloudprovider.ResourceNameMemory) &&
				errCoresMem != nil {
				// core resource info missing - no reason to proceed with scale up
				return scaleUpResourcesLimits{}, errCoresMem
			}
			switch {
			case resource == cloudprovider.ResourceNameCores:
				if errCoresMem != nil {
					resultScaleUpLimits[resource] = scaleUpLimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalCores, max)
				}

			case resource == cloudprovider.ResourceNameMemory:
				if errCoresMem != nil {
					resultScaleUpLimits[resource] = scaleUpLimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalMem, max)
				}

			case cloudprovider.IsGpuResource(resource):
				if totalGpusErr != nil {
					resultScaleUpLimits[resource] = scaleUpLimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalGpus[resource], max)
				}

			default:
				klog.Errorf("Scale up limits defined for unsupported resource '%s'", resource)
			}
		}
	}

	return resultScaleUpLimits, nil
}

// calculateScaleUpCoresMemoryTotal calculate the total cores and memory
func calculateScaleUpCoresMemoryTotal(
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node,
) (int64, int64, errors.AutoscalerError) {
	var coresTotal int64
	var memoryTotal int64

	for _, nodeGroup := range nodeGroups {
		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			return 0, 0, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
				"Failed to get node group size of %v:", nodeGroup.Id())
		}
		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			return 0, 0, errors.NewAutoscalerError(errors.CloudProviderError, "No node info for: %s", nodeGroup.Id())
		}
		if currentSize > 0 {
			nodeCPU, nodeMemory := getNodeInfoCoresAndMemory(nodeInfo)
			coresTotal += int64(currentSize) * nodeCPU
			memoryTotal += int64(currentSize) * nodeMemory
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		cores, memory := getNodeCoresAndMemory(node)
		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal, nil
}

// calculateScaleUpGpusTotal calculate the total gpu
func calculateScaleUpGpusTotal(
	gpuLabel string,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node,
) (map[string]int64, errors.AutoscalerError) {
	result := make(map[string]int64)
	for _, nodeGroup := range nodeGroups {
		currentSize, err := nodeGroup.TargetSize()
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
				"Failed to get node group size of %v:", nodeGroup.Id())
		}
		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			return nil, errors.NewAutoscalerError(errors.CloudProviderError, "No node info for: %s", nodeGroup.Id())
		}
		if currentSize > 0 {
			gpuType, gpuCount, err := gpu.GetNodeTargetGpus(gpuLabel, nodeInfo.Node(), nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
					"Failed to get target gpu for node group %v:", nodeGroup.Id())
			}
			if gpuType == "" {
				continue
			}
			result[gpuType] += gpuCount * int64(currentSize)
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		gpuType, gpuCount, err := gpu.GetNodeTargetGpus(gpuLabel, node, nil)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
				"Failed to get target gpu for node gpus count for node %v:", node.Name)
		}
		result[gpuType] += gpuCount
	}

	return result, nil
}

func computeBelowMax(total int64, max int64) int64 {
	if total < max {
		return max - total
	}
	return 0
}

// computeScaleUpResourcesDelta compute scale up resources delta
func computeScaleUpResourcesDelta(cp cloudprovider.CloudProvider, nodeInfo *schedulernodeinfo.NodeInfo,
	nodeGroup cloudprovider.NodeGroup, resourceLimiter *cloudprovider.ResourceLimiter) (scaleUpResourcesDelta,
	errors.AutoscalerError,
) {
	resultScaleUpDelta := make(scaleUpResourcesDelta)

	nodeCPU, nodeMemory := getNodeInfoCoresAndMemory(nodeInfo)
	resultScaleUpDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleUpDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	if cloudprovider.ContainsGpuResources(resourceLimiter.GetResources()) {
		gpuType, gpuCount, err := gpu.GetNodeTargetGpus(cp.GPULabel(), nodeInfo.Node(), nodeGroup)
		if err != nil {
			return scaleUpResourcesDelta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
				"Failed to get target gpu for node group %v:", nodeGroup.Id())
		}
		resultScaleUpDelta[gpuType] = gpuCount
	}

	return resultScaleUpDelta, nil
}

// scaleUpLimitsCheckResult the check result of scale up limits
type scaleUpLimitsCheckResult struct {
	exceeded          bool
	exceededResources []string
}

// scaleUpLimitsNotExceeded returns the result that limits not exceed
func scaleUpLimitsNotExceeded() scaleUpLimitsCheckResult {
	return scaleUpLimitsCheckResult{false, []string{}}
}

// checkScaleUpDeltaWithinLimits check whether the scale up delta is within limits
func (limits *scaleUpResourcesLimits) checkScaleUpDeltaWithinLimits(
	delta scaleUpResourcesDelta,
) scaleUpLimitsCheckResult {
	exceededResources := sets.NewString()
	for resource, resourceDelta := range delta {
		resourceLeft, found := (*limits)[resource]
		if found {
			if (resourceDelta > 0) && (resourceLeft == scaleUpLimitUnknown || resourceDelta > resourceLeft) {
				exceededResources.Insert(resource)
			}
		}
	}
	if len(exceededResources) > 0 {
		return scaleUpLimitsCheckResult{true, exceededResources.List()}
	}

	return scaleUpLimitsNotExceeded()
}

// getNodeInfoCoresAndMemory returns the cores and memory of nodeInfo
func getNodeInfoCoresAndMemory(nodeInfo *schedulernodeinfo.NodeInfo) (int64, int64) {
	return getNodeCoresAndMemory(nodeInfo.Node())
}

// skippedReasons contains the skipped reason
type skippedReasons struct {
	message []string
}

// Reasons returns the specific reason
func (sr *skippedReasons) Reasons() []string {
	return sr.message
}

var (
	backoffReason         = &skippedReasons{[]string{"in backoff after failed scale-up"}}
	maxLimitReachedReason = &skippedReasons{[]string{"max node group size reached"}}
	notReadyReason        = &skippedReasons{[]string{"not ready for scale-up"}}
)

// maxResourceLimitReached returns the reason of max resource limit reached
func maxResourceLimitReached(resources []string) *skippedReasons {
	return &skippedReasons{[]string{fmt.Sprintf("max cluster %s limit reached", strings.Join(resources, ", "))}}
}

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occurred. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
// nolint
// NOCC:golint/fnsize(设计如此)
func ScaleUp(context *contextinternal.Context, processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, unschedulablePods []*apiv1.Pod, nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet, nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	ignoredTaints taintKeySet, existingNodeInfos map[string]*schedulernodeinfo.NodeInfo,
	bufferNotEnough bool, maxBulkScaleUpCount int,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 && !bufferNotEnough {
		klog.V(1).Info("No unschedulable pods and buffer is enough")
		return &status.ScaleUpStatus{Result: status.ScaleUpNotNeeded}, nil
	}

	now := time.Now()

	loggingQuota := glogx.PodsLoggingQuota()

	for _, pod := range unschedulablePods {
		glogx.V(1).UpTo(loggingQuota).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}
	glogx.V(1).Over(loggingQuota).Infof("%v other pods are also unschedulable", -loggingQuota.Left())

	nodesFromNotAutoscaledGroups, filterErr := filterOutNodesFromNotAutoscaledGroups(nodes, context.CloudProvider)
	if filterErr != nil {
		return &status.ScaleUpStatus{Result: status.ScaleUpError}, filterErr.AddPrefix(
			"failed to filter out nodes which are from not autoscaled groups: ")
	}

	nodeGroups := context.CloudProvider.NodeGroups()
	gpuLabel := context.CloudProvider.GPULabel()
	availableGPUTypes := context.CloudProvider.GetAvailableGPUTypes()

	resourceLimiter, errCP := context.CloudProvider.GetResourceLimiter()
	if errCP != nil {
		return &status.ScaleUpStatus{Result: status.ScaleUpError}, errors.ToAutoscalerError(
			errors.CloudProviderError,
			errCP)
	}

	scaleUpResourcesLeft, errLimits := computeScaleUpResourcesLeftLimits(context.CloudProvider, nodeGroups,
		nodeInfos, nodesFromNotAutoscaledGroups, resourceLimiter)
	if errLimits != nil {
		return &status.ScaleUpStatus{Result: status.ScaleUpError}, errLimits.AddPrefix("Could not compute total resources: ")
	}

	upcomingNodes, err := generateUpcomingNodes(clusterStateRegistry, nodeInfos)
	if err != nil {
		return &status.ScaleUpStatus{Result: status.ScaleUpError}, errors.NewAutoscalerError(
			errors.InternalError, err.Error())
	}
	klog.V(4).Infof("Upcoming %d nodes", len(upcomingNodes))

	if processors != nil && processors.NodeGroupListProcessor != nil {
		var errProc error
		nodeGroups, nodeInfos, errProc = processors.NodeGroupListProcessor.Process(context.AutoscalingContext,
			nodeGroups, nodeInfos, unschedulablePods)
		if errProc != nil {
			return &status.ScaleUpStatus{Result: status.ScaleUpError}, errors.ToAutoscalerError(errors.InternalError, errProc)
		}
	}

	// Build pod equivalence groups for efficient scheduling checks
	podEquivalenceGroups := buildPodEquivalenceGroups(unschedulablePods)

	// Separate Pod-Triggered and Buffer-Triggered Scale-Up Paths
	// Priority: If pending pods have valid expansion options, handle pods first;
	//           Otherwise fall back to buffer expansion if buffer shortage exists

	var expansionOptions map[string]expander.Option
	var skippedNodeGroups map[string]status.Reasons

	// Determine trigger type with priority
	hasPendingPods := len(unschedulablePods) > 0
	hasBufferShortage := bufferNotEnough

	if hasPendingPods {
		// Pod-Triggered Expansion: Single NodeGroup selection via Expander
		klog.V(1).Infof("[Pod-Triggered] Scale-up triggered by %d unschedulable pod(s)", len(unschedulablePods))

		// Compute expansion options for ALL NodeGroups
		expansionOptions, skippedNodeGroups = computeExpansionOptionForNodeGroups(context,
			podEquivalenceGroups, nodeGroups, nodeInfos, existingNodeInfos, upcomingNodes,
			clusterStateRegistry, resourceLimiter, scaleUpResourcesLeft, now, false) // bufferNotEnough=false for pod path

		if len(expansionOptions) == 0 {
			// No expansion options for pending pods
			if hasBufferShortage {
				// Fall back to buffer expansion since pods cannot be scheduled
				klog.V(1).Infof("[Priority] Pending pods have no valid expansion options, falling back to buffer-triggered expansion")
			} else {
				// No buffer shortage either, nothing to do
				klog.V(1).Info("[Pod-Triggered] No expansion options available for pending pods")
				return &status.ScaleUpStatus{
					Result:                  status.ScaleUpNoOptionsAvailable,
					PodsRemainUnschedulable: getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups),
					ConsideredNodeGroups:    nodeGroups,
				}, nil
			}
		} else {
			// Has valid expansion options for pods - use pod path
			if hasBufferShortage {
				klog.V(1).Infof("[Priority] Both pending pods (%d) and buffer shortage detected - pods have valid options, buffer deferred to next cycle",
					len(unschedulablePods))
			}

			// Select single best option using Expander strategy
			options := make([]expander.Option, 0, len(expansionOptions))
			for _, o := range expansionOptions {
				options = append(options, o)
			}
			bestOption := context.ExpanderStrategy.BestOption(options, nodeInfos)
			if bestOption != nil && bestOption.NodeCount > 0 {
				klog.V(1).Infof("[Pod-Triggered] Selected NodeGroup %s for expansion (nodes: %d)",
					bestOption.NodeGroup.Id(), bestOption.NodeCount)
				return optimizeBestOption(context, processors, clusterStateRegistry, daemonSets, nodes,
					upcomingNodes, nodeInfos, scaleUpResourcesLeft, resourceLimiter, ignoredTaints,
					expansionOptions, podEquivalenceGroups, gpuLabel, availableGPUTypes, skippedNodeGroups, nodeGroups,
					maxBulkScaleUpCount, now, bestOption)
			}

			// Expander didn't select a valid option (shouldn't happen, but handle gracefully)
			if hasBufferShortage {
				klog.V(1).Info("[Priority] Expander didn't select valid option for pods, falling back to buffer expansion")
			} else {
				klog.V(1).Info("[Pod-Triggered] No valid expansion option selected by Expander")
				return &status.ScaleUpStatus{
					Result:                  status.ScaleUpNoOptionsAvailable,
					PodsRemainUnschedulable: getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups),
					ConsideredNodeGroups:    nodeGroups,
				}, nil
			}
		}
	}

	// Buffer-Triggered Expansion Path
	// Executed when:
	//   1. No pending pods, OR
	//   2. Pending pods exist but have no valid expansion options (fallback)
	if hasBufferShortage {
		// Buffer-Triggered Expansion: Multi-Set expansion
		klog.V(1).Infof("[Buffer-Triggered] Scale-up triggered by buffer shortage in %d Set(s)",
			len(context.TriggeredNodeGroupSets))

		// Filter NodeGroups to only those in triggered Sets
		filteredNodeGroups := filterNodeGroupsByTriggeredSets(context, nodeGroups)
		if len(filteredNodeGroups) == 0 {
			klog.V(1).Info("[Buffer-Triggered] No NodeGroups in triggered Sets")
			result := &status.ScaleUpStatus{
				Result:               status.ScaleUpNotNeeded,
				ConsideredNodeGroups: nodeGroups,
			}
			if hasPendingPods {
				result.PodsRemainUnschedulable = getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups)
			}
			return result, nil
		}

		// Compute expansion options for filtered NodeGroups
		bufferExpansionOptions, bufferSkippedNodeGroups := computeExpansionOptionForNodeGroups(context,
			podEquivalenceGroups, filteredNodeGroups, nodeInfos, existingNodeInfos, upcomingNodes,
			clusterStateRegistry, resourceLimiter, scaleUpResourcesLeft, now, true) // bufferNotEnough=true for buffer path

		if len(bufferExpansionOptions) == 0 {
			klog.V(1).Info("[Buffer-Triggered] No expansion options available")
			result := &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: nodeGroups,
			}
			if hasPendingPods {
				// Merge skipped node groups from both pod and buffer paths
				mergedSkipped := make(map[string]status.Reasons)
				for ng, reason := range skippedNodeGroups {
					mergedSkipped[ng] = reason
				}
				for ng, reason := range bufferSkippedNodeGroups {
					if _, exists := mergedSkipped[ng]; !exists {
						mergedSkipped[ng] = reason
					}
				}
				result.PodsRemainUnschedulable = getRemainingPodsEquivalenceGroups(podEquivalenceGroups, mergedSkipped)
			}
			return result, nil
		}

		// Check if multi-Set expansion is enabled
		useMultiSet := shouldUseMultiSetExpansion(context)

		if !useMultiSet {
			// Fallback: single NodeGroup selection if multi-Set disabled
			klog.V(2).Info("[Buffer-Triggered] Multi-Set expansion disabled, using single NodeGroup selection")
			options := make([]expander.Option, 0, len(bufferExpansionOptions))
			for _, o := range bufferExpansionOptions {
				options = append(options, o)
			}
			bestOption := context.ExpanderStrategy.BestOption(options, nodeInfos)
			if bestOption != nil && bestOption.NodeCount > 0 {
				return optimizeBestOption(context, processors, clusterStateRegistry, daemonSets, nodes,
					upcomingNodes, nodeInfos, scaleUpResourcesLeft, resourceLimiter, ignoredTaints,
					bufferExpansionOptions, podEquivalenceGroups, gpuLabel, availableGPUTypes,
					bufferSkippedNodeGroups, nodeGroups, maxBulkScaleUpCount, now, bestOption)
			}
			result := &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: nodeGroups,
			}
			if hasPendingPods {
				result.PodsRemainUnschedulable = getRemainingPodsEquivalenceGroups(podEquivalenceGroups, bufferSkippedNodeGroups)
			}
			return result, nil
		}

		// Multi-Set expansion: select best option per Set
		klog.V(1).Infof("[Buffer-Triggered] Using multi-Set expansion with %d expansion options", len(bufferExpansionOptions))

		setPlans := selectBestOptionsPerSet(context, bufferExpansionOptions, nodeInfos)
		klog.V(1).Infof("[Buffer-Triggered] Selected best options for %d Set(s)", len(setPlans))

		if len(setPlans) == 0 {
			klog.V(1).Info("[Buffer-Triggered] No Set plans available after selection")
			result := &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: nodeGroups,
			}
			if hasPendingPods {
				result.PodsRemainUnschedulable = getRemainingPodsEquivalenceGroups(podEquivalenceGroups, bufferSkippedNodeGroups)
			}
			return result, nil
		}

		// Select which Sets to expand based on resource limits and priority
		maxConcurrentSets := getMaxConcurrentSets()
		selectedSets := selectSetsToExpand(setPlans, scaleUpResourcesLeft, maxConcurrentSets)
		klog.V(1).Infof("[Buffer-Triggered] Selected %d Set(s) for concurrent expansion (from %d candidates)",
			len(selectedSets), len(setPlans))

		if len(selectedSets) == 0 {
			klog.V(1).Info("[Buffer-Triggered] No Sets selected (resource limits exceeded)")
			result := &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: nodeGroups,
			}
			if hasPendingPods {
				result.PodsRemainUnschedulable = getRemainingPodsEquivalenceGroups(podEquivalenceGroups, bufferSkippedNodeGroups)
			}
			return result, nil
		}

		// Execute multi-Set expansion
		return executeMultiSetExpansion(context, processors, clusterStateRegistry, daemonSets, nodes,
			upcomingNodes, nodeInfos, scaleUpResourcesLeft, resourceLimiter, ignoredTaints,
			podEquivalenceGroups, gpuLabel, availableGPUTypes, bufferSkippedNodeGroups, nodeGroups,
			maxBulkScaleUpCount, now, selectedSets)

	}

	// No trigger - should not happen as we checked at entry
	klog.V(1).Info("No scale-up trigger detected")
	result := &status.ScaleUpStatus{Result: status.ScaleUpNotNeeded}
	if hasPendingPods {
		result.PodsRemainUnschedulable = getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups)
		result.ConsideredNodeGroups = nodeGroups
	}
	return result, nil
}

// generateUpcomingNodes returns the nodeInfos of upcoming nodes
func generateUpcomingNodes(clusterStateRegistry *clusterstate.ClusterStateRegistry,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
) ([]*schedulernodeinfo.NodeInfo, error) {
	upcomingNodes := make([]*schedulernodeinfo.NodeInfo, 0)
	for nodeGroup, numberOfNodes := range clusterStateRegistry.GetUpcomingNodes() {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			return upcomingNodes, fmt.Errorf("failed to find template node for node group %s", nodeGroup)
		}
		for i := 0; i < numberOfNodes; i++ {
			upcomingNodes = append(upcomingNodes, nodeTemplate)
		}
	}
	return upcomingNodes, nil
}

// checkNodeGroupScalable check whether the nodegroup is scalable
func checkNodeGroupScalable(context *contextinternal.Context,
	nodeGroup cloudprovider.NodeGroup, clusterStateRegistry *clusterstate.ClusterStateRegistry,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo, resourceLimiter *cloudprovider.ResourceLimiter,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	skippedNodeGroups map[string]status.Reasons, now time.Time,
) (map[string]status.Reasons, error) {
	var checkErr error
	// Autoprovisioned node groups without nodes are created later so skip check for them.
	if nodeGroup.Exist() && !clusterStateRegistry.IsNodeGroupSafeToScaleUp(nodeGroup, now) {
		// Hack that depends on internals of IsNodeGroupSafeToScaleUp.
		if !clusterStateRegistry.IsNodeGroupHealthy(nodeGroup.Id()) {
			checkErr = fmt.Errorf("Node group %s is not ready for scaleup - unhealthy", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
		} else {
			checkErr = fmt.Errorf("Node group %s is not ready for scaleup - backoff", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = backoffReason
		}
		return skippedNodeGroups, checkErr
	}

	currentTargetSize, err := nodeGroup.TargetSize()
	if err != nil {
		checkErr = fmt.Errorf("Failed to get node group size: %v", err)
		skippedNodeGroups[nodeGroup.Id()] = notReadyReason
		return skippedNodeGroups, checkErr
	}
	if currentTargetSize >= nodeGroup.MaxSize() {
		checkErr = fmt.Errorf("Skipping node group %s - max size reached", nodeGroup.Id())
		skippedNodeGroups[nodeGroup.Id()] = maxLimitReachedReason
		return skippedNodeGroups, checkErr
	}

	nodeInfo, found := nodeInfos[nodeGroup.Id()]
	if !found {
		checkErr = fmt.Errorf("No node info for: %s", nodeGroup.Id())
		skippedNodeGroups[nodeGroup.Id()] = notReadyReason
		return skippedNodeGroups, checkErr
	}

	scaleUpResourcesDeltaNow, err := computeScaleUpResourcesDelta(context.CloudProvider, nodeInfo,
		nodeGroup, resourceLimiter)
	if err != nil {
		checkErr = fmt.Errorf("Skipping node group %s; error getting node group resources: %v", nodeGroup.Id(), err)
		skippedNodeGroups[nodeGroup.Id()] = notReadyReason
		return skippedNodeGroups, checkErr
	}

	checkResult := scaleUpResourcesLeft.checkScaleUpDeltaWithinLimits(scaleUpResourcesDeltaNow)
	if checkResult.exceeded {
		checkErr = fmt.Errorf("Skipping node group %s; maximal limit exceeded for %v", nodeGroup.Id(),
			checkResult.exceededResources)
		skippedNodeGroups[nodeGroup.Id()] = maxResourceLimitReached(checkResult.exceededResources)
		for _, resource := range checkResult.exceededResources {
			switch resource {
			case cloudprovider.ResourceNameCores:
				metrics.RegisterSkippedScaleUpCPU()
			case cloudprovider.ResourceNameMemory:
				metrics.RegisterSkippedScaleUpMemory()
			default:
				continue
			}
		}
		return skippedNodeGroups, checkErr
	}
	return skippedNodeGroups, checkErr
}

// optimizeBestOption generates the best scale up options using podEquivalenceGroups
// NOCC:golint/fnsize(设计如此)
// nolint
func optimizeBestOption(context *contextinternal.Context, processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, daemonSets []*appsv1.DaemonSet,
	nodes []*apiv1.Node, upcomingNodes []*schedulernodeinfo.NodeInfo,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	scaleUpResourcesLeft scaleUpResourcesLimits, resourceLimiter *cloudprovider.ResourceLimiter,
	ignoredTaints taintKeySet, expansionOptions map[string]expander.Option,
	podEquivalenceGroups []*podEquivalenceGroup, gpuLabel string,
	availableGPUTypes map[string]struct{}, skippedNodeGroups map[string]status.Reasons,
	nodeGroups []cloudprovider.NodeGroup, maxBulkScaleUpCount int,
	now time.Time, bestOption *expander.Option,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	klog.V(1).Infof("Best option to resize: %s", bestOption.NodeGroup.Id())
	if len(bestOption.Debug) > 0 {
		klog.V(1).Info(bestOption.Debug)
	}
	klog.V(1).Infof("Estimated %d nodes needed in %s", bestOption.NodeCount, bestOption.NodeGroup.Id())

	newNodes := bestOption.NodeCount

	if context.MaxNodesTotal > 0 && len(nodes)+newNodes+len(upcomingNodes) > context.MaxNodesTotal {
		klog.V(1).Infof("Capping size to max cluster total size (%d)", context.MaxNodesTotal)
		newNodes = context.MaxNodesTotal - len(nodes) - len(upcomingNodes)
		if newNodes < 1 {
			return &status.ScaleUpStatus{Result: status.ScaleUpError}, errors.NewAutoscalerError(
				errors.TransientError,
				"max node total count already reached")
		}
	}

	createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)
	if !bestOption.NodeGroup.Exist() {
		oldID := bestOption.NodeGroup.Id()
		createNodeGroupResult, asErr := processors.NodeGroupManager.CreateNodeGroup(context.AutoscalingContext,
			bestOption.NodeGroup)
		if asErr != nil {
			return &status.ScaleUpStatus{Result: status.ScaleUpError}, asErr
		}
		createNodeGroupResults = append(createNodeGroupResults, createNodeGroupResult)
		bestOption.NodeGroup = createNodeGroupResult.MainCreatedNodeGroup

		nodeInfos = renewNodeInfos(context, nodeInfos, ignoredTaints, daemonSets, bestOption,
			createNodeGroupResult, oldID)

		// Update ClusterStateRegistry so similar nodegroups rebalancing works.
		clusterStateRegistry.Recalculate()
	}
	nodeInfo, found := nodeInfos[bestOption.NodeGroup.Id()]
	if !found {
		klog.Errorf("No node info for: %s", bestOption.NodeGroup.Id())
		return &status.ScaleUpStatus{Result: status.ScaleUpError, CreateNodeGroupResults: createNodeGroupResults},
			errors.NewAutoscalerError(errors.CloudProviderError, "No node info for best expansion option!")
	}

	// apply upper limits for CPU and memory
	newNodes, err := applyScaleUpResourcesLimits(context.CloudProvider, newNodes, scaleUpResourcesLeft, nodeInfo,
		bestOption.NodeGroup, resourceLimiter)
	if err != nil {
		return &status.ScaleUpStatus{Result: status.ScaleUpError, CreateNodeGroupResults: createNodeGroupResults}, err
	}
	// apply scale up limits for node
	if maxBulkScaleUpCount > 0 && newNodes > maxBulkScaleUpCount {
		klog.Infof("newNodes(%d) is larger than maxBulkScaleUpCount(%d), set to maxBulkScaleUpCount",
			newNodes, maxBulkScaleUpCount)
		newNodes = maxBulkScaleUpCount
	}

	targetNodeGroups := []cloudprovider.NodeGroup{bestOption.NodeGroup}
	if context.BalanceSimilarNodeGroups {
		similarNodeGroups, typedErr := processors.NodeGroupSetProcessor.FindSimilarNodeGroups(context.AutoscalingContext,
			bestOption.NodeGroup, nodeInfos)
		if typedErr != nil {
			return &status.ScaleUpStatus{Result: status.ScaleUpError, CreateNodeGroupResults: createNodeGroupResults},
				typedErr.AddPrefix("Failed to find matching node groups: ")
		}
		// Filter by pods and safety
		similarNodeGroups = filterNodeGroupsByPodsFromOptions(similarNodeGroups, bestOption.Pods, expansionOptions)
		for _, ng := range similarNodeGroups {
			if clusterStateRegistry.IsNodeGroupSafeToScaleUp(ng, now) {
				targetNodeGroups = append(targetNodeGroups, ng)
			} else {
				klog.V(2).Infof("Ignoring node group %s when balancing: group is not ready for scaleup", ng.Id())
			}
		}
		if len(targetNodeGroups) > 1 {
			var buffer bytes.Buffer
			for i, ng := range targetNodeGroups {
				if i > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(ng.Id())
			}
			klog.V(1).Infof("Splitting scale-up between %v similar node groups: {%v}", len(targetNodeGroups), buffer.String())
		}
	}
	scaleUpInfos, typedErr := processors.NodeGroupSetProcessor.BalanceScaleUpBetweenGroups(
		context.AutoscalingContext, targetNodeGroups, newNodes)
	if typedErr != nil {
		return &status.ScaleUpStatus{Result: status.ScaleUpError, CreateNodeGroupResults: createNodeGroupResults}, typedErr
	}
	klog.V(1).Infof("Final scale-up plan: %v", scaleUpInfos)
	for _, info := range scaleUpInfos {
		typedErr := executeScaleUp(context.AutoscalingContext, clusterStateRegistry, info,
			gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, nodeInfo.Node(), nil), now)
		if typedErr != nil {
			return &status.ScaleUpStatus{Result: status.ScaleUpError, CreateNodeGroupResults: createNodeGroupResults}, typedErr
		}
	}

	clusterStateRegistry.Recalculate()
	return &status.ScaleUpStatus{
		Result:                  status.ScaleUpSuccessful,
		ScaleUpInfos:            scaleUpInfos,
		PodsRemainUnschedulable: getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups,
		CreateNodeGroupResults:  createNodeGroupResults,
		PodsTriggeredScaleUp:    bestOption.Pods,
		PodsAwaitEvaluation:     getPodsAwaitingEvaluationEquivalenceGroups(podEquivalenceGroups, bestOption.NodeGroup.Id()),
	}, nil
}

// filterNodeGroupsByPodsFromOptions filters node groups by pods using expansion options map
func filterNodeGroupsByPodsFromOptions(
	groups []cloudprovider.NodeGroup,
	podsRequiredToFit []*apiv1.Pod,
	expansionOptions map[string]expander.Option,
) []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0)

groupsloop:
	for _, group := range groups {
		option, found := expansionOptions[group.Id()]
		if !found {
			klog.V(1).Infof("No info about pods passing predicates found for group %v,"+
				" skipping it from scale-up consideration", group.Id())
			continue
		}
		fittingPods := option.Pods
		podSet := make(map[*apiv1.Pod]bool, len(fittingPods))
		for _, pod := range fittingPods {
			podSet[pod] = true
		}
		for _, pod := range podsRequiredToFit {
			if _, found := podSet[pod]; !found {
				klog.V(1).Infof("%.500s can't be scheduled on %s, removing from scale-up consideration",
					pod.Name, group.Id())
				continue groupsloop
			}
		}
		result = append(result, group)
	}
	return result
}

// renewNodeInfos renew nodeInfos
func renewNodeInfos(context *contextinternal.Context,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo, ignoredTaints taintKeySet,
	daemonSets []*appsv1.DaemonSet, bestOption *expander.Option,
	createNodeGroupResult nodegroups.CreateNodeGroupResult, oldID string,
) map[string]*schedulernodeinfo.NodeInfo {
	// If possible replace candidate node-info with node info based on crated node group. The latter
	// one should be more in line with nodes which will be created by node group.
	mainCreatedNodeInfo, asErr := getNodeInfoFromTemplate(createNodeGroupResult.MainCreatedNodeGroup, daemonSets,
		context.PredicateChecker, ignoredTaints)
	if asErr == nil {
		nodeInfos[createNodeGroupResult.MainCreatedNodeGroup.Id()] = mainCreatedNodeInfo
	} else {
		klog.Warningf("Cannot build node info for newly created main node group %v;"+
			" balancing similar node groups may not work; err=%v", createNodeGroupResult.MainCreatedNodeGroup.Id(), asErr)
		// Use node info based on expansion candidate but upadte Id which likely changed when node group was created.
		nodeInfos[bestOption.NodeGroup.Id()] = nodeInfos[oldID]
	}

	if oldID != createNodeGroupResult.MainCreatedNodeGroup.Id() {
		delete(nodeInfos, oldID)
	}

	for _, nodeGroup := range createNodeGroupResult.ExtraCreatedNodeGroups {
		nodeInfo, asErr := getNodeInfoFromTemplate(nodeGroup, daemonSets, context.PredicateChecker, ignoredTaints)

		if asErr != nil {
			klog.Warningf("Cannot build node info for newly created extra node group %v;"+
				" balancing similar node groups will not work; err=%v", nodeGroup.Id(), asErr)
			continue
		}
		nodeInfos[nodeGroup.Id()] = nodeInfo
	}
	return nodeInfos
}

// executeScaleUp execute scale up action with option
func executeScaleUp(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry,
	info nodegroupset.ScaleUpInfo, gpuType string, now time.Time,
) errors.AutoscalerError {
	klog.V(0).Infof("Scale-up: setting group %s size to %d", info.Group.Id(), info.NewSize)
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: setting group %s size to %d", info.Group.Id(), info.NewSize)
	increase := info.NewSize - info.CurrentSize
	if err := info.Group.IncreaseSize(increase); err != nil {
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "FailedToScaleUpGroup", "Scale-up failed for group %s: %v",
			info.Group.Id(), err)
		clusterStateRegistry.RegisterFailedScaleUp(info.Group, metrics.APIError, now)
		return errors.NewAutoscalerError(errors.CloudProviderError,
			"failed to increase node group size: %v", err)
	}
	clusterStateRegistry.RegisterOrUpdateScaleUp(
		info.Group,
		increase,
		time.Now())
	metrics.RegisterScaleUp(increase, gpuType)
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: group %s size set to %d", info.Group.Id(), info.NewSize)
	return nil
}

// applyScaleUpResourcesLimits applys scale up resources limits on scale up option
func applyScaleUpResourcesLimits(
	cp cloudprovider.CloudProvider,
	newNodes int,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	nodeInfo *schedulernodeinfo.NodeInfo,
	nodeGroup cloudprovider.NodeGroup,
	resourceLimiter *cloudprovider.ResourceLimiter,
) (int, errors.AutoscalerError) {
	delta, err := computeScaleUpResourcesDelta(cp, nodeInfo, nodeGroup, resourceLimiter)
	if err != nil {
		return 0, err
	}

	for resource, resourceDelta := range delta {
		limit, limitFound := scaleUpResourcesLeft[resource]
		if !limitFound {
			continue
		}
		if limit == scaleUpLimitUnknown {
			// should never happen - checked before
			return 0, errors.NewAutoscalerError(
				errors.InternalError,
				fmt.Sprintf("limit unknown for resource %s", resource))
		}
		if int64(newNodes)*resourceDelta <= limit {
			// no capping required
			continue
		}

		newNodes = int(limit / resourceDelta)
		klog.V(1).Infof("Capping scale-up size due to limit for resource %s", resource)
		if newNodes < 1 {
			// should never happen - checked before
			return 0, errors.NewAutoscalerError(
				errors.InternalError,
				fmt.Sprintf("cannot create any node; max limit for resource %s reached", resource))
		}
	}
	return newNodes, nil
}

// filterNodeGroupsByTriggeredSets filters NodeGroups to only those in triggered Sets
// Used by buffer-triggered expansion to target specific Sets
func filterNodeGroupsByTriggeredSets(
	context *contextinternal.Context,
	allNodeGroups []cloudprovider.NodeGroup,
) []cloudprovider.NodeGroup {
	if context.NodeGroupSetManager == nil || len(context.TriggeredNodeGroupSets) == 0 {
		return allNodeGroups
	}

	nodeGroupSetManager, ok := context.NodeGroupSetManager.(*NodeGroupSetManager)
	if !ok {
		klog.Warning("[Buffer-Triggered] NodeGroupSetManager has invalid type, using all NodeGroups")
		return allNodeGroups
	}

	// Build set of triggered Set names
	triggeredSetMap := make(map[string]bool)
	for _, setName := range context.TriggeredNodeGroupSets {
		triggeredSetMap[setName] = true
	}

	// Filter NodeGroups
	filteredNodeGroups := make([]cloudprovider.NodeGroup, 0)
	for _, ng := range allNodeGroups {
		setName := nodeGroupSetManager.GetNodeGroupSet(ng.Id())
		if triggeredSetMap[setName] {
			filteredNodeGroups = append(filteredNodeGroups, ng)
			klog.V(4).Infof("[Buffer-Triggered] Including NodeGroup %s from triggered Set %s",
				ng.Id(), setName)
		}
	}

	klog.V(3).Infof("[Buffer-Triggered] Filtered %d/%d NodeGroups from %d triggered Set(s)",
		len(filteredNodeGroups), len(allNodeGroups), len(context.TriggeredNodeGroupSets))

	return filteredNodeGroups
}

// createEstimatorForNodeGroup creates a dynamic Estimator for the given NodeGroup
// with Set-specific ratios and Set-scoped resource calculation
// nolint
func createEstimatorForNodeGroup(context *contextinternal.Context,
	nodeGroup cloudprovider.NodeGroup,
	existingNodeInfos map[string]*schedulernodeinfo.NodeInfo,
) estimator.Estimator {
	// Check if NodeGroupSetManager is configured
	if context.NodeGroupSetManager == nil {
		// Fall back to global estimator (backward compatibility)
		return context.ExtendedEstimatorBuilder(
			context.AutoscalingContext.PredicateChecker,
			existingNodeInfos)
	}

	manager, ok := context.NodeGroupSetManager.(*NodeGroupSetManager)
	if !ok {
		// Invalid type, fall back to global estimator
		klog.Warning("NodeGroupSetManager has invalid type, using global estimator")
		return context.ExtendedEstimatorBuilder(
			context.AutoscalingContext.PredicateChecker,
			existingNodeInfos)
	}

	// Get the Set this NodeGroup belongs to
	setName := manager.GetNodeGroupSet(nodeGroup.Id())
	if setName == "" {
		// NodeGroup not in any Set, use global estimator
		klog.V(4).Infof("NodeGroup %s not in any Set, using global estimator", nodeGroup.Id())
		return context.ExtendedEstimatorBuilder(
			context.AutoscalingContext.PredicateChecker,
			existingNodeInfos)
	}

	// GlobalSetName 表示使用全局集群级别配置
	if setName == GlobalSetName {
		klog.V(2).Infof("NodeGroup %s uses global cluster-level configuration (not in any Set, no DefaultSet)",
			nodeGroup.Id())

		// 使用全局 buffer ratio (来自命令行参数)
		cpuRatio := context.GlobalBufferedCPURatio
		memRatio := context.GlobalBufferedMemRatio
		resourceRatio := context.GlobalBufferedResourceRatio

		klog.V(4).Infof(
			"Creating cluster-scoped estimator for NodeGroup %s with global ratios: cpu=%.2f, mem=%.2f, resource=%.2f",
			nodeGroup.Id(), cpuRatio, memRatio, resourceRatio)

		// 创建 Cluster-scoped Estimator (不限制 Set 范围，使用全集群节点)
		return estimatorinternal.NewClusterResourceEstimatorWithSetScope(
			context.AutoscalingContext.PredicateChecker,
			existingNodeInfos,
			cpuRatio,
			memRatio,
			resourceRatio,
			nil, // setScope=nil 表示 Cluster-scoped，使用全集群节点
		)
	}

	// Get Set configuration
	setConfig, err := manager.GetSetConfig(setName)
	if err != nil {
		klog.Warningf("Failed to get config for Set %s: %v, using global estimator", setName, err)
		return context.ExtendedEstimatorBuilder(
			context.AutoscalingContext.PredicateChecker,
			existingNodeInfos)
	}

	// Get all NodeGroups in this Set for scope filtering
	setNodeGroups := manager.GetNodeGroupsInSet(setName)
	setNodeGroupIDs := make([]string, 0, len(setNodeGroups))
	for _, ng := range setNodeGroups {
		setNodeGroupIDs = append(setNodeGroupIDs, ng.Id())
	}

	// Use cached Node→NodeGroup mapping from Context if available
	var nodeToGroup map[string]string
	if len(context.NodeToGroupCache) > 0 {
		nodeToGroup = context.NodeToGroupCache
		klog.V(5).Infof("[NodeToGroup] Using cached mapping with %d nodes from Context", len(nodeToGroup))
	} else {
		// Fallback: Build mapping if cache not available
		klog.V(4).Info("[NodeToGroup] Cache not available, building mapping for estimator")
		nodeToGroup = make(map[string]string)
		for _, nodeInfo := range existingNodeInfos {
			node := nodeInfo.Node()
			if node != nil {
				if node.Annotations != nil &&
					node.Annotations[NodeUpcomingAnnotation] == valueTrue {
					nodeToGroup[node.Name] = node.Annotations[bcsNodeGroupAnnoKey]
				} else {
					tmpNodeGroup, getErr := context.CloudProvider.NodeGroupForNode(node)
					if getErr == nil && tmpNodeGroup != nil {
						nodeToGroup[node.Name] = tmpNodeGroup.Id()
					}
				}
			}
		}
	}

	// Create SetScope with accurate node mapping
	setScope := &estimatorinternal.EstimatorSetScope{
		Mode:          estimatorinternal.SetScoped,
		SetNodeGroups: setNodeGroupIDs,
		SetName:       setName,
		NodeToGroup:   nodeToGroup,
	}

	// Use Set-specific ratios
	cpuRatio := setConfig.BufferedRatios.CPURatio
	memRatio := setConfig.BufferedRatios.MemRatio
	resourceRatio := setConfig.BufferedRatios.ResourceRatio

	klog.V(4).Infof(
		"Creating Set-scoped estimator for NodeGroup %s in Set %s with ratios: cpu=%.2f, mem=%.2f, resource=%.2f",
		nodeGroup.Id(), setName, cpuRatio, memRatio, resourceRatio)

	// Create Set-scoped estimator with Set-specific ratios
	return estimatorinternal.NewClusterResourceEstimatorWithSetScope(
		context.AutoscalingContext.PredicateChecker,
		existingNodeInfos,
		cpuRatio,
		memRatio,
		resourceRatio,
		setScope,
	)
}

// shouldUseMultiSetExpansion determines if multi-Set concurrent expansion should be used
func shouldUseMultiSetExpansion(context *contextinternal.Context) bool {
	// Enable multi-Set if NodeGroupSetManager is configured
	if context.NodeGroupSetManager == nil {
		return false
	}

	manager, ok := context.NodeGroupSetManager.(*NodeGroupSetManager)
	if !ok {
		return false
	}

	// Check if we have multiple Sets configured
	allSets := manager.GetAllSets()
	if len(allSets) <= 1 {
		// Only one Set, no benefit from multi-Set logic
		return false
	}

	klog.V(3).Infof("Multi-Set expansion enabled with %d Set(s)", len(allSets))
	return true
}

// getMaxConcurrentSets returns the maximum number of Sets that can be scaled concurrently
func getMaxConcurrentSets() int {
	// TODO: Make this configurable via command-line flag
	// For now, return 0 which means unlimited
	return 0
}

// computeExpansionOption computes the expansion option for a single NodeGroup
func computeExpansionOption(context *contextinternal.Context, podEquivalenceGroups []*podEquivalenceGroup,
	nodeGroup cloudprovider.NodeGroup, nodeInfo *schedulernodeinfo.NodeInfo,
	existingNodeInfos map[string]*schedulernodeinfo.NodeInfo, upcomingNodes []*schedulernodeinfo.NodeInfo,
	bufferNotEnough bool,
) expander.Option {
	option := expander.Option{
		NodeGroup: nodeGroup,
		Pods:      make([]*apiv1.Pod, 0),
	}

	// Check each pod equivalence group for schedulability
	for _, eg := range podEquivalenceGroups {
		samplePod := eg.pods[0]
		err := context.PredicateChecker.CheckPredicates(samplePod, nil, nodeInfo)
		if err == nil {
			// add pods to option
			option.Pods = append(option.Pods, eg.pods...)
			// mark pod group as (theoretically) schedulable
			eg.schedulable = true
		} else {
			klog.V(2).Infof("Pod %s can't be scheduled on %s, predicate checking error: %v",
				samplePod.Name, nodeGroup.Id(), err.VerboseError())
			if podCount := len(eg.pods); podCount > 1 {
				klog.V(2).Infof("%d other pods similar to %s can't be scheduled on %s",
					podCount-1, samplePod.Name, nodeGroup.Id())
			}
			eg.schedulingErrors[nodeGroup.Id()] = err
		}
	}

	if len(option.Pods) > 0 || bufferNotEnough {
		// Create dynamic Estimator with Set-specific ratios and scope
		estimator := createEstimatorForNodeGroup(context, nodeGroup, existingNodeInfos)

		if bufferNotEnough && len(option.Pods) == 0 {
			// Buffer-triggered expansion: calculate based on buffer requirements
			option.NodeCount = estimator.Estimate([]*apiv1.Pod{}, nodeInfo, upcomingNodes)
			klog.V(4).Infof("[Buffer] NodeGroup %s buffer expansion needs %d nodes",
				nodeGroup.Id(), option.NodeCount)
		} else if bufferNotEnough {
			// Mixed mode: pods + buffer
			option.NodeCount = estimator.Estimate(option.Pods, nodeInfo, upcomingNodes)
			klog.V(4).Infof("[Mixed] NodeGroup %s expansion needs %d nodes for %d pods + buffer",
				nodeGroup.Id(), option.NodeCount, len(option.Pods))
		} else {
			// Pod-only expansion: use zero-ratio estimator
			podOnlyEstimator := estimatorinternal.NewClusterResourceEstimatorWithSetScope(
				context.AutoscalingContext.PredicateChecker,
				existingNodeInfos,
				0.0, // cpuRatio = 0 means no buffer
				0.0, // memRatio = 0
				0.0, // resourceRatio = 0
				nil, // No Set scope needed for pod-only calculation
			)
			option.NodeCount = podOnlyEstimator.Estimate(option.Pods, nodeInfo, upcomingNodes)
			klog.V(4).Infof("[Pod] NodeGroup %s pod expansion needs %d nodes for %d pods",
				nodeGroup.Id(), option.NodeCount, len(option.Pods))
		}
	}

	return option
}

// computeExpansionOptionForNodeGroups computes expansion options for multiple NodeGroups
func computeExpansionOptionForNodeGroups(context *contextinternal.Context, podEquivalenceGroups []*podEquivalenceGroup,
	nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	existingNodeInfos map[string]*schedulernodeinfo.NodeInfo, upcomingNodes []*schedulernodeinfo.NodeInfo,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	resourceLimiter *cloudprovider.ResourceLimiter, scaleUpResourcesLeft scaleUpResourcesLimits, now time.Time,
	bufferNotEnough bool,
) (map[string]expander.Option, map[string]status.Reasons) {
	expansionOptions := make(map[string]expander.Option, 0)
	skippedNodeGroups := map[string]status.Reasons{}
	var err error
	for _, nodeGroup := range nodeGroups {
		skippedNodeGroups, err = checkNodeGroupScalable(context, nodeGroup, clusterStateRegistry,
			nodeInfos, resourceLimiter, scaleUpResourcesLeft, skippedNodeGroups, now)
		if err != nil {
			klog.Info(err.Error())
			continue
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			klog.Errorf("No node info for: %s", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}

		option := computeExpansionOption(context, podEquivalenceGroups,
			nodeGroup, nodeInfo, existingNodeInfos, upcomingNodes, bufferNotEnough)

		if len(option.Pods) > 0 || bufferNotEnough {
			if option.NodeCount > 0 {
				expansionOptions[nodeGroup.Id()] = option
			} else {
				klog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
			}
		} else {
			klog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
		}
	}
	return expansionOptions, skippedNodeGroups
}

// getRemainingPodsEquivalenceGroups returns remaining pods from equivalence groups
func getRemainingPodsEquivalenceGroups(egs []*podEquivalenceGroup,
	skipped map[string]status.Reasons,
) []status.NoScaleUpInfo {
	remaining := []status.NoScaleUpInfo{}
	for _, eg := range egs {
		if eg.schedulable {
			continue
		}
		for _, pod := range eg.pods {
			noScaleUpInfo := status.NoScaleUpInfo{
				Pod:                pod,
				RejectedNodeGroups: eg.schedulingErrors,
				SkippedNodeGroups:  skipped,
			}
			remaining = append(remaining, noScaleUpInfo)
		}
	}
	return remaining
}

// getPodsAwaitingEvaluationEquivalenceGroups returns pods awaiting evaluation from equivalence groups
// NOCC:tosa/fn_length(设计如此)
func getPodsAwaitingEvaluationEquivalenceGroups(egs []*podEquivalenceGroup, bestOption string) []*apiv1.Pod {
	awaitsEvaluation := []*apiv1.Pod{}
	for _, eg := range egs {
		if eg.schedulable {
			if _, found := eg.schedulingErrors[bestOption]; found {
				// Schedulable, but not yet.
				awaitsEvaluation = append(awaitsEvaluation, eg.pods...)
			}
		}
	}
	return awaitsEvaluation
}

// executeMultiSetExpansion executes scale-up for multiple Sets concurrently
func executeMultiSetExpansion(
	context *contextinternal.Context,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	daemonSets []*appsv1.DaemonSet,
	nodes []*apiv1.Node,
	upcomingNodes []*schedulernodeinfo.NodeInfo,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	resourceLimiter *cloudprovider.ResourceLimiter,
	ignoredTaints taintKeySet,
	podEquivalenceGroups []*podEquivalenceGroup,
	gpuLabel string,
	availableGPUTypes map[string]struct{},
	skippedNodeGroups map[string]status.Reasons,
	nodeGroups []cloudprovider.NodeGroup,
	maxBulkScaleUpCount int,
	now time.Time,
	selectedSets []*SetExpansionPlan,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	klog.V(1).Infof("Executing multi-Set expansion for %d Set(s)", len(selectedSets))

	// Aggregate results
	allScaleUpInfos := make([]nodegroupset.ScaleUpInfo, 0)
	allCreateNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)
	var firstError errors.AutoscalerError

	// Execute scale-up for each selected Set
	for i, setPlan := range selectedSets {
		klog.V(1).Infof("Executing Set %d/%d: %s (NodeGroup: %s, nodes: %d)",
			i+1, len(selectedSets), setPlan.SetName, setPlan.BestOption.NodeGroup.Id(), setPlan.BestOption.NodeCount)

		// Execute scale-up for this Set's best option
		scaleUpStatus, err := executeSingleSetExpansion(
			context, processors, clusterStateRegistry, daemonSets, nodes,
			upcomingNodes, nodeInfos, scaleUpResourcesLeft, resourceLimiter,
			ignoredTaints, podEquivalenceGroups, gpuLabel, availableGPUTypes,
			skippedNodeGroups, nodeGroups, maxBulkScaleUpCount,
			now, setPlan)
		if err != nil {
			klog.Errorf("Failed to execute Set %s: %v", setPlan.SetName, err)
			if firstError == nil {
				firstError = err
			}
			// Continue with other Sets even if one fails
			continue
		}

		// Collect results
		if scaleUpStatus != nil {
			if scaleUpStatus.ScaleUpInfos != nil {
				allScaleUpInfos = append(allScaleUpInfos, scaleUpStatus.ScaleUpInfos...)
			}
			if scaleUpStatus.CreateNodeGroupResults != nil {
				allCreateNodeGroupResults = append(allCreateNodeGroupResults, scaleUpStatus.CreateNodeGroupResults...)
			}
		}
	}

	// Return aggregated results
	if len(allScaleUpInfos) == 0 {
		if firstError != nil {
			return &status.ScaleUpStatus{
				Result:                 status.ScaleUpError,
				CreateNodeGroupResults: allCreateNodeGroupResults,
			}, firstError
		}
		return &status.ScaleUpStatus{
			Result:                  status.ScaleUpNoOptionsAvailable,
			PodsRemainUnschedulable: getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups),
			ConsideredNodeGroups:    nodeGroups,
		}, nil
	}

	klog.V(1).Infof("[MultiSet] Multi-Set expansion completed: %d Set(s) scaled successfully", len(allScaleUpInfos))

	return &status.ScaleUpStatus{
		Result:                  status.ScaleUpSuccessful,
		ScaleUpInfos:            allScaleUpInfos,
		PodsRemainUnschedulable: getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups,
		CreateNodeGroupResults:  allCreateNodeGroupResults,
	}, nil
}

// executeSingleSetExpansion executes scale-up for a single Set
// nolint
func executeSingleSetExpansion(
	context *contextinternal.Context,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	daemonSets []*appsv1.DaemonSet,
	nodes []*apiv1.Node,
	upcomingNodes []*schedulernodeinfo.NodeInfo,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	resourceLimiter *cloudprovider.ResourceLimiter,
	ignoredTaints taintKeySet,
	podEquivalenceGroups []*podEquivalenceGroup,
	gpuLabel string,
	availableGPUTypes map[string]struct{},
	skippedNodeGroups map[string]status.Reasons,
	nodeGroups []cloudprovider.NodeGroup,
	maxBulkScaleUpCount int,
	now time.Time,
	setPlan *SetExpansionPlan,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	bestOption := setPlan.BestOption
	nodeInfo := setPlan.NodeInfo

	klog.V(2).Infof("Set %s: scaling NodeGroup %s by %d nodes",
		setPlan.SetName, bestOption.NodeGroup.Id(), bestOption.NodeCount)

	newNodes := bestOption.NodeCount

	// Apply max nodes total limit
	if context.MaxNodesTotal > 0 && len(nodes)+newNodes+len(upcomingNodes) > context.MaxNodesTotal {
		klog.V(2).Infof("Set %s: capping size to max cluster total size (%d)",
			setPlan.SetName, context.MaxNodesTotal)
		newNodes = context.MaxNodesTotal - len(nodes) - len(upcomingNodes)
		if newNodes < 1 {
			klog.Warningf("Set %s: max node total count already reached", setPlan.SetName)
			return nil, errors.NewAutoscalerError(
				errors.TransientError,
				"max node total count already reached")
		}
	}

	// Handle NodeGroup creation if needed
	createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)
	if !bestOption.NodeGroup.Exist() {
		klog.V(2).Infof("Set %s: creating new NodeGroup %s",
			setPlan.SetName, bestOption.NodeGroup.Id())

		createNodeGroupResult, asErr := processors.NodeGroupManager.CreateNodeGroup(
			context.AutoscalingContext, bestOption.NodeGroup)
		if asErr != nil {
			return &status.ScaleUpStatus{
				Result:                 status.ScaleUpError,
				CreateNodeGroupResults: createNodeGroupResults,
			}, asErr
		}
		createNodeGroupResults = append(createNodeGroupResults, createNodeGroupResult)
		bestOption.NodeGroup = createNodeGroupResult.MainCreatedNodeGroup

		// Update nodeInfo
		mainCreatedNodeInfo, err := getNodeInfoFromTemplate(
			createNodeGroupResult.MainCreatedNodeGroup,
			daemonSets, context.PredicateChecker, ignoredTaints)
		if err == nil {
			nodeInfo = mainCreatedNodeInfo
		}
	}

	// Apply resource limits
	newNodes, err := applyScaleUpResourcesLimits(context.CloudProvider, newNodes,
		scaleUpResourcesLeft, nodeInfo, bestOption.NodeGroup, resourceLimiter)
	if err != nil {
		return &status.ScaleUpStatus{
			Result:                 status.ScaleUpError,
			CreateNodeGroupResults: createNodeGroupResults,
		}, err
	}

	// Apply bulk scale-up limit
	if maxBulkScaleUpCount > 0 && newNodes > maxBulkScaleUpCount {
		klog.V(2).Infof("Set %s: capping newNodes(%d) to maxBulkScaleUpCount(%d)",
			setPlan.SetName, newNodes, maxBulkScaleUpCount)
		newNodes = maxBulkScaleUpCount
	}

	// Prepare target NodeGroups (with similar node group balancing if enabled)
	targetNodeGroups := []cloudprovider.NodeGroup{bestOption.NodeGroup}
	if context.BalanceSimilarNodeGroups {
		similarNodeGroups, typedErr := processors.NodeGroupSetProcessor.FindSimilarNodeGroups(
			context.AutoscalingContext, bestOption.NodeGroup, nodeInfos)
		if typedErr != nil {
			return &status.ScaleUpStatus{
				Result:                 status.ScaleUpError,
				CreateNodeGroupResults: createNodeGroupResults,
			}, typedErr.AddPrefix("Failed to find matching node groups: ")
		}

		// Filter by safety
		for _, ng := range similarNodeGroups {
			if clusterStateRegistry.IsNodeGroupSafeToScaleUp(ng, now) {
				targetNodeGroups = append(targetNodeGroups, ng)
			}
		}

		if len(targetNodeGroups) > 1 {
			klog.V(2).Infof("Set %s: balancing scale-up between %d similar node groups",
				setPlan.SetName, len(targetNodeGroups))
		}
	}

	// Balance scale-up between target NodeGroups
	scaleUpInfos, typedErr := processors.NodeGroupSetProcessor.BalanceScaleUpBetweenGroups(
		context.AutoscalingContext, targetNodeGroups, newNodes)
	if typedErr != nil {
		return &status.ScaleUpStatus{
			Result:                 status.ScaleUpError,
			CreateNodeGroupResults: createNodeGroupResults,
		}, typedErr
	}

	klog.V(1).Infof("Set %s: final scale-up plan: %v", setPlan.SetName, scaleUpInfos)

	// Execute scale-up
	for _, info := range scaleUpInfos {
		typedErr := executeScaleUp(context.AutoscalingContext, clusterStateRegistry, info,
			gpu.GetGpuTypeForMetrics(gpuLabel, availableGPUTypes, nodeInfo.Node(), nil), now)
		if typedErr != nil {
			return &status.ScaleUpStatus{
				Result:                 status.ScaleUpError,
				CreateNodeGroupResults: createNodeGroupResults,
			}, typedErr
		}
	}

	clusterStateRegistry.Recalculate()

	return &status.ScaleUpStatus{
		Result:                  status.ScaleUpSuccessful,
		ScaleUpInfos:            scaleUpInfos,
		PodsRemainUnschedulable: getRemainingPodsEquivalenceGroups(podEquivalenceGroups, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups,
		CreateNodeGroupResults:  createNodeGroupResults,
		PodsTriggeredScaleUp:    bestOption.Pods,
	}, nil
}
