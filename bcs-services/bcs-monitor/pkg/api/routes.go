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
	"encoding/json"
	"net/http"
	"path"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	ginTracing "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/route"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/docs" // docs xxx
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/logrule"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/pod"
	podmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/pod_monitor"
	service_monitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/servicemonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest/tracing"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// APIServer :
type APIServer struct { // nolint
	ctx      context.Context
	engine   *gin.Engine
	srv      *http.Server
	addr     string
	port     string
	addrIPv6 string
}

// NewAPIServer :
func NewAPIServer(ctx context.Context, addr, port, addrIPv6 string) (*APIServer, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	srv := &http.Server{Addr: addr, Handler: engine}

	s := &APIServer{
		ctx:      ctx,
		engine:   engine,
		srv:      srv,
		addr:     addr,
		port:     port,
		addrIPv6: addrIPv6,
	}
	s.newRoutes(engine)

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
// @Title    BCS-Monitor OpenAPI
// @BasePath /bcsapi/v4/monitor/api/projects/:projectId/clusters/:clusterId
func (a *APIServer) newRoutes(engine *gin.Engine) {
	// 添加 X-Request-Id 头部
	requestIdMiddleware := requestid.New(
		requestid.WithGenerator(tracing.RequestIdGenerator),
	)

	engine.Use(requestIdMiddleware, cors.Default())

	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	engine.GET("/-/healthy", HealthyHandler)
	engine.GET("/-/ready", ReadyHandler)

	// 注册 HTTP 请求
	registerRoutes(engine.Group(""))
	registerMetricsRoutes(engine.Group(""))

	if config.G.Web.RoutePrefix != "" {
		registerRoutes(engine.Group(config.G.Web.RoutePrefix))
		registerMetricsRoutes(engine.Group(config.G.Web.RoutePrefix))
	}
	registerRoutes(engine.Group(path.Join(config.G.Web.RoutePrefix, config.APIServicePrefix)))
	registerMetricsRoutes(engine.Group(path.Join(config.G.Web.RoutePrefix, config.APIServicePrefix)))
}

func registerRoutes(engine *gin.RouterGroup) {
	// 日志相关接口
	engine.Use(middleware.AuthenticationRequired(), middleware.ProjectParse(), middleware.NsScopeAuthorization())
	engine.Use(ginTracing.Middleware("bcs-monitor-api"))

	route := engine.Group("/projects/:projectId/clusters/:clusterId")
	{
		route.GET("/namespaces/:namespace/pods/:pod/containers", rest.RestHandlerFunc(pod.GetPodContainers))
		route.GET("/namespaces/:namespace/pods/:pod/logs", rest.RestHandlerFunc(pod.GetPodLog))
		route.GET("/namespaces/:namespace/pods/:pod/logs/download", rest.StreamHandler(pod.DownloadPodLog))

		// sse 实时日志流
		route.GET("/namespaces/:namespace/pods/:pod/logs/stream", rest.StreamHandler(pod.PodLogStream))

		// bk-log 日志采集规则
		route.POST("/log_collector/entrypoints", rest.RestHandlerFunc(logrule.GetEntrypoints))
		route.GET("/log_collector/rules", rest.RestHandlerFunc(logrule.ListLogCollectors))
		route.POST("/log_collector/rules", rest.RestHandlerFunc(logrule.CreateLogRule))
		route.GET("/log_collector/rules/:id", rest.RestHandlerFunc(logrule.GetLogRule))
		route.PUT("/log_collector/rules/:id", rest.RestHandlerFunc(logrule.UpdateLogRule))
		route.DELETE("/log_collector/rules/:id", rest.RestHandlerFunc(logrule.DeleteLogRule))
		route.POST("/log_collector/rules/:id/retry", rest.RestHandlerFunc(logrule.RetryLogRule))
		route.POST("/log_collector/rules/:id/enable", rest.RestHandlerFunc(logrule.EnableLogRule))
		route.POST("/log_collector/rules/:id/disable", rest.RestHandlerFunc(logrule.DisableLogRule))
		route.GET("/log_collector/storages/cluster_groups", rest.RestHandlerFunc(logrule.GetStorageClusters))
		route.POST("/log_collector/storages/switch_storage", rest.RestHandlerFunc(logrule.SwitchStorage))
	}
}

// registerMetricsRoutes metrics 相关接口
func registerMetricsRoutes(engine *gin.RouterGroup) {

	engine.Use(middleware.AuthenticationRequired(), middleware.ProjectParse(), middleware.ProjectAuthorization())
	engine.Use(ginTracing.Middleware("bcs-monitor-api"))

	// 命名规范
	// usage 代表 百分比
	// used 代表已使用
	// overview, info 数值量

	route := engine.Group("/metrics/projects/:projectCode/clusters/:clusterId")
	{
		route.GET("/overview", rest.RestHandlerFunc(metrics.GetClusterOverview))
		route.GET("/cpu_usage", rest.RestHandlerFunc(metrics.ClusterCPUUsage))
		route.GET("/cpu_request_usage", rest.RestHandlerFunc(metrics.ClusterCPURequestUsage))
		route.GET("/memory_usage", rest.RestHandlerFunc(metrics.ClusterMemoryUsage))
		route.GET("/memory_request_usage", rest.RestHandlerFunc(metrics.ClusterMemoryRequestUsage))
		route.GET("/disk_usage", rest.RestHandlerFunc(metrics.ClusterDiskUsage))
		route.GET("/diskio_usage", rest.RestHandlerFunc(metrics.ClusterDiskioUsage))
		route.GET("/pod_usage", rest.RestHandlerFunc(metrics.ClusterPodUsage))
		route.GET("/nodegroup/:nodegroup/node_num", rest.RestHandlerFunc(metrics.ClusterGroupNodeNum))
		route.GET("/nodegroup/:nodegroup/max_node_num", rest.RestHandlerFunc(metrics.ClusterGroupMaxNodeNum))
		route.GET("/nodes/:node/info", rest.RestHandlerFunc(metrics.GetNodeInfo))
		route.GET("/nodes/:node/overview", rest.RestHandlerFunc(metrics.GetNodeOverview))
		route.POST("/nodes/overviews", rest.RestHandlerFunc(metrics.ListNodeOverviews))
		route.GET("/nodes/:node/cpu_usage", rest.RestHandlerFunc(metrics.GetNodeCPUUsage))
		route.GET("/nodes/:node/cpu_request_usage", rest.RestHandlerFunc(metrics.GetNodeCPURequestUsage))
		route.GET("/nodes/:node/memory_usage", rest.RestHandlerFunc(metrics.GetNodeMemoryUsage))
		route.GET("/nodes/:node/memory_request_usage", rest.RestHandlerFunc(metrics.GetNodeMemoryRequestUsage))
		route.GET("/nodes/:node/network_receive", rest.RestHandlerFunc(metrics.GetNodeNetworkReceiveUsage))
		route.GET("/nodes/:node/network_transmit", rest.RestHandlerFunc(metrics.GetNodeNetworkTransmitUsage))
		route.GET("/nodes/:node/disk_usage", rest.RestHandlerFunc(metrics.GetNodeDiskUsage))
		route.GET("/nodes/:node/diskio_usage", rest.RestHandlerFunc(metrics.GetNodeDiskioUsage))
		route.POST("/namespaces/:namespace/pods/cpu_usage", rest.RestHandlerFunc(
			metrics.PodCPUUsage)) // 多个Pod场景, 可能有几十，上百Pod场景, 需要使用 Post 传递参数
		route.POST("/namespaces/:namespace/pods/cpu_limit_usage", rest.RestHandlerFunc(metrics.PodCPULimitUsage))
		route.POST("/namespaces/:namespace/pods/cpu_request_usage", rest.RestHandlerFunc(metrics.PodCPURequestUsage))
		route.POST("/namespaces/:namespace/pods/memory_used", rest.RestHandlerFunc(metrics.PodMemoryUsed))
		route.POST("/namespaces/:namespace/pods/network_receive", rest.RestHandlerFunc(metrics.PodNetworkReceive))
		route.POST("/namespaces/:namespace/pods/network_transmit", rest.RestHandlerFunc(metrics.PodNetworkTransmit))
		route.GET("/namespaces/:namespace/pods/:pod/containers/:container/cpu_usage",
			rest.RestHandlerFunc(metrics.ContainerCPUUsage))
		route.GET("/namespaces/:namespace/pods/:pod/containers/:container/memory_used",
			rest.RestHandlerFunc(metrics.ContainerMemoryUsed))
		route.GET("/namespaces/:namespace/pods/:pod/containers/:container/cpu_limit",
			rest.RestHandlerFunc(metrics.ContainerCPULimit))
		route.GET("/namespaces/:namespace/pods/:pod/containers/:container/memory_limit",
			rest.RestHandlerFunc(metrics.ContainerMemoryLimit))
		route.GET("/namespaces/:namespace/pods/:pod/containers/:container/disk_read_total",
			rest.RestHandlerFunc(metrics.ContainerDiskReadTotal))
		route.GET("/namespaces/:namespace/pods/:pod/containers/:container/disk_write_total",
			rest.RestHandlerFunc(metrics.ContainerDiskWriteTotal))

		route.GET("/namespaces/:namespace/service_monitors",
			rest.RestHandlerFunc(service_monitor.ListServiceMonitors))
		route.GET("/namespaces/:namespace/service_monitors/:name",
			rest.RestHandlerFunc(service_monitor.GetServiceMonitor))
		route.POST("/namespaces/:namespace/service_monitors",
			rest.RestHandlerFunc(service_monitor.CreateServiceMonitor))
		route.PUT("/namespaces/:namespace/service_monitors/:name",
			rest.RestHandlerFunc(service_monitor.UpdateServiceMonitor))
		route.DELETE("/namespaces/:namespace/service_monitors/:name",
			rest.RestHandlerFunc(service_monitor.DeleteServiceMonitor))
		route.GET("/service_monitors",
			rest.RestHandlerFunc(service_monitor.ListServiceMonitors))
		route.POST("/service_monitors/batchdelete",
			rest.RestHandlerFunc(service_monitor.BatchDeleteServiceMonitor))

		route.GET("/namespaces/:namespace/pod_monitors",
			rest.RestHandlerFunc(podmonitor.ListPodMonitors))
		route.GET("/namespaces/:namespace/pod_monitors/:name",
			rest.RestHandlerFunc(podmonitor.GetPodMonitor))
		route.POST("/namespaces/:namespace/pod_monitors",
			rest.RestHandlerFunc(podmonitor.CreatePodMonitor))
		route.PUT("/namespaces/:namespace/pod_monitors/:name",
			rest.RestHandlerFunc(podmonitor.UpdatePodMonitor))
		route.DELETE("/namespaces/:namespace/pod_monitors/:name",
			rest.RestHandlerFunc(podmonitor.DeletePodMonitor))
		route.GET("/pod_monitors",
			rest.RestHandlerFunc(podmonitor.ListPodMonitors))
		route.POST("/pod_monitors/batchdelete",
			rest.RestHandlerFunc(podmonitor.BatchDeletePodMonitor))
	}
}

// RegisterStoreGWRoutes 注册storegw http-sd
func RegisterStoreGWRoutes(gw *storegw.StoreGW) *route.Router {
	router := route.New()
	router.Get("/api/discovery/targetgroups", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(gw.TargetGroups())
	})

	return router
}

// HealthyHandler 健康检查
func HealthyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}

// ReadyHandler 健康检查
func ReadyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}
