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

// LocalizationActionExecutor implements ActionExecutor for Localization actions
type LocalizationActionExecutor struct {
	client client.Client
}

// NewLocalizationActionExecutor creates a new Localization action executor
func NewLocalizationActionExecutor(client client.Client) *LocalizationActionExecutor {
	return &LocalizationActionExecutor{client: client}
}

// Execute executes a Localization action
func (e *LocalizationActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing Localization action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: startTime},
	}

	if action.Localization == nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = "Localization configuration is nil"
		return status, fmt.Errorf("localization configuration is required")
	}
	if action.Localization.Spec == nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = "Localization.Spec is required"
		return status, fmt.Errorf("Localization.Spec is required")
	}

	templateData := &utils.TemplateData{Params: params}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	locName, err := render(action.Localization.Name)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to render Localization name: %v", err)
		return status, err
	}
	locNamespace, err := render(action.Localization.Namespace)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to render Localization namespace: %v", err)
		return status, err
	}

	spec := action.Localization.Spec
	specMap := map[string]interface{}{
		"priority": spec.Priority,
	}
	if spec.OverridePolicy != "" {
		specMap["overridePolicy"] = string(spec.OverridePolicy)
	}
	// Feed is embedded in clusternet LocalizationSpec
	if spec.APIVersion != "" || spec.Kind != "" || spec.Name != "" {
		feedAPI, err := render(spec.APIVersion)
		if err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to render Feed apiVersion: %v", err)
			return status, err
		}
		feedKind, err := render(spec.Kind)
		if err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to render Feed kind: %v", err)
			return status, err
		}
		feedName, err := render(spec.Name)
		if err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to render Feed name: %v", err)
			return status, err
		}
		feedNS, err := render(spec.Namespace)
		if err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to render Feed namespace: %v", err)
			return status, err
		}
		specMap["feed"] = map[string]interface{}{
			"apiVersion": feedAPI,
			"kind":       feedKind,
			"name":       feedName,
			"namespace":  feedNS,
		}
	}
	if len(spec.Overrides) > 0 {
		overrideList := make([]interface{}, 0, len(spec.Overrides))
		for i := range spec.Overrides {
			o := &spec.Overrides[i]
			on, err := render(o.Name)
			if err != nil {
				status.Phase = drv1alpha1.PhaseFailed
				status.CompletionTime = &metav1.Time{Time: time.Now()}
				status.Message = fmt.Sprintf("Failed to render override[%d] name: %v", i, err)
				return status, err
			}
			ot, err := render(string(o.Type))
			if err != nil {
				status.Phase = drv1alpha1.PhaseFailed
				status.CompletionTime = &metav1.Time{Time: time.Now()}
				status.Message = fmt.Sprintf("Failed to render override[%d] type: %v", i, err)
				return status, err
			}
			ov, err := render(o.Value)
			if err != nil {
				status.Phase = drv1alpha1.PhaseFailed
				status.CompletionTime = &metav1.Time{Time: time.Now()}
				status.Message = fmt.Sprintf("Failed to render override[%d] value: %v", i, err)
				return status, err
			}
			entry := map[string]interface{}{"name": on, "type": ot, "value": ov}
			if o.OverrideChart {
				entry["overrideChart"] = true
			}
			overrideList = append(overrideList, entry)
		}
		specMap["overrides"] = overrideList
	}

	loc := &unstructured.Unstructured{}
	loc.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Localization",
	})
	loc.SetName(locName)
	loc.SetNamespace(locNamespace)
	loc.Object["spec"] = specMap

	klog.V(4).Infof("Creating Localization %s/%s", loc.GetNamespace(), loc.GetName())
	if err := e.client.Create(ctx, loc); err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to create Localization: %v", err)
		return status, err
	}

	// Store localization reference
	status.Outputs = &drv1alpha1.ActionOutputs{
		LocalizationRef: &corev1.ObjectReference{
			Kind:       "Localization",
			APIVersion: "apps.clusternet.io/v1alpha1",
			Namespace:  loc.GetNamespace(),
			Name:       loc.GetName(),
			UID:        loc.GetUID(),
		},
	}

	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("Localization %s/%s created successfully", loc.GetNamespace(), loc.GetName())

	klog.Infof("Localization action %s completed", action.Name)
	return status, nil
}

// Rollback rolls back a Localization action by deleting the CR
func (e *LocalizationActionExecutor) Rollback(ctx context.Context, action *drv1alpha1.Action, actionStatus *drv1alpha1.ActionStatus, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back Localization action: %s", action.Name)

	// Create rollback status object
	rollbackStatus := &drv1alpha1.ActionStatus{
		Name:      actionStatus.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: time.Now()},
	}

	// Execute custom rollback if defined
	if action.Rollback != nil {
		klog.V(4).Infof("Executing custom rollback for Localization action %s", action.Name)
		customStatus, err := e.Execute(ctx, action.Rollback, params)
		if err != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Custom rollback failed: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, err
		}
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = "Rolled back: executed custom rollback action"
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		rollbackStatus.Outputs = customStatus.Outputs
		return rollbackStatus, nil
	}

	// Automatic rollback: delete the localization
	if actionStatus.Outputs != nil && actionStatus.Outputs.LocalizationRef != nil {
		loc := &unstructured.Unstructured{}
		loc.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "apps.clusternet.io",
			Version: "v1alpha1",
			Kind:    "Localization",
		})
		loc.SetName(actionStatus.Outputs.LocalizationRef.Name)
		loc.SetNamespace(actionStatus.Outputs.LocalizationRef.Namespace)

		klog.V(4).Infof("Deleting Localization %s/%s", loc.GetNamespace(), loc.GetName())
		if err := e.client.Delete(ctx, loc); client.IgnoreNotFound(err) != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to delete Localization: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, fmt.Errorf("failed to delete Localization: %w", err)
		}

		klog.Infof("Localization %s/%s deleted successfully", loc.GetNamespace(), loc.GetName())
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted Localization %s/%s",
			loc.GetNamespace(), loc.GetName())
	} else {
		// No localization to delete (e.g., action was Skipped)
		rollbackStatus.Phase = drv1alpha1.PhaseSkipped
		rollbackStatus.Message = "No Localization to rollback"
	}

	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	return rollbackStatus, nil
}

// Type returns the action type
func (e *LocalizationActionExecutor) Type() string {
	return "Localization"
}
