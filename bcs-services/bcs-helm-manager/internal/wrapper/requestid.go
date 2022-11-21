/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wrapper

import (
	"context"
	"reflect"

	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
)

// RequestIDWrapper get or generate request id
func RequestIDWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		requestID := getRequestID(ctx)
		ctx = context.WithValue(ctx, contextx.RequestIDContextKey, requestID)
		err = fn(ctx, req, rsp)
		v := reflect.ValueOf(rsp)
		if v.Elem().FieldByName("RequestID") != (reflect.Value{}) {
			v.Elem().FieldByName("RequestID").Set(reflect.ValueOf(&requestID))
		}
		return err
	}
}

// getRequestID 获取 request id
func getRequestID(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return stringx.GenUUID()
	}
	// 当request id不存在或者为空时，生成id
	requestID, ok := md.Get(contextx.RequestIDHeaderKey)
	if !ok || requestID == "" {
		return stringx.GenUUID()
	}

	return requestID
}
