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

// Package core implements the core methods of cluster autoscaler
package core

import (
	ctx "context"
	"fmt"
	"reflect"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	core_utils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
	estimatorinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/estimator"
	metricsinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/scalingconfig"
)

const (
	// How old the oldest unschedulable pod should be before starting scale up.
	unschedulablePodTimeBuffer = 2 * time.Second
	// How old the oldest unschedulable pod with GPU should be before starting scale up.
	// The idea is that nodes with GPU are very expensive and we're ready to sacrifice
	// a bit more latency to wait for more pods and make a more informed scale-up decision.
	unschedulablePodWithGpuTimeBuffer = 30 * time.Second

	// NodeUpcomingAnnotation is an annotation CA adds to nodes which are upcoming.
	NodeUpcomingAnnotation = "cluster-autoscaler.k8s.io/upcoming-node"
)

// BufferedAutoscaler is an autoscaler which has all the core functionality of a CA
// but without the reconfiguration feature
type BufferedAutoscaler struct {
	// AutoscalingContext consists of validated settings and options for this autoscaler
	*contextinternal.Context
	// ClusterState for maintaining the state of cluster nodes.
	clusterStateRegistry    *clusterstate.ClusterStateRegistry
	startTime               time.Time
	lastScaleUpTime         time.Time
	lastScaleDownDeleteTime time.Time
	lastScaleDownFailTime   time.Time
	scaleDown               *ScaleDown
	processors              *ca_processors.AutoscalingProcessors
	processorCallbacks      *bufferedAutoscalerProcessorCallbacks
	initialized             bool
	// last time that remove unregistered nodes
	lastRemoveUnregisteredTime time.Time
	scaleDownDelayAfterRemove  time.Duration
	// Caches nodeInfo computed for previously seen nodes
	nodeInfoCache       map[string]cacheItem
	ignoredTaints       taints.TaintKeySet
	CPURatio            float64
	MemRatio            float64
	ratio               float64
	webhook             Webhook
	maxBulkScaleUpCount int
}

type bufferedAutoscalerProcessorCallbacks struct {
	disableScaleDownForLoop bool
	extraValues             map[string]interface{}
	scaleDown               *ScaleDown
}

// ResetUnneededNodes TODO
func (callbacks *bufferedAutoscalerProcessorCallbacks) ResetUnneededNodes() {
	callbacks.scaleDown.CleanUpUnneededNodes()
}

// newBufferedAutoscalerProcessorCallbacks TODO
// NOCC:tosa/fn_length(设计如此)
func newBufferedAutoscalerProcessorCallbacks() *bufferedAutoscalerProcessorCallbacks {
	callbacks := &bufferedAutoscalerProcessorCallbacks{}
	callbacks.reset()
	return callbacks
}

// DisableScaleDownForLoop xxx
func (callbacks *bufferedAutoscalerProcessorCallbacks) DisableScaleDownForLoop() {
	callbacks.disableScaleDownForLoop = true
}

// SetExtraValue xxx
func (callbacks *bufferedAutoscalerProcessorCallbacks) SetExtraValue(key string, value interface{}) {
	callbacks.extraValues[key] = value
}

// GetExtraValue xxx
func (callbacks *bufferedAutoscalerProcessorCallbacks) GetExtraValue(key string) (value interface{}, found bool) {
	value, found = callbacks.extraValues[key]
	return
}

func (callbacks *bufferedAutoscalerProcessorCallbacks) reset() {
	callbacks.disableScaleDownForLoop = false
	callbacks.extraValues = make(map[string]interface{})
}

// NewBufferedAutoscaler creates an instance of Autoscaler filled with provided parameters
func NewBufferedAutoscaler(
	opts scalingconfig.Options,
	predicateChecker simulator.PredicateChecker,
	clusterSnapshot simulator.ClusterSnapshot,
	autoscalingKubeClients *context.AutoscalingKubeClients,
	processors *ca_processors.AutoscalingProcessors,
	cloudProvider cloudprovider.CloudProvider,
	expanderStrategy expander.Strategy,
	estimatorBuilder estimatorinternal.ExtendedEstimatorBuilder,
	backoff backoff.Backoff,
) core.Autoscaler {
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	autoscalingContext := contextinternal.NewAutoscalingContext(opts, predicateChecker,
		clusterSnapshot, autoscalingKubeClients, cloudProvider, expanderStrategy,
		estimatorBuilder, processorCallbacks)

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: opts.MaxTotalUnreadyPercentage,
		OkTotalUnreadyCount:       opts.OkTotalUnreadyCount,
		MaxNodeProvisionTime:      opts.MaxNodeProvisionTime,
		MaxNodeStartupTime:        opts.MaxNodeStartupTime,
		MaxNodeStartScheduleTime:  opts.MaxNodeStartScheduleTime,
	}

	ignoredTaints := make(taints.TaintKeySet)
	for _, taintKey := range opts.IgnoredTaints {
		klog.V(4).Infof("Ignoring taint %s on all NodeGroups", taintKey)
		ignoredTaints[taintKey] = true
	}

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(autoscalingContext.CloudProvider, clusterStateConfig,
		autoscalingContext.LogRecorder, backoff)

	scaleDown := NewScaleDown(autoscalingContext.AutoscalingContext, processors, clusterStateRegistry,
		opts.ExpendablePodsPriorityCutoff, opts.BufferedCPURatio, opts.BufferedMemRatio, opts.BufferedResourceRatio,
		opts.EvictLatest)
	processorCallbacks.scaleDown = scaleDown

	var webhook Webhook
	switch opts.WebhookMode {
	case WebMode:
		webhook = NewWebScaler(autoscalingContext.ClientSet, opts.ConfigNamespace,
			opts.WebhookModeConfig, opts.WebhookModeToken, opts.MaxBulkScaleUpCount, opts.BatchScaleUpCount)
		metricsinternal.RegisterWebhookParams("Web", opts.WebhookModeConfig)
	case ConfigMapMode:
		webhook = NewConfigMapScaler(autoscalingContext.ClientSet, opts.ConfigNamespace,
			opts.WebhookModeConfig, opts.MaxBulkScaleUpCount)
		metricsinternal.RegisterWebhookParams("ConfigMap", opts.WebhookModeConfig)
	default:
		webhook = nil
	}

	// Set the initial scale times to be less than the start time so as to
	// not start in cooldown mode.
	initialScaleTime := time.Now().Add(-time.Hour)
	return &BufferedAutoscaler{
		Context:                    autoscalingContext,
		lastScaleUpTime:            initialScaleTime,
		lastScaleDownDeleteTime:    initialScaleTime,
		lastScaleDownFailTime:      initialScaleTime,
		lastRemoveUnregisteredTime: initialScaleTime,
		scaleDownDelayAfterRemove:  opts.ScaleDownDelayAfterRemove,
		scaleDown:                  scaleDown,
		processors:                 processors,
		processorCallbacks:         processorCallbacks,
		clusterStateRegistry:       clusterStateRegistry,
		nodeInfoCache:              make(map[string]cacheItem),
		ignoredTaints:              ignoredTaints,
		CPURatio:                   opts.BufferedCPURatio,
		MemRatio:                   opts.BufferedMemRatio,
		ratio:                      opts.BufferedResourceRatio,
		webhook:                    webhook,
		maxBulkScaleUpCount:        opts.MaxBulkScaleUpCount,
	}
}

