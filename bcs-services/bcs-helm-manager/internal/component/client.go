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

// Package component xxxx
package component

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	goReq "github.com/parnurzeal/gorequest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
)

// Request request third api client
func Request(req goReq.SuperAgent, timeout int, proxy string, headers map[string]string) (string, error) {
	client := goReq.New().Timeout(time.Duration(timeout) * time.Second)
	// request by method
	client = client.CustomMethod(req.Method, req.Url)
	// set proxy
	if proxy != "" {
		client = client.Proxy(proxy)
	}
	// set headers
	for key, val := range headers {
		client = client.Set(key, val)
	}
	// request data
	client = client.Send(req.QueryData).Send(req.Data)
	client = client.SetDebug(req.Debug)
	_, body, errs := client.End()

	if len(errs) > 0 {
		blog.Error(
			"request api error, url: %s, method: %s, params: %s, data: %s, err: %v",
			req.Url, req.Method, req.QueryData, req.Data, errs,
		)
		return "", errors.New(stringx.Errs2String(errs))
	}
	return body, nil
}

// GetK8SClientByClusterID 通过集群 ID 获取 k8s client 对象
func GetK8SClientByClusterID(clusterID string) (*kubernetes.Clientset, error) {
	host := fmt.Sprintf("%s/clusters/%s", options.GlobalOptions.Release.APIServer, clusterID)
	config := &rest.Config{
		Host:            host,
		BearerToken:     options.GlobalOptions.Release.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

var (
	auditClient *audit.Client
	auditOnce   sync.Once
)

// GetAuditClient 获取审计客户端
func GetAuditClient() *audit.Client {
	if auditClient == nil {
		auditOnce.Do(func() {
			auditClient =
				audit.NewClient(options.GlobalOptions.Release.APIServer, options.GlobalOptions.Release.Token, nil)
		})
	}
	return auditClient
}

var (
	clientOnce   sync.Once
	globalClient *http.Client
	dialer       = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	// defaultTransport default transport
	defaultTransport http.RoundTripper = &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// NOCC:gas/tls(设计如此)
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
	}
)

// GetHttpClient : 新建Client
func GetHttpClient() *http.Client {
	if globalClient == nil {
		clientOnce.Do(func() {
			globalClient = &http.Client{
				Transport: defaultTransport,
			}
		})
	}
	return globalClient
}

// HttpRequest http 请求
func HttpRequest(ctx context.Context, url, method string, header http.Header, data []byte) ([]byte, error) {
	var req *http.Request
	var err error
	if data != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	resp, err := GetHttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
