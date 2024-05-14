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
	"math/rand"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	scheduler_utils "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	cloudproviderapi "k8s.io/cloud-provider/api"
	klog "k8s.io/klog/v2"
	nodeLabel "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	metricsinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/metrics"
)

const (
	// ReschedulerTaintKey is the name of the taint created by rescheduler.
	ReschedulerTaintKey = "CriticalAddonsOnly"

	gkeNodeTerminationHandlerTaint = "cloud.google.com/impending-node-termination"

	// filterNodeResourceAnnoKey filters nodes when calculating buffer and total resource
	filterNodeResourceAnnoKey = "io.tencent.bcs.dev/filter-node-resource"

	// nodeInstanceTypeLabelKey is the instance type of node
	nodeInstanceTypeLabelKey = "node.kubernetes.io/instance-type"
	// nodeInstanceTypeEklet indicates the instance type of node is eklet
	nodeInstanceTypeEklet = "eklet"

	// How long should Cluster Autoscaler wait for nodes to become ready after start.
	nodesNotReadyAfterStartTimeout = 10 * time.Minute

	valueTrue = "true"
)

var (
	nodeConditionTaints = taintKeySet{
		apiv1.TaintNodeNotReady:                     true,
		apiv1.TaintNodeUnreachable:                  true,
		apiv1.TaintNodeUnschedulable:                true,
		apiv1.TaintNodeMemoryPressure:               true,
		apiv1.TaintNodeDiskPressure:                 true,
		apiv1.TaintNodeNetworkUnavailable:           true,
		apiv1.TaintNodePIDPressure:                  true,
		cloudproviderapi.TaintExternalCloudProvider: true,
		cloudproviderapi.TaintNodeShutdown:          true,
		gkeNodeTerminationHandlerTaint:              true,
	}
)

var nodeInfoCacheExpiredTime = 10 * time.Minute

type cacheItem struct {
	*schedulerframework.NodeInfo
	added time.Time
}

func isCacheItemExpired(added time.Time) bool {
	return time.Since(added) > nodeInfoCacheExpiredTime
}

// Following data structure is used to avoid running predicates #pending_pods * #nodes
// times (which turned out to be very expensive if there are thousands of pending pods).
// This optimization is based on the assumption that if there are that many pods they're
// likely created by controllers (deployment, replication controller, ...).
// So instead of running all predicates for every pod we first check whether we've
// already seen identical pod (in this step we're not binpacking, just checking if
// the pod would fit anywhere right now) and if so we use the result we already
// calculated.
// To decide if two pods are similar enough we check if they have identical label
// and spec and are owned by the same controller. The problem is the whole
// podSchedulableInfo struct is not hashable and keeping a list and running deep
// equality checks would likely also be expensive. So instead we use controller
// UID as a key in initial lookup and only run full comparison on a set of
// podSchedulableInfos created for pods owned by this controller.
type podSchedulableInfo struct {
	spec            apiv1.PodSpec
	labels          map[string]string
	schedulingError *simulator.PredicateError
}

type podSchedulableMap map[string][]podSchedulableInfo

type taintKeySet map[string]bool

func (psi *podSchedulableInfo) match(pod *apiv1.Pod) bool {
	return reflect.DeepEqual(pod.Labels, psi.labels) && apiequality.Semantic.DeepEqual(pod.Spec, psi.spec)
}

func (podMap podSchedulableMap) get(pod *apiv1.Pod) (*simulator.PredicateError, bool) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return nil, false
	}
	uid := string(ref.UID)
	if infos, found := podMap[uid]; found {
		for _, info := range infos {
			if info.match(pod) {
				return info.schedulingError, true
			}
		}
	}
	return nil, false
}

func (podMap podSchedulableMap) set(pod *apiv1.Pod, err *simulator.PredicateError) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return
	}
	uid := string(ref.UID)
	podMap[uid] = append(podMap[uid], podSchedulableInfo{
		spec:            pod.Spec,
		labels:          pod.Labels,
		schedulingError: err,
	})
}

