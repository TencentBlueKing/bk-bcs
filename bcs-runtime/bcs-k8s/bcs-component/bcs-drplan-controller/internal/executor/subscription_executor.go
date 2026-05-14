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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

// SubscriptionActionExecutor implements ActionExecutor for Subscription actions
type SubscriptionActionExecutor struct {
	client             client.Client
	childClientFactory ChildClusterClientFactory
}

// NewSubscriptionActionExecutor creates a new Subscription action executor.
// parentConfig is used to build child-cluster clients via SocketProxy.
// NOCC:tosa/comment_ratio(设计如此)
func NewSubscriptionActionExecutor(
	c client.Client,
	parentConfig *rest.Config,
) *SubscriptionActionExecutor {
	return &SubscriptionActionExecutor{
		client:             c,
		childClientFactory: NewSocketProxyChildClusterClientFactory(c, parentConfig),
	}
}

// Execute executes a Subscription action
//
//nolint:funlen // orchestrates render/apply/wait/cleanup
func (e *SubscriptionActionExecutor) Execute(
	ctx context.Context,
	action *drv1alpha1.Action,
	params map[string]interface{},
) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing Subscription action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     drv1alpha1.PhaseRunning,
		StartTime: &metav1.Time{Time: startTime},
	}

	// Validate before rendering templates so user-facing errors point at the
	// missing Subscription fields instead of later unstructured object handling.
	if err := e.validateSubscriptionConfig(action); err != nil {
		return failSubscriptionStatus(status, err.Error()), err
	}

	// Subscription actions support DRPlan parameter placeholders in name,
	// namespace, and feeds. Rendering once here keeps create/apply and waitReady
	// checks using exactly the same resolved values.
	templateData := &utils.TemplateData{Params: params}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	subName, subNamespace, err := e.renderSubscriptionNameAndNamespace(action, render)
	if err != nil {
		return failSubscriptionStatus(status, err.Error()), err
	}

	// Delete mode deliberately skips spec rendering. Generated delete actions
	// and inferred delete-mode rollbacks only need a namespaced Subscription key.
	if action.Subscription.Operation == drv1alpha1.OperationDelete {
		sub := &unstructured.Unstructured{}
		sub.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "apps.clusternet.io",
			Version: "v1alpha1",
			Kind:    "Subscription",
		})
		sub.SetName(subName)
		sub.SetNamespace(subNamespace)

		if applyErr := e.applySubscription(ctx, sub, action.Subscription.Operation); applyErr != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to delete Subscription: %v", applyErr)
			return status, applyErr
		}

		status.Outputs = &drv1alpha1.ActionOutputs{
			SubscriptionRef: &corev1.ObjectReference{
				Kind:       "Subscription",
				APIVersion: "apps.clusternet.io/v1alpha1",
				Namespace:  sub.GetNamespace(),
				Name:       sub.GetName(),
			},
		}
		status.Phase = drv1alpha1.PhaseSucceeded
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Subscription %s/%s deleted successfully", sub.GetNamespace(), sub.GetName())
		return status, nil
	}

	if cleanupErr := e.applyHookPreCleanup(ctx, action, subNamespace, subName); cleanupErr != nil {
		return failSubscriptionStatus(status, fmt.Sprintf("hook pre-cleanup failed: %v", cleanupErr)), cleanupErr
	}

	// Build rendered subscription payload used both for create and waitReady
	// checks. This is important for feed readiness because child-cluster lookups
	// must use the same rendered namespace/name that Clusternet receives.
	renderedSub, err := e.renderSubscriptionAction(action.Subscription.Spec, render)
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
	sub.Object["spec"] = renderedSub.specMap

	if err := e.applySubscription(ctx, sub, action.Subscription.Operation); err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to create/apply Subscription: %v", err)
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

	if isPerClusterMode(action) {
		// PerCluster execution is coordinated by NativeWorkflowExecutor. This
		// executor only creates the parent Subscription and returns a Running
		// status so the workflow layer can aggregate per-cluster results.
		status.Phase = drv1alpha1.PhaseRunning
		status.Message = fmt.Sprintf("PerCluster mode: parent Subscription %s/%s created, per-cluster execution deferred to workflow executor", sub.GetNamespace(), sub.GetName())
		klog.Infof("PerCluster action %s: parent Subscription created, deferring per-cluster execution", action.Name)
		return status, nil
	}

	if action.WaitReady {
		// waitReady is synchronous only for regular Subscription actions. Hook
		// PerCluster actions wait in the workflow layer after each cluster has its
		// own subscription and child-cluster client context.
		waitDur, parseErr := parseActionTimeout(action.Timeout)
		if parseErr != nil {
			return failSubscriptionStatus(status, fmt.Sprintf("invalid timeout: %v", parseErr)), parseErr
		}
		if waitErr := e.waitForSubscriptionReady(ctx, subNamespace, subName, renderedSub.feeds, waitDur); waitErr != nil {
			if cleanupErr := e.applyHookPostCleanup(ctx, action, subNamespace, subName, drv1alpha1.PhaseFailed); cleanupErr != nil {
				return failSubscriptionStatus(
						status,
						fmt.Sprintf("waitReady failed: %v; hook cleanup failed: %v", waitErr, cleanupErr),
					), fmt.Errorf(
						"waitReady failed: %w; hook cleanup failed: %v",
						waitErr, cleanupErr,
					)
			}
			return failSubscriptionStatus(status, fmt.Sprintf("waitReady failed: %v", waitErr)), waitErr
		}
	}

	if cleanupErr := e.applyHookPostCleanup(ctx, action, subNamespace, subName, drv1alpha1.PhaseSucceeded); cleanupErr != nil {
		return failSubscriptionStatus(status, fmt.Sprintf("hook cleanup failed: %v", cleanupErr)), cleanupErr
	}

	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	verb := actionVerbCreated
	if action.Subscription.Operation == "Apply" {
		verb = actionVerbApplied
	}
	status.Message = fmt.Sprintf("Subscription %s/%s %s successfully", sub.GetNamespace(), sub.GetName(), verb)

	klog.Infof("Subscription action %s completed", action.Name)
	return status, nil
}

