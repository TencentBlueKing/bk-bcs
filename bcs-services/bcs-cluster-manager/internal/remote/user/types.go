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
 *
 */

package user

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

const (
	// ModuleUserManager default discovery usermanager module
	ModuleUserManager = "usermanager.bkbcs.tencent.com"
)

const (
	// PermissionKind kind
	PermissionKind = "permission"
	// PermissionVersion version
	PermissionVersion = "v1"
	// PermissionName name
	PermissionName = "user-permission"
)

const (
	// PermissionManagerRole manager
	PermissionManagerRole = "manager"
	// PermissionViewerRole manager
	PermissionViewerRole = "viewer"
)

const (
	// ResourceTypeClusterManager cm module
	ResourceTypeClusterManager = "clustermanager"
	// ResourceTypeCluster cluster module
	ResourceTypeCluster = "cluster"
)

const (
	// ResourceScopeAll all cluster
	ResourceScopeAll = "*"
)

// CreateTokenReq request
type CreateTokenReq struct {
	// default plain user
	UserType uint   `json:"usertype"`
	Username string `json:"username" validate:"required"`
	// token expiration second, -1: never expire
	Expiration int `json:"expiration" validate:"required"`
}

// CommonResp common resp
type CommonResp struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CreateTokenResp response
type CreateTokenResp struct {
	CommonResp `json:",inline"`
	Data       TokenResp `json:"data"`
}

// GetTokenResp response
type GetTokenResp struct {
	CommonResp `json:",inline"`
	Data       []TokenResp `json:"data"`
}

// TokenStatus is a enum for token status.
type TokenStatus uint8

const (
	// TokenStatusExpired mean that token is expired.
	TokenStatusExpired TokenStatus = iota
	// TokenStatusActive mean that token is active.
	TokenStatusActive
)

// TokenResp xxx
type TokenResp struct {
	Token     string     `json:"token"`
	ExpiredAt *time.Time `json:"expired_at"` // nil means never expired
}

// VerifyPermissionReq for permission v2 request
type VerifyPermissionReq struct {
	UserToken    string `json:"user_token" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	Resource     string `json:"resource"`
	Action       string `json:"action" validate:"required"`
}

// VerifyPermissionResponse xxx
type VerifyPermissionResponse struct {
	CommonResp `json:",inline"`
	Data       VerifyPermissionResult `json:"data"`
}

// VerifyPermissionResponse http verify response
type VerifyPermissionResult struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message"`
}

func buildUserPermission(permissions []types.Permission) types.BcsPermission {
	return types.BcsPermission{
		TypeMeta: types.TypeMeta{
			APIVersion: PermissionVersion,
			Kind:       PermissionKind,
		},
		ObjectMeta: types.ObjectMeta{
			Name: PermissionName,
		},
		Spec: types.BcsPermissionSpec{
			Permissions: permissions,
		},
	}
}
