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

// Package server xxx
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	grpccli "github.com/asim/go-micro/plugins/client/grpc/v4"
	"github.com/asim/go-micro/plugins/registry/etcd/v4"
	grpcsvr "github.com/asim/go-micro/plugins/server/grpc/v4"
	etcdsync "github.com/asim/go-micro/plugins/sync/etcd/v4"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/sync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/controller"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/resourcemgr"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
	mongstore "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage/mongo"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
)

const (
	defaultServiceName = "nodegroupmanager.bkbcs.tencent.com"
	gracefulPeriod     = 3
)

// NewServer create server instance
func NewServer(opt *Options) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		opt:        opt,
		svrContext: ctx,
		svrCancel:  cancel,
	}
}

// Server nodegroup manager entry
type Server struct {
	// server options
	opt *Options
	// context control for exit
	svrContext context.Context
	svrCancel  context.CancelFunc
	// etcdSync for leader election
	etcdSync sync.Sync
	// localStorage for database connection
	localStorage storage.Storage
	// resourceMgr resource-manager client
	resourceMgr resourcemgr.Client
	// clusterCli cluster client
	clusterCli cluster.Client
	// gatewayServer for grpc gateway
	gatewayServer http.Server
	// go-micro v4 grpc server
	microService micro.Service
	// controller for NodeGroup management
	nodeGroupCtl controller.Controller
	// taskController for task management
	taskController  controller.Controller
	nodeGroupCancel context.CancelFunc
	// extraServer for metric/pprof
	extraServer http.Server
}

// Init essential elements for running
func (s *Server) Init() error {
	initializer := []func() error{
		s.initStorage, s.initResourceMgr, s.initClusterCli,
		s.initMicroService, s.initGatewayServer, s.initExtraServers, s.initLeaderElection,
	}

	for _, initFunc := range initializer {
		if err := initFunc(); err != nil {
			return err
		}
	}
	return nil
}

// Run all background services and block
// all sub runner function are non-block.
// it's fatal if sub runner failed.
func (s *Server) Run() error {
	runners := []func(){
		s.startMicroService, s.startGatewayServer,
		s.startSignalHandler, s.startLeaderElection,
	}

	for _, run := range runners {
		time.Sleep(time.Second)
		go run()
	}

	<-s.svrContext.Done()
	blog.Infof("server is under graceful period, %d seconds...", gracefulPeriod)
	time.Sleep(time.Second * gracefulPeriod)
	return nil
}

// startGatewayServer start http server & metric server
func (s *Server) startGatewayServer() {
	err := s.gatewayServer.ListenAndServeTLS("", "")
	if err != nil {
		if http.ErrServerClosed == err {
			blog.Warnf("grpc http gateway graceful exit.")
			return
		}
		// start http gateway error, maybe port is conflict or something else
		blog.Fatal("server start grpc http gateway fatal, %s", err.Error())
	}
}

// stopGatewayServer  gracefully stop
func (s *Server) stopGatewayServer() {
	cxt, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if err := s.gatewayServer.Shutdown(cxt); err != nil {
		blog.Errorf("server gracefully shutdown grpc http gateway failure: %s", err.Error())
		return
	}
	blog.Infof("server shutdown grpc http gateway gracefully")
}

// startMicroService start go-micro service
func (s *Server) startMicroService() {
	// Run service
	if err := s.microService.Run(); err != nil {
		blog.Fatal("server start microservice fatal, %s", err.Error())
	}
}

// startSignalHandler trap system signal for exit
func (s *Server) startSignalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-ch
	blog.Infof("server traps terminated, signal: %s", sig.String())
	s.stopGatewayServer()
	// server cancel func cover go-micro/leaderElection
	s.svrCancel()
	tick := time.NewTicker(time.Second * 2)
	select {
	case doubleKill := <-ch:
		blog.Warnf("server trap system signal %s again, exit immediately", doubleKill.String())
		blog.CloseLogs()
		os.Exit(-1)
	case <-tick.C:
		return
	}
}