// filterOutExpendableAndSplit filters out expendable pods and splits into:
//   - waiting for lower priority pods preemption
//   - other pods.
func filterOutExpendableAndSplit(unschedulableCandidates []*apiv1.Pod,
	expendablePodsPriorityCutoff int) ([]*apiv1.Pod, []*apiv1.Pod) {
	var unschedulableNonExpendable []*apiv1.Pod
	var waitingForLowerPriorityPreemption []*apiv1.Pod
	for _, pod := range unschedulableCandidates {
		if pod.Spec.Priority != nil && int(*pod.Spec.Priority) < expendablePodsPriorityCutoff {
			klog.V(4).Infof("Pod %s has priority below %d (%d) and will scheduled when enough resources is free."+
				" Ignoring in scale up.", pod.Name, expendablePodsPriorityCutoff, *pod.Spec.Priority)
		} else if nominatedNodeName := pod.Status.NominatedNodeName; nominatedNodeName != "" {
			waitingForLowerPriorityPreemption = append(waitingForLowerPriorityPreemption, pod)
			klog.V(4).Infof("Pod %s will be scheduled after low priority pods are preempted on %s. Ignoring in scale up.",
				pod.Name, nominatedNodeName)
		} else {
			unschedulableNonExpendable = append(unschedulableNonExpendable, pod)
		}
	}
	return unschedulableNonExpendable, waitingForLowerPriorityPreemption
}

// filterOutExpendablePods filters out expendable pods.
func filterOutExpendablePods(pods []*apiv1.Pod, expendablePodsPriorityCutoff int) []*apiv1.Pod {
	var result []*apiv1.Pod
	for _, pod := range pods {
		if pod.Spec.Priority == nil || int(*pod.Spec.Priority) >= expendablePodsPriorityCutoff {
			result = append(result, pod)
		}
	}
	return result
}

// getNodeInfosForGroups finds NodeInfos for all node groups used to manage the given nodes.
// It also returns a node group to sample node mapping.
// DOTO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
// DOTO(mwielgus): Review error policy - sometimes we may continue with partial errors.
// NOCC:golint/fnsize(设计如此)
// nolint funlen
func getNodeInfosForGroups(nodes []*apiv1.Node, nodeInfoCache map[string]cacheItem,
	cloudProvider cloudprovider.CloudProvider, listers kube_util.ListerRegistry,
	daemonsets []*appsv1.DaemonSet, predicateChecker simulator.PredicateChecker,
	ignoredTaints taints.TaintKeySet) (map[string]*schedulerframework.NodeInfo, errors.AutoscalerError) {
	result := make(map[string]*schedulerframework.NodeInfo)
	seenGroups := make(map[string]bool)

	podsForNodes, err := getPodsForNodes(listers)
	if err != nil {
		return map[string]*schedulerframework.NodeInfo{}, err
	}

	// processNode returns information whether the nodeTemplate was generated and if there was an error.
	processNode := func(node *apiv1.Node) (bool, string, errors.AutoscalerError) {
		nodeGroup, getErr := cloudProvider.NodeGroupForNode(node)
		if getErr != nil {
			return false, "", errors.ToAutoscalerError(errors.CloudProviderError, getErr)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			return false, "", nil
		}
		id := nodeGroup.Id()
		if _, found := result[id]; !found {
			// Build nodeInfo.
			nodeInfo, buildErr := simulator.BuildNodeInfoForNode(node, podsForNodes)
			if buildErr != nil {
				return false, "", buildErr
			}
			sanitizedNodeInfo, sanErr := sanitizeNodeInfo(nodeInfo, id, ignoredTaints)
			if sanErr != nil {
				return false, "", sanErr
			}
			result[id] = sanitizedNodeInfo
			return true, id, nil
		}
		// if founded in result, but not found in cache, should add to node info cache
		if nodeInfoCache != nil {
			if _, found := nodeInfoCache[id]; !found {
				return true, id, nil
			}
		}
		return false, "", nil
	}

	seenGroups, result, err = getNodeInfos(cloudProvider, nodeInfoCache, daemonsets,
		predicateChecker, ignoredTaints, seenGroups, result)
	if err != nil {
		return map[string]*schedulerframework.NodeInfo{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	// if cannot get node template from provider, try to generate with real-world example
	for _, node := range nodes {
		// Broken nodes might have some stuff missing. Skipping.
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			continue
		}
		added, id, typedErr := processNode(node)
		if typedErr != nil {
			return map[string]*schedulerframework.NodeInfo{}, typedErr
		}
		// DOTO: support node pool label update
		if added && nodeInfoCache != nil {
			if nodeInfoCopy, err := deepCopyNodeInfo(result[id]); err == nil {
				nodeInfoCache[id] = cacheItem{NodeInfo: nodeInfoCopy, added: time.Now()}
			}
		}
	}

	// Remove invalid node groups from cache
	for id := range nodeInfoCache {
		if _, ok := seenGroups[id]; !ok {
			delete(nodeInfoCache, id)
		}
	}

	// Last resort - unready/unschedulable nodes.
	for _, node := range nodes {
		// Allowing broken nodes
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			added, _, typedErr := processNode(node)
			if typedErr != nil {
				return map[string]*schedulerframework.NodeInfo{}, typedErr
			}
			nodeGroup, err := cloudProvider.NodeGroupForNode(node)
			if err != nil {
				return map[string]*schedulerframework.NodeInfo{}, errors.ToAutoscalerError(
					errors.CloudProviderError, err)
			}
			if added {
				klog.Warningf("Built template for %s based on unready/unschedulable node %s", nodeGroup.Id(), node.Name)
			}
		}
	}

	return result, nil
}

