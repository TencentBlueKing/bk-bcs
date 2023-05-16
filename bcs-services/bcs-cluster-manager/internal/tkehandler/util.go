/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tkehandler

import (
	"errors"
	"fmt"

	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"

	"github.com/emicklei/go-restful"
	"gopkg.in/go-playground/validator.v9"
)

// error code to be compitable with interfaces in bcs-api
const (
	httpCodeClientError       = 400
	httpCodeUnauthorizedError = 401
	httpCodeForbiddenError    = 403
	httpCodeNotFoundError     = 404
	httpCodeServerError       = 500
)

// EmptyResponse empty response
type EmptyResponse struct{}

// ErrorResponse error response
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

// WriteClientError function for client error
var WriteClientError = WriteFuncFactory(httpCodeClientError)

// WriteUnauthorizedError function for unauthorizerd error
var WriteUnauthorizedError = WriteFuncFactory(httpCodeUnauthorizedError)

// WriteForbiddenError function for forbidden error
var WriteForbiddenError = WriteFuncFactory(httpCodeForbiddenError)

// WriteNotFoundError function for not found error
var WriteNotFoundError = WriteFuncFactory(httpCodeNotFoundError)

// WriteServerError function for server error
var WriteServerError = WriteFuncFactory(httpCodeServerError)

// FormatValidationError turn the original validatetion errors into error response, it will only use the FIRST
// errorField to construct the error message.
func FormatValidationError(errList error) *ErrorResponse {
	var message string
	for _, err := range errList.(validator.ValidationErrors) {
		if err.Tag() == "required" {
			message = fmt.Sprintf("errcode: %d, ",
				types.BcsErrClusterManagerInvalidParameter) + fmt.Sprintf(`field '%s' is required`, err.Field())
			break
		}
		message = fmt.Sprintf("errcode: %d, ", types.BcsErrClusterManagerInvalidParameter) +
			fmt.Sprintf(`'%s' failed on the '%s' tag`, err.Field(), err.Tag())
	}
	return &ErrorResponse{
		Result:  false,
		Code:    types.BcsErrClusterManagerInvalidParameter,
		Message: message,
		Data:    nil,
	}
}

// CreateResponeData common response
func CreateResponeData(err error, msg string, data interface{}) string {
	var rpyErr error
	if err != nil {
		rpyErr = bhttp.InternalError(types.BcsErrClusterManagerCommonErr, msg)
	} else {
		rpyErr = errors.New(bhttp.GetRespone(
			types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr, data))
	}
	return rpyErr.Error()
}
