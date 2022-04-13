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

package utils

import (
	"github.com/emicklei/go-restful"
)

// EmptyResponse is the empty response for restful response
type EmptyResponse struct{}

// ErrorResponse is the error response for restful response
type ErrorResponse struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// WriteFuncFactory builds WriteXXX shortcut functions
func WriteFuncFactory(statusCode int) func(response *restful.Response, code int, message string) {
	return func(response *restful.Response, code int, message string) {
		_ = response.WriteHeaderAndEntity(statusCode, ErrorResponse{
			Result:  false,
			Code:    code,
			Message: message,
			Data:    nil,
		})
	}
}

// WriteClientError writes client error
var WriteClientError = WriteFuncFactory(400)

// WriteUnauthorizedError writes unauthorized error
var WriteUnauthorizedError = WriteFuncFactory(401)

// WriteForbiddenError writes forbidden error
var WriteForbiddenError = WriteFuncFactory(403)

// WriteNotFoundError writes not found error
var WriteNotFoundError = WriteFuncFactory(404)

// WriteServerError writes internal error
var WriteServerError = WriteFuncFactory(500)
