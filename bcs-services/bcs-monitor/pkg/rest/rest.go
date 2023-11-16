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
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
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
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// HandlerFunc xxx
type HandlerFunc func(*Context) (interface{}, error)

// StreamHandlerFunc xxx
type StreamHandlerFunc func(*Context)

// AbortWithBadRequestError 请求失败
func AbortWithBadRequestError(c *Context, err error) {
	result := Result{Code: 1400, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusBadRequest, result)
}

// AbortWithUnauthorizedError 未登入
func AbortWithUnauthorizedError(c *Context, err error) {
	result := Result{Code: 1401, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusUnauthorized, result)
}

// AbortWithWithForbiddenError 没有权限
func AbortWithWithForbiddenError(c *Context, err error) {
	result := Result{Code: 1403, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusForbidden, result)
}

// AbortWithJSONError 目前的UI规范, 返回200状态码, 通过里面的code判断请求成功与否
func AbortWithJSONError(c *Context, err error) {
	result := Result{Code: 1400, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusOK, result)
}

// APIResponse 正常返回
func APIResponse(c *Context, data interface{}) {
	result := Result{Code: 0, Message: "OK", RequestId: c.RequestId, Data: data}
	c.JSON(http.StatusOK, result)
}

// InitRestContext :
func InitRestContext(c *gin.Context) *Context {
	requestId := requestid.Get(c)

	restContext := &Context{
		Context:     c,
		RequestId:   requestId,
		ClusterId:   c.Param("clusterId"),
		ProjectId:   c.Param("projectId"),
		ProjectCode: c.Param("projectCode"),
	}
	c.Set("rest_context", restContext)

	tracing.SetRequestIDValue(c.Request, requestId)
	ctx := store.WithRequestIDValue(c.Request.Context(), requestId)
	restContext.Request = restContext.Request.WithContext(ctx)
	return restContext
}

// GetRestContext 查询鉴权信息
func GetRestContext(c *gin.Context) (*Context, error) {
	ctxObj, ok := c.Get("rest_context")
	if !ok {
		return nil, ErrorUnauthorized
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, ErrorUnauthorized
	}

	return restContext, nil
}

// RestHandlerFunc rest handler
func RestHandlerFunc(handler HandlerFunc) gin.HandlerFunc { // nolint
	return func(c *gin.Context) {
		startTime := time.Now()
		// 需要在审计操作记录中对body进行解析
		reqBody := getRequestBody(c.Request)
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		result, err := handler(restContext)
		endTime := time.Now()
		if err != nil {
			// 添加审计中心操作记录
			go addAudit(restContext, reqBody, startTime, endTime, 1400, err.Error())
			AbortWithJSONError(restContext, err)
			return
		}
		// 添加审计中心操作记录
		go addAudit(restContext, reqBody, startTime, endTime, 0, "OK")
		APIResponse(restContext, result)
	}
}

// STDRestHandlerFunc 标准handler, 错误返回非200状态码
func STDRestHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		result, err := handler(restContext)
		if err != nil {
			AbortWithBadRequestError(restContext, err)
			return
		}

		APIResponse(restContext, result)
	}
}

// StreamHandler 流式 Handler
func StreamHandler(handler StreamHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		handler(restContext)
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
	resourceID.RuleID = ctx.Param("id")
	return resourceID
}

var auditFuncMap = map[string]func(b []byte, ctx *Context) (audit.Resource, audit.Action){
	"POST./projects/:projectId/clusters/:clusterId/log_collector/entrypoints": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_log_rule", ActivityType: audit.ActivityTypeView}
	},
	"POST./projects/:projectId/clusters/:clusterId/log_collector/rules": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		// resourceData解析
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "create_log_rule", ActivityType: audit.ActivityTypeCreate}
	},
	"GET./projects/:projectId/clusters/:clusterId/log_collector/rules/:id": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_log_rule", ActivityType: audit.ActivityTypeView}
	},
	"PUT./projects/:projectId/clusters/:clusterId/log_collector/rules/:id": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"DELETE./projects/:projectId/clusters/:clusterId/log_collector/rules/:id": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "delete_log_rule", ActivityType: audit.ActivityTypeDelete}
	},
	"POST./projects/:projectId/clusters/:clusterId/log_collector/rules/:id/retry": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "retry_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"POST./projects/:projectId/clusters/:clusterId/log_collector/rules/:id/enable": func(
		b []byte, ctx *Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "enable_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"POST./projects/:projectId/clusters/:clusterId/log_collector/rules/:id/disable": func(
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
	fn, ok := auditFuncMap[ctx.Request.Method+"."+ctx.FullPath()]
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
