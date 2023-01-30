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
	"net/http"
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
	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger(), cors.Default())

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
	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	engine.GET("/-/healthy", HealthyHandler)
	engine.GET("/-/ready", ReadyHandler)

	// 注册 HTTP 请求

	// 注册模板和静态资源
	engine.SetHTMLTemplate(bscp.WebTemplate())
	webFaviconPath := bscp.WebFaviconPath()
	rootGZipHandler := StaticGZipHandler("/web/", http.FS(bscp.WebStatic()))

	engine.Group("", StaticCacheHandler).StaticFileFS("/favicon.ico", webFaviconPath, http.FS(bscp.WebStatic()))
	// engine.Group("", StaticCacheHandler).StaticFS("/web", http.FS(bscp.WebStatic()))
	engine.Group("", StaticCacheHandler).HEAD("/web/*filepath", rootGZipHandler)
	engine.Group("", StaticCacheHandler).GET("/web/*filepath", rootGZipHandler)
	engine.GET("", w.IndexHandler)

	if config.G.Web.RoutePrefix != "" {
		prefixGZipHandler := StaticGZipHandler(path.Join(config.G.Web.RoutePrefix, "/web/"), http.FS(bscp.WebStatic()))
		engine.Group(config.G.Web.RoutePrefix, StaticCacheHandler).StaticFileFS("/favicon.ico", webFaviconPath, http.FS(bscp.WebStatic()))
		// engine.Group(config.G.Web.RoutePrefix, StaticCacheHandler).StaticFS("/web", http.FS(bscp.WebStatic()))
		engine.Group(config.G.Web.RoutePrefix, StaticCacheHandler).HEAD("/web/*filepath", prefixGZipHandler)
		engine.Group(config.G.Web.RoutePrefix, StaticCacheHandler).GET("/web/*filepath", prefixGZipHandler)
		engine.Group(config.G.Web.RoutePrefix).GET("", w.IndexHandler)
	}

	// 本地开发模式
	if config.G.IsDevMode() {
		engine.Any("/bscp/api/*path", ReverseAPIHandler("bscp_api", config.G.BCS.Host))
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

	// 本地开发模式
	if config.G.IsDevMode() {
		data["BK_BCS_BSCP_API"] = "/bscp"
	}

	c.HTML(http.StatusOK, "index.html", data)
}
