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

	"google.golang.org/grpc/metadata"
)

// ContextKey is the key for context
type ContextKey string

const (
	// LaneKey is the key for lane
	LaneKey ContextKey = "X-Lane"

	// LaneIDPrefix 染色的header前缀
	LaneIDPrefix = "X-Lane-"
)

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
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range md {
			tmpKey := textproto.CanonicalMIMEHeaderKey(k)
			if strings.HasPrefix(tmpKey, LaneIDPrefix) && len(v) > 0 {
				return map[string]string{tmpKey: md.Get(k)[0]}
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
