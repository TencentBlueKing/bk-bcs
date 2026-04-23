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
	sigyaml "sigs.k8s.io/yaml"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

var helmChartActionGVK = schema.GroupVersionKind{
	Group:   "apps.clusternet.io",
	Version: "v1alpha1",
	Kind:    "HelmChart",
}

// HelmChartActionExecutor implements ActionExecutor for HelmChart actions.
type HelmChartActionExecutor struct {
	client client.Client
}

// NewHelmChartActionExecutor creates a new HelmChart action executor.
func NewHelmChartActionExecutor(client client.Client) *HelmChartActionExecutor {
	return &HelmChartActionExecutor{client: client}
}

// Execute executes a HelmChart action.
// The executor always renders name/namespace first so both templated delete
// and templated apply/patch paths resolve to the same target object key.
func (e *HelmChartActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing HelmChart action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     drv1alpha1.PhaseRunning,
		StartTime: &metav1.Time{Time: startTime},
	}

	if err := e.validateHelmChartConfig(action); err != nil {
		return failHelmChartStatus(status, err.Error()), err
	}

	templateData := &utils.TemplateData{Params: params}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	name, err := render(action.HelmChart.Name)
	if err != nil {
		return failHelmChartStatus(status, fmt.Sprintf("failed to render HelmChart name: %v", err)), err
	}
	namespace, err := render(action.HelmChart.Namespace)
	if err != nil {
		return failHelmChartStatus(status, fmt.Sprintf("failed to render HelmChart namespace: %v", err)), err
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(helmChartActionGVK)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	operation := effectiveHelmChartOperation(action.HelmChart.Operation)
	if operation != drv1alpha1.OperationDelete {
		specMap, renderErr := e.renderHelmChartSpec(action.HelmChart, render)
		if renderErr != nil {
			return failHelmChartStatus(status, renderErr.Error()), renderErr
		}
		obj.Object["spec"] = specMap
	}

	if err := e.applyHelmChart(ctx, obj, operation); err != nil {
		return failHelmChartStatus(status, err.Error()), err
	}

	status.Outputs = &drv1alpha1.ActionOutputs{
		HelmChartRef: &corev1.ObjectReference{
			Kind:       helmChartActionGVK.Kind,
			APIVersion: helmChartActionGVK.GroupVersion().String(),
			Namespace:  obj.GetNamespace(),
			Name:       obj.GetName(),
			UID:        obj.GetUID(),
		},
	}
	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("HelmChart %s/%s %s successfully", obj.GetNamespace(), obj.GetName(), helmChartActionVerb(operation))

	return status, nil
}

func (e *HelmChartActionExecutor) validateHelmChartConfig(action *drv1alpha1.Action) error {
	if action.HelmChart == nil {
		return fmt.Errorf("helmChart configuration is required")
	}
	if action.HelmChart.Name == "" {
		return fmt.Errorf("HelmChart.Name is required")
	}
	if action.HelmChart.Namespace == "" {
		return fmt.Errorf("HelmChart.Namespace is required")
	}

	operation := effectiveHelmChartOperation(action.HelmChart.Operation)
	if operation == drv1alpha1.OperationDelete {
		return nil
	}
	if action.HelmChart.Spec == nil {
		return fmt.Errorf("HelmChart.Spec is required")
	}
	if operation != drv1alpha1.OperationPatch {
		if action.HelmChart.Spec.Repository == "" {
			return fmt.Errorf("HelmChart.Spec.Repo is required")
		}
		if action.HelmChart.Spec.Chart == "" {
			return fmt.Errorf("HelmChart.Spec.Chart is required")
		}
		if action.HelmChart.Spec.TargetNamespace == "" {
			return fmt.Errorf("HelmChart.Spec.TargetNamespace is required")
		}
	}
	return nil
}

