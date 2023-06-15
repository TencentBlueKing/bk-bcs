/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constants

const (
	// SecretTypeOpaque 普通类型
	SecretTypeOpaque = "Opaque"

	// SecretTypeDocker 镜像配置信息
	SecretTypeDocker = "kubernetes.io/dockerconfigjson" // NOCC:gas/crypto(误报)

	// SecretTypeBasicAuth 基础认证信息
	SecretTypeBasicAuth = "kubernetes.io/basic-auth"

	// SecretTypeSSHAuth SSH 身份认证
	SecretTypeSSHAuth = "kubernetes.io/ssh-auth"

	// SecretTypeTLS TLS 认证
	SecretTypeTLS = "kubernetes.io/tls"

	// SecretTypeSAToken ServiceAccount Token
	SecretTypeSAToken = "kubernetes.io/service-account-token"
)

// IngTLSCertEnabledSecretTypes 可用于 ingress tls 证书的 Secret 类型
var IngTLSCertEnabledSecretTypes = []string{SecretTypeOpaque, SecretTypeTLS}
