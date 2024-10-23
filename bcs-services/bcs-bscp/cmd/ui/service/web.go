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

// Package service provides functionality for displaying the UI interface.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"k8s.io/klog/v2"

	bscp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/docs"
	_ "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/docs" // 文档自动注册到 swagger
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/config"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/handler"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// WebServer :
type WebServer struct {
	ctx               context.Context
	srv               *http.Server
	addr              string
	addrs             []string
	embedWebServer    bscp.EmbedWebServer
	discover          serviced.Discover
	state             serviced.State
	authorizer        auth.Authorizer
	webAuthentication func(next http.Handler) http.Handler
}

// NewWebServer :
func NewWebServer(ctx context.Context, addr string, addrs []string) (*WebServer, error) {
	etcdOpt, err := config.G.EtcdConf()
	if err != nil {
		return nil, fmt.Errorf("get etcd config failed, err: %v", err)
	}

	// new discovery client.
	dis, err := serviced.NewDiscovery(*etcdOpt)
	if err != nil {
		return nil, fmt.Errorf("new discovery faield, err: %v", err)
	}

	state, ok := dis.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}

	// 鉴权器
	authorizer, err := auth.NewAuthorizer(dis, cc.TLSConfig{})
	if err != nil {
		return nil, fmt.Errorf("new authorizer failed, err: %v", err)
	}

	// 鉴权中间件
	webAuthentication := authorizer.WebAuthentication(config.G.Web.Host)

	s := &WebServer{
		ctx:               ctx,
		addr:              addr,
		addrs:             addrs,
		discover:          dis,
		state:             state,
		embedWebServer:    bscp.NewEmbedWeb(),
		authorizer:        authorizer,
		webAuthentication: webAuthentication,
	}

	srv := &http.Server{Addr: addr, Handler: s.newRouter()}
	s.srv = srv

	return s, nil
}

// Run :
func (s *WebServer) Run() error {
	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(s.addr); err != nil {
		return err
	}
	klog.Infof("http server listen address: %s", s.addr)

	for _, a := range s.addrs {
		if a == s.addr {
			continue
		}
		if err := dualStackListener.AddListenerWithAddr(a); err != nil {
			return err
		}
		klog.Infof("http serve listener with addr: %s", a)
	}

	return s.srv.Serve(dualStackListener)
}

// Close :
func (s *WebServer) Close() error {
	return s.srv.Shutdown(s.ctx)
}

func (s *WebServer) newRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// 注册 HTTP 请求
	r.Get("/-/healthy", s.HealthyHandler)
	r.Get("/-/ready", s.ReadyHandler)
	r.Get("/healthz", s.HealthzHandler)

	// init metrics
	metrics.InitMetrics(s.addr)
	metrics.RegisterHTTPMetrics()
	r.Get("/metrics", metrics.Handler().ServeHTTP)

	if config.G.Web.RoutePrefix != "/" && config.G.Web.RoutePrefix != "" {
		r.Mount(config.G.Web.RoutePrefix, http.StripPrefix(config.G.Web.RoutePrefix, s.subRouter()))
	}

	r.With(s.webAuthentication).Get(config.G.Web.RoutePrefix+"/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		ext := filepath.Ext(r.URL.Path)
		if ext == ".json" {
			w.Header().Set("Content-Type", "application/json")
			file, _ := docs.Assets.ReadFile("swagger/api.swagger.json")
			w.Write(file)
			return
		}
		httpSwagger.Handler(
			httpSwagger.UIConfig(map[string]string{
				"showExtensions": "true", // 显示扩展
			}),
			httpSwagger.URL("api.swagger.json"),
		).ServeHTTP(w, r)
	})
	r.Mount("/", s.subRouter())

	return r
}

// subRouter xxx
// @Title     BSCP-UI OpenAPI
// @BasePath  /bscp
func (s *WebServer) subRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/favicon.ico", s.embedWebServer.FaviconHandler)
	r.Get("/web/*", s.embedWebServer.StaticFileHandler("/web").ServeHTTP)

	shouldProxyAPI := config.G.IsDevMode()
	conf := &bscp.IndexConfig{
		StaticURL:            path.Join(config.G.Web.RoutePrefix, "/web"),
		RunEnv:               config.G.Base.RunEnv,
		ProxyAPI:             shouldProxyAPI,
		SiteURL:              config.G.Web.RoutePrefix,
		APIURL:               config.G.Frontend.Host.BSCPAPIURL,
		IAMHost:              config.G.Frontend.Host.BKIAMHost,
		CMDBHost:             config.G.Frontend.Host.BKCMDBHost,
		BKSharedResBaseJSURL: config.G.Frontend.Host.BKSharedResBaseJSURL,
		EnableBKNotice:       config.G.Frontend.EnableBKNotice,
		Helper:               config.G.Frontend.Helper,
		GrpcAddr:             config.G.Base.GrpcAddr,
		HttpAddr:             config.G.Base.HttpAddr,
		NodeManHost:          config.G.Frontend.Host.BKNODEMANHOST,
		UserManHost:          config.G.Frontend.Host.UserManHost,
	}

	if shouldProxyAPI {
		r.Mount("/bscp", handler.ReverseProxyHandler("bscp_api", "", config.G.Web.Host))
	}

	r.With(metrics.RequestCollect("no_permission")).Get("/403.html",
		s.embedWebServer.Render403Handler(conf).ServeHTTP)

	// vue 模版渲染
	r.With(metrics.RequestCollect("index"),
		s.webAuthentication).Get("/", s.embedWebServer.RenderIndexHandler(conf).ServeHTTP)
	r.With(metrics.RequestCollect("index"),
		s.webAuthentication).NotFound(s.embedWebServer.RenderIndexHandler(conf).ServeHTTP)

	return r
}

// HealthzHandler Healthz 接口
// @Summary  Healthz 接口
// @Tags     Healthz
// @Success  200  {string}  string
// @Router   /healthz [get]
func (s *WebServer) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.state.Healthz(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("OK"))
}

// HealthyHandler 健康检查
func (s *WebServer) HealthyHandler(w http.ResponseWriter, r *http.Request) {
	s.HealthzHandler(w, r)
}

// ReadyHandler 健康检查
func (s *WebServer) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
