/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	socketProxyPathFmt    = "/apis/proxies.clusternet.io/v1alpha1/sockets/%s/proxy/direct"
	impersonateUser       = "clusternet"
	tokenHeaderKey        = "Impersonate-Extra-Clusternet-Token"
	impersonateUserHeader = "Impersonate-User"
	childDeployerSecret   = "child-cluster-deployer"
	secretTokenKey        = "token"
	secretLegacyTokenKey  = "child-cluster-token"
)

// ChildClusterClientFactory builds child-cluster clients.
// Default implementation uses Clusternet SocketProxy;
// future implementations may use kubeconfig-based access.
type ChildClusterClientFactory interface {
	GetChildClient(ctx context.Context, clusterID, secretNamespace string) (client.Client, error)
}

// SocketProxyChildClusterClientFactory creates child clients by proxying through the parent apiserver.
type SocketProxyChildClusterClientFactory struct {
	parentClient client.Client
	parentConfig *rest.Config
}

// NewSocketProxyChildClusterClientFactory creates a new SocketProxy-based factory.
// NOCC:tosa/fn_length(设计如此)
func NewSocketProxyChildClusterClientFactory(
	parentClient client.Client,
	parentConfig *rest.Config,
) *SocketProxyChildClusterClientFactory {
	return &SocketProxyChildClusterClientFactory{
		parentClient: parentClient,
		parentConfig: parentConfig,
	}
}

// GetChildClient builds a controller-runtime client for a child cluster via SocketProxy.
func (f *SocketProxyChildClusterClientFactory) GetChildClient(
	ctx context.Context,
	clusterID,
	secretNamespace string,
) (client.Client, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("clusterID is required")
	}
	if secretNamespace == "" {
		return nil, fmt.Errorf("secret namespace is required")
	}

	token, err := f.getChildClusterToken(ctx, secretNamespace)
	if err != nil {
		return nil, err
	}

	cfg := rest.CopyConfig(f.parentConfig)
	cfg.Host = strings.TrimRight(f.parentConfig.Host, "/") +
		fmt.Sprintf(socketProxyPathFmt, clusterID)

	prevWrap := cfg.WrapTransport
	cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		if prevWrap != nil {
			rt = prevWrap(rt)
		}
		return &socketProxyAuthRoundTripper{base: rt, token: token}
	}

	childClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("create child client for cluster %s: %w", clusterID, err)
	}
	return childClient, nil
}

func (f *SocketProxyChildClusterClientFactory) getChildClusterToken(
	ctx context.Context,
	secretNamespace string,
) (string, error) {
	secret := &corev1.Secret{}
	if err := f.parentClient.Get(ctx, client.ObjectKey{
		Namespace: secretNamespace,
		Name:      childDeployerSecret,
	}, secret); err != nil {
		return "", fmt.Errorf("get secret %s/%s: %w", secretNamespace, childDeployerSecret, err)
	}

	if token, ok := secret.Data[secretTokenKey]; ok && len(token) > 0 {
		return string(token), nil
	}
	if token, ok := secret.Data[secretLegacyTokenKey]; ok && len(token) > 0 {
		return string(token), nil
	}
	return "", fmt.Errorf("secret %s/%s has no token key", secretNamespace, childDeployerSecret)
}

type socketProxyAuthRoundTripper struct {
	base  http.RoundTripper
	token string
}

func (r *socketProxyAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	clone.Header.Set(impersonateUserHeader, impersonateUser)
	clone.Header.Set(tokenHeaderKey, r.token)
	return r.base.RoundTrip(clone) //nolint:bodyclose // caller responsibility
}
