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

package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbas "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// initKitWithBKJWT 蓝鲸网关鉴权
func (a authorizer) initKitWithBKJWT(r *http.Request, k *kit.Kit, multiErr *multierror.Error) bool {
	if a.gwParser == nil {
		err := errors.New("gw pubkey is empty")
		multiErr.Errors = append(multiErr.Errors, errors.Wrap(err, "auth with bk_jwt"))
		return false
	}

	kt, err := a.gwParser.Parse(r.Context(), r.Header)
	if err != nil {
		multiErr.Errors = append(multiErr.Errors, errors.Wrap(err, "auth with bk_jwt"))
		return false
	}

	// jwt 只会从jwt里面解析出 app_code
	// user 会从jwt获取, fallback 从 X-Bkapi-User-Name 头部获取(app校验成功, 说明有权限, 网关使用场景)
	k.AppCode = kt.AppCode
	k.User = kt.User
	return true
}

// initKitWithCookie 蓝鲸统一登入Cookie鉴权
func (a authorizer) initKitWithCookie(r *http.Request, k *kit.Kit, multiErr *multierror.Error) bool {
	loginCred, err := a.authLoginClient.GetLoginCredentialFromCookies(r)
	if err != nil {
		multiErr.Errors = append(multiErr.Errors, errors.Wrap(err, "auth with cookie"))
		return false
	}

	req := &pbas.UserCredentialReq{Uid: loginCred.UID, Token: loginCred.Token}
	if req.Token == constant.BKTokenForTest {
		username := r.Header.Get(constant.UserKey)
		if username != "" {
			k.User = username
			return true
		}
	}

	resp, err := a.authClient.GetUserInfo(k.RpcCtx(), req)
	if err != nil {
		s := status.Convert(err)
		// 无权限的需要特殊跳转
		if s.Code() == codes.PermissionDenied {
			multiErr.Errors = append(multiErr.Errors, errors.Wrap(errf.ErrPermissionDenied, s.Message()))
		} else {
			multiErr.Errors = append(multiErr.Errors, errors.Wrap(errors.New(s.Message()), "auth with cookie"))
		}

		return false
	}

	// 登入态只支持用户名
	k.User = resp.Username
	return true
}

// initKitWithDevEnv Dev环境, 可以设置环境变量鉴权
func (a authorizer) initKitWithDevEnv(_ *http.Request, k *kit.Kit, _ *multierror.Error) bool {
	user := os.Getenv("BK_USER_FOR_TEST")
	appCode := os.Getenv("BK_APP_CODE_FOR_TEST")

	if user != "" && appCode != "" {
		k.User = user
		k.AppCode = appCode
		return true
	}

	return false
}

// UnifiedAuthentication HTTP API 鉴权, 异常返回json信息
func (a authorizer) UnifiedAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		k := &kit.Kit{
			Ctx: r.Context(),
			Rid: components.RequestIDValue(r.Context()),
		}
		k.Lang = tools.GetLangFromReq(r)
		multiErr := &multierror.Error{}

		switch {
		case a.initKitWithBKJWT(r, k, multiErr):
		case a.initKitWithCookie(r, k, multiErr):
		case a.initKitWithDevEnv(r, k, multiErr):
		default:
			// API类返回规范的JSON错误信息
			loginURL, loginPlainURL := a.authLoginClient.BuildLoginURL(r)
			render.Render(w, r, rest.NotLoggedInErr(multiErr, loginURL, loginPlainURL))
			return
		}

		ctx := kit.WithKit(r.Context(), k)
		r.Header.Set(constant.AppCodeKey, k.AppCode)
		r.Header.Set(constant.RidKey, k.Rid)
		r.Header.Set(constant.UserKey, k.User)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// WebAuthentication HTTP 前端鉴权, 异常跳转302到登入页面
