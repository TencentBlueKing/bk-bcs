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

// Package argoplugin xxx
package argoplugin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"
)

// VaultArgoPlugin is a struct for working with a Vault backend
type VaultArgoPlugin struct {
	AuthType      types.AuthType
	VaultClient   *api.Client
	KvVersion     string
	secretManager secret.SecretManagerWithVersion
	project       string
}

// NewVaultArgoPluginBackend initializes a new Vault Backend
func NewVaultArgoPluginBackend(
	auth types.AuthType,
	client *api.Client,
	kv string,
	secretManager secret.SecretManagerWithVersion,
	project string,
) *VaultArgoPlugin {
	vault := &VaultArgoPlugin{
		KvVersion:     kv,
		AuthType:      auth,
		VaultClient:   client,
		secretManager: secretManager,
		project:       project,
	}
	return vault
}

// Login authenticates with the auth type provided, it just a fake function
func (v *VaultArgoPlugin) Login() error {
	err := v.AuthType.Authenticate(v.VaultClient)
	if err != nil {
		return err
	}
	return nil
}

// GetSecrets gets secrets from vault and returns the formatted data
func (v *VaultArgoPlugin) GetSecrets(kvpath string, version string,
	annotations map[string]string) (map[string]interface{}, error) {
	_, secretPath := common.ParseKvPath(kvpath)
	if version != "" {
		ver, err := strconv.Atoi(version)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Get secret version [%s] error", version))
		}
		return v.secretManager.GetSecretWithVersion(context.Background(), &secret.SecretRequest{
			Project: v.project,
			Path:    secretPath,
		}, ver)
	}
	return v.secretManager.GetSecret(context.Background(), &secret.SecretRequest{
		Project: v.project,
		Path:    secretPath,
	})
}

// GetIndividualSecret will get the specific secret (placeholder) from the SM backend
// For Vault,
// we only support placeholders replaced from the k/v pairs of a secret which cannot be individually addressed
// So, we use GetSecrets and extract the specific placeholder we want
func (v *VaultArgoPlugin) GetIndividualSecret(kvpath, secret, version string,
	annotations map[string]string) (interface{}, error) {
	data, err := v.GetSecrets(kvpath, version, annotations)
	if err != nil {
		blog.Errorf("GetSecrets failed with error: %v", err)
		return nil, err
	}
	// 忽略path存在,secret不存在
	_, ok := data[secret]
	if !ok {
		return nil, fmt.Errorf("secret %v not found, path: %s", secret, kvpath)
	}
	return data[secret], nil
}
