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
	"testing"
	"time"

	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

func TestParseActionTimeout(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "default on empty", input: "", want: defaultWaitTimeout},
		{name: "valid 30s", input: "30s", want: 30 * time.Second},
		{name: "valid 2m", input: "2m", want: 2 * time.Minute},
		{name: "invalid string", input: "abc", wantErr: true},
		{name: "non-positive", input: "0s", wantErr: true},
		{name: "negative", input: "-1s", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseActionTimeout(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBindingCluster(t *testing.T) {
	ns, name, err := parseBindingCluster("clusternet-system/child-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ns != "clusternet-system" || name != "child-a" {
		t.Fatalf("got %s/%s", ns, name)
	}

	for _, bad := range []string{"invalid", "", "/"} {
		if _, _, err := parseBindingCluster(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestEvaluateResourceReadiness(t *testing.T) {
	t.Run("deployment ready", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "Deployment",
			"spec": map[string]interface{}{"replicas": int64(2)},
			"status": map[string]interface{}{
				"availableReplicas": int64(2),
				"updatedReplicas":   int64(2),
			},
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil || !ready {
			t.Fatalf("expected ready, got ready=%v err=%v", ready, err)
		}
	})

	t.Run("deployment not ready", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "Deployment",
			"spec": map[string]interface{}{"replicas": int64(3)},
			"status": map[string]interface{}{
				"availableReplicas": int64(1),
				"updatedReplicas":   int64(2),
			},
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ready {
			t.Fatal("expected not ready")
		}
	})

	t.Run("statefulset ready", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "StatefulSet",
			"spec": map[string]interface{}{"replicas": int64(3)},
			"status": map[string]interface{}{
				"readyReplicas": int64(3),
			},
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil || !ready {
			t.Fatalf("expected ready, got ready=%v err=%v", ready, err)
		}
	})

	t.Run("daemonset ready", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "DaemonSet",
			"status": map[string]interface{}{
				"desiredNumberScheduled": int64(5),
				"numberReady":            int64(5),
			},
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil || !ready {
			t.Fatalf("expected ready, got ready=%v err=%v", ready, err)
		}
	})

	t.Run("job complete", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "Job",
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{"type": "Complete", "status": "True"},
				},
			},
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil || !ready {
			t.Fatalf("expected ready, got ready=%v err=%v", ready, err)
		}
	})

	t.Run("job failed", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "Job",
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{"type": "Failed", "status": "True"},
				},
			},
		}}
		_, _, err := evaluateResourceReadiness(obj)
		if err == nil {
			t.Fatal("expected error for failed job")
		}
	})

	t.Run("job pending", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind":   "Job",
			"status": map[string]interface{}{},
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ready {
			t.Fatal("expected not ready")
		}
	})

	t.Run("configmap exists is ready", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "ConfigMap",
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil || !ready {
			t.Fatalf("expected ready, got ready=%v err=%v", ready, err)
		}
	})

	t.Run("service exists is ready", func(t *testing.T) {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"kind": "Service",
		}}
		ready, _, err := evaluateResourceReadiness(obj)
		if err != nil || !ready {
			t.Fatalf("expected ready, got ready=%v err=%v", ready, err)
		}
	})
}

// NOCC:tosa/fn_length(设计如此)
// NOCC:tosa/fn_length(设计如此)
func TestRenderSubscriptionActionRendersFeedsForWaitReady(t *testing.T) {
	executor := &SubscriptionActionExecutor{}
	templateData := &utils.TemplateData{
		Params: map[string]interface{}{
			"feedNamespace": "default",
			"feedName":      "demo-config",
		},
	}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	rendered, err := executor.renderSubscriptionAction(&clusternetapps.SubscriptionSpec{
		Feeds: []clusternetapps.Feed{
			{
				APIVersion: "v1",
				Kind:       "ConfigMap",
				Name:       "$(params.feedName)",
				Namespace:  "$(params.feedNamespace)",
			},
		},
	}, render)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(rendered.feeds) != 1 {
		t.Fatalf("expected 1 rendered feed, got %d", len(rendered.feeds))
	}
	if rendered.feeds[0].Name != "demo-config" {
		t.Fatalf("expected rendered feed name demo-config, got %q", rendered.feeds[0].Name)
	}
	if rendered.feeds[0].Namespace != "default" {
		t.Fatalf("expected rendered feed namespace default, got %q", rendered.feeds[0].Namespace)
	}
}

