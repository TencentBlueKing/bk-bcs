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

	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

var globalizationActionGVK = schema.GroupVersionKind{
	Group:   "apps.clusternet.io",
	Version: "v1alpha1",
	Kind:    "Globalization",
}

const (
	actionVerbDeleted = "deleted"
	actionVerbApplied = "applied"
	actionVerbPatched = "patched"
	actionVerbCreated = "created"
)

// GlobalizationActionExecutor implements ActionExecutor for Globalization actions.
type GlobalizationActionExecutor struct {
	client client.Client
}

// NewGlobalizationActionExecutor creates a new Globalization action executor.
func NewGlobalizationActionExecutor(client client.Client) *GlobalizationActionExecutor {
	return &GlobalizationActionExecutor{client: client}
}

// Execute executes a Globalization action.
func (e *GlobalizationActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing Globalization action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     drv1alpha1.PhaseRunning,
		StartTime: &metav1.Time{Time: startTime},
	}

	if err := e.validateGlobalizationConfig(action); err != nil {
		return failGlobalizationStatus(status, err.Error()), err
	}

	templateData := &utils.TemplateData{Params: params}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	name, err := render(action.Globalization.Name)
	if err != nil {
		return failGlobalizationStatus(status, fmt.Sprintf("failed to render Globalization name: %v", err)), err
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(globalizationActionGVK)
	obj.SetName(name)

	operation := effectiveGlobalizationOperation(action.Globalization.Operation)
	if operation != drv1alpha1.OperationDelete {
		specMap, renderErr := e.renderGlobalizationSpec(action.Globalization, operation, render)
		if renderErr != nil {
			return failGlobalizationStatus(status, renderErr.Error()), renderErr
		}
		obj.Object["spec"] = specMap
	}

	if err := e.applyGlobalization(ctx, obj, operation); err != nil {
		return failGlobalizationStatus(status, err.Error()), err
	}

	status.Outputs = &drv1alpha1.ActionOutputs{
		GlobalizationRef: &corev1.ObjectReference{
			Kind:       globalizationActionGVK.Kind,
			APIVersion: globalizationActionGVK.GroupVersion().String(),
			Name:       obj.GetName(),
			UID:        obj.GetUID(),
		},
	}
	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("Globalization %s %s successfully", obj.GetName(), globalizationActionVerb(operation))

	return status, nil
}

func (e *GlobalizationActionExecutor) validateGlobalizationConfig(action *drv1alpha1.Action) error {
	if action.Globalization == nil {
		return fmt.Errorf("globalization configuration is required")
	}
	if action.Globalization.Name == "" {
		return fmt.Errorf("Globalization.Name is required")
	}

	operation := effectiveGlobalizationOperation(action.Globalization.Operation)
	if operation == drv1alpha1.OperationDelete {
		return nil
	}
	if action.Globalization.Spec == nil {
		return fmt.Errorf("Globalization.Spec is required")
	}
	if operation != drv1alpha1.OperationPatch &&
		(action.Globalization.Spec.APIVersion == "" || action.Globalization.Spec.Kind == "" || action.Globalization.Spec.Name == "") {
		return fmt.Errorf("Globalization.Spec.Feed (apiVersion, kind, name) is required")
	}
	return nil
}

func (e *GlobalizationActionExecutor) renderGlobalizationSpec(
	action *drv1alpha1.GlobalizationAction,
	operation string,
	render func(string) (string, error),
) (map[string]interface{}, error) {
	spec := action.Spec
	specMap := map[string]interface{}{}

	if spec.OverridePolicy != "" {
		policy, err := render(string(spec.OverridePolicy))
		if err != nil {
			return nil, fmt.Errorf("failed to render Globalization overridePolicy: %w", err)
		}
		specMap["overridePolicy"] = policy
	}

	if spec.ClusterAffinity != nil {
		clusterAffinity, err := renderLabelSelector(spec.ClusterAffinity, render)
		if err != nil {
			return nil, fmt.Errorf("failed to render Globalization clusterAffinity: %w", err)
		}
		specMap["clusterAffinity"] = clusterAffinity
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

func renderLabelSelector(selector *metav1.LabelSelector, render func(string) (string, error)) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if len(selector.MatchLabels) > 0 {
		labels := make(map[string]interface{}, len(selector.MatchLabels))
		for key, value := range selector.MatchLabels {
			renderedKey, err := render(key)
			if err != nil {
				return nil, err
			}
			renderedValue, err := render(value)
			if err != nil {
				return nil, err
			}
			labels[renderedKey] = renderedValue
		}
		result["matchLabels"] = labels
	}

	if len(selector.MatchExpressions) > 0 {
		expressions := make([]interface{}, 0, len(selector.MatchExpressions))
		for i := range selector.MatchExpressions {
			expr := selector.MatchExpressions[i]
			renderedKey, err := render(expr.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to render matchExpressions[%d].key: %w", i, err)
			}
			entry := map[string]interface{}{
				"key":      renderedKey,
				"operator": string(expr.Operator),
			}
			if len(expr.Values) > 0 {
				values := make([]interface{}, 0, len(expr.Values))
				for j := range expr.Values {
					renderedValue, err := render(expr.Values[j])
					if err != nil {
						return nil, fmt.Errorf("failed to render matchExpressions[%d].values[%d]: %w", i, j, err)
					}
					values = append(values, renderedValue)
				}
				entry["values"] = values
			}
			expressions = append(expressions, entry)
		}
		result["matchExpressions"] = expressions
	}

	return result, nil
}

func renderOverrideConfig(
	override *clusternetapps.OverrideConfig,
	index int,
	render func(string) (string, error),
) (map[string]interface{}, error) {
	entry := map[string]interface{}{}

	if override.Name != "" {
		name, err := render(override.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to render override[%d] name: %w", index, err)
		}
		entry["name"] = name
	}

	overrideType, err := render(string(override.Type))
	if err != nil {
		return nil, fmt.Errorf("failed to render override[%d] type: %w", index, err)
	}
	entry["type"] = overrideType

	value, err := render(override.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to render override[%d] value: %w", index, err)
	}
	entry["value"] = value

	if override.OverrideChart {
		entry["overrideChart"] = true
	}

	return entry, nil
}

func renderFeed(feed clusternetapps.Feed, render func(string) (string, error)) (map[string]interface{}, error) {
	apiVersion, err := render(feed.APIVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to render feed apiVersion: %w", err)
	}
	kind, err := render(feed.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to render feed kind: %w", err)
	}
	name, err := render(feed.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to render feed name: %w", err)
	}

	entry := map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"name":       name,
	}
	if feed.Namespace != "" {
		namespace, err := render(feed.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to render feed namespace: %w", err)
		}
		entry["namespace"] = namespace
	}

	return entry, nil
}

