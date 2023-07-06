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
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
)

// Result 返回的标准结构
type Result struct {
	Code           int             `json:"code"`
	Message        string          `json:"message"`
	RequestId      string          `json:"request_id"`
	Data           interface{}     `json:"data"`
	WebAnnotations *WebAnnotations `json:"web_annotations"`
}

// WebAnnotations 权限信息
type WebAnnotations struct {
	Perms *Perms `json:"perms"`
}

// Perms 无权限返回的需要申请的权限信息
type Perms struct {
	ActionList []utils.ResourceAction `json:"action_list"`
	ApplyURL   string                 `json:"apply_url"`
}

// AbortWithBadRequest 请求参数错误
func AbortWithBadRequest(w http.ResponseWriter, r *http.Request, code int, msg string) {
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, Result{Code: code, Message: msg, RequestId: r.Header.Get(constants.RequestIDHeaderKey)})
}

// AbortWithUnauthorized 请求未认证
func AbortWithUnauthorized(w http.ResponseWriter, r *http.Request, code int, msg string) {
	render.Status(r, http.StatusUnauthorized)
	render.JSON(w, r, Result{Code: code, Message: msg, RequestId: r.Header.Get(constants.RequestIDHeaderKey)})
}

// AbortWithInternalServerError 请求内部错误
func AbortWithInternalServerError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, Result{Code: code, Message: msg, RequestId: r.Header.Get(constants.RequestIDHeaderKey)})
}

// AbortWithForbidden 请求无权限
func AbortWithForbidden(w http.ResponseWriter, r *http.Request, perms *Perms) {
	requestID := r.Header.Get(constants.RequestIDHeaderKey)
	var permissions []string
	for _, action := range perms.ActionList {
		permissions = append(permissions, action.Action)
	}
	msg := fmt.Sprintf("permission denied, need %s permission", strings.Join(permissions, ","))
	result := Result{Code: 40403, Message: msg, RequestId: requestID, WebAnnotations: &WebAnnotations{Perms: perms}}
	render.Status(r, http.StatusForbidden)
	render.JSON(w, r, result)
}

// Success 请求成功
func Success(w http.ResponseWriter, r *http.Request, data interface{}) {
	requestID := r.Header.Get(constants.RequestIDHeaderKey)
	result := Result{Code: 0, Message: "success", RequestId: requestID, Data: data}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, result)
}
