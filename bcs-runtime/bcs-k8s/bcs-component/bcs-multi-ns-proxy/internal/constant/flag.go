/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constant

const (
	// FlagKeyProxyPort listening port for proxy server
	FlagKeyProxyPort = "proxy-port"
	// FlagKeyProxyAddress listening address for proxy server
	FlagKeyProxyAddress = "proxy-address"
	// FlagKeyProxyServerCert cert file path for proxy server
	FlagKeyProxyServerCert = "proxy-servercert"
	// FlagKeyProxyServerKey key file path for proxy server
	FlagKeyProxyServerKey = "proxy-serverkey"

	// FlagKeyKubeconfigMode mode for proxy to get all kubeconfigs, available [secret, file]
	FlagKeyKubeconfigMode = "kubeconfig-mode"
	// FlagKeyKubeconfigSecretName k8s secret name for proxy to get all kubeconfigs when use secret mode
	FlagKeyKubeconfigSecretName = "kubeconfig-secretname" // nolint
	// FlagKeyKubeconfigSecretNamespace k8s secret namespace for proxy to get all kubeconfigs when use secret mode
	FlagKeyKubeconfigSecretNamespace = "kubeconfig-secretnamespace"
	// FlagKeyKubeconfigDir is the directory which holds all kubeconfigs for different namespaces
	FlagKeyKubeconfigDir = "kubeconfig-dir"
	// FlagKeyKubeconfigDefaultNs is the default namespace to use for non-namespaced api resource
	FlagKeyKubeconfigDefaultNs = "kubeconfig-defaultns"
	// FlagKeyKubeconfigCheckDuration interval for checking kubeconfig directory
	FlagKeyKubeconfigCheckDuration = "kubeconfig-checkduration"
	// FlagKeyConfigPath is config file path
	FlagKeyConfigPath = "config-path"
	// FlagKeyConfigName is config file name
	FlagKeyConfigName = "config-name"

	// KubeconfigModeFile find kubeconfig by file
	KubeconfigModeFile = "file"
	// KubeconfigModeSecret find kubeconfig by secret
	KubeconfigModeSecret = "secret"
)
