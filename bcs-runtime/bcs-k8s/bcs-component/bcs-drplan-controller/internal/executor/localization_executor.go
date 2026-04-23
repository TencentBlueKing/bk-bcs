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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

var localizationActionGVK = schema.GroupVersionKind{
	Group:   "apps.clusternet.io",
	Version: "v1alpha1",
	Kind:    "Localization",
}

// LocalizationActionExecutor implements ActionExecutor for Localization actions
type LocalizationActionExecutor struct {
	client client.Client
}

// NewLocalizationActionExecutor creates a new Localization action executor
func NewLocalizationActionExecutor(client client.Client) *LocalizationActionExecutor {
	return &LocalizationActionExecutor{client: client}
}

// Execute executes a Localization action
// Localization and Globalization intentionally share the same execution model:
// resolve target identity first, render spec only for non-delete operations,
// then dispatch to create/apply/patch/delete.
func (e *LocalizationActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing Localization action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     drv1alpha1.PhaseRunning,
		StartTime: &metav1.Time{Time: startTime},
	}

	if err := e.validateLocalizationConfig(action); err != nil {
		return failLocalizationStatus(status, err.Error()), err
	}

	templateData := &utils.TemplateData{Params: params}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	name, err := render(action.Localization.Name)
	if err != nil {
		return failLocalizationStatus(status, fmt.Sprintf("failed to render Localization name: %v", err)), err
	}
	namespace, err := render(action.Localization.Namespace)
	if err != nil {
		return failLocalizationStatus(status, fmt.Sprintf("failed to render Localization namespace: %v", err)), err
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(localizationActionGVK)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	operation := effectiveLocalizationOperation(action.Localization.Operation)
	if operation != drv1alpha1.OperationDelete {
		specMap, renderErr := e.renderLocalizationSpec(action.Localization, operation, render)
		if renderErr != nil {
			return failLocalizationStatus(status, renderErr.Error()), renderErr
		}
		obj.Object["spec"] = specMap
	}

	if err := e.applyLocalization(ctx, obj, operation); err != nil {
		return failLocalizationStatus(status, err.Error()), err
	}

	status.Outputs = &drv1alpha1.ActionOutputs{
		LocalizationRef: &corev1.ObjectReference{
			Kind:       localizationActionGVK.Kind,
			APIVersion: localizationActionGVK.GroupVersion().String(),
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
			UID:        obj.GetUID(),
		},
	}
	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("Localization %s/%s %s successfully", obj.GetNamespace(), obj.GetName(), localizationActionVerb(operation))

	return status, nil
}

func (e *LocalizationActionExecutor) validateLocalizationConfig(action *drv1alpha1.Action) error {
	if action.Localization == nil {
		return fmt.Errorf("localization configuration is required")
	}
	if action.Localization.Name == "" {
		return fmt.Errorf("Localization.Name is required")
	}
	if action.Localization.Namespace == "" {
		return fmt.Errorf("Localization.Namespace is required")
	}

	operation := effectiveLocalizationOperation(action.Localization.Operation)
	if operation == drv1alpha1.OperationDelete {
		return nil
	}
	if action.Localization.Spec == nil {
		return fmt.Errorf("Localization.Spec is required")
	}
	if operation != drv1alpha1.OperationPatch &&
		(action.Localization.Spec.APIVersion == "" || action.Localization.Spec.Kind == "" || action.Localization.Spec.Name == "") {
		return fmt.Errorf("Localization.Spec.Feed (apiVersion, kind, name) is required")
	}
	return nil
}

func (e *LocalizationActionExecutor) renderLocalizationSpec(
	action *drv1alpha1.LocalizationAction,
	operation string,
	render func(string) (string, error),
) (map[string]interface{}, error) {
	// Patch keeps feed optional so callers can change only priority/overrides
	// without resending the full spec payload.
	spec := action.Spec
	specMap := map[string]interface{}{}

	if spec.OverridePolicy != "" {
		policy, err := render(string(spec.OverridePolicy))
		if err != nil {
			return nil, fmt.Errorf("failed to render Localization overridePolicy: %w", err)
		}
		specMap["overridePolicy"] = policy
	}

	if len(spec.Overrides) > 0 {
		overrideList := make([]interface{}, 0, len(spec.Overrides))
		for i := range spec.Overrides {
			override, err := renderOverrideConfig(&spec.Overrides[i], i, render)
			if err != nil {
				return nil, err
			}
			overrideList = append(overrideList, override)
		}
		specMap["overrides"] = overrideList
	}

	if spec.Priority != 0 || operation != drv1alpha1.OperationPatch {
		specMap["priority"] = int64(spec.Priority)
	}

	if spec.APIVersion != "" || spec.Kind != "" || spec.Name != "" || spec.Namespace != "" {
		feed, err := renderFeed(spec.Feed, render)
		if err != nil {
			return nil, err
		}
		specMap["feed"] = feed
	}

	return specMap, nil
}

