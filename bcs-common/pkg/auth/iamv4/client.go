/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

// Package iamv4 提供蓝鲸权限中心 V4（bkiam 网关）OpenAPI 客户端。
package iamv4

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
)

const defaultTimeout = 60 * time.Second

// modelIDRe IAM V4 模型 ID：小写字母开头，小写字母或数字结尾，最长 32
var modelIDRe = regexp.MustCompile(`^[a-z]([a-z0-9_-]{0,30}[a-z0-9])?$`)

// Client IAM V4 客户端，覆盖 bkiam 网关全部开放接口
type Client struct {
	opt  *Options
	http *resty.Client
}

// NewClient 创建 IAM V4 客户端
func NewClient(opt *Options) (*Client, error) {
	if err := opt.validate(); err != nil {
		return nil, err
	}
	cli := resty.New().
		SetTimeout(defaultTimeout).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "bcs-iamv4")
	return &Client{opt: opt, http: cli}, nil
}

func (c *Client) systemID(id string) (string, error) {
	if c == nil || c.opt == nil {
		return "", ErrServerNotInit
	}
	if id == "" {
		id = c.opt.SystemID
	}
	return id, requireModelID("system_id", id)
}

func requireModelID(name, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s is required", name)
	}
	if !modelIDRe.MatchString(id) {
		return fmt.Errorf("%s %q is invalid", name, id)
	}
	return nil
}

func requireNonEmpty(name, val string) error {
	if strings.TrimSpace(val) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func (c *Client) authHeader() string {
	payload, _ := json.Marshal(map[string]string{
		"bk_app_code":   c.opt.AppCode,
		"bk_app_secret": c.opt.AppSecret,
	})
	return string(payload)
}

type rawResponse struct {
	Data      json.RawMessage `json:"data"`
	RequestID string          `json:"request_id"`
	Error     *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Code    json.RawMessage `json:"code"`
	Message string          `json:"message"`
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body interface{},
	extra http.Header, dest interface{}) error {
	if c == nil || c.opt == nil {
		return ErrServerNotInit
	}
	if ctx == nil {
		ctx = context.Background()
	}

	req := c.http.R().SetContext(ctx).
		SetHeader(HeaderBkAPIAuthorization, c.authHeader()).
		SetHeader(HeaderTenantID, c.opt.TenantID)
	for k, vs := range extra {
		for _, v := range vs {
			req.SetHeader(k, v)
		}
	}
	if len(query) > 0 {
		req.SetQueryParamsFromValues(query)
	}
	if body != nil {
		req.SetBody(body)
	}

	fullURL := c.opt.GateWayHost + path
	resp, err := req.Execute(method, fullURL)
	if err != nil {
		klog.Errorf("iamv4 %s %s request failed", method, path)
		return fmt.Errorf("iamv4 request %s %s: %w", method, path, err)
	}

	status := resp.StatusCode()
	raw := &rawResponse{}
	if len(resp.Body()) > 0 {
		if uerr := json.Unmarshal(resp.Body(), raw); uerr != nil {
			if status < 200 || status >= 300 {
				return &APIError{StatusCode: status, Message: "invalid response body"}
			}
			return fmt.Errorf("iamv4 decode %s %s: %w", method, path, uerr)
		}
	}

	if status < 200 || status >= 300 || (raw.Error != nil && raw.Error.Code != "") {
		apiErr := &APIError{StatusCode: status, RequestID: raw.RequestID, Message: raw.Message}
		if raw.Error != nil {
			apiErr.Code = raw.Error.Code
			if raw.Error.Message != "" {
				apiErr.Message = raw.Error.Message
			}
		}
		klog.Errorf("iamv4 %s %s failed: status=%d code=%s request_id=%s message=%s",
			method, path, status, apiErr.Code, apiErr.RequestID, apiErr.Message)
		return apiErr
	}

	if dest == nil {
		return nil
	}
	if len(raw.Data) > 0 && string(raw.Data) != "null" {
		if err := json.Unmarshal(raw.Data, dest); err != nil {
			return fmt.Errorf("iamv4 decode data %s %s: %w", method, path, err)
		}
		return nil
	}
	// 部分接口直接返回业务体（无 data / request_id 包裹）
	if len(resp.Body()) > 0 && raw.RequestID == "" && raw.Error == nil && raw.Message == "" {
		if err := json.Unmarshal(resp.Body(), dest); err != nil {
			return fmt.Errorf("iamv4 decode body %s %s: %w", method, path, err)
		}
	}
	return nil
}

func operatorHeader(operator string) (http.Header, error) {
	if err := requireNonEmpty("operator", operator); err != nil {
		return nil, ErrEmptyOperator
	}
	h := make(http.Header)
	h.Set(HeaderIAMOperator, operator)
	return h, nil
}

func isNotFound(err error) bool {
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		return false
	}
	if apiErr.StatusCode == http.StatusNotFound {
		return true
	}
	code := strings.ToUpper(apiErr.Code)
	return strings.Contains(code, "NOT_FOUND") || strings.Contains(code, "DOES_NOT_EXIST")
}

func isAlreadyExists(err error) bool {
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		return false
	}
	if apiErr.StatusCode == http.StatusConflict {
		return true
	}
	code := strings.ToUpper(apiErr.Code)
	msg := strings.ToLower(apiErr.Message)
	return strings.Contains(code, "EXIST") || strings.Contains(code, "CONFLICT") ||
		strings.Contains(msg, "already") || strings.Contains(msg, "exist")
}

func asAPIError(err error, target **APIError) bool {
	if err == nil {
		return false
	}
	var e *APIError
	if !errors.As(err, &e) {
		return false
	}
	*target = e
	return true
}
