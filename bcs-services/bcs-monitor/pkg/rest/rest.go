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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/ggicci/httpin"
	"github.com/gin-contrib/sse"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/thanos-io/thanos/pkg/store"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest/tracing"
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
type HandlerFunc[In, Out any] func(context.Context, In) (Out, error)

// StreamHandlerFunc  ServerStreaming or BidiStreaming handle function
type StreamHandlerFunc[In any] func(In, StreamingServer) error

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
	ctx = store.WithRequestIDValue(ctx, requestId)
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

// Handler handler
func Handler[In, Out any](handler HandlerFunc[In, Out]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		// 需要在审计操作记录中对body进行解析
		reqBody := getRequestBody(r)
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
		result, err := handler(r.Context(), *in)
		endTime := time.Now()
		if err != nil {
			// 添加审计中心操作记录
			go addAudit(restContext, reqBody, startTime, endTime, 1400, err.Error())
			_ = render.Render(w, r, AbortWithJSONError(restContext, err))
			return
		}
		// 添加审计中心操作记录
		go addAudit(restContext, reqBody, startTime, endTime, 0, "OK")
		_ = render.Render(w, r, APIResponse(restContext, result))
	}
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

// StreamHandler 流式 Handler
func StreamHandler[In any](handler StreamHandlerFunc[In]) func(w http.ResponseWriter, r *http.Request) {
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
		svr := &streamingServer{
			ResponseWriter:     w,
			ResponseController: http.NewResponseController(w),
			ctx:                r.Context(),
		}
		err = handler(*in, svr)
		if err != nil {
			_ = render.Render(w, r, AbortWithBadRequestError(restContext, err))
		}
	}
}

type resource struct {
	ClusterID string `json:"cluster_id" yaml:"cluster_id"`
	ProjectID string `json:"project_id" yaml:"project_id"`
	Name      string `json:"name" yaml:"name"`
	RuleID    string `json:"-" yaml:"-"`
}

// resource to map
func (r resource) toMap() map[string]any {
	result := make(map[string]any, 0)

	if r.ClusterID != "" {
		result["ClusterID"] = r.ClusterID
	}

	if r.ProjectID != "" {
		result["ProjectID"] = r.ProjectID
	}

	if r.Name != "" {
		result["Name"] = r.Name
	}

	if r.RuleID != "" {
		result["RuleID"] = r.RuleID
	}

	return result
}

// 获取resourceData 的资源
func getResourceID(b []byte, ctx *Context) resource {
	resourceID := resource{}
	_ = json.Unmarshal(b, &resourceID)
	resourceID.ClusterID = ctx.ClusterId
	resourceID.ProjectID = ctx.ProjectId
	resourceID.RuleID = chi.URLParam(ctx.Request, "id")
	return resourceID
}

var auditFuncMap = map[string]func(b []byte, ctx *Context) (audit.Resource, audit.Action){
	"POST./projects/{projectId}/clusters/{clusterId}/log_collector/entrypoints": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_log_rule", ActivityType: audit.ActivityTypeView}
	},
	"POST./projects/{projectId}/clusters/{clusterId}/log_collector/rules": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		// resourceData解析
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "create_log_rule", ActivityType: audit.ActivityTypeCreate}
	},
	"GET./projects/{projectId}/clusters/{clusterId}/log_collector/rules/{id}": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_log_rule", ActivityType: audit.ActivityTypeView}
	},
	"PUT./projects/{projectId}/clusters/{clusterId}/log_collector/rules/{id}": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"DELETE./projects/{projectId}/clusters/{clusterId}/log_collector/rules/{id}": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "delete_log_rule", ActivityType: audit.ActivityTypeDelete}
	},
	"POST./projects/{projectId}/clusters/{clusterId}/log_collector/rules/{id}/retry": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "retry_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"POST./projects/{projectId}/clusters/{clusterId}/log_collector/rules/{id}/enable": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "enable_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"POST./projects/{projectId}/clusters/{clusterId}/log_collector/rules/{id}/disable": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "disable_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
}

// 审计中心新增操作记录
func addAudit(ctx *Context, b []byte, startTime, endTime time.Time, code int, message string) {
	// get method audit func
	uri := chi.RouteContext(ctx.Request.Context()).RoutePatterns
	fn, ok := auditFuncMap[ctx.Request.Method+"."+getCompleteRoutePatterns(uri)]
	if !ok {
		return
	}

	res, act := fn(b, ctx)

	auditCtx := audit.RecorderContext{
		Username:  ctx.Username,
		RequestID: ctx.RequestId,
		StartTime: startTime,
		EndTime:   endTime,
	}
	resource := audit.Resource{
		ProjectCode:  ctx.ProjectCode,
		ResourceType: res.ResourceType,
		ResourceID:   res.ResourceID,
		ResourceName: res.ResourceName,
		ResourceData: res.ResourceData,
	}
	action := audit.Action{
		ActionID:     act.ActionID,
		ActivityType: act.ActivityType,
	}

	result := audit.ActionResult{
		Status:        audit.ActivityStatusSuccess,
		ResultCode:    code,
		ResultContent: message,
	}

	// code不为0的情况则为失败
	if code != 0 {
		result.Status = audit.ActivityStatusFailed
	}

	// add audit
	auditAction := component.GetAuditClient().R()
	// 查看类型不用记录activity
	if act.ActivityType == audit.ActivityTypeView {
		auditAction.DisableActivity()
	}
	_ = auditAction.SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}

// 获取请求体
func getRequestBody(r *http.Request) []byte {
	// 读取请求体
	body, _ := io.ReadAll(r.Body)
	// 恢复请求体
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

// 获取完整原始uri
func getCompleteRoutePatterns(s []string) string {
	return strings.ReplaceAll(path.Join(s...), "/*", "")
}
