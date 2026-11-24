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
	"math"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroups"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

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
	context *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulerframework.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node,
	resourceLimiter *cloudprovider.ResourceLimiter,
) (scaleUpResourcesLimits, errors.AutoscalerError) {
	totalCores, totalMem, errCoresMem := calculateScaleUpCoresMemoryTotal(nodeGroups, nodeInfos,
		nodesFromNotAutoscaledGroups)

	var totalResources map[string]int64
	var totalResourcesErr error
	if cloudprovider.ContainsCustomResources(resourceLimiter.GetResources()) {
		totalResources, totalResourcesErr = calculateScaleUpCustomResourcesTotal(context,
			processors, nodeGroups, nodeInfos, nodesFromNotAutoscaledGroups)
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

			case cloudprovider.IsCustomResource(resource):
				if totalResourcesErr != nil {
					resultScaleUpLimits[resource] = scaleUpLimitUnknown
				} else {
					resultScaleUpLimits[resource] = computeBelowMax(totalResources[resource], max)
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
	nodeInfos map[string]*schedulerframework.NodeInfo,
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
			return 0, 0, errors.NewAutoscalerError(errors.CloudProviderError,
				"No node info for: %s", nodeGroup.Id())
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
// NOCC:tosa/fn_length(设计如此)
func calculateScaleUpCustomResourcesTotal(
	context *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulerframework.NodeInfo,
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
			return nil, errors.NewAutoscalerError(errors.CloudProviderError,
				"No node info for: %s", nodeGroup.Id())
		}
		if currentSize > 0 {
			resourceTargets, err := processors.CustomResourcesProcessor.GetNodeResourceTargets(
				context, nodeInfo.Node(), nodeGroup)
			if err != nil {
				return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
					"Failed to get target gpu for node group %v:", nodeGroup.Id())
			}
			for _, resourceTarget := range resourceTargets {
				if resourceTarget.ResourceType == "" || resourceTarget.ResourceCount == 0 {
					continue
				}
				result[resourceTarget.ResourceType] += resourceTarget.ResourceCount * int64(currentSize)
			}
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		resourceTargets, err := processors.CustomResourcesProcessor.GetNodeResourceTargets(
			context, node, nil)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
				"Failed to get target gpu for node gpus count for node %v:", node.Name)
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

func computeBelowMax(total int64, max int64) int64 {
	if total < max {
		return max - total
	}
	return 0
}

// computeScaleUpResourcesDelta compute scale up resources delta
func computeScaleUpResourcesDelta(context *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	nodeInfo *schedulerframework.NodeInfo, nodeGroup cloudprovider.NodeGroup,
	resourceLimiter *cloudprovider.ResourceLimiter,
) (scaleUpResourcesDelta, errors.AutoscalerError) {
	resultScaleUpDelta := make(scaleUpResourcesDelta)

	nodeCPU, nodeMemory := getNodeInfoCoresAndMemory(nodeInfo)
	resultScaleUpDelta[cloudprovider.ResourceNameCores] = nodeCPU
	resultScaleUpDelta[cloudprovider.ResourceNameMemory] = nodeMemory

	if cloudprovider.ContainsCustomResources(resourceLimiter.GetResources()) {
		resourceTargets, err := processors.CustomResourcesProcessor.GetNodeResourceTargets(
			context, nodeInfo.Node(), nodeGroup)
		if err != nil {
			return scaleUpResourcesDelta{}, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
				"Failed to get target custom resources for node group %v:", nodeGroup.Id())
		}
		for _, resourceTarget := range resourceTargets {
			resultScaleUpDelta[resourceTarget.ResourceType] = resourceTarget.ResourceCount
		}
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
func getNodeInfoCoresAndMemory(nodeInfo *schedulerframework.NodeInfo) (int64, int64) {
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
	return &skippedReasons{[]string{fmt.Sprintf("max cluster %s limit reached",
		strings.Join(resources, ", "))}}
}

// computeExpansionOption computes scale up options
// nolint
func computeExpansionOption(context *contextinternal.Context, podEquivalenceGroups []*podEquivalenceGroup,
	nodeGroup cloudprovider.NodeGroup, nodeInfo *schedulerframework.NodeInfo,
	bufferNotEnough bool,
) (expander.Option, error) {
	option := expander.Option{
		NodeGroup: nodeGroup,
		Pods:      make([]*apiv1.Pod, 0),
	}

	if err := context.AutoscalingContext.ClusterSnapshot.Fork(); err != nil {
		klog.Errorf("Error while calling ClusterSnapshot.Fork; %v", err)
		return expander.Option{}, err
	}

	// add test node to snapshot
	var pods []*apiv1.Pod
	for _, podInfo := range nodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}
	if err := context.AutoscalingContext.ClusterSnapshot.AddNodeWithPods(nodeInfo.Node(), pods); err != nil {
		klog.Errorf("Error while adding test Node; %v", err)
		if revertErr := context.AutoscalingContext.ClusterSnapshot.Revert(); revertErr != nil {
			klog.Fatalf("Error while calling ClusterSnapshot.Revert; %v", revertErr)
		}
		// DOTO: Or should I just skip the node group? specifically if Revert fails it is fatal error.
		//       Maybe we should not return error from Revert as we cannot handle it in any way on the caller side?
		return expander.Option{}, err
	}

	for _, eg := range podEquivalenceGroups {
		samplePod := eg.pods[0]
		if err := context.AutoscalingContext.PredicateChecker.CheckPredicates(context.AutoscalingContext.ClusterSnapshot,
			samplePod, nodeInfo.Node().Name); err == nil {
			// add pods to option
			option.Pods = append(option.Pods, eg.pods...)
			// mark pod group as (theoretically) schedulable
			eg.schedulable = true
		} else {
			klog.V(2).Infof("Pod %s can't be scheduled on %s, predicate checking error: %v",
				samplePod.Name, nodeGroup.Id(), err.VerboseMessage())
			if podCount := len(eg.pods); podCount > 1 {
				klog.V(2).Infof("%d other pods similar to %s can't be scheduled on %s",
					podCount-1, samplePod.Name, nodeGroup.Id())
			}
			eg.schedulingErrors[nodeGroup.Id()] = err
		}
	}

	if err := context.AutoscalingContext.ClusterSnapshot.Revert(); err != nil {
		klog.Fatalf("Error while calling ClusterSnapshot.Revert; %v", err)
		return expander.Option{}, err
	}

	if len(option.Pods) > 0 || bufferNotEnough {
		// Create dynamic Estimator with Set-specific ratios and scope
		estimator := createEstimatorForNodeGroup(context, nodeGroup)

		if bufferNotEnough {
			// Buffer-triggered expansion: calculate based on buffer requirements
			// Pass empty pods list to get pure buffer calculation
			// The estimator will call estimateAccordingToLoad() which calculates buffer needs
			option.NodeCount = estimator.Estimate([]*apiv1.Pod{}, nodeInfo)
			klog.V(4).Infof("[Buffer] NodeGroup %s buffer expansion needs %d nodes",
				nodeGroup.Id(), option.NodeCount)
		} else {
			// Pod-triggered expansion: calculate based on pods only
			// We need to override the estimator to NOT calculate buffer
			// Use a zero-ratio estimator for pod-only calculation
			podOnlyEstimator := estimatorinternal.NewClusterResourceEstimatorWithSetScope(
				context.AutoscalingContext.PredicateChecker,
				context.AutoscalingContext.ClusterSnapshot,
				0.0, // cpuRatio = 0 means no buffer
				0.0, // memRatio = 0
				0.0, // resourceRatio = 0
				nil, // No Set scope needed for pod-only calculation
			)
			option.NodeCount = podOnlyEstimator.Estimate(option.Pods, nodeInfo)
			klog.V(4).Infof("[Pod] NodeGroup %s pod expansion needs %d nodes for %d pods",
				nodeGroup.Id(), option.NodeCount, len(option.Pods))
		}
	}

	return option, nil
}

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occurred. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
// NOCC:golint/fnsize(设计如此)
// nolint
func ScaleUp(context *contextinternal.Context, processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, unschedulablePods []*apiv1.Pod, nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet, nodeInfos map[string]*schedulerframework.NodeInfo,
	ignoredTaints taints.TaintKeySet, bufferNotEnough bool,
	maxBulkScaleUpCount int,
) (*status.ScaleUpStatus, errors.AutoscalerError) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(unschedulablePods) == 0 && !bufferNotEnough {
		klog.V(1).Info("No unschedulable pods and buffer is enough")
		return &status.ScaleUpStatus{Result: status.ScaleUpNotNeeded}, nil
	}

	now := time.Now()

	loggingQuota := klogx.PodsLoggingQuota()

	for _, pod := range unschedulablePods {
		klogx.V(1).UpTo(loggingQuota).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
	}
	klogx.V(1).Over(loggingQuota).Infof("%v other pods are also unschedulable", -loggingQuota.Left())

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

	scaleUpResourcesLeft, errLimits := computeScaleUpResourcesLeftLimits(context.AutoscalingContext, processors,
		nodeGroups, nodeInfos, nodesFromNotAutoscaledGroups, resourceLimiter)
	if errLimits != nil {
		return &status.ScaleUpStatus{Result: status.ScaleUpError},
			errLimits.AddPrefix("Could not compute total resources: ")
	}

	upcomingNodes := make([]*schedulerframework.NodeInfo, 0)
	for nodeGroup, numberOfNodes := range clusterStateRegistry.GetUpcomingNodes() {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			return &status.ScaleUpStatus{Result: status.ScaleUpError}, errors.NewAutoscalerError(
				errors.InternalError,
				"failed to find template node for node group %s",
				nodeGroup)
		}
		for i := 0; i < numberOfNodes; i++ {
			upcomingNodes = append(upcomingNodes, nodeTemplate)
		}
	}
	klog.V(4).Infof("Upcoming %d nodes", len(upcomingNodes))

	// expansionOptions := make(map[string]expander.Option, 0)

	if processors != nil && processors.NodeGroupListProcessor != nil {
		var errProc error
		nodeGroups, nodeInfos, errProc = processors.NodeGroupListProcessor.Process(context.AutoscalingContext,
			nodeGroups, nodeInfos, unschedulablePods)
		if errProc != nil {
			return &status.ScaleUpStatus{Result: status.ScaleUpError},
				errors.ToAutoscalerError(errors.InternalError, errProc)
		}
	}

	podEquivalenceGroups := buildPodEquivalenceGroups(unschedulablePods)

	// Separate Pod-Triggered and Buffer-Triggered Scale-Up Paths
	// Priority: If pending pods have valid expansion options, handle pods first;
	//           Otherwise fall back to buffer expansion if buffer shortage exists

	var expansionOptions map[string]expander.Option
	var skippedNodeGroups map[string]status.Reasons
	var scaleUpStatus *status.ScaleUpStatus
	var errCompute errors.AutoscalerError

	// Determine trigger type with priority
	hasPendingPods := len(unschedulablePods) > 0
	hasBufferShortage := bufferNotEnough

	if hasPendingPods {
		// Pod-Triggered Expansion: Single NodeGroup selection via Expander
		klog.V(1).Infof("[Pod-Triggered] Scale-up triggered by %d unschedulable pod(s)", len(unschedulablePods))

		// Compute expansion options for ALL NodeGroups
		expansionOptions, skippedNodeGroups, scaleUpStatus, errCompute = computeExpansionOptionForNodeGroups(context,
			podEquivalenceGroups, nodeGroups, nodeInfos, clusterStateRegistry, processors, resourceLimiter,
			scaleUpResourcesLeft, now, false) // bufferNotEnough=false for pod path
		if errCompute != nil {
			return scaleUpStatus, errCompute
		}

		if len(expansionOptions) == 0 {
			// No expansion options for pending pods
			if !hasBufferShortage {
				klog.V(1).Info("[Pod-Triggered] No expansion options available for pending pods")
				return &status.ScaleUpStatus{
					Result:                  status.ScaleUpNoOptionsAvailable,
					PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
					ConsideredNodeGroups:    nodeGroups,
				}, nil
			}
			// Fall back to buffer expansion since pods cannot be scheduled
			klog.V(1).Infof("[Priority] Pending pods have no valid expansion options, falling back to buffer-triggered expansion")
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
					upcomingNodes, nodeInfos, scaleUpResourcesLeft, resourceLimiter, ignoredTaints, expansionOptions,
					podEquivalenceGroups, gpuLabel, availableGPUTypes, skippedNodeGroups, nodeGroups,
					maxBulkScaleUpCount, false, now, bestOption)
			}

			// Expander didn't select a valid option (shouldn't happen, but handle gracefully)
			if !hasBufferShortage {
				klog.V(1).Info("[Pod-Triggered] No valid expansion option selected by Expander")
				return &status.ScaleUpStatus{
					Result:                  status.ScaleUpNoOptionsAvailable,
					PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
					ConsideredNodeGroups:    nodeGroups,
				}, nil
			}
			klog.V(1).Info("[Priority] Expander didn't select valid option for pods, falling back to buffer expansion")
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
				result.PodsRemainUnschedulable = getRemainingPods(podEquivalenceGroups, skippedNodeGroups)
			}
			return result, nil
		}

		// Compute expansion options for filtered NodeGroups
		bufferExpansionOptions, bufferSkippedNodeGroups, scaleUpStatus, errCompute := computeExpansionOptionForNodeGroups(context,
			podEquivalenceGroups, filteredNodeGroups, nodeInfos, clusterStateRegistry, processors, resourceLimiter,
			scaleUpResourcesLeft, now, true) // bufferNotEnough=true for buffer path
		if errCompute != nil {
			return scaleUpStatus, errCompute
		}

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
				result.PodsRemainUnschedulable = getRemainingPods(podEquivalenceGroups, mergedSkipped)
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
					upcomingNodes, nodeInfos, scaleUpResourcesLeft, resourceLimiter, ignoredTaints, bufferExpansionOptions,
					podEquivalenceGroups, gpuLabel, availableGPUTypes, bufferSkippedNodeGroups, nodeGroups,
					maxBulkScaleUpCount, true, now, bestOption)
			}
			result := &status.ScaleUpStatus{
				Result:               status.ScaleUpNoOptionsAvailable,
				ConsideredNodeGroups: nodeGroups,
			}
			if hasPendingPods {
				result.PodsRemainUnschedulable = getRemainingPods(podEquivalenceGroups, bufferSkippedNodeGroups)
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
				result.PodsRemainUnschedulable = getRemainingPods(podEquivalenceGroups, bufferSkippedNodeGroups)
			}
			return result, nil
		}

		// Select which Sets to expand based on resource limits and priority
		maxConcurrentSets := getMaxConcurrentSets(context)
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
				result.PodsRemainUnschedulable = getRemainingPods(podEquivalenceGroups, bufferSkippedNodeGroups)
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
		result.PodsRemainUnschedulable = getRemainingPods(podEquivalenceGroups, skippedNodeGroups)
		result.ConsideredNodeGroups = nodeGroups
	}
	return result, nil
}