func getNodeInfos(cloudProvider cloudprovider.CloudProvider,
	nodeInfoCache map[string]cacheItem,
	daemonsets []*appsv1.DaemonSet, predicateChecker simulator.PredicateChecker,
	ignoredTaints taints.TaintKeySet, seenGroups map[string]bool,
	result map[string]*schedulerframework.NodeInfo) (map[string]bool,
	map[string]*schedulerframework.NodeInfo, errors.AutoscalerError) {
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		seenGroups[id] = true
		if _, found := result[id]; found {
			continue
		}

		// No good template, check cache of previously running nodes.
		if nodeInfoCache != nil {
			if item, found := nodeInfoCache[id]; found {
				if isCacheItemExpired(item.added) {
					delete(nodeInfoCache, id)
				} else if nodeInfoCopy, err := deepCopyNodeInfo(item.NodeInfo); err == nil {
					result[id] = nodeInfoCopy
					continue
				}
			}
		}

		// No good template, trying to generate one.
		nodeInfo, err := getNodeInfoFromTemplate(nodeGroup, daemonsets, predicateChecker, ignoredTaints)
		if err != nil {
			if err == cloudprovider.ErrNotImplemented {
				continue
			} else {
				klog.Errorf("Unable to build proper template node for %s: %v", id, err)
				return seenGroups, result, errors.ToAutoscalerError(errors.CloudProviderError, err)
			}
		}
		result[id] = nodeInfo
	}
	return seenGroups, result, nil
}

func getPodsForNodes(listers kube_util.ListerRegistry) (map[string][]*apiv1.Pod, errors.AutoscalerError) {
	pods, err := listers.ScheduledPodLister().List()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	podsForNodes := map[string][]*apiv1.Pod{}
	for _, p := range pods {
		podsForNodes[p.Spec.NodeName] = append(podsForNodes[p.Spec.NodeName], p)
	}
	return podsForNodes, nil
}

// getNodeInfoFromTemplate returns NodeInfo object built base on TemplateNodeInfo
// returned by NodeGroup.TemplateNodeInfo().
func getNodeInfoFromTemplate(nodeGroup cloudprovider.NodeGroup, daemonsets []*appsv1.DaemonSet,
	predicateChecker simulator.PredicateChecker, ignoredTaints taints.TaintKeySet) (*schedulerframework.NodeInfo,
	errors.AutoscalerError) {
	id := nodeGroup.Id()
	baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	pods, err := daemonset.GetDaemonSetPodsForNode(baseNodeInfo, daemonsets, predicateChecker)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	for _, podInfo := range baseNodeInfo.Pods {
		pods = append(pods, podInfo.Pod)
	}
	fullNodeInfo := schedulerframework.NewNodeInfo(pods...)
	fullNodeInfo.SetNode(baseNodeInfo.Node())
	sanitizedNodeInfo, typedErr := sanitizeNodeInfo(fullNodeInfo, id, ignoredTaints)
	if typedErr != nil {
		return nil, typedErr
	}
	return sanitizedNodeInfo, nil
}

