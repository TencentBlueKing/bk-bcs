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

// Package handler provides the gRPC handlers for the PushManager service.
package handler

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// LogError provides a standard way to log errors.
func LogError(method string, err error) {
	blog.Infof("[%s] Error: %v", method, err)
}

// LogSuccess provides a standard way to log successful operations.
func LogSuccess(method string, message string) {
	blog.Infof("[%s] Success: %s", method, message)
}

// LogReceived provides a standard way to log received requests.
func LogReceived(method string, req interface{}) {
	blog.Infof("[%s] Received: %+v", method, req)
}

// FormatErrorMessage formats a standard error message string.
func FormatErrorMessage(operation string, err error) string {
	return fmt.Sprintf("failed to %s: %v", operation, err)
}

// StandardErrorResponse defines a standard structure for error responses.
type StandardErrorResponse struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new standard error response.
func NewErrorResponse(code uint32, message string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Code:    code,
		Message: message,
	}
}

// NewSuccessResponse creates a new standard success response.
func NewSuccessResponse() *StandardErrorResponse {
	return &StandardErrorResponse{
		Code:    0,
		Message: "success",
	}
}
