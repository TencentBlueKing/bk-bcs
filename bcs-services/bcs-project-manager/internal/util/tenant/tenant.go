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

// Package tenant xxx
package tenant

import (
	"context"
	"fmt"

	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/headerkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
)

// IsMultiTenantEnabled 检查是否启用了多租户模式
func IsMultiTenantEnabled() bool {
	return config.GlobalConf.EnableMultiTenant
}

// ResourceMetaData xxx
type ResourceMetaData struct {
	ProjectId   string
	ProjectCode string
	TenantId    string
}

// WithTenantIdByResourceForContext set tenantID by resource to context
func WithTenantIdByResourceForContext(ctx context.Context, resource ResourceMetaData) (context.Context, error) {
	if !IsMultiTenantEnabled() {
		return context.WithValue(ctx, headerkey.TenantIdKey, constant.DefaultTenantId), nil
	}

	// 优先使用resource中的租户ID
	if resource.TenantId != "" {
		return context.WithValue(ctx, headerkey.TenantIdKey, resource.TenantId), nil
	}

	var (
		projectIndex = resource.ProjectId
	)
	if projectIndex == "" && resource.ProjectCode != "" {
		projectIndex = resource.ProjectCode
	}

	if projectIndex == "" {
		return ctx, fmt.Errorf("projectIndex is empty")
	}

	pro, err := store.GetModel().GetProject(ctx, projectIndex)
	if err != nil {
		return ctx, err
	}

	if pro.TenantID == "" {
		pro.TenantID = constant.DefaultTenantId
	}

	// 注入租户信息
	return context.WithValue(ctx, headerkey.TenantIdKey, pro.TenantID), nil
}

// WithTenantIdFromContext set tenantID to context
func WithTenantIdFromContext(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, headerkey.TenantIdKey, tenantId)
}

// GetTenantIdFromContext get tenantId from context
func GetTenantIdFromContext(ctx context.Context) string {
	tenantId := ""

	if id, ok := ctx.Value(headerkey.TenantIdKey).(string); ok {
		tenantId = id
	}

	if tenantId == "" {
		tenantId = utils.DefaultTenantId
	}

	return tenantId
}

// GetUserFromCtx 通过 ctx 获取当前用户
func GetUserFromCtx(ctx context.Context) string {
	authUser, _ := middleauth.GetUserFromContext(ctx)
	return authUser.GetUsername()
}

// UserInfoCtx 用户信息
type UserInfoCtx struct {
	Username         string
	TenantId         string
	ResourceTenantId string
}

// GetUsername 获取当前用户
func (user UserInfoCtx) GetUsername() string {
	return user.Username
}

// GetUserTenantId 获取当前用户的租户ID
func (user UserInfoCtx) GetUserTenantId() string {
	return user.TenantId
}

// GetResourceTenantId 获取当前请求资源的租户ID
func (user UserInfoCtx) GetResourceTenantId() string {
	return user.ResourceTenantId
}

// GetAuthAndTenantInfoFromCtx 通过 ctx 获取当前用户和租户信息
func GetAuthAndTenantInfoFromCtx(ctx context.Context) UserInfoCtx {
	authUser, _ := GetAuthUserInfoFromCtx(ctx)

	user := UserInfoCtx{
		Username:         authUser.GetUsername(),
		TenantId:         GetTenantIdFromContext(ctx),
		ResourceTenantId: GetTenantIdFromContext(ctx),
	}

	// 兼容跨租户场景，不在租户的人员可以获取其他租户的资源
	if IsMultiTenantEnabled() {
		headerTenantId := getHeaderTenantIdFromCtx(ctx)
		if headerTenantId != "" {
			user.ResourceTenantId = headerTenantId
		}
	}

	if user.ResourceTenantId == "" {
		user.ResourceTenantId = utils.DefaultTenantId
	}

	return user
}

// getHeaderTenantIdFromCtx 通过 ctx 获取当前请求的租户信息
func getHeaderTenantIdFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	tenantId, _ := md.Get(string(headerkey.TenantIdKey))
	return tenantId
}

// GetAuthUserInfoFromCtx 通过 ctx 获取当前用户信息
func GetAuthUserInfoFromCtx(ctx context.Context) (*middleauth.AuthUser, error) {
	authUser, err := middleauth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return &authUser, nil
}
