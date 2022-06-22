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
	"time"

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/TencentBlueKing/bkmonitor-kits/logger/gokit"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	httpdiscovery "github.com/prometheus/prometheus/discovery/http"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/thanos-io/thanos/pkg/discovery/cache"
	"github.com/thanos-io/thanos/pkg/discovery/dns"
	"github.com/thanos-io/thanos/pkg/extgrpc"
	"github.com/thanos-io/thanos/pkg/extprom"
	"github.com/thanos-io/thanos/pkg/query"
	"github.com/thanos-io/thanos/pkg/runutil"
)

// DiscoveryClient 支持的服务发现, 包含静态配置, http-sd， 命令行和配置文件来源
type DiscoveryClient struct {
	endpoints *query.EndpointSet
}

// NewDiscoveryClient
func NewDiscoveryClient(ctx context.Context, reg *prometheus.Registry, tracer opentracing.Tracer, kitLogger gokit.Logger, storeList []string, httpSDURLs []string, g *run.Group) (*DiscoveryClient, error) {
	dnsStoreProvider := dns.NewProvider(
		kitLogger,
		extprom.WrapRegistererWithPrefix("bcs_monitor_query_store_apis_", reg),
		dns.ResolverType(dns.MiekgdnsResolverType),
	)

	dialOpts, err := extgrpc.StoreClientGRPCOpts(kitLogger, reg, tracer, false, false, "", "", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "building gRPC client")
	}

	endpoints := query.NewEndpointSet(
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

	client := &DiscoveryClient{endpoints: endpoints}

	return client, nil
}

// Endpoints 返回 EndpointSet
func (c *DiscoveryClient) Endpoints() *query.EndpointSet {
	return c.endpoints
}
