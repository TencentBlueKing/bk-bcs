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

package auth

import (
	"net/http"

	"bscp.io/pkg/kit"
	"bscp.io/pkg/rest"
	"github.com/go-chi/render"
)

// UnifiedAuthentication
// HTTP API 鉴权, 异常返回json信息
func (a authorizer) UnifiedAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		req, err := getUserCredentialFromCookies(r)
		if err != nil {
			render.Render(w, r, rest.UnauthorizedErrRender(err))
			return
		}
		resp, err := a.authClient.GetUserInfo(r.Context(), req)
		if err != nil {
			render.Render(w, r, rest.UnauthorizedErrRender(err))
			return
		}

		k := &kit.Kit{User: resp.Username}
		ctx := kit.WithKit(r.Context(), k)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// WebAuthentication
// HTTP 前端鉴权, 异常调整302到登入页面
func (a authorizer) WebAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
	}
	return http.HandlerFunc(fn)
}