// optimizeBestOption generates the best scale up options
// NOCC:golint/fnsize(设计如此)
// nolint funlen
func optimizeBestOption(context *contextinternal.Context, processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, daemonSets []*appsv1.DaemonSet,
	nodes []*apiv1.Node, upcomingNodes []*schedulerframework.NodeInfo,
	nodeInfos map[string]*schedulerframework.NodeInfo,
	scaleUpResourcesLeft scaleUpResourcesLimits, resourceLimiter *cloudprovider.ResourceLimiter,
	ignoredTaints taints.TaintKeySet, expansionOptions map[string]expander.Option,
	podEquivalenceGroups []*podEquivalenceGroup, gpuLabel string,
	availableGPUTypes map[string]struct{}, skippedNodeGroups map[string]status.Reasons,
	nodeGroups []cloudprovider.NodeGroup, maxBulkScaleUpCount int,
	bufferNotEnough bool, now time.Time, bestOption *expander.Option,
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
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "MaxNodesTotalReached",
			"Max total nodes in cluster reached: %v", context.MaxNodesTotal)
		if newNodes < 1 {
			return &status.ScaleUpStatus{Result: status.ScaleUpError}, errors.NewAutoscalerError(
				errors.TransientError,
				"max node total count already reached")
		}
	}

	createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)
	if !bestOption.NodeGroup.Exist() {
		expansionOptions, createNodeGroupResults, _, _ = computeExpansionOptionForCreatedNodeGroups(
			context, podEquivalenceGroups, nodeInfos, clusterStateRegistry, processors,
			bestOption, expansionOptions, daemonSets, ignoredTaints, bufferNotEnough)
	}

	nodeInfo, found := nodeInfos[bestOption.NodeGroup.Id()]
	if !found {
		// This should never happen, as we already should have retrieved
		// nodeInfo for any considered nodegroup.
		klog.Errorf("No node info for: %s", bestOption.NodeGroup.Id())
		return &status.ScaleUpStatus{Result: status.ScaleUpError, CreateNodeGroupResults: createNodeGroupResults},
			errors.NewAutoscalerError(errors.CloudProviderError,
				"No node info for best expansion option!")
	}

	// apply upper limits for CPU and memory
	newNodes, err := applyScaleUpResourcesLimits(context.AutoscalingContext, processors, newNodes,
		scaleUpResourcesLeft, nodeInfo, bestOption.NodeGroup, resourceLimiter)
	if err != nil {
		return &status.ScaleUpStatus{
			Result:                 status.ScaleUpError,
			CreateNodeGroupResults: createNodeGroupResults,
		}, err
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
			return &status.ScaleUpStatus{
					Result:                 status.ScaleUpError,
					CreateNodeGroupResults: createNodeGroupResults,
				},
				typedErr.AddPrefix("Failed to find matching node groups: ")
		}
		similarNodeGroups = filterNodeGroupsByPods(similarNodeGroups, bestOption.Pods, expansionOptions)
		for _, ng := range similarNodeGroups {
			if clusterStateRegistry.IsNodeGroupSafeToScaleUp(ng, now) {
				targetNodeGroups = append(targetNodeGroups, ng)
			} else {
				// This should never happen, as we will filter out the node group earlier on
				// because of missing entry in podsPassingPredicates, but double checking doesn't
				// really cost us anything
				klog.V(2).Infof("Ignoring node group %s when balancing: group is not ready for scaleup", ng.Id())
			}
		}
		if len(targetNodeGroups) > 1 {
			names := []string{}
			for _, ng := range targetNodeGroups {
				names = append(names, ng.Id())
			}
			klog.V(1).Infof("Splitting scale-up between %v similar node groups: {%v}",
				len(targetNodeGroups), strings.Join(names, ", "))
		}
	}
	scaleUpInfos, typedErr := processors.NodeGroupSetProcessor.BalanceScaleUpBetweenGroups(
		context.AutoscalingContext, targetNodeGroups, newNodes)
	if typedErr != nil {
		return &status.ScaleUpStatus{
			Result:                 status.ScaleUpError,
			CreateNodeGroupResults: createNodeGroupResults,
		}, typedErr
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
		PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups,
		CreateNodeGroupResults:  createNodeGroupResults,
		PodsTriggeredScaleUp:    bestOption.Pods,
		PodsAwaitEvaluation:     getPodsAwaitingEvaluation(podEquivalenceGroups, bestOption.NodeGroup.Id()),
	}, nil
}