// filterOutNodesFromNotAutoscaledGroups return subset of input nodes for which cloud provider does not
// return autoscaled node group.
// NOCC:tosa/fn_length(设计如此)
func filterOutNodesFromNotAutoscaledGroups(nodes []*apiv1.Node, cloudProvider cloudprovider.CloudProvider) (
	[]*apiv1.Node, errors.AutoscalerError) {
	result := make([]*apiv1.Node, 0)

	for _, node := range nodes {
		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			return []*apiv1.Node{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			result = append(result, node)
		}
	}
	return result, nil
}

// nolint
func deepCopyNodeInfo(nodeInfo *schedulerframework.NodeInfo) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
	newPods := make([]*apiv1.Pod, 0)
	for _, podInfo := range nodeInfo.Pods {
		newPods = append(newPods, podInfo.Pod.DeepCopy())
	}

	// Build a new node info.
	newNodeInfo := schedulerframework.NewNodeInfo(newPods...)
	newNodeInfo.SetNode(nodeInfo.Node().DeepCopy())
	return newNodeInfo, nil
}

func sanitizeNodeInfo(nodeInfo *schedulerframework.NodeInfo, nodeGroupName string,
	ignoredTaints taints.TaintKeySet) (*schedulerframework.NodeInfo, errors.AutoscalerError) {
	// Sanitize node name.
	sanitizedNode, err := sanitizeTemplateNode(nodeInfo.Node(), nodeGroupName, ignoredTaints)
	if err != nil {
		return nil, err
	}

	// Update nodename in pods.
	sanitizedPods := make([]*apiv1.Pod, 0)
	for _, podInfo := range nodeInfo.Pods {
		sanitizedPod := podInfo.Pod.DeepCopy()
		sanitizedPod.Spec.NodeName = sanitizedNode.Name
		sanitizedPods = append(sanitizedPods, sanitizedPod)
	}

	// Build a new node info.
	sanitizedNodeInfo := schedulerframework.NewNodeInfo(sanitizedPods...)
	sanitizedNodeInfo.SetNode(sanitizedNode)
	return sanitizedNodeInfo, nil
}

// nolint
func sanitizeTemplateNode(node *apiv1.Node, nodeGroup string,
	ignoredTaints taints.TaintKeySet) (*apiv1.Node, errors.AutoscalerError) {
	newNode := node.DeepCopy()
	nodeName := fmt.Sprintf("template-node-for-%s-%d", nodeGroup, rand.Int63()) // nolint
	newNode.Labels = make(map[string]string, len(node.Labels))
	for k, v := range node.Labels {
		// if !validLabel(k) {
		// 	continue
		// }
		if k != apiv1.LabelHostname {
			newNode.Labels[k] = v
		} else {
			newNode.Labels[k] = nodeName
		}
	}
	newNode.Name = nodeName
	newTaints := make([]apiv1.Taint, 0)
	for _, taint := range node.Spec.Taints {
		// Rescheduler can put this taint on a node while evicting non-critical pods.
		// New nodes will not have this taint and so we should strip it when creating
		// template node.
		switch taint.Key {
		case ReschedulerTaintKey:
			klog.V(4).Infof("Removing rescheduler taint when creating template from node %s", node.Name)
			continue
		case deletetaint.ToBeDeletedTaint:
			klog.V(4).Infof("Removing autoscaler taint when creating template from node %s", node.Name)
			continue
		case deletetaint.DeletionCandidateTaint:
			klog.V(4).Infof("Removing autoscaler soft taint when creating template from node %s", node.Name)
			continue
		}

		// ignore conditional taints as they represent a transient node state.
		if exists := nodeConditionTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing node condition taint %s, when creating template from node %s", taint.Key, node.Name)
			continue
		}

		if exists := ignoredTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing ignored taint %s, when creating template from node %s", taint.Key, node.Name)
			continue
		}

		newTaints = append(newTaints, taint)
	}
	newNode.Spec.Taints = newTaints
	return newNode, nil
}

