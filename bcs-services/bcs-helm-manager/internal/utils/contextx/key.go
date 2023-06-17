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

package contextx

// ContextKey xxx
type ContextKey string

const (
	// RequestIDContextKey 请求的requestID
	RequestIDContextKey ContextKey = "requestID"
	// TraceIDContextKey 链路跟踪需要的trace id
	TraceIDContextKey ContextKey = "traceID"
	// UsernamContextKey 用户名
	UsernamContextKey ContextKey = "username"
	// ProjectIDContextKey projectID context key
	ProjectIDContextKey ContextKey = "projectID"
	// ProjectCodeContextKey projectCode context key
	ProjectCodeContextKey ContextKey = "projectCode"
	// LangContectKey
	LangContectKey ContextKey = "lang"
)

// HeaderKey string
const (
	// RequestIDKey ...
	RequestIDHeaderKey = "X-Request-Id"
	// ContentDispositionKey content disposition key
	ContentDispositionKey = "content-disposition"
	// ContentDispositionKey content disposition key
	ContentDispositionCapKey = "Content-Disposition"
	// ContentDispositionValue contenct disposition value
	ContentDispositionValue = "Content-Disposition"
)
