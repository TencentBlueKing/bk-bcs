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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/executor"
)

const (
	// executionFinalizerName is the finalizer name for DRPlanExecution
	executionFinalizerName = "dr.bkbcs.tencent.com/execution-finalizer"
)

// DRPlanExecutionReconciler reconciles a DRPlanExecution object
type DRPlanExecutionReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	PlanExecutor executor.PlanExecutor
}

// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drplanexecutions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drplanexecutions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drplanexecutions/finalizers,verbs=update

// Reconcile manages DRPlanExecution lifecycle
// Refactored to reduce cyclomatic complexity by extracting helper methods
func (r *DRPlanExecutionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconciling DRPlanExecution: %s/%s", req.Namespace, req.Name)

	// Fetch the DRPlanExecution instance
	execution := &drv1alpha1.DRPlanExecution{}
	if err := r.Get(ctx, req.NamespacedName, execution); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("DRPlanExecution %s/%s not found, ignoring", req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		klog.Errorf("Failed to get DRPlanExecution %s/%s: %v", req.Namespace, req.Name, err)
		return ctrl.Result{}, err
	}

	// Handle deletion (finalizer logic)
	if !execution.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, execution)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(execution, executionFinalizerName) {
		klog.V(4).Infof("Adding finalizer to DRPlanExecution %s/%s", execution.Namespace, execution.Name)
		controllerutil.AddFinalizer(execution, executionFinalizerName)
		if err := r.Update(ctx, execution); err != nil {
			klog.Errorf("Failed to add finalizer to DRPlanExecution %s/%s: %v", execution.Namespace, execution.Name, err)
			return ctrl.Result{}, err
		}
		// Requeue to continue processing
		return ctrl.Result{Requeue: true}, nil
	}

	// Early return: Check if execution is in terminal state
	if r.isExecutionInTerminalState(execution) {
		klog.V(4).Infof("DRPlanExecution %s/%s is in terminal state: %s", req.Namespace, req.Name, execution.Status.Phase)

		// Before returning, ensure plan status is updated (in case it wasn't updated before)
		plan, err := r.fetchDRPlan(ctx, execution)
		if err != nil {
			klog.Warningf("Failed to fetch plan for completed execution %s/%s: %v", req.Namespace, req.Name, err)
		} else {
			if err := r.updatePlanAfterCompletion(ctx, execution, plan); err != nil {
				klog.Warningf("Failed to update plan after completion (terminal state check): %v", err)
			}
		}

		return ctrl.Result{}, nil
	}

	// Early return: Check for cancellation annotation
	if r.checkCancellationRequested(execution) {
		klog.Infof("DRPlanExecution %s/%s cancellation requested", req.Namespace, req.Name)
		if err := r.handleCancellation(ctx, execution); err != nil {
			klog.Errorf("Failed to cancel execution %s/%s: %v", req.Namespace, req.Name, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Fetch the DRPlan
	plan, err := r.fetchDRPlan(ctx, execution)
	if err != nil {
		return r.updateExecutionStatus(ctx, execution, "Failed", err.Error())
	}

	// Validate plan allows this execution (Ready for any op; Executed only for Revert)
	if err := r.validatePlanReady(plan, execution); err != nil {
		klog.Warningf("DRPlanExecution %s/%s: %v", req.Namespace, req.Name, err)
		return r.updateExecutionStatus(ctx, execution, "Failed", err.Error())
	}

	// Initialize execution if it's pending
	if execution.Status.Phase == "" || execution.Status.Phase == drv1alpha1.PhasePending {
		if err := r.initializeExecution(ctx, execution, plan); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Execute the operation
	if err := r.executeOperation(ctx, execution, plan); err != nil {
		klog.Errorf("DRPlanExecution %s/%s failed: %v", req.Namespace, req.Name, err)
		return r.updateExecutionStatus(ctx, execution, "Failed", err.Error())
	}

	// Refresh execution status
	if err := r.Get(ctx, req.NamespacedName, execution); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue if still running
	if r.shouldRequeue(execution) {
		klog.V(4).Infof("DRPlanExecution %s/%s still running, requeuing", req.Namespace, req.Name)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Update plan status after execution completes
	if err := r.updatePlanAfterCompletion(ctx, execution, plan); err != nil {
		klog.Warningf("Failed to update plan after completion: %v", err)
		// Don't return error, execution itself succeeded
	}

	klog.Infof("DRPlanExecution %s/%s completed with phase: %s", req.Namespace, req.Name, execution.Status.Phase)
	return ctrl.Result{}, nil
}

// handleCancellation handles execution cancellation
func (r *DRPlanExecutionReconciler) handleCancellation(ctx context.Context, execution *drv1alpha1.DRPlanExecution) error {
	if execution.Status.Phase == drv1alpha1.PhaseRunning {
		klog.Infof("Cancelling execution %s/%s", execution.Namespace, execution.Name)

		// Call executor to cancel
		if err := r.PlanExecutor.CancelExecution(ctx, execution); err != nil {
			return fmt.Errorf("failed to cancel execution: %w", err)
		}

		// Update status
		execution.Status.Phase = drv1alpha1.PhaseCancelled
		execution.Status.CompletionTime = &metav1.Time{Time: time.Now()}
		execution.Status.Message = "Execution cancelled by user"

		if err := r.Status().Update(ctx, execution); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		klog.Infof("Execution %s/%s cancelled successfully", execution.Namespace, execution.Name)
	}

	return nil
}

// updateExecutionStatus updates execution status with phase and message
func (r *DRPlanExecutionReconciler) updateExecutionStatus(ctx context.Context, execution *drv1alpha1.DRPlanExecution, phase, message string) (ctrl.Result, error) {
	execution.Status.Phase = phase
	execution.Status.Message = message
	if phase == drv1alpha1.PhaseFailed || phase == drv1alpha1.PhaseSucceeded || phase == drv1alpha1.PhaseCancelled {
		execution.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	}

	if err := r.Status().Update(ctx, execution); err != nil {
		klog.Errorf("Failed to update execution status: %v", err)
		return ctrl.Result{}, err
	}

	// Update the corresponding record in DRPlan's ExecutionHistory
	if err := r.updatePlanExecutionHistory(ctx, execution); err != nil {
		klog.Warningf("Failed to update plan execution history for %s/%s: %v", execution.Namespace, execution.Name, err)
		// Non-fatal: execution status is already updated
	}

	return ctrl.Result{}, nil
}

// updatePlanExecutionHistory updates the corresponding execution record in DRPlan's ExecutionHistory
func (r *DRPlanExecutionReconciler) updatePlanExecutionHistory(ctx context.Context, execution *drv1alpha1.DRPlanExecution) error {
	// Fetch the DRPlan
	plan := &drv1alpha1.DRPlan{}
	planKey := client.ObjectKey{
		Name:      execution.Spec.PlanRef,
		Namespace: execution.Namespace,
	}
	if err := r.Get(ctx, planKey, plan); err != nil {
		return fmt.Errorf("failed to get plan %s/%s: %w", execution.Namespace, execution.Spec.PlanRef, err)
	}

	// Find and update the matching execution record in history
	updated := false
	for i := range plan.Status.ExecutionHistory {
		record := &plan.Status.ExecutionHistory[i]
		if record.Name == execution.Name && record.Namespace == execution.Namespace {
			// Update phase and completion time
			record.Phase = execution.Status.Phase
			if execution.Status.CompletionTime != nil {
				record.CompletionTime = execution.Status.CompletionTime
			}
			updated = true
			break
		}
	}

	if !updated {
		klog.V(4).Infof("Execution %s/%s not found in plan history (may have been trimmed)", execution.Namespace, execution.Name)
		return nil
	}

	// Update plan status
	if err := r.Status().Update(ctx, plan); err != nil {
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	klog.V(4).Infof("Updated execution history for %s/%s in plan %s", execution.Namespace, execution.Name, plan.Name)
	return nil
}

// handleDeletion handles the deletion of DRPlanExecution with finalizer cleanup
func (r *DRPlanExecutionReconciler) handleDeletion(ctx context.Context, execution *drv1alpha1.DRPlanExecution) (ctrl.Result, error) {
	klog.Infof("Handling deletion of DRPlanExecution %s/%s", execution.Namespace, execution.Name)

	if controllerutil.ContainsFinalizer(execution, executionFinalizerName) {
		// Ensure execution history is updated with final status before deletion
		if err := r.ensureExecutionHistoryUpdated(ctx, execution); err != nil {
			klog.Errorf("Failed to update execution history before deletion for %s/%s: %v",
				execution.Namespace, execution.Name, err)
			return ctrl.Result{}, err
		}

		// Remove finalizer
		klog.V(4).Infof("Removing finalizer from DRPlanExecution %s/%s", execution.Namespace, execution.Name)
		controllerutil.RemoveFinalizer(execution, executionFinalizerName)
		if err := r.Update(ctx, execution); err != nil {
			klog.Errorf("Failed to remove finalizer from DRPlanExecution %s/%s: %v",
				execution.Namespace, execution.Name, err)
			return ctrl.Result{}, err
		}
	}

	klog.Infof("DRPlanExecution %s/%s finalizer cleanup completed", execution.Namespace, execution.Name)
	return ctrl.Result{}, nil
}

// ensureExecutionHistoryUpdated ensures the execution history in DRPlan is updated with final status
// This is called before deleting the DRPlanExecution CR to preserve accurate history even if CR is deleted
func (r *DRPlanExecutionReconciler) ensureExecutionHistoryUpdated(ctx context.Context, execution *drv1alpha1.DRPlanExecution) error {
	klog.V(4).Infof("Ensuring execution history is updated for %s/%s before deletion", execution.Namespace, execution.Name)

	// Fetch the DRPlan
	plan := &drv1alpha1.DRPlan{}
	planKey := client.ObjectKey{
		Name:      execution.Spec.PlanRef,
		Namespace: execution.Namespace,
	}
	if err := r.Get(ctx, planKey, plan); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("DRPlan %s/%s not found, skipping history update", execution.Namespace, execution.Spec.PlanRef)
			return nil
		}
		return fmt.Errorf("failed to get plan %s/%s: %w", execution.Namespace, execution.Spec.PlanRef, err)
	}

	// Find and update the matching execution record in history with final status
	updated := false
	for i := range plan.Status.ExecutionHistory {
		record := &plan.Status.ExecutionHistory[i]
		if record.Name == execution.Name && record.Namespace == execution.Namespace {
			klog.V(4).Infof("Updating execution history record for %s/%s with final status: %s",
				execution.Namespace, execution.Name, execution.Status.Phase)

			// Update phase
			if execution.Status.Phase != "" {
				record.Phase = execution.Status.Phase
			} else {
				// If execution was deleted before completion, mark as Cancelled
				record.Phase = drv1alpha1.PhaseCancelled
			}

			// Update completion time
			if execution.Status.CompletionTime != nil {
				record.CompletionTime = execution.Status.CompletionTime
			} else {
				// If no completion time (deleted while running), use current time
				now := metav1.Now()
				record.CompletionTime = &now
			}

			updated = true
			break
		}
	}

	if !updated {
		klog.V(4).Infof("Execution %s/%s not found in plan history (may have been trimmed or never started)",
			execution.Namespace, execution.Name)
		return nil
	}

	// Update plan status
	if err := r.Status().Update(ctx, plan); err != nil {
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	klog.Infof("Execution history updated for %s/%s in plan %s/%s before deletion",
		execution.Namespace, execution.Name, plan.Namespace, plan.Name)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DRPlanExecutionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&drv1alpha1.DRPlanExecution{}).
		Named("drplanexecution").
		Complete(r)
}
