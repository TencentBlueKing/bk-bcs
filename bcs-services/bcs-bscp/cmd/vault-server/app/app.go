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

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/vault-server/options"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/vault-server/service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	pbvs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/vault-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/brpc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Run start the config server
func Run(opt *options.Option) error {
	as := new(vaultService)
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

type vaultService struct {
	serve   *grpc.Server
	gwServe *http.Server
	service *service.Service
	sd      serviced.ServiceDiscover
}

// prepare do prepare jobs before run config server.
func (as *vaultService) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.VaultServer().Log.Logs())

	logs.Infof("load settings from config file success.")

	// init metrics
	metrics.InitMetrics(net.JoinHostPort(cc.VaultServer().Network.BindIP,
		strconv.Itoa(int(cc.VaultServer().Network.RpcPort))))

	etcdOpt, err := cc.VaultServer().Service.Etcd.ToConfig()
	if err != nil {
		return fmt.Errorf("get etcd config failed, err: %v", err)
	}

	// register vault server.
	svcOpt := serviced.ServiceOption{
		Name: cc.VaultServerName,
		IP:   cc.VaultServer().Network.BindIP,
		Port: cc.VaultServer().Network.RpcPort,
		Uid:  uuid.UUID(),
	}
	sd, err := serviced.NewServiceD(etcdOpt, svcOpt)
	if err != nil {
		return fmt.Errorf("new service discovery faield, err: %v", err)
	}

	as.sd = sd
	logs.Infof("create service discovery success.")

	return nil
}

// listenAndServe listen the grpc serve and set up the shutdown gracefully job.
func (as *vaultService) listenAndServe() error {
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
	network := cc.VaultServer().Network
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
	svc, err := service.NewService(as.sd)
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}
	pbvs.RegisterVaultServer(serve, svc)

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
		logs.Infof("start shutdown vault server grpc server gracefully...")

		as.serve.GracefulStop()
		notifier.Done()

		logs.Infof("shutdown vault server grpc server success...")

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
func (as *vaultService) gwListenAndServe() error {
	network := cc.VaultServer().Network
	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.HttpPort)))

	handler, err := as.service.Handler()
	if err != nil {
		return err
	}
	as.gwServe = &http.Server{Addr: addr, Handler: handler}

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		logs.Infof("start shutdown vault server http server gracefully...")

		_ = as.gwServe.Close()
		notifier.Done()

		logs.Infof("shutdown vault server http server success...")
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
			if err := as.gwServe.ListenAndServeTLS("", ""); err != nil {
				logs.Errorf("gateway https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()

	} else {
		go func() {
			if err := as.gwServe.ListenAndServe(); err != nil {
				logs.Errorf("gateway http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}

	logs.Infof("listen gateway server at %s now.", addr)

	return nil
}

func (as *vaultService) finalizer() {
	if err := as.sd.Deregister(); err != nil {
		logs.Errorf("process service shutdown, but deregister failed, err: %v", err)
		return
	}

	logs.Infof("shutting down service, deregister service success.")
}

// register the grpc serve.
func (as *vaultService) register() error {
	if err := as.sd.Register(); err != nil {
		return fmt.Errorf("register vault server failed, err: %v", err)
	}

	logs.Infof("register vault server to etcd success.")
	return nil
}
