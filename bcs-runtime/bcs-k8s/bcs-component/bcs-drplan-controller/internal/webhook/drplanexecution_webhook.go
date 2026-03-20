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

package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// NOCC:tosa/linelength(设计如此)
// +kubebuilder:webhook:path=/validate-dr-bkbcs-tencent-com-v1alpha1-drplanexecution,mutating=false,failurePolicy=fail,sideEffects=None,groups=dr.bkbcs.tencent.com,resources=drplanexecutions,verbs=create;update,versions=v1alpha1,name=vdrplanexecution.kb.io,admissionReviewVersions=v1

// DRPlanExecutionWebhook handles DRPlanExecution admission webhook requests
type DRPlanExecutionWebhook struct {
	Client client.Client
}

// SetupWebhookWithManager registers the webhook with the manager
func (w *DRPlanExecutionWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	w.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(&drv1alpha1.DRPlanExecution{}).
		WithValidator(w).
		Complete()
}

// ValidateCreate implements webhook.Validator
func (w *DRPlanExecutionWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	execution := obj.(*drv1alpha1.DRPlanExecution)
	klog.Infof("Validating create for DRPlanExecution: %s/%s", execution.Namespace, execution.Name)

	warnings, errors := w.validateExecution(ctx, execution)
	if len(errors) > 0 {
		return warnings, fmt.Errorf("validation failed: %v", errors)
	}

	return warnings, nil
}

// ValidateUpdate implements webhook.Validator
func (w *DRPlanExecutionWebhook) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldExecution := oldObj.(*drv1alpha1.DRPlanExecution)
	newExecution := newObj.(*drv1alpha1.DRPlanExecution)
	klog.Infof("Validating update for DRPlanExecution: %s/%s", newExecution.Namespace, newExecution.Name)

	// Prevent spec changes after creation
	if oldExecution.Spec.PlanRef != newExecution.Spec.PlanRef || oldExecution.Spec.OperationType != newExecution.Spec.OperationType {
		return nil, fmt.Errorf("cannot modify spec fields after creation")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator
func (w *DRPlanExecutionWebhook) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	execution := obj.(*drv1alpha1.DRPlanExecution)
	// Warn if deleting a running execution
	if execution.Status.Phase == drv1alpha1.PhaseRunning {
		return []string{"Deleting a running execution may leave resources in inconsistent state"}, nil
	}
	return nil, nil
}

// validateExecution performs comprehensive validation
func (w *DRPlanExecutionWebhook) validateExecution(ctx context.Context, execution *drv1alpha1.DRPlanExecution) ([]string, []string) {
	var warnings []string
	var errors []string

	// Validate planRef
	if execution.Spec.PlanRef == "" {
		errors = append(errors, "planRef is required")
		return warnings, errors
	}

	// Get the DRPlan
	plan := &drv1alpha1.DRPlan{}
	planKey := client.ObjectKey{
		Name:      execution.Spec.PlanRef,
		Namespace: execution.Namespace,
	}
	if err := w.Client.Get(ctx, planKey, plan); err != nil {
		errors = append(errors, fmt.Sprintf("plan %s not found: %v", execution.Spec.PlanRef, err))
		return warnings, errors
	}

	// Check plan allows new execution: Ready allows any operation; Executed only allows Revert
	allowCreate := plan.Status.Phase == drv1alpha1.PlanPhaseReady ||
		(plan.Status.Phase == drv1alpha1.PlanPhaseExecuted && execution.Spec.OperationType == drv1alpha1.OperationTypeRevert)
	if !allowCreate {
		errors = append(errors, fmt.Sprintf("plan %s is not ready (phase=%s)", execution.Spec.PlanRef, plan.Status.Phase))
	}

	// Validate revert operation
	if execution.Spec.OperationType == drv1alpha1.OperationTypeRevert {
		// revertExecutionRef is required for Revert
		if execution.Spec.RevertExecutionRef == "" {
			errors = append(errors, "revertExecutionRef is required for Revert operation")
		} else {
			// Validate that the referenced execution exists and is valid
			targetExecution := &drv1alpha1.DRPlanExecution{}
			targetKey := client.ObjectKey{
				Name:      execution.Spec.RevertExecutionRef,
				Namespace: execution.Namespace,
			}
			if err := w.Client.Get(ctx, targetKey, targetExecution); err != nil {
				errors = append(errors, fmt.Sprintf("referenced execution %s not found: %v", execution.Spec.RevertExecutionRef, err))
			} else {
				// Validate target execution is an Execute operation
				if targetExecution.Spec.OperationType != drv1alpha1.OperationTypeExecute {
					errors = append(errors, fmt.Sprintf("referenced execution %s must be an Execute operation, got %s", execution.Spec.RevertExecutionRef, targetExecution.Spec.OperationType))
				}
				// Validate target execution succeeded
				if targetExecution.Status.Phase != drv1alpha1.PhaseSucceeded {
					errors = append(errors, fmt.Sprintf("referenced execution %s must be in Succeeded phase, got %s", execution.Spec.RevertExecutionRef, targetExecution.Status.Phase))
				}
				// Validate target execution references the same plan
				if targetExecution.Spec.PlanRef != execution.Spec.PlanRef {
					// NOCC:tosa/linelength (设计如此)
					errors = append(errors, fmt.Sprintf("referenced execution %s must belong to the same plan (expected %s, got %s)", execution.Spec.RevertExecutionRef, execution.Spec.PlanRef, targetExecution.Spec.PlanRef))
				}
			}
		}
	}

	// Check for concurrent executions
	if plan.Status.CurrentExecution != nil {
		errors = append(errors, fmt.Sprintf("plan %s already has a running execution: %s/%s",
			execution.Spec.PlanRef,
			plan.Status.CurrentExecution.Namespace,
			plan.Status.CurrentExecution.Name))
	}

	return warnings, errors
}