// NOCC:tosa/fn_length(设计如此)
// NOCC:tosa/fn_length(设计如此)
func TestRenderSubscriptionActionReturnsErrorForInvalidFeedNamespaceTemplate(t *testing.T) {
	executor := &SubscriptionActionExecutor{}
	templateData := &utils.TemplateData{Params: map[string]interface{}{}}
	render := func(s string) (string, error) { return utils.RenderTemplate(s, templateData) }

	_, err := executor.renderSubscriptionAction(&clusternetapps.SubscriptionSpec{
		Feeds: []clusternetapps.Feed{
			{
				APIVersion: "v1",
				Kind:       "ConfigMap",
				Name:       "demo-config",
				Namespace:  "$(params.feedNamespace)",
			},
		},
	}, render)
	if err == nil {
		t.Fatal("expected error for invalid feed namespace template")
	}
}

// NOCC:tosa/fn_length(设计如此)
// NOCC:tosa/fn_length(设计如此)
func TestCheckSubscriptionReadyOnceUsesRenderedFeeds(t *testing.T) {
	scheme := runtime.NewScheme()
	subscription := &unstructured.Unstructured{}
	subscription.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	subscription.SetNamespace("default")
	subscription.SetName("demo-sub")
	subscription.Object["status"] = map[string]interface{}{
		"bindingClusters": []interface{}{"clusternet-abc/child-a"},
	}

	managedCluster := &unstructured.Unstructured{}
	managedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "clusters.clusternet.io",
		Version: "v1beta1",
		Kind:    "ManagedCluster",
	})
	managedCluster.SetNamespace("clusternet-abc")
	managedCluster.SetName("child-a")
	managedCluster.Object["spec"] = map[string]interface{}{
		"clusterId": "cluster-1",
	}

	childResource := &unstructured.Unstructured{}
	childResource.SetGroupVersionKind(schema.GroupVersionKind{
		Version: "v1",
		Kind:    "ConfigMap",
	})
	childResource.SetNamespace("default")
	childResource.SetName("demo-config")

	executor := &SubscriptionActionExecutor{
		client: fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(subscription, managedCluster).Build(),
		childClientFactory: &fakeChildClusterClientFactory{
			clients: map[string]client.Client{
				"cluster-1": fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childResource).Build(),
			},
		},
	}

	ready, _, err := executor.checkSubscriptionReadyOnce(context.Background(), "default", "demo-sub", []clusternetapps.Feed{
		{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Name:       "demo-config",
			Namespace:  "default",
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !ready {
		t.Fatal("expected subscription feeds to be ready")
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestCheckSubscriptionReadyOnceReturnsDescriptionFailure(t *testing.T) {
	scheme := runtime.NewScheme()

	subscription := &unstructured.Unstructured{}
	subscription.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	subscription.SetNamespace("default")
	subscription.SetName("demo-sub")
	subscription.SetUID("sub-uid-1")
	subscription.Object["status"] = map[string]interface{}{
		"bindingClusters": []interface{}{"clusternet-abc/child-a"},
	}

	description := &unstructured.Unstructured{}
	description.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Description",
	})
	description.SetNamespace("clusternet-abc")
	description.SetName("demo-sub-generic")
	description.SetLabels(map[string]string{
		"apps.clusternet.io/subs.uid": "sub-uid-1",
	})
	description.Object["status"] = map[string]interface{}{
		"phase":  "Failure",
		"reason": "job invalid",
	}

	executor := &SubscriptionActionExecutor{
		client: fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(subscription, description).Build(),
		childClientFactory: &fakeChildClusterClientFactory{
			clients: map[string]client.Client{},
		},
	}

	ready, _, err := executor.checkSubscriptionReadyOnce(context.Background(), "default", "demo-sub", []clusternetapps.Feed{
		{
			APIVersion: "batch/v1",
			Kind:       "Job",
			Name:       "release-name-db-migrate",
			Namespace:  "default",
		},
	})
	if err == nil {
		t.Fatal("expected Description failure to be returned")
	}
	if ready {
		t.Fatal("expected ready=false when Description has failed")
	}
}

type fakeChildClusterClientFactory struct {
	clients map[string]client.Client
	err     error
}

func (f *fakeChildClusterClientFactory) GetChildClient(_ context.Context, clusterID, _ string) (client.Client, error) {
	if f.err != nil {
		return nil, f.err
	}
	childClient, ok := f.clients[clusterID]
	if !ok {
		return nil, fmt.Errorf("child client for cluster %s not found", clusterID)
	}
	return childClient, nil
}
