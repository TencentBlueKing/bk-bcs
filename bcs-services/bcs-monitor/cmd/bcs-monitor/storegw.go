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

package main

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/extgrpc"
	"github.com/thanos-io/thanos/pkg/extprom"
	"github.com/thanos-io/thanos/pkg/prober"
	"github.com/thanos-io/thanos/pkg/query"
	"github.com/thanos-io/thanos/pkg/runutil"
	grpcserver "github.com/thanos-io/thanos/pkg/server/grpc"
	httpserver "github.com/thanos-io/thanos/pkg/server/http"
	"github.com/thanos-io/thanos/pkg/store"
	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

var (
	grpcPort                  string
	grpcAdvertisePortRangeStr string
	grpcAdvertiseIP           string
)

// StoreGWCmd StoreGW 命令
func StoreGWCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storegw",
		Short: "Heterogeneous storage gateway",
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runCmd(cmd, runStoreGW)
	}

	flags := cmd.Flags()
	flags.StringVar(&grpcPort, "grpc-port", grpcPort, "Listen host:port for grpc endpoints.")
	flags.StringVar(&grpcAdvertiseIP, "grpc-advertise-ip", "127.0.0.1", "grpc advertise ip")
	flags.StringVar(&grpcAdvertisePortRangeStr, "grpc-advertise-port-range", "28000-29000", "grpc advertise port range")

	return cmd
}

// runStoreGW
func runStoreGW(ctx context.Context, g *run.Group, opt *option) error {
	logKit := blog.LogKit
	gw, err := storegw.NewStoreGW(ctx, logKit, opt.reg, grpcAdvertiseIP, grpcAdvertisePortRangeStr,
		config.G.StoreGWList, storegw.GetStoreSvr)
	if err != nil {
		return err
	}

	// 可用性评估，必须全部grpc端口,可用，才ready

	// http 服务
	{
		httpProbe := prober.NewHTTP()
		statusProber := prober.Combine(
			httpProbe,
			prober.NewInstrumentation(component.Store, logKit,
				extprom.WrapRegistererWithPrefix("bcsmonitor_", opt.reg)),
		)

		httpSrv := httpserver.New(logKit, opt.reg, component.Store, httpProbe,
			httpserver.WithListen(utils.GetListenAddr(bindAddress, httpPort)),
			httpserver.WithGracePeriod(time.Second*5),
		)

		router := api.RegisterStoreGWRoutes(gw)
		httpSrv.Handle("/", router)

		g.Add(func() error {
			statusProber.Healthy()
			statusProber.Ready()

			return httpSrv.ListenAndServe()
		}, func(err error) {
			defer statusProber.NotHealthy(err)
			defer statusProber.NotReady(err)

			httpSrv.Shutdown(err)
		})
	}

	// Periodically update the store set with the addresses we see in our cluster.
	var endpoints *query.EndpointSet
	{
		var dialOpts []grpc.DialOption
		dialOpts, err = extgrpc.StoreClientGRPCOpts(logKit, opt.reg, opt.tracer, false, false, "", "", "", "")
		if err != nil {
			return errors.Wrap(err, "building gRPC client")
		}

		// 现在的模式 thanos_store_nodes_grpc_connections metric 会有大量的 external_labels 且无实际用途, 使用一个临时 reg drop 掉
		_reg := prometheus.NewRegistry()

		endpoints = query.NewEndpointSet(logKit, _reg,
			func() (specs []*query.GRPCEndpointSpec) {
				for _, addr := range gw.GetStoreAddrs() {
					specs = append(specs, query.NewGRPCEndpointSpec(addr, true))
				}
				return specs
			},
			dialOpts,
			time.Second*30,
		)

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

	// proxyStore grpc 服务
	registryProxyStore(g, logKit, opt, endpoints)

	// 自定义 store grpc 服务
	{
		g.Add(func() error {
			return gw.Run()
		}, func(err error) {
			gw.Shutdown(err)
		})
	}

	bcs.CacheListClusters()

	return err
}

// registryProxyStore registry proxy store
func registryProxyStore(g *run.Group, logKit blog.GlogKit, opt *option, endpoints *query.EndpointSet) {

	proxyStore := store.NewProxyStore(logKit, opt.reg, endpoints.GetStoreClients, component.Query, nil,
		time.Minute*2)
	grpcProbe := prober.NewGRPC()
	grpcSrv := grpcserver.New(logKit, opt.reg, nil, nil, nil, component.Store, grpcProbe,
		grpcserver.WithServer(store.RegisterStoreServer(proxyStore)),
		grpcserver.WithListen(utils.GetListenAddr(bindAddress, grpcPort)),
		grpcserver.WithGracePeriod(time.Duration(0)),
		grpcserver.WithMaxConnAge(time.Minute*5), // 5分钟主动重连, pod 扩容等需要
	)

	g.Add(func() error {
		grpcProbe.Healthy()
		grpcProbe.Ready()

		return grpcSrv.ListenAndServe()
	}, func(err error) {
		defer grpcProbe.NotHealthy(err)
		defer grpcProbe.NotReady(err)

		grpcSrv.Shutdown(err)
	})
}
