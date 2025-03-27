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

// Package middleware xxx
package middleware

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

const (
	// AuthUserKey is the key for user in context
	AuthUserKey ContextValueKey = "X-Bcs-User"
	// InnerClientHeaderKey is the key for client in header
	InnerClientHeaderKey = "X-Bcs-Client"
	// AuthorizationHeaderKey is the key for authorization in header
	AuthorizationHeaderKey = "Authorization"
	// CustomUsernameHeaderKey is the key for custom username in header
	CustomUsernameHeaderKey = "X-Bcs-Username"
)

// ContextValueKey is the key for context value
type ContextValueKey string

// AuthUser is the user info in context
type AuthUser struct {
	InnerClient string
	ClientName  string
	Username    string
	TenantId    string
}

// IsInner returns true if the user is inner client
func (u AuthUser) IsInner() bool {
	return len(u.InnerClient) > 0
}

// GetUsername returns the username
func (u AuthUser) GetUsername() string {
	if len(u.Username) != 0 {
		return u.Username
	}
	if len(u.ClientName) != 0 {
		return u.ClientName
	}
	return u.InnerClient
}

// GetTenantId returns the tenant id
func (u AuthUser) GetTenantId() string {
	if u.TenantId == "" {
		return utils.DefaultTenantId
	}

	return u.TenantId
}

// GetUserFromContext returns the user info in context
func GetUserFromContext(ctx context.Context) (AuthUser, error) {
	authUser, ok := ctx.Value(AuthUserKey).(AuthUser)
	if !ok {
		return AuthUser{}, errors.New("get user from context failed, user not found")
	}
	return authUser, nil
}
