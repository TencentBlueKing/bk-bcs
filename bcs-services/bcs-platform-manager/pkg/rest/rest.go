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

// Package rest xxx
package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/ggicci/httpin"
	"github.com/gin-contrib/sse"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/rest/tracing"
)

var (
	// ErrorUnauthorized 错误
	ErrorUnauthorized = errors.New("用户未登入")
)

// Result 返回的标准结构
type Result struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
	Data      any    `json:"data"`
	HTTPCode  int    `json:"-"` // http response status code
}

// Render chi render interface implementation
func (e *Result) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPCode)
	return nil
}

// HandlerFunc xxx
type HandlerFunc[In, Out any] func(context.Context, *In) (*Out, error)

// StreamHandlerFunc  ServerStreaming or BidiStreaming handle function
type StreamHandlerFunc[In any] func(*In, StreamingServer) error

// AbortWithBadRequestError 请求失败
func AbortWithBadRequestError(c *Context, err error) render.Renderer {
	return &Result{Code: 1400, Message: err.Error(), RequestId: c.RequestId, HTTPCode: http.StatusBadRequest}
}

// AbortWithUnauthorizedError 未登入
func AbortWithUnauthorizedError(c *Context, err error) render.Renderer {
	return &Result{Code: 1401, Message: err.Error(), RequestId: c.RequestId, HTTPCode: http.StatusUnauthorized}
}

// AbortWithWithForbiddenError 没有权限
func AbortWithWithForbiddenError(c *Context, err error) render.Renderer {
	return &Result{Code: 1403, Message: err.Error(), RequestId: c.RequestId, HTTPCode: http.StatusForbidden}
}

// AbortWithJSONError 目前的UI规范, 返回200状态码, 通过里面的code判断请求成功与否
func AbortWithJSONError(c *Context, err error) render.Renderer {
	return &Result{Code: 1400, Message: err.Error(), RequestId: c.RequestId, HTTPCode: http.StatusOK}
}

// APIResponse 正常返回
func APIResponse(c *Context, data any) render.Renderer {
	return &Result{Code: 0, Message: "OK", RequestId: c.RequestId, Data: data, HTTPCode: http.StatusOK}
}

// Event sse event
type Event struct {
	HTTPCode int `json:"-"` // http response status code
	sse.Event
}

// Render chi render interface implementation
func (e *Event) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPCode)
	return e.Event.Render(w)
}

// restContext
type restCtx string

const (
	restContextKey restCtx = "rest_context"
)

// InitRestContext :
func InitRestContext(w http.ResponseWriter, r *http.Request) *Context {
	requestId := tracing.GetRequestIDResp(w)

	restContext := &Context{
		Request:     r,
		RequestId:   requestId,
		ClusterId:   chi.URLParam(r, "clusterId"),
		ProjectId:   chi.URLParam(r, "projectId"),
		ProjectCode: chi.URLParam(r, "projectCode"),
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, restContextKey, restContext)

	tracing.SetRequestIDValue(r, requestId)
	restContext.Request = r.WithContext(ctx)
	return restContext
}

// SetRestContext 设置鉴权信息
func SetRestContext(ctx context.Context, rctx *Context) context.Context {
	return context.WithValue(ctx, restContextKey, rctx)
}

// GetRestContext 查询鉴权信息
func GetRestContext(ctx context.Context) (*Context, error) {
	restContext, ok := ctx.Value(restContextKey).(*Context)
	if !ok {
		return nil, ErrorUnauthorized
	}

	return restContext, nil
}

// Handle generic handle
func Handle[In, Out any](handle HandlerFunc[In, Out]) http.HandlerFunc {
	handleName := getHandleName(handle)
	return func(w http.ResponseWriter, r *http.Request) {
		restContext, err := GetRestContext(r.Context())
		if err != nil {
			_ = render.Render(w, r, AbortWithUnauthorizedError(InitRestContext(w, r), err))
			return
		}
		restContext.HandleName = handleName

		in, err := decodeReq[In](r)
		if err != nil {
			blog.Errorf("handle decode request failed, err: %s", err)
			_ = render.Render(w, r, AbortWithJSONError(restContext, err))
			return
		}

		err = Struct(r.Context(), in)
		if err != nil {
			blog.Errorf("valid request param failed, err: %s", err)
			_ = render.Render(w, r, AbortWithJSONError(restContext, err))
			return
		}

		result, err := handle(r.Context(), in)
		if err != nil {
			_ = render.Render(w, r, AbortWithJSONError(restContext, err))
			return
		}
		_ = render.Render(w, r, APIResponse(restContext, result))
	}
}

// getHandleName 获取FuncHandle/StreamHandle函数名
func getHandleName(fn any) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	if fullName == "" {
		panic("get func name is empty")
	}

	parts := strings.Split(fullName, ".")
	lastPart := parts[len(parts)-1]
	name := strings.TrimSuffix(lastPart, "-fm")
	return name
}

// decodeReq ...
func decodeReq[T any](r *http.Request) (*T, error) {
	in := new(T)
	var err error

	// http.Request 直接返回
	if _, ok := any(in).(*http.Request); ok {
		return any(r).(*T), nil
	}

	// 空值不需要反序列化
	if _, ok := any(in).(*EmptyReq); ok {
		return in, nil
	}

	in, err = httpin.Decode[T](r)
	if err != nil {
		return nil, err
	}

	// Get/Delete 请求, 请求参数从url中获取
	if r.Method == http.MethodGet || r.Method == http.MethodDelete {
		return in, nil
	}

	// Post 请求等, 从body中获取
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, in); err != nil {
		return nil, fmt.Errorf("unmarshal json body: %s", err)
	}
	return in, nil
}

var validate = validator.New()

// Struct request validate
func Struct(ctx context.Context, in any) error {
	return validate.StructCtx(ctx, in)
}

// EmptyReq 空的请求
type EmptyReq struct{}

// StreamingServer server or bidi streaming server
type StreamingServer interface {
	http.ResponseWriter
	Context() context.Context
	Flush() error
}

type streamingServer struct {
	http.ResponseWriter
	*http.ResponseController
	ctx context.Context
}

// Context return svr's context
func (s *streamingServer) Context() context.Context {
	return s.ctx
}

// Flush return svr's Flush
func (s *streamingServer) Flush() error {
	return s.ResponseController.Flush()
}

// Stream 流式 Handle
func Stream[In any](handler StreamHandlerFunc[In]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		restContext, err := GetRestContext(r.Context())
		if err != nil {
			_ = render.Render(w, r, AbortWithUnauthorizedError(InitRestContext(w, r), err))
			return
		}

		in, err := decodeReq[In](r)
		if err != nil {
			blog.Errorf("handle decode request failed, err: %s", err)
			_ = render.Render(w, r, AbortWithJSONError(restContext, err))
			return
		}
		err = Struct(r.Context(), in)
		if err != nil {
			blog.Errorf("valid request param failed, err: %s", err)
			_ = render.Render(w, r, AbortWithJSONError(restContext, err))
			return
		}
		svr := &streamingServer{
			ResponseWriter:     w,
			ResponseController: http.NewResponseController(w),
			ctx:                r.Context(),
		}
		err = handler(in, svr)
		if err != nil {
			_ = render.Render(w, r, AbortWithBadRequestError(restContext, err))
		}
	}
}
