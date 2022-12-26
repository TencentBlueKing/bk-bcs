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

// Package query xxx
package query

import (
	"context"
	"math"
	"path"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/TencentBlueKing/bkmonitor-kits/logger/gokit"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/route"
	"github.com/prometheus/prometheus/promql"
	v1 "github.com/thanos-io/thanos/pkg/api/query"
	"github.com/thanos-io/thanos/pkg/compact/downsample"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/extprom"
	extpromhttp "github.com/thanos-io/thanos/pkg/extprom/http"
	"github.com/thanos-io/thanos/pkg/gate"
	"github.com/thanos-io/thanos/pkg/logging"
	"github.com/thanos-io/thanos/pkg/prober"
	"github.com/thanos-io/thanos/pkg/query"
	httpserver "github.com/thanos-io/thanos/pkg/server/http"
	"github.com/thanos-io/thanos/pkg/store"
	"github.com/thanos-io/thanos/pkg/ui"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// QueryAPI promql api 服务, 封装 thaons 的API使用
type QueryAPI struct {
	StoresList   []string
	endpoints    *query.EndpointSet
	srv          *httpserver.Server
	httpAddr     string
	addrIPv6     string
	statusProber prober.Probe
	ctx          context.Context
}

// NewQueryAPI 这个包对thanos的query做一些封装，重新调用等
// 使用配置文件配置
// 启动 query 模块，暴露http
// query模块对应我们的store
func NewQueryAPI(
	ctx context.Context,
	reg *prometheus.Registry,
	tracer opentracing.Tracer,
	kitLogger gokit.Logger,
	httpAddr string,
	addrIPv6 string,
	strictStoreList []string,
	storeList []string,
	httpSDURLs []string,
	g *run.Group,
) (*QueryAPI, error) {
	discoveryClient, err := NewDiscoveryClient(ctx, reg, tracer, kitLogger, strictStoreList, storeList, httpSDURLs, g)
	if err != nil {
		return nil, err
	}

	queryableCreator := NewQueryableCreator(reg, kitLogger, discoveryClient)
	queryEngine := NewQueryEngine(reg, kitLogger)

	apiServer := &QueryAPI{
		ctx:       ctx,
		endpoints: discoveryClient.Endpoints(),
		httpAddr:  httpAddr,
		addrIPv6:  addrIPv6,
	}
	logger.Infof("store list: [%v]", storeList)

	var comp = component.Query

	httpProbe := prober.NewHTTP()
	grpcProbe := prober.NewGRPC()
	apiServer.statusProber = prober.Combine(
		httpProbe,
		grpcProbe,
		prober.NewInstrumentation(comp, kitLogger, extprom.WrapRegistererWithPrefix("bcs_monitor_", reg)),
	)

	// Start query API + UI HTTP server.
	router := route.New()

	// Configure Request Logging for HTTP calls.
	logMiddleware := logging.NewHTTPServerMiddleware(kitLogger)

	ins := extpromhttp.NewInstrumentationMiddleware(reg, nil)
	tenantAuthMiddleware, _ := NewTenantAuthMiddleware(ctx, ins)

	// 启动一个ui界面
	var prefix = ""
	if !config.G.IsDevMode() {
		// 正式环境, 接入到网关后面
		prefix = path.Join(config.G.Web.RoutePrefix, config.QueryServicePrefix)
	}
	ui.NewQueryUI(kitLogger, discoveryClient.Endpoints(), prefix, "", "").Register(router, ins)
	engineOpts := promql.EngineOpts{
		Logger:        kitLogger,
		Reg:           reg,
		MaxSamples:    math.MaxInt32,
		Timeout:       queryTimeout,
		LookbackDelta: lookbackDelta,
		NoStepSubqueryIntervalFn: func(int64) int64 {
			return defaultEvaluationInterval.Milliseconds()
		},
	}
	lookbackDeltaCreator := LookbackDeltaFactory(engineOpts, dynamicLookbackDelta)
	queryTelemetryDurationQuantiles := []float64{0.1, 0.25, 0.75, 1.25, 1.75, 2.5, 3, 5, 10}
	queryTelemetrySamplesQuantiles := []int64{100, 1000, 10000, 100000, 1000000}
	queryTelemetrySeriesQuantiles := []int64{10, 100, 1000, 10000, 100000}
	api := v1.NewQueryAPI(
		kitLogger,
		discoveryClient.Endpoints().GetEndpointStatus,
		queryEngine(math.MaxInt32),
		lookbackDeltaCreator,
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
		/*	queryTelemetryDurationQuantiles := cmd.Flag("query.telemetry.request-duration-seconds-quantiles", "The quantiles for exporting metrics about the request duration quantiles.").Default("0.1", "0.25", "0.75", "1.25", "1.75", "2.5", "3", "5", "10").Float64List()
			queryTelemetrySamplesQuantiles := cmd.Flag("query.telemetry.request-samples-quantiles", "The quantiles for exporting metrics about the samples count quantiles.").Default("100", "1000", "10000", "100000", "1000000").Int64List()
			queryTelemetrySeriesQuantiles := cmd.Flag("query.telemetry.request-series-seconds-quantiles", "The quantiles for exporting metrics about the series count quantiles.").Default("10", "100", "1000", "10000", "100000").Int64List()*/
		store.NewSeriesStatsAggregator(
			reg,
			queryTelemetryDurationQuantiles,
			queryTelemetrySamplesQuantiles,
			queryTelemetrySeriesQuantiles,
		),
		reg,
	)

	api.Register(router.WithPrefix("/api/v1"), tracer, kitLogger, tenantAuthMiddleware, logMiddleware)

	srv := httpserver.New(kitLogger, reg, comp, httpProbe,
		httpserver.WithListen(httpAddr),
		httpserver.WithGracePeriod(time.Minute*2),
	)
	srv.Handle("/", router)

	apiServer.srv = srv

	logger.Infof("starting query node")
	return apiServer, nil
}

// Run 启动服务
func (a *QueryAPI) Run() error {
	a.statusProber.Healthy()
	a.statusProber.Ready()

	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(a.httpAddr); err != nil {
		return err
	}

	if a.addrIPv6 != "" {
		if err := dualStackListener.AddListenerWithAddr(a.addrIPv6); err != nil {
			return err
		}
		logger.Infof("query serve dualStackListener with ipv6: %s", a.addrIPv6)
	}

	return a.srv.Serve(dualStackListener)
}

// Close 停止服务
func (a *QueryAPI) Close(err error) {
	a.statusProber.NotHealthy(err)
	a.statusProber.NotReady(err)
	a.srv.Shutdown(err)
}

// LookbackDeltaFactory creates from 1 to 3 lookback deltas depending on
// dynamicLookbackDelta and eo.LookbackDelta and returns a function
// that returns appropriate lookback delta for given maxSourceResolutionMillis.
func LookbackDeltaFactory(
	eo promql.EngineOpts,
	dynamicLookbackDelta bool,
) func(int64) time.Duration {
	resolutions := []int64{downsample.ResLevel0}
	if dynamicLookbackDelta {
		resolutions = []int64{downsample.ResLevel0, downsample.ResLevel1, downsample.ResLevel2}
	}
	var (
		lds = make([]time.Duration, len(resolutions))
		ld  = eo.LookbackDelta.Milliseconds()
	)

	lookbackDelta := eo.LookbackDelta
	for i, r := range resolutions {
		if ld < r {
			lookbackDelta = time.Duration(r) * time.Millisecond
		}

		lds[i] = lookbackDelta
	}
	return func(maxSourceResolutionMillis int64) time.Duration {
		for i := len(resolutions) - 1; i >= 1; i-- {
			left := resolutions[i-1]
			if resolutions[i-1] < ld {
				left = ld
			}
			if left < maxSourceResolutionMillis {
				return lds[i]
			}
		}
		return lds[0]
	}
}
