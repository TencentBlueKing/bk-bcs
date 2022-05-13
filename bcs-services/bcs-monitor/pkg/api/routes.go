/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-监控平台 (Blueking - Monitor) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package api

import (
	"context"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/pod"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest/middleware"
)

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
	registerRoutes(engine)

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

func registerRoutes(engine *gin.Engine) {
	// 添加X-Request-Id 头部
	requestIdMiddleware := requestid.New(
		requestid.WithGenerator(func() string {
			return rest.RequestIdGenerator()
		}),
	)

	engine.Use(requestIdMiddleware)
	engine.Use(middleware.AuthRequired())

	// 日志相关接口
	route := engine.Group("/projects/:projectId/clusters/:clusterId/namespaces/:namespace/pods/:pod")
	{
		route.GET("/containers", rest.RestHandlerFunc(pod.GetPodContainers))
		route.GET("/logs", rest.RestHandlerFunc(pod.GetPodLog))
		route.GET("/logs/download", rest.StreamHandler(pod.DownloadPodLog))

		// sse 实时日志流
		route.GET("/logs/stream", rest.StreamHandler(pod.PodLogStream))
	}
}