// removeOldUnregisteredNodes removes unregistered nodes if needed. Returns true if anything
// was removed and error if such occurred.
func removeOldUnregisteredNodes(unregisteredNodes []clusterstate.UnregisteredNode, context *context.AutoscalingContext,
	csr *clusterstate.ClusterStateRegistry, currentTime time.Time, logRecorder *utils.LogEventRecorder) (bool, error) {
	removedAny := false
	for _, unregisteredNode := range unregisteredNodes {
		if unregisteredNode.UnregisteredSince.Add(context.MaxNodeProvisionTime).Before(currentTime) {
			klog.V(0).Infof("Removing unregistered node %v", unregisteredNode.Node.Name)
			nodeGroup, err := context.CloudProvider.NodeGroupForNode(unregisteredNode.Node)
			if err != nil {
				klog.Warningf("Failed to get node group for %s: %v", unregisteredNode.Node.Name, err)
				return removedAny, err
			}
			if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
				klog.Warningf("No node group for node %s, skipping", unregisteredNode.Node.Name)
				continue
			}
			size, err := nodeGroup.TargetSize()
			if err != nil {
				klog.Warningf("Failed to get node group size; unregisteredNode=%v; nodeGroup=%v; err=%v",
					unregisteredNode.Node.Name, nodeGroup.Id(), err)
				continue
			}
			if nodeGroup.MinSize() >= size {
				klog.Warningf("Failed to remove node %s: node group min size reached, skipping unregistered node removal",
					unregisteredNode.Node.Name)
				continue
			}
			err = nodeGroup.DeleteNodes([]*apiv1.Node{unregisteredNode.Node})
			csr.InvalidateNodeInstancesCacheEntry(nodeGroup)
			if err != nil {
				klog.Warningf("Failed to remove node %s: %v", unregisteredNode.Node.Name, err)
				logRecorder.Eventf(apiv1.EventTypeWarning, "DeleteUnregisteredFailed",
					"Failed to remove node %s: %v", unregisteredNode.Node.Name, err)
				return removedAny, err
			}
			logRecorder.Eventf(apiv1.EventTypeNormal, "DeleteUnregistered",
				"Removed unregistered node %v", unregisteredNode.Node.Name)
			metrics.RegisterOldUnregisteredNodesRemoved(1)
			removedAny = true
		}
	}
	return removedAny, nil
}

// fixNodeGroupSize sets the target size of node groups to the current number of nodes in them
// if the difference was constant for a prolonged time. Returns true if managed
// to fix something.
func fixNodeGroupSize(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry,
	currentTime time.Time) (bool, error) {
	fixed := false
	for _, nodeGroup := range context.CloudProvider.NodeGroups() {
		incorrectSize := clusterStateRegistry.GetIncorrectNodeGroupSize(nodeGroup.Id())
		if incorrectSize == nil {
			continue
		}
		// set the target size of node groups to the current number of nodes
		// may decrease or increase the target size
		if incorrectSize.FirstObserved.Add(context.MaxNodeProvisionTime).Before(currentTime) {
			delta := incorrectSize.CurrentSize - incorrectSize.ExpectedSize
			klog.V(0).Infof("Fix size of %s, expected=%d current=%d delta=%d", nodeGroup.Id(),
				incorrectSize.ExpectedSize, incorrectSize.CurrentSize, delta)
			if err := nodeGroup.DecreaseTargetSize(delta); err != nil {
				return fixed, fmt.Errorf("failed to fix %s: %v", nodeGroup.Id(), err)
			}
			fixed = true
		}
	}
	return fixed, nil
}

