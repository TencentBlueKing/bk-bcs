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

package service

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/audit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/view"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/handler"
)

// routers return router config handler
// nolint: funlen
func (p *proxy) routers() http.Handler {
	r := chi.NewRouter()
	r.Use(handler.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handler.CORS)
	r.Use(audit.Audit)
	// r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/-/healthy", p.HealthyHandler)
	r.Get("/-/ready", p.ReadyHandler)
	r.Get("/healthz", p.Healthz)
	r.Mount("/", handler.RegisterCommonHandler())

	// iam 回调接口
	r.Route("/api/v1/auth/iam/find/resource", func(r chi.Router) {
		r.Use(handler.RequestBodyLogger())
		r.Use(auth.IAMVerified)
		r.Mount("/", p.authSvrMux)
	})

	// 用户信息
	r.With(p.authorizer.UnifiedAuthentication).Get("/api/v1/auth/user/info", UserInfoHandler)
	r.With(p.authorizer.UnifiedAuthentication).Get("/api/v1/feature_flags", FeatureFlagsHandler)
	// 登入接口, 不带鉴权信息
	r.Get("/api/v1/logout", p.LogoutHandler)

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

	// 模版空间相关接口，检查当前业务下的默认空间，没有则创建
	r.Route("/api/v1/config/biz/{biz_id}/template_spaces", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(view.Generic(p.authorizer))
		r.Use(p.CheckDefaultTmplSpace)
		r.Mount("/", p.cfgSvrMux)
	})

	r.Route("/api/v1/config/", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(view.Generic(p.authorizer))
		r.Mount("/", p.cfgSvrMux)
	})

	// repo 上传 API, 此处因兼容老版本而保留，后续统一使用新接口
	r.Route("/api/v1/api/create/content/upload", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.With(p.authorizer.BizVerified, p.authorizer.AppVerified,
			p.HttpServerHandledTotal("", "Upload")).
			Put("/biz_id/{biz_id}/app_id/{app_id}",
				p.repo.UploadFile)

	})

	// repo 下载 API, 此处因兼容老版本而保留，后续统一使用新接口
	r.Route("/api/v1/api/get/content/download", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.With(p.authorizer.BizVerified, p.authorizer.AppVerified,
			p.HttpServerHandledTotal("", "Download")).
			Get("/biz_id/{biz_id}/app_id/{app_id}",
				p.repo.DownloadFile)
	})

	// repo 获取二进制元数据 API, 此处因兼容老版本而保留，后续统一使用新接口
	r.Route("/api/v1/api/get/content/metadata", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.With(p.authorizer.BizVerified, p.authorizer.AppVerified,
			p.HttpServerHandledTotal("", "Metadata")).
			Get("/biz_id/{biz_id}/app_id/{app_id}",
				p.repo.FileMetadata)
	})

	// 新的内容上传、下载相关接口，后续统一使用这组新接口
	// 服务下的配置项内容需要进行服务鉴权，模版空间下的模版配置项内容需要进行模版空间鉴权
	// app_id和template_space_id信息放在header中，用于鉴权，和sign保持一致
	r.Route("/api/v1/biz/{biz_id}/content", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(p.authorizer.ContentVerified)
		// 内容上传API
		r.Route("/upload", func(r chi.Router) {
			r.Use(p.HttpServerHandledTotal("", "Upload"))
			r.Put("/", p.repo.UploadFile)
		})
		// 内容下载API
		r.Route("/download", func(r chi.Router) {
			r.Use(p.HttpServerHandledTotal("", "Download"))
			r.Get("/", p.repo.DownloadFile)
		})
		// 获取二进制内容元数据API
		r.Route("/metadata", func(r chi.Router) {
			r.Use(p.HttpServerHandledTotal("", "Metadata"))
			r.Get("/", p.repo.FileMetadata)
		})
	})

	// 导入模板压缩包
	r.Route("/api/v1/config/biz/{biz_id}/template_spaces/{template_space_id}/templates/import", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(p.HttpServerHandledTotal("", "TemplateConfigFileImport"))
		r.Post("/", p.configImportService.TemplateConfigFileImport)

	})

	// 导入配置压缩包
	r.Route("/api/v1/config/biz/{biz_id}/apps/{app_id}/config_item/import", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Use(p.authorizer.BizVerified)
		r.Use(p.HttpServerHandledTotal("", "ConfigFileImport"))
		r.Post("/", p.configImportService.ConfigFileImport)
	})

	return r
}
