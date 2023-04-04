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
	"errors"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"google.golang.org/grpc/status"

	"bscp.io/pkg/components"
	"bscp.io/pkg/components/bkpaas"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	pbas "bscp.io/pkg/protocol/auth-server"
	"bscp.io/pkg/rest"
)

// UnifiedAuthentication
// HTTP API 鉴权, 异常返回json信息
func (a authorizer) UnifiedAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		req, err := getUserCredentialFromCookies(r)
		if err != nil {
			render.Render(w, r, rest.UnauthorizedErr(err))
			return
		}
		var username string
		if req.Token == constant.BKTokenForTest {
			username = r.Header.Get(constant.UserKey)
		} else {
			resp, err := a.authClient.GetUserInfo(r.Context(), req)
			if err != nil {
				s := status.Convert(err)
				render.Render(w, r, rest.UnauthorizedErr(errors.New(s.Message())))
				return
			}
			username = resp.Username
		}

		k := &kit.Kit{
			Ctx:         r.Context(),
			User:        username,
			Rid:         components.RequestIDValue(r.Context()),
			AppId:       chi.URLParam(r, "app_id"),
			AppCode:     "dummyApp", // 测试 App
			SpaceID:     "",
			SpaceTypeID: "",
		}
		ctx := kit.WithKit(r.Context(), k)

		r.Header.Set(constant.AppCodeKey, k.AppCode)
		r.Header.Set(constant.RidKey, k.Rid)
		r.Header.Set(constant.UserKey, k.User)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// WebAuthentication
// HTTP 前端鉴权, 异常调整302到登入页面
func (a authorizer) WebAuthentication(webHost, loginHost string) func(http.Handler) http.Handler {
	ignoreExtMap := map[string]struct{}{
		".js":  {},
		".css": {},
		".map": {},
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// 静态资源过滤, 注意不会带鉴权信息
			fileExt := filepath.Ext(r.URL.Path)
			if _, ok := ignoreExtMap[fileExt]; ok {
				next.ServeHTTP(w, r)
				return
			}

			req, err := getUserCredentialFromCookies(r)
			if err != nil {
				http.Redirect(w, r, bkpaas.BuildLoginRedirectURL(r, webHost, loginHost), http.StatusFound)
				return
			}

			resp, err := a.authClient.GetUserInfo(r.Context(), req)
			if err != nil {
				http.Redirect(w, r, bkpaas.BuildLoginRedirectURL(r, webHost, loginHost), http.StatusFound)
				return
			}

			k := &kit.Kit{User: resp.Username}
			ctx := kit.WithKit(r.Context(), k)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// AppVerified App校验中间件, 需要放到 UnifiedAuthentication 后面, url 需要添加 {app_id} 变量
func (a authorizer) AppVerified(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		kt := kit.MustGetKit(r.Context())
		appIDStr := chi.URLParam(r, "app_id")
		if appIDStr == "" {
			err := errors.New("app_id is required in url params")
			render.Render(w, r, rest.BadRequest(err))
			return
		}

		appID, err := strconv.Atoi(appIDStr)
		if err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}

		space, err := a.authClient.QuerySpaceByAppID(r.Context(), &pbas.QuerySpaceByAppIDReq{AppId: uint32(appID)})
		if err != nil {
			s := status.Convert(err)
			render.Render(w, r, rest.BadRequest(errors.New(s.Message())))
			return
		}

		kt.SpaceID = space.SpaceId
		kt.SpaceTypeID = space.SpaceTypeId
		ctx := kit.WithKit(r.Context(), kt)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// dummyVerified dummy鉴权方式，测试使用
func dummyVerified(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		k := &kit.Kit{
			Ctx:         r.Context(),
			User:        "",
			Rid:         components.RequestIDValue(r.Context()),
			AppId:       "",
			AppCode:     "dummyApp", // 测试 App
			SpaceID:     "",
			SpaceTypeID: "",
		}
		ctx := kit.WithKit(r.Context(), k)

		r.Header.Set(constant.AppCodeKey, k.AppCode)
		r.Header.Set(constant.RidKey, k.Rid)
		r.Header.Set(constant.UserKey, k.User)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// IAMVerified IAM 回调鉴权
func IAMVerified(next http.Handler) http.Handler {
	return dummyVerified(next)
}

// BKRepoVerified bk_repo 回调鉴权
func BKRepoVerified(next http.Handler) http.Handler {
	return dummyVerified(next)
}
