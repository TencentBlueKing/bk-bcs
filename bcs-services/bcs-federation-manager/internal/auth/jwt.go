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

package auth

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"go-micro.dev/v4/metadata"
)

// jwtClient is the jwt client
var jwtClient *jwt.JWTClient

// JWTOptions jwt options
type JWTOptions struct {
	PublicKeyFile  string
	PrivateKeyFile string
}

// InitJWTClient init jwt client
func InitJWTClient(op *JWTOptions) error {
	cli, err := jwt.NewJWTClient(jwt.JWTOptions{
		VerifyKeyFile: op.PublicKeyFile,
		SignKeyFile:   op.PrivateKeyFile,
	})
	if err != nil {
		return err
	}
	jwtClient = cli
	return nil
}

// GetJWTClient get jwt client
func GetJWTClient() *jwt.JWTClient {
	return jwtClient
}

// GetUserFromCtx 通过 ctx 获取当前用户
func GetUserFromCtx(ctx context.Context) string {
	authUser, _ := middleauth.GetUserFromContext(ctx)
	return authUser.GetUsername()
}

// GetRealUserFromCtx 通过 ctx 判断当前用户是否是真实用户
func GetRealUserFromCtx(ctx context.Context) string {
	authUser, _ := middleauth.GetUserFromContext(ctx)
	return authUser.Username
}

// TenantIDFromCtx 解析本次请求的租户。优先级：请求头、JWT，空则 default。
// 依赖的 bcs-auth AuthUser 还没有 TenantId 字段，JWT 租户从 token 再解一次。
func TenantIDFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return iam.DefaultTenantId
	}
	if id, _ := md.Get(iam.HeaderTenantId); id != "" {
		return id
	}
	if token, _ := md.Get(middleauth.AuthorizationHeaderKey); token != "" {
		if id := tenantIDFromJWT(token); id != "" {
			return id
		}
	}
	return iam.DefaultTenantId
}

func tenantIDFromJWT(jwtToken string) string {
	if GetJWTClient() == nil || !strings.HasPrefix(jwtToken, "Bearer ") {
		return ""
	}
	claims, err := GetJWTClient().JWTDecode(jwtToken[len("Bearer "):])
	if err != nil || claims == nil {
		return ""
	}
	return claims.TenantId
}
