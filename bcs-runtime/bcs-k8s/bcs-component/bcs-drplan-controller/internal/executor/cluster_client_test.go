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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NOCC:tosa/fn_length(设计如此)
// NOCC:tosa/fn_length(设计如此)
func TestSocketProxyChildClusterClientFactory_GetChildClient(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      childDeployerSecret,
			Namespace: "clusternet-abc",
		},
		Data: map[string][]byte{
			secretTokenKey: []byte("test-token-123"),
		},
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
	parentCfg := &rest.Config{Host: "https://10.0.0.1:6443"}

	factory := NewSocketProxyChildClusterClientFactory(fakeClient, parentCfg)

	t.Run("successful client creation", func(t *testing.T) {
		childClient, err := factory.GetChildClient(context.Background(), "test-cluster-id", "clusternet-abc")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if childClient == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("empty clusterID returns error", func(t *testing.T) {
		_, err := factory.GetChildClient(context.Background(), "", "clusternet-abc")
		if err == nil {
			t.Fatal("expected error for empty clusterID")
		}
	})

	t.Run("empty secretNamespace returns error", func(t *testing.T) {
		_, err := factory.GetChildClient(context.Background(), "test-id", "")
		if err == nil {
			t.Fatal("expected error for empty secretNamespace")
		}
	})

	t.Run("missing secret returns error", func(t *testing.T) {
		_, err := factory.GetChildClient(context.Background(), "test-id", "nonexistent-ns")
		if err == nil {
			t.Fatal("expected error for missing secret")
		}
	})

	t.Run("secret without token key returns error", func(t *testing.T) {
		emptySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      childDeployerSecret,
				Namespace: "empty-ns",
			},
			Data: map[string][]byte{
				"other-key": []byte("value"),
			},
		}
		emptyClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(emptySecret).Build()
		emptyFactory := NewSocketProxyChildClusterClientFactory(emptyClient, parentCfg)
		_, err := emptyFactory.GetChildClient(context.Background(), "test-id", "empty-ns")
		if err == nil {
			t.Fatal("expected error for secret without token key")
		}
	})

	t.Run("legacy token key works", func(t *testing.T) {
		legacySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      childDeployerSecret,
				Namespace: "legacy-ns",
			},
			Data: map[string][]byte{
				secretLegacyTokenKey: []byte("legacy-token"),
			},
		}
		legacyClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(legacySecret).Build()
		legacyFactory := NewSocketProxyChildClusterClientFactory(legacyClient, parentCfg)
		client, err := legacyFactory.GetChildClient(context.Background(), "test-id", "legacy-ns")
		if err != nil {
			t.Fatalf("expected no error with legacy token, got: %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client with legacy token")
		}
	})
}
