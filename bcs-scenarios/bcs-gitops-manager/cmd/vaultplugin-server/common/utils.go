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

import (
	"fmt"
	"os"
	"strings"
)

// GetSecretType secret type
func GetSecretType() string {
	if addr := os.Getenv(secretTypeEnvKey); len(addr) > 0 {
		return addr
	}
	return defaultSecretType
}

// GetVaultAddr vault地址，默认是通过k8s内部svc访问
func GetVaultAddr() string {
	if addr := os.Getenv(secretEndpointEnvKey); len(addr) > 0 {
		return addr
	}
	return defaultSecretEndpoint
}

// GetVaultCa vault的私有ca地址，部署的vault使用的是私有证书，客户端链接需要指定私有证书路径才能正常访问
// 这里需要注意的是有两个地方的引用，一个是bcs-gitops-manager本身调用vault api需要的ca，另一个是写到secret给avp使用的ca
func GetVaultCa() string {
	if ca := os.Getenv(secretCaDirEnvKey); len(ca) > 0 {
		return ca
	}
	return defaultSecretCaDir
}

// GetVaultSecretNamespace vault secret密钥存储的ns
func GetVaultSecretNamespace() string {
	if ns := os.Getenv(vaultSecretNamespaceEnvKey); len(ns) > 0 {
		return ns
	}
	return defaultVaultSecretNamespace
}

// GetVaultSecretName vault secret密钥存储的name
func GetVaultSecretName(project string) string {
	return fmt.Sprintf(defaultProjectSecretName, project)
}

// GetVaultSecForProAnno vault secret存储在app annotations中的信息,主要是 secret_ns:secret_name 的格式
func GetVaultSecForProAnno(project string) string {
	secretNs := GetVaultSecretNamespace()
	secretName := GetVaultSecretName(project)
	return fmt.Sprintf(VaultSecretPattern, secretNs, secretName)
}

// GetVaultTokenForServer vault token
func GetVaultTokenForServer() string {
	if token := os.Getenv(vaultTokenForServerEnvKey); len(token) > 0 {
		return token
	}
	return ""
}

// GetVaultProjectRule vault project rule
func GetVaultProjectRule(project string) string {
	return fmt.Sprintf(vaultPolicyRule, project)
}

// GetGitopsService get gitops service name from environment variable
func GetGitopsService() string {
	if service := os.Getenv(GitopsServiceEnvKey); len(service) > 0 {
		return service
	}
	return ""
}

// GetGitopsAdminNamespace get gitops service's namespace from environment variable
func GetGitopsAdminNamespace() string {
	if ns := os.Getenv(GitopsAdminNamespace); len(ns) > 0 {
		return ns
	}
	return ""
}

// GetGitopsUser get gitops user from environment variable
func GetGitopsUser() string {
	if user := os.Getenv(GitopsUser); len(user) > 0 {
		return user
	}
	return ""
}

// GetGitopsPassword get gitops password from environment variable
func GetGitopsPassword() string {
	if pwd := os.Getenv(GitopsPassword); len(pwd) > 0 {
		return pwd
	}
	return ""
}

// ParseKvPath Parse kvpath into mountpath and secretpath
func ParseKvPath(kvpath string) (mountPath string, secretPath string) {
	tmp := strings.Split(kvpath, "/data/")
	return tmp[0], tmp[1]
}
