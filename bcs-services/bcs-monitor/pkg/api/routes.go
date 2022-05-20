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

package api

import (
	"context"
	"net/http"
	"path"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/docs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/pod"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest/middleware"
)

// 子模块前缀
const apiModulePrefix = "/api"

// APIServer
type APIServer struct {
	ctx    context.Context
	engine *gin.Engine
	srv    *http.Server
}

// NewAPIServer
func NewAPIServer(ctx context.Context, addr string) (*APIServer, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	srv := &http.Server{Addr: addr, Handler: engine}

	s := &APIServer{
		ctx:    ctx,
		engine: engine,
		srv:    srv,
	}
	s.newRoutes(engine)

	return s, nil
}

// Run
func (a *APIServer) Run() error {
	return a.srv.ListenAndServe()
}

// Close
func (a *APIServer) Close() error {
	return a.srv.Shutdown(a.ctx)
}

// @Title     BCS-Monitor OpenAPI
// @BasePath  /bcsapi/v4/monitor/api/projects/:projectId/clusters/:clusterId
func (a *APIServer) newRoutes(engine *gin.Engine) {
	// 添加 X-Request-Id 头部
	requestIdMiddleware := requestid.New(
		requestid.WithGenerator(func() string {
			return rest.RequestIdGenerator()
		}),
	)

	engine.Use(requestIdMiddleware, cors.Default())
	engine.Use(middleware.AuthRequired())

	// openapi 文档
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 注册 HTTP 请求
	registerRoutes(engine.Group(apiModulePrefix))
	registerRoutes(engine.Group(path.Join(config.G.Web.RoutePrefix, apiModulePrefix)))
}

func registerRoutes(engine *gin.RouterGroup) {
	// 日志相关接口
	route := engine.Group("/projects/:projectId/clusters/:clusterId")
	{
		route.GET("/namespaces/:namespace/pods/:pod/containers", rest.RestHandlerFunc(pod.GetPodContainers))
		route.GET("/namespaces/:namespace/pods/:pod/logs", rest.RestHandlerFunc(pod.GetPodLog))
		route.GET("/namespaces/:namespace/pods/:pod/logs/download", rest.StreamHandler(pod.DownloadPodLog))

		// sse 实时日志流
		route.GET("/namespaces/:namespace/pods/:pod/logs/stream", rest.StreamHandler(pod.PodLogStream))
	}
}
