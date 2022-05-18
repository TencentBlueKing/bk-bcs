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
 *
 */

package rest

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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

// HandlerFunc
type HandlerFunc func(*Context) (interface{}, error)

// StreamHandlerFunc
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

// APIResponse 正常返回
func APIResponse(c *Context, data interface{}) {
	result := Result{Code: 0, Message: "OK", RequestId: c.RequestId, Data: data}
	c.JSON(http.StatusOK, result)
}

// RequestIdGenerator
func RequestIdGenerator() string {
	uid := uuid.New().String()
	requestId := strings.Replace(uid, "-", "", -1)
	return requestId
}

// InitRestContext
func InitRestContext(c *gin.Context) *Context {
	restContext := &Context{
		Context:   c,
		RequestId: requestid.Get(c),
	}
	c.Set("rest_context", restContext)
	return restContext
}

// GetAuthContext 查询鉴权信息
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
