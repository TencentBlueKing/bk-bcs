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
	"errors"
	"strings"

	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common/headerkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/util/stringx"
)

// NewInjectContextWrapper 生成 request id, 用于操作审计等便于跟踪
func NewInjectContextWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		ctx = context.WithValue(ctx, ctxkey.RequestIDKey, getRequestID(ctx))

		// 解析jwt，获取username，并注入到context
		var username string
		if auth.CanExemptAuth(req.Endpoint()) {
			username = common.AnonymousUsername
		} else {
			md, ok := metadata.FromContext(ctx)
			if !ok {
				return errors.New("failed to get micro's metadata")
			}
			// 解析到jwt
			jwtToken, ok := md.Get("Authorization")
			if !ok {
				return errors.New("failed to get authorization token!")
			}
			// 判断jwt格式正确
			if len(jwtToken) == 0 || !strings.HasPrefix(jwtToken, "Bearer ") {
				return errors.New("authorization token error")
			}
			authUser, err := auth.ParseUserFromJWT(jwtToken[7:])
			if err != nil {
				return err
			}
			username = authUser.Username
			// 如果jwt token解析的username为空，并且clientid在豁免的白名单中，通过指定header获取用户名
			if username == "" && isClientIDInWhitelist(authUser.ClientID) {
				username, ok = md.Get(headerkey.UsernameKey)
				if !ok {
					return errors.New("username not found from header")
				}
			}
			if username == "" {
				return errors.New("username is null")
			}
		}
		ctx = context.WithValue(ctx, ctxkey.UsernameKey, username)
		return fn(ctx, req, rsp)
	}
}

// 获取 request id
func getRequestID(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return stringx.GenUUID()
	}
	// 当request id不存在或者为空时，生成id
	requestID, ok := md.Get(headerkey.RequestIDKey)
	if !ok || requestID == "" {
		return stringx.GenUUID()
	}

	return requestID
}

func isClientIDInWhitelist(clientID string) bool {
	if clientID == "" {
		return false
	}
	clientIDs := stringx.SplitString(auth.JWTConfig.ExemptClients)
	return stringx.StringInSlice(clientID, clientIDs)
}
