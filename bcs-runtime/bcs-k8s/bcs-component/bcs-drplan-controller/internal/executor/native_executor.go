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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// NativeWorkflowExecutor implements WorkflowExecutor for native execution
type NativeWorkflowExecutor struct {
	client   client.Client
	registry ExecutorRegistry
}

// NewNativeWorkflowExecutor creates a new NativeWorkflowExecutor
func NewNativeWorkflowExecutor(client client.Client, registry ExecutorRegistry) *NativeWorkflowExecutor {
	return &NativeWorkflowExecutor{
		client:   client,
		registry: registry,
	}
}

// ExecuteWorkflow executes a workflow sequentially
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

	// Execute actions sequentially
	totalActions := len(workflow.Spec.Actions)
	for i, action := range workflow.Spec.Actions {
		klog.Infof("Executing action %d/%d: %s (type=%s)", i+1, totalActions, action.Name, action.Type)

		status.CurrentAction = action.Name
		status.Progress = fmt.Sprintf("%d/%d actions completed", i, totalActions)

		// Get action executor
		actionExecutor, err := e.registry.GetExecutor(action.Type)
		if err != nil {
			klog.Errorf("Failed to get executor for action %s (type=%s): %v", action.Name, action.Type, err)
			actionStatus := drv1alpha1.ActionStatus{
				Name:           action.Name,
				Phase:          "Failed",
				StartTime:      &metav1.Time{Time: time.Now()},
				CompletionTime: &metav1.Time{Time: time.Now()},
				Message:        fmt.Sprintf("Executor not found: %v", err),
			}
			status.ActionStatuses = append(status.ActionStatuses, actionStatus)

			// Handle failure policy
			if workflow.Spec.FailurePolicy == "FailFast" || workflow.Spec.FailurePolicy == "" {
				status.Phase = drv1alpha1.PhaseFailed
				status.Message = fmt.Sprintf("Action %s failed: %v", action.Name, err)
				return e.finalizeWorkflowStatus(status), err
			}
			// Continue with next action
			continue
		}

		// Execute action
		actionStatus, err := actionExecutor.Execute(ctx, &action, params)
		if err != nil {
			klog.Errorf("Action %s failed: %v", action.Name, err)
			actionStatus.Phase = drv1alpha1.PhaseFailed
			actionStatus.Message = err.Error()
		}

		// Append action status
		status.ActionStatuses = append(status.ActionStatuses, *actionStatus)

		// Check if action failed
		if actionStatus.Phase == drv1alpha1.PhaseFailed {
			// Handle failure policy
			if workflow.Spec.FailurePolicy == "FailFast" || workflow.Spec.FailurePolicy == "" {
				status.Phase = drv1alpha1.PhaseFailed
				status.Message = fmt.Sprintf("Action %s failed: %s", action.Name, actionStatus.Message)
				klog.Warningf("Workflow %s/%s failed at action %s (FailFast)", workflow.Namespace, workflow.Name, action.Name)
				return e.finalizeWorkflowStatus(status), fmt.Errorf("action %s failed: %s", action.Name, actionStatus.Message)
			}
			// Continue with next action if policy is Continue
			klog.V(4).Infof("Action %s failed but continuing due to FailurePolicy=Continue", action.Name)
		}
	}

	// Check overall success
	allSucceeded := true
	for _, as := range status.ActionStatuses {
		if as.Phase != drv1alpha1.PhaseSucceeded {
			allSucceeded = false
			break
		}
	}

	if allSucceeded {
		status.Phase = drv1alpha1.PhaseSucceeded
		status.Message = "All actions completed successfully"
		klog.Infof("Workflow %s/%s succeeded", workflow.Namespace, workflow.Name)
	} else {
		status.Phase = drv1alpha1.PhaseFailed
		status.Message = "One or more actions failed"
		klog.Warningf("Workflow %s/%s completed with failures", workflow.Namespace, workflow.Name)
	}

	return e.finalizeWorkflowStatus(status), nil
}