// startLeaderElection choose leader for
// starting controller logic
func (s *Server) startLeaderElection() {
	cxt, cancel := context.WithCancel(s.svrContext)
	defer cancel()
	leaderID := defaultServiceName
	leader, err := s.etcdSync.Leader(leaderID)
	if err != nil {
		blog.Errorf("server %s campaign leader role met error, %s", leaderID, err.Error())
		// maybe network error, server can start elect
		// again after backoff strategy
		time.Sleep(time.Second * gracefulPeriod)
		go s.startLeaderElection()
		return
	}
	blog.Infof("server %s become leader, starting controller...", leaderID)
	lost := leader.Status()
	// server campaign leader successfully, start all controllers.
	// controller must start successfully or there is bug
	if err := s.startController(); err != nil {
		blog.Fatalf("server start controller fatal, %s", err.Error())
		return
	}
	// start controller successfully, check leader status continuously.
	// when server lost leader role, stop controller
	select {
	case <-lost:
		s.stopController()
		go s.startLeaderElection()
	case <-cxt.Done():
		// server exit
		if err := leader.Resign(); err != nil {
			blog.Errorf("server %s resign leader failure, %s, other servers wait until timeout",
				leaderID, err.Error())
			return
		}
		blog.Infof("server %s resign leader successfully", leaderID)
	}
}

// startController start essential controller when campaign leader successfully
func (s *Server) startController() error {
	option := &controller.Options{
		Interval:        s.opt.ControllerLoop,
		Storage:         s.localStorage,
		ResourceManager: s.resourceMgr,
		ClusterClient:   s.clusterCli,
	}
	s.nodeGroupCtl = controller.NewController(option)
	if err := s.nodeGroupCtl.Init(); err != nil {
		return err
	}
	s.taskController = controller.NewTaskController(option)
	if err := s.taskController.Init(); err != nil {
		return err
	}
	cxt, cancel := context.WithCancel(s.svrContext)
	go s.nodeGroupCtl.Run(cxt)
	go s.taskController.Run(cxt)
	s.nodeGroupCancel = cancel
	return nil
}

// stopController stop essential controller when lost leader
func (s *Server) stopController() {
	// stop NodeGroupController
	s.nodeGroupCancel()

	// stop more controller
}

// initStorage for database
func (s *Server) initStorage() error {
	// convert to mongo options
	option := &mongo.Options{
		Hosts:                 strings.Split(s.opt.Storage.Endpoints, ","),
		ConnectTimeoutSeconds: 3,
		Database:              s.opt.Storage.Database,
		Username:              s.opt.Storage.UserName,
		Password:              s.opt.Storage.Password,
	}
	mInst, err := mongo.NewDB(option)
	if err != nil {
		return fmt.Errorf("storage create mongo instance failed, %s", err.Error())
	}
	if err := mInst.Ping(); err != nil {
		return fmt.Errorf("storage connection test failed, %s", err.Error())
	}
	s.localStorage = mongstore.NewServer(mInst)
	return nil
}

// initResourceMgr client for resource-manager
func (s *Server) initResourceMgr() error {
	opt := &resourcemgr.ClientOptions{
		Name:         s.opt.ResourceManager,
		Etcd:         strings.Split(s.opt.Registry.Endpoints, ","),
		EtcdConfig:   s.opt.Registry.tlsConfig,
		ClientConfig: s.opt.clientTLS,
	}
	s.resourceMgr = resourcemgr.New(opt)
	return nil
}

