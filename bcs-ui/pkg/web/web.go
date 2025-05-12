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

// Package web xxx
package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"k8s.io/klog/v2"

	bcsui "github.com/Tencent/bk-bcs/bcs-ui"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/notice"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/middleware"
)

// WebServer :
type WebServer struct { // nolint
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

	// 注册系统到通知中心
	if config.G.BKNotice.Enable {
		if err := notice.RegisterSystem(ctx); err != nil {
			klog.Infof("register system to notice center failed: %s", err.Error())
		}
	}

	srv := &http.Server{Addr: addr, Handler: s.newRouter()}
	s.srv = srv

	return s, nil
}

// Run :
func (w *WebServer) Run() error {
	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(w.srv.Addr); err != nil {
		return err
	}

	if w.addrIPv6 != "" && w.addrIPv6 != w.srv.Addr {
		if err := dualStackListener.AddListenerWithAddr(w.addrIPv6); err != nil {
			return err
		}
		klog.Infof("api serve dualStackListener with ipv6: %s", w.addrIPv6)
	}

	return w.srv.Serve(dualStackListener)
}

// Close :
func (w *WebServer) Close() error {
	return w.srv.Shutdown(w.ctx)
}

// newRoutes xxx
func (w *WebServer) newRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(chimid.Logger)
	r.Use(chimid.Recoverer)

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

	// 注册到网关的地址, 默认/bcsapi/v4/ui
	routePrefix := config.G.Web.RoutePrefix
	if routePrefix != "" && routePrefix != "/" {
		r.Mount(routePrefix+"/", http.StripPrefix(routePrefix, w.subRouter()))
	}

	r.Mount("/", w.subRouter())

	return r
}

func (w *WebServer) subRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Tracing)
	r.Use(middleware.TenantHandler)

	r.Get("/favicon.ico", w.embedWebServer.FaviconHandler)

	r.With(metrics.RequestCollect("FeatureFlagsHandler"), middleware.NeedProjectAuthorization).
		Get("/feature_flags", w.FeatureFlagsHandler)

	r.With(metrics.RequestCollect("GetCurrentAnnouncements")).Get("/announcements", w.GetCurrentAnnouncements)

	// ai
	r.With(metrics.RequestCollect("Assistant"), middleware.Authentication, middleware.BKTicket).
		Post("/assistant", w.Assistant)

	// 静态资源
	r.Get("/web/*", w.embedWebServer.StaticFileHandler("/web").ServeHTTP)
	r.With(metrics.RequestCollect("ReleaseNoteHandler")).Get("/release_note", w.ReleaseNoteHandler)
	r.With(metrics.RequestCollect("ReportHandler")).Post("/report", ReportHandler)

	r.With(metrics.RequestCollect("no_permission")).Get("/403.html", w.embedWebServer.Render403Handler().ServeHTTP)

	// vue 模版渲染
	r.With(metrics.RequestCollect("IndexHandler")).Get("/", w.embedWebServer.IndexHandler().ServeHTTP)
	r.With(metrics.RequestCollect("IndexHandler")).NotFound(w.embedWebServer.IndexHandler().ServeHTTP)

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

// OKResponse ok response
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

// ReportHandler 通过标准输出落地到日志采集，前端埋点接口
func ReportHandler(w http.ResponseWriter, r *http.Request) {
	okResponse := &OKResponse{Message: "OK", RequestID: r.Header.Get(constants.RequestIDHeaderKey)}
	// response message
	defer render.JSON(w, r, okResponse)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		// failure return
		okResponse.Code = http.StatusBadRequest
		okResponse.Message = err.Error()
		return
	}
	var stringBuffer bytes.Buffer
	err = json.Compact(&stringBuffer, b)
	if err != nil {
		// failure return
		okResponse.Code = http.StatusBadRequest
		okResponse.Message = err.Error()
		return
	}
	// 通过标准输出落地到日志采集
	fmt.Println(stringBuffer.String())
}