func (e *GlobalizationActionExecutor) applyGlobalization(ctx context.Context, obj *unstructured.Unstructured, operation string) error {
	switch operation {
	case drv1alpha1.OperationDelete:
		return client.IgnoreNotFound(e.client.Delete(ctx, obj))
	case drv1alpha1.OperationApply:
		return e.client.Patch(ctx, obj, client.Apply, client.FieldOwner("drplan-controller"), client.ForceOwnership)
	case drv1alpha1.OperationPatch:
		return e.patchGlobalization(ctx, obj)
	default:
		return e.client.Create(ctx, obj)
	}
}

func (e *GlobalizationActionExecutor) patchGlobalization(ctx context.Context, desired *unstructured.Unstructured) error {
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(globalizationActionGVK)
	if err := e.client.Get(ctx, client.ObjectKey{Name: desired.GetName()}, existing); err != nil {
		return fmt.Errorf("get existing Globalization %s: %w", desired.GetName(), err)
	}

	base := existing.DeepCopy()
	existingSpec, found, err := unstructured.NestedMap(existing.Object, "spec")
	if err != nil {
		return fmt.Errorf("read existing Globalization spec: %w", err)
	}
	if !found {
		existingSpec = map[string]interface{}{}
	}

	desiredSpec, found, err := unstructured.NestedMap(desired.Object, "spec")
	if err != nil {
		return fmt.Errorf("read desired Globalization spec: %w", err)
	}
	if found {
		mergeNestedMaps(existingSpec, desiredSpec)
	}
	existing.Object["spec"] = existingSpec

	if err := e.client.Patch(ctx, existing, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("patch Globalization %s: %w", desired.GetName(), err)
	}
	desired.SetUID(existing.GetUID())
	return nil
}

func mergeNestedMaps(dst, src map[string]interface{}) {
	for key, value := range src {
		srcMap, srcIsMap := value.(map[string]interface{})
		dstMap, dstIsMap := dst[key].(map[string]interface{})
		if srcIsMap && dstIsMap {
			mergeNestedMaps(dstMap, srcMap)
			dst[key] = dstMap
			continue
		}
		dst[key] = value
	}
}

// Rollback rolls back a Globalization action.
func (e *GlobalizationActionExecutor) Rollback(
	ctx context.Context,
	action *drv1alpha1.Action,
	actionStatus *drv1alpha1.ActionStatus,
	params map[string]interface{},
) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back Globalization action: %s", action.Name)

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

	if action.Globalization != nil &&
		effectiveGlobalizationOperation(action.Globalization.Operation) == drv1alpha1.OperationCreate &&
		actionStatus.Outputs != nil &&
		actionStatus.Outputs.GlobalizationRef != nil {
		glob := &unstructured.Unstructured{}
		glob.SetGroupVersionKind(globalizationActionGVK)
		glob.SetName(actionStatus.Outputs.GlobalizationRef.Name)

		if err := e.client.Delete(ctx, glob); client.IgnoreNotFound(err) != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to delete Globalization: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, fmt.Errorf("failed to delete Globalization: %w", err)
		}

		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted Globalization %s", glob.GetName())
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		return rollbackStatus, nil
	}

	rollbackStatus.Phase = drv1alpha1.PhaseSkipped
	rollbackStatus.Message = "No Globalization to rollback"
	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	return rollbackStatus, nil
}

// Type returns the action type.
func (e *GlobalizationActionExecutor) Type() string {
	return drv1alpha1.ActionTypeGlobalization
}

func failGlobalizationStatus(status *drv1alpha1.ActionStatus, message string) *drv1alpha1.ActionStatus {
	status.Phase = drv1alpha1.PhaseFailed
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = message
	return status
}

func effectiveGlobalizationOperation(operation string) string {
	if operation == "" {
		return drv1alpha1.OperationCreate
	}
	return operation
}

func globalizationActionVerb(operation string) string {
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