func (e *SubscriptionActionExecutor) applyHookPreCleanup(
	ctx context.Context,
	action *drv1alpha1.Action,
	subNamespace, subName string,
) error {
	// Helm hooks are commonly re-run with stable names. BeforeCreate cleanup
	// removes stale subscriptions and their distributed objects before the next
	// hook run creates a fresh subscription with the same name.
	if action == nil || action.HookCleanup == nil || !action.HookCleanup.BeforeCreate {
		return nil
	}
	return e.deleteSubscriptionIfExists(ctx, subNamespace, subName)
}

func (e *SubscriptionActionExecutor) applyHookPostCleanup(
	ctx context.Context,
	action *drv1alpha1.Action,
	subNamespace, subName, phase string,
) error {
	if action == nil || action.HookCleanup == nil {
		return nil
	}

	shouldDelete := false
	switch phase {
	case drv1alpha1.PhaseSucceeded:
		// Keep success cleanup opt-in. Operators often need the generated
		// Description and child-cluster Job for auditing hook execution.
		shouldDelete = action.HookCleanup.OnSuccess
	case drv1alpha1.PhaseFailed:
		// Failure cleanup is also opt-in so failed hook artifacts remain
		// available for troubleshooting unless explicitly configured otherwise.
		shouldDelete = action.HookCleanup.OnFailure
	}
	if !shouldDelete {
		return nil
	}

	return e.deleteSubscriptionIfExists(ctx, subNamespace, subName)
}

