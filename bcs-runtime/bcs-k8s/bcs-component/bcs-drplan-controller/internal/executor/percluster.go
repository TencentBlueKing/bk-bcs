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
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

const (
	childSeparator      = "--"
	maxK8sNameLength    = 253
	subscriptionGroup   = "apps.clusternet.io"
	subscriptionVersion = "v1alpha1"
	subscriptionKind    = "Subscription"
)

// isPerClusterMode returns true only when all three conditions are met:
// 1. ClusterExecutionMode == PerCluster
// 2. Type == Subscription
// 3. WaitReady == true
func isPerClusterMode(action *drv1alpha1.Action) bool {
	return action.ClusterExecutionMode == drv1alpha1.ClusterExecutionModePerCluster &&
		action.Type == drv1alpha1.ActionTypeSubscription &&
		action.WaitReady
}

// buildChildSubscription constructs an unstructured child Subscription scoped to a single cluster.
// The child inherits the rendered namespace and feeds, but has a single-cluster subscriber.
func (e *SubscriptionActionExecutor) buildChildSubscription(
	parentName, parentNamespace string,
	clusterBinding string,
	clusterID string,
	feeds []interface{},
	parentSpecMap map[string]interface{},
) (*unstructured.Unstructured, error) {
	_, clusterName, err := parseBindingCluster(clusterBinding)
	if err != nil {
		return nil, fmt.Errorf("parse binding cluster %q: %w", clusterBinding, err)
	}

	childName := truncateK8sName(parentName + childSeparator + clusterName)

	child := &unstructured.Unstructured{}
	child.SetGroupVersionKind(schema.GroupVersionKind{
		Group: subscriptionGroup, Version: subscriptionVersion, Kind: subscriptionKind,
	})
	child.SetName(childName)
	child.SetNamespace(parentNamespace)

	childSpec := make(map[string]interface{})
	for k, v := range parentSpecMap {
		if k != "subscribers" {
			childSpec[k] = v
		}
	}

	childSpec["subscribers"] = []interface{}{
		map[string]interface{}{
			"clusterAffinity": map[string]interface{}{
				"matchExpressions": []interface{}{
					map[string]interface{}{
						"key":      "clusters.clusternet.io/cluster-id",
						"operator": "In",
						"values":   []interface{}{clusterID},
					},
				},
			},
		},
	}

	childSpec["schedulingStrategy"] = "Replication"
	childSpec["feeds"] = feeds

	child.Object["spec"] = childSpec
	return child, nil
}

func (e *SubscriptionActionExecutor) resolveTargetClusters(
	ctx context.Context,
	spec *clusternetapps.SubscriptionSpec,
) ([]string, error) {
	clusterList := &unstructured.UnstructuredList{}
	clusterList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "clusters.clusternet.io",
		Version: "v1beta1",
		Kind:    "ManagedClusterList",
	})
	if err := e.client.List(ctx, clusterList); err != nil {
		return nil, fmt.Errorf("list ManagedClusters: %w", err)
	}

	subscribers := spec.Subscribers
	if len(subscribers) == 0 {
		subscribers = []clusternetapps.Subscriber{{ClusterAffinity: &metav1.LabelSelector{}}}
	}

	targets := make(map[string]struct{})
	for _, subscriber := range subscribers {
		selector, err := metav1.LabelSelectorAsSelector(subscriber.ClusterAffinity)
		if err != nil {
			return nil, fmt.Errorf("parse subscriber clusterAffinity: %w", err)
		}
		for i := range clusterList.Items {
			cluster := &clusterList.Items[i]
			if !selector.Matches(labels.Set(cluster.GetLabels())) {
				continue
			}
			targets[cluster.GetNamespace()+"/"+cluster.GetName()] = struct{}{}
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no ManagedClusters matched subscription subscribers")
	}

	bindings := make([]string, 0, len(targets))
	for binding := range targets {
		bindings = append(bindings, binding)
	}
	sort.Strings(bindings)
	return bindings, nil
}

func truncateK8sName(name string) string {
	if len(name) <= maxK8sNameLength {
		return name
	}
	return strings.TrimRight(name[:maxK8sNameLength], "-.")
}

