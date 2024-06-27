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

// Package roundtrip client
package roundtrip

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	ctxkey "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/header"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
)

// Client http client(自带重试)
type Client interface {
	// Get http get method
	Get(ctx context.Context, traceId, url string, data []byte) ([]byte, error)

	// Post http post method
	Post(ctx context.Context, traceId, url string, data []byte) ([]byte, error)

	// Put http put method
	Put(ctx context.Context, traceId, url string, data []byte) ([]byte, error)

	// Delete http delete method
	Delete(ctx context.Context, traceId, url string, data []byte) ([]byte, error)
}

// NewClient return Client
func NewClient(config *options.Config) Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			// NOCC:gas/tls(设计如此)
			InsecureSkipVerify: config.InsecureSkipVerify, // nolint
		},
	}

	c := &http.Client{
		Transport: transport,
	}

	return &client{
		client: c,
		config: config,
	}
}

// client impl Client
type client struct {
	// config 配置
	config *options.Config

	// client http client
	client *http.Client
}

// pre 检查 RequestIDKey 或 设置RequestIDKey
func (h *client) pre(traceId string) string {
	if len(traceId) == 0 {
		return ctxkey.GenUUID()
	}

	return traceId
}

// Get http get method
func (h *client) Get(ctx context.Context, traceId, url string, data []byte) ([]byte, error) {
	return h.roundTrip(ctx, h.pre(traceId), http.MethodGet, url, data)
}

// Post http post method
func (h *client) Post(ctx context.Context, traceId, url string, data []byte) ([]byte, error) {
	return h.roundTrip(ctx, h.pre(traceId), http.MethodPost, url, data)
}

// Put http put method
func (h *client) Put(ctx context.Context, traceId, url string, data []byte) ([]byte, error) {
	return h.roundTrip(ctx, h.pre(traceId), http.MethodPut, url, data)
}

// Delete http delete method
func (h *client) Delete(ctx context.Context, traceId, url string, data []byte) ([]byte, error) {
	return h.roundTrip(ctx, h.pre(traceId), http.MethodDelete, url, data)
}

// roundTrip 请求
func (h *client) roundTrip(ctx context.Context, traceId, method, url string, data []byte) ([]byte, error) {
	var err error
	var body []byte

	if body, err = h.doRequest(ctx, traceId, method, url, data); err == nil {
		return body, nil
	}

	if !strings.Contains(err.Error(), "status code is not 200") {
		return nil, err
	}

	return body, nil
}

// setHeader 设置请求头
func (h *client) setHeader(req *http.Request, traceId string) {
	req.Header.Add(ctxkey.RequestIDKey, traceId)
	req.Header.Add(ctxkey.UsernameKey, h.config.Username)
	req.Header.Add(ctxkey.Authorization, fmt.Sprintf("Bearer %s", h.config.Token))
}

// doRequest 请求
func (h *client) doRequest(ctx context.Context, traceId, method, url string, data []byte) ([]byte, error) {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrapf(err, "new request failed, traceId: %s, url: %s, body: %s",
			traceId, url, string(data))
	}
	h.setHeader(req, traceId)

	// 请求server
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "http request failed, traceId: %s, url: %s, body: %s",
			traceId, url, string(data))
	}
	defer resp.Body.Close()
	blog.Infof("traceId: %s, statusCode: %d", traceId, resp.StatusCode) // note: debug

	// 读取body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read response body failed, traceId: %s, url: %s", traceId, url)
	}

	// 状态判断
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(
			"request failed, status code is not 200, traceId: %s, statusCode: %d, url: %s, content: %s", traceId,
			resp.StatusCode, url, cast.ToString(body))
	}
	blog.Infof("request success, traceId: %s, body: %s", traceId, cast.ToString(body)) // note: debug

	return body, nil
}
