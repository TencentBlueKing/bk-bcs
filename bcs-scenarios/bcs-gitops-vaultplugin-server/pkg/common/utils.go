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
	"strings"
)

// GetVaultSecretName vault secret密钥存储的name
func GetVaultSecretName(project string) string {
	return fmt.Sprintf(defaultProjectSecretName, project)
}

// GetVaultSecForProAnno vault secret存储在app annotations中的信息,主要是 secret_ns:secret_name 的格式
func GetVaultSecForProAnno(secretNs, project string) string {
	secretName := GetVaultSecretName(project)
	return fmt.Sprintf(VaultSecretPattern, secretNs, secretName)
}

// GetVaultProjectRule vault project rule
func GetVaultProjectRule(project string) string {
	return fmt.Sprintf(vaultPolicyRule, project)
}

// ParseKvPath Parse kvpath into mountpath and secretpath
func ParseKvPath(kvpath string) (mountPath string, secretPath string) {
	tmp := strings.Split(kvpath, "/data/")
	return tmp[0], tmp[1]
}
