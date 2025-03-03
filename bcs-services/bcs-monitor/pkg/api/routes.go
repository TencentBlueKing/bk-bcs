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
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/common/route"
	httpSwagger "github.com/swaggo/http-swagger"

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
	srv      *http.Server
	addr     string
	port     string
	addrIPv6 string
}

// NewAPIServer :
func NewAPIServer(ctx context.Context, addr, port, addrIPv6 string) (*APIServer, error) {

	s := &APIServer{
		ctx:      ctx,
		addr:     addr,
		port:     port,
		addrIPv6: addrIPv6,
	}
	srv := &http.Server{Addr: addr, Handler: s.newRoutes()}
	s.srv = srv
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
// @BasePath /bcsapi/v4/monitor/api/projects/{projectId}/clusters/{clusterId}
func (a *APIServer) newRoutes() http.Handler {
	r := chi.NewRouter()

	// 添加 X-Request-Id 头部
	r.Use(tracing.RequestIdMiddleware)

	// openapi 文档
	// 访问 swagger/index.html, swagger/doc.json
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	r.Get("/-/healthy", HealthyHandler)
	r.Get("/-/ready", ReadyHandler)

	// 注册 HTTP 请求
	r.Mount("/", registerRoutes())
	r.Mount("/metrics", registerMetricsRoutes())

	// 注册到网关的地址
	routePrefix := config.G.Web.RoutePrefix
	if routePrefix != "" && routePrefix != "/" {
		r.Mount(routePrefix+"/", http.StripPrefix(routePrefix, registerRoutes()))
		r.Mount(routePrefix+"/metrics", http.StripPrefix(routePrefix, registerMetricsRoutes()))
	}
	webApiPrefix := path.Join(config.G.Web.RoutePrefix, config.APIServicePrefix)
	r.Mount(webApiPrefix+"/", http.StripPrefix(webApiPrefix, registerRoutes()))
	r.Mount(webApiPrefix+"/metrics", http.StripPrefix(webApiPrefix, registerMetricsRoutes()))
	return r
}

func registerRoutes() http.Handler {
	r := chi.NewRouter()
	// 日志相关接口

	r.Route("/projects/{projectId}/clusters/{clusterId}", func(route chi.Router) {
		route.Use(middleware.AuthenticationRequired, middleware.ProjectParse, middleware.ClusterAuthorization)
		route.Use(middleware.Tracing, middleware.Audit)

		route.Get("/namespaces/{namespace}/pods/{pod}/containers", rest.Handle(pod.GetPodContainers))
		route.Get("/namespaces/{namespace}/pods/{pod}/logs", rest.Handle(pod.GetPodLog))
		route.Get("/namespaces/{namespace}/pods/{pod}/logs/download", rest.Stream(pod.DownloadPodLog))

		// sse 实时日志流
		route.Get("/namespaces/{namespace}/pods/{pod}/logs/stream", rest.Stream(pod.PodLogStream))

		// bk-log 日志采集规则
		route.Post("/log_collector/entrypoints", rest.Handle(logrule.GetEntrypoints))
		route.Get("/log_collector/rules", rest.Handle(logrule.ListLogCollectors))
		route.Post("/log_collector/rules", rest.Handle(logrule.CreateLogRule))
		route.Get("/log_collector/rules/{id}", rest.Handle(logrule.GetLogRule))
		route.Put("/log_collector/rules/{id}", rest.Handle(logrule.UpdateLogRule))
		route.Delete("/log_collector/rules/{id}", rest.Handle(logrule.DeleteLogRule))
		route.Post("/log_collector/rules/{id}/retry", rest.Handle(logrule.RetryLogRule))
		route.Post("/log_collector/rules/{id}/enable", rest.Handle(logrule.EnableLogRule))
		route.Post("/log_collector/rules/{id}/disable", rest.Handle(logrule.DisableLogRule))
		route.Get("/log_collector/storages/cluster_groups", rest.Handle(logrule.GetStorageClusters))
		route.Post("/log_collector/storages/switch_storage", rest.Handle(logrule.SwitchStorage))
	})
	return r
}

// registerMetricsRoutes metrics 相关接口
func registerMetricsRoutes() http.Handler {
	r := chi.NewRouter()
	// 命名规范
	// usage 代表 百分比
	// used 代表已使用
	// overview, info 数值量

	r.Route("/projects/{projectCode}/clusters/{clusterId}", func(route chi.Router) {
		route.Use(middleware.AuthenticationRequired, middleware.ProjectParse, middleware.ProjectAuthorization)
		route.Use(middleware.Tracing)

		route.Get("/overview", rest.Handle(metrics.GetClusterOverview))
		route.Get("/cpu_usage", rest.Handle(metrics.ClusterCPUUsage))
		route.Get("/cpu_request_usage", rest.Handle(metrics.ClusterCPURequestUsage))
		route.Get("/memory_usage", rest.Handle(metrics.ClusterMemoryUsage))
		route.Get("/memory_request_usage", rest.Handle(metrics.ClusterMemoryRequestUsage))
		route.Get("/disk_usage", rest.Handle(metrics.ClusterDiskUsage))
		route.Get("/diskio_usage", rest.Handle(metrics.ClusterDiskioUsage))
		route.Get("/pod_usage", rest.Handle(metrics.ClusterPodUsage))
		route.Get("/nodegroup/{nodegroup}/node_num", rest.Handle(metrics.ClusterGroupNodeNum))
		route.Get("/nodegroup/{nodegroup}/max_node_num", rest.Handle(metrics.ClusterGroupMaxNodeNum))
		route.Get("/nodes/{node}/info", rest.Handle(metrics.GetNodeInfo))
		route.Get("/nodes/{node}/overview", rest.Handle(metrics.GetNodeOverview))
		route.Post("/nodes/overviews", rest.Handle(metrics.ListNodeOverviews))
		route.Get("/nodes/{node}/cpu_usage", rest.Handle(metrics.GetNodeCPUUsage))
		route.Get("/nodes/{node}/cpu_request_usage", rest.Handle(metrics.GetNodeCPURequestUsage))
		route.Get("/nodes/{node}/memory_usage", rest.Handle(metrics.GetNodeMemoryUsage))
		route.Get("/nodes/{node}/memory_request_usage", rest.Handle(metrics.GetNodeMemoryRequestUsage))
		route.Get("/nodes/{node}/network_receive", rest.Handle(metrics.GetNodeNetworkReceiveUsage))
		route.Get("/nodes/{node}/network_transmit", rest.Handle(metrics.GetNodeNetworkTransmitUsage))
		route.Get("/nodes/{node}/disk_usage", rest.Handle(metrics.GetNodeDiskUsage))
		route.Get("/nodes/{node}/diskio_usage", rest.Handle(metrics.GetNodeDiskioUsage))
		route.Post("/namespaces/{namespace}/pods/cpu_usage", rest.Handle(
			metrics.PodCPUUsage)) // 多个Pod场景, 可能有几十，上百Pod场景, 需要使用 Post 传递参数
		route.Post("/namespaces/{namespace}/pods/cpu_limit_usage", rest.Handle(metrics.PodCPULimitUsage))
		route.Post("/namespaces/{namespace}/pods/cpu_request_usage", rest.Handle(metrics.PodCPURequestUsage))
		route.Post("/namespaces/{namespace}/pods/memory_used", rest.Handle(metrics.PodMemoryUsed))
		route.Post("/namespaces/{namespace}/pods/network_receive", rest.Handle(metrics.PodNetworkReceive))
		route.Post("/namespaces/{namespace}/pods/network_transmit", rest.Handle(metrics.PodNetworkTransmit))
		route.Get("/namespaces/{namespace}/pods/{pod}/containers/{container}/cpu_usage",
			rest.Handle(metrics.ContainerCPUUsage))
		route.Get("/namespaces/{namespace}/pods/{pod}/containers/{container}/memory_used",
			rest.Handle(metrics.ContainerMemoryUsed))
		route.Get("/namespaces/{namespace}/pods/{pod}/containers/{container}/cpu_limit",
			rest.Handle(metrics.ContainerCPULimit))
		route.Get("/namespaces/{namespace}/pods/{pod}/containers/{container}/memory_limit",
			rest.Handle(metrics.ContainerMemoryLimit))
		route.Get("/namespaces/{namespace}/pods/{pod}/containers/{container}/disk_read_total",
			rest.Handle(metrics.ContainerDiskReadTotal))
		route.Get("/namespaces/{namespace}/pods/{pod}/containers/{container}/disk_write_total",
			rest.Handle(metrics.ContainerDiskWriteTotal))

		route.Get("/namespaces/{namespace}/service_monitors",
			rest.Handle(service_monitor.ListServiceMonitors))
		route.Get("/namespaces/{namespace}/service_monitors/{name}",
			rest.Handle(service_monitor.GetServiceMonitor))
		route.Post("/namespaces/{namespace}/service_monitors",
			rest.Handle(service_monitor.CreateServiceMonitor))
		route.Put("/namespaces/{namespace}/service_monitors/{name}",
			rest.Handle(service_monitor.UpdateServiceMonitor))
		route.Delete("/namespaces/{namespace}/service_monitors/{name}",
			rest.Handle(service_monitor.DeleteServiceMonitor))
		route.Get("/service_monitors",
			rest.Handle(service_monitor.ListServiceMonitors))
		route.Post("/service_monitors/batchdelete",
			rest.Handle(service_monitor.BatchDeleteServiceMonitor))

		route.Get("/namespaces/{namespace}/pod_monitors",
			rest.Handle(podmonitor.ListPodMonitors))
		route.Get("/namespaces/{namespace}/pod_monitors/{name}",
			rest.Handle(podmonitor.GetPodMonitor))
		route.Post("/namespaces/{namespace}/pod_monitors",
			rest.Handle(podmonitor.CreatePodMonitor))
		route.Put("/namespaces/{namespace}/pod_monitors/{name}",
			rest.Handle(podmonitor.UpdatePodMonitor))
		route.Delete("/namespaces/{namespace}/pod_monitors/{name}",
			rest.Handle(podmonitor.DeletePodMonitor))
		route.Get("/pod_monitors",
			rest.Handle(podmonitor.ListPodMonitors))
		route.Post("/pod_monitors/batchdelete",
			rest.Handle(podmonitor.BatchDeletePodMonitor))

		route.Get("/event_data_id", rest.Handle(metrics.GetClusterEventDataId))
	})
	return r
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
func HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ReadyHandler 健康检查
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
