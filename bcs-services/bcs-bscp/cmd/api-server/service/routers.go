/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/rest/view"
	"bscp.io/pkg/runtime/handler"
)

// routers return router config handler
func (p *proxy) routers() http.Handler {
	r := chi.NewRouter()
	r.Use(handler.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handler.CORS)
	// r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/-/healthy", p.HealthyHandler)
	r.Get("/-/ready", p.ReadyHandler)
	r.Get("/healthz", p.Healthz)
	r.Mount("/", handler.RegisterCommonHandler())

	// iam 回调接口
	r.Route("/api/v1/auth/iam/find/resource", func(r chi.Router) {
		r.Use(handler.RequestBodyLogger())
		r.Use(view.Generic(p.authorizer))
		r.Use(auth.IAMVerified)
		r.Mount("/", p.authSvrMux)
	})

	// 用户信息
	r.With(p.authorizer.UnifiedAuthentication).Get("/api/v1/auth/user/info", UserInfoHandler)

	// authserver通用接口
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(view.Generic(p.authorizer))
		r.Mount("/", p.authSvrMux)
	})

	r.Route("/api/v1/config/biz/{biz_id}/apps/query", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(view.Generic(p.authorizer))
		r.Mount("/", p.cfgSvrMux)
	})

	// 规范后的路由，url 需要包含 {app_id} 变量, 使用 AppVerified 中间件校验和初始化 kit.SpaceID 变量
	r.Route("/api/v1/config/biz/{biz_id}/apps/{app_id}", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(p.authorizer.AppVerified)
		r.Use(view.Generic(p.authorizer))
		r.Mount("/", p.cfgSvrMux)
	})

	r.Route("/api/v1/config/", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(view.Generic(p.authorizer))
		r.Mount("/", p.cfgSvrMux)
	})

	// repo 上传 API
	r.Route("/api/v1/api/create/content/upload", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.With(p.authorizer.BizVerified, p.authorizer.AppVerified).Put("/biz_id/{biz_id}/app_id/{app_id}", p.repoRevProxy.UploadFile)
	})

	// repo 下载 API
	r.Route("/api/v1/api/get/content/download", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.With(p.authorizer.BizVerified, p.authorizer.AppVerified).Get("/biz_id/{biz_id}/app_id/{app_id}", p.repoRevProxy.DownloadFile)
	})

	// repo 获取二进制元数据 API
	r.Route("/api/v1/api/get/content/metadata", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.With(p.authorizer.BizVerified, p.authorizer.AppVerified).Get("/biz_id/{biz_id}/app_id/{app_id}", p.repoRevProxy.FileMetadata)
	})

	return r
}