func (e *HelmChartActionExecutor) renderHelmChartSpec(
	action *drv1alpha1.HelmChartAction,
	render func(string) (string, error),
) (map[string]interface{}, error) {
	// Render the strongly typed spec through YAML instead of field-by-field
	// assignment so new HelmChart fields automatically inherit template support.
	specBytes, err := sigyaml.Marshal(action.Spec)
	if err != nil {
		return nil, fmt.Errorf("marshal HelmChart spec: %w", err)
	}

	renderedSpec, err := render(string(specBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to render HelmChart spec: %w", err)
	}

	specMap := map[string]interface{}{}
	if err := sigyaml.Unmarshal([]byte(renderedSpec), &specMap); err != nil {
		return nil, fmt.Errorf("unmarshal rendered HelmChart spec: %w", err)
	}

	return specMap, nil
}

// applyHelmChart keeps operation dispatch in one place so Execute and Rollback
// can share the same create/apply/patch/delete semantics.
func (e *HelmChartActionExecutor) applyHelmChart(ctx context.Context, obj *unstructured.Unstructured, operation string) error {
	switch operation {
	case drv1alpha1.OperationDelete:
		return client.IgnoreNotFound(e.client.Delete(ctx, obj))
	case drv1alpha1.OperationApply:
		return e.client.Patch(ctx, obj, client.Apply, client.FieldOwner("drplan-controller"), client.ForceOwnership)
	case drv1alpha1.OperationPatch:
		return e.patchHelmChart(ctx, obj)
	default:
		return e.client.Create(ctx, obj)
	}
}

// patchHelmChart performs a shallow strategic merge on spec content.
// Patch mode is intentionally "update only provided fields" rather than SSA.
func (e *HelmChartActionExecutor) patchHelmChart(ctx context.Context, desired *unstructured.Unstructured) error {
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(helmChartActionGVK)
	if err := e.client.Get(ctx, client.ObjectKey{Name: desired.GetName(), Namespace: desired.GetNamespace()}, existing); err != nil {
		return fmt.Errorf("get existing HelmChart %s/%s: %w", desired.GetNamespace(), desired.GetName(), err)
	}

	base := existing.DeepCopy()
	existingSpec, found, err := unstructured.NestedMap(existing.Object, "spec")
	if err != nil {
		return fmt.Errorf("read existing HelmChart spec: %w", err)
	}
	if !found {
		existingSpec = map[string]interface{}{}
	}

	desiredSpec, found, err := unstructured.NestedMap(desired.Object, "spec")
	if err != nil {
		return fmt.Errorf("read desired HelmChart spec: %w", err)
	}
	if found {
		mergeNestedMaps(existingSpec, desiredSpec)
	}
	existing.Object["spec"] = existingSpec

	if err := e.client.Patch(ctx, existing, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("patch HelmChart %s/%s: %w", desired.GetNamespace(), desired.GetName(), err)
	}
	desired.SetUID(existing.GetUID())
	return nil
}

// Rollback rolls back a HelmChart action.
// By default only resources created in the forward Create path are deleted.
// Apply/Patch need explicit rollback actions because they may target pre-existing objects.
func (e *HelmChartActionExecutor) Rollback(
	ctx context.Context,
	action *drv1alpha1.Action,
	actionStatus *drv1alpha1.ActionStatus,
	params map[string]interface{},
) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back HelmChart action: %s", action.Name)

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

	if action.HelmChart != nil &&
		effectiveHelmChartOperation(action.HelmChart.Operation) == drv1alpha1.OperationCreate &&
		actionStatus.Outputs != nil &&
		actionStatus.Outputs.HelmChartRef != nil {
		chart := &unstructured.Unstructured{}
		chart.SetGroupVersionKind(helmChartActionGVK)
		chart.SetName(actionStatus.Outputs.HelmChartRef.Name)
		chart.SetNamespace(actionStatus.Outputs.HelmChartRef.Namespace)

		if err := e.client.Delete(ctx, chart); client.IgnoreNotFound(err) != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to delete HelmChart: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, fmt.Errorf("failed to delete HelmChart: %w", err)
		}

		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted HelmChart %s/%s", chart.GetNamespace(), chart.GetName())
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		return rollbackStatus, nil
	}

	rollbackStatus.Phase = drv1alpha1.PhaseSkipped
	rollbackStatus.Message = "No HelmChart to rollback"
	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	return rollbackStatus, nil
}

// Type returns the action type.
func (e *HelmChartActionExecutor) Type() string {
	return drv1alpha1.ActionTypeHelmChart
}

// failHelmChartStatus centralizes failure status shaping so every error path
// reports a completion timestamp and user-facing message consistently.
func failHelmChartStatus(status *drv1alpha1.ActionStatus, message string) *drv1alpha1.ActionStatus {
	status.Phase = drv1alpha1.PhaseFailed
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = message
	return status
}

// effectiveHelmChartOperation preserves backwards compatibility with old
// workflows that omitted operation and implicitly meant Create.
func effectiveHelmChartOperation(operation string) string {
	if operation == "" {
		return drv1alpha1.OperationCreate
	}
	return operation
}

// helmChartActionVerb is only for status messaging; operation dispatch stays in
// applyHelmChart/effectiveHelmChartOperation to avoid stringly typed behavior.
func helmChartActionVerb(operation string) string {
	switch operation {
	case drv1alpha1.OperationDelete:
		return "deleted"
	case drv1alpha1.OperationApply:
		return "applied"
	case drv1alpha1.OperationPatch:
		return "patched"
	default:
		return "created"
	}
}
