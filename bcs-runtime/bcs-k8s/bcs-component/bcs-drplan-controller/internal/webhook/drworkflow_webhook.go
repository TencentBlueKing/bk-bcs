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
// +kubebuilder:webhook:path=/mutate-dr-bkbcs-tencent-com-v1alpha1-drworkflow,mutating=true,failurePolicy=fail,sideEffects=None,groups=dr.bkbcs.tencent.com,resources=drworkflows,verbs=create;update,versions=v1alpha1,name=mdrworkflow.kb.io,admissionReviewVersions=v1

// NOCC:tosa/linelength(设计如此)
// +kubebuilder:webhook:path=/validate-dr-bkbcs-tencent-com-v1alpha1-drworkflow,mutating=false,failurePolicy=fail,sideEffects=None,groups=dr.bkbcs.tencent.com,resources=drworkflows,verbs=create;update;delete,versions=v1alpha1,name=vdrworkflow.kb.io,admissionReviewVersions=v1

// DRWorkflowWebhook handles DRWorkflow admission webhook requests
type DRWorkflowWebhook struct {
	Client client.Client
}

// SetupWebhookWithManager registers the webhook with the manager
func (w *DRWorkflowWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	w.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(&drv1alpha1.DRWorkflow{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements webhook.Defaulter
func (w *DRWorkflowWebhook) Default(_ context.Context, obj runtime.Object) error {
	workflow := obj.(*drv1alpha1.DRWorkflow)
	klog.V(4).Infof("Defaulting DRWorkflow: %s/%s", workflow.Namespace, workflow.Name)

	// Set default failure policy
	if workflow.Spec.FailurePolicy == "" {
		workflow.Spec.FailurePolicy = "FailFast"
	}

	// Set default timeout for actions
	for i := range workflow.Spec.Actions {
		action := &workflow.Spec.Actions[i]
		if action.Timeout == "" {
			action.Timeout = "5m"
		}

		// Set default retry policy
		if action.RetryPolicy == nil {
			action.RetryPolicy = &drv1alpha1.RetryPolicy{
				Limit:             3,
				Interval:          "5s",
				BackoffMultiplier: "2.0",
			}
		}

		// Set default operation for action types
		switch action.Type {
		case "HTTP":
			if action.HTTP != nil && action.HTTP.Method == "" {
				action.HTTP.Method = "GET"
			}
		case "Localization":
			if action.Localization != nil {
				if action.Localization.Operation == "" {
					action.Localization.Operation = "Create"
				}
				if action.Localization.Spec != nil && action.Localization.Spec.Priority == 0 {
					action.Localization.Spec.Priority = 500
				}
			}
		case "Subscription":
			if action.Subscription != nil && action.Subscription.Operation == "" {
				action.Subscription.Operation = "Create"
			}
		case "KubernetesResource":
			if action.Resource != nil && action.Resource.Operation == "" {
				action.Resource.Operation = "Create"
			}
		}
	}

	return nil
}

// ValidateCreate implements webhook.Validator
func (w *DRWorkflowWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	workflow := obj.(*drv1alpha1.DRWorkflow)
	klog.Infof("Validating create for DRWorkflow: %s/%s", workflow.Namespace, workflow.Name)

	warnings, errors := w.validateWorkflow(workflow)
	if len(errors) > 0 {
		return warnings, fmt.Errorf("validation failed: %v", errors)
	}

	return warnings, nil
}

// ValidateUpdate implements webhook.Validator
func (w *DRWorkflowWebhook) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	_ = oldObj.(*drv1alpha1.DRWorkflow) // oldWorkflow not used currently
	newWorkflow := newObj.(*drv1alpha1.DRWorkflow)
	klog.Infof("Validating update for DRWorkflow: %s/%s", newWorkflow.Namespace, newWorkflow.Name)

	warnings, errors := w.validateWorkflow(newWorkflow)
	if len(errors) > 0 {
		return warnings, fmt.Errorf("validation failed: %v", errors)
	}

	return warnings, nil
}

// ValidateDelete implements webhook.Validator
func (w *DRWorkflowWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	workflow := obj.(*drv1alpha1.DRWorkflow)
	klog.Infof("Validating delete for DRWorkflow: %s/%s", workflow.Namespace, workflow.Name)

	// Check if this workflow is referenced by any DRPlan
	referencingPlans, err := w.findReferencingPlans(ctx, workflow)
	if err != nil {
		klog.Errorf("Failed to check referencing plans for DRWorkflow %s/%s: %v", workflow.Namespace, workflow.Name, err)
		return nil, fmt.Errorf("failed to check references: %w", err)
	}

	if len(referencingPlans) > 0 {
		planNames := make([]string, len(referencingPlans))
		for i, plan := range referencingPlans {
			planNames[i] = fmt.Sprintf("%s/%s", plan.Namespace, plan.Name)
		}
		return []string{fmt.Sprintf("Workflow is referenced by %d plan(s)", len(referencingPlans))},
			fmt.Errorf("cannot delete DRWorkflow %s/%s: referenced by DRPlan(s): %v",
				workflow.Namespace, workflow.Name, planNames)
	}

	return nil, nil
}

// findReferencingPlans finds all DRPlans that reference this workflow
func (w *DRWorkflowWebhook) findReferencingPlans(ctx context.Context, workflow *drv1alpha1.DRWorkflow) ([]*drv1alpha1.DRPlan, error) {
	// List all DRPlans in the same namespace
	planList := &drv1alpha1.DRPlanList{}
	if err := w.Client.List(ctx, planList, client.InNamespace(workflow.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list DRPlans: %w", err)
	}

	var referencingPlans []*drv1alpha1.DRPlan
	for i := range planList.Items {
		plan := &planList.Items[i]

		// Check if this plan references the workflow
		for _, stage := range plan.Spec.Stages {
			for _, wfRef := range stage.Workflows {
				if wfRef.WorkflowRef.Name == workflow.Name &&
					(wfRef.WorkflowRef.Namespace == "" || wfRef.WorkflowRef.Namespace == workflow.Namespace) {
					referencingPlans = append(referencingPlans, plan)
					goto nextPlan // Found reference in this plan, move to next plan
				}
			}
		}
	nextPlan:
	}

	return referencingPlans, nil
}

// validateWorkflow performs comprehensive validation
func (w *DRWorkflowWebhook) validateWorkflow(workflow *drv1alpha1.DRWorkflow) ([]string, []string) {
	var warnings []string
	var errors []string

	// Validate actions
	if len(workflow.Spec.Actions) == 0 {
		errors = append(errors, "at least one action is required")
		return warnings, errors
	}

	// Track action names for uniqueness check
	actionNames := make(map[string]bool)

	for i, action := range workflow.Spec.Actions {
		// Check unique action names
		if actionNames[action.Name] {
			errors = append(errors, fmt.Sprintf("action[%d]: duplicate action name '%s'", i, action.Name))
		}
		actionNames[action.Name] = true

		// Validate action name
		if action.Name == "" {
			errors = append(errors, fmt.Sprintf("action[%d]: name is required", i))
		}

		// Validate action type and configuration
		actionErrors := w.validateAction(&action, i)
		errors = append(errors, actionErrors...)

		// Validate rollback requirement for Patch operations
		rollbackWarnings, rollbackErrors := w.validateRollbackRequirement(&action, i)
		warnings = append(warnings, rollbackWarnings...)
		errors = append(errors, rollbackErrors...)
	}

	// Validate parameters
	paramNames := make(map[string]bool)
	for i, param := range workflow.Spec.Parameters {
		if param.Name == "" {
			errors = append(errors, fmt.Sprintf("parameter[%d]: name is required", i))
		}
		if paramNames[param.Name] {
			errors = append(errors, fmt.Sprintf("parameter[%d]: duplicate parameter name '%s'", i, param.Name))
		}
		paramNames[param.Name] = true
	}

	return warnings, errors
}

// validateAction validates action type-specific configuration
func (w *DRWorkflowWebhook) validateAction(action *drv1alpha1.Action, index int) []string {
	var errors []string

	switch action.Type {
	case "HTTP":
		if action.HTTP == nil {
			errors = append(errors, fmt.Sprintf("action[%d] '%s': HTTP configuration is required", index, action.Name))
		} else if action.HTTP.URL == "" {
			errors = append(errors, fmt.Sprintf("action[%d] '%s': HTTP.URL is required", index, action.Name))
		}
	case "Job":
		if action.Job == nil {
			errors = append(errors, fmt.Sprintf("action[%d] '%s': Job configuration is required", index, action.Name))
		}
	case "Localization":
		errors = append(errors, w.validateLocalization(action, index)...)
	case "Subscription":
		errors = append(errors, w.validateSubscription(action, index)...)
	case "KubernetesResource":
		errors = append(errors, w.validateKubernetesResource(action, index)...)
	default:
		errors = append(errors, fmt.Sprintf("action[%d] '%s': unsupported action type '%s'", index, action.Name, action.Type))
	}

	return errors
}

// validateLocalization validates Localization action
func (w *DRWorkflowWebhook) validateLocalization(action *drv1alpha1.Action, index int) []string {
	var errors []string

	if action.Localization == nil {
		return append(errors, fmt.Sprintf("action[%d] '%s': Localization configuration is required", index, action.Name))
	}

	loc := action.Localization
	if loc.Name == "" {
		errors = append(errors, fmt.Sprintf("action[%d] '%s': Localization.Name is required", index, action.Name))
	}
	if loc.Namespace == "" {
		errors = append(errors, fmt.Sprintf("action[%d] '%s': Localization.Namespace is required", index, action.Name))
	}

	// Validate operation-specific requirements
	if loc.Operation == "Create" {
		if loc.Spec == nil {
			errors = append(errors, fmt.Sprintf("action[%d] '%s': Localization.Spec is required when operation=Create", index, action.Name))
		} else if loc.Spec.APIVersion == "" || loc.Spec.Kind == "" || loc.Spec.Name == "" {
			errors = append(errors, fmt.Sprintf("action[%d] '%s': Localization.Spec.Feed (apiVersion, kind, name) is required when operation=Create", index, action.Name))
		}
	}

	return errors
}

// validateSubscription validates Subscription action
func (w *DRWorkflowWebhook) validateSubscription(action *drv1alpha1.Action, index int) []string {
	var errors []string

	if action.Subscription == nil {
		return append(errors, fmt.Sprintf("action[%d] '%s': Subscription configuration is required", index, action.Name))
	}

	sub := action.Subscription
	if sub.Name == "" {
		errors = append(errors, fmt.Sprintf("action[%d] '%s': Subscription.Name is required", index, action.Name))
	}

	// Validate Spec when operation is Create
	if sub.Operation == "Create" {
		if sub.Spec == nil {
			errors = append(errors, fmt.Sprintf("action[%d] '%s': Subscription.Spec is required when operation=Create", index, action.Name))
		} else {
			if len(sub.Spec.Feeds) == 0 {
				errors = append(errors, fmt.Sprintf("action[%d] '%s': Subscription.Spec.Feeds is required (at least one feed)", index, action.Name))
			}
			if len(sub.Spec.Subscribers) == 0 {
				errors = append(errors, fmt.Sprintf("action[%d] '%s': Subscription.Spec.Subscribers is required (at least one subscriber)", index, action.Name))
			}
		}
	}

	return errors
}

// validateKubernetesResource validates KubernetesResource action
func (w *DRWorkflowWebhook) validateKubernetesResource(action *drv1alpha1.Action, index int) []string {
	var errors []string

	if action.Resource == nil {
		return append(errors, fmt.Sprintf("action[%d] '%s': KubernetesResource configuration is required", index, action.Name))
	}

	if action.Resource.Manifest == "" {
		errors = append(errors, fmt.Sprintf("action[%d] '%s': KubernetesResource.Manifest is required", index, action.Name))
	}

	return errors
}

// rollbackRule defines rollback validation rules for an action type
type rollbackRule struct {
	checkRequired  func(action *drv1alpha1.Action) (bool, string) // returns (required, operationType)
	checkAutomatic func(action *drv1alpha1.Action) bool           // returns true if has automatic rollback
}

// getRollbackRules returns rollback validation rules for all action types
func getRollbackRules() map[string]rollbackRule {
	return map[string]rollbackRule{
		"Localization": {
			checkRequired: func(action *drv1alpha1.Action) (bool, string) {
				if action.Localization != nil && action.Localization.Operation == "Patch" {
					return true, "Localization Patch"
				}
				return false, ""
			},
			checkAutomatic: func(action *drv1alpha1.Action) bool {
				return action.Localization != nil && action.Localization.Operation == "Create"
			},
		},
		"Subscription": {
			checkRequired: func(action *drv1alpha1.Action) (bool, string) {
				if action.Subscription != nil && action.Subscription.Operation == "Patch" {
					return true, "Subscription Patch"
				}
				return false, ""
			},
			checkAutomatic: func(action *drv1alpha1.Action) bool {
				return action.Subscription != nil && action.Subscription.Operation == "Create"
			},
		},
		"KubernetesResource": {
			checkRequired: func(action *drv1alpha1.Action) (bool, string) {
				if action.Resource != nil && (action.Resource.Operation == "Patch" || action.Resource.Operation == "Apply") {
					return true, fmt.Sprintf("KubernetesResource %s", action.Resource.Operation)
				}
				return false, ""
			},
			checkAutomatic: func(action *drv1alpha1.Action) bool {
				return action.Resource != nil && action.Resource.Operation == "Create"
			},
		},
		"Job": {
			checkRequired: func(_ *drv1alpha1.Action) (bool, string) {
				return false, ""
			},
			checkAutomatic: func(_ *drv1alpha1.Action) bool {
				return true // Job always has automatic rollback
			},
		},
	}
}

// validateRollbackRequirement validates that Patch operations have rollback defined
// Refactored to reduce cyclomatic complexity by using rule-based validation
func (w *DRWorkflowWebhook) validateRollbackRequirement(action *drv1alpha1.Action, index int) ([]string, []string) {
	var warnings []string
	var errors []string

	rules := getRollbackRules()
	rule, exists := rules[action.Type]
	if !exists {
		// No rollback rules for this action type
		return warnings, errors
	}

	// Check if rollback is required
	if required, operationType := rule.checkRequired(action); required && action.Rollback == nil {
		errors = append(errors, fmt.Sprintf("action[%d] '%s': rollback is required for %s operation",
			index, action.Name, operationType))
	}

	// Warn if using automatic rollback
	if action.Rollback == nil && rule.checkAutomatic(action) {
		warnings = append(warnings, fmt.Sprintf("action[%d] '%s': no custom rollback defined, will use automatic rollback (delete resource)",
			index, action.Name))
	}

	return warnings, errors
}
