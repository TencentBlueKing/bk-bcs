/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// whenModePattern limits supported when syntax to mode/operation equality checks.
var whenModePattern = regexp.MustCompile(`^\s*(?:mode|operation)\s*==\s*["']?([a-zA-Z]+)["']?\s*$`)

// modeRollback is the normalized runtime mode used during revert execution.
const modeRollback = "rollback"

// NativeWorkflowExecutor implements WorkflowExecutor for native execution
type NativeWorkflowExecutor struct {
	client   client.Client
	registry Registry
}

// NewNativeWorkflowExecutor creates a new NativeWorkflowExecutor
func NewNativeWorkflowExecutor(client client.Client, registry Registry) *NativeWorkflowExecutor {
	return &NativeWorkflowExecutor{
		client:   client,
		registry: registry,
	}
}

// ExecuteWorkflow executes a workflow using batch-based scheduling.
// When any action specifies DependsOn, the workflow switches to DAG-based scheduling
// where actions with no inter-dependencies execute concurrently.
// Otherwise, consecutive PerCluster actions are grouped and executed concurrently per binding cluster,
// while Global actions execute sequentially (preserving backward compatibility).
// NOCC:tosa/fn_length(设计如此)
func (e *NativeWorkflowExecutor) ExecuteWorkflow(ctx context.Context, workflow *drv1alpha1.DRWorkflow, params map[string]interface{}) (*drv1alpha1.WorkflowExecutionStatus, error) {
	klog.Infof("Executing workflow: %s/%s", workflow.Namespace, workflow.Name)

	status := &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef: drv1alpha1.ObjectReference{
			Name:      workflow.Name,
			Namespace: workflow.Namespace,
		},
		Phase:          "Running",
		StartTime:      &metav1.Time{Time: time.Now()},
		ActionStatuses: make([]drv1alpha1.ActionStatus, 0, len(workflow.Spec.Actions)),
	}

	if hasDependsOn(workflow.Spec.Actions) {
		return e.executeDAGWorkflow(ctx, workflow, params, status)
	}

	return e.executeBatchWorkflow(ctx, workflow, params, status)
}

// executeBatchWorkflow is the original batch-based execution path (no DependsOn).
func (e *NativeWorkflowExecutor) executeBatchWorkflow(ctx context.Context, workflow *drv1alpha1.DRWorkflow, params map[string]interface{}, status *drv1alpha1.WorkflowExecutionStatus) (*drv1alpha1.WorkflowExecutionStatus, error) {
	batches := groupActionBatches(workflow.Spec.Actions)

	for batchIdx, batch := range batches {
		klog.V(2).Infof("Executing batch %d/%d (perCluster=%v, actions=%d)",
			batchIdx+1, len(batches), batch.perCluster, len(batch.actions))

		if batch.perCluster {
			batchStatuses, err := e.executePerClusterBatch(ctx, batch.actions, params, workflow.Spec.FailurePolicy)
			status.ActionStatuses = append(status.ActionStatuses, batchStatuses...)
			if err != nil {
				status.Phase = drv1alpha1.PhaseFailed
				status.Message = fmt.Sprintf("PerCluster batch failed: %v", err)
				return e.finalizeWorkflowStatus(status), err
			}
		} else {
			for _, action := range batch.actions {
				actionStatus, shouldStop, err := e.executeSingleGlobalAction(ctx, action, params, workflow, status)
				status.ActionStatuses = append(status.ActionStatuses, actionStatus)
				if shouldStop {
					return e.finalizeWorkflowStatus(status), err
				}
			}
		}
	}

	updateWorkflowFinalPhase(status, workflow.Namespace, workflow.Name)
	return e.finalizeWorkflowStatus(status), nil
}

// executeDAGWorkflow runs actions according to DependsOn DAG topology.
// Actions in the same layer (no inter-dependencies) execute concurrently.
// Each layer may contain a mix of Global and PerCluster actions; both groups
// are dispatched concurrently and merged in original definition order.
// NOCC:tosa/fn_length(设计如此)
func (e *NativeWorkflowExecutor) executeDAGWorkflow(
	ctx context.Context,
	workflow *drv1alpha1.DRWorkflow,
	params map[string]interface{},
	status *drv1alpha1.WorkflowExecutionStatus,
) (*drv1alpha1.WorkflowExecutionStatus, error) {
	forward, inDegree, err := buildActionGraph(workflow.Spec.Actions)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.Message = fmt.Sprintf("Invalid dependsOn: %v", err)
		klog.Errorf("Workflow %s/%s DAG validation failed: %v", workflow.Namespace, workflow.Name, err)
		return e.finalizeWorkflowStatus(status), err
	}

	layers, err := topoSortLayers(workflow.Spec.Actions, forward, inDegree)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.Message = fmt.Sprintf("Invalid dependsOn: %v", err)
		klog.Errorf("Workflow %s/%s DAG topological sort failed: %v", workflow.Namespace, workflow.Name, err)
		return e.finalizeWorkflowStatus(status), err
	}

	klog.Infof("Workflow %s/%s DAG: %d layers", workflow.Namespace, workflow.Name, len(layers))

	for layerIdx, layer := range layers {
		klog.V(2).Infof("Executing DAG layer %d/%d (%d actions)", layerIdx+1, len(layers), len(layer))

		layerStatuses, layerErr := e.executeDAGLayer(ctx, layer, params, workflow)
		status.ActionStatuses = append(status.ActionStatuses, layerStatuses...)
		if layerErr != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.Message = fmt.Sprintf("DAG layer %d failed: %v", layerIdx+1, layerErr)
			return e.finalizeWorkflowStatus(status), layerErr
		}
	}

	updateWorkflowFinalPhase(status, workflow.Namespace, workflow.Name)
	return e.finalizeWorkflowStatus(status), nil
}

// dagLayerSlot tracks an action's position in the original layer for ordered result merging.
type dagLayerSlot struct {
	index  int
	action drv1alpha1.Action
}

// executeDAGLayer runs a set of independent actions within one DAG layer.
// It splits the layer into PerCluster and Global groups, executes them concurrently,
// and merges results in the original definition order.
// NOCC:tosa/fn_length(设计如此)
func (e *NativeWorkflowExecutor) executeDAGLayer(
	ctx context.Context,
	layer []drv1alpha1.Action,
	params map[string]interface{},
	workflow *drv1alpha1.DRWorkflow,
) ([]drv1alpha1.ActionStatus, error) {
	var pcSlots []dagLayerSlot
	var globalSlots []dagLayerSlot
	for i, action := range layer {
		if isPerClusterMode(&layer[i]) {
			pcSlots = append(pcSlots, dagLayerSlot{index: i, action: action})
		} else {
			globalSlots = append(globalSlots, dagLayerSlot{index: i, action: action})
		}
	}

	results := make([]drv1alpha1.ActionStatus, len(layer))
	layerCtx, layerCancel := context.WithCancel(ctx)
	defer layerCancel()

	var wg sync.WaitGroup
	// pcErr is only written by the single pcSlots goroutine; safe to read after wg.Wait().
	var pcErr error

	if len(pcSlots) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pcActions := make([]drv1alpha1.Action, len(pcSlots))
			for i, s := range pcSlots {
				pcActions[i] = s.action
			}
			statuses, err := e.executePerClusterBatch(layerCtx, pcActions, params, workflow.Spec.FailurePolicy)
			for i, s := range pcSlots {
				if i < len(statuses) {
					results[s.index] = statuses[i]
				}
			}
			if err != nil {
				pcErr = err
				if isFailFast(workflow.Spec.FailurePolicy) {
					layerCancel()
				}
			}
		}()
	}

	if len(globalSlots) > 0 {
		for _, slot := range globalSlots {
			wg.Add(1)
			go func(s dagLayerSlot) {
				defer wg.Done()
				results[s.index] = e.executeDAGGlobalAction(layerCtx, s.action, params, workflow, &layerCancel)
			}(slot)
		}
	}

	wg.Wait()

	if pcErr != nil {
		return results, pcErr
	}
	for i := range results {
		if results[i].Phase == drv1alpha1.PhaseFailed && isFailFast(workflow.Spec.FailurePolicy) {
			return results, fmt.Errorf("action %s failed: %s", results[i].Name, results[i].Message)
		}
	}
	return results, nil
}

