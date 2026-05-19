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
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

type executionProgressContextKey struct{}
type stageProgressContextKey struct{}

type stageProgressContext struct {
	name      string
	parallel  bool
	dependsOn []string
	startTime *metav1.Time
}

type executionProgressRecorder struct {
	executor  *NativePlanExecutor
	execution *drv1alpha1.DRPlanExecution
	mu        sync.Mutex
}

func newExecutionProgressRecorder(
	executor *NativePlanExecutor,
	execution *drv1alpha1.DRPlanExecution,
) *executionProgressRecorder {
	return &executionProgressRecorder{executor: executor, execution: execution}
}

func (e *NativePlanExecutor) initializePendingStageStatuses(
	execution *drv1alpha1.DRPlanExecution,
	stages []drv1alpha1.Stage,
) {
	if len(execution.Status.StageStatuses) > 0 {
		return
	}
	execution.Status.StageStatuses = make([]drv1alpha1.StageStatus, 0, len(stages))
	for i := range stages {
		execution.Status.StageStatuses = append(execution.Status.StageStatuses, pendingStageStatus(stages[i]))
	}
	e.updateExecutionSummary(execution)
}

func pendingStageStatus(stage drv1alpha1.Stage) drv1alpha1.StageStatus {
	return drv1alpha1.StageStatus{
		Name:               stage.Name,
		Phase:              drv1alpha1.PhasePending,
		Parallel:           stage.Parallel,
		DependsOn:          append([]string(nil), stage.DependsOn...),
		WorkflowExecutions: pendingWorkflowStatuses(stage.Workflows),
	}
}

func runningStageStatus(stage drv1alpha1.Stage) drv1alpha1.StageStatus {
	status := pendingStageStatus(stage)
	now := metav1.Now()
	status.Phase = drv1alpha1.PhaseRunning
	status.StartTime = &now
	return status
}

func ptrToStageStatus(status drv1alpha1.StageStatus) *drv1alpha1.StageStatus {
	return &status
}

func pendingWorkflowStatuses(refs []drv1alpha1.WorkflowReference) []drv1alpha1.WorkflowExecutionStatus {
	statuses := make([]drv1alpha1.WorkflowExecutionStatus, 0, len(refs))
	for i := range refs {
		statuses = append(statuses, drv1alpha1.WorkflowExecutionStatus{
			WorkflowRef: refs[i].WorkflowRef,
			Phase:       drv1alpha1.PhasePending,
			Message:     "pending",
		})
	}
	return statuses
}

func withExecutionProgressRecorder(ctx context.Context, recorder *executionProgressRecorder) context.Context {
	if recorder == nil {
		return ctx
	}
	return context.WithValue(ctx, executionProgressContextKey{}, recorder)
}

func progressRecorderFrom(ctx context.Context) *executionProgressRecorder {
	recorder, _ := ctx.Value(executionProgressContextKey{}).(*executionProgressRecorder)
	return recorder
}

func withStageProgressContext(ctx context.Context, stageStatus *drv1alpha1.StageStatus) context.Context {
	if stageStatus == nil {
		return ctx
	}
	startTime := stageStatus.StartTime
	if startTime == nil {
		now := metav1.Now()
		startTime = &now
	}
	meta := stageProgressContext{
		name:      stageStatus.Name,
		parallel:  stageStatus.Parallel,
		dependsOn: append([]string(nil), stageStatus.DependsOn...),
		startTime: startTime,
	}
	return context.WithValue(ctx, stageProgressContextKey{}, meta)
}

func stageProgressFromContext(ctx context.Context) (stageProgressContext, bool) {
	meta, ok := ctx.Value(stageProgressContextKey{}).(stageProgressContext)
	return meta, ok
}

func (r *executionProgressRecorder) reportStage(ctx context.Context, stageStatus *drv1alpha1.StageStatus) {
	if r == nil || r.executor == nil || r.execution == nil || stageStatus == nil {
		return
	}
	r.mu.Lock()
	r.mergeStageLocked(stageStatus)
	r.ensureExecutionRunningLocked()
	r.executor.updateExecutionSummary(r.execution)
	err := r.executor.updateExecutionStatus(ctx, r.execution)
	r.mu.Unlock()
	if err != nil {
		klog.Warningf("Failed to persist stage progress %s: %v", stageStatus.Name, err)
	}
}

func (r *executionProgressRecorder) reportWorkflow(ctx context.Context, workflowStatus *drv1alpha1.WorkflowExecutionStatus) {
	if r == nil || r.executor == nil || r.execution == nil || workflowStatus == nil {
		return
	}
	meta, ok := stageProgressFromContext(ctx)
	if !ok || meta.name == "" {
		return
	}
	r.mu.Lock()
	stageStatus := r.ensureStageLocked(meta)
	if !isProgressTerminalPhase(stageStatus.Phase) {
		stageStatus.Phase = drv1alpha1.PhaseRunning
	}
	stageStatus.WorkflowExecutions = upsertWorkflowProgress(stageStatus.WorkflowExecutions, *workflowStatus.DeepCopy())
	r.ensureExecutionRunningLocked()
	r.executor.updateExecutionSummary(r.execution)
	err := r.executor.updateExecutionStatus(ctx, r.execution)
	r.mu.Unlock()
	if err != nil {
		klog.Warningf("Failed to persist workflow progress %s/%s: %v",
			workflowStatus.WorkflowRef.Namespace, workflowStatus.WorkflowRef.Name, err)
	}
}