// Start starts components running in background.
func (b *BufferedAutoscaler) Start() error {
	b.clusterStateRegistry.Start()
	return nil
}

// cleanUpIfRequired removes ToBeDeleted taints added by a previous run of CA
// the taints are removed only once per runtime
func (b *BufferedAutoscaler) cleanUpIfRequired() {
	if b.initialized {
		return
	}

	// CA can die at any time. Removing taints that might have been left from the previous run.
	if readyNodes, err := b.ReadyNodeLister().List(); err != nil {
		klog.Errorf("Failed to list ready nodes, not cleaning up taints: %v", err)
	} else {
		deletetaint.CleanAllToBeDeleted(readyNodes,
			b.AutoscalingContext.ClientSet, b.Recorder, b.CordonNodeBeforeTerminate)
		if b.AutoscalingContext.AutoscalingOptions.MaxBulkSoftTaintCount == 0 {
			// Clean old taints if soft taints handling is disabled
			deletetaint.CleanAllDeletionCandidates(readyNodes,
				b.AutoscalingContext.ClientSet, b.Recorder)
		}
	}
	b.initialized = true
}

func (b *BufferedAutoscaler) initializeClusterSnapshot(nodes []*apiv1.Node,
	scheduledPods []*apiv1.Pod,
) errors.AutoscalerError {
	b.ClusterSnapshot.Clear()

	knownNodes := make(map[string]bool)
	for _, node := range nodes {
		if err := b.ClusterSnapshot.AddNode(node); err != nil {
			klog.Errorf("Failed to add node %s to cluster snapshot: %v", node.Name, err)
			return errors.ToAutoscalerError(errors.InternalError, err)
		}
		knownNodes[node.Name] = true
	}
	for _, pod := range scheduledPods {
		if knownNodes[pod.Spec.NodeName] {
			if err := b.ClusterSnapshot.AddPod(pod, pod.Spec.NodeName); err != nil {
				klog.Errorf("Failed to add pod %s scheduled to node %s to cluster snapshot: %v", pod.Name, pod.Spec.NodeName, err)
				return errors.ToAutoscalerError(errors.InternalError, err)
			}
		}
	}
	return nil
}