// applyLocalization keeps operation routing explicit and matches the webhook /
// validator semantics: Apply uses SSA, Patch merges existing spec, Delete is idempotent.
func (e *LocalizationActionExecutor) applyLocalization(ctx context.Context, obj *unstructured.Unstructured, operation string) error {
	switch operation {
	case drv1alpha1.OperationDelete:
		return client.IgnoreNotFound(e.client.Delete(ctx, obj))
	case drv1alpha1.OperationApply:
		return e.client.Patch(ctx, obj, client.Apply, client.FieldOwner("drplan-controller"), client.ForceOwnership)
	case drv1alpha1.OperationPatch:
		return e.patchLocalization(ctx, obj)
	default:
		return e.client.Create(ctx, obj)
	}
}

// patchLocalization merges desired spec fields into the current object.
// This mirrors the user expectation of "partial update" for DRWorkflow Patch actions.
func (e *LocalizationActionExecutor) patchLocalization(ctx context.Context, desired *unstructured.Unstructured) error {
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(localizationActionGVK)
	key := client.ObjectKey{Name: desired.GetName(), Namespace: desired.GetNamespace()}
	if err := e.client.Get(ctx, key, existing); err != nil {
		return fmt.Errorf("get existing Localization %s/%s: %w", desired.GetNamespace(), desired.GetName(), err)
	}

	base := existing.DeepCopy()
	existingSpec, found, err := unstructured.NestedMap(existing.Object, "spec")
	if err != nil {
		return fmt.Errorf("read existing Localization spec: %w", err)
	}
	if !found {
		existingSpec = map[string]interface{}{}
	}

	desiredSpec, found, err := unstructured.NestedMap(desired.Object, "spec")
	if err != nil {
		return fmt.Errorf("read desired Localization spec: %w", err)
	}
	if found {
		mergeNestedMaps(existingSpec, desiredSpec)
	}
	existing.Object["spec"] = existingSpec

	if err := e.client.Patch(ctx, existing, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("patch Localization %s/%s: %w", desired.GetNamespace(), desired.GetName(), err)
	}
	desired.SetUID(existing.GetUID())
	return nil
}

// Rollback rolls back a Localization action by deleting the CR
// Default rollback only deletes resources that were created by the forward Create path.
// Apply/Patch remain opt-in via explicit rollback actions because they may mutate existing objects.
func (e *LocalizationActionExecutor) Rollback(
	ctx context.Context,
	action *drv1alpha1.Action,
	actionStatus *drv1alpha1.ActionStatus,
	params map[string]interface{},
) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back Localization action: %s", action.Name)

	rollbackStatus := &drv1alpha1.ActionStatus{
		Name:      actionStatus.Name,
		Phase:     drv1alpha1.PhaseRunning,
		StartTime: &metav1.Time{Time: time.Now()},
	}

	if action.Rollback != nil {
		customStatus, err := e.Execute(ctx, action.Rollback, params)
		if err != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Custom rollback failed: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, err
		}
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = drv1alpha1.MessageRollbackSuccess
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		rollbackStatus.Outputs = customStatus.Outputs
		return rollbackStatus, nil
	}

	if action.Localization != nil &&
		effectiveLocalizationOperation(action.Localization.Operation) == drv1alpha1.OperationCreate &&
		actionStatus.Outputs != nil &&
		actionStatus.Outputs.LocalizationRef != nil {
		loc := &unstructured.Unstructured{}
		loc.SetGroupVersionKind(localizationActionGVK)
		loc.SetName(actionStatus.Outputs.LocalizationRef.Name)
		loc.SetNamespace(actionStatus.Outputs.LocalizationRef.Namespace)

		if err := e.client.Delete(ctx, loc); client.IgnoreNotFound(err) != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to delete Localization: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, fmt.Errorf("failed to delete Localization: %w", err)
		}

		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted Localization %s/%s", loc.GetNamespace(), loc.GetName())
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		return rollbackStatus, nil
	}

	rollbackStatus.Phase = drv1alpha1.PhaseSkipped
	rollbackStatus.Message = "No Localization to rollback"
	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	return rollbackStatus, nil
}

// Type returns the action type this executor handles
func (e *LocalizationActionExecutor) Type() string {
	return drv1alpha1.ActionTypeLocalization
}

// failLocalizationStatus keeps failure handling uniform across render, validation,
// CRUD, and rollback branches.
func failLocalizationStatus(status *drv1alpha1.ActionStatus, message string) *drv1alpha1.ActionStatus {
	status.Phase = drv1alpha1.PhaseFailed
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = message
	return status
}

// effectiveLocalizationOperation preserves the historical "empty means Create"
// behavior for older workflows and generated manifests.
func effectiveLocalizationOperation(operation string) string {
	if operation == "" {
		return drv1alpha1.OperationCreate
	}
	return operation
}

// localizationActionVerb is restricted to status messages so callers do not
// accidentally depend on it for real operation branching.
func localizationActionVerb(operation string) string {
	switch operation {
	case drv1alpha1.OperationDelete:
		return actionVerbDeleted
	case drv1alpha1.OperationApply:
		return actionVerbApplied
	case drv1alpha1.OperationPatch:
		return actionVerbPatched
	default:
		return actionVerbCreated
	}
}