// applySubscription creates, applies, or replaces a Subscription CR based on the operation type.
// "Apply" uses Server-Side Apply for idempotent create-or-update.
// "Replace" deletes the existing resource (waiting for complete removal) then creates a new one.
// All other values (including "") fall back to Create.
func (e *SubscriptionActionExecutor) applySubscription(
	ctx context.Context,
	sub *unstructured.Unstructured,
	operation string,
) error {
	switch operation {
	case drv1alpha1.OperationDelete:
		return client.IgnoreNotFound(e.client.Delete(ctx, sub))
	case drv1alpha1.OperationApply:
		klog.V(4).Infof("Applying (SSA) Subscription %s/%s", sub.GetNamespace(), sub.GetName())
		return e.client.Patch(ctx, sub, client.Apply,
			client.FieldOwner("drplan-controller"),
			client.ForceOwnership)
	case drv1alpha1.OperationReplace:
		return e.replaceSubscription(ctx, sub)
	default:
		klog.V(4).Infof("Creating Subscription %s/%s", sub.GetNamespace(), sub.GetName())
		return e.client.Create(ctx, sub)
	}
}

// replaceSubscription deletes an existing Subscription (if any), waits for it to be fully
// removed (including Clusternet GC of distributed feeds), then creates a fresh one.
func (e *SubscriptionActionExecutor) replaceSubscription(ctx context.Context, sub *unstructured.Unstructured) error {
	key := client.ObjectKeyFromObject(sub)
	klog.Infof("Replace: deleting existing Subscription %s (if any)", key)

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(sub.GroupVersionKind())
	if err := e.client.Get(ctx, key, existing); err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Infof("Replace: no existing Subscription %s, proceeding to create", key)
			return e.client.Create(ctx, sub)
		}
		return fmt.Errorf("replace: get existing Subscription %s: %w", key, err)
	}

	if err := e.client.Delete(ctx, existing); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("replace: delete Subscription %s: %w", key, err)
	}

	if err := e.waitForDeletion(ctx, key, sub.GroupVersionKind()); err != nil {
		return fmt.Errorf("replace: waiting for Subscription %s deletion: %w", key, err)
	}

	klog.V(4).Infof("Replace: creating fresh Subscription %s", key)
	return e.client.Create(ctx, sub)
}

func (e *SubscriptionActionExecutor) deleteSubscriptionIfExists(ctx context.Context, namespace, name string) error {
	key := client.ObjectKey{Namespace: namespace, Name: name}
	klog.V(4).Infof("Deleting existing Subscription %s/%s if present", namespace, name)

	// Clusternet cleanup is asynchronous. Deleting the hub Subscription is not
	// enough for immediate re-create flows, so callers wait for the object to be
	// fully gone before applying the replacement.
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	if err := e.client.Get(ctx, key, existing); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get Subscription %s: %w", key, err)
	}

	if err := e.client.Delete(ctx, existing); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("delete Subscription %s: %w", key, err)
	}

	return e.waitForDeletion(ctx, key, existing.GroupVersionKind())
}