// executeDAGGlobalAction executes a single Global action within a DAG layer goroutine.
// It handles when-filtering, context cancellation, and FailFast signaling.
func (e *NativeWorkflowExecutor) executeDAGGlobalAction(
	ctx context.Context,
	action drv1alpha1.Action,
	params map[string]interface{},
	workflow *drv1alpha1.DRWorkflow,
	cancelFn *context.CancelFunc,
) drv1alpha1.ActionStatus {
	if ctx.Err() != nil {
		return drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseFailed,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: "canceled by FailFast",
		}
	}

	shouldExecute, skipReason, whenErr := shouldExecuteActionByWhen(action.When, params)
	if whenErr != nil {
		as := drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseFailed,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: fmt.Sprintf("invalid when expression %q: %v", action.When, whenErr),
		}
		if isFailFast(workflow.Spec.FailurePolicy) {
			(*cancelFn)()
		}
		return as
	}
	if !shouldExecute {
		klog.Infof("Skipping action %s due to when condition: %s", action.Name, skipReason)
		return drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseSkipped,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: skipReason,
		}
	}

	actionExecutor, getErr := e.registry.GetExecutor(action.Type)
	if getErr != nil {
		as := drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseFailed,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: fmt.Sprintf("Executor not found: %v", getErr),
		}
		if isFailFast(workflow.Spec.FailurePolicy) {
			(*cancelFn)()
		}
		return as
	}

	actionStatus, execErr := actionExecutor.Execute(ctx, &action, params)
	if execErr != nil {
		klog.Errorf("Action %s failed: %v", action.Name, execErr)
		if actionStatus != nil {
			actionStatus.Phase = drv1alpha1.PhaseFailed
			actionStatus.Message = execErr.Error()
		}
	}

	if actionStatus == nil {
		actionStatus = &drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseFailed,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: "executor returned nil status",
		}
	}

	if actionStatus.Phase == drv1alpha1.PhaseFailed && isFailFast(workflow.Spec.FailurePolicy) {
		(*cancelFn)()
	}
	return *actionStatus
}

// executeSingleGlobalAction executes one Global action and returns its status.
// Returns (status, shouldStop, error). shouldStop=true means workflow should abort.
func (e *NativeWorkflowExecutor) executeSingleGlobalAction(
	ctx context.Context,
	action drv1alpha1.Action,
	params map[string]interface{},
	workflow *drv1alpha1.DRWorkflow,
	wfStatus *drv1alpha1.WorkflowExecutionStatus,
) (drv1alpha1.ActionStatus, bool, error) {
	totalActions := len(workflow.Spec.Actions)
	completedSoFar := len(wfStatus.ActionStatuses)
	klog.Infof("Executing action %d/%d: %s (type=%s)", completedSoFar+1, totalActions, action.Name, action.Type)
	wfStatus.CurrentAction = action.Name
	wfStatus.Progress = fmt.Sprintf("%d/%d actions completed", completedSoFar, totalActions)

	shouldExecute, skipReason, whenErr := shouldExecuteActionByWhen(action.When, params)
	if whenErr != nil {
		as := drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseFailed,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: fmt.Sprintf("invalid when expression %q: %v", action.When, whenErr),
		}
		if isFailFast(workflow.Spec.FailurePolicy) {
			wfStatus.Phase = drv1alpha1.PhaseFailed
			wfStatus.Message = fmt.Sprintf("Action %s failed: invalid when expression", action.Name)
			return as, true, fmt.Errorf("invalid when on action %s: %w", action.Name, whenErr)
		}
		return as, false, nil
	}
	if !shouldExecute {
		klog.Infof("Skipping action %s due to when condition: %s", action.Name, skipReason)
		return drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseSkipped,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: skipReason,
		}, false, nil
	}

	actionExecutor, err := e.registry.GetExecutor(action.Type)
	if err != nil {
		klog.Errorf("Failed to get executor for action %s (type=%s): %v", action.Name, action.Type, err)
		as := drv1alpha1.ActionStatus{
			Name: action.Name, Phase: drv1alpha1.PhaseFailed,
			StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
			Message: fmt.Sprintf("Executor not found: %v", err),
		}
		if isFailFast(workflow.Spec.FailurePolicy) {
			wfStatus.Phase = drv1alpha1.PhaseFailed
			wfStatus.Message = fmt.Sprintf("Action %s failed: %v", action.Name, err)
			return as, true, err
		}
		return as, false, nil
	}

	actionStatus, err := actionExecutor.Execute(ctx, &action, params)
	if err != nil {
		klog.Errorf("Action %s failed: %v", action.Name, err)
		actionStatus.Phase = drv1alpha1.PhaseFailed
		actionStatus.Message = err.Error()
	}

	if actionStatus.Phase == drv1alpha1.PhaseFailed {
		if isFailFast(workflow.Spec.FailurePolicy) {
			wfStatus.Phase = drv1alpha1.PhaseFailed
			wfStatus.Message = fmt.Sprintf("Action %s failed: %s", action.Name, actionStatus.Message)
			klog.Warningf("Workflow %s/%s failed at action %s (FailFast)", workflow.Namespace, workflow.Name, action.Name)
			return *actionStatus, true, fmt.Errorf("action %s failed: %s", action.Name, actionStatus.Message)
		}
		klog.V(4).Infof("Action %s failed but continuing due to FailurePolicy=Continue", action.Name)
	}

	return *actionStatus, false, nil
}

// executePerClusterBatch executes a batch of PerCluster actions concurrently across target clusters.
// The executor resolves target clusters directly from subscription selectors and creates only child
// Subscriptions, avoiding parent subscriptions that duplicate Clusternet distribution records.
// NOCC:tosa/fn_length(设计如此)
func (e *NativeWorkflowExecutor) executePerClusterBatch(
	ctx context.Context,
	actions []drv1alpha1.Action,
	params map[string]interface{},
	failurePolicy string,
) ([]drv1alpha1.ActionStatus, error) {
	if len(actions) == 0 {
		return nil, nil
	}

	batchStartTime := metav1.Now()
	precheckedStatuses, executableActions, err := precheckPerClusterActions(actions, params, &batchStartTime)
	if err != nil {
		return precheckedStatuses, err
	}
	if len(executableActions) == 0 {
		return precheckedStatuses, nil
	}

	for _, action := range executableActions {
		if action.Subscription == nil {
			return nil, fmt.Errorf("PerCluster action %s has nil Subscription spec", action.Name)
		}
	}

	subExec, err := e.getSubscriptionExecutor()
	if err != nil {
		klog.Warningf("PerCluster batch: %v, falling back to Global", err)
		executableStatuses, fallbackErr := e.executeGlobalBatchFallback(ctx, executableActions, params, failurePolicy)
		return mergePerClusterBatchStatuses(actions, precheckedStatuses, executableStatuses), fallbackErr
	}

	actionTargets, bindings, err := e.collectPerClusterBindings(ctx, subExec, executableActions)
	if err != nil {
		return mergePerClusterBatchStatuses(actions, precheckedStatuses, nil), err
	}

	klog.Infof("PerCluster batch: %d actions × %d clusters", len(executableActions), len(bindings))

	batchCtx, batchCancel := context.WithCancel(ctx)
	defer batchCancel()

	statusMap := make(map[string][]drv1alpha1.ClusterActionStatus)
	childRefMap := make(map[string][]corev1.ObjectReference)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, binding := range bindings {
		wg.Add(1)
		go func(clusterBinding string) { // NOCC:tosa/fn_length(设计如此)
			defer wg.Done()
			e.executeClusterActions(
				batchCtx, subExec, executableActions, actionTargets, clusterBinding, params,
				failurePolicy, &batchStartTime, &mu, statusMap, childRefMap, batchCancel,
			)
		}(binding)
	}

	wg.Wait()

	executableStatuses, buildErr := e.buildPerClusterResult(executableActions, statusMap, childRefMap, &batchStartTime, failurePolicy)
	if cleanupErr := e.applyPerClusterHookCleanup(ctx, subExec, executableActions, executableStatuses); cleanupErr != nil {
		if buildErr != nil {
			return mergePerClusterBatchStatuses(actions, precheckedStatuses, executableStatuses),
				fmt.Errorf("%v; hook cleanup failed: %w", buildErr, cleanupErr)
		}
		return mergePerClusterBatchStatuses(actions, precheckedStatuses, executableStatuses), cleanupErr
	}
	return mergePerClusterBatchStatuses(actions, precheckedStatuses, executableStatuses), buildErr
}

