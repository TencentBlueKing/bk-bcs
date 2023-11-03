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
	"os"
	"strconv"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/hashicorp/vault/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"bscp.io/cmd/data-service/options"
	"bscp.io/cmd/data-service/service"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/uuid"
	"bscp.io/pkg/dal/dao"
	"bscp.io/pkg/dal/vault"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/brpc"
	"bscp.io/pkg/runtime/ctl"
	"bscp.io/pkg/runtime/shutdown"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
)

// RunServer run the data service
func RunServer(sysOpt *cc.SysOption) {
	opts := options.InitOptions(sysOpt)
	if err := Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "start data service failed, err: %v", err)
		logs.CloseLogs()
		os.Exit(1)
	}
}

// Run start the data service
func Run(opt *options.Option) error {
	ds := new(dataService)
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

type dataService struct {
	serve   *grpc.Server
	gwServe *http.Server
	service *service.Service
	sd      serviced.Service
	daoSet  dao.Set
	vault   vault.Set
}

// prepare do prepare jobs before run data service.
func (ds *dataService) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.DataService().Log.Logs())

	logs.Infof("load settings from config file success.")

	// init metrics
	metrics.InitMetrics(net.JoinHostPort(cc.DataService().Network.BindIP,
		strconv.Itoa(int(cc.DataService().Network.RpcPort))))

	etcdOpt, err := cc.DataService().Service.Etcd.ToConfig()
	if err != nil {
		return fmt.Errorf("get etcd config failed, err: %v", err)
	}

	// register data service.
	svcOpt := serviced.ServiceOption{
		Name: cc.DataServiceName,
		IP:   cc.DataService().Network.BindIP,
		Port: cc.DataService().Network.RpcPort,
		Uid:  uuid.UUID(),
	}
	sd, err := serviced.NewService(etcdOpt, svcOpt)
	if err != nil {
		return fmt.Errorf("new service faield, err: %v", err)
	}

	ds.sd = sd

	// init bscp control tool
	if err = ctl.LoadCtl(ctl.WithBasics(sd)...); err != nil {
		return fmt.Errorf("load control tool failed, err: %v", err)
	}

	// initial DAO set
	set, err := dao.NewDaoSet(cc.DataService().Sharding, cc.DataService().Credential)
	if err != nil {
		return fmt.Errorf("initial dao set failed, err: %v", err)
	}

	ds.daoSet = set

	// initial Vault set
	vaultSet, err := vault.NewSet(cc.DataService().Vault)
	if err != nil {
		return fmt.Errorf("initial vault set failed, err: %v", err)
	}
	// 挂载目录
	exists, err := vaultSet.IsMountPathExists(vault.MountPath)
	if err != nil {
		return fmt.Errorf("error checking mount path: %v", err)
	}
	if !exists {
		mountConfig := &api.MountInput{
			Type: "kv-v2",
		}
		if err = vaultSet.CreateMountPath(vault.MountPath, mountConfig); err != nil {
			return fmt.Errorf("initial vault mount path failed, err: %v", err)
		}
	}

	ds.vault = vaultSet

	return nil
}

// listenAndServe listen the grpc serve and set up the shutdown gracefully job.
func (ds *dataService) listenAndServe() error {
	// generate standard grpc server grpcMetrics.
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcMetrics.EnableHandlingTimeHistogram(metrics.GrpcBuckets)

	recoveryOpt := grpc_recovery.WithRecoveryHandlerContext(brpc.RecoveryHandlerFuncContext)

	opts := []grpc.ServerOption{grpc.MaxRecvMsgSize(math.MaxInt32),
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

	network := cc.DataService().Network
	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init grpc tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.Creds(cred))
	}

	serve := grpc.NewServer(opts...)
	svc, err := service.NewService(ds.sd, ds.daoSet, ds.vault)
	if err != nil {
		return err
	}

	pbds.RegisterDataServer(serve, svc)

	// initialize and register standard grpc server grpcMetrics.
	grpcMetrics.InitializeMetrics(serve)
	if err = metrics.Register().Register(grpcMetrics); err != nil {
		return fmt.Errorf("register metrics failed, err: %v", err)
	}

	ds.service = svc
	ds.serve = serve

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		logs.Infof("start shutdown grpc server gracefully...")

		ds.serve.GracefulStop()
		notifier.Done()

		logs.Infof("shutdown grpc server success...")

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

func (ds *dataService) finalizer() {
	if err := ds.sd.Deregister(); err != nil {
		logs.Errorf("process service shutdown, but deregister failed, err: %v", err)
		return
	}

	logs.Infof("shutting down service, deregister service success.")
}

// gwListenAndServe listen the http serve and set up the shutdown gracefully job.
func (ds *dataService) gwListenAndServe() error {
	network := cc.DataService().Network
	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.HttpPort)))

	handler, err := ds.service.Handler()
	if err != nil {
		return err
	}

	ds.gwServe = &http.Server{Addr: addr, Handler: handler}

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		logs.Infof("start shutdown data service http server gracefully...")

		_ = ds.gwServe.Close()
		notifier.Done()

		logs.Infof("shutdown data service http server success...")
	}()

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init tls config failed, err: %v", err)
		}

		ds.gwServe.TLSConfig = tlsC

		go func() {
			if err := ds.gwServe.ListenAndServeTLS("", ""); err != nil {
				logs.Errorf("gateway https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	} else {
		go func() {
			if err := ds.gwServe.ListenAndServe(); err != nil {
				logs.Errorf("gateway http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}

	logs.Infof("listen gateway server at %s now.", addr)

	return nil
}

// register the grpc serve.
func (ds *dataService) register() error {
	if err := ds.sd.Register(); err != nil {
		return fmt.Errorf("register service failed, err: %v", err)
	}

	logs.Infof("register data service to etcd success.")
	return nil
}
