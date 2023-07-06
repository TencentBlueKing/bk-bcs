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
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	bcsui "github.com/Tencent/bk-bcs/bcs-ui"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/tracing"
)

// WebServer :
type WebServer struct {
	ctx            context.Context
	srv            *http.Server
	addrIPv6       string
	embedWebServer bcsui.EmbedWebServer
	releaseNote    ReleaseNoteLang
}

// NewWebServer :
func NewWebServer(ctx context.Context, addr string, addrIPv6 string) (*WebServer, error) {
	s := &WebServer{
		ctx:            ctx,
		addrIPv6:       addrIPv6,
		embedWebServer: bcsui.NewEmbedWeb(),
	}

	// 初始化版本日志和特性
	if err := s.initReleaseNote(); err != nil {
		return nil, err
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

	if a.addrIPv6 != "" && a.addrIPv6 != a.srv.Addr {
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
func (w *WebServer) newRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Get("/healthz", HealthzHandler)
	r.Get("/-/healthy", HealthyHandler)
	r.Get("/-/ready", ReadyHandler)

	// metrics 配置
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	if config.G.IsLocalDevMode() {
		r.Mount("/backend", ReverseAPIHandler("bcs_saas_api_url", config.G.FrontendConf.Host.DevOpsBCSAPIURL))
		r.Mount("/bcsapi", ReverseAPIHandler("bcs_host", config.G.BCS.Host))
	}

	r.Mount("/", w.subRouter())

	return r
}

func (w *WebServer) subRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(tracing.MiddleWareTracing)

	r.Get("/favicon.ico", w.embedWebServer.FaviconHandler)

	r.Route("/release_note", func(r chi.Router) {
		// 单独使用metrics中间件的方式收集请求量、耗时
		r.Use(metrics.RequestCollect("ReleaseNoteHandler"))
		r.Get("/", w.ReleaseNoteHandler)
	})

	r.Get("/web/*", w.embedWebServer.StaticFileHandler("/web").ServeHTTP)

	// vue 模版渲染
	r.Route("/", func(r chi.Router) {
		r.Use(metrics.RequestCollect("IndexHandler"))
		r.Get("/", w.embedWebServer.IndexHandler().ServeHTTP)
	})
	r.NotFound(w.embedWebServer.IndexHandler().ServeHTTP)

	r.Put("/switch_language", w.CookieSwitchLanguage)
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

// OKResponse
type OKResponse struct {
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
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
