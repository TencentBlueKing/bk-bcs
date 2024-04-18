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

// Package restyclient for http client
package restyclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	timeout = time.Second * 30
)

var (
	clientOnce   sync.Once
	globalClient *resty.Client
)

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}

// defaultTransport default transport
var defaultTransport http.RoundTripper = &http.Transport{
	DialContext:           dialer.DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	// NOCC:gas/tls(设计如此)
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
}

// New : 新建 Client, 设置公共参数，tracing 等; 每次新建，cookies不复用
func New() *resty.Client {
	if globalClient == nil {
		clientOnce.Do(func() {
			globalClient = resty.New().
				SetTransport(otelhttp.NewTransport(defaultTransport)).
				SetTimeout(timeout).
				SetCookieJar(nil).
				SetDebugBodyLimit(1024).
				SetHeader("User-Agent", "bcs-restyclient")
		})
	}
	return globalClient
}

// R : New().R() 快捷方式, 已设置公共参数，tracing 等
func R() *resty.Request {
	return New().R()
}
