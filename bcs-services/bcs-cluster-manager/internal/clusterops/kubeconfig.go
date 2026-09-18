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

package clusterops

import (
	"errors"
	"fmt"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	// ErrExecCredentialForbidden kubeConfig 声明 exec 凭证插件。client-go 会在发起请求前 fork 子进程
	// 执行 exec.command 获取凭证，外部传入的 kubeConfig 一旦带该字段即等价于在本机执行任意命令。
	ErrExecCredentialForbidden = errors.New("exec credential plugin is forbidden in kubeConfig")

	// ErrAuthProviderForbidden kubeConfig 声明 auth-provider。部分 provider（如 gcp）会执行
	// cmd-path 指定的本地命令，且平台侧的 provider 凭证不应由用户输入指定。
	ErrAuthProviderForbidden = errors.New("auth-provider is forbidden in kubeConfig")

	// ErrLocalFileRefForbidden kubeConfig 通过文件路径引用凭证（tokenFile/client-certificate/
	// client-key/certificate-authority）。这些路径在服务端解析，会读取本服务所在机器的本地文件，
	// 例如 /var/run/secrets/kubernetes.io/serviceaccount/token。凭证必须以内联 data 形式提供。
	ErrLocalFileRefForbidden = errors.New("referencing local file for credential is forbidden in kubeConfig")

	// ErrUnsupportedAuthType kubeConfig 未提供任何受支持的认证方式
	ErrUnsupportedAuthType = errors.New("no supported authentication type found in kubeConfig")

	// ErrProxyURLForbidden kubeConfig 声明 proxy-url。该字段会让本服务经由用户指定的代理出网，
	// 平台侧访问用户集群不应由用户输入决定出口。
	ErrProxyURLForbidden = errors.New("proxy-url is forbidden in kubeConfig")

	// ErrImpersonateForbidden kubeConfig 声明 act-as 系列模拟身份字段。导入集群的凭证不需要模拟身份，
	// 该字段只会让目标集群侧的审计记录与实际调用方不一致。
	ErrImpersonateForbidden = errors.New("impersonation is forbidden in kubeConfig")
)

// ValidateKubeConfig 校验外部传入的 kubeConfig。
//
// user 段仅放通白名单内的三类认证方式：
//  1. 客户端证书认证：client-certificate-data + client-key-data
//  2. Bearer Token 认证：token（不支持 tokenFile）
//  3. Basic Auth 认证：username + password
//
// exec、auth-provider、act-as 模拟身份，以及一切通过本地文件路径引用凭证的写法均拒绝。
// cluster 段拒绝 certificate-authority 文件路径与 proxy-url。
// 所有由外部输入（接口请求、DB 存量数据、任务参数）构造 k8s 客户端的路径都必须先经过该校验。
func ValidateKubeConfig(data []byte) error {
	apiConfig, err := clientcmd.Load(data)
	if err != nil {
		return fmt.Errorf("load kubeConfig failed: %w", err)
	}

	if len(apiConfig.AuthInfos) == 0 {
		return fmt.Errorf("%w: kubeConfig declares no user", ErrUnsupportedAuthType)
	}

	// 遍历前排序，保证多个条目同时非法时返回的错误信息稳定
	authNames := make([]string, 0, len(apiConfig.AuthInfos))
	for name := range apiConfig.AuthInfos {
		authNames = append(authNames, name)
	}
	sort.Strings(authNames)
	for _, name := range authNames {
		if err := validateAuthInfo(apiConfig.AuthInfos[name]); err != nil {
			return fmt.Errorf("kubeConfig user %q is invalid: %w", name, err)
		}
	}

	clusterNames := make([]string, 0, len(apiConfig.Clusters))
	for name := range apiConfig.Clusters {
		clusterNames = append(clusterNames, name)
	}
	sort.Strings(clusterNames)
	for _, name := range clusterNames {
		if err := validateClusterInfo(apiConfig.Clusters[name]); err != nil {
			return fmt.Errorf("kubeConfig cluster %q is invalid: %w", name, err)
		}
	}

	return nil
}

// validateClusterInfo 对单个 cluster 条目做校验
func validateClusterInfo(cluster *clientcmdapi.Cluster) error {
	if cluster == nil {
		return nil
	}

	if cluster.CertificateAuthority != "" {
		return fmt.Errorf("%w: use certificate-authority-data instead", ErrLocalFileRefForbidden)
	}
	if cluster.ProxyURL != "" {
		return ErrProxyURLForbidden
	}

	return nil
}

// validateAuthInfo 对单个 user 条目做白名单校验
func validateAuthInfo(authInfo *clientcmdapi.AuthInfo) error {
	if authInfo == nil {
		return fmt.Errorf("%w: empty user", ErrUnsupportedAuthType)
	}

	// 先拒绝高危写法，再判断是否命中白名单，避免 “带 token 同时带 exec” 绕过
	if err := rejectForbiddenAuthFields(authInfo); err != nil {
		return err
	}

	return checkSupportedAuthType(authInfo)
}

// rejectForbiddenAuthFields 拒绝会拉起本地进程、读取本机文件或伪造身份的认证写法
func rejectForbiddenAuthFields(authInfo *clientcmdapi.AuthInfo) error {
	if authInfo.Exec != nil {
		return ErrExecCredentialForbidden
	}
	if authInfo.AuthProvider != nil {
		return ErrAuthProviderForbidden
	}
	if authInfo.TokenFile != "" {
		return fmt.Errorf("%w: tokenFile is not allowed, use token instead", ErrLocalFileRefForbidden)
	}
	if authInfo.ClientCertificate != "" || authInfo.ClientKey != "" {
		return fmt.Errorf("%w: use client-certificate-data and client-key-data instead",
			ErrLocalFileRefForbidden)
	}
	if hasImpersonation(authInfo) {
		return ErrImpersonateForbidden
	}

	return nil
}

// hasImpersonation 判断 user 条目是否声明了 act-as 系列模拟身份字段
func hasImpersonation(authInfo *clientcmdapi.AuthInfo) bool {
	return authInfo.Impersonate != "" || authInfo.ImpersonateUID != "" ||
		len(authInfo.ImpersonateGroups) > 0 || len(authInfo.ImpersonateUserExtra) > 0
}

// checkSupportedAuthType 判断是否命中白名单内的三类认证方式，其中成对字段必须同时提供
func checkSupportedAuthType(authInfo *clientcmdapi.AuthInfo) error {
	hasClientCert := len(authInfo.ClientCertificateData) > 0 || len(authInfo.ClientKeyData) > 0
	if hasClientCert && (len(authInfo.ClientCertificateData) == 0 || len(authInfo.ClientKeyData) == 0) {
		return errors.New("client-certificate-data and client-key-data must be provided together")
	}

	hasBasicAuth := authInfo.Username != "" || authInfo.Password != ""
	if hasBasicAuth && (authInfo.Username == "" || authInfo.Password == "") {
		return errors.New("username and password must be provided together")
	}

	if !hasClientCert && !hasBasicAuth && authInfo.Token == "" {
		return fmt.Errorf("%w: expect client-certificate-data/client-key-data, token or username/password",
			ErrUnsupportedAuthType)
	}

	return nil
}
