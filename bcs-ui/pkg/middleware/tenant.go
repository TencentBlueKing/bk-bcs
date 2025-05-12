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

package middleware

import (
	"context"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
)

// TenantHandler middleware tenant
func TenantHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// 默认default
		tenantId := constants.DefaultTenantId
		if !config.G.BCS.EnableMultiTenantMode {
			ctx = context.WithValue(ctx, constants.TenantIdCtxKey, tenantId)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		headerTenantId := r.Header.Get(constants.HeaderTenantId)
		if headerTenantId != "" {
			tenantId = headerTenantId
		}
		ctx = context.WithValue(ctx, constants.TenantIdCtxKey, tenantId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
