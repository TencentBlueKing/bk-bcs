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
// +kubebuilder:webhook:path=/mutate-dr-bkbcs-tencent-com-v1alpha1-drplan,mutating=true,failurePolicy=fail,sideEffects=None,groups=dr.bkbcs.tencent.com,resources=drplans,verbs=create;update,versions=v1alpha1,name=mdrplan.kb.io,admissionReviewVersions=v1

// NOCC:tosa/linelength(设计如此)
// +kubebuilder:webhook:path=/validate-dr-bkbcs-tencent-com-v1alpha1-drplan,mutating=false,failurePolicy=fail,sideEffects=None,groups=dr.bkbcs.tencent.com,resources=drplans,verbs=create;update;delete,versions=v1alpha1,name=vdrplan.kb.io,admissionReviewVersions=v1

// DRPlanWebhook handles DRPlan admission webhook requests
type DRPlanWebhook struct {
	Client client.Client
}

// SetupWebhookWithManager registers the webhook with the manager
func (w *DRPlanWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	w.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(&drv1alpha1.DRPlan{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements webhook.Defaulter
func (w *DRPlanWebhook) Default(_ context.Context, obj runtime.Object) error {
	plan := obj.(*drv1alpha1.DRPlan)
	klog.V(4).Infof("Defaulting DRPlan: %s/%s", plan.Namespace, plan.Name)

	// Set default failure policy
	if plan.Spec.FailurePolicy == "" {
		plan.Spec.FailurePolicy = "Stop"
	}

	// Set defaults for each stage
	for i := range plan.Spec.Stages {
		stage := &plan.Spec.Stages[i]

		// Set default parallel to false
		if !stage.Parallel {
			stage.Parallel = false
		}

		// Set default failure policy for stage
		if stage.FailurePolicy == "" {
			if stage.Parallel {
				stage.FailurePolicy = "FailFast"
			} else {
				stage.FailurePolicy = plan.Spec.FailurePolicy
			}
		}

		// Set default namespace for workflow references
		for j := range stage.Workflows {
			wfRef := &stage.Workflows[j]
			if wfRef.WorkflowRef.Namespace == "" {
				wfRef.WorkflowRef.Namespace = plan.Namespace
			}
		}
	}

	return nil
}

// ValidateCreate implements webhook.Validator
func (w *DRPlanWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	plan := obj.(*drv1alpha1.DRPlan)
	klog.Infof("Validating create for DRPlan: %s/%s", plan.Namespace, plan.Name)

	warnings, errors := w.validatePlan(ctx, plan)
	if len(errors) > 0 {
		return warnings, fmt.Errorf("validation failed: %v", errors)
	}

	return warnings, nil
}

// ValidateUpdate implements webhook.Validator
func (w *DRPlanWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	_ = oldObj.(*drv1alpha1.DRPlan) // oldPlan not used currently
	newPlan := newObj.(*drv1alpha1.DRPlan)
	klog.Infof("Validating update for DRPlan: %s/%s", newPlan.Namespace, newPlan.Name)

	warnings, errors := w.validatePlan(ctx, newPlan)
	if len(errors) > 0 {
		return warnings, fmt.Errorf("validation failed: %v", errors)
	}

	// Check if there's a running execution
	if newPlan.Status.CurrentExecution != nil {
		warnings = append(warnings, "Plan is currently being executed, changes may not take effect until next execution")
	}

	return warnings, nil
}

// ValidateDelete implements webhook.Validator
func (w *DRPlanWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	plan := obj.(*drv1alpha1.DRPlan)
	klog.Infof("Validating delete for DRPlan: %s/%s", plan.Namespace, plan.Name)

	// Check 1: Fast path - check currentExecution from status
	if plan.Status.CurrentExecution != nil {
		return []string{"Plan has a running execution"},
			fmt.Errorf("cannot delete plan with running execution: %s/%s",
				plan.Status.CurrentExecution.Namespace, plan.Status.CurrentExecution.Name)
	}

	// Check 2: Comprehensive check - list all executions for this plan
	// This catches race conditions where execution was created but status not yet updated
	runningExecutions, err := w.findRunningExecutions(ctx, plan)
	if err != nil {
		klog.Errorf("Failed to check running executions for DRPlan %s/%s: %v", plan.Namespace, plan.Name, err)
		return nil, fmt.Errorf("failed to check running executions: %w", err)
	}

	if len(runningExecutions) > 0 {
		execNames := make([]string, len(runningExecutions))
		for i, exec := range runningExecutions {
			execNames[i] = fmt.Sprintf("%s/%s (phase=%s)", exec.Namespace, exec.Name, exec.Status.Phase)
		}
		return []string{fmt.Sprintf("Plan has %d running execution(s)", len(runningExecutions))},
			fmt.Errorf("cannot delete DRPlan %s/%s: has running executions: %v",
				plan.Namespace, plan.Name, execNames)
	}

	return nil, nil
}

// findRunningExecutions finds all running executions for this plan
func (w *DRPlanWebhook) findRunningExecutions(ctx context.Context, plan *drv1alpha1.DRPlan) ([]*drv1alpha1.DRPlanExecution, error) {
	// List all executions in the same namespace
	execList := &drv1alpha1.DRPlanExecutionList{}
	if err := w.Client.List(ctx, execList, client.InNamespace(plan.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list DRPlanExecutions: %w", err)
	}

	var runningExecutions []*drv1alpha1.DRPlanExecution
	for i := range execList.Items {
		exec := &execList.Items[i]

		// Check if this execution references this plan
		if exec.Spec.PlanRef != plan.Name {
			continue
		}

		// Check if execution is in running state (not terminal)
		phase := exec.Status.Phase
		if phase == "" || phase == drv1alpha1.PhasePending ||
			phase == drv1alpha1.PhaseRunning {
			runningExecutions = append(runningExecutions, exec)
		}
	}

	return runningExecutions, nil
}

// validatePlan performs comprehensive validation
// Refactored to reduce cyclomatic complexity by extracting validation sub-functions
func (w *DRPlanWebhook) validatePlan(ctx context.Context, plan *drv1alpha1.DRPlan) ([]string, []string) {
	var warnings []string
	var errors []string

	// Early return: Validate stages exist
	if len(plan.Spec.Stages) == 0 {
		errors = append(errors, "at least one stage is required")
		return warnings, errors
	}

	// Build stage name map and check for duplicates
	stageNames, nameErrors := buildStageNameMap(plan.Spec.Stages)
	errors = append(errors, nameErrors...)

	// Validate each stage
	for i, stage := range plan.Spec.Stages {
		// Validate stage basics
		errors = append(errors, validateStageBasics(stage, i)...)

		// Validate workflow references
		wfWarnings, wfErrors := w.validateStageWorkflows(ctx, stage, i, plan.Namespace)
		warnings = append(warnings, wfWarnings...)
		errors = append(errors, wfErrors...)

		// Validate parallel stage constraints
		warnings = append(warnings, validateParallelStage(stage, i)...)
	}

	// Validate stage dependencies
	errors = append(errors, validateStageDependencies(plan.Spec.Stages, stageNames)...)

	// Validate stage dependency cycles
	errors = append(errors, w.validateStageDependencyCycles(plan.Spec.Stages)...)

	// Validate global parameters
	errors = append(errors, validateGlobalParameters(plan.Spec.GlobalParams)...)

	return warnings, errors
}

// validateStageDependencyCycles detects cycles in stage dependencies
func (w *DRPlanWebhook) validateStageDependencyCycles(stages []drv1alpha1.Stage) []string {
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
