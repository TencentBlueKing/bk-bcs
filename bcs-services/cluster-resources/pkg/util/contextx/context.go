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

// Package contextx xxx
package contextx

import (
	"context"
	"net/http"
	"net/textproto"
	"strings"

	"go-micro.dev/v4/metadata"
)

// GetRequestIDFromCtx 通过 ctx 获取 requestID
func GetRequestIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDContextKey).(string)
	return id
}

// GetProjectIDFromCtx 通过 ctx 获取 projectID
func GetProjectIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(ProjectIDContextKey).(string)
	return id
}

// GetProjectCodeFromCtx 通过 ctx 获取 projectCode
func GetProjectCodeFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(ProjectCodeContextKey).(string)
	return id
}

// GetClusterIDFromCtx 通过 ctx 获取 clusterID
func GetClusterIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(ClusterIDContextKey).(string)
	return id
}

// GetSourceIPFromCtx 通过 ctx 获取 sourceIP
func GetSourceIPFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	forwarded, _ := md.Get(ForwardedForHeaderKey)
	return forwarded
}

// GetUserAgentFromCtx 通过 ctx 获取 userAgent
func GetUserAgentFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	userAgent, _ := md.Get(UserAgentHeaderKey)
	return userAgent
}

// GetLaneIDByCtx get lane id by ctx
func GetLaneIDByCtx(ctx context.Context) map[string]string {
	// http 格式的以key value方式存放，eg: key: X-Lane value: X-Lane-xxx:xxx
	v, ok := ctx.Value(LaneKey).(string)
	if ok || v != "" {
		result := strings.Split(v, ":")
		if len(result) != 2 {
			return nil
		}
		return map[string]string{result[0]: result[1]}
	}
	if !ok || v == "" {
		return grpcLaneIDValue(ctx)
	}
	return nil
}

// grpcLaneIDValue grpc lane id 处理
func grpcLaneIDValue(ctx context.Context) map[string]string {
	md, ok := metadata.FromContext(ctx)
	if ok {
		for k, v := range md {
			tmpKey := textproto.CanonicalMIMEHeaderKey(k)
			if strings.HasPrefix(tmpKey, LaneIDPrefix) {
				return map[string]string{tmpKey: v}
			}
		}
	}
	return nil
}

// WithLaneIdCtx ctx lane id
func WithLaneIdCtx(ctx context.Context, h http.Header) context.Context {
	for k, v := range h {
		if strings.HasPrefix(k, LaneIDPrefix) && len(v) > 0 {
			ctx = context.WithValue(ctx, LaneKey, k+":"+v[0])
		}
	}
	return ctx
}
