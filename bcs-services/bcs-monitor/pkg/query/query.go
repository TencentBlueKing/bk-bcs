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

package query

import (
	"context"
	"math"
	"path"
	"strconv"
	"time"

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/TencentBlueKing/bkmonitor-kits/logger/gokit"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/route"
	httpdiscovery "github.com/prometheus/prometheus/discovery/http"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/promql"
	v1 "github.com/thanos-io/thanos/pkg/api/query"
	"github.com/thanos-io/thanos/pkg/compact/downsample"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/discovery/cache"
	"github.com/thanos-io/thanos/pkg/discovery/dns"
	"github.com/thanos-io/thanos/pkg/extgrpc"
	"github.com/thanos-io/thanos/pkg/extprom"
	extpromhttp "github.com/thanos-io/thanos/pkg/extprom/http"
	"github.com/thanos-io/thanos/pkg/gate"
	"github.com/thanos-io/thanos/pkg/logging"
	"github.com/thanos-io/thanos/pkg/prober"
	"github.com/thanos-io/thanos/pkg/query"
	"github.com/thanos-io/thanos/pkg/runutil"
	httpserver "github.com/thanos-io/thanos/pkg/server/http"
	"github.com/thanos-io/thanos/pkg/store"
	"github.com/thanos-io/thanos/pkg/tracing/client"
	"github.com/thanos-io/thanos/pkg/ui"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// QueryAPI promql api 服务, 封装 thaons 的API使用
type QueryAPI struct {
	StoresList   []string
	endpoints    *query.EndpointSet
	srv          *httpserver.Server
	statusProber prober.Probe
	ctx          context.Context
}

// 这个包对thanos的query做一些封装，重新调用等
// 使用配置文件配置
// 启动 query 模块，暴露http
// query模块对应我们的store
func NewQueryAPI(
	ctx context.Context,
	reg *prometheus.Registry,
	tracer opentracing.Tracer,
	kitLogger gokit.Logger,
	httpAddr string,
	storeList []string,
	httpSDURLs []string,
	g *run.Group,
) (*QueryAPI, error) {

	dnsStoreProvider := dns.NewProvider(
		kitLogger,
		extprom.WrapRegistererWithPrefix("bcs_monitor_query_store_apis_", reg),
		dns.ResolverType(dns.MiekgdnsResolverType),
	)

	dialOpts, err := extgrpc.StoreClientGRPCOpts(kitLogger, reg, tracer, false,
		false, "", "", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "building gRPC client")
	}
	var (
		apiServer = &QueryAPI{ctx: ctx}
		comp      = component.Query
		endpoints = query.NewEndpointSet(
			kitLogger,
			reg,
			func() (specs []*query.GRPCEndpointSpec) {
				// Add DNS resolved addresses from static flags and file SD.
				for _, addr := range dnsStoreProvider.Addresses() {
					specs = append(specs, query.NewGRPCEndpointSpec(addr, false))
				}
				return specs
			},
			dialOpts,
			unhealthyStoreTimeout,
		)

		proxy = store.NewProxyStore(kitLogger, reg, endpoints.GetStoreClients, component.Query, nil, storeResponseTimeout)

		queryableCreator = query.NewQueryableCreator(
			kitLogger,
			extprom.WrapRegistererWithPrefix("bcs_monitor_api_", reg),
			proxy,
			maxConcurrentSelects,
			queryTimeout,
		)

		engineOpts = promql.EngineOpts{
			Logger:        kitLogger,
			Reg:           reg,
			MaxSamples:    math.MaxInt32,
			Timeout:       queryTimeout,
			LookbackDelta: lookbackDelta,
			NoStepSubqueryIntervalFn: func(int64) int64 {
				return defaultEvaluationInterval.Milliseconds()
			},
		}
	)
	apiServer.endpoints = endpoints
	logger.Infof("store list: [%v]", storeList)

	// Periodically update the store set with the addresses we see in our cluster.
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return runutil.Repeat(5*time.Second, ctx.Done(), func() error {
				endpoints.Update(ctx)
				return nil
			})
		}, func(error) {
			cancel()
			endpoints.Close()
		})
	}

	// Run File Service Discovery and update the store set when the files are modified.
	httpSD, err := httpdiscovery.NewDiscovery(nil, kitLogger)
	if err != nil {
		return nil, err
	}
	httpSDCache := cache.New()

	{
		ctxRun, cancelRun := context.WithCancel(context.Background())
		httpSDUpdates := make(chan []*targetgroup.Group)

		g.Add(func() error {
			httpSD.Run(ctxRun, httpSDUpdates)
			return nil
		}, func(error) {
			cancelRun()
		})

		ctxUpdate, cancelUpdate := context.WithCancel(context.Background())
		g.Add(func() error {
			for {
				select {
				case update := <-httpSDUpdates:
					// Discoverers sometimes send nil updates so need to check for it to avoid panics.
					if update == nil {
						continue
					}
					httpSDCache.Update(update)
					endpoints.Update(ctxUpdate)

					if err := dnsStoreProvider.Resolve(ctxUpdate, append(httpSDCache.Addresses(), storeList...)); err != nil {
						level.Error(kitLogger).Log("msg", "failed to resolve addresses for storeAPIs", "err", err)
					}

				case <-ctxUpdate.Done():
					return nil
				}
			}
		}, func(error) {
			cancelUpdate()
		})
	}

	// Periodically update the addresses from static flags and file SD by resolving them using DNS SD if necessary.
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return runutil.Repeat(time.Second*30, ctx.Done(), func() error {
				resolveCtx, resolveCancel := context.WithTimeout(ctx, time.Second*30)
				defer resolveCancel()
				if err := dnsStoreProvider.Resolve(resolveCtx, append(httpSDCache.Addresses(), storeList...)); err != nil {
					logger.Errorw("failed to resolve addresses for storeAPIs", "err", err)
				}
				return nil
			})
		}, func(error) {
			cancel()
		})
	}

	httpProbe := prober.NewHTTP()
	grpcProbe := prober.NewGRPC()
	apiServer.statusProber = prober.Combine(
		httpProbe,
		grpcProbe,
		prober.NewInstrumentation(comp, kitLogger, extprom.WrapRegistererWithPrefix("bcs_monitor_", reg)),
	)

	if tracer == nil {
		tracer = client.NoopTracer()
	}

	// Start query API + UI HTTP server.
	{
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
		ui.NewQueryUI(kitLogger, endpoints, prefix, "", "").Register(router, ins)

		api := v1.NewQueryAPI(
			kitLogger,
			endpoints.GetEndpointStatus,
			engineFactory(promql.NewEngine, engineOpts, dynamicLookbackDelta),
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
				extprom.WrapRegistererWithPrefix("bcs_monitor_api_concurrent_", reg),
				maxConcurrentQueries,
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
	}

	logger.Infof("starting query node")
	return apiServer, nil
}

