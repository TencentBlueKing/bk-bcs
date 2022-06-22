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
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/TencentBlueKing/bkmonitor-kits/logger/gokit"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
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
	endpoints        *query.EndpointSet
	dnsStoreProvider *dns.Provider
	storeCacheMap    map[string]*cache.Cache
	mtx              sync.RWMutex
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

	client := &DiscoveryClient{
		endpoints:        endpoints,
		dnsStoreProvider: dnsStoreProvider,
		storeCacheMap:    map[string]*cache.Cache{},
	}

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

	cmdStore := parseStaticStore(storeList)
	client.addStaticDiscovery("cmd", cmdStore)
	client.addStaticDiscovery("conf", config.G.QueryStore.StaticConfigs)

	// Run File Service Discovery and update the store set when the files are modified.
	httpSDConfs, err := parseHttpSDURLs(httpSDURLs)
	if err != nil {
		return nil, err
	}

	for _, conf := range httpSDConfs {

		if err := client.addHTTPDiscovery(ctx, kitLogger, conf, g); err != nil {
			return nil, err
		}
	}

	for _, conf := range config.G.QueryStore.HTTPSDConfigs {
		if err := client.addHTTPDiscovery(ctx, kitLogger, conf, g); err != nil {
			return nil, err
		}
	}

	// Periodically update the addresses from static flags and file SD by resolving them using DNS SD if necessary.
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return runutil.Repeat(time.Second*30, ctx.Done(), func() error {
				resolveCtx, resolveCancel := context.WithTimeout(ctx, time.Second*30)
				defer resolveCancel()
				if err := dnsStoreProvider.Resolve(resolveCtx, client.Addresses()); err != nil {
					logger.Errorw("failed to resolve addresses for storeAPIs", "err", err)
				}
				return nil
			})
		}, func(error) {
			cancel()
		})
	}

	return client, nil
}

func (c *DiscoveryClient) addStaticDiscovery(name string, tgs []*targetgroup.Group) error {
	httpSDCache := cache.New()

	c.mtx.Lock()
	c.storeCacheMap[name] = httpSDCache
	c.mtx.Unlock()

	httpSDCache.Update(tgs)
	return nil
}

func (c *DiscoveryClient) addHTTPDiscovery(ctx context.Context, kitLogger gokit.Logger, conf *httpdiscovery.SDConfig, g *run.Group) error {
	// Run File Service Discovery and update the store set when the files are modified.
	httpSD, err := httpdiscovery.NewDiscovery(conf, kitLogger)
	if err != nil {
		return err
	}

	httpSDCache := cache.New()

	c.mtx.Lock()
	name := fmt.Sprintf("%s:%s", conf.Name(), conf.URL)
	c.storeCacheMap[name] = httpSDCache
	c.mtx.Unlock()

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

				c.endpoints.Update(ctxUpdate)
				if err := c.dnsStoreProvider.Resolve(ctx, c.Addresses()); err != nil {
					level.Error(kitLogger).Log("msg", "failed to resolve addresses for storeAPIs", "err", err)
				}

			case <-ctxUpdate.Done():
				return nil
			}
		}
	}, func(error) {
		cancelUpdate()
	})

	return nil
}

func (c *DiscoveryClient) Addresses() []string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	addresses := []string{}
	for _, c := range c.storeCacheMap {
		addresses = append(addresses, c.Addresses()...)
	}
	return addresses
}

// Endpoints 返回 EndpointSet
func (c *DiscoveryClient) Endpoints() *query.EndpointSet {
	return c.endpoints
}

// parseStaticStore 解析静态IP配置, 命令行来源
func parseStaticStore(storeList []string) []*targetgroup.Group {
	tgs := make([]*targetgroup.Group, 0, len(storeList))
	for _, store := range storeList {
		tgs = append(tgs, &targetgroup.Group{
			Targets: []model.LabelSet{
				{model.AddressLabel: model.LabelValue(store)},
			},
		})
	}
	return tgs
}

// parseHttpSDURLs 解析url配置, 命令行来源
func parseHttpSDURLs(httpSDURLs []string) ([]*httpdiscovery.SDConfig, error) {
	confs := make([]*httpdiscovery.SDConfig, 0, len(httpSDURLs))
	for _, rawURL := range httpSDURLs {
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}

		username := u.User.Username()
		pwd, ok := u.User.Password()

		// 清理掉鉴权信息
		u.User = nil
		conf := &httpdiscovery.SDConfig{
			HTTPClientConfig: promconfig.HTTPClientConfig{
				FollowRedirects: true,
				EnableHTTP2:     true,
			},
			RefreshInterval: model.Duration(time.Second * 10),
			URL:             u.String(),
		}

		if ok {
			if username == "bearer_token" {
				conf.HTTPClientConfig.BearerToken = promconfig.Secret(pwd)
			} else {
				conf.HTTPClientConfig.BasicAuth = &promconfig.BasicAuth{
					Username: username,
					Password: promconfig.Secret(pwd),
				}
			}
		}

		confs = append(confs, conf)
	}
	return confs, nil
}
