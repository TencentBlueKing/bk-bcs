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

// Package tenantutils provides helpers to carry tenant id in context
package tenantutils

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/constants"
)

// WithTenantIdFromContext set tenantID to context
func WithTenantIdFromContext(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, constants.BkTenantIdHeaderKey, tenantId)
}

// GetTenantIdFromContext get tenantId from context
func GetTenantIdFromContext(ctx context.Context) string {
	tenantId := ""

	if id, ok := ctx.Value(constants.BkTenantIdHeaderKey).(string); ok {
		tenantId = id
	}

	if tenantId == "" {
		tenantId = string(constants.DefaultTenantId)
	}

	return tenantId
}
