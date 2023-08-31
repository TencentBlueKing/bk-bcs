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

package common

const (
	// vaultPolicyRule vault policy rule for token for every project
	vaultPolicyRule = `
	path "%s/*" {
	 capabilities = ["create", "read", "update", "patch", "delete", "list"]
	}
	`

	//VaultVersion vault engine version, for secret version management and control.
	VaultVersion = "2"

	// default secret plugin type
	secretTypeEnvKey  = "secretType"
	defaultSecretType = "vault"

	// default secret plugin addr
	secretEndpointEnvKey = "secretEndpoints"
	// NOCC:gas/crypto(工具误报)
	defaultSecretEndpoint = "https://bcs-gitops-vault.default.svc.cluster.local:8200"

	// default secret car dir
	secretCaDirEnvKey  = "vaultCacert"
	defaultSecretCaDir = "/data/bcs/certs/vaultca"

	// default vault secret namespace
	vaultSecretNamespaceEnvKey  = "vaultSecretNamespace"
	defaultVaultSecretNamespace = "default"

	defaultProjectSecretName = "vault-secret-%s"
	// VaultSecretPattern vault token info for project, mounts, token, ca
	VaultSecretPattern = "%s:%s"

	vaultTokenForServerEnvKey = "secretToken"

	// GitopsServiceEnvKey gitops service name environment
	GitopsServiceEnvKey = "GITOPS_SERVICE"
	// GitopsUser gitops service user name environment
	GitopsUser = "GITOPS_USER"
	// GitopsPassword gitops service password environment
	GitopsPassword = "GITOPS_PASSWORD"
	// GitopsAdminNamespace gitops service namespace environment
	GitopsAdminNamespace = "GITOPS_ADMIN_NAMESPACE"
)