func (r *executionProgressRecorder) mergeStageLocked(stageStatus *drv1alpha1.StageStatus) {
	stage := r.ensureStageLocked(stageProgressContext{
		name:      stageStatus.Name,
		parallel:  stageStatus.Parallel,
		dependsOn: stageStatus.DependsOn,
		startTime: stageStatus.StartTime,
	})
	stage.Phase = stageStatus.Phase
	stage.Message = stageStatus.Message
	stage.Parallel = stageStatus.Parallel
	stage.DependsOn = append([]string(nil), stageStatus.DependsOn...)
	if stageStatus.StartTime != nil {
		stage.StartTime = stageStatus.StartTime.DeepCopy()
	}
	if stageStatus.CompletionTime != nil {
		stage.CompletionTime = stageStatus.CompletionTime.DeepCopy()
	}
	stage.Duration = stageStatus.Duration
	for i := range stageStatus.WorkflowExecutions {
		stage.WorkflowExecutions = upsertWorkflowProgress(stage.WorkflowExecutions, stageStatus.WorkflowExecutions[i])
	}
}

func (r *executionProgressRecorder) ensureStageLocked(meta stageProgressContext) *drv1alpha1.StageStatus {
	for i := range r.execution.Status.StageStatuses {
		if r.execution.Status.StageStatuses[i].Name == meta.name {
			stage := &r.execution.Status.StageStatuses[i]
			stage.Parallel = meta.parallel
			stage.DependsOn = append([]string(nil), meta.dependsOn...)
			if stage.StartTime == nil && meta.startTime != nil {
				stage.StartTime = meta.startTime.DeepCopy()
			}
			return stage
		}
	}
	stage := drv1alpha1.StageStatus{
		Name:               meta.name,
		Phase:              drv1alpha1.PhasePending,
		Parallel:           meta.parallel,
		DependsOn:          append([]string(nil), meta.dependsOn...),
		StartTime:          meta.startTime,
		WorkflowExecutions: []drv1alpha1.WorkflowExecutionStatus{},
	}
	r.execution.Status.StageStatuses = append(r.execution.Status.StageStatuses, stage)
	return &r.execution.Status.StageStatuses[len(r.execution.Status.StageStatuses)-1]
}

func (r *executionProgressRecorder) ensureExecutionRunningLocked() {
	if r.execution.Status.Phase == "" || r.execution.Status.Phase == drv1alpha1.PhasePending {
		r.execution.Status.Phase = drv1alpha1.PhaseRunning
	}
}

func upsertWorkflowProgress(
	statuses []drv1alpha1.WorkflowExecutionStatus,
	status drv1alpha1.WorkflowExecutionStatus,
) []drv1alpha1.WorkflowExecutionStatus {
	for i := range statuses {
		if statuses[i].WorkflowRef.Name == status.WorkflowRef.Name &&
			statuses[i].WorkflowRef.Namespace == status.WorkflowRef.Namespace {
			statuses[i] = status
			return statuses
		}
	}
	return append(statuses, status)
}

func upsertActionProgress(statuses []drv1alpha1.ActionStatus, status drv1alpha1.ActionStatus) []drv1alpha1.ActionStatus {
	for i := range statuses {
		if statuses[i].Name == status.Name {
			statuses[i] = status
			return statuses
		}
	}
	return append(statuses, status)
}

func workflowProgressSnapshot(
	base *drv1alpha1.WorkflowExecutionStatus,
	actions []drv1alpha1.Action,
	runningStatuses ...drv1alpha1.ActionStatus,
) *drv1alpha1.WorkflowExecutionStatus {
	snapshot := base.DeepCopy()
	for i := range runningStatuses {
		snapshot.ActionStatuses = upsertActionProgress(snapshot.ActionStatuses, runningStatuses[i])
	}
	for i := range actions {
		if hasActionStatus(snapshot.ActionStatuses, actions[i].Name) {
			continue
		}
		snapshot.ActionStatuses = append(snapshot.ActionStatuses, drv1alpha1.ActionStatus{
			Name:    actions[i].Name,
			Phase:   drv1alpha1.PhasePending,
			Message: "pending",
		})
	}
	return snapshot
}

func runningActionStatus(action drv1alpha1.Action) drv1alpha1.ActionStatus {
	now := metav1.Time{Time: time.Now()}
	return drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     drv1alpha1.PhaseRunning,
		StartTime: &now,
		Message:   fmt.Sprintf("action %s running", action.Name),
	}
}

func runningActionStatuses(actions []drv1alpha1.Action) []drv1alpha1.ActionStatus {
	statuses := make([]drv1alpha1.ActionStatus, 0, len(actions))
	for i := range actions {
		statuses = append(statuses, runningActionStatus(actions[i]))
	}
	return statuses
}

func hasActionStatus(statuses []drv1alpha1.ActionStatus, name string) bool {
	for i := range statuses {
		if statuses[i].Name == name {
			return true
		}
	}
	return false
}

func isProgressTerminalPhase(phase string) bool {
	return phase == drv1alpha1.PhaseSucceeded ||
		phase == drv1alpha1.PhaseFailed ||
		phase == drv1alpha1.PhaseSkipped ||
		phase == drv1alpha1.PhaseCancelled
}
