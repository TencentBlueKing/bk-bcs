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

// Package utils xxx
package utils

import (
	"context"

	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// ContextKey xxx
type ContextKey string

const (
	// ProjectIDContextKey projectID context key
	ProjectIDContextKey ContextKey = "projectID"
	// ProjectCodeContextKey projectCode context key
	ProjectCodeContextKey ContextKey = "projectCode"
	// RequestIDContextKey requestID for request
	RequestIDContextKey ContextKey = "requestID"
	// TraceIDContextKey trace id for tracing
	TraceIDContextKey ContextKey = "traceID"
	// UsernameContextKey username for request
	UsernameContextKey ContextKey = "username"
	// MeshIDContextKey meshID context key
	MeshIDContextKey ContextKey = "meshID"
	// PrimaryClustersContextKey primaryClusters context key
	PrimaryClustersContextKey ContextKey = "primaryClusters"
	// RemoteClustersContextKey remoteClusters context key
	RemoteClustersContextKey ContextKey = "remoteClusters"
)

// GetProjectIDFromCtx 通过 ctx 获取 projectID
func GetProjectIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(ProjectIDContextKey).(string)
	return id
}

// GetMeshIDFromCtx 通过 ctx 获取 meshID
func GetMeshIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(MeshIDContextKey).(string)
	return id
}

// GetPrimaryClustersFromCtx 通过 ctx 获取 primaryClusters
func GetPrimaryClustersFromCtx(ctx context.Context) []string {
	clusters, _ := ctx.Value(PrimaryClustersContextKey).([]string)
	return clusters
}

// GetRemoteClustersFromCtx 通过 ctx 获取 remoteClusters
func GetRemoteClustersFromCtx(ctx context.Context) []string {
	clusters, _ := ctx.Value(RemoteClustersContextKey).([]string)
	return clusters
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
