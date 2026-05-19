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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
)

var localizationGVK = schema.GroupVersionKind{
	Group:   "apps.clusternet.io",
	Version: "v1alpha1",
	Kind:    "Localization",
}

func buildLocalizationAction(operation string) drv1alpha1.Action {
	return drv1alpha1.Action{
		Name: "test-localization",
		Type: drv1alpha1.ActionTypeLocalization,
		Localization: &drv1alpha1.LocalizationAction{
			Operation: operation,
			Name:      "demo-localization",
			Namespace: "cluster-a",
			Spec: &clusternetapps.LocalizationSpec{
				Priority: 500,
				Feed: clusternetapps.Feed{
					APIVersion: "apps.clusternet.io/v1alpha1",
					Kind:       "HelmChart",
					Namespace:  "default",
					Name:       "demo-app",
				},
			},
		},
	}
}

func TestLocalizationExecutor_Create(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &LocalizationActionExecutor{client: fakeClient}

	action := buildLocalizationAction("")
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if status.Outputs == nil || status.Outputs.LocalizationRef == nil {
		t.Fatalf("expected LocalizationRef to be recorded, got %#v", status.Outputs)
	}

	loc := &unstructured.Unstructured{}
	loc.SetGroupVersionKind(localizationGVK)
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-localization", Namespace: "cluster-a"}, loc); err != nil {
		t.Fatalf("expected Localization to be created: %v", err)
	}
}

func TestLocalizationExecutor_Apply_UsesPatch(t *testing.T) {
	sc := newTestScheme()
	recorder := &patchRecordingClient{
		Client: fakeclient.NewClientBuilder().WithScheme(sc).Build(),
	}
	ex := &LocalizationActionExecutor{client: recorder}

	action := buildLocalizationAction(drv1alpha1.OperationApply)
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Apply operation failed: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if !recorder.patchCalled {
		t.Fatal("expected Patch to be called for operation=Apply")
	}
	if recorder.createCalled {
		t.Fatal("expected Create not to be called for operation=Apply")
	}
}

func TestLocalizationExecutor_Delete_DoesNotRequireSpec(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(localizationGVK)
	existing.SetName("demo-localization")
	existing.SetNamespace("cluster-a")
	existing.Object["spec"] = map[string]interface{}{
		"priority": int64(500),
		"feed": map[string]interface{}{
			"apiVersion": "apps.clusternet.io/v1alpha1",
			"kind":       "HelmChart",
			"name":       "demo-app",
			"namespace":  "default",
		},
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &LocalizationActionExecutor{client: fakeClient}

	action := drv1alpha1.Action{
		Name: "delete-localization",
		Type: drv1alpha1.ActionTypeLocalization,
		Localization: &drv1alpha1.LocalizationAction{
			Operation: drv1alpha1.OperationDelete,
			Name:      "demo-localization",
			Namespace: "cluster-a",
		},
	}

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	loc := &unstructured.Unstructured{}
	loc.SetGroupVersionKind(localizationGVK)
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-localization", Namespace: "cluster-a"}, loc)
	if err == nil {
		t.Fatal("expected Localization to be deleted")
	}
}

func TestLocalizationExecutor_Patch_UpdatesExistingSpec(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(localizationGVK)
	existing.SetName("demo-localization")
	existing.SetNamespace("cluster-a")
	existing.Object["spec"] = map[string]interface{}{
		"priority":       int64(100),
		"overridePolicy": "ApplyLater",
		"feed": map[string]interface{}{
			"apiVersion": "apps.clusternet.io/v1alpha1",
			"kind":       "HelmChart",
			"name":       "demo-app",
			"namespace":  "default",
		},
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &LocalizationActionExecutor{client: fakeClient}

	action := buildLocalizationAction(drv1alpha1.OperationPatch)
	action.Rollback = &drv1alpha1.Action{
		Name: "rollback",
		Type: drv1alpha1.ActionTypeLocalization,
		Localization: &drv1alpha1.LocalizationAction{
			Operation: drv1alpha1.OperationDelete,
			Name:      "demo-localization",
			Namespace: "cluster-a",
		},
	}
	action.Localization.Spec.Priority = 700
	action.Localization.Spec.OverridePolicy = clusternetapps.ApplyNow

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(localizationGVK)
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-localization", Namespace: "cluster-a"}, updated); err != nil {
		t.Fatalf("get updated Localization: %v", err)
	}

	priority, found, err := unstructured.NestedInt64(updated.Object, "spec", "priority")
	if err != nil || !found {
		t.Fatalf("expected spec.priority to exist, found=%v err=%v", found, err)
	}
	if priority != 700 {
		t.Fatalf("expected patched priority=700, got %d", priority)
	}
}

func TestLocalizationExecutor_Rollback_CreateDeletesResource(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &LocalizationActionExecutor{client: fakeClient}

	action := buildLocalizationAction(drv1alpha1.OperationCreate)
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("execute create action: %v", err)
	}

	rollbackStatus, err := ex.Rollback(context.Background(), &action, status, map[string]interface{}{})
	if err != nil {
		t.Fatalf("rollback create action: %v", err)
	}
	if rollbackStatus.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected rollback succeeded, got %s: %s", rollbackStatus.Phase, rollbackStatus.Message)
	}

	loc := &unstructured.Unstructured{}
	loc.SetGroupVersionKind(localizationGVK)
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-localization", Namespace: "cluster-a"}, loc)
	if err == nil {
		t.Fatal("expected Localization to be deleted by rollback")
	}
}