func getNodeCoresAndMemory(node *apiv1.Node) (int64, int64) {
	// filter eklet node
	if node.Labels[nodeInstanceTypeLabelKey] == nodeInstanceTypeEklet {
		return 0, 0
	}
	if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
		return 0, 0
	}
	cores := getNodeResource(node, apiv1.ResourceCPU)
	memory := getNodeResource(node, apiv1.ResourceMemory)
	return cores, memory
}

func getNodeResource(node *apiv1.Node, resource apiv1.ResourceName) int64 {
	nodeCapacity, found := node.Status.Capacity[resource]
	if !found {
		return 0
	}

	nodeCapacityValue := nodeCapacity.Value()
	if nodeCapacityValue < 0 {
		nodeCapacityValue = 0
	}

	return nodeCapacityValue
}

// UpdateClusterStateMetrics updates metrics related to cluster state
func UpdateClusterStateMetrics(csr *clusterstate.ClusterStateRegistry) {
	if csr == nil || reflect.ValueOf(csr).IsNil() {
		return
	}
	metrics.UpdateClusterSafeToAutoscale(csr.IsClusterHealthy())
	readiness := csr.GetClusterReadiness()
	// fix(bcs): 删除中节点也应统计到指标中, 为减少改动, 添加到 Unregistered 中
	metrics.UpdateNodesCount(readiness.Ready, readiness.Unready, readiness.NotStarted,
		readiness.LongUnregistered, readiness.Unregistered+readiness.Deleted)
}

func getOldestCreateTime(pods []*apiv1.Pod) time.Time {
	oldest := time.Now()
	for _, pod := range pods {
		if oldest.After(pod.CreationTimestamp.Time) {
			oldest = pod.CreationTimestamp.Time
		}
	}
	return oldest
}

func getOldestCreateTimeWithGpu(pods []*apiv1.Pod) (bool, time.Time) {
	oldest := time.Now()
	gpuFound := false
	for _, pod := range pods {
		if gpu.PodRequestsGpu(pod) {
			gpuFound = true
			if oldest.After(pod.CreationTimestamp.Time) {
				oldest = pod.CreationTimestamp.Time
			}
		}
	}
	return gpuFound, oldest
}

// updateEmptyClusterStateMetrics updates metrics related to empty cluster's state.
// DOTO(aleksandra-malinowska): use long unregistered value from ClusterStateRegistry.
func updateEmptyClusterStateMetrics() {
	metrics.UpdateClusterSafeToAutoscale(false)
	metrics.UpdateNodesCount(0, 0, 0, 0, 0)
}

func allPodsAreNew(pods []*apiv1.Pod, currentTime time.Time) bool {
	if getOldestCreateTime(pods).Add(unschedulablePodTimeBuffer).After(currentTime) {
		return true
	}
	found, oldest := getOldestCreateTimeWithGpu(pods)
	return found && oldest.Add(unschedulablePodWithGpuTimeBuffer).After(currentTime)
}

func getUpcomingNodeInfos(registry *clusterstate.ClusterStateRegistry,
	nodeInfos map[string]*schedulerframework.NodeInfo) []*schedulerframework.NodeInfo {
	upcomingNodes := make([]*schedulerframework.NodeInfo, 0)
	for nodeGroup, numberOfNodes := range registry.GetUpcomingNodes() {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			klog.Warningf("Couldn't find template for node group %s", nodeGroup)
			continue
		}

		if nodeTemplate.Node().Annotations == nil {
			nodeTemplate.Node().Annotations = make(map[string]string)
		}
		nodeTemplate.Node().Annotations[NodeUpcomingAnnotation] = valueTrue

		for i := 0; i < numberOfNodes; i++ {
			// Ensure new nodes have different names because nodeName
			// will be used as a map key. Also deep copy pods (daemonsets &
			// any pods added by cloud provider on template).
			upcomingNodes = append(upcomingNodes,
				scheduler_utils.DeepCopyTemplateNode(nodeTemplate, fmt.Sprintf("upcoming-%d", i)))
		}
	}
	return upcomingNodes
}