// ExecuteForCluster creates a child Subscription for a single cluster and waits for it to become ready.
// NOCC:tosa/fn_length(设计如此)
func (e *SubscriptionActionExecutor) ExecuteForCluster( //nolint:funlen
	ctx context.Context,
	action *drv1alpha1.Action,
	clusterBinding string,
	params map[string]interface{},
) (*drv1alpha1.ClusterActionStatus, *corev1.ObjectReference, error) {
	clusterNS, clusterName, parseErr := parseBindingCluster(clusterBinding)
	if parseErr != nil {
		return &drv1alpha1.ClusterActionStatus{
			Cluster: clusterBinding, Phase: drv1alpha1.PhaseFailed,
			StartTime: timeNowPtr(), CompletionTime: timeNowPtr(),
			Message: fmt.Sprintf("invalid binding cluster: %v", parseErr),
		}, nil, parseErr
	}
	clusterID := clusterNS + "/" + clusterName

	cs := &drv1alpha1.ClusterActionStatus{
		Cluster: clusterBinding, ClusterID: clusterID,
		Phase: drv1alpha1.PhaseRunning, StartTime: timeNowPtr(),
	}

	if err := e.validateSubscriptionConfig(action); err != nil {
		cs.Phase = drv1alpha1.PhaseFailed
		cs.CompletionTime = timeNowPtr()
		cs.Message = err.Error()
		return cs, nil, err
	}

	templateData := &utils.TemplateData{Params: params}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	subName, subNS, err := e.renderSubscriptionNameAndNamespace(action, render)
	if err != nil {
		cs.Phase = drv1alpha1.PhaseFailed
		cs.CompletionTime = timeNowPtr()
		cs.Message = fmt.Sprintf("render name failed: %v", err)
		return cs, nil, err
	}

	renderedSub, err := e.renderSubscriptionAction(action.Subscription.Spec, render)
	if err != nil {
		cs.Phase = drv1alpha1.PhaseFailed
		cs.CompletionTime = timeNowPtr()
		cs.Message = fmt.Sprintf("render spec failed: %v", err)
		return cs, nil, err
	}

	managedClusterID, err := e.getManagedClusterID(ctx, clusterNS, clusterName)
	if err != nil {
		cs.Phase = drv1alpha1.PhaseFailed
		cs.CompletionTime = timeNowPtr()
		cs.Message = fmt.Sprintf("get ManagedCluster %s/%s clusterId: %v", clusterNS, clusterName, err)
		return cs, nil, err
	}
	cs.ClusterID = managedClusterID

	feeds, _ := renderedSub.specMap["feeds"].([]interface{})

	child, err := e.buildChildSubscription(subName, subNS, clusterBinding, managedClusterID, feeds, renderedSub.specMap)
	if err != nil {
		cs.Phase = drv1alpha1.PhaseFailed
		cs.CompletionTime = timeNowPtr()
		cs.Message = fmt.Sprintf("build child Subscription: %v", err)
		return cs, nil, err
	}
	childRef := &corev1.ObjectReference{
		APIVersion: "apps.clusternet.io/v1alpha1",
		Kind:       "Subscription",
		Namespace:  child.GetNamespace(),
		Name:       child.GetName(),
		UID:        child.GetUID(),
	}

	createErr := e.client.Create(ctx, child)
	if createErr != nil {
		if !errors.IsAlreadyExists(createErr) {
			cs.Phase = drv1alpha1.PhaseFailed
			cs.CompletionTime = timeNowPtr()
			cs.Message = fmt.Sprintf("create child Subscription: %v", createErr)
			return cs, nil, createErr
		}
		klog.V(3).Infof("Child Subscription %s/%s already exists, reusing", child.GetNamespace(), child.GetName())
	}

	waitDur, parseTimeoutErr := parseActionTimeout(action.Timeout)
	if parseTimeoutErr != nil {
		cs.Phase = drv1alpha1.PhaseFailed
		cs.CompletionTime = timeNowPtr()
		cs.Message = fmt.Sprintf("invalid timeout: %v", parseTimeoutErr)
		return cs, childRef, parseTimeoutErr
	}

	if waitErr := e.waitForChildSubscriptionReady(ctx, child, clusterBinding, renderedSub.feeds, waitDur); waitErr != nil {
		cs.Phase = drv1alpha1.PhaseFailed
		cs.CompletionTime = timeNowPtr()
		cs.Message = fmt.Sprintf("waitReady failed for child %s/%s: %v", child.GetNamespace(), child.GetName(), waitErr)
		return cs, childRef, waitErr
	}

	cs.Phase = drv1alpha1.PhaseSucceeded
	cs.CompletionTime = timeNowPtr()
	cs.Message = fmt.Sprintf("child Subscription %s/%s ready", child.GetNamespace(), child.GetName())
	return cs, childRef, nil
}

func (e *SubscriptionActionExecutor) waitForChildSubscriptionReady(
	ctx context.Context,
	child *unstructured.Unstructured,
	clusterBinding string,
	feeds []clusternetapps.Feed,
	timeout time.Duration,
) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(waitPollInterval)
	defer ticker.Stop()

	for {
		ready, reason, err := e.evaluateSubscriptionReadiness(waitCtx, child, feeds, []string{clusterBinding})
		if err != nil {
			return err
		}
		if ready {
			klog.Infof("Child Subscription %s/%s waitReady succeeded: %s", child.GetNamespace(), child.GetName(), reason)
			return nil
		}
		klog.V(3).Infof("Child Subscription %s/%s waitReady pending: %s", child.GetNamespace(), child.GetName(), reason)

		select {
		case <-waitCtx.Done():
			return fmt.Errorf("timeout waiting for child Subscription %s/%s ready: %w",
				child.GetNamespace(), child.GetName(), waitCtx.Err())
		case <-ticker.C:
		}
	}
}