func (a authorizer) WebAuthentication(webHost string) func(http.Handler) http.Handler {
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

			k := &kit.Kit{
				Ctx: r.Context(),
				Rid: components.RequestIDValue(r.Context()),
			}
			multiErr := &multierror.Error{}

			switch {
			case a.initKitWithCookie(r, k, multiErr):
			default:
				// 如果无权限, 跳转到403页面
				for _, err := range multiErr.Errors {
					if errors.Is(err, errf.ErrPermissionDenied) {
						msg := base64.StdEncoding.EncodeToString([]byte(errf.GetErrMsg(err)))
						redirectURL := fmt.Sprintf("/403.html?msg=%s", url.QueryEscape(msg))
						http.Redirect(w, r, redirectURL, http.StatusFound)
						return
					}
				}

				// web类型做302跳转登入
				http.Redirect(w, r, a.authLoginClient.BuildLoginRedirectURL(r, webHost), http.StatusFound)
				return
			}

			ctx := kit.WithKit(r.Context(), k)
			r.Header.Set(constant.AppCodeKey, k.AppCode)
			r.Header.Set(constant.RidKey, k.Rid)
			r.Header.Set(constant.UserKey, k.User)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// ContentVerified 内容操作校验中间件, 需要放到UnifiedAuthentication和BizVerified后面
// 服务下的配置项内容需要校验服务权限，模版空间下的模版配置项内容需要校验模版空间权限
func (a authorizer) ContentVerified(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		kt := kit.MustGetKit(r.Context())

		appID, tmplSpaceID, err := repository.GetContentLevelID(r)
		if err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}

		if appID > 0 {
			// NOTE: authenticate app on iam

			space, err := a.authClient.QuerySpaceByAppID(kt.RpcCtx(), &pbas.QuerySpaceByAppIDReq{AppId: appID})
			if err != nil {
				render.Render(w, r, rest.BadRequest(err))
				return
			}
			kt.AppID = appID
			kt.SpaceID = space.SpaceId
			kt.SpaceTypeID = space.SpaceTypeId
		}

		if tmplSpaceID > 0 {
			// NOTE: authenticate template space on iam

			kt.TmplSpaceID = tmplSpaceID
		}

		ctx := kit.WithKit(r.Context(), kt)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
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
		space, err := a.authClient.QuerySpaceByAppID(kt.RpcCtx(), &pbas.QuerySpaceByAppIDReq{AppId: uint32(appID)})
		if err != nil {
			render.Render(w, r, rest.GRPCErr(err))
			return
		}

		kt.AppID = uint32(appID)
		kt.SpaceID = space.SpaceId
		kt.SpaceTypeID = space.SpaceTypeId
		ctx := kit.WithKit(r.Context(), kt)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func getBizID(r *http.Request) string {
	bizIDStr := chi.URLParam(r, "biz_id")
	if bizIDStr != "" {
		return bizIDStr
	}

	parts := strings.Split(r.URL.Path, "/")
	for idx, v := range parts {
		if v == "biz_id" && len(parts) > idx+1 {
			return parts[idx+1]
		}
		if v == "biz" && len(parts) > idx+1 {
			return parts[idx+1]
		}
	}

	return ""
}

// BizVerified 业务ID鉴权, url必须满足/{biz_id}; /biz_id/{n} 或者 /biz/{n}
func (a authorizer) BizVerified(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		kt := kit.MustGetKit(r.Context())
		bizIDStr := getBizID(r)
		if bizIDStr == "" {
			err := errors.New("biz id is required in url params")
			render.Render(w, r, rest.BadRequest(err))
			return
		}

		// 设置语言
		lang := tools.GetLangFromReq(r)
		kt.Lang = lang

		bizID, err := strconv.Atoi(bizIDStr)
		if err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}
		kt.BizID = uint32(bizID)

		if !a.HasBiz(uint32(bizID)) {
			err := fmt.Errorf("biz id %d does not exist", bizID)
			render.Render(w, r, rest.BadRequest(err))
			return
		}

		// skip validate biz permission when user is for test
		if strings.HasPrefix(kt.User, constant.BKUserForTestPrefix) {
			ctx := kit.WithKit(r.Context(), kt)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		kt.OperateWay = r.Header.Get(constant.OperateWayKey)
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
			AppID:       0,
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
