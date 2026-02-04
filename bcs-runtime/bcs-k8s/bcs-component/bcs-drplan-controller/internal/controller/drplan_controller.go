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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// DRPlanReconciler reconciles a DRPlan object
type DRPlanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drplans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drplans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dr.bkbcs.tencent.com,resources=drplans/finalizers,verbs=update

// Reconcile validates and manages the lifecycle of DRPlan resources
func (r *DRPlanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconciling DRPlan: %s/%s", req.Namespace, req.Name)

	// Fetch the DRPlan instance
	plan := &drv1alpha1.DRPlan{}
	if err := r.Get(ctx, req.NamespacedName, plan); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("DRPlan %s/%s not found, ignoring", req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		klog.Errorf("Failed to get DRPlan %s/%s: %v", req.Namespace, req.Name, err)
		return ctrl.Result{}, err
	}

	// Check if the plan has been updated
	if plan.Status.ObservedGeneration == plan.Generation && plan.Status.Phase != "" {
		klog.V(4).Infof("DRPlan %s/%s is up-to-date", req.Namespace, req.Name)
		return ctrl.Result{}, nil
	}

	// Validate the plan
	validationErrors := r.validatePlan(ctx, plan)

	// Update status
	oldStatus := plan.Status.DeepCopy()
	if len(validationErrors) > 0 {
		plan.Status.Phase = drv1alpha1.PlanPhaseInvalid
		plan.Status.Conditions = []metav1.Condition{
			{
				Type:               "Ready",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: plan.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "ValidationFailed",
				Message:            fmt.Sprintf("Validation failed: %v", validationErrors),
			},
		}
		klog.Warningf("DRPlan %s/%s validation failed: %v", req.Namespace, req.Name, validationErrors)
	} else {
		plan.Status.Phase = drv1alpha1.PlanPhaseReady
		plan.Status.Conditions = []metav1.Condition{
			{
				Type:               "Ready",
				Status:             metav1.ConditionTrue,
				ObservedGeneration: plan.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "ValidationSucceeded",
				Message:            "Plan is valid and ready to execute",
			},
		}
		klog.Infof("DRPlan %s/%s validated successfully", req.Namespace, req.Name)
	}
	plan.Status.ObservedGeneration = plan.Generation

	// Update status if changed
	if !drPlanStatusEqual(oldStatus, &plan.Status) {
		if err := r.Status().Update(ctx, plan); err != nil {
			klog.Errorf("Failed to update DRPlan status %s/%s: %v", req.Namespace, req.Name, err)
			return ctrl.Result{}, err
		}
		klog.V(4).Infof("DRPlan %s/%s status updated to %s", req.Namespace, req.Name, plan.Status.Phase)
	}

	return ctrl.Result{}, nil
}

// validatePlan validates the plan definition
func (r *DRPlanReconciler) validatePlan(ctx context.Context, plan *drv1alpha1.DRPlan) []string {
	var errors []string

	// Validate stages
	if len(plan.Spec.Stages) == 0 {
		errors = append(errors, "at least one stage is required")
		return errors
	}

	// Validate each stage
	stageNames := make(map[string]bool)
	for i, stage := range plan.Spec.Stages {
		// Check unique stage names
		if stageNames[stage.Name] {
			errors = append(errors, fmt.Sprintf("duplicate stage name: %s", stage.Name))
		}
		stageNames[stage.Name] = true

		// Validate stage
		if stage.Name == "" {
			errors = append(errors, fmt.Sprintf("stage[%d]: name is required", i))
		}

		// Validate workflows
		if len(stage.Workflows) == 0 {
			errors = append(errors, fmt.Sprintf("stage[%d] %s: at least one workflow is required", i, stage.Name))
		}

		// Validate workflow references
		for j, wfRef := range stage.Workflows {
			if wfRef.WorkflowRef.Name == "" {
				errors = append(errors, fmt.Sprintf("stage[%d] %s workflow[%d]: workflowRef.name is required", i, stage.Name, j))
			}

			// Check if workflow exists
			workflow := &drv1alpha1.DRWorkflow{}
			namespace := wfRef.WorkflowRef.Namespace
			if namespace == "" {
				namespace = plan.Namespace
			}
			key := client.ObjectKey{
				Name:      wfRef.WorkflowRef.Name,
				Namespace: namespace,
			}
			if err := r.Get(ctx, key, workflow); err != nil {
				if apierrors.IsNotFound(err) {
					errors = append(errors, fmt.Sprintf("stage[%d] %s workflow[%d]: workflow %s/%s not found", i, stage.Name, j, namespace, wfRef.WorkflowRef.Name))
				} else {
					errors = append(errors, fmt.Sprintf("stage[%d] %s workflow[%d]: failed to get workflow %s/%s: %v", i, stage.Name, j, namespace, wfRef.WorkflowRef.Name, err))
				}
			} else if workflow.Status.Phase != drv1alpha1.PlanPhaseReady {
				errors = append(errors, fmt.Sprintf(
					"stage[%d] %s workflow[%d]: workflow %s/%s is not ready (phase=%s)",
					i, stage.Name, j, namespace, wfRef.WorkflowRef.Name, workflow.Status.Phase))
			}
		}

		// Validate stage dependencies
		for _, depName := range stage.DependsOn {
			if !stageNames[depName] {
				errors = append(errors, fmt.Sprintf("stage[%d] %s: depends on non-existent stage '%s'", i, stage.Name, depName))
			}
			if depName == stage.Name {
				errors = append(errors, fmt.Sprintf("stage[%d] %s: cannot depend on itself", i, stage.Name))
			}
		}
	}

	// Validate stage dependency graph (detect cycles)
	if cycleErrors := r.validateStageDependencyCycles(plan.Spec.Stages); len(cycleErrors) > 0 {
		errors = append(errors, cycleErrors...)
	}

	// Validate global parameters
	paramNames := make(map[string]bool)
	for i, param := range plan.Spec.GlobalParams {
		if param.Name == "" {
			errors = append(errors, fmt.Sprintf("globalParams[%d]: name is required", i))
		}
		if paramNames[param.Name] {
			errors = append(errors, fmt.Sprintf("globalParams[%d]: duplicate parameter name '%s'", i, param.Name))
		}
		paramNames[param.Name] = true
	}

	return errors
}

// validateStageDependencyCycles detects cycles in stage dependencies
func (r *DRPlanReconciler) validateStageDependencyCycles(stages []drv1alpha1.Stage) []string {
	var errors []string

	// Build adjacency list
	graph := make(map[string][]string)
	for _, stage := range stages {
		graph[stage.Name] = stage.DependsOn
	}

	// Track visited nodes and recursion stack
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	// DFS to detect cycles
	var dfs func(string, []string) bool
	dfs = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range graph[node] {
			if !visited[dep] {
				if dfs(dep, path) {
					return true
				}
			} else if recStack[dep] {
				// Found cycle
				cyclePath := append([]string(nil), append(path, dep)...)
				errors = append(errors, fmt.Sprintf("cycle detected in stage dependencies: %v", cyclePath))
				return true
			}
		}

		recStack[node] = false
		return false
	}

	// Check each stage
	for _, stage := range stages {
		if !visited[stage.Name] {
			if dfs(stage.Name, []string{}) {
				break // Stop after finding first cycle
			}
		}
	}

	return errors
}

// drPlanStatusEqual checks if two DRPlan status objects are equal
func drPlanStatusEqual(a, b *drv1alpha1.DRPlanStatus) bool {
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
func (r *DRPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&drv1alpha1.DRPlan{}).
		Named("drplan").
		Complete(r)
}