// RevertWorkflow reverts a workflow by executing rollback actions in reverse order
func (e *NativeWorkflowExecutor) RevertWorkflow(ctx context.Context, workflow *drv1alpha1.DRWorkflow, workflowStatus *drv1alpha1.WorkflowExecutionStatus) (*drv1alpha1.WorkflowExecutionStatus, error) {
	klog.Infof("Reverting workflow: %s/%s", workflow.Namespace, workflow.Name)

	// Create rollback workflow status object
	rollbackStatus := &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef:    workflowStatus.WorkflowRef,
		Phase:          "Running",
		StartTime:      &metav1.Time{Time: time.Now()},
		ActionStatuses: []drv1alpha1.ActionStatus{},
	}

	// Revert actions in reverse order
	succeededCount := 0
	totalCount := 0
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
			rollbackStatus.ActionStatuses = append([]drv1alpha1.ActionStatus{skippedStatus}, rollbackStatus.ActionStatuses...)
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
			rollbackStatus.ActionStatuses = append([]drv1alpha1.ActionStatus{skippedStatus}, rollbackStatus.ActionStatuses...)
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
			rollbackStatus.ActionStatuses = append([]drv1alpha1.ActionStatus{failedStatus}, rollbackStatus.ActionStatuses...)
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to rollback action %s", action.Name)
			return e.finalizeWorkflowStatus(rollbackStatus), fmt.Errorf("failed to get executor for action %s: %w", action.Name, err)
		}

		// Execute rollback and get status
		actionRollbackStatus, err := actionExecutor.Rollback(ctx, action, &actionStatus, nil)
		if err != nil {
			klog.Errorf("Failed to rollback action %s: %v", action.Name, err)
			// Add the failed action status to rollback status
			rollbackStatus.ActionStatuses = append([]drv1alpha1.ActionStatus{*actionRollbackStatus}, rollbackStatus.ActionStatuses...)
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to rollback action %s", action.Name)
			return e.finalizeWorkflowStatus(rollbackStatus), fmt.Errorf("failed to rollback action %s: %w", action.Name, err)
		}

		// Add successful action status to rollback status (prepend to maintain reverse order in status)
		rollbackStatus.ActionStatuses = append([]drv1alpha1.ActionStatus{*actionRollbackStatus}, rollbackStatus.ActionStatuses...)
		if actionRollbackStatus.Phase == drv1alpha1.PhaseSucceeded {
			succeededCount++
		}
		klog.Infof("Action %s rolled back successfully", action.Name)
	}

	// Update progress and finalize
	rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
	rollbackStatus.Progress = fmt.Sprintf("%d/%d actions rolled back", succeededCount, totalCount)
	rollbackStatus.Message = "Workflow reverted successfully"

	klog.Infof("Workflow %s/%s reverted successfully", workflow.Namespace, workflow.Name)
	return e.finalizeWorkflowStatus(rollbackStatus), nil
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
		if as.Phase == drv1alpha1.PhaseSucceeded || as.Phase == drv1alpha1.PhaseFailed {
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
}