// precheckPerClusterActions evaluates when expressions before any cluster fan-out.
// It returns pre-generated skipped/failed statuses plus actions that should really execute.
func precheckPerClusterActions(
	actions []drv1alpha1.Action,
	params map[string]interface{},
	batchStartTime *metav1.Time,
) ([]drv1alpha1.ActionStatus, []drv1alpha1.Action, error) {
	precheckedStatuses := make([]drv1alpha1.ActionStatus, 0, len(actions))
	executableActions := make([]drv1alpha1.Action, 0, len(actions))

	for _, action := range actions {
		shouldExecute, skipReason, whenErr := shouldExecuteActionByWhen(action.When, params)
		if whenErr != nil {
			as := drv1alpha1.ActionStatus{
				Name:           action.Name,
				Phase:          drv1alpha1.PhaseFailed,
				StartTime:      batchStartTime,
				CompletionTime: timeNowPtr(),
				Message:        fmt.Sprintf("invalid when expression %q: %v", action.When, whenErr),
			}
			precheckedStatuses = append(precheckedStatuses, as)
			return precheckedStatuses, nil, fmt.Errorf("invalid when on action %s: %w", action.Name, whenErr)
		}
		if !shouldExecute {
			precheckedStatuses = append(precheckedStatuses, drv1alpha1.ActionStatus{
				Name:           action.Name,
				Phase:          drv1alpha1.PhaseSkipped,
				StartTime:      batchStartTime,
				CompletionTime: timeNowPtr(),
				Message:        skipReason,
			})
			continue
		}
		executableActions = append(executableActions, action)
	}

	return precheckedStatuses, executableActions, nil
}

// mergePerClusterBatchStatuses restores action order and combines precheck/execution results.
func mergePerClusterBatchStatuses(
	originalActions []drv1alpha1.Action,
	precheckedStatuses []drv1alpha1.ActionStatus,
	executableStatuses []drv1alpha1.ActionStatus,
) []drv1alpha1.ActionStatus {
	mergedByName := make(map[string]drv1alpha1.ActionStatus, len(precheckedStatuses)+len(executableStatuses))
	for _, status := range precheckedStatuses {
		mergedByName[status.Name] = status
	}
	for _, status := range executableStatuses {
		mergedByName[status.Name] = status
	}

	merged := make([]drv1alpha1.ActionStatus, 0, len(originalActions))
	for _, action := range originalActions {
		if status, ok := mergedByName[action.Name]; ok {
			merged = append(merged, status)
		}
	}
	return merged
}

type perClusterActionTargets map[string]map[string]struct{}

// getSubscriptionExecutor fetches the typed Subscription executor from registry.
func (e *NativeWorkflowExecutor) getSubscriptionExecutor() (*SubscriptionActionExecutor, error) {
	executor, err := e.registry.GetExecutor(drv1alpha1.ActionTypeSubscription)
	if err != nil {
		return nil, fmt.Errorf("get Subscription executor: %w", err)
	}
	subExec, ok := executor.(*SubscriptionActionExecutor)
	if !ok {
		return nil, fmt.Errorf("executor is not SubscriptionActionExecutor")
	}
	return subExec, nil
}

// collectPerClusterBindings resolves each action's cluster scope and returns:
// 1) per-action target set, 2) union cluster list used for goroutine fan-out.
func (e *NativeWorkflowExecutor) collectPerClusterBindings(
	ctx context.Context,
	subExec *SubscriptionActionExecutor,
	actions []drv1alpha1.Action,
) (perClusterActionTargets, []string, error) {
	targets := make(perClusterActionTargets, len(actions))
	seen := make(map[string]struct{})
	union := make([]string, 0)

	for _, action := range actions {
		bindings, err := subExec.resolveTargetClusters(ctx, action.Subscription.Spec)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve target clusters for PerCluster action %s: %w", action.Name, err)
		}

		targetSet := make(map[string]struct{}, len(bindings))
		for _, binding := range bindings {
			targetSet[binding] = struct{}{}
			if _, exists := seen[binding]; exists {
				continue
			}
			seen[binding] = struct{}{}
			union = append(union, binding)
		}
		targets[action.Name] = targetSet
	}

	return targets, union, nil
}

// executeClusterActions runs all batch actions sequentially for a single cluster.
// NOCC:tosa/fn_length(设计如此)
func (e *NativeWorkflowExecutor) executeClusterActions(
	ctx context.Context,
	subExec *SubscriptionActionExecutor,
	actions []drv1alpha1.Action,
	actionTargets perClusterActionTargets,
	clusterBinding string,
	params map[string]interface{},
	failurePolicy string,
	batchStartTime *metav1.Time,
	mu *sync.Mutex,
	statusMap map[string][]drv1alpha1.ClusterActionStatus,
	childRefMap map[string][]corev1.ObjectReference,
	cancelFn context.CancelFunc,
) {
	clusterID := clusterBinding
	if clusterNS, clusterName, err := parseBindingCluster(clusterBinding); err == nil {
		clusterID = clusterNS + "/" + clusterName
	}

	for _, action := range actions {
		if ctx.Err() != nil {
			cs := drv1alpha1.ClusterActionStatus{
				Cluster: clusterBinding, ClusterID: clusterID,
				Phase:     drv1alpha1.PhaseFailed,
				StartTime: batchStartTime, CompletionTime: timeNowPtr(),
				Message: "canceled",
			}
			mu.Lock()
			statusMap[action.Name] = append(statusMap[action.Name], cs)
			mu.Unlock()
			return
		}

		if !clusterTargeted(actionTargets, action.Name, clusterBinding) {
			cs := drv1alpha1.ClusterActionStatus{
				Cluster: clusterBinding, ClusterID: clusterID,
				Phase:     drv1alpha1.PhaseSkipped,
				StartTime: timeNowPtr(), CompletionTime: timeNowPtr(),
				Message: "cluster not targeted by action",
			}
			mu.Lock()
			statusMap[action.Name] = append(statusMap[action.Name], cs)
			mu.Unlock()
			continue
		}

		shouldExec, skipReason, whenErr := shouldExecuteActionByWhen(action.When, params)
		if whenErr != nil {
			cs := drv1alpha1.ClusterActionStatus{
				Cluster: clusterBinding, ClusterID: clusterID,
				Phase:     drv1alpha1.PhaseFailed,
				StartTime: timeNowPtr(), CompletionTime: timeNowPtr(),
				Message: fmt.Sprintf("invalid when expression: %v", whenErr),
			}
			mu.Lock()
			statusMap[action.Name] = append(statusMap[action.Name], cs)
			mu.Unlock()
			break
		}
		if !shouldExec {
			cs := drv1alpha1.ClusterActionStatus{
				Cluster: clusterBinding, ClusterID: clusterID,
				Phase:     drv1alpha1.PhaseSkipped,
				StartTime: timeNowPtr(), CompletionTime: timeNowPtr(),
				Message: skipReason,
			}
			mu.Lock()
			statusMap[action.Name] = append(statusMap[action.Name], cs)
			mu.Unlock()
			continue
		}

		cs, childRef, _ := subExec.ExecuteForCluster(ctx, &action, clusterBinding, params)
		if cs == nil {
			cs = &drv1alpha1.ClusterActionStatus{
				Cluster: clusterBinding, Phase: drv1alpha1.PhaseFailed,
				StartTime: timeNowPtr(), CompletionTime: timeNowPtr(),
				Message: "ExecuteForCluster returned nil",
			}
		}

		mu.Lock()
		statusMap[action.Name] = append(statusMap[action.Name], *cs)
		if childRef != nil {
			childRefMap[action.Name] = append(childRefMap[action.Name], *childRef)
		}
		mu.Unlock()

		if cs.Phase == drv1alpha1.PhaseFailed {
			if isFailFast(failurePolicy) {
				cancelFn()
			}
			break
		}
	}
}

// clusterTargeted reports whether a cluster should run a specific action.
// Empty actionTargets means no per-action filtering (all clusters allowed).
func clusterTargeted(actionTargets perClusterActionTargets, actionName, clusterBinding string) bool {
	if len(actionTargets) == 0 {
		return true
	}
	targets, ok := actionTargets[actionName]
	if !ok || len(targets) == 0 {
		return false
	}
	_, targeted := targets[clusterBinding]
	return targeted
}