// RunOnce iterates over node groups and scales them up/down if necessary
// NOCC:golint/fnsize(设计如此)
// nolint
func (b *BufferedAutoscaler) RunOnce(currentTime time.Time) errors.AutoscalerError {
	stateUpdateStart := time.Now()
	klog.V(4).Info("Starting main loop")

	allNodes, readyNodes, typedErr := b.preRun(currentTime)
	if typedErr != nil {
		return typedErr
	}

	scheduledPodLister := b.ScheduledPodLister()
	scaleDown := b.scaleDown
	autoscalingContext := b.AutoscalingContext

	originalScheduledPods, err := scheduledPodLister.List()
	if err != nil {
		klog.Errorf("Failed to list scheduled pods: %v", err)
		return errors.ToAutoscalerError(errors.ApiCallError, err)
	}

	daemonsets, err := b.ListerRegistry.DaemonSetLister().List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to get daemonset list")
		return errors.ToAutoscalerError(errors.ApiCallError, err)
	}

	nonExpendableScheduledPods := core_utils.FilterOutExpendablePods(originalScheduledPods, b.ExpendablePodsPriorityCutoff)
	// Initialize cluster state to ClusterSnapshot
	if typedErr = b.initializeClusterSnapshot(allNodes, nonExpendableScheduledPods); typedErr != nil {
		return typedErr.AddPrefix("Initialize ClusterSnapshot")
	}

	nodeInfosForGroups, autoscalerError := getNodeInfosForGroups(
		readyNodes, b.nodeInfoCache, autoscalingContext.CloudProvider,
		autoscalingContext.ListerRegistry, daemonsets, autoscalingContext.PredicateChecker,
		b.ignoredTaints)
	if autoscalerError != nil {
		klog.Errorf("Failed to get node infos for groups: %v", autoscalerError)
		return autoscalerError.AddPrefix("failed to build node infos for node groups: ")
	}

	nodeInfosForGroups, err = b.processors.NodeInfoProcessor.Process(autoscalingContext, nodeInfosForGroups)
	if err != nil {
		klog.Errorf("Failed to process nodeInfos: %v", err)
		return errors.ToAutoscalerError(errors.InternalError, err)
	}

	if typedErr = b.updateClusterState(allNodes, nodeInfosForGroups, currentTime); typedErr != nil {
		klog.Errorf("Failed to update cluster state: %v", typedErr)
		return typedErr
	}
	metrics.UpdateDurationFromStart(metrics.UpdateState, stateUpdateStart)

	scaleUpStatus := &status.ScaleUpStatus{Result: status.ScaleUpNotTried}
	scaleUpStatusProcessorAlreadyCalled := false
	scaleDownStatus := &status.ScaleDownStatus{Result: status.ScaleDownNotTried}
	scaleDownStatusProcessorAlreadyCalled := false

	defer func() {
		// Update status information when the loop is done (regardless of reason)
		if autoscalingContext.WriteStatusConfigMap {
			CAStatus := b.clusterStateRegistry.GetStatus(currentTime)
			_, _ = utils.WriteStatusConfigMap(autoscalingContext.ClientSet, autoscalingContext.ConfigNamespace,
				CAStatus.GetReadableString(), b.AutoscalingContext.LogRecorder, b.AutoscalingContext.StatusConfigMapName)
		}

		// This deferred processor execution allows the processors to handle a situation when a scale-(up|down)
		// wasn't even attempted because e.g. the iteration exited earlier.
		if !scaleUpStatusProcessorAlreadyCalled && b.processors != nil && b.processors.ScaleUpStatusProcessor != nil {
			b.processors.ScaleUpStatusProcessor.Process(b.AutoscalingContext, scaleUpStatus)
		}
		if !scaleDownStatusProcessorAlreadyCalled && b.processors != nil && b.processors.ScaleDownStatusProcessor != nil {
			scaleDownStatus.SetUnremovableNodesInfo(scaleDown.unremovableNodeReasons,
				scaleDown.nodeUtilizationMap, scaleDown.context.CloudProvider)
			b.processors.ScaleDownStatusProcessor.Process(b.AutoscalingContext, scaleDownStatus)
		}

		processErr := b.processors.AutoscalingStatusProcessor.Process(b.AutoscalingContext,
			b.clusterStateRegistry, currentTime)
		if processErr != nil {
			klog.Errorf("AutoscalingStatusProcessor error: %v.", processErr)
		}
	}()

	skip, checkErr := b.checkClusterState(b.AutoscalingContext, currentTime, scaleDown, allNodes)
	if skip || checkErr != nil {
		return checkErr
	}
	metrics.UpdateLastTime(metrics.Autoscaling, time.Now())

	// execute webhook mode
	if b.webhook != nil {
		return b.webhook.DoWebhook(b.Context, b.clusterStateRegistry, b.scaleDown, allNodes, originalScheduledPods)
	}

	scaleUpStatus, scaleUpStatusProcessorAlreadyCalled, typedErr = b.doScaleUp(b.Context, currentTime,
		allNodes, readyNodes, nodeInfosForGroups, daemonsets)
	if typedErr != nil {
		return typedErr
	}
	if scaleUpStatus.Result == status.ScaleUpSuccessful {
		b.lastScaleUpTime = currentTime
		// No scale down in this iteration.
		scaleDownStatus.Result = status.ScaleDownInCooldown
		return nil
	}

	scaleDownStatus, scaleDownStatusProcessorAlreadyCalled, typedErr = b.doScaleDown(autoscalingContext,
		currentTime, allNodes)

	return typedErr
}

