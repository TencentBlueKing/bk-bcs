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

/*
在M1芯片的mac中使用gomonkey需要注意！！！
https://github.com/agiledragon/gomonkey/issues/70

需要把 $GOMODCACHE/github.com/agiledragon/gomonkey/v2@v2.9.0/modify_binary_darwin.go 文件的
modifyBinary(target uintptr, bytes []byte) 函数

err := syscall.Mprotect(page, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC)
修改为
err := syscall.Mprotect(page, syscall.PROT_READ|syscall.PROT_WRITE)

否则会出现报错:
panic: permission denied [recovered]
        panic: permission denied

*/

package secret

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

// TestInitProject init project
func TestInitProject(t *testing.T) {
	manager := &VaultSecretManager{
		client: &api.Client{},
	}
	patches := gomonkey.NewPatches()

	patches.ApplyMethod(manager.client.Sys(), "PutPolicyWithContext",
		func(_ *api.Sys, _ context.Context, _ string, _ string) error {
			return nil
		})

	patches.ApplyMethodSeq(reflect.TypeOf(manager.client.Auth().Token()), "CreateWithContext", []gomonkey.OutputCell{
		{
			Values: []interface{}{
				&api.Secret{
					Auth: &api.SecretAuth{
						ClientToken: "test-token",
					},
				},
				nil,
			},
		},
	})

	defer patches.Reset()

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
	am := gomonkey.ApplyMethod(manager.client.KVv2(""), "Get",
		func(_ *api.KVv2, _ context.Context, _ string) (*api.KVSecret, error) {
			return mockSecret, fmt.Errorf("errors")
		})
	defer am.Reset()

	req := SecretRequest{}
	secretData, err := manager.GetSecret(context.Background(), &req)

	assert.NoError(t, err)
	assert.Equal(t, len(secretData), len(mockData))
}

// TestGetMetadata test get metadata
func TestGetMetadata(t *testing.T) {
	manager := &VaultSecretManager{client: &api.Client{}}

	mockMetadata := &SecretMetadata{
		CreateTime:     time.Time{},
		UpdatedTime:    time.Time{},
		CurrentVersion: 0,
		Version:        nil,
	}
	patches := gomonkey.ApplyMethod(manager.client.KVv2(""), "GetMetadata",
		func(_ *api.KVv2, _ context.Context, _ string) (*SecretMetadata, error) {
			return mockMetadata, nil
		})
	defer patches.Reset()

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
	patches := gomonkey.ApplyMethod(manager.client.Logical(), "ListWithContext",
		func(_ *api.Logical, _ context.Context, _ string) (*api.Secret, error) {
			return mockPath, nil
		})
	defer patches.Reset()

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
	am := gomonkey.ApplyMethod(manager.client.KVv2(""), "Put",
		func(_ *api.KVv2, _ context.Context, _ string, _ map[string]interface{}) (*api.KVSecret, error) {
			return mockSecret, fmt.Errorf("errors")
		})
	defer am.Reset()

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

	am := gomonkey.ApplyMethod(manager.client.KVv2(""), "DeleteMetadata",
		func(_ *api.KVv2, _ context.Context, _ string) error {
			return nil
		})
	defer am.Reset()

	req := SecretRequest{}
	err := manager.DeleteSecret(context.Background(), &req)

	assert.NoError(t, err)
}
