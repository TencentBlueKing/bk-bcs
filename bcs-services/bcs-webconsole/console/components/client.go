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

// Package components xxx
package components

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/dustin/go-humanize"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/tracing"
)

type ctxKey int

const (
	timeout = time.Second * 30
	// BKAPIRequestIDHeader 蓝鲸网关的请求ID
	BKAPIRequestIDHeader = "X-Bkapi-Request-Id"
	userAgent            = "bcs-webconsole/v1.0"
	requestIDCtxKey      = ctxKey(1)
	requestIDHeaderKey   = "X-Request-Id"
)

var (
	maskKeys = map[string]struct{}{
		"bk_app_secret": {},
	}
	clientOnce   sync.Once
	globalClient *resty.Client

	// dialer with timeout
	dialer = &net.Dialer{
		Timeout:   timeout,
		KeepAlive: timeout,
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

// WithRequestIDValue 设置 RequestId 值
func WithRequestIDValue(ctx context.Context, id string) context.Context {
	newCtx := context.WithValue(ctx, requestIDCtxKey, id)
	return metadata.AppendToOutgoingContext(newCtx, requestIDHeaderKey, id)
}

// RequestIDValue 获取 RequestId 值
func RequestIDValue(ctx context.Context) string {
	v, ok := ctx.Value(requestIDCtxKey).(string)
	if !ok || v == "" {
		return grpcRequestIDValue(ctx)
	}

	return v
}

// grpcRequestIDValue grpc 需要单独处理
func grpcRequestIDValue(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(requestIDHeaderKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// SetRequestIDHeaderValue 设置 RequestId 值到头部
func SetRequestIDHeaderValue(req *http.Request, id string) {
	req.Header.Set(requestIDHeaderKey, id)
}

// restyReqToCurl curl 格式的请求日志
func restyReqToCurl(r *resty.Request) string {
	headers := ""
	for key, values := range r.Header {
		for _, value := range values {
			headers += fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %s", key, value))
		}
	}

	// 过滤掉敏感信息
	rawURL := *r.RawRequest.URL
	queryValue := rawURL.Query()
	for key := range queryValue {
		if _, ok := maskKeys[key]; ok {
			queryValue.Set(key, "<masked>")
		}
	}
	rawURL.RawQuery = queryValue.Encode()

	reqMsg := fmt.Sprintf("curl -X %s '%s'%s", r.Method, rawURL.String(), headers)
	if r.Body != nil {
		switch body := r.Body.(type) {
		case []byte:
			reqMsg += fmt.Sprintf(" -d %q", body)
		case string:
			reqMsg += fmt.Sprintf(" -d %q", body)
		case io.Reader:
			reqMsg += fmt.Sprintf(" -d %q (io.Reader)", body)
		default:
			prtBodyBytes, err := json.Marshal(body)
			if err != nil {
				reqMsg += fmt.Sprintf(" -d %q (MarshalErr %s)", body, err)
			} else {
				reqMsg += fmt.Sprintf(" -d '%s'", prtBodyBytes)
			}
		}
	}
	if r.FormData.Encode() != "" {
		encodeStr := r.FormData.Encode()
		reqMsg += fmt.Sprintf(" -d %q", encodeStr)
		rawStr, _ := url.QueryUnescape(encodeStr)
		reqMsg += fmt.Sprintf(" -raw `%s`", rawStr)
	}

	return reqMsg
}

// restyResponseToCurl 返回日志
func restyResponseToCurl(resp *resty.Response) string {
	// 最大打印 1024 个字符
	body := string(resp.Body())
	if len(body) > 1024 {
		body = fmt.Sprintf("%s...(Total %s)", body[:1024], humanize.Bytes(uint64(len(body))))
	}

	respMsg := fmt.Sprintf("[%s] %s %s", resp.Status(), resp.Time(), body)

	// 请求蓝鲸网关记录RequestID
	bkAPIRequestID := resp.RawResponse.Header.Get(BKAPIRequestIDHeader)
	if bkAPIRequestID != "" {
		respMsg = fmt.Sprintf("[%s] %s bkapi_request_id=%s %s", resp.Status(), resp.Time(), bkAPIRequestID, body)
	}

	return respMsg
}

func restyErrHook(r *resty.Request, err error) {
	blog.Infof("[%s] REQ: %s", RequestIDValue(r.RawRequest.Context()), restyReqToCurl(r))
	blog.Infof("[%s] RESP: [err] %s", RequestIDValue(r.RawRequest.Context()), err)
}

func restyAfterResponseHook(c *resty.Client, r *resty.Response) error {
	blog.V(6).Infof("[%s] REQ: %s", RequestIDValue(r.Request.Context()), restyReqToCurl(r.Request))
	blog.V(6).Infof("[%s] RESP: %s", RequestIDValue(r.Request.Context()), restyResponseToCurl(r))
	return nil
}

func restyBeforeRequestHook(c *resty.Client, r *resty.Request) error {
	SetRequestIDHeaderValue(r.RawRequest, RequestIDValue(r.Context()))
	return nil
}

// GetClient : 新建Client, 设置公共参数，每次新建，cookies不复用
func GetClient() *resty.Client {
	if globalClient == nil {
		clientOnce.Do(func() {
			globalClient = resty.New().
				SetTimeout(timeout).
				SetTransport(tracing.NewTracingTransport(defaultTransport)). // 设置tracingTransport 传递tracing
				SetDebug(false).                                             // 更多详情, 可以开启为 true
				SetCookieJar(nil).                                           // 后台API去掉 cookie 记录
				SetDebugBodyLimit(1024).
				OnAfterResponse(restyAfterResponseHook).
				SetPreRequestHook(restyBeforeRequestHook).
				OnError(restyErrHook).
				SetHeader("User-Agent", userAgent)
		})
	}
	return globalClient
}

// BKResult 蓝鲸返回规范的结构体
type BKResult struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UnmarshalBKResult 反序列化为蓝鲸返回规范
func UnmarshalBKResult(resp *resty.Response, data interface{}) error {
	if resp.StatusCode() != http.StatusOK {
		return errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	// 部分接口，如 usermanager 返回的content-type不是json, 需要手动Unmarshal
	bkResult := &BKResult{Data: data}
	if err := json.Unmarshal(resp.Body(), bkResult); err != nil {
		return err
	}

	if err := bkResult.ValidateCode(); err != nil {
		return err
	}

	return nil
}

// ValidateCode 返回结果是否OK
func (r *BKResult) ValidateCode() error {
	var resultCode int

	switch code := r.Code.(type) {
	case int:
		resultCode = code
	case float64:
		resultCode = int(code)
	case string:
		c, err := strconv.Atoi(code)
		if err != nil {
			return err
		}
		resultCode = c
	default:
		return errors.Errorf("conversion to int from %T not supported", code)
	}

	if resultCode != 0 {
		return errors.Errorf("resp code %d != 0, %s", resultCode, r.Message)
	}
	return nil
}
