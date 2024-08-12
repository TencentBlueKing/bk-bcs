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

// Package http xxx
package http

import (
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// CustomHeaderMatcher 自定义 HTTP Header Matcher
func CustomHeaderMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return "X-Request-Id", true
	case "Traceparent":
		// http -> grpc Traceparent
		return "Traceparent", true
	case ctxkey.CustomUsernameHeaderKey:
		return ctxkey.CustomUsernameHeaderKey, true
	case ctxkey.InnerClientHeaderKey:
		return ctxkey.CustomUsernameHeaderKey, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// 会在 websocket 连接中被转发的 Header Key（可按需添加）
var wsHeadersToForward = []string{"origin", "referer", "authorization", "cookie"}

// WSHeaderForwarder websocket Headers 转发规则
func WSHeaderForwarder(header string) bool {
	return slice.StringInSlice(strings.ToLower(header), wsHeadersToForward)
}
