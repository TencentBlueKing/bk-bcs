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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

func TestKubernetesResourceRollbackApplyDefaultsToDelete(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 scheme: %v", err)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
		},
	}
	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(configMap).Build()
	executor := NewKubernetesResourceActionExecutor(k8sClient)

	action := &drv1alpha1.Action{
		Name: "apply-configmap",
		Type: drv1alpha1.ActionTypeKubernetesResource,
		Resource: &drv1alpha1.KubernetesResourceAction{
			Operation: drv1alpha1.OperationApply,
		},
	}
	actionStatus := &drv1alpha1.ActionStatus{
		Name:  "apply-configmap",
		Phase: drv1alpha1.PhaseSucceeded,
		Outputs: &drv1alpha1.ActionOutputs{
			ResourceRef: &corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "ConfigMap",
				Namespace:  "default",
				Name:       "demo",
			},
		},
	}

	rollbackStatus, err := executor.Rollback(context.Background(), action, actionStatus, nil)
	if err != nil {
		t.Fatalf("expected rollback to succeed, got %v", err)
	}
	if rollbackStatus.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("expected rollback phase Succeeded, got %s", rollbackStatus.Phase)
	}

	got := &corev1.ConfigMap{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: "demo", Namespace: "default"}, got)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected ConfigMap to be deleted, got err=%v", err)
	}
}
