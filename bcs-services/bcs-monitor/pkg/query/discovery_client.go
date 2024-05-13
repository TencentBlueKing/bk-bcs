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

package query

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// DiscoveryClient 支持的服务发现, 包含静态配置, http-sd， 命令行和配置文件来源
type DiscoveryClient struct {
	endpoints            *query.EndpointSet
	reg                  *prometheus.Registry
	dnsStoreProvider     *dns.Provider
	storeCacheMap        map[string]*cache.Cache
	httpSDClientGroupMap map[string]*HTTPSDClientGroup
	mtx                  sync.RWMutex
}

// NewDiscoveryClient xxx
func NewDiscoveryClient(ctx context.Context, reg *prometheus.Registry, tracer opentracing.Tracer,
	logKit blog.GlogKit, strictStoreList []string, storeList []string, httpSDURLs []string,
	g *run.Group) (*DiscoveryClient, error) {

	// 检查静态 store 配置
	for _, endpoint := range strictStoreList {
		if dns.IsDynamicNode(endpoint) {
			return nil, errors.Errorf("%s is a dynamically specified endpoint i.e. it uses SD and that is not "+
				"permitted under strict mode. Use --store for this", endpoint)
		}
	}

	dnsStoreProvider := dns.NewProvider(
		logKit,
		extprom.WrapRegistererWithPrefix("bcs_monitor_query_store_apis_", reg),
		dns.MiekgdnsResolverType,
	)

	dialOpts, err := extgrpc.StoreClientGRPCOpts(logKit, reg, tracer, false, false, "", "", "", "")
	// 高可用 添加重试逻辑
	opts := []grpc_retry.CallOption{
		grpc_retry.WithCodes(codes.Unavailable),
		grpc_retry.WithMax(3),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
	}
	dialOpts = append(dialOpts,
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(opts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // 负载均衡
	)

	if err != nil {
		return nil, errors.Wrap(err, "building gRPC client")
	}

	endpoints := getEndpoints(logKit, reg, strictStoreList, dnsStoreProvider, dialOpts)
	client := &DiscoveryClient{
		reg:                  reg,
		endpoints:            endpoints,
		dnsStoreProvider:     dnsStoreProvider,
		storeCacheMap:        map[string]*cache.Cache{},
		httpSDClientGroupMap: map[string]*HTTPSDClientGroup{},
	}

	// Periodically update the store set with the addresses we see in our cluster.
	{
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
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
	httpSDConfs, err := parseHTTPSDURLs(httpSDURLs)
	if err != nil {
		return nil, err
	}

	for _, conf := range httpSDConfs {
		if err := client.addHTTPDiscovery(ctx, logKit, conf, g); err != nil {
			return nil, err
		}
	}

	for _, conf := range config.G.QueryStore.HTTPSDConfigs {
		if err := client.addHTTPDiscovery(ctx, logKit, conf, g); err != nil {
			return nil, err
		}
	}

	// Periodically update the addresses from static flags and file SD by resolving them using DNS SD if necessary.
	resolveStoreProvider(ctx, g, dnsStoreProvider, client)

	return client, nil
}

func resolveStoreProvider(ctxm context.Context, g *run.Group, dnsStoreProvider *dns.Provider, client *DiscoveryClient) {
	ctx, cancel := context.WithCancel(ctxm)
	g.Add(func() error {
		return runutil.Repeat(time.Second*30, ctx.Done(), func() error {
			resolveCtx, resolveCancel := context.WithTimeout(ctx, time.Second*30)
			defer resolveCancel()
			if err := dnsStoreProvider.Resolve(resolveCtx, client.Addresses()); err != nil {
				blog.Errorw("failed to resolve addresses for storeAPIs", "err", err)
			}
			return nil
		})
	}, func(error) {
		cancel()
	})
}

func getEndpoints(logKit blog.GlogKit, reg *prometheus.Registry, strictStoreList []string,
	dnsStoreProvider *dns.Provider, dialOpts []grpc.DialOption) *query.EndpointSet {
	return query.NewEndpointSet(
		logKit,
		reg,
		func() (specs []*query.GRPCEndpointSpec) {
			// Add strict & static nodes.
			for _, addr := range strictStoreList {
				specs = append(specs, query.NewGRPCEndpointSpec(addr, true))
			}

			// Add DNS resolved addresses from static flags and file SD.
			for _, addr := range dnsStoreProvider.Addresses() {
				specs = append(specs, query.NewGRPCEndpointSpec(addr, false))
			}
			return specs
		},
		dialOpts,
		unhealthyStoreTimeout,
	)
}

func (c *DiscoveryClient) addStaticDiscovery(name string, tgs []*targetgroup.Group) {
	httpSDCache := cache.New()

	c.mtx.Lock()
	c.storeCacheMap[name] = httpSDCache
	c.mtx.Unlock()

	httpSDCache.Update(tgs)
}

func (c *DiscoveryClient) addHTTPDiscovery(
	ctx context.Context, logKit blog.GlogKit, conf *httpdiscovery.SDConfig, g *run.Group) error {
	client, err := NewHTTPSDClientGroup(ctx, logKit, c.reg, conf, g, c.ForceRefreshEndpoints)
	if err != nil {
		return err
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.httpSDClientGroupMap[client.id] = client
	return nil
}

// ForceRefreshEndpoints xxx
func (c *DiscoveryClient) ForceRefreshEndpoints(ctx context.Context) {
	c.endpoints.Update(ctx)

	if err := c.dnsStoreProvider.Resolve(ctx, c.Addresses()); err != nil {
		blog.Errorw("force reresh endpoints failed to resolve addresses for storeAPIs", "err", err)
	}
}

// Addresses 服务发现所有地址
func (c *DiscoveryClient) Addresses() []string {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	addresses := []string{}
	for _, c := range c.storeCacheMap {
		addresses = append(addresses, c.Addresses()...)
	}

	for _, c := range c.httpSDClientGroupMap {
		addresses = append(addresses, c.Addresses()...)
	}
	return addresses
}

// Endpoints 返回 EndpointSet
func (c *DiscoveryClient) Endpoints() *query.EndpointSet {
	return c.endpoints
}

// HTTPSDClientGroup http sd client group
type HTTPSDClientGroup struct {
	mtx             sync.RWMutex
	id              string
	logKit          blog.GlogKit
	ctx             context.Context
	reg             *prometheus.Registry
	httpSDClientMap map[string]*HTTPSDClient
}

// NewHTTPSDClientGroup xxx
func NewHTTPSDClientGroup(ctx context.Context, logKit blog.GlogKit, reg *prometheus.Registry,
	conf *httpdiscovery.SDConfig, g *run.Group, forceRefreshFunc func(ctx context.Context)) (*HTTPSDClientGroup, error) {
	id := fmt.Sprintf("%s:%s", conf.Name(), conf.URL)

	httpSDClientMap := map[string]*HTTPSDClient{}

	c := &HTTPSDClientGroup{
		ctx:             ctx,
		logKit:          logKit,
		reg:             reg,
		id:              id,
		httpSDClientMap: httpSDClientMap,
	}

	updateCtx, updateCancel := context.WithCancel(ctx)
	g.Add(func() error {
		return runutil.Repeat(time.Second*30, ctx.Done(), func() error {
			addrMap, err := c.parseURLHost(conf.URL)
			if err != nil {
				klog.ErrorS(err, "resolve http sd addresses failed", "url", conf.URL)
				return nil
			}

			// 新增client
			for addr, u := range addrMap {
				_, ok := httpSDClientMap[addr]
				if ok {
					continue
				}
				s, err := NewHTTPSDClient(updateCtx, logKit, *conf, addr, *u, forceRefreshFunc)
				if err != nil {
					klog.ErrorS(err, "create http sd client failed", "url", conf.URL, "addr", addr)
					continue
				}
				_ = s.Run()

				c.mtx.Lock()
				httpSDClientMap[addr] = s
				c.mtx.Unlock()
			}

			// 清理掉不需要的client
			for name, client := range httpSDClientMap {
				_, ok := addrMap[name]
				if !ok {
					client.Close()
					delete(httpSDClientMap, name)
				}
			}
			return nil
		})
	}, func(err error) {
		updateCancel()
	})

	return c, nil
}

func (c *HTTPSDClientGroup) parseURLHost(rawURL string) (map[string]*url.URL, error) {
	dnsStoreProvider := dns.NewProvider(
		c.logKit,
		nil,
		dns.MiekgdnsResolverType,
	)
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// Split the host and port if present.
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		// The host could be missing a port.
		host, port = u.Host, ""
	}
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// dnsStoreProvider 解析需要带上端口
	addr := fmt.Sprintf("%s:%s", host, port)

	if err := dnsStoreProvider.Resolve(c.ctx, []string{addr}); err != nil {
		return nil, err
	}

	addrMap := map[string]*url.URL{}
	for _, v := range dnsStoreProvider.Addresses() {
		addrMap[v] = u
	}

	return addrMap, nil
}

// Addresses xxx
func (c *HTTPSDClientGroup) Addresses() []string {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	addresses := []string{}
	for _, c := range c.httpSDClientMap {
		addresses = append(addresses, c.storeCache.Addresses()...)
	}
	return addresses
}

// HTTPSDClient http sd client
type HTTPSDClient struct {
	ctx              context.Context
	storeCache       *cache.Cache
	sdConf           *httpdiscovery.SDConfig
	sdClient         *httpdiscovery.Discovery
	addr             string // 保证唯一性
	cancel           func()
	forceRefreshFunc func(ctx context.Context)
}

// NewHTTPSDClient xxx
func NewHTTPSDClient(
	ctx context.Context, logKit blog.GlogKit, conf httpdiscovery.SDConfig, addr string, u url.URL,
	forceRefreshFunc func(ctx context.Context)) (*HTTPSDClient, error) {
	// Run File Service Discovery and update the store set when the files are modified.
	u.Host = addr
	conf.URL = u.String()
	sdClient, err := httpdiscovery.NewDiscovery(&conf, logKit)
	if err != nil {
		return nil, err
	}

	storeCache := cache.New()

	updateCtx, updateCancel := context.WithCancel(ctx)
	c := HTTPSDClient{
		ctx:              updateCtx,
		storeCache:       storeCache,
		sdConf:           &conf,
		sdClient:         sdClient,
		addr:             addr,
		cancel:           updateCancel,
		forceRefreshFunc: forceRefreshFunc,
	}
	return &c, nil
}

// Close xxx
func (c *HTTPSDClient) Close() {
	c.cancel()
}

// Run xxx
func (c *HTTPSDClient) Run() error {
	httpSDUpdates := make(chan []*targetgroup.Group)

	go func() {
		c.sdClient.Run(c.ctx, httpSDUpdates)
	}()

	firstRun := true

	go func() {
		for {
			select {
			case update := <-httpSDUpdates:
				// Discoverers sometimes send nil updates so need to check for it to avoid panics.
				if update == nil {
					continue
				}
				c.storeCache.Update(update)

				if firstRun {
					c.forceRefreshFunc(c.ctx)
					firstRun = false
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()

	return nil
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

// parseHTTPSDURLs 解析url配置, 命令行来源
func parseHTTPSDURLs(httpSDURLs []string) ([]*httpdiscovery.SDConfig, error) {
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
