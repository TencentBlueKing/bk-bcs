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

	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
)

// Result 返回的标准结构
type Result struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// AbortWithUnauthorized 请求未认证
func AbortWithUnauthorized(w http.ResponseWriter, r *http.Request, code int, msg string) {
	render.Status(r, http.StatusUnauthorized)
	render.JSON(w, r, Result{Code: code, Message: msg, RequestId: r.Header.Get(constants.RequestIDHeaderKey)})
}

// Success 请求成功
func Success(w http.ResponseWriter, r *http.Request, data interface{}) {
	requestID := r.Header.Get(constants.RequestIDHeaderKey)
	result := Result{Code: 0, Message: "success", RequestId: requestID, Data: data}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, result)
}
