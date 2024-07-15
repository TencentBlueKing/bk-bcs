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

import "context"

// ContextKey xxx
type ContextKey string

const (
	// RequestIDContextKey 请求的requestID
	RequestIDContextKey ContextKey = "requestID"
	// TraceIDContextKey 链路跟踪需要的trace id
	TraceIDContextKey ContextKey = "traceID"
	// UsernameContextKey 用户名
	UsernameContextKey ContextKey = "username"
)

// HeaderKey string
const (
	// RequestIDHeaderKey xxx
	RequestIDHeaderKey = "X-Request-Id"
)

const (
	// TraceID resource trace
	TraceID        = "traceID"
	defaultTraceID = "1234567890zxcvbnm"
	taskID         = "taskID"
)

// GetTaskIDFromContext get taskID from context
func GetTaskIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(taskID).(string); ok {
		return id
	}

	return ""
}

// GetTraceIDFromContext get traceID from context
func GetTraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(TraceID).(string); ok {
		return id
	}

	return defaultTraceID
}

// WithTraceIDForContext inject traceID to context
func WithTraceIDForContext(ctx context.Context, traceID string) context.Context {
	// NOCC:golint/type(设计如此)
	return context.WithValue(ctx, TraceID, traceID) // nolint
}