func (b *BufferedAutoscaler) preRun(currentTime time.Time) ([]*corev1.Node, []*corev1.Node, errors.AutoscalerError) {
	b.cleanUpIfRequired()
	b.processorCallbacks.reset()
	b.clusterStateRegistry.PeriodicCleanup()

	// Get nodes and pods currently living on cluster
	allNodes, readyNodes, typedErr := b.obtainNodeLists(b.CloudProvider)
	if typedErr != nil {
		return nil, nil, typedErr
	}
	if b.actOnEmptyCluster(allNodes, readyNodes, currentTime) {
		return nil, nil, nil
	}

	// Update cluster resource usage metrics
	coresTotal, memoryTotal := calculateScaleDownCoresMemoryTotal(allNodes, currentTime)
	metrics.UpdateClusterCPUCurrentCores(coresTotal)
	metrics.UpdateClusterMemoryCurrentBytes(memoryTotal)

	// Call CloudProvider.Refresh before any other calls to cloud provider.
	refreshStart := time.Now()
	err := b.AutoscalingContext.CloudProvider.Refresh()
	metrics.UpdateDurationFromStart(metrics.CloudProviderRefresh, refreshStart)
	if err != nil {
		klog.Errorf("Failed to refresh cloud provider config: %v", err)
		return nil, nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	// execute cron mode
	// move here before updating nodegroup's metrics, prevent reporting incorrect min sizes
	err = b.doCron(b.Context, b.clusterStateRegistry, currentTime)
	if err != nil {
		klog.Errorf("Failed in cron mode: %v", err)
	}

	// Update node groups min/max/current after cloud provider refresh
	for _, nodeGroup := range b.AutoscalingContext.CloudProvider.NodeGroups() {
		metrics.UpdateNodeGroupMin(nodeGroup.Id(), nodeGroup.MinSize())
		metrics.UpdateNodeGroupMax(nodeGroup.Id(), nodeGroup.MaxSize())
		if cur, err := nodeGroup.TargetSize(); err == nil {
			metrics.UpdateNodeGroupCurrent(nodeGroup.Id(), cur)
		}
	}
	// reset unremovable node metrics
	// metrics.ResetUnremovableNodes()

	return allNodes, readyNodes, nil
}

func (b *BufferedAutoscaler) checkClusterState(autoscalingContext *context.AutoscalingContext,
	currentTime time.Time, scaleDown *ScaleDown, allNodes []*apiv1.Node,
) (bool, errors.AutoscalerError) {
	// Check if there are any nodes that failed to register in Kubernetes
	// master.
	unregisteredNodes := b.clusterStateRegistry.GetUnregisteredNodes()
	if len(unregisteredNodes) > 0 && b.lastRemoveUnregisteredTime.Add(b.scaleDownDelayAfterRemove).Before(currentTime) {
		klog.V(1).Infof("%d unregistered nodes present", len(unregisteredNodes))
		removedAny, err := removeOldUnregisteredNodes(unregisteredNodes, autoscalingContext,
			b.clusterStateRegistry, currentTime, autoscalingContext.LogRecorder)
		// There was a problem with removing unregistered nodes. Retry in the next loop.
		if err != nil {
			klog.Warningf("Failed to remove unregistered nodes: %v", err)
		}
		if removedAny {
			b.lastRemoveUnregisteredTime = currentTime
			klog.V(0).Infof("Some unregistered nodes were removed")
		}
	}

	if !b.clusterStateRegistry.IsClusterHealthy() {
		klog.Warning("Cluster is not ready for autoscaling")
		scaleDown.CleanUpUnneededNodes()
		autoscalingContext.LogRecorder.Eventf(corev1.EventTypeWarning,
			"ClusterUnhealthy", "Cluster is unhealthy")
		return true, nil
	}

	b.deleteCreatedNodesWithErrors(allNodes)

	// Check if there has been a constant difference between the number of nodes in k8s and
	// the number of nodes on the cloud provider side.
	// DOTO: andrewskim - add protection for ready AWS nodes.
	fixedSomething, err := fixNodeGroupSize(autoscalingContext, b.clusterStateRegistry, currentTime)
	if err != nil {
		klog.Errorf("Failed to fix node group sizes: %v", err)
		return false, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}
	if fixedSomething {
		klog.V(0).Infof("Some node group target size was fixed, skipping the iteration")
		return true, nil
	}
	return false, nil
}

// doScaleUp TODO
// NOCC:golint/fnsize(设计如此)
// nolint
func (b *BufferedAutoscaler) doScaleUp(autoscalingContext *contextinternal.Context,
	currentTime time.Time, allNodes []*corev1.Node, readyNodes []*corev1.Node,
	nodeInfosForGroups map[string]*schedulerframework.NodeInfo,
	daemonsets []*appsv1.DaemonSet,
) (*status.ScaleUpStatus, bool, errors.AutoscalerError) {
	var scaleUpStatusProcessorAlreadyCalled bool
	var typedErr errors.AutoscalerError
	scaleUpStatus := &status.ScaleUpStatus{Result: status.ScaleUpNotTried}

	unschedulablePods, err := b.UnschedulablePodLister().List()
	if err != nil {
		klog.Errorf("Failed to list unscheduled pods: %v", err)
		return scaleUpStatus, scaleUpStatusProcessorAlreadyCalled,
			errors.ToAutoscalerError(errors.ApiCallError, err)
	}

	unschedulablePods = tpu.ClearTPURequests(unschedulablePods)

	// DOTO: move split and append below to separate PodListProcessor
	// Some unschedulable pods can be waiting for lower priority pods preemption so they have nominated node to run.
	// Such pods don't require scale up but should be considered during scale down.
	unschedulablePods, unschedulableWaitingForLowerPriorityPreemption := core_utils.FilterOutExpendableAndSplit(
		unschedulablePods, allNodes, b.ExpendablePodsPriorityCutoff)

	metrics.UpdateUnschedulablePodsCount(len(unschedulablePods))

	// modify the snapshot simulating scheduling of pods waiting for preemption.
	// this is not strictly correct as we are not simulating preemption itself but it matches
	// CA logic from before migration to scheduler framework. So let's keep it for now
	for _, p := range unschedulableWaitingForLowerPriorityPreemption {
		if addErr := b.ClusterSnapshot.AddPod(p, p.Status.NominatedNodeName); addErr != nil {
			klog.Errorf("Failed to update snapshot with pod %s waiting for preemption", addErr)
			return scaleUpStatus, scaleUpStatusProcessorAlreadyCalled,
				errors.ToAutoscalerError(errors.InternalError, addErr)
		}
	}

	// add upcoming nodes to ClusterSnapshot
	upcomingNodes := getUpcomingNodeInfos(b.clusterStateRegistry, nodeInfosForGroups)
	for _, upcomingNode := range upcomingNodes {
		var pods []*apiv1.Pod
		for _, podInfo := range upcomingNode.Pods {
			pods = append(pods, podInfo.Pod)
		}
		err = b.ClusterSnapshot.AddNodeWithPods(upcomingNode.Node(), pods)
		if err != nil {
			klog.Errorf("Failed to add upcoming node %s to cluster snapshot: %v", upcomingNode.Node().Name, err)
			return scaleUpStatus, scaleUpStatusProcessorAlreadyCalled,
				errors.ToAutoscalerError(errors.InternalError, err)
		}
	}

	// 过滤特殊资源
	prunedUnschedulablePods := make([]*apiv1.Pod, 0)
	for i := range unschedulablePods {
		pod := unschedulablePods[i].DeepCopy()
		for j := range pod.Spec.Containers {
			delete(pod.Spec.Containers[j].Resources.Requests, "cloud.bkbcs.tencent.com/eip")
			delete(pod.Spec.Containers[j].Resources.Requests, "tke.cloud.tencent.com/eni-ip")
			delete(pod.Spec.Containers[j].Resources.Requests, "tke.cloud.tencent.com/direct-eni")
			delete(pod.Spec.Containers[j].Resources.Requests, "ephemeral-storage")
			delete(pod.Spec.Containers[j].Resources.Requests, "bkbcs.tencent.com/cpuset")
			delete(pod.Spec.Containers[j].Resources.Requests, "extend.resource/diskIO")
			delete(pod.Spec.Containers[j].Resources.Requests, "extend.resource/disk")
			delete(pod.Spec.Containers[j].Resources.Requests, "extend.resource/netIO")
			delete(pod.Spec.Containers[j].Resources.Requests, "extend.resource/other")
		}
		for j := range pod.Spec.InitContainers {
			delete(pod.Spec.InitContainers[j].Resources.Requests, "cloud.bkbcs.tencent.com/eip")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "tke.cloud.tencent.com/eni-ip")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "tke.cloud.tencent.com/direct-eni")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "ephemeral-storage")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "bkbcs.tencent.com/cpuset")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "extend.resource/diskIO")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "extend.resource/disk")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "extend.resource/netIO")
			delete(pod.Spec.InitContainers[j].Resources.Requests, "extend.resource/other")
		}
		prunedUnschedulablePods = append(prunedUnschedulablePods, pod)
	}

	unschedulablePodsToHelp, _ := b.processors.PodListProcessor.Process(b.AutoscalingContext, prunedUnschedulablePods)

	// finally, filter out pods that are too "young" to safely be considered for a scale-up (delay is configurable)
	unschedulablePodsToHelp = b.filterOutYoungPods(unschedulablePodsToHelp, currentTime)

	allNodeInfos, err := b.ClusterSnapshot.NodeInfos().List()
	if err != nil {
		klog.Errorf("Unable to fetch ClusterNode List for Debugging Snapshot, %v", err)
	}
	nodesInfoWithoutMaster := filterOutMastersNodeInfo(allNodeInfos)
	bufferNotEnough := checkResourceNotEnough(nodesInfoWithoutMaster, nil, b.CPURatio, b.MemRatio, b.ratio)
	shouldScaleUp := false

	// nolint
	if len(unschedulablePodsToHelp) == 0 {
		scaleUpStatus.Result = status.ScaleUpNotNeeded
		klog.V(1).Info("No unschedulable pods")
	} else if b.MaxNodesTotal > 0 && len(readyNodes) >= b.MaxNodesTotal {
		scaleUpStatus.Result = status.ScaleUpNoOptionsAvailable
		klog.V(1).Info("Max total nodes in cluster reached")
	} else if allPodsAreNew(unschedulablePodsToHelp, currentTime) {
		// The assumption here is that these pods have been created very recently and probably there
		// is more pods to come. In theory we could check the newest pod time but then if pod were created
		// slowly but at the pace of 1 every 2 seconds then no scale up would be triggered for long time.
		// We also want to skip a real scale down (just like if the pods were handled).
		b.processorCallbacks.DisableScaleDownForLoop()
		scaleUpStatus.Result = status.ScaleUpInCooldown
		klog.V(1).Info("Unschedulable pods are very new, waiting one iteration for more")
	} else {
		shouldScaleUp = true
	}
	if bufferNotEnough || shouldScaleUp {
		klog.V(4).Infof("Will try scale up,  bufferNotEnough: %v, shouldScaleUp: %v", bufferNotEnough, shouldScaleUp)
		scaleUpStart := time.Now()
		metrics.UpdateLastTime(metrics.ScaleUp, scaleUpStart)

		scaleUpStatus, typedErr = ScaleUp(autoscalingContext, b.processors, b.clusterStateRegistry,
			unschedulablePodsToHelp, readyNodes, daemonsets, nodeInfosForGroups,
			b.ignoredTaints, bufferNotEnough, b.maxBulkScaleUpCount)

		metrics.UpdateDurationFromStart(metrics.ScaleUp, scaleUpStart)

		if b.processors != nil && b.processors.ScaleUpStatusProcessor != nil {
			b.processors.ScaleUpStatusProcessor.Process(autoscalingContext.AutoscalingContext, scaleUpStatus)
			scaleUpStatusProcessorAlreadyCalled = true
		}

		if typedErr != nil {
			klog.Errorf("Failed to scale up: %v", typedErr)
		}
	}
	return scaleUpStatus, scaleUpStatusProcessorAlreadyCalled, typedErr
}