// engineFactory creates from 1 to 3 promql.Engines depending on
// dynamicLookbackDelta and eo.LookbackDelta and returns a function
// that returns appropriate engine for given maxSourceResolutionMillis.
//
// instead of creating several Engines here.
func engineFactory(
	newEngine func(promql.EngineOpts) *promql.Engine,
	eo promql.EngineOpts,
	dynamicLookbackDelta bool,
) func(int64) *promql.Engine {
	resolutions := []int64{downsample.ResLevel0}
	if dynamicLookbackDelta {
		resolutions = []int64{downsample.ResLevel0, downsample.ResLevel1, downsample.ResLevel2}
	}
	var (
		engines = make([]*promql.Engine, len(resolutions))
		ld      = eo.LookbackDelta.Milliseconds()
	)
	wrapReg := func(engineNum int) prometheus.Registerer {
		return extprom.WrapRegistererWith(map[string]string{"engine": strconv.Itoa(engineNum)}, eo.Reg)
	}

	lookbackDelta := eo.LookbackDelta
	for i, r := range resolutions {
		if ld < r {
			lookbackDelta = time.Duration(r) * time.Millisecond
		}
		engines[i] = newEngine(promql.EngineOpts{
			Logger:                   eo.Logger,
			Reg:                      wrapReg(i),
			MaxSamples:               eo.MaxSamples,
			Timeout:                  eo.Timeout,
			ActiveQueryTracker:       eo.ActiveQueryTracker,
			LookbackDelta:            lookbackDelta,
			NoStepSubqueryIntervalFn: eo.NoStepSubqueryIntervalFn,
		})
	}
	return func(maxSourceResolutionMillis int64) *promql.Engine {
		for i := len(resolutions) - 1; i >= 1; i-- {
			left := resolutions[i-1]
			if resolutions[i-1] < ld {
				left = ld
			}
			if left < maxSourceResolutionMillis {
				return engines[i]
			}
		}
		return engines[0]
	}
}

// RunHttp
func (a *QueryAPI) RunHttp() error {
	a.statusProber.Healthy()
	a.statusProber.Ready()
	return a.srv.ListenAndServe()
}

// ShutDownHttp
func (a *QueryAPI) ShutDownHttp(err error) {
	a.statusProber.NotHealthy(err)
	a.statusProber.NotReady(err)
	a.srv.Shutdown(err)
}

// RunGetStore 周期性对store进行健康检查，剔除不健康的stores
func (a *QueryAPI) RunGetStore() error {
	return runutil.Repeat(5*time.Second, a.ctx.Done(), func() error {
		a.endpoints.Update(a.ctx)
		return nil
	})
}

// ShutDownGetStore
func (a *QueryAPI) ShutDownGetStore(_ error) {
	a.endpoints.Close()
}
