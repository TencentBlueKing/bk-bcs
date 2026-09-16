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

package iamv4

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"sync"

	restful "github.com/emicklei/go-restful/v3"
)

const basicUser = "bk_iam"

var (
	// ErrNotConfigured 未配置 V4 客户端
	ErrNotConfigured = errors.New("iam v4 not configured")

	// FetchToken 获取 V4 回调 token，由 auth 包注入，禁止写入日志
	FetchToken = func(ctx context.Context) (string, error) {
		return "", ErrNotConfigured
	}

	tokenMu     sync.RWMutex
	cachedToken string
)

// AuthFunc IAM V4 provider 回调鉴权
func AuthFunc(rb *restful.RouteBuilder) *restful.RouteBuilder {
	rb.Filter(Authenticate)
	return rb
}

// Authenticate 校验 Basic Auth：用户名 bk_iam，密码为 V4 auth-token
func Authenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	user, password, ok := request.Request.BasicAuth()
	if !ok {
		writeError(response, http.StatusUnauthorized, "UNAUTHORIZED", "missing basic auth")
		return
	}
	if subtle.ConstantTimeCompare([]byte(user), []byte(basicUser)) != 1 {
		writeError(response, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
		return
	}

	token, err := cachedFetchToken(request.Request.Context())
	if err != nil {
		if errors.Is(err, ErrNotConfigured) {
			writeError(response, http.StatusUnauthorized, "UNAUTHORIZED", "iam v4 not configured")
			return
		}
		writeError(response, http.StatusUnauthorized, "UNAUTHORIZED", "get token from iam v4 failed")
		return
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(password)) != 1 {
		writeError(response, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
		return
	}
	chain.ProcessFilter(request, response)
}

func cachedFetchToken(ctx context.Context) (string, error) {
	tokenMu.RLock()
	if cachedToken != "" {
		token := cachedToken
		tokenMu.RUnlock()
		return token, nil
	}
	tokenMu.RUnlock()

	tokenMu.Lock()
	defer tokenMu.Unlock()
	if cachedToken != "" {
		return cachedToken, nil
	}
	token, err := FetchToken(ctx)
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", errors.New("iam v4 auth token is empty")
	}
	cachedToken = token
	return cachedToken, nil
}

func resetTokenCache() {
	tokenMu.Lock()
	cachedToken = ""
	tokenMu.Unlock()
}