// doScaleDown TODO
// NOCC:golint/fnsize(设计如此)
// nolint funlen
func (b *BufferedAutoscaler) doScaleDown(autoscalingContext *context.AutoscalingContext,
	currentTime time.Time, allNodes []*corev1.Node) (
	*status.ScaleDownStatus, bool, errors.AutoscalerError,
) {
	scaleDown := b.scaleDown

	scaleDownStatus := &status.ScaleDownStatus{Result: status.ScaleDownNotTried}
	scaleDownStatusProcessorAlreadyCalled := false

	if b.ScaleDownEnabled {
		pdbs, err := b.PodDisruptionBudgetLister().List()
		if err != nil {
			scaleDownStatus.Result = status.ScaleDownError
			klog.Errorf("Failed to list pod disruption budgets: %v", err)
			return scaleDownStatus, scaleDownStatusProcessorAlreadyCalled,
				errors.ToAutoscalerError(errors.ApiCallError, err)
		}

		unneededStart := time.Now()

		klog.V(4).Infof("Calculating unneeded nodes")

		scaleDownCandidates, podDestinations, processErr := b.processNodes(autoscalingContext, allNodes)
		if processErr != nil {
			return scaleDownStatus, scaleDownStatusProcessorAlreadyCalled, processErr
		}

		// We use scheduledPods (not originalScheduledPods) here, so artificial scheduled pods introduced by processors
		// (e.g unscheduled pods with nominated node name) can block scaledown of given node.
		if typedErr := scaleDown.UpdateUnneededNodes(podDestinations, scaleDownCandidates,
			currentTime, pdbs); typedErr != nil {
			scaleDownStatus.Result = status.ScaleDownError
			klog.Errorf("Failed to scale down: %v", typedErr)
			return scaleDownStatus, scaleDownStatusProcessorAlreadyCalled, typedErr
		}

		metrics.UpdateDurationFromStart(metrics.FindUnneeded, unneededStart)

		scaleDownInCooldown := b.processorCallbacks.disableScaleDownForLoop ||
			b.lastScaleUpTime.Add(b.ScaleDownDelayAfterAdd).After(currentTime) ||
			b.lastScaleDownFailTime.Add(b.ScaleDownDelayAfterFailure).After(currentTime) ||
			b.lastScaleDownDeleteTime.Add(b.ScaleDownDelayAfterDelete).After(currentTime)

		calculateUnneededOnly := scaleDownInCooldown || scaleDown.nodeDeletionTracker.IsNonEmptyNodeDeleteInProgress()

		klog.V(4).Infof("Scale down status: unneededOnly=%v lastScaleUpTime=%s "+
			"lastScaleDownDeleteTime=%v lastScaleDownFailTime=%s scaleDownForbidden=%v "+
			"isDeleteInProgress=%v scaleDownInCooldown=%v",
			calculateUnneededOnly, b.lastScaleUpTime,
			b.lastScaleDownDeleteTime, b.lastScaleDownFailTime, b.processorCallbacks.disableScaleDownForLoop,
			scaleDown.nodeDeletionTracker.IsNonEmptyNodeDeleteInProgress(), scaleDownInCooldown)
		metrics.UpdateScaleDownInCooldown(scaleDownInCooldown)

		if scaleDownInCooldown {
			scaleDownStatus.Result = status.ScaleDownInCooldown
		} else if scaleDown.nodeDeletionTracker.IsNonEmptyNodeDeleteInProgress() {
			scaleDownStatus.Result = status.ScaleDownInProgress
		} else {
			klog.V(4).Infof("Starting scale down")
			// We want to delete unneeded Node Groups only if there was no recent scale up,
			// and there is no current delete in progress and there was no recent errors.
			removedNodeGroups, err := b.processors.NodeGroupManager.RemoveUnneededNodeGroups(autoscalingContext)
			if err != nil {
				klog.Errorf("Error while removing unneeded node groups: %v", err)
			}

			scaleDownStart := time.Now()
			metrics.UpdateLastTime(metrics.ScaleDown, scaleDownStart)
			var downError errors.AutoscalerError
			scaleDownStatus, downError = scaleDown.TryToScaleDown(currentTime, pdbs)
			metrics.UpdateDurationFromStart(metrics.ScaleDown, scaleDownStart)
			metrics.UpdateUnremovableNodesCount(scaleDown.getUnremovableNodesCount())

			scaleDownStatus.RemovedNodeGroups = removedNodeGroups

			if scaleDownStatus.Result == status.ScaleDownNodeDeleteStarted {
				b.lastScaleDownDeleteTime = currentTime
				b.clusterStateRegistry.Recalculate()
			}

			if (scaleDownStatus.Result == status.ScaleDownNoNodeDeleted ||
				scaleDownStatus.Result == status.ScaleDownNoUnneeded) &&
				b.AutoscalingContext.AutoscalingOptions.MaxBulkSoftTaintCount != 0 {
				// 按序加污点, 与空节点列表排序逻辑保持一致
				sort.SliceStable(allNodes, func(i, j int) bool {
					return allNodes[i].Name < allNodes[j].Name
				})
				scaleDown.SoftTaintUnneededNodes(allNodes)
			}

			if b.processors != nil && b.processors.ScaleDownStatusProcessor != nil {
				scaleDownStatus.SetUnremovableNodesInfo(scaleDown.unremovableNodeReasons,
					scaleDown.nodeUtilizationMap, scaleDown.context.CloudProvider)
				b.processors.ScaleDownStatusProcessor.Process(autoscalingContext, scaleDownStatus)
				scaleDownStatusProcessorAlreadyCalled = true

				if downError != nil {
					klog.Errorf("Failed to scale down: %v", downError)
					b.lastScaleDownFailTime = currentTime
					return scaleDownStatus, scaleDownStatusProcessorAlreadyCalled, downError
				}

			}
		}
	}
	return scaleDownStatus, scaleDownStatusProcessorAlreadyCalled, nil
}

