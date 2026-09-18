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
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// fakeCert 仅用于填充 client-certificate-data / client-key-data，不需要是合法 PEM，
// 因为白名单校验只判断字段是否存在
const fakeCert = "ZmFrZS1jZXJ0" // base64("fake-cert")

// kubeConfigWithUser 以指定的 user 段拼出完整 kubeConfig
func kubeConfigWithUser(user string) string {
	return kubeConfigWith("    server: https://10.0.0.1:6443\n    insecure-skip-tls-verify: true", user)
}

// kubeConfigWith 以指定的 cluster 段和 user 段拼出完整 kubeConfig
func kubeConfigWith(cluster, user string) string {
	return fmt.Sprintf(`apiVersion: v1
 kind: Config
 clusters:
 - name: c
   cluster:
 %s
 users:
 - name: u
   user:
 %s
 contexts:
 - name: ctx
   context:
	 cluster: c
	 user: u
 current-context: ctx
 `, cluster, user)
}

func TestKubeConfigAllowAuthTypes(t *testing.T) {
	cases := []struct {
		name string
		user string
	}{
		{
			name: "client certificate",
			user: "    client-certificate-data: " + fakeCert + "\n    client-key-data: " + fakeCert,
		},
		{
			name: "bearer token",
			user: "    token: fake-token",
		},
		{
			name: "basic auth",
			user: "    username: admin\n    password: fake-password",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := ValidateKubeConfig([]byte(kubeConfigWithUser(c.user))); err != nil {
				t.Fatalf("expect %s to be allowed, got %v", c.name, err)
			}
		})
	}
}

func TestKubeConfigRejectAuthTypes(t *testing.T) {
	cases := []struct {
		name    string
		user    string
		wantErr error
	}{
		{
			name: "exec credential plugin",
			user: "    exec:\n      apiVersion: client.authentication.k8s.io/v1beta1\n" +
				"      command: /bin/sh\n      args: [\"-c\", \"id\"]\n      interactiveMode: Never",
			wantErr: ErrExecCredentialForbidden,
		},
		{
			name:    "auth provider",
			user:    "    auth-provider:\n      name: gcp",
			wantErr: ErrAuthProviderForbidden,
		},
		{
			name:    "service account tokenFile",
			user:    "    tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token",
			wantErr: ErrLocalFileRefForbidden,
		},
		{
			name:    "client certificate file path",
			user:    "    client-certificate: /etc/kubernetes/pki/admin.crt\n    client-key: /etc/kubernetes/pki/admin.key",
			wantErr: ErrLocalFileRefForbidden,
		},
		{
			name:    "no credential at all",
			user:    "",
			wantErr: ErrUnsupportedAuthType,
		},
		{
			name:    "exec combined with token does not bypass",
			user:    "    token: fake-token\n    exec:\n      command: /bin/sh\n      interactiveMode: Never",
			wantErr: ErrExecCredentialForbidden,
		},
		{
			name:    "impersonation",
			user:    "    token: fake-token\n    as: system:admin",
			wantErr: ErrImpersonateForbidden,
		},
		{
			name:    "impersonation groups",
			user:    "    token: fake-token\n    as-groups: [\"system:masters\"]",
			wantErr: ErrImpersonateForbidden,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateKubeConfig([]byte(kubeConfigWithUser(c.user)))
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expect %v, got %v", c.wantErr, err)
			}
		})
	}
}

func TestKubeConfigRejectClusterFields(t *testing.T) {
	const token = "    token: fake-token"

	cases := []struct {
		name    string
		cluster string
		wantErr error
	}{
		{
			name:    "certificate authority file path",
			cluster: "    server: https://10.0.0.1:6443\n    certificate-authority: /etc/kubernetes/pki/ca.crt",
			wantErr: ErrLocalFileRefForbidden,
		},
		{
			name:    "proxy url",
			cluster: "    server: https://10.0.0.1:6443\n    proxy-url: socks5://attacker.example.com:1080",
			wantErr: ErrProxyURLForbidden,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateKubeConfig([]byte(kubeConfigWith(c.cluster, token)))
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expect %v, got %v", c.wantErr, err)
			}
		})
	}
}

// TestKubeConfigCredentialPairing 证书与 basic auth 必须成对提供，
// 否则应给出明确错误而不是落到「无可用认证方式」
func TestKubeConfigCredentialPairing(t *testing.T) {
	cases := []struct {
		name string
		user string
	}{
		{name: "cert without key", user: "    client-certificate-data: " + fakeCert},
		{name: "key without cert", user: "    client-key-data: " + fakeCert},
		{name: "username without password", user: "    username: admin"},
		{name: "password without username", user: "    password: fake-password"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateKubeConfig([]byte(kubeConfigWithUser(c.user)))
			if err == nil {
				t.Fatal("expect pairing error, got nil")
			}
			if errors.Is(err, ErrUnsupportedAuthType) {
				t.Fatalf("expect explicit pairing error, got %v", err)
			}
		})
	}
}

func TestRESTConfigFromKubeConfig(t *testing.T) {
	config, err := RESTConfigFromKubeConfig([]byte(kubeConfigWithUser("    token: fake-token")))
	if err != nil {
		t.Fatalf("token kubeConfig should pass, got %v", err)
	}
	if config.ExecProvider != nil || config.AuthProvider != nil {
		t.Fatal("rest config should not carry exec or auth provider")
	}
	if config.BearerToken != "fake-token" {
		t.Fatalf("expect bearer token to be preserved, got %q", config.BearerToken)
	}
}

// TestNewKubeClientNoLocalExec 复现漏洞利用流程：构造客户端并发起请求，
// client-go 在首个请求发出前执行 exec.command，这里验证本机不会留下任何执行副作用
func TestNewKubeClientNoLocalExec(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "rce-marker.txt")
	user := fmt.Sprintf("    exec:\n      apiVersion: client.authentication.k8s.io/v1beta1\n"+
		"      command: /bin/sh\n      args: [\"-c\", \"id > %s\"]\n      interactiveMode: Never", marker)

	cli, err := NewKubeClient(base64.StdEncoding.EncodeToString([]byte(kubeConfigWithUser(user))))
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, _ = cli.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		t.Error("NewKubeClient should reject kubeConfig with exec credential plugin")
	} else if !errors.Is(err, ErrExecCredentialForbidden) {
		t.Fatalf("expect ErrExecCredentialForbidden, got %v", err)
	}

	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("exec credential plugin was executed, marker file %s created", marker)
	}
}

// TestNewKubeClientRestRejectExec 验证 exec 全局开关对直接传入 rest.Config 同样生效。
// 这里绕过 RESTConfigFromKubeConfig，直接用 client-go 原生函数构造，模拟未经校验的调用方。
func TestNewKubeClientRestRejectExec(t *testing.T) {
	user := "    exec:\n      apiVersion: client.authentication.k8s.io/v1beta1\n" +
		"      command: /bin/sh\n      interactiveMode: Never"
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfigWithUser(user)))
	if err != nil {
		t.Fatalf("build rest config failed: %v", err)
	}
	if config.ExecProvider == nil {
		t.Fatal("precondition failed: rest config should carry exec provider")
	}

	if _, err := NewKubeClientByRestConfig(config); !errors.Is(err, ErrExecCredentialForbidden) {
		t.Fatalf("expect ErrExecCredentialForbidden, got %v", err)
	}
}
