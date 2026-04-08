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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
)

// patchRecordingClient wraps a fake client and records if Patch was called with Apply type.
type patchRecordingClient struct {
	client.Client
	patchCalled  bool
	createCalled bool
}

func (c *patchRecordingClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	c.patchCalled = true
	return nil // Simulate success
}

func (c *patchRecordingClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	c.createCalled = true
	return c.Client.Create(ctx, obj, opts...)
}

var subscriptionGVK = schema.GroupVersionKind{
	Group:   "apps.clusternet.io",
	Version: "v1alpha1",
	Kind:    "Subscription",
}

func newTestScheme() *runtime.Scheme {
	sc := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(sc)
	return sc
}

func buildSubscriptionAction(operation string) drv1alpha1.Action {
	return drv1alpha1.Action{
		Name: "test-sub",
		Type: "Subscription",
		Subscription: &drv1alpha1.SubscriptionAction{
			Operation: operation,
			Name:      "my-sub",
			Namespace: "default",
			Spec: &clusternetapps.SubscriptionSpec{
				Subscribers: []clusternetapps.Subscriber{},
				Feeds:       []clusternetapps.Feed{},
			},
		},
	}
}

// TestSubscriptionExecutor_Create verifies default Create operation creates new Subscription.
func TestSubscriptionExecutor_Create(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &SubscriptionActionExecutor{client: fakeClient}

	action := buildSubscriptionAction("Create")
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
}

// TestSubscriptionExecutor_Create_AlreadyExists verifies Create fails when Subscription already exists.
// NOCC:tosa/fn_length(设计如此)
func TestSubscriptionExecutor_Create_AlreadyExists(t *testing.T) {
	sc := newTestScheme()

	// Pre-create the Subscription
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(subscriptionGVK)
	existing.SetName("my-sub")
	existing.SetNamespace("default")

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &SubscriptionActionExecutor{client: fakeClient}

	action := buildSubscriptionAction("Create")
	_, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err == nil {
		t.Error("expected AlreadyExists error, got nil")
	}
}

// TestSubscriptionExecutor_Apply_UsesPatch verifies Apply operation calls Patch (not Create).
// NOCC:tosa/fn_length(设计如此)
func TestSubscriptionExecutor_Apply_UsesPatch(t *testing.T) {
	sc := newTestScheme()
	recorder := &patchRecordingClient{
		Client: fakeclient.NewClientBuilder().WithScheme(sc).Build(),
	}
	ex := &SubscriptionActionExecutor{client: recorder}

	action := buildSubscriptionAction("Apply")
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Apply operation failed: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if !recorder.patchCalled {
		t.Error("expected Patch to be called for operation=Apply, but it was not")
	}
	if recorder.createCalled {
		t.Error("expected Create NOT to be called for operation=Apply")
	}
}

// TestSubscriptionExecutor_Create_UsesCreate verifies default Create operation calls Create (not Patch).
// NOCC:tosa/fn_length(设计如此)
func TestSubscriptionExecutor_Create_UsesCreate(t *testing.T) {
	sc := newTestScheme()
	recorder := &patchRecordingClient{
		Client: fakeclient.NewClientBuilder().WithScheme(sc).Build(),
	}
	ex := &SubscriptionActionExecutor{client: recorder}

	action := buildSubscriptionAction("")
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Create operation failed: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if recorder.patchCalled {
		t.Error("expected Patch NOT to be called for default Create operation")
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestSubscriptionExecutor_Delete_DoesNotRequireSpec(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(subscriptionGVK)
	existing.SetName("my-sub")
	existing.SetNamespace("default")

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &SubscriptionActionExecutor{client: fakeClient}

	action := drv1alpha1.Action{
		Name: "delete-sub",
		Type: drv1alpha1.ActionTypeSubscription,
		Subscription: &drv1alpha1.SubscriptionAction{
			Operation: drv1alpha1.OperationDelete,
			Name:      "my-sub",
			Namespace: "default",
		},
	}

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	sub := &unstructured.Unstructured{}
	sub.SetGroupVersionKind(subscriptionGVK)
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "my-sub", Namespace: "default"}, sub)
	if err == nil {
		t.Fatal("expected Subscription to be deleted")
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestSubscriptionExecutor_BeforeCreateCleanupDeletesExistingSubscription(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(subscriptionGVK)
	existing.SetName("my-sub")
	existing.SetNamespace("default")

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &SubscriptionActionExecutor{client: fakeClient}

	action := buildSubscriptionAction(drv1alpha1.OperationCreate)
	action.HookCleanup = &drv1alpha1.HookCleanupPolicy{BeforeCreate: true}

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestSubscriptionExecutor_OnSuccessCleanupDeletesCreatedSubscription(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &SubscriptionActionExecutor{client: fakeClient}

	action := buildSubscriptionAction(drv1alpha1.OperationCreate)
	action.HookCleanup = &drv1alpha1.HookCleanupPolicy{OnSuccess: true}

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	sub := &unstructured.Unstructured{}
	sub.SetGroupVersionKind(subscriptionGVK)
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "my-sub", Namespace: "default"}, sub)
	if err == nil {
		t.Fatal("expected Subscription to be deleted on success cleanup")
	}
}
