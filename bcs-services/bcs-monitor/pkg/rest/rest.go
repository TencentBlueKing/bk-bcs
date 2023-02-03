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
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/thanos-io/thanos/pkg/store"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest/tracing"
)

var (
	// UnauthorizedError 错误
	UnauthorizedError = errors.New("用户未登入")
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
		return nil, UnauthorizedError
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, UnauthorizedError
	}

	return restContext, nil
}

// RestHandlerFunc rest handler
func RestHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		result, err := handler(restContext)
		if err != nil {
			AbortWithJSONError(restContext, err)
			return
		}

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
