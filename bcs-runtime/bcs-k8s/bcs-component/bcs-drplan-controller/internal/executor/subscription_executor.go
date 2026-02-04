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

// SubscriptionActionExecutor implements ActionExecutor for Subscription actions
type SubscriptionActionExecutor struct {
	client client.Client
}

// NewSubscriptionActionExecutor creates a new Subscription action executor
func NewSubscriptionActionExecutor(client client.Client) *SubscriptionActionExecutor {
	return &SubscriptionActionExecutor{client: client}
}

// Execute executes a Subscription action
func (e *SubscriptionActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing Subscription action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: startTime},
	}

	// Validate configuration
	if err := e.validateSubscriptionConfig(action); err != nil {
		return failSubscriptionStatus(status, err.Error()), err
	}

	// Prepare template renderer
	templateData := &utils.TemplateData{Params: params}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	// Render name and namespace
	subName, subNamespace, err := e.renderSubscriptionNameAndNamespace(action, render)
	if err != nil {
		return failSubscriptionStatus(status, err.Error()), err
	}

	// Build spec map
	specMap, err := e.buildSubscriptionSpecMap(action.Subscription.Spec, render)
	if err != nil {
		return failSubscriptionStatus(status, err.Error()), err
	}

	sub := &unstructured.Unstructured{}
	sub.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	sub.SetName(subName)
	sub.SetNamespace(subNamespace)
	sub.Object["spec"] = specMap

	klog.V(4).Infof("Creating Subscription %s/%s", sub.GetNamespace(), sub.GetName())
	if err := e.client.Create(ctx, sub); err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to create Subscription: %v", err)
		return status, err
	}

	// Store subscription reference
	status.Outputs = &drv1alpha1.ActionOutputs{
		SubscriptionRef: &corev1.ObjectReference{
			Kind:       "Subscription",
			APIVersion: "apps.clusternet.io/v1alpha1",
			Namespace:  sub.GetNamespace(),
			Name:       sub.GetName(),
			UID:        sub.GetUID(),
		},
	}

	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("Subscription %s/%s created successfully", sub.GetNamespace(), sub.GetName())

	klog.Infof("Subscription action %s completed", action.Name)
	return status, nil
}

// Rollback rolls back a Subscription action by deleting the CR
func (e *SubscriptionActionExecutor) Rollback(ctx context.Context, action *drv1alpha1.Action, actionStatus *drv1alpha1.ActionStatus, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back Subscription action: %s", action.Name)

	// Create rollback status object
	rollbackStatus := &drv1alpha1.ActionStatus{
		Name:      actionStatus.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: time.Now()},
	}

	// Execute custom rollback if defined
	if action.Rollback != nil {
		klog.V(4).Infof("Executing custom rollback for Subscription action %s", action.Name)
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

	// Automatic rollback: delete the subscription
	if actionStatus.Outputs != nil && actionStatus.Outputs.SubscriptionRef != nil {
		sub := &unstructured.Unstructured{}
		sub.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "apps.clusternet.io",
			Version: "v1alpha1",
			Kind:    "Subscription",
		})
		sub.SetName(actionStatus.Outputs.SubscriptionRef.Name)
		sub.SetNamespace(actionStatus.Outputs.SubscriptionRef.Namespace)

		klog.V(4).Infof("Deleting Subscription %s/%s", sub.GetNamespace(), sub.GetName())
		if err := e.client.Delete(ctx, sub); client.IgnoreNotFound(err) != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to delete Subscription: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, fmt.Errorf("failed to delete Subscription: %w", err)
		}

		klog.Infof("Subscription %s/%s deleted successfully", sub.GetNamespace(), sub.GetName())
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted Subscription %s/%s",
			sub.GetNamespace(), sub.GetName())
	} else {
		// No subscription to delete
		rollbackStatus.Phase = drv1alpha1.PhaseSkipped
		rollbackStatus.Message = "No Subscription to rollback"
	}

	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	return rollbackStatus, nil
}

// Type returns the action type
func (e *SubscriptionActionExecutor) Type() string {
	return "Subscription"
}

// Helper functions to reduce cyclomatic complexity

// validateSubscriptionConfig validates the subscription configuration
func (e *SubscriptionActionExecutor) validateSubscriptionConfig(action *drv1alpha1.Action) error {
	if action.Subscription == nil {
		return fmt.Errorf("subscription configuration is required")
	}
	if action.Subscription.Spec == nil {
		return fmt.Errorf("Subscription.Spec is required")
	}
	return nil
}

