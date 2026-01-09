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
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/rest"
)

// TanantCheck middleware for tenant check
func TanantCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, err := decodeBCSJwtFromContext(ctx, r)
		if err != nil {
			rest.AbortWithUnauthorized(w, r, http.StatusUnauthorized, err.Error())
			return
		}

		// get tenant id
		headerTenantId := r.Header.Get(constants.TenantIDHeaderKey)
		tenantId := func() string {
			if headerTenantId != "" {
				return headerTenantId
			}
			if claims.TenantId != "" {
				return claims.TenantId
			}
			return constants.DefaultTenantID
		}()

		r = r.WithContext(context.WithValue(r.Context(), constants.ContextTenantKey, tenantId))
		next.ServeHTTP(w, r)
	})
}