// buildPerClusterResult aggregates per-cluster statuses into ActionStatuses.
func (e *NativeWorkflowExecutor) buildPerClusterResult(
	actions []drv1alpha1.Action,
	statusMap map[string][]drv1alpha1.ClusterActionStatus,
	childRefMap map[string][]corev1.ObjectReference,
	batchStartTime *metav1.Time,
	failurePolicy string,
) ([]drv1alpha1.ActionStatus, error) {
	completionTime := metav1.Now()
	var result []drv1alpha1.ActionStatus
	for _, action := range actions {
		clusterStatuses := statusMap[action.Name]
		phase := aggregateClusterStatuses(clusterStatuses)
		as := drv1alpha1.ActionStatus{
			Name:            action.Name,
			Phase:           phase,
			StartTime:       batchStartTime,
			CompletionTime:  &completionTime,
			ClusterStatuses: clusterStatuses,
		}
		if childRefs := childRefMap[action.Name]; len(childRefs) > 0 {
			as.Outputs = &drv1alpha1.ActionOutputs{
				SubscriptionRefs: append([]corev1.ObjectReference(nil), childRefs...),
			}
		}
		switch phase {
		case drv1alpha1.PhaseFailed:
			as.Message = "one or more clusters failed"
		case drv1alpha1.PhaseSucceeded:
			as.Message = fmt.Sprintf("all %d clusters succeeded", len(clusterStatuses))
		case drv1alpha1.PhaseSkipped:
			as.Message = "all clusters skipped"
		}
		result = append(result, as)

		if phase == drv1alpha1.PhaseFailed && isFailFast(failurePolicy) {
			return result, fmt.Errorf("PerCluster action %s failed", action.Name)
		}
	}
	return result, nil
}

// timeNowPtr returns a fresh metav1.Now pointer for status timestamps.
func timeNowPtr() *metav1.Time {
	t := metav1.Now()
	return &t
}

// applyPerClusterHookCleanup executes post-cleanup policy for each child subscription reference.
// Any cleanup failure upgrades the corresponding action status to Failed.
func (e *NativeWorkflowExecutor) applyPerClusterHookCleanup(
	ctx context.Context,
	subExec *SubscriptionActionExecutor,
	actions []drv1alpha1.Action,
	statuses []drv1alpha1.ActionStatus,
) error {
	actionByName := make(map[string]drv1alpha1.Action, len(actions))
	for _, action := range actions {
		actionByName[action.Name] = action
	}

	for i := range statuses {
		status := &statuses[i]
		action, ok := actionByName[status.Name]
		if !ok || action.HookCleanup == nil || status.Outputs == nil {
			continue
		}

		refs := make([]corev1.ObjectReference, 0, len(status.Outputs.SubscriptionRefs)+1)
		refs = append(refs, status.Outputs.SubscriptionRefs...)
		if len(refs) == 0 && status.Outputs.SubscriptionRef != nil {
			refs = append(refs, *status.Outputs.SubscriptionRef)
		}
		for _, ref := range refs {
			if err := subExec.applyHookPostCleanup(ctx, &action, ref.Namespace, ref.Name, status.Phase); err != nil {
				status.Phase = drv1alpha1.PhaseFailed
				status.Message = fmt.Sprintf("hook cleanup failed: %v", err)
				status.CompletionTime = timeNowPtr()
				return fmt.Errorf("PerCluster hook cleanup failed for %s: %w", action.Name, err)
			}
		}
	}

	return nil
}

// executeGlobalBatchFallback executes actions sequentially as Global when per-cluster split is not possible.
func (e *NativeWorkflowExecutor) executeGlobalBatchFallback(
	ctx context.Context,
	actions []drv1alpha1.Action,
	params map[string]interface{},
	failurePolicy string,
) ([]drv1alpha1.ActionStatus, error) {
	var result []drv1alpha1.ActionStatus
	for _, action := range actions {
		actionExecutor, err := e.registry.GetExecutor(action.Type)
		if err != nil {
			as := drv1alpha1.ActionStatus{
				Name: action.Name, Phase: drv1alpha1.PhaseFailed,
				StartTime: &metav1.Time{Time: time.Now()}, CompletionTime: &metav1.Time{Time: time.Now()},
				Message: fmt.Sprintf("Executor not found: %v", err),
			}
			result = append(result, as)
			if isFailFast(failurePolicy) {
				return result, err
			}
			continue
		}

		as, err := actionExecutor.Execute(ctx, &action, params)
		if err != nil && as != nil {
			as.Phase = drv1alpha1.PhaseFailed
			as.Message = err.Error()
		}
		if as != nil && !isTerminalActionPhase(as.Phase) {
			originalPhase := as.Phase
			as.Phase = drv1alpha1.PhaseFailed
			as.CompletionTime = &metav1.Time{Time: time.Now()}
			as.Message = fmt.Sprintf("action %s did not reach terminal phase in fallback path: %s", action.Name, originalPhase)
		}
		if as != nil {
			result = append(result, *as)
		}
		if as != nil && as.Phase == drv1alpha1.PhaseFailed && isFailFast(failurePolicy) {
			return result, fmt.Errorf("action %s failed: %s", action.Name, as.Message)
		}
	}
	return result, nil
}

// isFailFast treats empty policy as FailFast for backward compatibility.
func isFailFast(policy string) bool {
	return policy == drv1alpha1.FailurePolicyFailFast || policy == ""
}

// updateWorkflowFinalPhase derives workflow final phase from all action phases.
// Priority: Failed > Running > Pending > Skipped(all) > Succeeded.
func updateWorkflowFinalPhase(status *drv1alpha1.WorkflowExecutionStatus, namespace, name string) {
	hasRunning := false
	hasPending := false
	allSkipped := len(status.ActionStatuses) > 0
	for _, as := range status.ActionStatuses {
		if as.Phase == drv1alpha1.PhaseFailed {
			status.Phase = drv1alpha1.PhaseFailed
			status.Message = "One or more actions failed"
			klog.Warningf("Workflow %s/%s completed with failures", namespace, name)
			return
		}
		if as.Phase == drv1alpha1.PhaseRunning {
			hasRunning = true
			allSkipped = false
			continue
		}
		if as.Phase == drv1alpha1.PhasePending {
			hasPending = true
			allSkipped = false
			continue
		}
		if as.Phase != drv1alpha1.PhaseSkipped {
			allSkipped = false
		}
	}
	if hasRunning {
		status.Phase = drv1alpha1.PhaseRunning
		status.Message = "One or more actions are still running"
		klog.Infof("Workflow %s/%s still running", namespace, name)
		return
	}
	if hasPending {
		status.Phase = drv1alpha1.PhasePending
		status.Message = "One or more actions are still pending"
		klog.Infof("Workflow %s/%s still pending", namespace, name)
		return
	}
	if allSkipped {
		status.Phase = drv1alpha1.PhaseSkipped
		status.Message = "All actions were skipped"
		klog.Infof("Workflow %s/%s skipped", namespace, name)
		return
	}
	status.Phase = drv1alpha1.PhaseSucceeded
	status.Message = "All actions completed successfully"
	klog.Infof("Workflow %s/%s succeeded", namespace, name)
}

// isTerminalActionPhase defines phases accepted by global fallback path.
func isTerminalActionPhase(phase string) bool {
	return phase == drv1alpha1.PhaseSucceeded ||
		phase == drv1alpha1.PhaseFailed ||
		phase == drv1alpha1.PhaseSkipped
}

// shouldExecuteActionByWhen evaluates restricted when expressions against params["mode"].
// It intentionally supports only a safe subset: mode/operation == value (optionally joined by ||).
func shouldExecuteActionByWhen(when string, params map[string]interface{}) (bool, string, error) {
	when = strings.TrimSpace(when)
	if when == "" {
		return true, "", nil
	}
	mode := ""
	if params != nil {
		if val, ok := params["mode"]; ok && val != nil {
			mode = strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", val)))
		} else if val, ok := params["operation"]; ok && val != nil {
			mode = strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", val)))
		}
	}
	// Backward compatibility: when mode is not provided, do not filter actions.
	if mode == "" {
		return true, "mode not provided, compatibility mode", nil
	}
	expectedModes := make([]string, 0, 2)
	for _, clause := range strings.Split(when, "||") {
		clause = strings.TrimSpace(clause)
		match := whenModePattern.FindStringSubmatch(clause)
		if len(match) != 2 {
			return false, "", fmt.Errorf("unsupported expression, only `mode == \"install|upgrade|delete|rollback\"` joined by `||` is allowed")
		}
		expected := strings.ToLower(strings.TrimSpace(match[1]))
		switch expected {
		case "install", "upgrade", "delete", modeRollback:
		default:
			return false, "", fmt.Errorf("unsupported mode value %q in when", expected)
		}
		expectedModes = append(expectedModes, expected)
		if mode == expected {
			return true, "", nil
		}
	}
	return false, fmt.Sprintf("when not matched: mode=%s, expected one of [%s]", mode, strings.Join(expectedModes, ",")), nil
}