// initMicroService init grpc service
func (s *Server) initMicroService() error {
	// context for microservice
	ctx, _ := context.WithCancel(s.svrContext) // nolint
	// registry init
	hosts := strings.Split(s.opt.Registry.Endpoints, ",")
	globalRegistry := etcd.NewRegistry(
		registry.Addrs(hosts...),
		registry.TLSConfig(s.opt.Registry.tlsConfig),
	)
	// Create service
	s.microService = micro.NewService(
		micro.Server(grpcsvr.NewServer(
			grpcsvr.AuthTLS(s.opt.serverTLS),
		)),
		micro.Client(grpccli.NewClient(
			grpccli.AuthTLS(s.opt.clientTLS),
		)),
		micro.Name(defaultServiceName),
		// context for exit control
		micro.Context(ctx),
		micro.Metadata(map[string]string{
			"httpport": fmt.Sprintf("%d", s.opt.HTTPPort),
			// feature(DeveloperJim): IPv6 supporte
		}),
		micro.Address(fmt.Sprintf("%s:%d", s.opt.Address, s.opt.Port)),
		micro.Version(version.BcsVersion),
		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(25*time.Second),
		micro.Registry(globalRegistry),
	)
	// nothing to init
	// s.microService.Init()
	// Register handler
	if err := pb.RegisterNodegroupManagerHandler(s.microService.Server(), handler.New(s.localStorage)); err != nil {
		return fmt.Errorf("micro service registry handle failed, %s", err.Error())
	}
	return nil
}

// initGatewayServer init http gateway for grpc service proxy
func (s *Server) initGatewayServer() error {
	ctx, _ := context.WithCancel(s.svrContext) // nolint
	// register grpc server information
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(s.opt.clientTLS))}
	err := pb.RegisterNodegroupManagerGwFromEndpoint(
		ctx, mux, fmt.Sprintf("%s:%d", s.opt.Address, s.opt.Port), opts)
	if err != nil {
		return fmt.Errorf("http gateway register grpc service failed, %s", err.Error())
	}
	s.gatewayServer = http.Server{
		Addr:      fmt.Sprintf("%s:%d", s.opt.Address, s.opt.HTTPPort),
		Handler:   mux,
		TLSConfig: s.opt.serverTLS,
	}
	return nil
}

// initLeaderElection init etcd leader election impelementation
func (s *Server) initLeaderElection() error {
	hosts := strings.Split(s.opt.Registry.Endpoints, ",")
	// feature(DeveloperJim): tracing Sync.Leader implementation
	// for Context control
	s.etcdSync = etcdsync.NewSync(
		sync.WithTLS(s.opt.Registry.tlsConfig),
		sync.Nodes(hosts...),
	)
	return s.etcdSync.Init()
}

func (s *Server) initClusterCli() error {
	realAuthToken, err := encrypt.DesDecryptFromBase([]byte(s.opt.Gateway.Token))
	if err != nil {
		return fmt.Errorf("init clusterCli failed, encrypt token error:%s", err.Error())
	}
	clusterOpts := &cluster.ClusterClientOptions{
		Endpoint: s.opt.Gateway.Endpoint,
		Token:    string(realAuthToken),
		Sender:   requester.NewRequester(),
	}
	clusterManagerOpts := &cluster.ClusterManagerClientOptions{
		Name:         s.opt.ClusterManager,
		Etcd:         strings.Split(s.opt.Registry.Endpoints, ","),
		EtcdConfig:   s.opt.Registry.tlsConfig,
		ClientConfig: s.opt.clientTLS,
	}
	s.clusterCli = cluster.NewClient(clusterOpts, clusterManagerOpts)
	return nil
}

func (s *Server) initExtraServers() error {
	extraMux := http.NewServeMux()
	s.initMetric(extraMux)
	extraServerEndpoint := s.opt.Address + ":" + strconv.Itoa(int(s.opt.MetricPort))
	s.extraServer = http.Server{
		Addr:    extraServerEndpoint,
		Handler: extraMux,
	}

	go func() {
		var err error
		blog.Infof("start extra modules [pprof, metric] server %s", extraServerEndpoint)
		err = s.extraServer.ListenAndServe()
		if err != nil {
			blog.Errorf("extra modules server listen failed, err %s", err.Error())
			s.svrCancel()
		}
	}()
	return nil
}

func (s *Server) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}
