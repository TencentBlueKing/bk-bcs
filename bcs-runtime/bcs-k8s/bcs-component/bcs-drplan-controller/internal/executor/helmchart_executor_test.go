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

var helmChartGVK = schema.GroupVersionKind{
	Group:   "apps.clusternet.io",
	Version: "v1alpha1",
	Kind:    "HelmChart",
}

func buildHelmChartAction(operation string) drv1alpha1.Action {
	return drv1alpha1.Action{
		Name: "test-helmchart",
		Type: drv1alpha1.ActionTypeHelmChart,
		HelmChart: &drv1alpha1.HelmChartAction{
			Operation: operation,
			Name:      "demo-chart",
			Namespace: "default",
			Spec: &clusternetapps.HelmChartSpec{
				HelmOptions: clusternetapps.HelmOptions{
					Repository:   "oci://registry.example.com/charts",
					Chart:        "demo-app",
					ChartVersion: "1.2.3",
				},
				TargetNamespace: "default",
			},
		},
	}
}

func TestHelmChartExecutor_Create(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &HelmChartActionExecutor{client: fakeClient}

	action := buildHelmChartAction("")
	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if status.Outputs == nil || status.Outputs.HelmChartRef == nil {
		t.Fatalf("expected HelmChartRef to be recorded, got %#v", status.Outputs)
	}

	chart := &unstructured.Unstructured{}
	chart.SetGroupVersionKind(helmChartGVK)
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-chart", Namespace: "default"}, chart); err != nil {
		t.Fatalf("expected HelmChart to be created: %v", err)
	}
}

func TestHelmChartExecutor_Create_PreservesPlainHTTP(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &HelmChartActionExecutor{client: fakeClient}

	action := buildHelmChartAction("")
	action.HelmChart.Spec.PlainHTTP = boolPtr(true)

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	chart := &unstructured.Unstructured{}
	chart.SetGroupVersionKind(helmChartGVK)
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-chart", Namespace: "default"}, chart); err != nil {
		t.Fatalf("expected HelmChart to be created: %v", err)
	}

	got, found, err := unstructured.NestedBool(chart.Object, "spec", "plainHTTP")
	if err != nil {
		t.Fatalf("read spec.plainHTTP: %v", err)
	}
	if !found {
		t.Fatal("expected spec.plainHTTP to be rendered")
	}
	if !got {
		t.Fatalf("spec.plainHTTP = %v, want true", got)
	}
}

func TestHelmChartExecutor_Apply_UsesPatch(t *testing.T) {
	sc := newTestScheme()
	recorder := &patchRecordingClient{
		Client: fakeclient.NewClientBuilder().WithScheme(sc).Build(),
	}
	ex := &HelmChartActionExecutor{client: recorder}

	action := buildHelmChartAction(drv1alpha1.OperationApply)
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

func TestHelmChartExecutor_Delete_DoesNotRequireSpec(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(helmChartGVK)
	existing.SetName("demo-chart")
	existing.SetNamespace("default")
	existing.Object["spec"] = map[string]interface{}{
		"repo":            "oci://registry.example.com/charts",
		"chart":           "demo-app",
		"version":         "1.2.3",
		"targetNamespace": "default",
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &HelmChartActionExecutor{client: fakeClient}

	action := drv1alpha1.Action{
		Name: "delete-helmchart",
		Type: drv1alpha1.ActionTypeHelmChart,
		HelmChart: &drv1alpha1.HelmChartAction{
			Operation: drv1alpha1.OperationDelete,
			Name:      "demo-chart",
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

	chart := &unstructured.Unstructured{}
	chart.SetGroupVersionKind(helmChartGVK)
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-chart", Namespace: "default"}, chart)
	if err == nil {
		t.Fatal("expected HelmChart to be deleted")
	}
}

func TestHelmChartExecutor_Patch_UpdatesExistingSpec(t *testing.T) {
	sc := newTestScheme()

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(helmChartGVK)
	existing.SetName("demo-chart")
	existing.SetNamespace("default")
	existing.Object["spec"] = map[string]interface{}{
		"repo":            "oci://registry.example.com/charts",
		"chart":           "demo-app",
		"version":         "1.2.3",
		"targetNamespace": "default",
		"wait":            false,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).WithObjects(existing).Build()
	ex := &HelmChartActionExecutor{client: fakeClient}

	action := buildHelmChartAction(drv1alpha1.OperationPatch)
	action.HelmChart.Spec = &clusternetapps.HelmChartSpec{
		HelmOptions: clusternetapps.HelmOptions{
			ChartVersion: "2.0.0",
			Wait:         boolPtr(true),
		},
	}

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(helmChartGVK)
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-chart", Namespace: "default"}, updated); err != nil {
		t.Fatalf("get updated HelmChart: %v", err)
	}

	version, found, err := unstructured.NestedString(updated.Object, "spec", "version")
	if err != nil || !found {
		t.Fatalf("expected spec.version to exist, found=%v err=%v", found, err)
	}
	if version != "2.0.0" {
		t.Fatalf("expected patched version=2.0.0, got %q", version)
	}

	wait, found, err := unstructured.NestedBool(updated.Object, "spec", "wait")
	if err != nil || !found {
		t.Fatalf("expected spec.wait to exist, found=%v err=%v", found, err)
	}
	if !wait {
		t.Fatalf("expected patched wait=true, got %v", wait)
	}
}

func TestHelmChartExecutor_Rollback_CreateDeletesResource(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &HelmChartActionExecutor{client: fakeClient}

	action := buildHelmChartAction(drv1alpha1.OperationCreate)
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

	chart := &unstructured.Unstructured{}
	chart.SetGroupVersionKind(helmChartGVK)
	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "demo-chart", Namespace: "default"}, chart)
	if err == nil {
		t.Fatal("expected HelmChart to be deleted by rollback")
	}
}

func boolPtr(v bool) *bool {
	return &v
}
