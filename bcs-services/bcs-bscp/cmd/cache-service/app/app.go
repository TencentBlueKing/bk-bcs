/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package app NOTES
package app

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"bscp.io/cmd/cache-service/options"
	"bscp.io/cmd/cache-service/service"
	"bscp.io/cmd/cache-service/service/cache/client"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/uuid"
	"bscp.io/pkg/dal/bedis"
	"bscp.io/pkg/dal/dao"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	pbcs "bscp.io/pkg/protocol/cache-service"
	"bscp.io/pkg/runtime/brpc"
	"bscp.io/pkg/runtime/ctl"
	"bscp.io/pkg/runtime/ctl/cmd"
	"bscp.io/pkg/runtime/shutdown"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
)

// Run start the cache service
func Run(opt *options.Option) error {
	ds := new(cacheService)
	if err := ds.prepare(opt); err != nil {
		return err
	}

	if err := ds.listenAndServe(); err != nil {
		return err
	}

	if err := ds.gwListenAndServe(); err != nil {
		return err
	}

	if err := ds.register(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(ds.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

type cacheService struct {
	serve   *grpc.Server
	gwServe *http.Server
	service *service.Service
	sd      serviced.ServiceDiscover
	daoSet  dao.Set
	bds     bedis.Client
	op      client.Interface
}

// prepare do prepare jobs before run cache service.
func (cs *cacheService) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.CacheService().Log.Logs())

	logs.Infof("load settings from config file success.")

	// init metrics
	metrics.InitMetrics(net.JoinHostPort(cc.CacheService().Network.BindIP,
		strconv.Itoa(int(cc.CacheService().Network.RpcPort))))

	etcdOpt, err := cc.CacheService().Service.Etcd.ToConfig()
	if err != nil {
		return fmt.Errorf("get etcd config failed, err: %v", err)
	}

	// register cache service.
	svcOpt := serviced.ServiceOption{
		Name: cc.CacheServiceName,
		IP:   cc.CacheService().Network.BindIP,
		Port: cc.CacheService().Network.RpcPort,
		Uid:  uuid.UUID(),
	}
	sd, err := serviced.NewServiceD(etcdOpt, svcOpt)
	if err != nil {
		return fmt.Errorf("new service discovery faield, err: %v", err)
	}

	cs.sd = sd

	// init redis client
	bds, err := bedis.NewRedisCache(cc.CacheService().RedisCluster)
	if err != nil {
		return fmt.Errorf("new redis cluster failed, err: %v", err)
	}
	cs.bds = bds

	// initial DAO set
	set, err := dao.NewDaoSet(cc.CacheService().Sharding, cc.CacheService().Credential)
	if err != nil {
		return fmt.Errorf("initial dao set failed, err: %v", err)
	}

	cs.daoSet = set

	// init bscp control tool
	cs.op, err = client.New(set, bds)
	if err != nil {
		return fmt.Errorf("new cache client failed, err: %v", err)
	}

	if err := ctl.LoadCtl(append(ctl.WithBasics(sd), cmd.WithRefreshCache(cs.op))...); err != nil {
		return fmt.Errorf("load control tool failed, err: %v", err)
	}

	return nil
}

// listenAndServe listen the grpc serve and set up the shutdown gracefully job.
func (cs *cacheService) listenAndServe() error {
	// generate standard grpc server grpcMetrics.
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcMetrics.EnableHandlingTimeHistogram(metrics.GrpcBuckets)

	recoveryOpt := grpc_recovery.WithRecoveryHandlerContext(brpc.RecoveryHandlerFuncContext)

	opts := []grpc.ServerOption{grpc.MaxRecvMsgSize(4 * 1024 * 1024),
		// add bscp unary interceptor and standard grpc server metrics interceptor.
		grpc.ChainUnaryInterceptor(
			brpc.LogUnaryServerInterceptor(),
			grpcMetrics.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(recoveryOpt),
		),
		grpc.ChainStreamInterceptor(
			grpcMetrics.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(recoveryOpt),
		),
		grpc.ReadBufferSize(8 * 1024 * 1024),
		grpc.WriteBufferSize(16 * 1024 * 1024),
		grpc.InitialConnWindowSize(16 * 1024 * 1024),
	}

	network := cc.CacheService().Network
	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.Creds(cred))
	}

	serve := grpc.NewServer(opts...)
	svc, err := service.NewService(cs.sd, cs.daoSet, cs.bds, cs.op)
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}
	pbcs.RegisterCacheServer(serve, svc)

	// initialize and register standard grpc server grpcMetrics.
	grpcMetrics.InitializeMetrics(serve)
	if err = metrics.Register().Register(grpcMetrics); err != nil {
		return fmt.Errorf("register metrics failed, err: %v", err)
	}

	cs.service = svc
	cs.serve = serve

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			logs.Infof("start shutdown cache service grpc server gracefully...")

			cs.serve.GracefulStop()
			notifier.Done()

			logs.Infof("shutdown cache service grpc server success...")

		}
	}()

	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.RpcPort)))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("lisen addr: %s failed, err: %v", addr, err)
	}

	go func() {
		if err := serve.Serve(listener); err != nil {
			logs.Errorf("serve grpc server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()

	logs.Infof("listen grpc server at %s now.", addr)

	return nil
}

// gwListenAndServe listen the http serve and set up the shutdown gracefully job.
func (cs *cacheService) gwListenAndServe() error {
	network := cc.CacheService().Network
	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.HttpPort)))

	handler, err := cs.service.Handler()
	if err != nil {
		return err
	}

	cs.gwServe = &http.Server{Addr: addr, Handler: handler}

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			logs.Infof("start shutdown cache service http server gracefully...")

			cs.gwServe.Close()
			notifier.Done()

			logs.Infof("shutdown cache service http server success...")
		}
	}()

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init tls config failed, err: %v", err)
		}

		cs.gwServe.TLSConfig = tlsC

		go func() {
			if err := cs.gwServe.ListenAndServeTLS("", ""); err != nil {
				logs.Errorf("gateway https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	} else {
		go func() {
			if err := cs.gwServe.ListenAndServe(); err != nil {
				logs.Errorf("gateway http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}

	logs.Infof("listen gateway server at %s now.", addr)

	return nil
}

func (cs *cacheService) finalizer() {
	if err := cs.sd.Deregister(); err != nil {
		logs.Errorf("process service shutdown, but deregister failed, err: %v", err)
		return
	}

	logs.Infof("shutting down service, deregister service success.")
}

// register the grpc serve.
func (cs *cacheService) register() error {
	if err := cs.sd.Register(); err != nil {
		return fmt.Errorf("register cache service failed, err: %v", err)
	}

	logs.Infof("register cache service to etcd success.")
	return nil
}
