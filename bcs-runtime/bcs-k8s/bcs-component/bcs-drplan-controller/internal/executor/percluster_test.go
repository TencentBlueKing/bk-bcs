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
	"testing"

	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

func TestIsPerClusterMode(t *testing.T) {
	tests := []struct {
		name     string
		action   *drv1alpha1.Action
		expected bool
	}{
		{
			name: "all three conditions met: PerCluster + Subscription + waitReady",
			action: &drv1alpha1.Action{
				Type:                 drv1alpha1.ActionTypeSubscription,
				WaitReady:            true,
				ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			},
			expected: true,
		},
		{
			name: "empty clusterExecutionMode defaults to Global",
			action: &drv1alpha1.Action{
				Type:                 drv1alpha1.ActionTypeSubscription,
				WaitReady:            true,
				ClusterExecutionMode: "",
			},
			expected: false,
		},
		{
			name: "explicit Global mode",
			action: &drv1alpha1.Action{
				Type:                 drv1alpha1.ActionTypeSubscription,
				WaitReady:            true,
				ClusterExecutionMode: drv1alpha1.ClusterExecutionModeGlobal,
			},
			expected: false,
		},
		{
			name: "PerCluster but not Subscription type",
			action: &drv1alpha1.Action{
				Type:                 drv1alpha1.ActionTypeJob,
				WaitReady:            true,
				ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			},
			expected: false,
		},
		{
			name: "PerCluster + Subscription but waitReady=false",
			action: &drv1alpha1.Action{
				Type:                 drv1alpha1.ActionTypeSubscription,
				WaitReady:            false,
				ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			},
			expected: false,
		},
		{
			name: "backward compat: legacy action with no new fields",
			action: &drv1alpha1.Action{
				Type: drv1alpha1.ActionTypeSubscription,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPerClusterMode(tt.action)
			if got != tt.expected {
				t.Errorf("isPerClusterMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func newSubscriptionGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "apps.clusternet.io", Version: "v1alpha1", Kind: "Subscription"}
}

// NOCC:tosa/fn_length(设计如此)
func TestCreateChildSubscription(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
	executor := &SubscriptionActionExecutor{client: fakeClient}

	feedsList := []interface{}{
		map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"name":       "db-migrate",
			"namespace":  "demo-ns",
		},
	}

	specMap := map[string]interface{}{
		"schedulingStrategy": "Replication",
		"feeds":              feedsList,
	}

	t.Run("builds child with correct name and single-cluster subscriber", func(t *testing.T) {
		child, err := executor.buildChildSubscription(
			"demo-app-db-migrate-sub", "default",
			"clusternet-abc12/clusternet-cluster-v5x8k", "cluster-id-123",
			feedsList, specMap,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedName := "demo-app-db-migrate-sub--clusternet-cluster-v5x8k"
		if child.GetName() != expectedName {
			t.Errorf("name = %q, want %q", child.GetName(), expectedName)
		}
		if child.GetNamespace() != "default" {
			t.Errorf("namespace = %q, want %q", child.GetNamespace(), "default")
		}

		subscribers, found, _ := unstructured.NestedSlice(child.Object, "spec", "subscribers")
		if !found || len(subscribers) != 1 {
			t.Fatalf("subscribers count = %d, want 1", len(subscribers))
		}
		subscriber, ok := subscribers[0].(map[string]interface{})
		if !ok {
			t.Fatalf("subscriber type = %T, want map[string]interface{}", subscribers[0])
		}
		clusterAffinity, ok := subscriber["clusterAffinity"].(map[string]interface{})
		if !ok {
			t.Fatalf("clusterAffinity type = %T, want map[string]interface{}", subscriber["clusterAffinity"])
		}
		matchExpressions, ok := clusterAffinity["matchExpressions"].([]interface{})
		if !ok || len(matchExpressions) != 1 {
			t.Fatalf("matchExpressions = %v, want one entry", clusterAffinity["matchExpressions"])
		}
		expr, ok := matchExpressions[0].(map[string]interface{})
		if !ok {
			t.Fatalf("matchExpression type = %T, want map[string]interface{}", matchExpressions[0])
		}
		values, ok := expr["values"].([]interface{})
		if !ok || len(values) != 1 || values[0] != "cluster-id-123" {
			t.Fatalf("clusterAffinity values = %v, want [cluster-id-123]", expr["values"])
		}
	})

	t.Run("truncates long names to 253 chars", func(t *testing.T) {
		longName := ""
		for i := 0; i < 250; i++ {
			longName += "a"
		}

		child, err := executor.buildChildSubscription(longName, "default", "ns/cluster-name", "cluster-id-long", feedsList, specMap)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(child.GetName()) > 253 {
			t.Errorf("name length = %d, want <= 253", len(child.GetName()))
		}
	})
}

func TestResolveTargetClusters(t *testing.T) {
	scheme := runtime.NewScheme()

	clusterA := &unstructured.Unstructured{}
	clusterA.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "clusters.clusternet.io",
		Version: "v1beta1",
		Kind:    "ManagedCluster",
	})
	clusterA.SetNamespace("clusternet-ns1")
	clusterA.SetName("cluster-a")
	clusterA.SetLabels(map[string]string{"env": "prod"})

	clusterB := &unstructured.Unstructured{}
	clusterB.SetGroupVersionKind(clusterA.GroupVersionKind())
	clusterB.SetNamespace("clusternet-ns2")
	clusterB.SetName("cluster-b")
	clusterB.SetLabels(map[string]string{"env": "test"})

	t.Run("matches subscribers against ManagedCluster labels", func(t *testing.T) {
		fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(clusterA, clusterB).Build()
		executor := &SubscriptionActionExecutor{client: fakeClient}

		clusters, err := executor.resolveTargetClusters(context.Background(), &clusternetapps.SubscriptionSpec{
			Subscribers: []clusternetapps.Subscriber{{ClusterAffinity: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(clusters) != 1 {
			t.Fatalf("clusters count = %d, want 1", len(clusters))
		}
		if clusters[0] != "clusternet-ns1/cluster-a" {
			t.Errorf("clusters[0] = %q, want %q", clusters[0], "clusternet-ns1/cluster-a")
		}
	})

	t.Run("empty subscribers match all clusters", func(t *testing.T) {
		fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(clusterA, clusterB).Build()
		executor := &SubscriptionActionExecutor{client: fakeClient}

		clusters, err := executor.resolveTargetClusters(context.Background(), &clusternetapps.SubscriptionSpec{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(clusters) != 2 {
			t.Fatalf("clusters count = %d, want 2", len(clusters))
		}
	})

	t.Run("returns error when selector matches no clusters", func(t *testing.T) {
		fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(clusterA, clusterB).Build()
		executor := &SubscriptionActionExecutor{client: fakeClient}

		_, err := executor.resolveTargetClusters(context.Background(), &clusternetapps.SubscriptionSpec{
			Subscribers: []clusternetapps.Subscriber{{ClusterAffinity: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "staging"},
			}}},
		})
		if err == nil {
			t.Fatal("expected selector mismatch error")
		}
	})
}

// NOCC:tosa/fn_length(设计如此)
func TestExecutePerClusterCreateChildSub(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	executor := &SubscriptionActionExecutor{
		client:             fakeClient,
		childClientFactory: nil,
	}

	feeds := []interface{}{
		map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"name":       "test-job",
			"namespace":  "test-ns",
		},
	}
	specMap := map[string]interface{}{
		"schedulingStrategy": "Replication",
		"feeds":              feeds,
	}

	t.Run("creates child subscription via client", func(t *testing.T) {
		child, err := executor.buildChildSubscription("parent-sub", "default", "ns1/cluster-a", "cluster-a-id", feeds, specMap)
		if err != nil {
			t.Fatalf("buildChildSubscription failed: %v", err)
		}

		err = fakeClient.Create(context.Background(), child)
		if err != nil {
			t.Fatalf("create child subscription failed: %v", err)
		}

		got := &unstructured.Unstructured{}
		got.SetGroupVersionKind(newSubscriptionGVK())
		err = fakeClient.Get(context.Background(), client.ObjectKey{
			Namespace: "default",
			Name:      "parent-sub--cluster-a",
		}, got)
		if err != nil {
			t.Fatalf("get child subscription failed: %v", err)
		}
		if got.GetName() != "parent-sub--cluster-a" {
			t.Errorf("child name = %q, want %q", got.GetName(), "parent-sub--cluster-a")
		}
	})
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteForCluster(t *testing.T) {
	scheme := runtime.NewScheme()

	t.Run("creates child subscription for the given cluster", func(t *testing.T) {
		managedCluster := &unstructured.Unstructured{}
		managedCluster.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "clusters.clusternet.io",
			Version: "v1beta1",
			Kind:    "ManagedCluster",
		})
		managedCluster.SetNamespace("ns1")
		managedCluster.SetName("cluster-a")
		managedCluster.Object["spec"] = map[string]interface{}{
			"clusterId": "cluster-a-id",
		}

		fakeClient := fakeclient.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(managedCluster).
			Build()

		childJob := &unstructured.Unstructured{}
		childJob.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "batch",
			Version: "v1",
			Kind:    "Job",
		})
		childJob.SetNamespace("app-ns")
		childJob.SetName("hook-job")
		childJob.Object["status"] = map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{"type": "Complete", "status": "True"},
			},
		}

		executor := &SubscriptionActionExecutor{
			client: fakeClient,
			childClientFactory: &fakeChildClusterClientFactory{
				clients: map[string]client.Client{
					"cluster-a-id": fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childJob).Build(),
				},
			},
		}

		action := &drv1alpha1.Action{
			Name:                 "hook-action",
			Type:                 drv1alpha1.ActionTypeSubscription,
			WaitReady:            true,
			ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			Subscription: &drv1alpha1.SubscriptionAction{
				Name:      "hook-sub",
				Namespace: "default",
				Spec: &clusternetapps.SubscriptionSpec{
					SchedulingStrategy: clusternetapps.ReplicaSchedulingStrategyType,
					Feeds: []clusternetapps.Feed{
						{APIVersion: "batch/v1", Kind: "Job", Name: "hook-job", Namespace: "app-ns"},
					},
				},
			},
			Timeout: "10s",
		}

		cs, childRef, err := executor.ExecuteForCluster(context.Background(), action, "ns1/cluster-a", nil)

		if cs == nil {
			t.Fatal("expected non-nil ClusterActionStatus")
		}

		// Child should exist in the fake client
		child := &unstructured.Unstructured{}
		child.SetGroupVersionKind(newSubscriptionGVK())
		getErr := fakeClient.Get(context.Background(), client.ObjectKey{
			Namespace: "default", Name: "hook-sub--cluster-a",
		}, child)
		if getErr != nil {
			t.Fatalf("child subscription not created: %v", getErr)
		}
		if child.GetName() != "hook-sub--cluster-a" {
			t.Errorf("child name = %q, want %q", child.GetName(), "hook-sub--cluster-a")
		}
		if childRef == nil || childRef.Name != "hook-sub--cluster-a" {
			t.Fatalf("unexpected childRef: %#v", childRef)
		}

		if cs.Phase != drv1alpha1.PhaseSucceeded {
			t.Errorf("phase = %q, want %q", cs.Phase, drv1alpha1.PhaseSucceeded)
		}
		if cs.ClusterID != "cluster-a-id" {
			t.Errorf("clusterID = %q, want %q", cs.ClusterID, "cluster-a-id")
		}
		if err != nil {
			t.Fatalf("ExecuteForCluster returned error: %v", err)
		}
	})

	t.Run("returns error for invalid binding cluster", func(t *testing.T) {
		fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
		executor := &SubscriptionActionExecutor{
			client:             fakeClient,
			childClientFactory: nil,
		}

		action := &drv1alpha1.Action{
			Name:                 "test",
			Type:                 drv1alpha1.ActionTypeSubscription,
			WaitReady:            true,
			ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			Subscription: &drv1alpha1.SubscriptionAction{
				Name: "test-sub",
				Spec: &clusternetapps.SubscriptionSpec{},
			},
		}

		cs, _, err := executor.ExecuteForCluster(context.Background(), action, "invalid-no-slash", nil)
		if err == nil {
			t.Error("expected error for invalid binding cluster")
		}
		if cs == nil || cs.Phase != drv1alpha1.PhaseFailed {
			t.Error("expected Failed phase for invalid binding")
		}
	})

	t.Run("returns error when target ManagedCluster not found", func(t *testing.T) {
		fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
		executor := &SubscriptionActionExecutor{
			client:             fakeClient,
			childClientFactory: nil,
		}

		action := &drv1alpha1.Action{
			Name:                 "test",
			Type:                 drv1alpha1.ActionTypeSubscription,
			WaitReady:            true,
			ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			Subscription: &drv1alpha1.SubscriptionAction{
				Name:      "nonexistent-sub",
				Namespace: "default",
				Spec: &clusternetapps.SubscriptionSpec{
					SchedulingStrategy: clusternetapps.ReplicaSchedulingStrategyType,
				},
			},
		}

		cs, _, err := executor.ExecuteForCluster(context.Background(), action, "ns1/cluster-a", nil)
		if err == nil {
			t.Error("expected error when ManagedCluster not found")
		}
		if cs == nil || cs.Phase != drv1alpha1.PhaseFailed {
			t.Error("expected Failed phase when ManagedCluster not found")
		}
	})
}

