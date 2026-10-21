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

// Package ctxkey xxx
package ctxkey

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/types"
)

const (
	// RequestIDKey xxx
	RequestIDKey = types.ContextKey("requestID")
	// UsernameKey xxx
	UsernameKey = types.ContextKey("username")
	// TenantIDKey xxx
	TenantIDKey = types.ContextKey("tenantID")
	// UserinfoKey xxx
	UserinfoKey = types.ContextKey("userinfo")
	// ProjKey xxx
	ProjKey = types.ContextKey("project")
	// ClusterKey xxx
	ClusterKey = types.ContextKey("cluster")
	// LangKey 语言版本
	LangKey = types.ContextKey("lang")
	// UserAgentHeaderKey is the header name of User-Agent.
	UserAgentHeaderKey = "Grpcgateway-User-Agent"
	// ForwardedForHeaderKey is the header name of X-Forwarded-For.
	ForwardedForHeaderKey = "X-Forwarded-For"
	// InnerClientHeaderKey is the key for client in header
	InnerClientHeaderKey = "X-Bcs-Client"
	// AuthorizationHeaderKey is the key for authorization in header
	AuthorizationHeaderKey = "Authorization"
	// CustomUsernameHeaderKey is the key for custom username in header
	CustomUsernameHeaderKey = "X-Bcs-Username"
	// TenantIdHeaderKey is the key for tenant id in header
	TenantIdHeaderKey = "X-Bk-Tenant-Id"
)

// GetUsernameFromCtx 通过 ctx 获取 username
func GetUsernameFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(UsernameKey).(string)
	return id
}

// GetTenantIDFromCtx 通过 ctx 获取 tenant id
func GetTenantIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(TenantIDKey).(string)
	fmt.Println("GetTenantIDFromCtx", id)
	return id
}
