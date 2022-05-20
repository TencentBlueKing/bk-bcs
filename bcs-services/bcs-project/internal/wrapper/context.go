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
	"strings"

	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/auth"
	constant "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/headerkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/stringx"
)

// NewInjectContextWrapper 生成 request id, 用于操作审计等便于跟踪
func NewInjectContextWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		// generate uuid， e.g. 40a05290d67a4a39a04c705a0ee56add
		// TODO: trace id by opentelemetry
		uuid := stringx.GenUUID()
		ctx = context.WithValue(ctx, ctxkey.RequestIDKey, uuid)

		// 解析jwt，获取username，并注入到context
		var username string
		if auth.CanExemptAuth(req.Endpoint()) {
			username = constant.AnonymousUsername
		} else {
			md, ok := metadata.FromContext(ctx)
			if !ok {
				return RenderResponse(rsp, uuid, errorx.NewAuthErr("failed to get micro's metadata"))
			}
			// 解析到jwt
			jwtToken, ok := md.Get("Authorization")
			if !ok {
				return errorx.NewAuthErr("failed to get authorization token!")
			}
			// 判断jwt格式正确
			if len(jwtToken) == 0 || !strings.HasPrefix(jwtToken, "Bearer ") {
				return errorx.NewAuthErr("authorization token error")
			}
			authUser, err := auth.ParseUserFromJWT(jwtToken[7:])
			if err != nil {
				return RenderResponse(rsp, uuid, err)
			}
			username = authUser.Username
			// NOTE: 现阶段兼容处理非用户态token
			// 当通过认证后，认为是合法的Token，然后判断用户类型为非用户态时，通过header中获取真正的操作者
			if (*authUser).UserType != auth.UserType {
				username, ok = md.Get(headerkey.UsernameKey)
				if !ok {
					return errorx.NewAuthErr("not found username from header")
				}
			}
			// 注入client ID
			ctx = context.WithValue(ctx, ctxkey.ClientID, authUser.ClientID)
		}
		ctx = context.WithValue(ctx, ctxkey.UsernameKey, username)
		return fn(ctx, req, rsp)
	}
}

// NewLogWrapper 记录流水
func NewLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		requestIDKey := ctxkey.RequestIDKey
		md, _ := metadata.FromContext(ctx)
		logging.Info("request func %s, request_id: %s, ctx: %v", req.Endpoint(), ctx.Value(requestIDKey), md)
		if err := fn(ctx, req, rsp); err != nil {
			logging.Error("request func %s failed, request_id: %s, ctx: %v, body: %v", req.Endpoint(), ctx.Value(requestIDKey), md, req.Body())
			return err
		}
		return nil
	}
}
