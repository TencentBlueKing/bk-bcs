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
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

// KubernetesResourceActionExecutor implements ActionExecutor for KubernetesResource actions
type KubernetesResourceActionExecutor struct {
	client client.Client
}

// NewKubernetesResourceActionExecutor creates a new KubernetesResource action executor
func NewKubernetesResourceActionExecutor(client client.Client) *KubernetesResourceActionExecutor {
	return &KubernetesResourceActionExecutor{client: client}
}

// Execute executes a KubernetesResource action
func (e *KubernetesResourceActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing KubernetesResource action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: startTime},
	}

	if action.Resource == nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = "KubernetesResource configuration is nil"
		return status, fmt.Errorf("KubernetesResource configuration is required")
	}

	// Render manifest with parameters
	templateData := &utils.TemplateData{Params: params}
	manifest, err := utils.RenderTemplate(action.Resource.Manifest, templateData)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to render manifest: %v", err)
		return status, err
	}

	// Parse YAML to unstructured object
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to parse manifest: %v", err)
		return status, err
	}

	operation, err := utils.RenderTemplate(action.Resource.Operation, templateData)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to render operation: %v", err)
		return status, err
	}
	if operation == "" {
		operation = "Create"
	}

	klog.V(4).Infof("Performing %s operation on %s %s/%s", operation, obj.GetKind(), obj.GetNamespace(), obj.GetName())

	// Perform operation
	switch operation {
	case "Create":
		if err := e.client.Create(ctx, obj); err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to create resource: %v", err)
			return status, err
		}
	case "Apply":
		if err := e.client.Patch(ctx, obj, client.Apply, client.ForceOwnership, client.FieldOwner("drplan-controller")); err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to apply resource: %v", err)
			return status, err
		}
	case "Delete":
		if err := e.client.Delete(ctx, obj); err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to delete resource: %v", err)
			return status, err
		}
	default:
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Unsupported operation: %s", operation)
		return status, fmt.Errorf("unsupported operation: %s", operation)
	}

	// Store resource reference
	status.Outputs = &drv1alpha1.ActionOutputs{
		ResourceRef: &corev1.ObjectReference{
			Kind:       obj.GetKind(),
			APIVersion: obj.GetAPIVersion(),
			Namespace:  obj.GetNamespace(),
			Name:       obj.GetName(),
			UID:        obj.GetUID(),
		},
	}

	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("Resource %s %s/%s %s successfully", obj.GetKind(), obj.GetNamespace(), obj.GetName(), operation)

	klog.Infof("KubernetesResource action %s completed", action.Name)
	return status, nil
}

// Rollback rolls back a KubernetesResource action
func (e *KubernetesResourceActionExecutor) Rollback(ctx context.Context, action *drv1alpha1.Action, actionStatus *drv1alpha1.ActionStatus, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back KubernetesResource action: %s", action.Name)

	// Create rollback status object
	rollbackStatus := &drv1alpha1.ActionStatus{
		Name:      actionStatus.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: time.Now()},
	}

	// Execute custom rollback if defined
	if action.Rollback != nil {
		klog.V(4).Infof("Executing custom rollback for KubernetesResource action %s", action.Name)
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

	// Automatic rollback for Create operation: delete the resource
	if action.Resource.Operation == "Create" && actionStatus.Outputs != nil && actionStatus.Outputs.ResourceRef != nil {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(obj.GroupVersionKind())
		obj.SetName(actionStatus.Outputs.ResourceRef.Name)
		obj.SetNamespace(actionStatus.Outputs.ResourceRef.Namespace)

		klog.V(4).Infof("Deleting resource %s %s/%s", obj.GetKind(), obj.GetNamespace(), obj.GetName())
		if err := e.client.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to delete resource: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, fmt.Errorf("failed to delete resource: %w", err)
		}

		klog.Infof("Resource %s %s/%s deleted successfully", obj.GetKind(), obj.GetNamespace(), obj.GetName())
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted resource %s %s/%s",
			obj.GetKind(), obj.GetNamespace(), obj.GetName())
	} else {
		// No automatic rollback for non-Create operations
		rollbackStatus.Phase = drv1alpha1.PhaseSkipped
		rollbackStatus.Message = "No automatic rollback for non-Create operation"
	}

	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	return rollbackStatus, nil
}

// Type returns the action type
func (e *KubernetesResourceActionExecutor) Type() string {
	return "KubernetesResource"
}