// waitForDeletion polls until the resource is fully removed or the context expires.
func (e *SubscriptionActionExecutor) waitForDeletion(ctx context.Context, key client.ObjectKey, gvk schema.GroupVersionKind) error {
	const pollInterval = 2 * time.Second
	probe := &unstructured.Unstructured{}
	probe.SetGroupVersionKind(gvk)

	for {
		if err := e.client.Get(ctx, key, probe); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		klog.V(4).Infof("Waiting for %s %s to be deleted...", gvk.Kind, key)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// Rollback rolls back a Subscription action by deleting the CR
func (e *SubscriptionActionExecutor) Rollback(
	ctx context.Context,
	action *drv1alpha1.Action,
	actionStatus *drv1alpha1.ActionStatus,
	params map[string]interface{},
) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back Subscription action: %s", action.Name)

	// Create rollback status object
	rollbackStatus := &drv1alpha1.ActionStatus{
		Name:      actionStatus.Name,
		Phase:     drv1alpha1.PhaseRunning,
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
		rollbackStatus.Message = drv1alpha1.MessageRollbackSuccess
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		rollbackStatus.Outputs = customStatus.Outputs
		return rollbackStatus, nil
	}

	// Automatic rollback deletes all recorded subscriptions. PerCluster actions
	// can produce multiple child-specific SubscriptionRefs, while older paths
	// still populate the singular SubscriptionRef field.
	if actionStatus.Outputs != nil {
		refs := make([]corev1.ObjectReference, 0, len(actionStatus.Outputs.SubscriptionRefs)+1)
		refs = append(refs, actionStatus.Outputs.SubscriptionRefs...)
		if len(refs) == 0 && actionStatus.Outputs.SubscriptionRef != nil {
			refs = append(refs, *actionStatus.Outputs.SubscriptionRef)
		}
		if len(refs) > 0 {
			deleted := make([]string, 0, len(refs))
			for _, ref := range refs {
				sub := &unstructured.Unstructured{}
				sub.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   "apps.clusternet.io",
					Version: "v1alpha1",
					Kind:    "Subscription",
				})
				sub.SetName(ref.Name)
				sub.SetNamespace(ref.Namespace)

				klog.V(4).Infof("Deleting Subscription %s/%s", sub.GetNamespace(), sub.GetName())
				if err := e.client.Delete(ctx, sub); client.IgnoreNotFound(err) != nil {
					rollbackStatus.Phase = drv1alpha1.PhaseFailed
					rollbackStatus.Message = fmt.Sprintf("Failed to delete Subscription: %v", err)
					rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
					return rollbackStatus, fmt.Errorf("failed to delete Subscription: %w", err)
				}
				deleted = append(deleted, fmt.Sprintf("%s/%s", sub.GetNamespace(), sub.GetName()))
			}

			rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
			if len(deleted) == 1 {
				rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted Subscription %s", deleted[0])
			} else {
				rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted %d Subscriptions", len(deleted))
			}
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, nil
		}
	}

	{
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
	// Delete needs only identity. Requiring Spec here would make inferred delete
	// mode verbose and would diverge from Kubernetes delete semantics.
	if action.Subscription.Operation == drv1alpha1.OperationDelete {
		if action.Subscription.Name == "" {
			return fmt.Errorf("Subscription.Name is required")
		}
		return nil
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

	subNamespace := drv1alpha1.DefaultNamespace
	if action.Subscription.Namespace != "" {
		// Namespace defaults after rendering name so existing manifests that omit
		// namespace keep the historical "default" behavior.
		subNamespace, err = render(action.Subscription.Namespace)
		if err != nil {
			return "", "", fmt.Errorf("failed to render Subscription namespace: %w", err)
		}
	}

	return subName, subNamespace, nil
}

type renderedSubscriptionAction struct {
	specMap map[string]interface{}
	feeds   []clusternetapps.Feed
}

// renderSubscriptionAction builds the rendered subscription payload used for create and waitReady checks.
func (e *SubscriptionActionExecutor) renderSubscriptionAction(
	spec *clusternetapps.SubscriptionSpec,
	render func(string) (string, error),
) (*renderedSubscriptionAction, error) {
	// renderedSubscriptionAction keeps both the unstructured spec map and typed
	// feeds. The map is sent to Kubernetes; the typed feeds drive waitReady.
	specMap := make(map[string]interface{})
	rendered := &renderedSubscriptionAction{
		specMap: specMap,
		feeds:   make([]clusternetapps.Feed, 0, len(spec.Feeds)),
	}

	// Preserve Clusternet fields instead of reconstructing only the fields used
	// by current generator output. This keeps the action executor useful for
	// hand-written workflows that use advanced scheduling options.
	if err := e.setSimpleSpecFields(spec, specMap, render); err != nil {
		return nil, err
	}

	// Feeds are the only nested fields that need template rendering today.
	// Subscribers are copied as-is because their affinity structures are not
	// string-templated by the public API.
	if err := e.renderFeeds(spec.Feeds, specMap, &rendered.feeds, render); err != nil {
		return nil, err
	}

	// Set subscribers and tolerations
	if len(spec.Subscribers) > 0 {
		specMap["subscribers"] = spec.Subscribers
	}
	if len(spec.ClusterTolerations) > 0 {
		specMap["clusterTolerations"] = spec.ClusterTolerations
	}

	return rendered, nil
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
func (e *SubscriptionActionExecutor) renderFeeds(
	feeds []clusternetapps.Feed,
	specMap map[string]interface{},
	renderedFeeds *[]clusternetapps.Feed,
	render func(string) (string, error),
) error {
	if len(feeds) == 0 {
		return nil
	}

	feedsList := make([]interface{}, 0, len(feeds))
	for i := range feeds {
		renderedFeed, feedMap, err := e.renderSingleFeed(&feeds[i], i, render)
		if err != nil {
			return err
		}
		*renderedFeeds = append(*renderedFeeds, renderedFeed)
		feedsList = append(feedsList, feedMap)
	}
	specMap["feeds"] = feedsList
	return nil
}

// renderSingleFeed renders a single feed
func (e *SubscriptionActionExecutor) renderSingleFeed(
	f *clusternetapps.Feed,
	index int,
	render func(string) (string, error),
) (clusternetapps.Feed, map[string]interface{}, error) {
	apiVer, err := render(f.APIVersion)
	if err != nil {
		return clusternetapps.Feed{}, nil, fmt.Errorf("failed to render feed[%d] apiVersion: %w", index, err)
	}

	kind, err := render(f.Kind)
	if err != nil {
		return clusternetapps.Feed{}, nil, fmt.Errorf("failed to render feed[%d] kind: %w", index, err)
	}

	name, err := render(f.Name)
	if err != nil {
		return clusternetapps.Feed{}, nil, fmt.Errorf("failed to render feed[%d] name: %w", index, err)
	}

	ns := ""
	if f.Namespace != "" {
		ns, err = render(f.Namespace)
		if err != nil {
			return clusternetapps.Feed{}, nil, fmt.Errorf("failed to render feed[%d] namespace: %w", index, err)
		}
	}

	renderedFeed := clusternetapps.Feed{
		APIVersion: apiVer,
		Kind:       kind,
		Name:       name,
		Namespace:  ns,
	}

	return renderedFeed, map[string]interface{}{
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

// waitForSubscriptionReady polls until all feeds are ready on all binding clusters.
func (e *SubscriptionActionExecutor) waitForSubscriptionReady(
	ctx context.Context,
	namespace, name string,
	feeds []clusternetapps.Feed,
	timeout time.Duration,
) error {
	// Polling starts from the hub Subscription because Clusternet first records
	// binding clusters there, then creates Description objects and child-cluster
	// resources asynchronously.
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(waitPollInterval)
	defer ticker.Stop()

	for {
		ready, reason, err := e.checkSubscriptionReadyOnce(waitCtx, namespace, name, feeds)
		if err != nil {
			return err
		}
		if ready {
			klog.Infof("Subscription %s/%s waitReady succeeded: %s", namespace, name, reason)
			return nil
		}
		klog.V(3).Infof("Subscription %s/%s waitReady pending: %s", namespace, name, reason)

		select {
		case <-waitCtx.Done():
			return fmt.Errorf("timeout waiting for Subscription %s/%s ready: %w", namespace, name, waitCtx.Err())
		case <-ticker.C:
		}
	}
}

func (e *SubscriptionActionExecutor) checkSubscriptionReadyOnce(
	ctx context.Context,
	namespace, name string,
	feeds []clusternetapps.Feed,
) (bool, string, error) {
	// NotFound is a pending state rather than an error because the caller may
	// begin waiting immediately after apply while the apiserver cache catches up.
	sub := &unstructured.Unstructured{}
	sub.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "apps.clusternet.io", Version: "v1alpha1", Kind: "Subscription",
	})
	if err := e.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, sub); err != nil {
		if errors.IsNotFound(err) {
			return false, "subscription not found yet", nil
		}
		return false, "", fmt.Errorf("get Subscription %s/%s: %w", namespace, name, err)
	}

	bindings, found, err := unstructured.NestedStringSlice(sub.Object, "status", "bindingClusters")
	if err != nil {
		return false, "", fmt.Errorf("parse status.bindingClusters: %w", err)
	}
	if !found || len(bindings) == 0 {
		return false, "status.bindingClusters is empty", nil
	}

	return e.evaluateSubscriptionReadiness(ctx, sub, feeds, bindings)
}

func (e *SubscriptionActionExecutor) checkDescriptionFailures(
	ctx context.Context,
	sub *unstructured.Unstructured,
	bindings []string,
) error {
	// Description failures surface scheduling or distribution errors that may
	// occur before the child resource exists. Checking them first gives a clear
	// failure reason instead of waiting until the overall timeout expires.
	subUID := string(sub.GetUID())
	if subUID == "" {
		return nil
	}

	for _, binding := range bindings {
		clusterNS, _, err := parseBindingCluster(binding)
		if err != nil {
			return err
		}

		descriptionList := &unstructured.UnstructuredList{}
		descriptionList.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "apps.clusternet.io",
			Version: "v1alpha1",
			Kind:    "DescriptionList",
		})
		if err := e.client.List(ctx, descriptionList,
			client.InNamespace(clusterNS),
			client.MatchingLabels{"apps.clusternet.io/subs.uid": subUID},
		); err != nil {
			return fmt.Errorf("list Descriptions for Subscription %s/%s in %s: %w",
				sub.GetNamespace(), sub.GetName(), clusterNS, err)
		}

		for i := range descriptionList.Items {
			phase, _, err := unstructured.NestedString(descriptionList.Items[i].Object, "status", "phase")
			if err != nil {
				return fmt.Errorf("parse Description %s/%s status.phase: %w",
					descriptionList.Items[i].GetNamespace(), descriptionList.Items[i].GetName(), err)
			}
			if phase != "Failure" {
				continue
			}
			reason, _, err := unstructured.NestedString(descriptionList.Items[i].Object, "status", "reason")
			if err != nil {
				return fmt.Errorf("parse Description %s/%s status.reason: %w",
					descriptionList.Items[i].GetNamespace(), descriptionList.Items[i].GetName(), err)
			}
			if reason == "" {
				reason = "description deployment failed"
			}
			return fmt.Errorf("description %s/%s failed: %s",
				descriptionList.Items[i].GetNamespace(), descriptionList.Items[i].GetName(), reason)
		}
	}

	return nil
}