func (b *BufferedAutoscaler) processNodes(autoscalingContext *context.AutoscalingContext, allNodes []*apiv1.Node) (
	[]*apiv1.Node, []*apiv1.Node, errors.AutoscalerError,
) {
	var scaleDownCandidates []*apiv1.Node
	var podDestinations []*apiv1.Node

	// podDestinations and scaleDownCandidates are initialized based on allNodes variable, which contains only
	// registered nodes in cluster.
	// It does not include any upcoming nodes which can be part of clusterSnapshot. As an alternative to using
	// allNodes here, we could use nodes from clusterSnapshot and explicitly filter out upcoming nodes here but it
	// is of little (if any) benefit.

	if b.processors == nil || b.processors.ScaleDownNodeProcessor == nil {
		scaleDownCandidates = allNodes
		podDestinations = allNodes
	} else {
		var err errors.AutoscalerError
		scaleDownCandidates, err = b.processors.ScaleDownNodeProcessor.GetScaleDownCandidates(
			autoscalingContext, allNodes)
		if err != nil {
			klog.Error(err)
			return scaleDownCandidates, podDestinations, err
		}
		podDestinations, err = b.processors.ScaleDownNodeProcessor.GetPodDestinationCandidates(autoscalingContext, allNodes)
		if err != nil {
			klog.Error(err)
			return scaleDownCandidates, podDestinations, err
		}
	}
	return scaleDownCandidates, podDestinations, nil
}