// RevertWorkflow reverts a workflow by executing rollback actions in reverse order
func (e *NativeWorkflowExecutor) RevertWorkflow( //nolint:funlen // rollback flow intentionally spans multiple guarded steps
	ctx context.Context,
	workflow *drv1alpha1.DRWorkflow,
	workflowStatus *drv1alpha1.WorkflowExecutionStatus,
	params map[string]interface{},
) (*drv1alpha1.WorkflowExecutionStatus, error) {
	klog.Infof("Reverting workflow: %s/%s", workflow.Namespace, workflow.Name)

	// Create rollback workflow status object
	rollbackStatus := &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef:    workflowStatus.WorkflowRef,
		Phase:          "Running",
		StartTime:      &metav1.Time{Time: time.Now()},
		ActionStatuses: []drv1alpha1.ActionStatus{},
	}

	preRollbackHooks, postRollbackHooks := splitRollbackHooks(workflow.Spec.Actions)
	if len(preRollbackHooks) > 0 {
		preStatuses, err := e.executeRollbackHooks(ctx, preRollbackHooks, params)
		rollbackStatus.ActionStatuses = append(rollbackStatus.ActionStatuses, preStatuses...)
		if err != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = "Pre-rollback hook failed"
			return e.finalizeWorkflowStatus(rollbackStatus), err
		}
	}

	// Revert actions in reverse order
	succeededCount := 0
	totalCount := 0
	revertedStatuses := make([]drv1alpha1.ActionStatus, 0, len(workflowStatus.ActionStatuses))
	for i := len(workflowStatus.ActionStatuses) - 1; i >= 0; i-- {
		actionStatus := workflowStatus.ActionStatuses[i]
		totalCount++

		// Only revert succeeded actions
		if actionStatus.Phase != drv1alpha1.PhaseSucceeded {
			klog.V(4).Infof("Skipping revert for action %s (phase=%s)", actionStatus.Name, actionStatus.Phase)
			// Record skipped action in rollback status
			skippedStatus := drv1alpha1.ActionStatus{
				Name:    actionStatus.Name,
				Phase:   "Skipped",
				Message: fmt.Sprintf("Original action phase was %s, skipped rollback", actionStatus.Phase),
			}
			revertedStatuses = append([]drv1alpha1.ActionStatus{skippedStatus}, revertedStatuses...)
			continue
		}

		// Find action definition
		var action *drv1alpha1.Action
		for j := range workflow.Spec.Actions {
			if workflow.Spec.Actions[j].Name == actionStatus.Name {
				action = &workflow.Spec.Actions[j]
				break
			}
		}

		if action == nil {
			klog.Warningf("Action %s not found in workflow definition", actionStatus.Name)
			skippedStatus := drv1alpha1.ActionStatus{
				Name:    actionStatus.Name,
				Phase:   "Skipped",
				Message: "Action not found in workflow definition",
			}
			revertedStatuses = append([]drv1alpha1.ActionStatus{skippedStatus}, revertedStatuses...)
			continue
		}
		if isRollbackHookAction(action) {
			klog.V(4).Infof("Skipping rollback hook action %s during reverse rollback loop", action.Name)
			continue
		}

		klog.Infof("Reverting action: %s (type=%s)", action.Name, action.Type)

		// Get action executor
		actionExecutor, err := e.registry.GetExecutor(action.Type)
		if err != nil {
			klog.Errorf("Failed to get executor for action %s: %v", action.Name, err)
			failedStatus := drv1alpha1.ActionStatus{
				Name:           actionStatus.Name,
				Phase:          "Failed",
				Message:        fmt.Sprintf("Failed to get executor: %v", err),
				StartTime:      &metav1.Time{Time: time.Now()},
				CompletionTime: &metav1.Time{Time: time.Now()},
			}
			revertedStatuses = append([]drv1alpha1.ActionStatus{failedStatus}, revertedStatuses...)
			rollbackStatus.ActionStatuses = append(rollbackStatus.ActionStatuses, revertedStatuses...)
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to rollback action %s", action.Name)
			return e.finalizeWorkflowStatus(rollbackStatus), fmt.Errorf("failed to get executor for action %s: %w", action.Name, err)
		}

		// Execute rollback and get status
		actionRollbackStatus, err := actionExecutor.Rollback(ctx, action, &actionStatus, params)
		if err != nil {
			klog.Errorf("Failed to rollback action %s: %v", action.Name, err)
			// Add the failed action status to rollback status
			revertedStatuses = append([]drv1alpha1.ActionStatus{*actionRollbackStatus}, revertedStatuses...)
			rollbackStatus.ActionStatuses = append(rollbackStatus.ActionStatuses, revertedStatuses...)
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to rollback action %s", action.Name)
			return e.finalizeWorkflowStatus(rollbackStatus), fmt.Errorf("failed to rollback action %s: %w", action.Name, err)
		}

		// Add successful action status to rollback status (prepend to maintain reverse order in status)
		revertedStatuses = append([]drv1alpha1.ActionStatus{*actionRollbackStatus}, revertedStatuses...)
		if actionRollbackStatus.Phase == drv1alpha1.PhaseSucceeded {
			succeededCount++
		}
		klog.Infof("Action %s rolled back successfully", action.Name)
	}

	rollbackStatus.ActionStatuses = append(rollbackStatus.ActionStatuses, revertedStatuses...)

	if len(postRollbackHooks) > 0 {
		postStatuses, err := e.executeRollbackHooks(ctx, postRollbackHooks, params)
		rollbackStatus.ActionStatuses = append(rollbackStatus.ActionStatuses, postStatuses...)
		if err != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = "Post-rollback hook failed"
			return e.finalizeWorkflowStatus(rollbackStatus), err
		}
	}

	// Update progress and finalize
	rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
	rollbackStatus.Progress = fmt.Sprintf("%d/%d actions rolled back", succeededCount, totalCount)
	rollbackStatus.Message = "Workflow reverted successfully"

	klog.Infof("Workflow %s/%s reverted successfully", workflow.Namespace, workflow.Name)
	return e.finalizeWorkflowStatus(rollbackStatus), nil
}

func splitRollbackHooks(actions []drv1alpha1.Action) ([]drv1alpha1.Action, []drv1alpha1.Action) {
	var pre []drv1alpha1.Action
	var post []drv1alpha1.Action
	for i := range actions {
		switch actions[i].HookType {
		case "pre-rollback":
			pre = append(pre, actions[i])
		case "post-rollback":
			post = append(post, actions[i])
		}
	}
	return pre, post
}

func isRollbackHookAction(action *drv1alpha1.Action) bool {
	if action == nil {
		return false
	}
	return action.HookType == "pre-rollback" || action.HookType == "post-rollback"
}

func (e *NativeWorkflowExecutor) executeRollbackHooks(
	ctx context.Context,
	actions []drv1alpha1.Action,
	params map[string]interface{},
) ([]drv1alpha1.ActionStatus, error) {
	statuses := make([]drv1alpha1.ActionStatus, 0, len(actions))
	for i := range actions {
		action := &actions[i]
		shouldExec, skipReason, whenErr := shouldExecuteActionByWhen(action.When, params)
		if whenErr != nil {
			return statuses, fmt.Errorf("evaluating rollback hook %s when condition: %w", action.Name, whenErr)
		}
		if !shouldExec {
			statuses = append(statuses, drv1alpha1.ActionStatus{
				Name:           action.Name,
				Phase:          drv1alpha1.PhaseSkipped,
				Message:        skipReason,
				StartTime:      &metav1.Time{Time: time.Now()},
				CompletionTime: &metav1.Time{Time: time.Now()},
			})
			continue
		}

		actionExecutor, err := e.registry.GetExecutor(action.Type)
		if err != nil {
			return append(statuses, drv1alpha1.ActionStatus{
				Name:           action.Name,
				Phase:          drv1alpha1.PhaseFailed,
				Message:        fmt.Sprintf("Failed to get executor: %v", err),
				StartTime:      &metav1.Time{Time: time.Now()},
				CompletionTime: &metav1.Time{Time: time.Now()},
			}), fmt.Errorf("failed to get executor for rollback hook %s: %w", action.Name, err)
		}

		status, err := actionExecutor.Execute(ctx, action, params)
		if status != nil {
			statuses = append(statuses, *status)
		}
		if err != nil {
			if status == nil {
				statuses = append(statuses, drv1alpha1.ActionStatus{
					Name:           action.Name,
					Phase:          drv1alpha1.PhaseFailed,
					Message:        err.Error(),
					StartTime:      &metav1.Time{Time: time.Now()},
					CompletionTime: &metav1.Time{Time: time.Now()},
				})
			}
			return statuses, fmt.Errorf("rollback hook %s failed: %w", action.Name, err)
		}
		if status == nil {
			statuses = append(statuses, drv1alpha1.ActionStatus{
				Name:           action.Name,
				Phase:          drv1alpha1.PhaseSucceeded,
				Message:        "Rollback hook executed successfully",
				StartTime:      &metav1.Time{Time: time.Now()},
				CompletionTime: &metav1.Time{Time: time.Now()},
			})
		}
	}
	return statuses, nil
}