func (e *SubscriptionActionExecutor) evaluateSubscriptionReadiness(
	ctx context.Context,
	sub *unstructured.Unstructured,
	feeds []clusternetapps.Feed,
	bindings []string,
) (bool, string, error) {
	if err := e.checkDescriptionFailures(ctx, sub, bindings); err != nil {
		return false, "", err
	}

	if len(feeds) == 0 {
		// A Subscription with no feeds is still useful for scheduling checks in
		// tests and hand-written workflows. Once bindings exist and there are no
		// Description failures, there is no child resource left to wait for.
		return true, "no feeds configured, scheduling confirmed", nil
	}

	return e.checkAllFeedsReady(ctx, feeds, bindings)
}

func (e *SubscriptionActionExecutor) checkAllFeedsReady(
	ctx context.Context,
	feeds []clusternetapps.Feed,
	bindings []string,
) (bool, string, error) {
	// Readiness is all-or-nothing across clusters and feeds. The first pending
	// feed returns a reason so wait logs show which cluster/resource is blocking.
	for _, binding := range bindings {
		clusterNS, clusterName, err := parseBindingCluster(binding)
		if err != nil {
			return false, "", err
		}
		clusterID, err := e.getManagedClusterID(ctx, clusterNS, clusterName)
		if err != nil {
			return false, "", err
		}
		childClient, err := e.childClientFactory.GetChildClient(ctx, clusterID, clusterNS)
		if err != nil {
			return false, "", err
		}
		for i := range feeds {
			feed := feeds[i]
			ready, reason, feedErr := e.isFeedReadyInChildCluster(ctx, childClient, feed)
			if feedErr != nil {
				return false, "", fmt.Errorf("cluster %s feed %s/%s %s: %w",
					clusterName, feed.Namespace, feed.Name, feed.Kind, feedErr)
			}
			if !ready {
				return false, fmt.Sprintf("cluster %s feed %s/%s %s not ready: %s",
					clusterName, feed.Namespace, feed.Name, feed.Kind, reason), nil
			}
		}
	}
	return true, fmt.Sprintf("all feeds ready in %d binding clusters", len(bindings)), nil
}