// NewNativePlanExecutor creates a new NativePlanExecutor
func NewNativePlanExecutor(client client.Client, stageExecutor StageExecutor, workflowExecutor WorkflowExecutor) *NativePlanExecutor {
	return &NativePlanExecutor{
		client:           client,
		stageExecutor:    stageExecutor,
		workflowExecutor: workflowExecutor,
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

	// Build global params map (plan only uses value, not default)
	globalParams := make(map[string]interface{})
	for _, param := range plan.Spec.GlobalParams {
		if param.Value != "" {
			globalParams[param.Name] = param.Value
		}
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

			stageStatus, err := e.stageExecutor.ExecuteStage(ctx, plan, &stage, globalParams)
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
	} else {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = "One or more stages failed"
	}

	return e.updateExecutionStatus(ctx, execution, nil)
}

// RevertPlan reverts a DR plan
func (e *NativePlanExecutor) RevertPlan(ctx context.Context, plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution) error {
	klog.Infof("Reverting plan: %s/%s", plan.Namespace, plan.Name)

	// Get the target execution to revert (must be explicitly specified)
	if execution.Spec.RevertExecutionRef == "" {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = "revertExecutionRef must be specified for Revert operation"
		klog.Errorf("RevertExecutionRef not specified for execution %s/%s", execution.Namespace, execution.Name)
		return e.updateExecutionStatus(ctx, execution, nil)
	}

	targetExecutionName := execution.Spec.RevertExecutionRef
	klog.Infof("Reverting explicitly specified execution: %s", targetExecutionName)

	// Fetch the target execution to get its stage statuses
	targetExecution := &drv1alpha1.DRPlanExecution{}
	targetExecKey := client.ObjectKey{
		Name:      targetExecutionName,
		Namespace: plan.Namespace,
	}
	if err := e.client.Get(ctx, targetExecKey, targetExecution); err != nil {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = fmt.Sprintf("Failed to get target execution %s: %v", targetExecutionName, err)
		klog.Errorf("Failed to get target execution %s/%s: %v", plan.Namespace, targetExecutionName, err)
		return e.updateExecutionStatus(ctx, execution, nil)
	}

	// Validate that target execution was an Execute operation
	if targetExecution.Spec.OperationType != drv1alpha1.OperationTypeExecute {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = fmt.Sprintf("Cannot revert non-Execute operation (target is %s)", targetExecution.Spec.OperationType)
		klog.Errorf("Target execution %s/%s is not an Execute operation: %s", targetExecution.Namespace, targetExecution.Name, targetExecution.Spec.OperationType)
		return e.updateExecutionStatus(ctx, execution, nil)
	}

	// Validate that target execution succeeded
	if targetExecution.Status.Phase != drv1alpha1.PhaseSucceeded {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = fmt.Sprintf("Cannot revert execution in phase %s (must be Succeeded)", targetExecution.Status.Phase)
		klog.Errorf("Target execution %s/%s is not in Succeeded phase: %s", targetExecution.Namespace, targetExecution.Name, targetExecution.Status.Phase)
		return e.updateExecutionStatus(ctx, execution, nil)
	}

	if len(targetExecution.Status.StageStatuses) == 0 {
		execution.Status.Phase = drv1alpha1.PhaseFailed
		execution.Status.Message = "Target execution has no stage statuses to revert"
		klog.Warningf("Target execution %s/%s has no stage statuses", targetExecution.Namespace, targetExecution.Name)
		return e.updateExecutionStatus(ctx, execution, nil)
	}

	klog.Infof("Reverting based on target execution %s/%s with %d stages",
		targetExecution.Namespace, targetExecution.Name, len(targetExecution.Status.StageStatuses))

	// Initialize Revert execution status
	if execution.Status.StageStatuses == nil {
		execution.Status.StageStatuses = make([]drv1alpha1.StageStatus, 0, len(targetExecution.Status.StageStatuses))
	}
	if execution.Status.Summary == nil {
		execution.Status.Summary = &drv1alpha1.ExecutionSummary{
			TotalStages: len(targetExecution.Status.StageStatuses),
		}
	}

	// Revert stages in reverse order from the target execution
	succeededStages := 0
	skippedStages := 0
	totalActions := 0

	for i := len(targetExecution.Status.StageStatuses) - 1; i >= 0; i-- {
		originalStageStatus := targetExecution.Status.StageStatuses[i]

		if originalStageStatus.Phase != drv1alpha1.PhaseSucceeded {
			klog.V(4).Infof("Skipping revert for stage %s (phase=%s)", originalStageStatus.Name, originalStageStatus.Phase)
			// Record skipped stage in execution status
			skippedStageStatus := drv1alpha1.StageStatus{
				Name:    originalStageStatus.Name,
				Phase:   "Skipped",
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
		rollbackStageStatus, err := e.stageExecutor.RevertStage(ctx, plan, stage, &originalStageStatus)

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

	// Generate detailed success message
	execution.Status.Phase = drv1alpha1.PhaseSucceeded
	execution.Status.Message = fmt.Sprintf(
		"Plan reverted successfully: %d stage(s) rolled back, %d action(s) rolled back, %d stage(s) skipped",
		succeededStages, totalActions, skippedStages)

	klog.Infof("Plan %s/%s reverted successfully: %d stages, %d actions, %d skipped",
		plan.Namespace, plan.Name, succeededStages, totalActions, skippedStages)

	return e.updateExecutionStatus(ctx, execution, nil)
}

// CancelExecution cancels an ongoing execution
func (e *NativePlanExecutor) CancelExecution(_ context.Context, execution *drv1alpha1.DRPlanExecution) error {
	klog.Infof("Cancelling execution: %s/%s", execution.Namespace, execution.Name)

	// Mark all running/pending stages as cancelled
	for i := range execution.Status.StageStatuses {
		if execution.Status.StageStatuses[i].Phase == drv1alpha1.PhaseRunning || execution.Status.StageStatuses[i].Phase == drv1alpha1.PhasePending {
			execution.Status.StageStatuses[i].Phase = drv1alpha1.PhaseSkipped
			execution.Status.StageStatuses[i].Message = "Cancelled by user"
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
