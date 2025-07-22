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

package utils

import (
	"fmt"
	"strings"

	restful "github.com/emicklei/go-restful/v3"
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

// -----v2-----

// SuccessResponse is the success response for restful response
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ErrorResponseV2 is the error response for restful response
type ErrorResponseV2 struct {
	Error Error `json:"error"`
}

// Error is the error response for restful response
type Error struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data"`
	Details []ErrorDetail `json:"details"`
}

// ErrorDetail is the error detail response for restful response
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PermDeniedError permission denied,user need to apply
type PermDeniedError struct {
	Perms PermData `json:"perms"`
}

// PermData permission data for no permission
type PermData struct {
	ApplyURL   string           `json:"apply_url"`
	ActionList []ResourceAction `json:"action_list"`
}

// Error return error message with perm actions
func (e *PermDeniedError) Error() string {
	var actionList []string
	for _, action := range e.Perms.ActionList {
		actionList = append(actionList, action.Action)
	}
	actions := strings.Join(actionList, ",")
	return fmt.Sprintf("permission denied, need %s permission", actions)
}

// ResourceAction for multi action multi resources
type ResourceAction struct {
	Resource string `json:"-"`
	Type     string `json:"resource_type"`
	Action   string `json:"action_id"`
}

// ResponseOK response ok
func ResponseOK(response *restful.Response, data interface{}) {
	_ = response.WriteHeaderAndEntity(200, SuccessResponse{
		Data: data,
	})
}

// ResponseAuthError response auth error
func ResponseAuthError(response *restful.Response) {
	_ = response.WriteHeaderAndEntity(401, ErrorResponseV2{
		Error: Error{
			Code:    "UNAUTHENTICATED",
			Message: "user not authenticated",
			Data:    make([]ParamsErrorData, 0),
			Details: make([]ErrorDetail, 0),
		},
	})
}

// ResponsePermissionError response permission error
func ResponsePermissionError(response *restful.Response, permsErr error) {
	switch err := permsErr.(type) {
	case *PermDeniedError:
		_ = response.WriteHeaderAndEntity(403, ErrorResponseV2{
			Error: Error{
				Code:    "IAM_NO_PERMISSION",
				Message: err.Error(),
				Data:    err.Perms,
				Details: make([]ErrorDetail, 0),
			},
		})
	default:
		_ = response.WriteHeaderAndEntity(403, ErrorResponseV2{
			Error: Error{
				Code:    "NO_PERMISSION",
				Message: "user has no permission",
				Data:    make([]ParamsErrorData, 0),
				Details: []ErrorDetail{{Code: "NO_PERMISSION", Message: err.Error()}},
			},
		})
	}
}

// ResponseSuccess response success
func ResponseSuccess(response *restful.Response, code int, data interface{}) {
	_ = response.WriteHeaderAndEntity(code, SuccessResponse{
		Data: data,
	})
}

// ResponseParamsError response params error
func ResponseParamsError(response *restful.Response, err error) {
	validationError := ParseValidationError(err)
	if len(validationError) == 0 {
		_ = response.WriteHeaderAndEntity(400, ErrorResponseV2{
			Error: Error{
				Code:    "INVALID_ARGUMENT",
				Message: err.Error(),
				Data:    make([]ParamsErrorData, 0),
				Details: make([]ErrorDetail, 0),
			},
		})
		return
	}
	_ = response.WriteHeaderAndEntity(400, ErrorResponseV2{
		Error: Error{
			Code:    "INVALID_ARGUMENT",
			Message: "invalid argument",
			Data:    validationError,
			Details: make([]ErrorDetail, 0),
		},
	})
}

// ResponseDBError response db error
func ResponseDBError(response *restful.Response, err error) {
	_ = response.WriteHeaderAndEntity(500, ErrorResponseV2{
		Error: Error{
			Code:    "INTERNAL",
			Message: "database error",
			Data:    make([]ParamsErrorData, 0),
			Details: []ErrorDetail{
				{Code: "INTERNAL", Message: err.Error()},
			},
		},
	})
}

// ResponseSystemError response system error
func ResponseSystemError(response *restful.Response, err error) {
	_ = response.WriteHeaderAndEntity(500, ErrorResponseV2{
		Error: Error{
			Code:    "INTERNAL",
			Message: "api system error",
			Data:    make([]ParamsErrorData, 0),
			Details: []ErrorDetail{
				{Code: "INTERNAL", Message: err.Error()},
			},
		},
	})
}