func (b *BufferedAutoscaler) deleteCreatedNodesWithErrors(allNodes []*apiv1.Node) {
	// We always schedule deleting of incoming errornous nodes
	// DOTO[lukaszos] Consider adding logic to not retry delete every loop iteration
	nodes := b.clusterStateRegistry.GetCreatedNodesWithErrors()

	nodeGroups := b.nodeGroupsByID()
	nodesToBeDeletedByNodeGroupID := make(map[string][]*corev1.Node)

	for _, node := range nodes {
		nodeGroup, err := b.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			id := "<nil>"
			if node != nil {
				id = node.Name
			}
			klog.Warningf("Cannot determine nodeGroup for node %v; %v", id, err)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.Warningf("Skipped %s from delete with errors consideration - the node is not autoscaled", node.Name)
			continue
		}
		nodesToBeDeletedByNodeGroupID[nodeGroup.Id()] = append(nodesToBeDeletedByNodeGroupID[nodeGroup.Id()], node)
	}

	deletedAny := false
	for nodeGroupsID, nodesToBeDeleted := range nodesToBeDeletedByNodeGroupID {
		var err error
		klog.V(1).Infof("Deleting %v from %v node group because of create errors",
			len(nodesToBeDeleted), nodeGroupsID)

		nodeGroup := nodeGroups[nodeGroupsID]
		if nodeGroup == nil {
			err = fmt.Errorf("node group %s not found", nodeGroupsID)
		} else {
			// 扩容失败节点的缩容也需走 Pod 驱逐流程，防止意外情况
			for i := range nodesToBeDeleted {
				go func(node *apiv1.Node) {
					var result status.NodeDeleteResult
					defer func() { b.scaleDown.nodeDeletionTracker.AddNodeDeleteResult(node.Name, result) }()
					defer b.scaleDown.nodeDeletionTracker.SetNonEmptyNodeDeleteInProgress(false)
					name, ok := findNameFromAllNodes(node, allNodes)
					if !ok {
						klog.Errorf("Failed to get %s nodeName from allNodes", node.Name)
						return
					}
					freshNode, getErr := b.ClientSet.CoreV1().Nodes().Get(ctx.TODO(), name, metav1.GetOptions{})
					if getErr != nil || freshNode == nil {
						klog.Warningf("Error while get fresh node %v: %v", name, getErr)
						return
					}
					// 重新获取需驱逐 Pod 列表
					podList, listErr := b.ClientSet.CoreV1().Pods(metav1.NamespaceAll).List(ctx.TODO(), metav1.ListOptions{
						FieldSelector: fmt.Sprintf("spec.nodeName=%s", freshNode.Name),
					})
					if listErr != nil {
						klog.Warningf("Error while list pods on node %v: %v", name, listErr)
						return
					}
					toEvictPods, daemonsetPods := []*apiv1.Pod{}, []*apiv1.Pod{}
					for i := range podList.Items {
						if podList.Items[i].DeletionTimestamp != nil {
							continue
						}
						if pod_util.IsDaemonSetPod(&podList.Items[i]) {
							daemonsetPods = append(daemonsetPods, &podList.Items[i])
							continue
						}
						toEvictPods = append(toEvictPods, &podList.Items[i])
					}
					// deleteNode
					result = b.scaleDown.deleteNode(freshNode, toEvictPods, daemonsetPods, nodeGroup)
					if result.ResultType != status.NodeDeleteOk {
						klog.Errorf("Failed to delete %s: %v", name, result.Err)
						return
					}
				}(nodesToBeDeleted[i])
			}
		}

		if err != nil {
			klog.Warningf("Error while trying to delete nodes from %v: %v", nodeGroupsID, err)
		}
		deletedAny = deletedAny || err == nil
		b.clusterStateRegistry.InvalidateNodeInstancesCacheEntry(nodeGroup)
	}
}

func findNameFromAllNodes(node *apiv1.Node, allNodes []*apiv1.Node) (string, bool) {
	// clusterstate 中获取的 node, Name 字段填充的是 InternalIP，需要重新解析
	for i := range allNodes {
		for _, adr := range allNodes[i].Status.Addresses {
			if adr.Type == apiv1.NodeInternalIP && adr.Address == node.Name {
				return allNodes[i].Name, true
			}
		}
		// 如果获取不到实际节点 InternalIP, 但 nodeName 对得上, 可返回该值
		if allNodes[i].Name == node.Name {
			return allNodes[i].Name, true
		}
	}
	return "", false
}