// computeExpansionOptionForNodeGroups computes scale up option for each nodegroup
func computeExpansionOptionForNodeGroups(context *contextinternal.Context, podEquivalenceGroups []*podEquivalenceGroup,
	nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*schedulerframework.NodeInfo,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, processors *ca_processors.AutoscalingProcessors,
	resourceLimiter *cloudprovider.ResourceLimiter, scaleUpResourcesLeft scaleUpResourcesLimits, now time.Time,
	bufferNotEnough bool) (map[string]expander.Option,
	map[string]status.Reasons, *status.ScaleUpStatus, errors.AutoscalerError,
) {
	expansionOptions := make(map[string]expander.Option, 0)
	skippedNodeGroups := map[string]status.Reasons{}
	for _, nodeGroup := range nodeGroups {
		// Autoprovisioned node groups without nodes are created later so skip check for them.
		if nodeGroup.Exist() && !clusterStateRegistry.IsNodeGroupSafeToScaleUp(nodeGroup, now) {
			// Hack that depends on internals of IsNodeGroupSafeToScaleUp.
			if !clusterStateRegistry.IsNodeGroupHealthy(nodeGroup.Id()) {
				klog.Warningf("Node group %s is not ready for scaleup - unhealthy", nodeGroup.Id())
				skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			} else {
				klog.Warningf("Node group %s is not ready for scaleup - backoff", nodeGroup.Id())
				skippedNodeGroups[nodeGroup.Id()] = backoffReason
			}
			continue
		}

		currentTargetSize, err := nodeGroup.TargetSize()
		if err != nil {
			klog.Errorf("Failed to get node group size: %v", err)
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}
		if currentTargetSize >= nodeGroup.MaxSize() {
			klog.V(4).Infof("Skipping node group %s - max size reached", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = maxLimitReachedReason
			continue
		}

		nodeInfo, found := nodeInfos[nodeGroup.Id()]
		if !found {
			klog.Errorf("No node info for: %s", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}

		scaleUpResourceDelta, err := computeScaleUpResourcesDelta(context.AutoscalingContext, processors,
			nodeInfo, nodeGroup, resourceLimiter)
		if err != nil {
			klog.Errorf("Skipping node group %s; error getting node group resources: %v", nodeGroup.Id(), err)
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}
		checkResult := scaleUpResourcesLeft.checkScaleUpDeltaWithinLimits(scaleUpResourceDelta)
		if checkResult.exceeded {
			klog.V(4).Infof("Skipping node group %s; maximal limit exceeded for %v", nodeGroup.Id(),
				checkResult.exceededResources)
			skippedNodeGroups[nodeGroup.Id()] = maxResourceLimitReached(checkResult.exceededResources)
			continue
		}

		option, err := computeExpansionOption(context, podEquivalenceGroups,
			nodeGroup, nodeInfo, bufferNotEnough)
		if err != nil {
			return expansionOptions, skippedNodeGroups, &status.ScaleUpStatus{Result: status.ScaleUpError},
				errors.ToAutoscalerError(errors.InternalError, err)
		}

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
	return expansionOptions, skippedNodeGroups, &status.ScaleUpStatus{Result: status.ScaleUpNotTried}, nil
}

// computeExpansionOptionForCreatedNodeGroups computes expansion option for created nodegroups
// NOCC:tosa/fn_length(设计如此)
func computeExpansionOptionForCreatedNodeGroups(context *contextinternal.Context,
	podEquivalenceGroups []*podEquivalenceGroup, nodeInfos map[string]*schedulerframework.NodeInfo,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, processors *ca_processors.AutoscalingProcessors,
	bestOption *expander.Option, expansionOptions map[string]expander.Option,
	daemonSets []*appsv1.DaemonSet, ignoredTaints taints.TaintKeySet,
	bufferNotEnough bool) (map[string]expander.Option,
	[]nodegroups.CreateNodeGroupResult, *status.ScaleUpStatus, errors.AutoscalerError,
) {
	createNodeGroupResults := make([]nodegroups.CreateNodeGroupResult, 0)

	oldID := bestOption.NodeGroup.Id()
	createNodeGroupResult, asErr := processors.NodeGroupManager.CreateNodeGroup(context.AutoscalingContext,
		bestOption.NodeGroup)
	if asErr != nil {
		return expansionOptions, createNodeGroupResults, &status.ScaleUpStatus{Result: status.ScaleUpError}, asErr
	}
	createNodeGroupResults = append(createNodeGroupResults, createNodeGroupResult)
	bestOption.NodeGroup = createNodeGroupResult.MainCreatedNodeGroup

	// If possible replace candidate node-info with node info based on crated node group. The latter
	// one should be more in line with nodes which will be created by node group.
	mainCreatedNodeInfo, err := utils.GetNodeInfoFromTemplate(createNodeGroupResult.MainCreatedNodeGroup,
		daemonSets, context.PredicateChecker, ignoredTaints)
	if err == nil {
		nodeInfos[createNodeGroupResult.MainCreatedNodeGroup.Id()] = mainCreatedNodeInfo
	} else {
		klog.Warningf("Cannot build node info for newly created main node group %v;"+
			" balancing similar node groups may not work; err=%v",
			createNodeGroupResult.MainCreatedNodeGroup.Id(), err)
		// Use node info based on expansion candidate but upadte Id which likely changed when node group was created.
		nodeInfos[bestOption.NodeGroup.Id()] = nodeInfos[oldID]
	}

	if oldID != createNodeGroupResult.MainCreatedNodeGroup.Id() {
		delete(nodeInfos, oldID)
	}

	for _, nodeGroup := range createNodeGroupResult.ExtraCreatedNodeGroups {
		nodeInfo, err := utils.GetNodeInfoFromTemplate(nodeGroup, daemonSets,
			context.PredicateChecker, ignoredTaints)
		if err != nil {
			klog.Warningf("Cannot build node info for newly created extra node group %v;"+
				" balancing similar node groups will not work; err=%v", nodeGroup.Id(), asErr)
			continue
		}
		nodeInfos[nodeGroup.Id()] = nodeInfo

		option, err2 := computeExpansionOption(context, podEquivalenceGroups,
			nodeGroup, nodeInfo, bufferNotEnough)
		if err2 != nil {
			return expansionOptions, createNodeGroupResults,
				&status.ScaleUpStatus{Result: status.ScaleUpError},
				errors.ToAutoscalerError(errors.InternalError, err2)
		}

		if len(option.Pods) > 0 && option.NodeCount > 0 {
			expansionOptions[nodeGroup.Id()] = option
		}
	}

	// Update ClusterStateRegistry so similar nodegroups rebalancing works.
	// DOTO(lukaszos) when pursuing scalability update this call with one which takes list of changed node groups so
	//                we do not do extra API calls. (the call at the bottom of ScaleUp() could be also changed then)
	clusterStateRegistry.Recalculate()
	return expansionOptions, createNodeGroupResults, &status.ScaleUpStatus{}, nil
}

// getRemainingPods return remaining pods
func getRemainingPods(egs []*podEquivalenceGroup,
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

// getPodsAwaitingEvaluation return pods awaiting evaluation
func getPodsAwaitingEvaluation(egs []*podEquivalenceGroup, bestOption string) []*apiv1.Pod {
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

// filterNodeGroupsByPods return suitable nodegroups for pods
func filterNodeGroupsByPods(
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
				klog.V(1).Infof("Group %v, can't fit pod %v/%v, removing from scale-up consideration",
					group.Id(), pod.Namespace, pod.Name)
				continue groupsloop
			}
		}
		result = append(result, group)
	}
	return result
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
		aerr := errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix(
			"failed to increase node group size: ")
		clusterStateRegistry.RegisterFailedScaleUp(info.Group,
			metrics.FailedScaleUpReason(string(aerr.Type())), now)
		return aerr
	}
	clusterStateRegistry.RegisterOrUpdateScaleUp(
		info.Group,
		increase,
		time.Now())
	metrics.RegisterScaleUp(increase, gpuType)
	context.LogRecorder.Eventf(apiv1.EventTypeNormal, "ScaledUpGroup",
		"Scale-up: group %s size set to %d instead of %d (max: %d)",
		info.Group.Id(), info.NewSize, info.CurrentSize, info.MaxSize)
	return nil
}

// applyScaleUpResourcesLimits applys scale up resources limits on scale up option
func applyScaleUpResourcesLimits(
	context *context.AutoscalingContext,
	processors *ca_processors.AutoscalingProcessors,
	newNodes int,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	nodeInfo *schedulerframework.NodeInfo,
	nodeGroup cloudprovider.NodeGroup,
	resourceLimiter *cloudprovider.ResourceLimiter,
) (int, errors.AutoscalerError) {
	delta, err := computeScaleUpResourcesDelta(context, processors, nodeInfo,
		nodeGroup, resourceLimiter)
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
) estimator.Estimator {
	// Check if NodeGroupSetManager is configured
	if context.NodeGroupSetManager == nil {
		// Fall back to global estimator (backward compatibility)
		return context.ExtendedEstimatorBuilder(
			context.AutoscalingContext.PredicateChecker,
			context.AutoscalingContext.ClusterSnapshot)
	}

	manager, ok := context.NodeGroupSetManager.(*NodeGroupSetManager)
	if !ok {
		// Invalid type, fall back to global estimator
		klog.Warning("NodeGroupSetManager has invalid type, using global estimator")
		return context.ExtendedEstimatorBuilder(
			context.AutoscalingContext.PredicateChecker,
			context.AutoscalingContext.ClusterSnapshot)
	}

	// Get the Set this NodeGroup belongs to
	setName := manager.GetNodeGroupSet(nodeGroup.Id())
	if setName == "" {
		// NodeGroup not in any Set, use global estimator
		klog.V(4).Infof("NodeGroup %s not in any Set, using global estimator", nodeGroup.Id())
		return context.ExtendedEstimatorBuilder(
			context.AutoscalingContext.PredicateChecker,
			context.AutoscalingContext.ClusterSnapshot)
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
			context.AutoscalingContext.ClusterSnapshot,
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
			context.AutoscalingContext.ClusterSnapshot)
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
		// This happens when NodeGroupSet is disabled or snapshot failed
		klog.V(4).Info("[NodeToGroup] Cache not available, building mapping for estimator")
		nodeToGroup = make(map[string]string)
		allNodes, err := context.AutoscalingContext.ClusterSnapshot.NodeInfos().List()
		if err == nil {
			for _, nodeInfo := range allNodes {
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
		} else {
			klog.Warningf("Failed to list nodes for building NodeToGroup mapping: %v", err)
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
		context.AutoscalingContext.ClusterSnapshot,
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
func getMaxConcurrentSets(context *contextinternal.Context) int {
	// TODO: Make this configurable via command-line flag
	// For now, return 0 which means unlimited
	return 0
}

// executeMultiSetExpansion executes scale-up for multiple Sets concurrently
func executeMultiSetExpansion(
	context *contextinternal.Context,
	processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	daemonSets []*appsv1.DaemonSet,
	nodes []*apiv1.Node,
	upcomingNodes []*schedulerframework.NodeInfo,
	nodeInfos map[string]*schedulerframework.NodeInfo,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	resourceLimiter *cloudprovider.ResourceLimiter,
	ignoredTaints taints.TaintKeySet,
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
		// Re-use the existing optimizeBestOption logic
		status, err := executeSingleSetExpansion(
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
		if status != nil {
			if status.ScaleUpInfos != nil {
				allScaleUpInfos = append(allScaleUpInfos, status.ScaleUpInfos...)
			}
			if status.CreateNodeGroupResults != nil {
				allCreateNodeGroupResults = append(allCreateNodeGroupResults, status.CreateNodeGroupResults...)
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
			PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
			ConsideredNodeGroups:    nodeGroups,
		}, nil
	}

	klog.V(1).Infof("[Phase2] Multi-Set expansion completed: %d Set(s) scaled successfully", len(allScaleUpInfos))

	return &status.ScaleUpStatus{
		Result:                  status.ScaleUpSuccessful,
		ScaleUpInfos:            allScaleUpInfos,
		PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
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
	upcomingNodes []*schedulerframework.NodeInfo,
	nodeInfos map[string]*schedulerframework.NodeInfo,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	resourceLimiter *cloudprovider.ResourceLimiter,
	ignoredTaints taints.TaintKeySet,
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

	klog.V(2).Infof(" Set %s: scaling NodeGroup %s by %d nodes",
		setPlan.SetName, bestOption.NodeGroup.Id(), bestOption.NodeCount)

	newNodes := bestOption.NodeCount

	// Apply max nodes total limit
	if context.MaxNodesTotal > 0 && len(nodes)+newNodes+len(upcomingNodes) > context.MaxNodesTotal {
		klog.V(2).Infof("Set %s: capping size to max cluster total size (%d)",
			setPlan.SetName, context.MaxNodesTotal)
		newNodes = context.MaxNodesTotal - len(nodes) - len(upcomingNodes)
		if newNodes < 1 {
			klog.Warningf(" Set %s: max node total count already reached", setPlan.SetName)
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
		mainCreatedNodeInfo, err := utils.GetNodeInfoFromTemplate(
			createNodeGroupResult.MainCreatedNodeGroup,
			daemonSets, context.PredicateChecker, ignoredTaints)
		if err == nil {
			nodeInfo = mainCreatedNodeInfo
		}
	}

	// Apply resource limits
	newNodes, err := applyScaleUpResourcesLimits(context.AutoscalingContext, processors, newNodes,
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
		PodsRemainUnschedulable: getRemainingPods(podEquivalenceGroups, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups,
		CreateNodeGroupResults:  createNodeGroupResults,
		PodsTriggeredScaleUp:    bestOption.Pods,
	}, nil
}