// finalizeWorkflowStatus finalizes workflow status with completion time and duration
func (e *NativeWorkflowExecutor) finalizeWorkflowStatus(status *drv1alpha1.WorkflowExecutionStatus) *drv1alpha1.WorkflowExecutionStatus {
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	if status.StartTime != nil {
		status.Duration = status.CompletionTime.Sub(status.StartTime.Time).String()
	}
	status.CurrentAction = ""

	// Update final progress
	completed := 0
	for _, as := range status.ActionStatuses {
		if as.Phase == drv1alpha1.PhaseSucceeded ||
			as.Phase == drv1alpha1.PhaseFailed ||
			as.Phase == drv1alpha1.PhaseSkipped {
			completed++
		}
	}
	status.Progress = fmt.Sprintf("%d/%d actions completed", completed, len(status.ActionStatuses))

	return status
}

// NativePlanExecutor implements PlanExecutor for native execution
type NativePlanExecutor struct {
	client           client.Client
	stageExecutor    StageExecutor
	workflowExecutor WorkflowExecutor
	dynamicClient    dynamic.Interface
	mapper           paramRESTMapper
}

// NewNativePlanExecutor creates a new NativePlanExecutor.
// dynamicClient and mapper are used for valueFrom.manifestRef parameter resolution;
// pass nil to disable valueFrom support (backward compatible).
func NewNativePlanExecutor(
	client client.Client,
	stageExecutor StageExecutor,
	workflowExecutor WorkflowExecutor,
	dynamicClient dynamic.Interface,
	mapper paramRESTMapper,
) *NativePlanExecutor {
	return &NativePlanExecutor{
		client:           client,
		stageExecutor:    stageExecutor,
		workflowExecutor: workflowExecutor,
		dynamicClient:    dynamicClient,
		mapper:           mapper,
	}
}

// ExecutePlan executes a DR plan
func (e *NativePlanExecutor) ExecutePlan(ctx context.Context, plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution) error {
	klog.Infof("Executing plan: %s/%s", plan.Namespace, plan.Name)

	// Initialize execution status
	if execution.Status.StageStatuses == nil {
		execution.Status.StageStatuses = make([]drv1alpha1.StageStatus, 0, len(plan.Spec.Stages))
	}
	if execution.Status.Summary == nil {
		execution.Status.Summary = &drv1alpha1.ExecutionSummary{
			TotalStages: len(plan.Spec.Stages),
		}
	}

	globalParams, executionOverrides, err := e.resolveExecutionParams(ctx, plan, execution)
	if err != nil {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = err.Error()
		return e.updateExecutionStatus(ctx, execution, nil)
	}

	// Execute stages based on dependencies
	executedStages := make(map[string]bool)

	for {
		// Find stages that can be executed (dependencies met)
		readyStages := e.findReadyStages(plan.Spec.Stages, executedStages)
		if len(readyStages) == 0 {
			break // All stages executed or no more can be executed
		}

		// Execute ready stages
		for _, stage := range readyStages {
			klog.Infof("Executing stage: %s", stage.Name)

			stageStatus, err := e.stageExecutor.ExecuteStage(ctx, plan, &stage, globalParams, executionOverrides)
			if err != nil {
				klog.Errorf("Stage %s failed: %v", stage.Name, err)
				execution.Status.Phase = drv1alpha1.PhaseFailed
				execution.Status.Message = fmt.Sprintf("Stage %s failed: %v", stage.Name, err)
				return e.updateExecutionStatus(ctx, execution, stageStatus)
			}

			// Update stage status in execution
			e.updateStageStatusInExecution(execution, stageStatus)

			// Check if stage failed
			if stageStatus.Phase == drv1alpha1.PhaseFailed {
				if plan.Spec.FailurePolicy == "Stop" || plan.Spec.FailurePolicy == "" {
					execution.Status.Phase = drv1alpha1.PhaseFailed
					execution.Status.Message = fmt.Sprintf("Stage %s failed", stage.Name)
					return e.updateExecutionStatus(ctx, execution, nil)
				}
			}

			executedStages[stage.Name] = true
		}
	}

	// Check overall success
	allSucceeded := true
	for _, stageStatus := range execution.Status.StageStatuses {
		if stageStatus.Phase != drv1alpha1.PhaseSucceeded {
			allSucceeded = false
			break
		}
	}

	if allSucceeded {
		execution.Status.Phase = drv1alpha1.PhaseSucceeded
		execution.Status.Message = "All stages completed successfully"
		if shouldCleanupHistoricalSubscriptions(execution) {
			cleaned, err := e.cleanupHistoricalSubscriptionOutputs(ctx, plan, execution)
			if err != nil {
				execution.Status.Phase = drv1alpha1.PhaseFailed
				execution.Status.Message = fmt.Sprintf("delete cleanup failed: %v", err)
				return e.updateExecutionStatus(ctx, execution, nil)
			}
			if cleaned > 0 {
				execution.Status.Message = fmt.Sprintf("All stages completed successfully (cleaned %d historical subscriptions)", cleaned)
			}
		}
	} else {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = "One or more stages failed"
	}

	return e.updateExecutionStatus(ctx, execution, nil)
}


// shouldCleanupHistoricalSubscriptions returns true only for Execute(Delete mode).
// It gates historical subscription cleanup to avoid side-effects in other modes.
// NOCC:tosa/fn_length(设计如此)
func shouldCleanupHistoricalSubscriptions(execution *drv1alpha1.DRPlanExecution) bool {
	if execution == nil {
		return false
	}
	if execution.Spec.OperationType != drv1alpha1.OperationTypeExecute {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(execution.Spec.Mode), "delete")
}


// cleanupHistoricalSubscriptionOutputs deletes subscriptions referenced by prior executions
// of the same plan, and returns number of successfully deleted resources.
// NOCC:tosa/fn_length(设计如此)
func (e *NativePlanExecutor) cleanupHistoricalSubscriptionOutputs(
	ctx context.Context,
	plan *drv1alpha1.DRPlan,
	execution *drv1alpha1.DRPlanExecution,
) (int, error) {
	refs, err := e.collectHistoricalSubscriptionRefs(ctx, plan, execution)
	if err != nil {
		return 0, err
	}
	if len(refs) == 0 {
		return 0, nil
	}

	deleted := 0
	for _, ref := range refs {
		sub := newSubscriptionObjectRef(ref)
		if err := e.client.Delete(ctx, sub); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return deleted, fmt.Errorf("delete Subscription %s/%s: %w", sub.GetNamespace(), sub.GetName(), err)
		}
		deleted++
	}
	return deleted, nil
}

// collectHistoricalSubscriptionRefs collects unique subscription refs from plan execution history,
// excluding current execution and non-Execute records.
func (e *NativePlanExecutor) collectHistoricalSubscriptionRefs(
	ctx context.Context,
	plan *drv1alpha1.DRPlan,
	execution *drv1alpha1.DRPlanExecution,
) ([]corev1.ObjectReference, error) {
	if plan == nil {
		return nil, nil
	}

	refsByKey := make(map[string]corev1.ObjectReference)
	for _, record := range plan.Status.ExecutionHistory {
		if record.OperationType != drv1alpha1.OperationTypeExecute {
			continue
		}
		recordName := strings.TrimSpace(record.Name)
		recordNamespace := strings.TrimSpace(record.Namespace)
		if recordNamespace == "" {
			recordNamespace = execution.Namespace
		}
		if recordName == execution.Name && recordNamespace == execution.Namespace {
			continue
		}
		if recordName == "" {
			continue
		}

		historyExec := &drv1alpha1.DRPlanExecution{}
		key := client.ObjectKey{Name: recordName, Namespace: recordNamespace}
		if err := e.client.Get(ctx, key, historyExec); err != nil {
			if apierrors.IsNotFound(err) {
				klog.V(4).Infof("Execution history record %s/%s not found, skip subscription cleanup collection", key.Namespace, key.Name)
				continue
			}
			return nil, fmt.Errorf("get historical execution %s/%s: %w", key.Namespace, key.Name, err)
		}
		if historyExec.Spec.PlanRef != plan.Name {
			continue
		}

		for _, ref := range extractSubscriptionRefs(historyExec, execution.Namespace) {
			refsByKey[fmt.Sprintf("%s/%s", ref.Namespace, ref.Name)] = ref
		}
	}

	if len(refsByKey) == 0 {
		return nil, nil
	}
	refs := make([]corev1.ObjectReference, 0, len(refsByKey))
	for _, ref := range refsByKey {
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Namespace != refs[j].Namespace {
			return refs[i].Namespace < refs[j].Namespace
		}
		return refs[i].Name < refs[j].Name
	})
	return refs, nil
}

