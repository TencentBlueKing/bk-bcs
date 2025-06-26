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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

// jwtClient is the jwt client
var jwtClient *jwt.JWTClient

// InitJWTClient init jwt client
func InitJWTClient(op *options.ClusterManagerOptions) error {
	cli, err := jwt.NewJWTClient(jwt.JWTOptions{
		VerifyKeyFile: op.Auth.PublicKeyFile,
		SignKeyFile:   op.Auth.PrivateKeyFile,
	})
	jwtClient = cli
	if err != nil {
		return err
	}
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

// GetHeaderTenantIdFromCtx 通过 ctx 获取当前请求的租户信息
func GetHeaderTenantIdFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	tenantId, _ := md.Get(common.BkTenantIdHeaderKey)
	return tenantId
}

// GetAuthUserInfoFromCtx 通过 ctx 获取当前用户信息
func GetAuthUserInfoFromCtx(ctx context.Context) middleauth.AuthUser {
	authUser, _ := middleauth.GetUserFromContext(ctx)
	return authUser
}

// UserInfoCtx 用户信息
type UserInfoCtx struct {
	Username         string
	TenantId         string
	ResourceTenantId string
}

// GetAuthAndTenantInfoFromCtx 通过 ctx 获取当前用户和租户信息
func GetAuthAndTenantInfoFromCtx(ctx context.Context) UserInfoCtx {
	authUser := GetAuthUserInfoFromCtx(ctx)

	user := UserInfoCtx{
		Username:         authUser.GetUsername(),
		TenantId:         authUser.GetTenantId(),
		ResourceTenantId: authUser.GetTenantId(),
	}

	// 兼容跨租户场景，不在租户的人员可以获取其他租户的资源
	if options.GetGlobalCMOptions().TenantConfig.EnableMultiTenantMode {
		headerTenantId := GetHeaderTenantIdFromCtx(ctx)
		if headerTenantId != "" {
			user.ResourceTenantId = headerTenantId
		}
	}

	if user.ResourceTenantId == "" {
		user.ResourceTenantId = utils.DefaultTenantId
	}

	return user
}