// NOCC:tosa/fn_length(设计如此)
func TestPerClusterRollbackDeletesChildren(t *testing.T) {
	scheme := runtime.NewScheme()
	childA := &unstructured.Unstructured{}
	childA.SetGroupVersionKind(newSubscriptionGVK())
	childA.SetName("hook-sub--cluster-a")
	childA.SetNamespace("default")

	childB := &unstructured.Unstructured{}
	childB.SetGroupVersionKind(newSubscriptionGVK())
	childB.SetName("hook-sub--cluster-b")
	childB.SetNamespace("default")

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(childA, childB).Build()

	executor := &SubscriptionActionExecutor{
		client:             fakeClient,
		childClientFactory: nil,
	}

	action := &drv1alpha1.Action{
		Name:                 "hook-action",
		Type:                 drv1alpha1.ActionTypeSubscription,
		WaitReady:            true,
		ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
	}

	prevStatus := &drv1alpha1.ActionStatus{
		Name:  "hook-action",
		Phase: drv1alpha1.PhaseSucceeded,
		Outputs: &drv1alpha1.ActionOutputs{
			SubscriptionRefs: []corev1.ObjectReference{
				{Kind: "Subscription", APIVersion: "apps.clusternet.io/v1alpha1", Namespace: "default", Name: "hook-sub--cluster-a"},
				{Kind: "Subscription", APIVersion: "apps.clusternet.io/v1alpha1", Namespace: "default", Name: "hook-sub--cluster-b"},
			},
		},
	}

	t.Run("rollback deletes all child subscriptions", func(t *testing.T) {
		rbStatus, err := executor.Rollback(context.Background(), action, prevStatus, nil)
		if err != nil {
			t.Fatalf("rollback failed: %v", err)
		}
		if rbStatus.Phase != drv1alpha1.PhaseSucceeded {
			t.Errorf("rollback phase = %q, want %q", rbStatus.Phase, drv1alpha1.PhaseSucceeded)
		}

		check := &unstructured.Unstructured{}
		check.SetGroupVersionKind(newSubscriptionGVK())
		for _, name := range []string{"hook-sub--cluster-a", "hook-sub--cluster-b"} {
			err = fakeClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: name}, check)
			if err == nil {
				t.Errorf("expected child subscription %s to be deleted, but it still exists", name)
			}
		}
	})
}
