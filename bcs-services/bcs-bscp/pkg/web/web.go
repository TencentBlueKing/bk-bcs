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
	"path"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"k8s.io/klog/v2"

	bscp "bscp.io"
	"bscp.io/pkg/config"
)

// WebServer :
type WebServer struct {
	ctx      context.Context
	engine   *gin.Engine
	srv      *http.Server
	addrIPv6 string
}

// NewWebServer :
func NewWebServer(ctx context.Context, addr string, addrIPv6 string) (*WebServer, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	srv := &http.Server{Addr: addr, Handler: engine}

	s := &WebServer{
		ctx:      ctx,
		engine:   engine,
		srv:      srv,
		addrIPv6: addrIPv6,
	}
	s.newRoutes(engine)

	return s, nil
}

// Run :
func (w *WebServer) Run() error {
	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(w.srv.Addr); err != nil {
		return err
	}

	if w.addrIPv6 != "" {
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
// @Title     BCS-Monitor OpenAPI
// @BasePath  /bcsapi/v4/monitor/api/projects/:projectId/clusters/:clusterId
func (w *WebServer) newRoutes(engine *gin.Engine) {
	engine.Use(gin.Recovery(), gin.Logger(), cors.Default())

	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	engine.GET("/-/healthy", HealthyHandler)
	engine.GET("/-/ready", ReadyHandler)

	// 注册 HTTP 请求

	// 注册模板和静态资源
	engine.SetHTMLTemplate(bscp.WebTemplate())

	engine.Group("").StaticFS("/web", http.FS(bscp.WebStatic()))
	engine.GET("", w.IndexHandler)

	if config.G.Web.RoutePrefix != "" {
		engine.Group(config.G.Web.RoutePrefix).StaticFS("/web", http.FS(bscp.WebStatic()))
		engine.Group(config.G.Web.RoutePrefix).GET("", w.IndexHandler)
	}

	// vue 自定义路由, 前端返回404
	engine.NoRoute(w.IndexHandler)
}

// IndexHandler Vue 模板渲染
func (w *WebServer) IndexHandler(c *gin.Context) {
	data := gin.H{
		"BK_STATIC_URL":   path.Join(config.G.Web.RoutePrefix, "/web"),
		"RUN_ENV":         config.G.Base.RunEnv,
		"BK_BCS_BSCP_API": config.G.BCS.Host + "/bscp",
	}

	c.HTML(http.StatusOK, "index.html", data)
}

// ReverseAPIHandler 代理请求， CORS 跨域问题
func ReverseAPIHandler(name, remoteURL string) gin.HandlerFunc {
	remote, err := url.Parse(remoteURL)
	if err != nil {
		panic(err)
	}

	if remote.Scheme != "http" && remote.Scheme != "https" {
		panic(fmt.Errorf("%s '%s' scheme not supported", name, remoteURL))
	}

	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthyHandler 健康检查
func HealthyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}

// ReadyHandler 健康检查
func ReadyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}