// extractSubscriptionRefs scans stage/workflow/action outputs and normalizes object references.
// Empty namespaces are backfilled with defaultNamespace.
func extractSubscriptionRefs(execution *drv1alpha1.DRPlanExecution, defaultNamespace string) []corev1.ObjectReference {
	if execution == nil {
		return nil
	}
	refs := make([]corev1.ObjectReference, 0)
	addRef := func(ref corev1.ObjectReference) {
		name := strings.TrimSpace(ref.Name)
		if name == "" {
			return
		}
		ref.Name = name
		ref.Namespace = strings.TrimSpace(ref.Namespace)
		if ref.Namespace == "" {
			ref.Namespace = defaultNamespace
		}
		refs = append(refs, ref)
	}

	for _, stageStatus := range execution.Status.StageStatuses {
		for _, wfStatus := range stageStatus.WorkflowExecutions {
			for _, actionStatus := range wfStatus.ActionStatuses {
				if actionStatus.Outputs == nil {
					continue
				}
				for _, ref := range actionStatus.Outputs.SubscriptionRefs {
					addRef(ref)
				}
				if actionStatus.Outputs.SubscriptionRef != nil {
					addRef(*actionStatus.Outputs.SubscriptionRef)
				}
			}
		}
	}
	return refs
}

// newSubscriptionObjectRef converts an ObjectReference to unstructured Subscription object.
func newSubscriptionObjectRef(ref corev1.ObjectReference) *unstructured.Unstructured {
	sub := &unstructured.Unstructured{}
	sub.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	sub.SetNamespace(ref.Namespace)
	sub.SetName(ref.Name)
	return sub
}

// resolveExecutionParams resolves plan-level globals first, then execution-level overrides.
// Reserved key "mode" is sourced from execution.Spec.Mode and not overridden by Params.
func (e *NativePlanExecutor) resolveExecutionParams(
	ctx context.Context,
	plan *drv1alpha1.DRPlan,
	execution *drv1alpha1.DRPlanExecution,
) (map[string]interface{}, map[string]interface{}, error) {
	globalParams := make(map[string]interface{})
	if len(plan.Spec.GlobalParams) > 0 {
		planParams, err := resolveParams(ctx, e.dynamicClient, e.mapper, plan.Spec.GlobalParams, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("resolving plan global params: %v", err)
		}
		for k, v := range planParams {
			globalParams[k] = v
		}
	}

	if mode := strings.TrimSpace(execution.Spec.Mode); mode != "" {
		globalParams["mode"] = strings.ToLower(mode)
	}

	executionOverrides := make(map[string]interface{})
	if len(execution.Spec.Params) > 0 {
		resolved, err := resolveParams(ctx, e.dynamicClient, e.mapper, execution.Spec.Params, globalParams)
		if err != nil {
			return nil, nil, fmt.Errorf("resolving execution params: %v", err)
		}
		for k, v := range resolved {
			if k == "mode" {
				continue
			}
			executionOverrides[k] = v
		}
	}

	return globalParams, executionOverrides, nil
}

// RevertPlan reverts a DR plan
func (e *NativePlanExecutor) RevertPlan(ctx context.Context, plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution) error { //nolint:funlen,gocyclo
	klog.Infof("Reverting plan: %s/%s", plan.Namespace, plan.Name)

	targetExecution, err := e.validateAndFetchRevertTarget(ctx, plan, execution)
	if err != nil {
		return e.updateExecutionStatus(ctx, execution, nil)
	}

	klog.Infof("Reverting based on target execution %s/%s with %d stages",
		targetExecution.Namespace, targetExecution.Name, len(targetExecution.Status.StageStatuses))

	globalParams, executionOverrides, err := e.resolveExecutionParams(ctx, plan, targetExecution)
	if err != nil {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = err.Error()
		return e.updateExecutionStatus(ctx, execution, nil)
	}
	globalParams["mode"] = modeRollback
	if len(execution.Spec.Params) > 0 {
		revertOverrides, err := resolveParams(ctx, e.dynamicClient, e.mapper, execution.Spec.Params, globalParams)
		if err != nil {
			execution.Status.Phase = drv1alpha1.PhaseFailed
			execution.Status.Message = fmt.Sprintf("resolving revert execution params: %v", err)
			return e.updateExecutionStatus(ctx, execution, nil)
		}
		for k, v := range revertOverrides {
			if k == "mode" {
				continue
			}
			executionOverrides[k] = v
		}
	}

	e.initializeRevertExecutionStatus(execution, len(targetExecution.Status.StageStatuses))

	// Revert stages in reverse order from the target execution
	succeededStages := 0
	skippedStages := 0
	totalActions := 0

	for i := len(targetExecution.Status.StageStatuses) - 1; i >= 0; i-- {
		originalStageStatus := targetExecution.Status.StageStatuses[i]

		// Skip stages that never ran or were intentionally bypassed.
		// Failed stages may contain partially-succeeded workflows that have side-effects,
		// so they must go through RevertStage which handles workflow-level skipping internally.
		isNonTerminalOrSkipped := originalStageStatus.Phase == drv1alpha1.PhasePending ||
			originalStageStatus.Phase == drv1alpha1.PhaseRunning ||
			originalStageStatus.Phase == drv1alpha1.PhaseSkipped ||
			originalStageStatus.Phase == drv1alpha1.PhaseCancelled
		if isNonTerminalOrSkipped {
			klog.V(4).Infof("Skipping revert for stage %s (phase=%s)", originalStageStatus.Name, originalStageStatus.Phase)
			skippedStageStatus := drv1alpha1.StageStatus{
				Name:    originalStageStatus.Name,
				Phase:   drv1alpha1.PhaseSkipped,
				Message: fmt.Sprintf("Original stage phase was %s, skipped rollback", originalStageStatus.Phase),
			}
			e.updateStageStatusInExecution(execution, &skippedStageStatus)
			skippedStages++
			continue
		}

		// Find stage definition
		var stage *drv1alpha1.Stage
		for j := range plan.Spec.Stages {
			if plan.Spec.Stages[j].Name == originalStageStatus.Name {
				stage = &plan.Spec.Stages[j]
				break
			}
		}

		if stage == nil {
			klog.Warningf("Stage %s not found in plan definition", originalStageStatus.Name)
			skippedStageStatus := drv1alpha1.StageStatus{
				Name:    originalStageStatus.Name,
				Phase:   "Skipped",
				Message: "Stage not found in plan definition",
			}
			e.updateStageStatusInExecution(execution, &skippedStageStatus)
			skippedStages++
			continue
		}

		klog.Infof("Reverting stage: %s", stage.Name)

		// Execute revert and get rollback status
		rollbackStageStatus, err := e.stageExecutor.RevertStage(ctx, plan, stage, &originalStageStatus, globalParams, executionOverrides)

		// Update execution status with rollback stage status
		e.updateStageStatusInExecution(execution, rollbackStageStatus)

		if err != nil {
			klog.Errorf("Failed to revert stage %s: %v", stage.Name, err)
			execution.Status.Phase = drv1alpha1.PhaseFailed
			execution.Status.Message = fmt.Sprintf("Failed to revert stage %s: %v", stage.Name, err)
			return e.updateExecutionStatus(ctx, execution, nil)
		}

		// Check stage rollback result
		if rollbackStageStatus.Phase == drv1alpha1.PhaseFailed {
			execution.Status.Phase = drv1alpha1.PhaseFailed
			execution.Status.Message = fmt.Sprintf("Stage %s rollback failed", stage.Name)
			return e.updateExecutionStatus(ctx, execution, nil)
		}

		if rollbackStageStatus.Phase == drv1alpha1.PhaseSucceeded {
			succeededStages++
			// Count total actions rolled back
			for _, wfStatus := range rollbackStageStatus.WorkflowExecutions {
				for _, actionStatus := range wfStatus.ActionStatuses {
					if actionStatus.Phase == drv1alpha1.PhaseSucceeded {
						totalActions++
					}
				}
			}
		}
	}

	e.finalizeRevertSuccess(plan, execution, succeededStages, totalActions, skippedStages)
	return e.updateExecutionStatus(ctx, execution, nil)
}

