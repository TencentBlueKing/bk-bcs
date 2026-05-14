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

// Package header contains header related constants and functions
package header

import (
	"context"
	"net/textproto"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"k8s.io/klog/v2"
)

// ContextValueKey is the key for context value
type ContextValueKey string

const (
	// AuthUserKey is the key for user in context
	AuthUserKey ContextValueKey = "X-Bcs-User"
	// InnerClientHeaderKey is the key for client in header
	InnerClientHeaderKey = "X-Bcs-Client"
	// AuthorizationHeaderKey is the key for authorization in header
	AuthorizationHeaderKey = "Authorization"
	// CustomUsernameHeaderKey is the key for custom username in header
	CustomUsernameHeaderKey = "X-Bcs-Username"

	// RequestIDKey x-request-id header key
	RequestIDKey = "X-Request-Id"
	// UsernameKey x-project-username header key
	UsernameKey = "X-Project-Username"

	//TraceparentHeaderKey traceparent header key
	TraceparentKey = "Traceparent"
	//TracestateHeaderKey tracestate header key
	TracestateKey = "Tracestate"

	// BkTenantIdHeaderKey is the header name of X-Bk-Tenant-Id.
	BkTenantIdHeaderKey = "X-Bk-Tenant-Id"

	//LaneIDPrefix 染色的header前缀
	LaneIDPrefix = "X-Lane-"
)

var (
	passthroughHeaderKeys = []string{RequestIDKey, TraceparentKey, TracestateKey}
)

// CustomHeaderMatcher for http header
func CustomHeaderMatcher(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	if strings.HasPrefix(key, LaneIDPrefix) {
		return key, true
	}
	switch key {
	case RequestIDKey:
		return RequestIDKey, true
	case UsernameKey:
		return UsernameKey, true
	case InnerClientHeaderKey:
		return InnerClientHeaderKey, true
	case CustomUsernameHeaderKey:
		return CustomUsernameHeaderKey, true
	case TraceparentKey:
		return TraceparentKey, true
	case TracestateKey:
		return TracestateKey, true
	case BkTenantIdHeaderKey:
		return BkTenantIdHeaderKey, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// LaneHeaderInterceptor 透传泳道header
// x-lane-xxx,x-request-id,traceparent
// 从incoming context获取request-id,以及包含泳道前缀的header
func LaneHeaderInterceptor() grpc.UnaryClientInterceptor {
	// 创建一个集合用于快速查找需要透传的键
	passthroughHeaderKeysSet := make(map[string]struct{}, len(passthroughHeaderKeys))
	for _, key := range passthroughHeaderKeys {
		passthroughHeaderKeysSet[textproto.CanonicalMIMEHeaderKey(key)] = struct{}{}
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		passthroughHeader := map[string]string{}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			for k, v := range md {
				if len(v) > 0 {
					tmpKey := textproto.CanonicalMIMEHeaderKey(k)
					// 检查是否包含laneID前缀
					if strings.HasPrefix(tmpKey, LaneIDPrefix) {
						passthroughHeader[k] = v[0]
						continue // 已找到需要透传的，无需继续检查此键
					}
					// 检查是否在需要透传的键集合中
					if _, exists := passthroughHeaderKeysSet[tmpKey]; exists {
						passthroughHeader[k] = v[0]
					}
				}
			}
		}
		klog.Infof("passthrough header: %v", passthroughHeader)
		outgoingCtx := metadata.NewOutgoingContext(ctx, metadata.New(passthroughHeader))
		// 调用下一个处理器
		return invoker(outgoingCtx, method, req, reply, cc, opts...)
	}
}