func (b *BufferedAutoscaler) nodeGroupsByID() map[string]cloudprovider.NodeGroup {
	nodeGroups := make(map[string]cloudprovider.NodeGroup)
	for _, nodeGroup := range b.CloudProvider.NodeGroups() {
		nodeGroups[nodeGroup.Id()] = nodeGroup
	}
	return nodeGroups
}

// filterOutYoungPods TODO
// don't consider pods newer than newPodScaleUpDelay seconds old as unschedulable
func (b *BufferedAutoscaler) filterOutYoungPods(allUnschedulablePods []*corev1.Pod,
	currentTime time.Time,
) []*corev1.Pod {
	var oldUnschedulablePods []*corev1.Pod
	newPodScaleUpDelay := b.AutoscalingOptions.NewPodScaleUpDelay
	for _, pod := range allUnschedulablePods {
		podAge := currentTime.Sub(pod.CreationTimestamp.Time)
		if podAge > newPodScaleUpDelay {
			oldUnschedulablePods = append(oldUnschedulablePods, pod)
		} else {
			klog.V(3).Infof("Pod %s is %.3f seconds old, too new to consider unschedulable", pod.Name, podAge.Seconds())
		}
	}
	return oldUnschedulablePods
}

// ExitCleanUp performs all necessary clean-ups when the autoscaler's exiting.
func (b *BufferedAutoscaler) ExitCleanUp() {
	b.processors.CleanUp()

	if !b.AutoscalingContext.WriteStatusConfigMap {
		return
	}

	_ = utils.DeleteStatusConfigMap(b.AutoscalingContext.ClientSet,
		b.AutoscalingContext.ConfigNamespace, b.AutoscalingContext.StatusConfigMapName)

	b.clusterStateRegistry.Stop()
}

// obtainNodeLists TODO
// nolint
func (b *BufferedAutoscaler) obtainNodeLists(cp cloudprovider.CloudProvider) ([]*corev1.Node, []*corev1.Node,
	errors.AutoscalerError,
) {
	allNodes, err := b.AllNodeLister().List()
	if err != nil {
		klog.Errorf("Failed to list all nodes: %v", err)
		return nil, nil, errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	readyNodes, err := b.ReadyNodeLister().List()
	if err != nil {
		klog.Errorf("Failed to list ready nodes: %v", err)
		return nil, nil, errors.ToAutoscalerError(errors.ApiCallError, err)
	}

	// Handle GPU case - allocatable GPU may be equal to 0 up to 15 minutes after
	// node registers as ready. See https://github.com/kubernetes/kubernetes/issues/54959
	// Treat those nodes as unready until GPU actually becomes available and let
	// our normal handling for booting up nodes deal with this.
	// DOTO: Remove this call when we handle dynamically provisioned resources.
	allNodes, readyNodes = b.processors.CustomResourcesProcessor.FilterOutNodesWithUnreadyResources(
		b.AutoscalingContext, allNodes, readyNodes)
	allNodes, readyNodes = taints.FilterOutNodesWithIgnoredTaints(b.ignoredTaints, allNodes, readyNodes)
	return allNodes, readyNodes, nil
}

// actOnEmptyCluster returns true if the cluster was empty and thus acted upon
func (b *BufferedAutoscaler) actOnEmptyCluster(allNodes, readyNodes []*corev1.Node, currentTime time.Time) bool {
	if b.AutoscalingContext.AutoscalingOptions.ScaleUpFromZero {
		return false
	}
	if len(allNodes) == 0 {
		b.onEmptyCluster("Cluster has no nodes.", true)
		return true
	}
	if len(readyNodes) == 0 {
		// Cluster Autoscaler may start running before nodes are ready.
		// Timeout ensures no ClusterUnhealthy events are published immediately in this case.
		b.onEmptyCluster("Cluster has no ready nodes.", currentTime.After(b.startTime.Add(nodesNotReadyAfterStartTimeout)))
		return true
	}
	// the cluster is not empty
	return false
}

func (b *BufferedAutoscaler) updateClusterState(allNodes []*corev1.Node,
	nodeInfosForGroups map[string]*schedulerframework.NodeInfo, currentTime time.Time,
) errors.AutoscalerError {
	err := b.clusterStateRegistry.UpdateNodes(allNodes, nodeInfosForGroups, currentTime)
	if err != nil {
		klog.Errorf("Failed to update node registry: %v", err)
		b.scaleDown.CleanUpUnneededNodes()
		return errors.ToAutoscalerError(errors.CloudProviderError, err)
	}
	core_utils.UpdateClusterStateMetrics(b.clusterStateRegistry)

	// Update node groups upcoming after cluster registry refresh
	upcoming := b.clusterStateRegistry.GetUpcomingNodes()
	for _, nodeGroup := range b.AutoscalingContext.CloudProvider.NodeGroups() {
		metrics.UpdateNodeGroupUpcoming(nodeGroup.Id(), upcoming[nodeGroup.Id()])
	}

	return nil
}

func (b *BufferedAutoscaler) onEmptyCluster(status string, emitEvent bool) {
	klog.Warningf(status)
	b.scaleDown.CleanUpUnneededNodes()
	updateEmptyClusterStateMetrics()
	if b.AutoscalingContext.WriteStatusConfigMap {
		_, err := utils.WriteStatusConfigMap(b.AutoscalingContext.ClientSet, b.AutoscalingContext.ConfigNamespace, status,
			b.AutoscalingContext.LogRecorder, b.AutoscalingContext.StatusConfigMapName)
		if err != nil {
			klog.Errorf("WriteStatusConfigMap failed. Error: %v", err)
		}
	}
	if emitEvent {
		b.AutoscalingContext.LogRecorder.Eventf(corev1.EventTypeWarning, "ClusterUnhealthy", status)
	}
}
