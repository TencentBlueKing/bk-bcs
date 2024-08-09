/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package secret

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/options"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type MockSecret struct {
	corev1.SecretInterface
}
type MockCoreV1 struct {
	corev1.CoreV1Interface
	secret MockSecret
}

func (c MockCoreV1) Secrets(namespace string) corev1.SecretInterface {
	return c.secret
}
func (c MockSecret) Create(ctx context.Context, secret *v1.Secret, opts metav1.CreateOptions) (*v1.Secret, error) {
	return nil, nil
}

// TestInitProject init project
func TestInitProject(t *testing.T) {
	manager := &VaultSecretManager{
		option:  &options.Options{},
		client:  &api.Client{},
		kclient: &kubernetes.Clientset{},
	}

	mock.Patch(manager.hasInitProject, func(project string) bool {
		return false
	})
	mock.Patch((*api.Sys).Mount, func(_ *api.Sys, path string, mountInfo *api.MountInput) error {
		return nil
	})
	mock.Patch((*api.Sys).PutPolicyWithContext, func(_ *api.Sys, _ context.Context, _ string, _ string) error {
		return nil
	})
	mock.Patch((*api.TokenAuth).CreateWithContext, func(_ *api.TokenAuth, ctx context.Context, opts *api.TokenCreateRequest) (*api.Secret, error) {
		return &api.Secret{
			Auth: &api.SecretAuth{
				ClientToken: "test-token",
			},
		}, nil
	})
	secretIntf := MockSecret{}
	corev1Intf := MockCoreV1{secret: secretIntf}
	mock.Patch(manager.kclient.CoreV1, func() corev1.CoreV1Interface {
		return corev1Intf
	})

	err := manager.InitProject("")

	assert.NoError(t, err)
}

// TestGetSecret test get secret
func TestGetSecret(t *testing.T) {
	manager := &VaultSecretManager{client: &api.Client{}}

	mockData := map[string]interface{}{
		"key": "val",
	}
	mockSecret := &api.KVSecret{
		Data:            mockData,
		VersionMetadata: nil,
		CustomMetadata:  nil,
		Raw:             nil,
	}
	mock.Patch((*api.KVv2).Get, func(_ *api.KVv2, _ context.Context, _ string) (*api.KVSecret, error) {
		return mockSecret, nil
	})

	req := SecretRequest{}
	secretData, err := manager.GetSecret(context.Background(), &req)

	assert.NoError(t, err)
	assert.Equal(t, len(secretData), len(mockData))
}

// TestGetMetadata test get metadata
func TestGetMetadata(t *testing.T) {
	manager := &VaultSecretManager{client: &api.Client{}}

	mockMetadata := &api.KVMetadata{
		// CreateTime:     time.Time{},
		UpdatedTime:    time.Time{},
		CurrentVersion: 123,
		// Version:        nil,
	}

	mock.Patch((*api.KVv2).GetMetadata, func(_ *api.KVv2, _ context.Context, _ string) (*api.KVMetadata, error) {
		return mockMetadata, nil
	})

	req := SecretRequest{}
	metadata, err := manager.GetMetadata(context.Background(), &req)

	assert.NoError(t, err)
	assert.Equal(t, metadata.CurrentVersion, mockMetadata.CurrentVersion)
}

// TestListSecret test list secret
func TestListSecret(t *testing.T) {
	manager := &VaultSecretManager{client: &api.Client{}}

	mockPath := &api.Secret{
		Data: map[string]interface{}{"keys": []interface{}{"hello", "qa70"}},
	}

	mock.Patch((*api.Logical).ListWithContext, func(_ *api.Logical, _ context.Context, _ string) (*api.Secret, error) {
		return mockPath, nil
	})

	req := SecretRequest{}
	path, err := manager.ListSecret(context.Background(), &req)

	assert.NoError(t, err)
	assert.Equal(t, path, []string{"hello", "qa70"})
}

// TestCreatePutSecret test create put secret
func TestCreatePutSecret(t *testing.T) {
	manager := &VaultSecretManager{client: &api.Client{}}

	mockData := map[string]interface{}{
		"key": "val",
	}
	mockSecret := &api.KVSecret{
		Data:            mockData,
		VersionMetadata: nil,
		CustomMetadata:  nil,
		Raw:             nil,
	}
	mock.Mock((*api.KVv2).Put, func(ctx context.Context, fn *core.FuncInfo, args, results core.Object) error {
		results.GetFieldIndex(0).Set(mockSecret)
		return nil
	})

	req := SecretRequest{
		Data: map[string]interface{}{
			"key": "val",
		},
	}
	err := manager.CreateSecret(context.Background(), &req)

	assert.NoError(t, err)
	//assert.Equal(t, len(secretData), len(mockData))
}

// TestDeleteSecret test delete secret.
func TestDeleteSecret(t *testing.T) {
	manager := &VaultSecretManager{client: &api.Client{}}

	mock.Patch((*api.KVv2).DeleteMetadata, func(_ *api.KVv2, _ context.Context, _ string) error {
		return nil
	})

	req := SecretRequest{}
	err := manager.DeleteSecret(context.Background(), &req)

	assert.NoError(t, err)
}
