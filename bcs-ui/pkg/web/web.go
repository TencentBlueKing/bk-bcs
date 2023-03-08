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

package web

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	bcsui "github.com/Tencent/bk-bcs/bcs-ui"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
)

// WebServer :
type WebServer struct {
	ctx            context.Context
	srv            *http.Server
	addrIPv6       string
	embedWebServer bcsui.EmbedWebServer
}

// NewWebServer :
func NewWebServer(ctx context.Context, addr string, addrIPv6 string) (*WebServer, error) {

	s := &WebServer{
		ctx:            ctx,
		addrIPv6:       addrIPv6,
		embedWebServer: bcsui.NewEmbedWeb(),
	}
	srv := &http.Server{Addr: addr, Handler: s.newRouter()}
	s.srv = srv

	return s, nil
}

// Run :
func (a *WebServer) Run() error {
	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(a.srv.Addr); err != nil {
		return err
	}

	if a.addrIPv6 != "" {
		if err := dualStackListener.AddListenerWithAddr(a.addrIPv6); err != nil {
			return err
		}
		klog.Infof("api serve dualStackListener with ipv6: %s", a.addrIPv6)
	}

	return a.srv.Serve(dualStackListener)
}

// Close :
func (a *WebServer) Close() error {
	return a.srv.Shutdown(a.ctx)
}

// newRoutes xxx
// @Title     BCS-Monitor OpenAPI
// @BasePath  /bcsapi/v4/monitor/api/projects/:projectId/clusters/:clusterId
func (w *WebServer) newRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(CORS)

	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Get("/healthz", HealthzHandler)
	r.Get("/-/healthy", HealthyHandler)
	r.Get("/-/ready", ReadyHandler)

	// 注册 HTTP 请求
	// 正确路由
	r.Get("/bcs", w.embedWebServer.IndexHandler().ServeHTTP)
	r.Get("/bcs/*", w.embedWebServer.IndexHandler().ServeHTTP)

	if config.G.IsDevMode() {
		r.Mount("/backend", ReverseAPIHandler("bcs_saas_api_url", config.G.FrontendConf.Host.DevOpsBCSAPIURL))
		r.Mount("/bcsapi", ReverseAPIHandler("bcs_host", config.G.BCS.Host))
	}

	// vue 自定义路由, 前端返回404
	r.NotFound(w.embedWebServer.IndexHandler().ServeHTTP)

	return r
}

// ReverseAPIHandler api代理
func ReverseAPIHandler(name, remoteURL string) http.Handler {
	remote, err := url.Parse(remoteURL)
	if err != nil {
		panic(fmt.Errorf("%s '%s' not valid: %s", name, remoteURL, err))
	}

	if remote.Scheme != "http" && remote.Scheme != "https" {
		panic(fmt.Errorf("%s '%s' scheme not supported", name, remoteURL))
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = r.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			klog.InfoS("forward request", "name", name, "url", req.URL)
		}

		proxy.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// HealthyHandler 健康检查
func HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ReadyHandler 健康检查
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// HealthzHandler 健康检查
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// CORS 跨域
func CORS(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// cors 处理
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		allowHeaders := []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"X-Requested-With",
		}
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ","))

		allowMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ","))

		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