// CancelExecution cancels an ongoing execution
func (e *NativePlanExecutor) CancelExecution(_ context.Context, execution *drv1alpha1.DRPlanExecution) error {
	klog.Infof("Canceling execution: %s/%s", execution.Namespace, execution.Name)

	// Mark all running/pending stages as canceled
	for i := range execution.Status.StageStatuses {
		if execution.Status.StageStatuses[i].Phase == drv1alpha1.PhaseRunning || execution.Status.StageStatuses[i].Phase == drv1alpha1.PhasePending {
			execution.Status.StageStatuses[i].Phase = drv1alpha1.PhaseSkipped
			execution.Status.StageStatuses[i].Message = "Canceled by user"
		}
	}

	return nil
}

// findReadyStages finds stages that can be executed (all dependencies met)
func (e *NativePlanExecutor) findReadyStages(stages []drv1alpha1.Stage, executedStages map[string]bool) []drv1alpha1.Stage {
	var ready []drv1alpha1.Stage

	for _, stage := range stages {
		// Skip already executed stages
		if executedStages[stage.Name] {
			continue
		}

		// Check if all dependencies are met
		allDepsMet := true
		for _, dep := range stage.DependsOn {
			if !executedStages[dep] {
				allDepsMet = false
				break
			}
		}

		if allDepsMet {
			ready = append(ready, stage)
		}
	}

	return ready
}

// updateStageStatusInExecution updates or appends stage status in execution
func (e *NativePlanExecutor) updateStageStatusInExecution(execution *drv1alpha1.DRPlanExecution, stageStatus *drv1alpha1.StageStatus) {
	found := false
	for i := range execution.Status.StageStatuses {
		if execution.Status.StageStatuses[i].Name == stageStatus.Name {
			execution.Status.StageStatuses[i] = *stageStatus
			found = true
			break
		}
	}
	if !found {
		execution.Status.StageStatuses = append(execution.Status.StageStatuses, *stageStatus)
	}

	// Update summary
	e.updateExecutionSummary(execution)
}

// validateAndFetchRevertTarget validates RevertExecutionRef and returns target execution.
// Only Execute operations in terminal phases (Succeeded/Failed) are accepted.
func (e *NativePlanExecutor) validateAndFetchRevertTarget(ctx context.Context, plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution) (*drv1alpha1.DRPlanExecution, error) {
	if execution.Spec.RevertExecutionRef == "" {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = "revertExecutionRef must be specified for Revert operation"
		klog.Errorf("RevertExecutionRef not specified for execution %s/%s", execution.Namespace, execution.Name)
		return nil, fmt.Errorf("revertExecutionRef is required")
	}

	targetExecutionName := execution.Spec.RevertExecutionRef
	klog.Infof("Reverting explicitly specified execution: %s", targetExecutionName)

	targetExecution := &drv1alpha1.DRPlanExecution{}
	targetExecKey := client.ObjectKey{
		Name:      targetExecutionName,
		Namespace: plan.Namespace,
	}
	if err := e.client.Get(ctx, targetExecKey, targetExecution); err != nil {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = fmt.Sprintf("Failed to get target execution %s: %v", targetExecutionName, err)
		klog.Errorf("Failed to get target execution %s/%s: %v", plan.Namespace, targetExecutionName, err)
		return nil, err
	}

	if targetExecution.Spec.OperationType != drv1alpha1.OperationTypeExecute {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = fmt.Sprintf("Cannot revert non-Execute operation (target is %s)", targetExecution.Spec.OperationType)
		klog.Errorf("Target execution %s/%s is not an Execute operation: %s", targetExecution.Namespace, targetExecution.Name, targetExecution.Spec.OperationType)
		return nil, fmt.Errorf("target must be an Execute operation")
	}

	// Allow revert for terminal phases only. A Failed execution may have partially
	// applied changes that need to be cleaned up, so both Succeeded and Failed are valid.
	isTerminal := targetExecution.Status.Phase == drv1alpha1.PhaseSucceeded ||
		targetExecution.Status.Phase == drv1alpha1.PhaseFailed
	if !isTerminal {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = fmt.Sprintf(
			"Cannot revert execution in phase %s (must be Succeeded or Failed)", targetExecution.Status.Phase)
		klog.Errorf("Target execution %s/%s is not in a terminal phase: %s",
			targetExecution.Namespace, targetExecution.Name, targetExecution.Status.Phase)
		return nil, fmt.Errorf("target must be in a terminal phase (Succeeded or Failed)")
	}

	if len(targetExecution.Status.StageStatuses) == 0 {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = "Target execution has no stage statuses to revert"
		klog.Warningf("Target execution %s/%s has no stage statuses", targetExecution.Namespace, targetExecution.Name)
		return nil, fmt.Errorf("target has no stage statuses")
	}

	return targetExecution, nil
}

// initializeRevertExecutionStatus ensures status slices/summary are initialized for revert flow.
func (e *NativePlanExecutor) initializeRevertExecutionStatus(execution *drv1alpha1.DRPlanExecution, totalStages int) {
	if execution.Status.StageStatuses == nil {
		execution.Status.StageStatuses = make([]drv1alpha1.StageStatus, 0, totalStages)
	}
	if execution.Status.Summary == nil {
		execution.Status.Summary = &drv1alpha1.ExecutionSummary{
			TotalStages: totalStages,
		}
	}
}

// finalizeRevertSuccess writes aggregated rollback counters into final success message.
func (e *NativePlanExecutor) finalizeRevertSuccess(plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution, succeededStages, totalActions, skippedStages int) {
	execution.Status.Phase = drv1alpha1.PhaseSucceeded
	execution.Status.Message = fmt.Sprintf(
		"Plan reverted successfully: %d stage(s) rolled back, %d action(s) rolled back, %d stage(s) skipped",
		succeededStages, totalActions, skippedStages)
	klog.Infof("Plan %s/%s reverted successfully: %d stages, %d actions, %d skipped",
		plan.Namespace, plan.Name, succeededStages, totalActions, skippedStages)
}

// updateExecutionSummary updates the execution summary
func (e *NativePlanExecutor) updateExecutionSummary(execution *drv1alpha1.DRPlanExecution) {
	if execution.Status.Summary == nil {
		execution.Status.Summary = &drv1alpha1.ExecutionSummary{}
	}

	summary := execution.Status.Summary
	summary.TotalStages = len(execution.Status.StageStatuses)
	summary.CompletedStages = 0
	summary.RunningStages = 0
	summary.PendingStages = 0
	summary.FailedStages = 0
	summary.TotalWorkflows = 0
	summary.CompletedWorkflows = 0
	summary.RunningWorkflows = 0
	summary.PendingWorkflows = 0
	summary.FailedWorkflows = 0

	for _, stageStatus := range execution.Status.StageStatuses {
		switch stageStatus.Phase {
		case "Succeeded":
			summary.CompletedStages++
		case "Running":
			summary.RunningStages++
		case "Pending":
			summary.PendingStages++
		case "Failed":
			summary.FailedStages++
		}

		// Count workflows
		summary.TotalWorkflows += len(stageStatus.WorkflowExecutions)
		for _, wfStatus := range stageStatus.WorkflowExecutions {
			switch wfStatus.Phase {
			case "Succeeded":
				summary.CompletedWorkflows++
			case "Running":
				summary.RunningWorkflows++
			case "Pending":
				summary.PendingWorkflows++
			case "Failed":
				summary.FailedWorkflows++
			}
		}
	}
}

// updateExecutionStatus updates execution status in API server
func (e *NativePlanExecutor) updateExecutionStatus(ctx context.Context, execution *drv1alpha1.DRPlanExecution, stageStatus *drv1alpha1.StageStatus) error {
	if stageStatus != nil {
		e.updateStageStatusInExecution(execution, stageStatus)
	}

	if err := e.client.Status().Update(ctx, execution); err != nil {
		klog.Errorf("Failed to update execution status: %v", err)
		return err
	}

	return nil
}
