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
 *
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
)

type scaleUpResourcesLimits map[string]int64
type scaleUpResourcesDelta map[string]int64

// used as a value in scaleUpResourcesLimits if actual limit could not be obtained
// due to errors talking to cloud provider
const scaleUpLimitUnknown = math.MaxInt64

func computeScaleUpResourcesLeftLimits(
	cp cloudprovider.CloudProvider,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node,
	resourceLimiter *cloudprovider.ResourceLimiter) (scaleUpResourcesLimits, errors.AutoscalerError) {
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

func calculateScaleUpCoresMemoryTotal(
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node) (int64, int64, errors.AutoscalerError) {
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
			coresTotal = coresTotal + int64(currentSize)*nodeCPU
			memoryTotal = memoryTotal + int64(currentSize)*nodeMemory
		}
	}

	for _, node := range nodesFromNotAutoscaledGroups {
		cores, memory := getNodeCoresAndMemory(node)
		coresTotal += cores
		memoryTotal += memory
	}

	return coresTotal, memoryTotal, nil
}

func calculateScaleUpGpusTotal(
	GPULabel string,
	nodeGroups []cloudprovider.NodeGroup,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	nodesFromNotAutoscaledGroups []*apiv1.Node) (map[string]int64, errors.AutoscalerError) {

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
			gpuType, gpuCount, err := gpu.GetNodeTargetGpus(GPULabel, nodeInfo.Node(), nodeGroup)
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
		gpuType, gpuCount, err := gpu.GetNodeTargetGpus(GPULabel, node, nil)
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

func computeScaleUpResourcesDelta(cp cloudprovider.CloudProvider, nodeInfo *schedulernodeinfo.NodeInfo,
	nodeGroup cloudprovider.NodeGroup, resourceLimiter *cloudprovider.ResourceLimiter) (scaleUpResourcesDelta,
	errors.AutoscalerError) {
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

type scaleUpLimitsCheckResult struct {
	exceeded          bool
	exceededResources []string
}

func scaleUpLimitsNotExceeded() scaleUpLimitsCheckResult {
	return scaleUpLimitsCheckResult{false, []string{}}
}

func (limits *scaleUpResourcesLimits) checkScaleUpDeltaWithinLimits(
	delta scaleUpResourcesDelta) scaleUpLimitsCheckResult {
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

func getNodeInfoCoresAndMemory(nodeInfo *schedulernodeinfo.NodeInfo) (int64, int64) {
	return getNodeCoresAndMemory(nodeInfo.Node())
}

type skippedReasons struct {
	message []string
}

func (sr *skippedReasons) Reasons() []string {
	return sr.message
}

var (
	backoffReason         = &skippedReasons{[]string{"in backoff after failed scale-up"}}
	maxLimitReachedReason = &skippedReasons{[]string{"max node group size reached"}}
	notReadyReason        = &skippedReasons{[]string{"not ready for scale-up"}}
)

func maxResourceLimitReached(resources []string) *skippedReasons {
	return &skippedReasons{[]string{fmt.Sprintf("max cluster %s limit reached", strings.Join(resources, ", "))}}
}

// ScaleUp tries to scale the cluster up. Return true if it found a way to increase the size,
// false if it didn't and error if an error occurred. Assumes that all nodes in the cluster are
// ready and in sync with instance groups.
func ScaleUp(context *contextinternal.Context, processors *ca_processors.AutoscalingProcessors,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, originUnschedulablePods []*apiv1.Pod, nodes []*apiv1.Node,
	daemonSets []*appsv1.DaemonSet, nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	ignoredTaints taintKeySet, existingNodeInfos map[string]*schedulernodeinfo.NodeInfo,
	bufferNotEnough bool) (*status.ScaleUpStatus, errors.AutoscalerError) {
	// From now on we only care about unschedulable pods that were marked after the newest
	// node became available for the scheduler.
	if len(originUnschedulablePods) == 0 && !bufferNotEnough {
		klog.V(1).Info("No unschedulable pods")
		return &status.ScaleUpStatus{Result: status.ScaleUpNotNeeded}, nil
	}

	// 去除 eip 资源
	unschedulablePods := make([]*apiv1.Pod, 0)
	for i := range originUnschedulablePods {
		pod := originUnschedulablePods[i].DeepCopy()
		for j := range pod.Spec.Containers {
			delete(pod.Spec.Containers[j].Resources.Requests, "cloud.bkbcs.tencent.com/eip")
			delete(pod.Spec.Containers[j].Resources.Requests, "tke.cloud.tencent.com/eni-ip")
			delete(pod.Spec.Containers[j].Resources.Requests, "tke.cloud.tencent.com/direct-eni")
		}
		unschedulablePods = append(unschedulablePods, pod)
	}

	now := time.Now()

	loggingQuota := glogx.PodsLoggingQuota()

	podsRemainUnschedulable := make(map[*apiv1.Pod]map[string]status.Reasons)

	for _, pod := range unschedulablePods {
		glogx.V(1).UpTo(loggingQuota).Infof("Pod %s/%s is unschedulable", pod.Namespace, pod.Name)
		podsRemainUnschedulable[pod] = make(map[string]status.Reasons)
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

	upcomingNodes := make([]*schedulernodeinfo.NodeInfo, 0)
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

	expansionOptions := make([]expander.Option, 0)

	if processors != nil && processors.NodeGroupListProcessor != nil {
		var errProc error
		nodeGroups, nodeInfos, errProc = processors.NodeGroupListProcessor.Process(context.AutoscalingContext,
			nodeGroups, nodeInfos, unschedulablePods)
		if errProc != nil {
			return &status.ScaleUpStatus{Result: status.ScaleUpError}, errors.ToAutoscalerError(errors.InternalError, errProc)
		}
	}

	podsPredicatePassingCheckFunctions := getPodsPredicatePassingCheckFunctions(context.AutoscalingContext,
		unschedulablePods, nodeInfos)
	getPodsPassingPredicates := podsPredicatePassingCheckFunctions.getPodsPassingPredicates
	getPodsNotPassingPredicates := podsPredicatePassingCheckFunctions.getPodsNotPassingPredicates

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

		scaleUpResourcesDeltaNow, err := computeScaleUpResourcesDelta(context.CloudProvider, nodeInfo,
			nodeGroup, resourceLimiter)
		if err != nil {
			klog.Errorf("Skipping node group %s; error getting node group resources: %v", nodeGroup.Id(), err)
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}
		checkResult := scaleUpResourcesLeft.checkScaleUpDeltaWithinLimits(scaleUpResourcesDeltaNow)
		if checkResult.exceeded {
			klog.V(4).Infof("Skipping node group %s; maximal limit exceeded for %v", nodeGroup.Id(),
				checkResult.exceededResources)
			skippedNodeGroups[nodeGroup.Id()] = maxResourceLimitReached(checkResult.exceededResources)
			continue
		}

		option := expander.Option{
			NodeGroup: nodeGroup,
			Pods:      make([]*apiv1.Pod, 0),
		}

		// add list of pods which pass predicates to option
		podsPassing, err := getPodsPassingPredicates(nodeGroup.Id())
		if err != nil {
			klog.V(4).Infof("Skipping node group %s; cannot compute pods passing predicates", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		} else {
			option.Pods = make([]*apiv1.Pod, len(podsPassing))
			copy(option.Pods, podsPassing)
		}

		// update information why we cannot schedule pods for which we did not find a working extension option so far
		podsNotPassing, err := getPodsNotPassingPredicates(nodeGroup.Id())
		if err != nil {
			klog.V(4).Infof("Skipping node group %s; cannot compute pods not passing predicates", nodeGroup.Id())
			skippedNodeGroups[nodeGroup.Id()] = notReadyReason
			continue
		}

		// mark that there is a scheduling option for pods which can be scheduled to node from currently analyzed node group
		for _, pod := range podsPassing {
			delete(podsRemainUnschedulable, pod)
		}

		for pod, err := range podsNotPassing {
			_, found := podsRemainUnschedulable[pod]
			if found {
				podsRemainUnschedulable[pod][nodeGroup.Id()] = err
			}
		}

		if len(option.Pods) > 0 || bufferNotEnough {
			estimator := context.ExtendedEstimatorBuilder(context.PredicateChecker, existingNodeInfos)
			if len(upcomingNodes) > 0 {
				klog.Infof("option.Pods: %+v, upcomingNodes: %v", len(upcomingNodes), upcomingNodes[0])
			} else {
				klog.Infof("option.Pods: %+v, upcomingNodes: %v", len(upcomingNodes), upcomingNodes)
			}
			option.NodeCount = estimator.Estimate(option.Pods, nodeInfo, upcomingNodes)
			if option.NodeCount > 0 {
				expansionOptions = append(expansionOptions, option)
			} else {
				klog.V(2).Infof("No need for any nodes in %s", nodeGroup.Id())
			}
		} else {
			klog.V(4).Infof("No pod can fit to %s", nodeGroup.Id())
		}
	}

	if len(expansionOptions) == 0 {
		klog.V(1).Info("No expansion options")
		return &status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable,
			PodsRemainUnschedulable: getRemainingPods(podsRemainUnschedulable, skippedNodeGroups),
			ConsideredNodeGroups:    nodeGroups}, nil
	}

	// Pick some expansion option.
	bestOption := context.ExpanderStrategy.BestOption(expansionOptions, nodeInfos)
	if bestOption != nil && bestOption.NodeCount > 0 {
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

			// Update ClusterStateRegistry so similar nodegroups rebalancing works.
			// TODO(lukaszos) when pursuing scalability update this call with one which takes list of changed node groups so
			//                we do not do extra API calls. (the call at the bottom of ScaleUp() could be also changed then)
			clusterStateRegistry.Recalculate()
		}

		nodeInfo, found := nodeInfos[bestOption.NodeGroup.Id()]
		if !found {
			// This should never happen, as we already should have retrieved
			// nodeInfo for any considered nodegroup.
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

		targetNodeGroups := []cloudprovider.NodeGroup{bestOption.NodeGroup}
		if context.BalanceSimilarNodeGroups {
			similarNodeGroups, typedErr := processors.NodeGroupSetProcessor.FindSimilarNodeGroups(context.AutoscalingContext,
				bestOption.NodeGroup, nodeInfos)
			if typedErr != nil {
				return &status.ScaleUpStatus{Result: status.ScaleUpError, CreateNodeGroupResults: createNodeGroupResults},
					typedErr.AddPrefix("Failed to find matching node groups: ")
			}
			similarNodeGroups = filterNodeGroupsByPods(similarNodeGroups, bestOption.Pods, getPodsPassingPredicates)
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
				PodsRemainUnschedulable: getRemainingPods(podsRemainUnschedulable, skippedNodeGroups),
				ConsideredNodeGroups:    nodeGroups,
				CreateNodeGroupResults:  createNodeGroupResults,
				PodsTriggeredScaleUp:    bestOption.Pods,
				PodsAwaitEvaluation:     getPodsAwaitingEvaluation(unschedulablePods, podsRemainUnschedulable, bestOption.Pods)},
			nil
	}

	return &status.ScaleUpStatus{Result: status.ScaleUpNoOptionsAvailable,
		PodsRemainUnschedulable: getRemainingPods(podsRemainUnschedulable, skippedNodeGroups),
		ConsideredNodeGroups:    nodeGroups}, nil
}

type podsPredicatePassingCheckFunctions struct {
	getPodsPassingPredicates    func(nodeGroupId string) ([]*apiv1.Pod, error)
	getPodsNotPassingPredicates func(nodeGroupId string) (map[*apiv1.Pod]status.Reasons, error)
}

func getPodsPredicatePassingCheckFunctions(
	context *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo) podsPredicatePassingCheckFunctions {

	podsPassingPredicatesCache := make(map[string][]*apiv1.Pod)
	podsNotPassingPredicatesCache := make(map[string]map[*apiv1.Pod]status.Reasons)
	errorsCache := make(map[string]error)
	checker := newPodsSchedulableOnNodeChecker(context, unschedulablePods)

	computeCaches := func(nodeGroupId string) {
		nodeInfo, found := nodeInfos[nodeGroupId]
		if !found {
			errorsCache[nodeGroupId] = errors.NewAutoscalerError(errors.InternalError,
				"NodeInfo not found for node group %v", nodeGroupId)
			return
		}

		podsPassing := make([]*apiv1.Pod, 0)
		podsNotPassing := make(map[*apiv1.Pod]status.Reasons)
		schedulableOnNode := checker.checkPodsSchedulableOnNode(nodeGroupId, nodeInfo)
		for pod, err := range schedulableOnNode {
			if err == nil {
				podsPassing = append(podsPassing, pod)
			} else {
				podsNotPassing[pod] = err
			}
		}
		podsPassingPredicatesCache[nodeGroupId] = podsPassing
		podsNotPassingPredicatesCache[nodeGroupId] = podsNotPassing
	}

	return podsPredicatePassingCheckFunctions{

		getPodsPassingPredicates: func(nodeGroupId string) ([]*apiv1.Pod, error) {
			_, passingFound := podsPassingPredicatesCache[nodeGroupId]
			_, errorFound := errorsCache[nodeGroupId]

			if !passingFound && !errorFound {
				computeCaches(nodeGroupId)
			}
			err, found := errorsCache[nodeGroupId]
			if found {
				return []*apiv1.Pod{}, err
			}
			pods, found := podsPassingPredicatesCache[nodeGroupId]
			if found {
				return pods, nil
			}
			return []*apiv1.Pod{}, errors.NewAutoscalerError(errors.InternalError,
				"Pods passing predicate entry not found in cache for node group %s", nodeGroupId)
		},

		getPodsNotPassingPredicates: func(nodeGroupId string) (map[*apiv1.Pod]status.Reasons, error) {
			_, notPassingFound := podsNotPassingPredicatesCache[nodeGroupId]
			_, errorFound := errorsCache[nodeGroupId]

			if !notPassingFound && !errorFound {
				computeCaches(nodeGroupId)
			}
			err, found := errorsCache[nodeGroupId]
			if found {
				return map[*apiv1.Pod]status.Reasons{}, err
			}
			pods, found := podsNotPassingPredicatesCache[nodeGroupId]
			if found {
				return pods, nil
			}
			return map[*apiv1.Pod]status.Reasons{}, errors.NewAutoscalerError(errors.InternalError,
				"Pods not passing predicate entry not found in cache for node group %s", nodeGroupId)
		},
	}
}

func getRemainingPods(schedulingErrors map[*apiv1.Pod]map[string]status.Reasons,
	skipped map[string]status.Reasons) []status.NoScaleUpInfo {
	remaining := []status.NoScaleUpInfo{}
	for pod, errs := range schedulingErrors {
		noScaleUpInfo := status.NoScaleUpInfo{
			Pod:                pod,
			RejectedNodeGroups: errs,
			SkippedNodeGroups:  skipped,
		}
		remaining = append(remaining, noScaleUpInfo)
	}
	return remaining
}

func getPodsAwaitingEvaluation(allPods []*apiv1.Pod, unschedulable map[*apiv1.Pod]map[string]status.Reasons,
	bestOption []*apiv1.Pod) []*apiv1.Pod {
	awaitsEvaluation := make(map[*apiv1.Pod]bool, len(allPods))
	for _, pod := range allPods {
		if _, found := unschedulable[pod]; !found {
			awaitsEvaluation[pod] = true
		}
	}
	for _, pod := range bestOption {
		delete(awaitsEvaluation, pod)
	}

	result := make([]*apiv1.Pod, 0)
	for pod := range awaitsEvaluation {
		result = append(result, pod)
	}
	return result
}

func filterNodeGroupsByPods(
	groups []cloudprovider.NodeGroup,
	podsRequiredToFit []*apiv1.Pod,
	fittingPodsPerNodeGroup func(groupId string) ([]*apiv1.Pod, error)) []cloudprovider.NodeGroup {

	result := make([]cloudprovider.NodeGroup, 0)

groupsloop:
	for _, group := range groups {
		fittingPods, err := fittingPodsPerNodeGroup(group.Id())
		if err != nil {
			klog.V(1).Infof("No info about pods passing predicates found for group %v,"+
				" skipping it from scale-up consideration; err=%v", group.Id(), err)
			continue
		}
		podSet := make(map[*apiv1.Pod]bool, len(fittingPods))
		for _, pod := range fittingPods {
			podSet[pod] = true
		}
		for _, pod := range podsRequiredToFit {
			if _, found := podSet[pod]; !found {
				klog.V(1).Infof("Group %v, can't fit pod %v/%v, removing from scale-up consideration", group.Id(),
					pod.Namespace, pod.Name)
				continue groupsloop
			}
		}
		result = append(result, group)
	}
	return result
}

func executeScaleUp(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry,
	info nodegroupset.ScaleUpInfo, gpuType string, now time.Time) errors.AutoscalerError {
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

func applyScaleUpResourcesLimits(
	cp cloudprovider.CloudProvider,
	newNodes int,
	scaleUpResourcesLeft scaleUpResourcesLimits,
	nodeInfo *schedulernodeinfo.NodeInfo,
	nodeGroup cloudprovider.NodeGroup,
	resourceLimiter *cloudprovider.ResourceLimiter) (int, errors.AutoscalerError) {

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
