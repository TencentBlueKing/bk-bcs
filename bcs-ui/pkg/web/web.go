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

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"k8s.io/klog/v2"

	bcsui "github.com/Tencent/bk-bcs/bcs-ui"
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
func (a *WebServer) newRoutes(engine *gin.Engine) {
	engine.Use(gin.Recovery(), gin.Logger(), cors.Default())

	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	engine.GET("/-/healthy", HealthyHandler)
	engine.GET("/-/ready", ReadyHandler)

	// 注册 HTTP 请求

	// 注册模板和静态资源
	engine.SetHTMLTemplate(bcsui.WebTemplate())

	// router.Group(routePrefix).StaticFS("/web/static", http.FS(web.WebStatic()))
	engine.Group("").StaticFS("/web/static", http.FS(bcsui.WebStatic()))

	// 正确路由
	engine.GET("/bcs", IndexHandler)
	engine.GET("/bcs/*path", IndexHandler)

	if config.G.IsDevMode() {
		engine.Any("/backend/*path", ReverseAPIHandler("bcs_saas_api_url", config.G.FrontendConf.Host.DevOpsBCSAPIURL))
		engine.Any("/bcsapi/*path", ReverseAPIHandler("bcs_host", config.G.BCS.Host))
	}

	// vue 自定义路由, 前端返回404
	engine.NoRoute(IndexHandler)
}

// IndexHandler Vue 模板渲染
func IndexHandler(c *gin.Context) {
	data := gin.H{
		"STATIC_URL":              STATIC_URL,
		"SITE_URL":                SITE_URL,
		"REGION":                  "ce",
		"RUN_ENV":                 config.G.Base.RunEnv,
		"PREFERRED_DOMAINS":       config.G.Web.PreferredDomains,
		"DEVOPS_HOST":             config.G.FrontendConf.Host.DevOpsHost,
		"DEVOPS_BCS_API_URL":      config.G.FrontendConf.Host.DevOpsBCSAPIURL,
		"DEVOPS_ARTIFACTORY_HOST": config.G.FrontendConf.Host.DevOpsArtifactoryHost,
		"BK_IAM_APP_URL":          config.G.FrontendConf.Host.BKIAMAppURL,
		"PAAS_HOST":               config.G.FrontendConf.Host.PaaSHost,
		"BKMONITOR_HOST":          config.G.FrontendConf.Host.BKMonitorHOst,
		"BCS_API_HOST":            config.G.BCS.Host,
		"BK_CC_HOST":              config.G.FrontendConf.Host.BKCMDBHost,
	}

	if config.G.IsDevMode() {
		data["DEVOPS_BCS_API_URL"] = fmt.Sprintf("%s/backend", config.G.Web.Host)
		data["BCS_API_HOST"] = config.G.Web.Host
	}
	c.HTML(http.StatusOK, "index.html", data)
}

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