func (e *SubscriptionActionExecutor) getManagedClusterID(ctx context.Context, namespace, clusterName string) (string, error) {
	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "clusters.clusternet.io", Version: "v1beta1", Kind: "ManagedCluster",
	})
	if err := e.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: clusterName}, cluster); err != nil {
		return "", fmt.Errorf("get ManagedCluster %s/%s: %w", namespace, clusterName, err)
	}
	clusterID, found, err := unstructured.NestedString(cluster.Object, "spec", "clusterId")
	if err != nil {
		return "", fmt.Errorf("parse ManagedCluster %s spec.clusterId: %w", clusterName, err)
	}
	if !found || clusterID == "" {
		return "", fmt.Errorf("ManagedCluster %s has empty spec.clusterId", clusterName)
	}
	return clusterID, nil
}

func (e *SubscriptionActionExecutor) isFeedReadyInChildCluster(
	ctx context.Context,
	childClient client.Client,
	feed clusternetapps.Feed,
) (bool, string, error) {
	// Manifest feeds are hub-side wrappers. Resolve them before child-cluster
	// lookup so waitReady checks the real object distributed by Clusternet.
	resolvedFeed, found, reason, err := e.resolveFeedTarget(ctx, feed)
	if err != nil {
		return false, "", err
	}
	if !found {
		return false, reason, nil
	}

	gv, err := schema.ParseGroupVersion(resolvedFeed.APIVersion)
	if err != nil {
		return false, "", fmt.Errorf("parse feed apiVersion %q: %w", resolvedFeed.APIVersion, err)
	}
	target := &unstructured.Unstructured{}
	target.SetGroupVersionKind(schema.GroupVersionKind{
		Group: gv.Group, Version: gv.Version, Kind: resolvedFeed.Kind,
	})
	if err := childClient.Get(ctx, client.ObjectKey{
		Namespace: resolvedFeed.Namespace, Name: resolvedFeed.Name,
	}, target); err != nil {
		if errors.IsNotFound(err) {
			return false, "resource not found in child cluster", nil
		}
		return false, "", err
	}
	return evaluateResourceReadiness(target)
}

