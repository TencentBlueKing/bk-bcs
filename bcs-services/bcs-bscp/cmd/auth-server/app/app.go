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

// Package app NOTES
package app

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/options"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	pbas "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/brpc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/ctl"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/ctl/cmd"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Run start the config server
func Run(opt *options.Option) error {
	as := new(authService)
	if err := as.prepare(opt); err != nil {
		return err
	}

	if err := as.listenAndServe(); err != nil {
		return err
	}

	if err := as.gwListenAndServe(); err != nil {
		return err
	}

	if err := as.register(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(as.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

type authService struct {
	serve   *grpc.Server
	gwServe *http.Server
	service *service.Service
	sd      serviced.ServiceDiscover
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
	// disableWriteOpt defines which biz's write operation needs to be disabled
	disableWriteOpt *options.DisableWriteOption
}

// prepare do prepare jobs before run config server.
func (as *authService) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.AuthServer().Log.Logs())

	logs.Infof("load settings from config file success.")

	// init metrics
	metrics.InitMetrics(net.JoinHostPort(cc.AuthServer().Network.BindIP,
		strconv.Itoa(int(cc.AuthServer().Network.RpcPort))))

	etcdOpt, err := cc.AuthServer().Service.Etcd.ToConfig()
	if err != nil {
		return fmt.Errorf("get etcd config failed, err: %v", err)
	}

	// register auth server.
	svcOpt := serviced.ServiceOption{
		Name: cc.AuthServerName,
		IP:   cc.AuthServer().Network.BindIP,
		Port: cc.AuthServer().Network.RpcPort,
		Uid:  uuid.UUID(),
	}
	sd, err := serviced.NewServiceD(etcdOpt, svcOpt)
	if err != nil {
		return fmt.Errorf("new service discovery faield, err: %v", err)
	}

	as.sd = sd
	logs.Infof("create service discovery success.")

	as.disableWriteOpt = &options.DisableWriteOption{
		IsDisabled: false,
		IsAll:      false,
		BizIDMap:   sync.Map{},
	}

	// init bscp control tool
	if err := ctl.LoadCtl(append(ctl.WithBasics(sd), cmd.WithWrites(as.disableWriteOpt)...)...); err != nil {
		return fmt.Errorf("load control tool failed, err: %v", err)
	}

	as.disableAuth = opt.DisableAuth
	if opt.DisableAuth {
		logs.Infof("authorize function is disabled.")
	}

	return nil
}

// listenAndServe listen the grpc serve and set up the shutdown gracefully job.
func (as *authService) listenAndServe() error {
	// generate standard grpc server grpcMetrics.
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcMetrics.EnableHandlingTimeHistogram(metrics.GrpcBuckets)

	recoveryOpt := grpc_recovery.WithRecoveryHandlerContext(brpc.RecoveryHandlerFuncContext)

	opts := []grpc.ServerOption{grpc.MaxRecvMsgSize(math.MaxInt32),
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
	}
	network := cc.AuthServer().Network
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
	svc, err := service.NewService(as.sd, cc.AuthServer().IAM, as.disableAuth, as.disableWriteOpt)
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}
	pbas.RegisterAuthServer(serve, svc)

	// initialize and register standard grpc server grpcMetrics.
	grpcMetrics.InitializeMetrics(serve)
	if err = metrics.Register().Register(grpcMetrics); err != nil {
		return fmt.Errorf("register metrics failed, err: %v", err)
	}

	as.service = svc
	as.serve = serve

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		logs.Infof("start shutdown auth server grpc server gracefully...")

		as.serve.GracefulStop()
		notifier.Done()

		logs.Infof("shutdown auth server grpc server success...")
	}()

	addr := tools.GetListenAddr(network.BindIP, int(network.RpcPort))
	ipv6Addr := tools.GetListenAddr(network.BindIPv6, int(network.RpcPort))
	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(addr); err != nil {
		return err
	}

	if network.BindIPv6 != "" && network.BindIPv6 != network.BindIP {
		if err := dualStackListener.AddListenerWithAddr(ipv6Addr); err != nil {
			return err
		}
		logs.Infof("grpc serve dualStackListener with ipv6: %s", ipv6Addr)
	}

	go func() {
		if err := serve.Serve(dualStackListener); err != nil {
			logs.Errorf("serve grpc server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()

	logs.Infof("listen grpc server at %s now.", addr)

	return nil
}

// gwListenAndServe listen the http serve and set up the shutdown gracefully job.
func (as *authService) gwListenAndServe() error {
	network := cc.AuthServer().Network
	addr := tools.GetListenAddr(network.BindIP, int(network.HttpPort))
	ipv6Addr := tools.GetListenAddr(network.BindIPv6, int(network.HttpPort))
	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(addr); err != nil {
		return err
	}

	if network.BindIPv6 != "" && network.BindIPv6 != network.BindIP {
		if err := dualStackListener.AddListenerWithAddr(ipv6Addr); err != nil {
			return err
		}
		logs.Infof("api serve dualStackListener with ipv6: %s", ipv6Addr)
	}
	handler, err := as.service.Handler()
	if err != nil {
		return err
	}
	as.gwServe = &http.Server{Addr: addr, Handler: handler}

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		logs.Infof("start shutdown auth server http server gracefully...")

		_ = as.gwServe.Close()
		notifier.Done()

		logs.Infof("shutdown auth server http server success...")
	}()

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init tls config failed, err: %v", err)
		}

		as.gwServe.TLSConfig = tlsC

		go func() {
			if err := as.gwServe.ServeTLS(dualStackListener, "", ""); err != nil {
				logs.Errorf("gateway https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()

	} else {
		go func() {
			if err := as.gwServe.Serve(dualStackListener); err != nil {
				logs.Errorf("gateway http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}

	logs.Infof("listen gateway server at %s now.", addr)

	return nil
}

func (as *authService) finalizer() {
	if err := as.sd.Deregister(); err != nil {
		logs.Errorf("process service shutdown, but deregister failed, err: %v", err)
		return
	}

	logs.Infof("shutting down service, deregister service success.")
}

// register the grpc serve.
func (as *authService) register() error {
	if err := as.sd.Register(); err != nil {
		return fmt.Errorf("register auth server failed, err: %v", err)
	}

	logs.Infof("register auth server to etcd success.")
	return nil
}
