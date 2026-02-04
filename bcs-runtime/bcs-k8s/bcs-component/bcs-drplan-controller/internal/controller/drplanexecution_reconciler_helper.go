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

package controller

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// isExecutionInTerminalState checks if execution is in a terminal state
func (r *DRPlanExecutionReconciler) isExecutionInTerminalState(execution *drv1alpha1.DRPlanExecution) bool {
	return execution.Status.Phase == drv1alpha1.PhaseSucceeded ||
		execution.Status.Phase == drv1alpha1.PhaseFailed ||
		execution.Status.Phase == drv1alpha1.PhaseCancelled
}

// checkCancellationRequested checks if cancellation is requested
func (r *DRPlanExecutionReconciler) checkCancellationRequested(execution *drv1alpha1.DRPlanExecution) bool {
	return execution.Annotations != nil &&
		execution.Annotations["dr.bkbcs.tencent.com/cancel"] == "true"
}

// fetchDRPlan fetches the associated DRPlan
func (r *DRPlanExecutionReconciler) fetchDRPlan(ctx context.Context, execution *drv1alpha1.DRPlanExecution) (*drv1alpha1.DRPlan, error) {
	plan := &drv1alpha1.DRPlan{}
	planKey := client.ObjectKey{
		Name:      execution.Spec.PlanRef,
		Namespace: execution.Namespace,
	}

	if err := r.Get(ctx, planKey, plan); err != nil {
		klog.Errorf("Failed to get DRPlan %s/%s: %v", execution.Namespace, execution.Spec.PlanRef, err)
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	return plan, nil
}

// validatePlanReady validates that the plan allows the execution: Ready allows any operation; Executed only allows Revert.
func (r *DRPlanExecutionReconciler) validatePlanReady(plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution) error {
	allow := plan.Status.Phase == drv1alpha1.PlanPhaseReady ||
		(plan.Status.Phase == drv1alpha1.PlanPhaseExecuted && execution.Spec.OperationType == drv1alpha1.OperationTypeRevert)
	if !allow {
		return fmt.Errorf("plan %s is not ready (phase=%s)", plan.Name, plan.Status.Phase)
	}
	return nil
}

// initializeExecution initializes a pending execution
func (r *DRPlanExecutionReconciler) initializeExecution(ctx context.Context, execution *drv1alpha1.DRPlanExecution, plan *drv1alpha1.DRPlan) error {
	startTime := &metav1.Time{Time: metav1.Now().Time}

	// Add to execution history (if not already present)
	r.addToExecutionHistory(plan, execution.Name, execution.Namespace, execution.Spec.OperationType, startTime)

	// Update plan's currentExecution reference
	plan.Status.CurrentExecution = &drv1alpha1.ObjectReference{
		Name:      execution.Name,
		Namespace: execution.Namespace,
	}
	if err := r.Status().Update(ctx, plan); err != nil {
		klog.Errorf("Failed to update plan currentExecution %s/%s: %v", plan.Namespace, plan.Name, err)
		return err
	}

	// Update execution status
	execution.Status.Phase = drv1alpha1.PhaseRunning
	execution.Status.StartTime = startTime
	if err := r.Status().Update(ctx, execution); err != nil {
		klog.Errorf("Failed to update execution status %s/%s: %v", execution.Namespace, execution.Name, err)
		return err
	}

	klog.Infof("DRPlanExecution %s/%s started", execution.Namespace, execution.Name)
	return nil
}

// executeOperation executes the requested operation (Execute or Revert)
func (r *DRPlanExecutionReconciler) executeOperation(ctx context.Context, execution *drv1alpha1.DRPlanExecution, plan *drv1alpha1.DRPlan) error {
	switch execution.Spec.OperationType {
	case drv1alpha1.OperationTypeExecute:
		return r.PlanExecutor.ExecutePlan(ctx, plan, execution)
	case drv1alpha1.OperationTypeRevert:
		return r.PlanExecutor.RevertPlan(ctx, plan, execution)
	default:
		return fmt.Errorf("unsupported operation type: %s", execution.Spec.OperationType)
	}
}

// shouldRequeue checks if the execution should be requeued
func (r *DRPlanExecutionReconciler) shouldRequeue(execution *drv1alpha1.DRPlanExecution) bool {
	return execution.Status.Phase == drv1alpha1.PhaseRunning
}

// addToExecutionHistory adds an execution record to the plan's execution history (max 10, newest first)
// This is idempotent - if the execution already exists in history AND is not in terminal state, it won't be added again
// If the same name exists but in terminal state (Succeeded/Failed/Cancelled), a new record will be added (for re-execution)
func (r *DRPlanExecutionReconciler) addToExecutionHistory(plan *drv1alpha1.DRPlan, execName, execNamespace, operationType string, startTime *metav1.Time) {
	// Check if already in history (idempotent check with terminal state consideration)
	for _, record := range plan.Status.ExecutionHistory {
		if record.Name == execName && record.Namespace == execNamespace {
			// If the existing record is in terminal state, allow adding a new record (for re-execution)
			if record.Phase == drv1alpha1.PhaseSucceeded ||
				record.Phase == drv1alpha1.PhaseFailed ||
				record.Phase == drv1alpha1.PhaseCancelled {
				klog.V(4).Infof("Execution %s/%s exists in history with terminal phase %s, allowing re-execution record",
					execNamespace, execName, record.Phase)
				// Continue to add new record (fall through)
				break
			}
			// If not in terminal state (Pending/Running), skip to prevent duplicate
			klog.V(4).Infof("Execution %s/%s already in history with non-terminal phase %s, skipping",
				execNamespace, execName, record.Phase)
			return
		}
	}

	newRecord := drv1alpha1.ExecutionRecord{
		Name:          execName,
		Namespace:     execNamespace,
		OperationType: operationType,
		Phase:         drv1alpha1.PhasePending,
		StartTime:     startTime,
	}

	// Prepend to history (newest first)
	plan.Status.ExecutionHistory = append([]drv1alpha1.ExecutionRecord{newRecord}, plan.Status.ExecutionHistory...)

	// Keep only the most recent 10 records
	if len(plan.Status.ExecutionHistory) > 10 {
		plan.Status.ExecutionHistory = plan.Status.ExecutionHistory[:10]
	}

	klog.V(4).Infof("Added execution %s/%s to plan %s/%s history (total: %d)",
		execNamespace, execName, plan.Namespace, plan.Name, len(plan.Status.ExecutionHistory))
}

// updatePlanAfterCompletion updates plan status after execution completes
func (r *DRPlanExecutionReconciler) updatePlanAfterCompletion(ctx context.Context, execution *drv1alpha1.DRPlanExecution, plan *drv1alpha1.DRPlan) error {
	klog.V(4).Infof("updatePlanAfterCompletion called for execution %s/%s (phase=%s, operationType=%s)",
		execution.Namespace, execution.Name, execution.Status.Phase, execution.Spec.OperationType)

	if !r.isExecutionInTerminalState(execution) {
		klog.V(4).Infof("Execution %s/%s not in terminal state, skipping plan update", execution.Namespace, execution.Name)
		return nil
	}

	// Refresh plan to get latest status
	planKey := client.ObjectKey{Name: plan.Name, Namespace: plan.Namespace}
	if err := r.Get(ctx, planKey, plan); err != nil {
		klog.Warningf("Failed to refresh plan %s/%s: %v", plan.Namespace, plan.Name, err)
		return err
	}

	// Idempotency check: if this execution was already processed, skip
	switch execution.Spec.OperationType {
	case drv1alpha1.OperationTypeExecute:
		if plan.Status.LastExecutionRef == execution.Name && plan.Status.Phase == drv1alpha1.PlanPhaseExecuted {
			klog.V(4).Infof("Plan %s/%s already updated for execution %s, skipping",
				plan.Namespace, plan.Name, execution.Name)
			return nil
		}
	case drv1alpha1.OperationTypeRevert:
		// For revert, if currentExecution is nil and phase is Ready, likely already processed
		if plan.Status.CurrentExecution == nil && plan.Status.Phase == drv1alpha1.PlanPhaseReady {
			klog.V(4).Infof("Plan %s/%s likely already updated for revert execution %s, skipping",
				plan.Namespace, plan.Name, execution.Name)
			return nil
		}
	}

	klog.Infof("Updating plan %s/%s after execution %s completion (execution phase=%s, operation=%s)",
		plan.Namespace, plan.Name, execution.Name, execution.Status.Phase, execution.Spec.OperationType)

	// Clear currentExecution
	plan.Status.CurrentExecution = nil

	// Update plan phase and lastExecutionRef based on execution result
	if execution.Status.Phase == drv1alpha1.PhaseSucceeded {
		// Always update lastExecutionRef regardless of operation type (Execute or Revert)
		// This ensures lastExecutionRef always points to the most recent successful operation
		plan.Status.LastExecutionRef = execution.Name
		plan.Status.LastExecutionTime = execution.Status.CompletionTime

		switch execution.Spec.OperationType {
		case drv1alpha1.OperationTypeExecute:
			plan.Status.Phase = drv1alpha1.PlanPhaseExecuted
			klog.Infof("Setting plan %s/%s: phase=Executed, lastExecutionRef=%s",
				plan.Namespace, plan.Name, execution.Name)
		case drv1alpha1.OperationTypeRevert:
			plan.Status.Phase = drv1alpha1.PlanPhaseReady
			klog.Infof("Setting plan %s/%s: phase=Ready (after revert), lastExecutionRef=%s",
				plan.Namespace, plan.Name, execution.Name)
		}
	}

	// Also update the execution record in history with final phase and completion time
	for i := range plan.Status.ExecutionHistory {
		record := &plan.Status.ExecutionHistory[i]
		if record.Name == execution.Name && record.Namespace == execution.Namespace {
			klog.V(4).Infof("Updating execution history record for %s/%s: phase=%s -> %s",
				execution.Namespace, execution.Name, record.Phase, execution.Status.Phase)
			record.Phase = execution.Status.Phase
			if execution.Status.CompletionTime != nil {
				record.CompletionTime = execution.Status.CompletionTime
			}
			break
		}
	}

	if err := r.Status().Update(ctx, plan); err != nil {
		klog.Errorf("Failed to update plan status %s/%s: %v", plan.Namespace, plan.Name, err)
		return err
	}

	klog.Infof("Successfully updated plan %s/%s status: phase=%s, lastExecutionRef=%s, currentExecution=%v",
		plan.Namespace, plan.Name, plan.Status.Phase, plan.Status.LastExecutionRef, plan.Status.CurrentExecution)
	return nil
}