func (e *SubscriptionActionExecutor) resolveFeedTarget(
	ctx context.Context,
	feed clusternetapps.Feed,
) (clusternetapps.Feed, bool, string, error) {
	// Non-Manifest feeds can be checked directly in the child cluster. Manifest
	// feeds need hub lookup because the child cluster receives only the embedded
	// template object, not the Manifest wrapper.
	if !isClusternetManifestFeed(feed) {
		return feed, true, "", nil
	}
	return e.resolveFeedFromManifest(ctx, feed)
}

func isClusternetManifestFeed(feed clusternetapps.Feed) bool {
	// Match the group rather than a single version so future v1beta1 Manifest
	// feeds keep the same readiness path if the template shape remains stable.
	if !strings.EqualFold(feed.Kind, "Manifest") {
		return false
	}
	gv, err := schema.ParseGroupVersion(feed.APIVersion)
	if err != nil {
		return false
	}
	return gv.Group == "apps.clusternet.io"
}

func (e *SubscriptionActionExecutor) resolveFeedFromManifest(
	ctx context.Context,
	feed clusternetapps.Feed,
) (clusternetapps.Feed, bool, string, error) {
	gv, err := schema.ParseGroupVersion(feed.APIVersion)
	if err != nil {
		return clusternetapps.Feed{}, false, "", fmt.Errorf("parse manifest apiVersion %q: %w", feed.APIVersion, err)
	}
	manifest := &unstructured.Unstructured{}
	manifest.SetGroupVersionKind(schema.GroupVersionKind{
		Group: gv.Group, Version: gv.Version, Kind: feed.Kind,
	})
	getErr := e.client.Get(ctx, client.ObjectKey{Namespace: feed.Namespace, Name: feed.Name}, manifest)
	if getErr != nil {
		if errors.IsNotFound(getErr) {
			// Missing hub Manifest is pending because the Subscription may have
			// been observed before its feed action became visible.
			return clusternetapps.Feed{}, false, "resource not found in hub cluster", nil
		}
		return clusternetapps.Feed{}, false, "", fmt.Errorf("get Manifest %s/%s: %w", feed.Namespace, feed.Name, getErr)
	}

	// Clusternet Manifest keeps the raw template at the top level. This is not
	// Kubernetes-style spec.template; using spec.template breaks v0.18 CRDs.
	template, found, err := unstructured.NestedMap(manifest.Object, "template")
	if err != nil {
		return clusternetapps.Feed{}, false, "", fmt.Errorf("parse Manifest %s/%s template: %w", feed.Namespace, feed.Name, err)
	}
	if !found {
		return clusternetapps.Feed{}, false, "", fmt.Errorf("manifest %s/%s template is required", feed.Namespace, feed.Name)
	}

	target, err := feedFromManifestTemplate(template)
	if err != nil {
		return clusternetapps.Feed{}, false, "", fmt.Errorf("resolve Manifest %s/%s template: %w", feed.Namespace, feed.Name, err)
	}
	return target, true, "", nil
}

