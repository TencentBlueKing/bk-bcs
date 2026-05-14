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

var globalizationGVK = schema.GroupVersionKind{
	Group:   "apps.clusternet.io",
	Version: "v1alpha1",
	Kind:    "Globalization",
}

func buildGlobalizationAction(operation string) drv1alpha1.Action {
	return drv1alpha1.Action{
		Name: "test-globalization",
		Type: drv1alpha1.ActionTypeGlobalization,
		Globalization: &drv1alpha1.GlobalizationAction{
			Operation: operation,
			Name:      "demo-global-values",
			Spec: &clusternetapps.GlobalizationSpec{
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

func TestGlobalizationExecutor_Create(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &GlobalizationActionExecutor{client: fakeClient}

	action := buildGlobalizationAction("")
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if status.Outputs == nil || status.Outputs.GlobalizationRef == nil {
		t.Fatalf("expected GlobalizationRef to be recorded, got %#v", status.Outputs)
	}
	if status.Outputs.GlobalizationRef.Namespace != "" {
		t.Fatalf("expected cluster-scoped reference, got namespace %q", status.Outputs.GlobalizationRef.Namespace)
	}

	glob := &unstructured.Unstructured{}
	glob.SetGroupVersionKind(globalizationGVK)
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-global-values"}, glob); err != nil {
		t.Fatalf("expected Globalization to be created: %v", err)
	}
}

func TestGlobalizationExecutor_Apply_UsesPatch(t *testing.T) {
	sc := newTestScheme()
	recorder := &patchRecordingClient{
		Client: fakeclient.NewClientBuilder().WithScheme(sc).Build(),
	}
	ex := &GlobalizationActionExecutor{client: recorder}

	action := buildGlobalizationAction(drv1alpha1.OperationApply)
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

func TestGlobalizationExecutor_Delete_DoesNotRequireSpec(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(globalizationGVK)
	existing.SetName("demo-global-values")
	existing.Object["spec"] = map[string]interface{}{
		"priority": int64(500),
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &GlobalizationActionExecutor{client: fakeClient}

	action := drv1alpha1.Action{
		Name: "delete-globalization",
		Type: drv1alpha1.ActionTypeGlobalization,
		Globalization: &drv1alpha1.GlobalizationAction{
			Operation: drv1alpha1.OperationDelete,
			Name:      "demo-global-values",
		},
	}

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	glob := &unstructured.Unstructured{}
	glob.SetGroupVersionKind(globalizationGVK)
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-global-values"}, glob)
	if err == nil {
		t.Fatal("expected Globalization to be deleted")
	}
}

func TestGlobalizationExecutor_Patch_UpdatesExistingSpec(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(globalizationGVK)
	existing.SetName("demo-global-values")
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
	ex := &GlobalizationActionExecutor{client: fakeClient}

	action := buildGlobalizationAction(drv1alpha1.OperationPatch)
	action.Rollback = &drv1alpha1.Action{
		Name: "rollback",
		Type: drv1alpha1.ActionTypeGlobalization,
		Globalization: &drv1alpha1.GlobalizationAction{
			Operation: drv1alpha1.OperationDelete,
			Name:      "demo-global-values",
		},
	}
	action.Globalization.Spec.Priority = 700
	action.Globalization.Spec.OverridePolicy = clusternetapps.ApplyNow

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(globalizationGVK)
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-global-values"}, updated); err != nil {
		t.Fatalf("get updated Globalization: %v", err)
	}

	priority, found, err := unstructured.NestedInt64(updated.Object, "spec", "priority")
	if err != nil || !found {
		t.Fatalf("expected spec.priority to exist, found=%v err=%v", found, err)
	}
	if priority != 700 {
		t.Fatalf("expected patched priority=700, got %d", priority)
	}
}
