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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"

	constant "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
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
		if canExemptAuth(req) {
			username = constant.AnonymousUsername
		} else {
			md, ok := metadata.FromContext(ctx)
			if !ok {
				return RenderResponse(rsp, uuid, errorx.New(errcode.UnauthErr, "failed to get micro's metadata"))
			}

			username, err = parseUsername(md)
			if err != nil {
				return RenderResponse(rsp, uuid, err)
			}
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

// NoAuthEndpoints 不需要用户身份认证的方法
var NoAuthEndpoints = []string{
	"Healthz.Ping",
	"Healthz.Healthz",
}

// 检查当前请求是否允许免除用户认证
func canExemptAuth(req server.Request) bool {
	// 禁用身份认证
	if !config.G.JWT.Enable {
		return true
	}
	// 特殊指定的Handler，不需要认证的方法
	return stringx.StringInSlice(req.Endpoint(), NoAuthEndpoints)
}

func getJWTOpt(md metadata.Metadata) (*jwt.JWTOptions, error) {
	jwtOpt := &jwt.JWTOptions{
		VerifyKeyFile: config.G.JWT.PublicKeyFile,
		SignKeyFile:   config.G.JWT.PrivateKeyFile,
	}
	publicKey := config.G.JWT.PublicKey
	privateKey := config.G.JWT.PrivateKey

	if publicKey != "" {
		key, err := jwtGo.ParseRSAPublicKeyFromPEM([]byte(publicKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.VerifyKey = key
	}
	if privateKey != "" {
		key, err := jwtGo.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.SignKey = key
	}
	return jwtOpt, nil
}

func parseUsername(md metadata.Metadata) (string, error) {
	jwtToken, ok := md.Get("Authorization")
	if !ok {
		return "", errorx.New(errcode.UnauthErr, "failed to get authorization token!")
	}
	if len(jwtToken) == 0 || !strings.HasPrefix(jwtToken, "Bearer ") {
		return "", errorx.New(errcode.UnauthErr, "authorization token error")
	}
	// 组装 jwt client
	jwtOpt, err := getJWTOpt(md)
	if err != nil {
		return "", errorx.New(errcode.UnauthErr, "parse jwt key error", err.Error())
	}
	jwtClient, err := jwt.NewJWTClient(*jwtOpt)
	if err != nil {
		return "", err
	}
	// 解析token
	claims, err := jwtClient.JWTDecode(jwtToken[7:])
	if err != nil {
		return "", err
	}
	return claims.UserName, nil
}