func feedFromManifestTemplate(template map[string]interface{}) (clusternetapps.Feed, error) {
	// The embedded object must be resolvable to a namespaced feed. The generated
	// hook Job always includes namespace, and requiring it here prevents waiting
	// on an ambiguous child-cluster object.
	apiVersion, found, err := unstructured.NestedString(template, "apiVersion")
	if err != nil {
		return clusternetapps.Feed{}, fmt.Errorf("parse template apiVersion: %w", err)
	}
	if !found || strings.TrimSpace(apiVersion) == "" {
		return clusternetapps.Feed{}, fmt.Errorf("template apiVersion is required")
	}

	kind, found, err := unstructured.NestedString(template, "kind")
	if err != nil {
		return clusternetapps.Feed{}, fmt.Errorf("parse template kind: %w", err)
	}
	if !found || strings.TrimSpace(kind) == "" {
		return clusternetapps.Feed{}, fmt.Errorf("template kind is required")
	}

	name, found, err := unstructured.NestedString(template, "metadata", "name")
	if err != nil {
		return clusternetapps.Feed{}, fmt.Errorf("parse template metadata.name: %w", err)
	}
	if !found || strings.TrimSpace(name) == "" {
		return clusternetapps.Feed{}, fmt.Errorf("template metadata.name is required")
	}

	namespace, found, err := unstructured.NestedString(template, "metadata", "namespace")
	if err != nil {
		return clusternetapps.Feed{}, fmt.Errorf("parse template metadata.namespace: %w", err)
	}
	if !found || strings.TrimSpace(namespace) == "" {
		return clusternetapps.Feed{}, fmt.Errorf("template metadata.namespace is required")
	}

	return clusternetapps.Feed{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Namespace:  namespace,
	}, nil
}
