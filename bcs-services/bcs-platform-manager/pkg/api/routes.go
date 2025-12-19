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

// Package api router
package api

import (
	"context"
	"net/http"
	"path"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/api/pod"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/rest/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/rest/tracing"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

// APIServer :
type APIServer struct { // nolint
	ctx      context.Context
	srv      *http.Server
	addr     string
	port     string
	addrIPv6 string
}

// NewAPIServer :
func NewAPIServer(ctx context.Context, addr, port, addrIPv6 string) (*APIServer, error) {

	s := &APIServer{
		ctx:      ctx,
		addr:     addr,
		port:     port,
		addrIPv6: addrIPv6,
	}
	srv := &http.Server{Addr: addr, Handler: s.newRoutes()}
	s.srv = srv
	return s, nil
}

// Run :
func (a *APIServer) Run() error {
	dualStackListener := listener.NewDualStackListener()
	addr := utils.GetListenAddr(a.addr, a.port)
	if err := dualStackListener.AddListenerWithAddr(utils.GetListenAddr(a.addr, a.port)); err != nil {
		return err
	}
	blog.Infow("listening for requests and metrics", "address", addr)

	if a.addrIPv6 != "" && a.addrIPv6 != a.addr {
		v6Addr := utils.GetListenAddr(a.addrIPv6, a.port)
		if err := dualStackListener.AddListenerWithAddr(v6Addr); err != nil {
			return err
		}
		blog.Infof("api serve dualStackListener with ipv6: %s", v6Addr)
	}

	return a.srv.Serve(dualStackListener)
}

// Close :
func (a *APIServer) Close() error {
	return a.srv.Shutdown(a.ctx)
}

// newRoutes xxx
// @Title    BCS-Platform-Manager OpenAPI
// @BasePath /bcsapi/v4/platform-manager/api/projects/{projectId}/clusters/{clusterId}
func (a *APIServer) newRoutes() http.Handler {
	r := chi.NewRouter()

	// 添加 X-Request-Id 头部
	r.Use(tracing.RequestIdMiddleware)

	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	r.Get("/-/healthy", HealthyHandler)
	r.Get("/-/ready", ReadyHandler)

	// 注册 HTTP 请求
	r.Mount("/", registerRoutes())

	// 注册到网关的地址
	routePrefix := config.G.Web.RoutePrefix
	if routePrefix != "" && routePrefix != "/" {
		r.Mount(routePrefix+"/", http.StripPrefix(routePrefix, registerRoutes()))
	}
	webApiPrefix := path.Join(config.G.Web.RoutePrefix, config.APIServicePrefix)
	r.Mount(webApiPrefix+"/", http.StripPrefix(webApiPrefix, registerRoutes()))
	return r
}

func registerRoutes() http.Handler {
	r := chi.NewRouter()
	// 日志相关接口

	r.Route("/projects/{projectId}/clusters/{clusterId}", func(route chi.Router) {
		route.Use(middleware.AuthenticationRequired, middleware.ProjectParse, middleware.ClusterAuthorization)
		route.Use(middleware.VisitorsRequired, middleware.Tracing, middleware.Audit)

		route.Get("/containers", rest.Handle(pod.GetPodContainers))
	})
	return r
}

// HealthyHandler 健康检查
func HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ReadyHandler 健康检查
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
