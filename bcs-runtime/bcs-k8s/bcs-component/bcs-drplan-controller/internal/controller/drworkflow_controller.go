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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// DRWorkflowReconciler reconciles a DRWorkflow object
type DRWorkflowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drworkflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drworkflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drworkflows/finalizers,verbs=update

// Reconcile validates and manages the lifecycle of DRWorkflow resources
func (r *DRWorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconciling DRWorkflow: %s/%s", req.Namespace, req.Name)

	// Fetch the DRWorkflow instance
	workflow := &drv1alpha1.DRWorkflow{}
	if err := r.Get(ctx, req.NamespacedName, workflow); err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Infof("DRWorkflow %s/%s not found, ignoring", req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		klog.Errorf("Failed to get DRWorkflow %s/%s: %v", req.Namespace, req.Name, err)
		return ctrl.Result{}, err
	}

	// Check if the workflow has been updated
	if workflow.Status.ObservedGeneration == workflow.Generation && workflow.Status.Phase != "" {
		klog.V(4).Infof("DRWorkflow %s/%s is up-to-date", req.Namespace, req.Name)
		return ctrl.Result{}, nil
	}

	// Validate the workflow
	validationErrors := r.validateWorkflow(workflow)

	// Update status
	oldStatus := workflow.Status.DeepCopy()
	if len(validationErrors) > 0 {
		workflow.Status.Phase = drv1alpha1.PlanPhaseInvalid
		workflow.Status.Conditions = []metav1.Condition{
			{
				Type:               "Ready",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: workflow.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "ValidationFailed",
				Message:            fmt.Sprintf("Validation failed: %v", validationErrors),
			},
		}
		klog.Warningf("DRWorkflow %s/%s validation failed: %v", req.Namespace, req.Name, validationErrors)
	} else {
		workflow.Status.Phase = drv1alpha1.PlanPhaseReady
		workflow.Status.Conditions = []metav1.Condition{
			{
				Type:               "Ready",
				Status:             metav1.ConditionTrue,
				ObservedGeneration: workflow.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "ValidationSucceeded",
				Message:            "Workflow is valid and ready to execute",
			},
		}
		klog.Infof("DRWorkflow %s/%s validated successfully", req.Namespace, req.Name)
	}
	workflow.Status.ObservedGeneration = workflow.Generation

	// Update status if changed
	if !statusEqual(oldStatus, &workflow.Status) {
		if err := r.Status().Update(ctx, workflow); err != nil {
			klog.Errorf("Failed to update DRWorkflow status %s/%s: %v", req.Namespace, req.Name, err)
			return ctrl.Result{}, err
		}
		klog.V(4).Infof("DRWorkflow %s/%s status updated to %s", req.Namespace, req.Name, workflow.Status.Phase)
	}

	return ctrl.Result{}, nil
}

// validateWorkflow validates the workflow definition
// Refactored to reduce cyclomatic complexity by using validator pattern
func (r *DRWorkflowReconciler) validateWorkflow(workflow *drv1alpha1.DRWorkflow) []string {
	var errors []string

	// Validate actions exist
	if len(workflow.Spec.Actions) == 0 {
		errors = append(errors, "at least one action is required")
		return errors
	}

	// Validate action names uniqueness
	errors = append(errors, validateActionNames(workflow.Spec.Actions)...)

	// Validate each action using validator registry
	validatorRegistry := NewActionValidatorRegistry()
	for i, action := range workflow.Spec.Actions {
		actionErrors := validatorRegistry.ValidateAction(&action, i)
		errors = append(errors, actionErrors...)
	}

	// Validate parameters
	errors = append(errors, validateParameters(workflow.Spec.Parameters)...)

	return errors
}

// statusEqual checks if two status objects are equal
func statusEqual(a, b *drv1alpha1.DRWorkflowStatus) bool {
	if a.Phase != b.Phase {
		return false
	}
	if a.ObservedGeneration != b.ObservedGeneration {
		return false
	}
	if len(a.Conditions) != len(b.Conditions) {
		return false
	}
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *DRWorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&drv1alpha1.DRWorkflow{}).
		Named("drworkflow").
		Complete(r)
}
