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

// Package query for query
package query

import (
	"context"
	"path"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/route"
	v1 "github.com/thanos-io/thanos/pkg/api/query"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/extprom"
	extpromhttp "github.com/thanos-io/thanos/pkg/extprom/http"
	"github.com/thanos-io/thanos/pkg/gate"
	"github.com/thanos-io/thanos/pkg/logging"
	"github.com/thanos-io/thanos/pkg/prober"
	"github.com/thanos-io/thanos/pkg/query"
	httpserver "github.com/thanos-io/thanos/pkg/server/http"
	"github.com/thanos-io/thanos/pkg/ui"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// QueryAPI promql api 服务, 封装 thaons 的API使用
type QueryAPI struct { // nolint
	StoresList   []string
	endpoints    *query.EndpointSet
	srv          *httpserver.Server
	httpAddr     string
	httpPort     string
	addrIPv6     string
	statusProber prober.Probe
	ctx          context.Context
}

// NewQueryAPI 这个包对thanos的query做一些封装，重新调用等
// 使用配置文件配置
// 启动 query 模块，暴露http
// query模块对应我们的store
func NewQueryAPI(ctx context.Context, reg *prometheus.Registry, tracer opentracing.Tracer, logKit blog.GlogKit,
	httpAddr string, httpPort string, addrIPv6 string, strictStoreList []string, storeList []string,
	httpSDURLs []string, g *run.Group,
) (*QueryAPI, error) {
	discoveryClient, err := NewDiscoveryClient(ctx, reg, tracer, logKit, strictStoreList, storeList, httpSDURLs, g)
	if err != nil {
		return nil, err
	}

	queryableCreator := NewQueryableCreator(reg, logKit, discoveryClient)
	queryEngine := NewQueryEngine(reg, logKit)

	apiServer := &QueryAPI{
		ctx:       ctx,
		endpoints: discoveryClient.Endpoints(),
		httpAddr:  httpAddr,
		httpPort:  httpPort,
		addrIPv6:  addrIPv6,
	}
	blog.Infof("store list: [%v]", storeList)

	var comp = component.Query

	httpProbe := prober.NewHTTP()
	grpcProbe := prober.NewGRPC()
	apiServer.statusProber = prober.Combine(
		httpProbe,
		grpcProbe,
		prober.NewInstrumentation(comp, logKit, extprom.WrapRegistererWithPrefix("bcs_monitor_", reg)),
	)

	// Start query API + UI HTTP server.
	router := route.New()

	// Configure Request Logging for HTTP calls.
	logMiddleware := logging.NewHTTPServerMiddleware(logKit)

	ins := extpromhttp.NewInstrumentationMiddleware(reg, nil)
	tenantAuthMiddleware, _ := NewTenantAuthMiddleware(ctx, ins)

	// 启动一个ui界面
	var prefix = ""
	if !config.G.IsDevMode() {
		// 正式环境, 接入到网关后面
		prefix = path.Join(config.G.Web.RoutePrefix, config.QueryServicePrefix)
	}
	ui.NewQueryUI(logKit, discoveryClient.Endpoints(), prefix, "", "").Register(router, ins)

	api := v1.NewQueryAPI(
		logKit,
		discoveryClient.Endpoints().GetEndpointStatus,
		queryEngine,
		queryableCreator,
		NewEmptyRuleClient(),
		NewEmptyTargetClient(),
		NewEmptyMetaDataClient(),
		NewEmptyExemplarClient(),
		enableAutodownsampling,
		enableQueryPartialResponse,
		true, // 用不到rule接口
		true, // 用不到target接口
		true, // 用不到 metadata接口
		true, // enableExemplarPartialResponse
		true, // enableQueryPushdown
		queryReplicaLabels,
		nil,
		defaultRangeQueryStep,
		instantDefaultMaxSourceResolution,
		defaultMetadataTimeRange,
		false, // disableCORS
		gate.New(
			extprom.WrapRegistererWithPrefix("bcs_monitor_query_concurrent_", reg),
			maxConcurrentQueries,
		),
		reg,
	)

	api.Register(router.WithPrefix("/api/v1"), tracer, logKit, tenantAuthMiddleware, logMiddleware)

	srv := httpserver.New(logKit, reg, comp, httpProbe, httpserver.WithGracePeriod(time.Minute*2))
	srv.Handle("/", router)

	apiServer.srv = srv

	blog.Infof("starting query node")
	return apiServer, nil
}

// Run 启动服务
func (a *QueryAPI) Run() error {
	a.statusProber.Healthy()
	a.statusProber.Ready()

	dualStackListener := listener.NewDualStackListener()
	addr := utils.GetListenAddr(a.httpAddr, a.httpPort)
	if err := dualStackListener.AddListenerWithAddr(addr); err != nil {
		return err
	}
	blog.Infow("listening for requests and metrics", "service", "query", "address", addr)

	if a.addrIPv6 != "" && a.addrIPv6 != a.httpAddr {
		v6Addr := utils.GetListenAddr(a.addrIPv6, a.httpPort)
		if err := dualStackListener.AddListenerWithAddr(v6Addr); err != nil {
			return err
		}
		blog.Infof("query serve dualStackListener with ipv6: %s", v6Addr)
	}

	return a.srv.Serve(dualStackListener)
}

// Close 停止服务
func (a *QueryAPI) Close(err error) {
	a.statusProber.NotHealthy(err)
	a.statusProber.NotReady(err)
	a.srv.Shutdown(err)
}
