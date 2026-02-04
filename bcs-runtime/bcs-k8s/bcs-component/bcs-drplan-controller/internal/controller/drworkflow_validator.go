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
	"fmt"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// ActionValidator interface for validating specific action types
type ActionValidator interface {
	Validate(action *drv1alpha1.Action, index int) []string
}

// HTTPActionValidator validates HTTP actions
type HTTPActionValidator struct{}

// Validate validates HTTP action configuration
func (v *HTTPActionValidator) Validate(action *drv1alpha1.Action, index int) []string {
	var errors []string
	if action.HTTP == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: HTTP configuration is required", index, action.Name))
		return errors
	}
	if action.HTTP.URL == "" {
		errors = append(errors, fmt.Sprintf("action[%d] %s: HTTP.URL is required", index, action.Name))
	}
	return errors
}

// JobActionValidator validates Job actions
type JobActionValidator struct{}

// Validate validates Job action configuration
func (v *JobActionValidator) Validate(action *drv1alpha1.Action, index int) []string {
	var errors []string
	if action.Job == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: Job configuration is required", index, action.Name))
	}
	return errors
}

// LocalizationActionValidator validates Localization actions
type LocalizationActionValidator struct{}

// Validate validates Localization action configuration
func (v *LocalizationActionValidator) Validate(action *drv1alpha1.Action, index int) []string {
	var errors []string

	if action.Localization == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: Localization configuration is required", index, action.Name))
		return errors
	}

	loc := action.Localization
	if loc.Name == "" {
		errors = append(errors, fmt.Sprintf("action[%d] %s: Localization.Name is required", index, action.Name))
	}
	if loc.Namespace == "" {
		errors = append(errors, fmt.Sprintf("action[%d] %s: Localization.Namespace is required", index, action.Name))
	}
	if loc.Operation == "Create" {
		if loc.Spec == nil {
			errors = append(errors, fmt.Sprintf("action[%d] %s: Localization.Spec is required when operation=Create", index, action.Name))
		} else if loc.Spec.APIVersion == "" || loc.Spec.Kind == "" || loc.Spec.Name == "" {
			errors = append(errors, fmt.Sprintf("action[%d] %s: Localization.Spec.Feed (apiVersion, kind, name) is required when operation=Create", index, action.Name))
		}
	}
	if loc.Operation == "Patch" && action.Rollback == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: rollback is required for Localization Patch operation", index, action.Name))
	}

	return errors
}

// SubscriptionActionValidator validates Subscription actions
type SubscriptionActionValidator struct{}

// Validate validates Subscription action configuration
func (v *SubscriptionActionValidator) Validate(action *drv1alpha1.Action, index int) []string {
	var errors []string

	if action.Subscription == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: Subscription configuration is required", index, action.Name))
		return errors
	}

	if action.Subscription.Name == "" {
		errors = append(errors, fmt.Sprintf("action[%d] %s: Subscription.Name is required", index, action.Name))
	}
	if action.Subscription.Operation == "Patch" && action.Rollback == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: rollback is required for Subscription Patch operation", index, action.Name))
	}
	if action.Subscription.Operation == "Create" {
		if action.Subscription.Spec == nil {
			errors = append(errors, fmt.Sprintf("action[%d] %s: Subscription.Spec is required when operation=Create", index, action.Name))
		} else {
			if len(action.Subscription.Spec.Feeds) == 0 {
				errors = append(errors, fmt.Sprintf("action[%d] %s: Subscription.Spec.Feeds is required (at least one)", index, action.Name))
			}
			if len(action.Subscription.Spec.Subscribers) == 0 {
				errors = append(errors, fmt.Sprintf("action[%d] %s: Subscription.Spec.Subscribers is required (at least one)", index, action.Name))
			}
		}
	}

	return errors
}

// KubernetesResourceActionValidator validates KubernetesResource actions
type KubernetesResourceActionValidator struct{}

// Validate validates KubernetesResource action configuration
func (v *KubernetesResourceActionValidator) Validate(action *drv1alpha1.Action, index int) []string {
	var errors []string

	if action.Resource == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: KubernetesResource configuration is required", index, action.Name))
		return errors
	}

	if action.Resource.Manifest == "" {
		errors = append(errors, fmt.Sprintf("action[%d] %s: KubernetesResource.Manifest is required", index, action.Name))
	}

	needsRollback := action.Resource.Operation == "Patch" || action.Resource.Operation == "Apply"
	if needsRollback && action.Rollback == nil {
		errors = append(errors, fmt.Sprintf("action[%d] %s: rollback is required for KubernetesResource %s operation",
			index, action.Name, action.Resource.Operation))
	}

	return errors
}

// ActionValidatorRegistry manages action validators
type ActionValidatorRegistry struct {
	validators map[string]ActionValidator
}

// NewActionValidatorRegistry creates a new validator registry
func NewActionValidatorRegistry() *ActionValidatorRegistry {
	registry := &ActionValidatorRegistry{
		validators: make(map[string]ActionValidator),
	}

	// Register validators
	registry.validators["HTTP"] = &HTTPActionValidator{}
	registry.validators["Job"] = &JobActionValidator{}
	registry.validators["Localization"] = &LocalizationActionValidator{}
	registry.validators["Subscription"] = &SubscriptionActionValidator{}
	registry.validators["KubernetesResource"] = &KubernetesResourceActionValidator{}

	return registry
}

// ValidateAction validates an action using the registered validator
func (r *ActionValidatorRegistry) ValidateAction(action *drv1alpha1.Action, index int) []string {
	validator, exists := r.validators[action.Type]
	if !exists {
		return []string{fmt.Sprintf("action[%d] %s: unsupported action type: %s", index, action.Name, action.Type)}
	}
	return validator.Validate(action, index)
}

// validateActionNames checks for duplicate action names
func validateActionNames(actions []drv1alpha1.Action) []string {
	var errors []string
	actionNames := make(map[string]bool)

	for _, action := range actions {
		if actionNames[action.Name] {
			errors = append(errors, fmt.Sprintf("duplicate action name: %s", action.Name))
		}
		actionNames[action.Name] = true
	}

	return errors
}

// validateParameters checks for duplicate parameter names
func validateParameters(parameters []drv1alpha1.Parameter) []string {
	var errors []string
	paramNames := make(map[string]bool)

	for i, param := range parameters {
		if paramNames[param.Name] {
			errors = append(errors, fmt.Sprintf("duplicate parameter name: %s", param.Name))
		}
		paramNames[param.Name] = true

		if param.Name == "" {
			errors = append(errors, fmt.Sprintf("parameter[%d]: name is required", i))
		}
	}

	return errors
}