// renderSubscriptionNameAndNamespace renders the subscription name and namespace
func (e *SubscriptionActionExecutor) renderSubscriptionNameAndNamespace(action *drv1alpha1.Action, render func(string) (string, error)) (string, string, error) {
	subName, err := render(action.Subscription.Name)
	if err != nil {
		return "", "", fmt.Errorf("failed to render Subscription name: %w", err)
	}

	subNamespace := "default"
	if action.Subscription.Namespace != "" {
		subNamespace, err = render(action.Subscription.Namespace)
		if err != nil {
			return "", "", fmt.Errorf("failed to render Subscription namespace: %w", err)
		}
	}

	return subName, subNamespace, nil
}

// buildSubscriptionSpecMap builds the subscription spec map
func (e *SubscriptionActionExecutor) buildSubscriptionSpecMap(spec *clusternetapps.SubscriptionSpec, render func(string) (string, error)) (map[string]interface{}, error) {
	specMap := make(map[string]interface{})

	// Set simple fields
	if err := e.setSimpleSpecFields(spec, specMap, render); err != nil {
		return nil, err
	}

	// Render feeds
	if err := e.renderFeeds(spec.Feeds, specMap, render); err != nil {
		return nil, err
	}

	// Set subscribers and tolerations
	if len(spec.Subscribers) > 0 {
		specMap["subscribers"] = spec.Subscribers
	}
	if len(spec.ClusterTolerations) > 0 {
		specMap["clusterTolerations"] = spec.ClusterTolerations
	}

	return specMap, nil
}

// setSimpleSpecFields sets simple spec fields
func (e *SubscriptionActionExecutor) setSimpleSpecFields(spec *clusternetapps.SubscriptionSpec, specMap map[string]interface{}, render func(string) (string, error)) error {
	if spec.SchedulerName != "" {
		schedulerName, err := render(spec.SchedulerName)
		if err != nil {
			return fmt.Errorf("failed to render SchedulerName: %w", err)
		}
		specMap["schedulerName"] = schedulerName
	}
	if spec.SchedulingBySubGroup != nil {
		specMap["schedulingBySubGroup"] = *spec.SchedulingBySubGroup
	}
	if spec.SchedulingStrategy != "" {
		specMap["schedulingStrategy"] = string(spec.SchedulingStrategy)
	}
	if spec.DividingScheduling != nil {
		specMap["dividingScheduling"] = spec.DividingScheduling
	}
	if spec.Priority != nil {
		specMap["priority"] = *spec.Priority
	}
	if spec.PreemptionPolicy != nil {
		specMap["preemptionPolicy"] = string(*spec.PreemptionPolicy)
	}
	return nil
}

// renderFeeds renders the feeds array
func (e *SubscriptionActionExecutor) renderFeeds(feeds []clusternetapps.Feed, specMap map[string]interface{}, render func(string) (string, error)) error {
	if len(feeds) == 0 {
		return nil
	}

	feedsList := make([]interface{}, 0, len(feeds))
	for i := range feeds {
		feedMap, err := e.renderSingleFeed(&feeds[i], i, render)
		if err != nil {
			return err
		}
		feedsList = append(feedsList, feedMap)
	}
	specMap["feeds"] = feedsList
	return nil
}

// renderSingleFeed renders a single feed
func (e *SubscriptionActionExecutor) renderSingleFeed(f *clusternetapps.Feed, index int, render func(string) (string, error)) (map[string]interface{}, error) {
	apiVer, err := render(f.APIVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to render feed[%d] apiVersion: %w", index, err)
	}

	kind, err := render(f.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to render feed[%d] kind: %w", index, err)
	}

	name, err := render(f.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to render feed[%d] name: %w", index, err)
	}

	ns, _ := render(f.Namespace)

	return map[string]interface{}{
		"apiVersion": apiVer,
		"kind":       kind,
		"name":       name,
		"namespace":  ns,
	}, nil
}

// failSubscriptionStatus sets failure status
func failSubscriptionStatus(status *drv1alpha1.ActionStatus, message string) *drv1alpha1.ActionStatus {
	status.Phase = drv1alpha1.PhaseFailed
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = message
	return status
}