func checkResourceNotEnough(nodes map[string]*schedulerframework.NodeInfo,
	podsToReschedule []*apiv1.Pod, cpuRatio, memRatio, ratio float64) bool {
	sumResources := &schedulerframework.Resource{}
	leftResources := &schedulerframework.Resource{}
	for _, nodeInfo := range nodes {
		node := nodeInfo.Node()
		if node == nil {
			continue
		}
		if node.Spec.Unschedulable {
			continue
		}
		if node.Labels[apiv1.LabelInstanceTypeStable] == "eklet" {
			continue
		}
		if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
			continue
		}
		if node.Labels[nodeLabel.LabelNodeRoleControlPlane] == valueTrue {
			continue
		}
		if node.Labels[nodeLabel.LabelNodeRoleOldControlPlane] == valueTrue {
			continue
		}
		sumResources.Add(scheduler.ResourceToResourceList(nodeInfo.Allocatable))
		leftResources.Add(singleNodeResource(nodeInfo))
	}

	if len(podsToReschedule) > 0 {
		leftResources = substractRescheduledPodResources(leftResources,
			podsToReschedule)
	}

	leftResourcesList := scheduler.ResourceToResourceList(leftResources)
	sumResourcesList := scheduler.ResourceToResourceList(sumResources)

	for name, sum := range sumResourcesList {
		left, ok := leftResourcesList[name]
		if !ok {
			continue
		}
		if sum.IsZero() {
			continue
		}
		r := float64(left.MilliValue()) / float64(sum.MilliValue())
		metricsinternal.UpdateResourceUsedRatio(name.String(), 1.0-r)
		switch name {
		case apiv1.ResourceCPU:
			klog.V(4).Infof("%v ratio %v, desired CPU ratio %v", name, r, cpuRatio)
			if r < cpuRatio {
				return true
			}
		case apiv1.ResourceMemory:
			klog.V(4).Infof("%v ratio %v, desired Memory ratio %v", name, r, memRatio)
			if r < memRatio {
				return true
			}
		default:
			klog.V(4).Infof("%v ratio %v, desired ratio %v", name, r, ratio)
			if r < ratio {
				return true
			}
		}
	}
	return false
}

func substractRescheduledPodResources(leftResources *schedulerframework.Resource,
	podsToReschedule []*apiv1.Pod) *schedulerframework.Resource {

	podResources := &schedulerframework.Resource{
		ScalarResources: make(map[apiv1.ResourceName]int64),
	}
	for _, pod := range podsToReschedule {
		for _, container := range pod.Spec.Containers {
			podResources.Add(container.Resources.Requests)
		}
	}

	leftResources.AllowedPodNumber -= len(podsToReschedule)
	leftResources.MilliCPU -= podResources.MilliCPU
	leftResources.Memory -= podResources.Memory
	leftResources.EphemeralStorage -= podResources.EphemeralStorage

	// calculate extend resources
	for k, v := range podResources.ScalarResources {
		_, ok := leftResources.ScalarResources[k]
		if ok {
			leftResources.ScalarResources[k] -= v
		}
	}

	return leftResources
}

func singleNodeResource(info *schedulerframework.NodeInfo) apiv1.ResourceList {
	leftResource := &schedulerframework.Resource{
		ScalarResources: make(map[apiv1.ResourceName]int64),
	}

	allocatable := info.Allocatable
	requested := info.Requested

	podCount := requested.AllowedPodNumber
	if podCount == 0 {
		podCount = len(info.Pods)
	}

	if allocatable == nil { // nolint
		klog.Warningf("allocatable is nil: %v", info.Node().Name)
	}

	leftResource.AllowedPodNumber = allocatable.AllowedPodNumber - podCount                   // nolint
	leftResource.MilliCPU = allocatable.MilliCPU - requested.MilliCPU                         // nolint
	leftResource.Memory = allocatable.Memory - requested.Memory                               // nolint
	leftResource.EphemeralStorage = allocatable.EphemeralStorage - requested.EphemeralStorage // nolint

	// calculate extend resources
	for k, allocatableEx := range allocatable.ScalarResources { // nolint
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
