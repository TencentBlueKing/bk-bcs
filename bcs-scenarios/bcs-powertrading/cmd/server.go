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

package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/go-micro/plugins/v4/registry/etcd"
	microgrpcsvc "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	microsvc "go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	grpccred "google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/handler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bkcc"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bksops"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/cr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/job"
	requester2 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/resourcemgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes/data"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes/supply"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
	mongostore "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage/mongo"
	powertrading "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/proto"
)

const (
	// RpcServerName the name of PredictApproval server
	RpcServerName = "powertrading.bkbcs.tencent.com"
	// RegisterTTL register ttl
	RegisterTTL = 20
	// RegisterInterval register interval
	RegisterInterval = 10
)

// Server powertrading server
type Server struct {
	op *Options

	httpServer      *http.Server
	rpcServer       microsvc.Service
	metricServer    *http.Server
	metricListeners []net.Listener

	internalCtx    context.Context
	internalCancel context.CancelFunc
	internalWg     *sync.WaitGroup

	cmCli          clustermgr.Client
	rmCli          resourcemgr.Client
	ccCli          bkcc.Client
	bksopsCli      bksops.Client
	crCli          cr.Client
	jobCli         job.Client
	storage        storage.Storage
	taskController scenes.Controller
	dataController scenes.Controller
	dataService    data.Service
}

// NewServer new server
func NewServer(opt *Options) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		op:             opt,
		internalCtx:    ctx,
		internalCancel: cancel,
	}
}

// Init will parse options, and init grpc_server/http_server/metric_server
func (s *Server) Init(ctx context.Context) error {
	if err := s.initClient(); err != nil {
		return errors.Wrapf(err, "init clients failed")
	}
	if err := s.initStorage(); err != nil {
		return errors.Wrapf(err, "init storage failed")
	}
	s.initDataService()
	s.initMetricServer()
	if err := s.initHTTPServer(); err != nil {
		return errors.Wrapf(err, "init http server failed")
	}
	if err := s.initRpcServer(); err != nil {
		return errors.Wrapf(err, "init rpc server failed")
	}
	if err := s.initTaskController(); err != nil {
		return errors.Wrapf(err, "initTaskController failed")
	}
	if err := s.initDataController(); err != nil {
		return errors.Wrapf(err, "initDataController failed")
	}
	return nil
}

// Run all the server that had init
func (s *Server) Run() error {
	errChan := s.run(s.runHttpServer, s.runMetricServer, s.runRpcServer, s.runTaskController, s.runDataController)

	defer func() {
		s.internalCancel()
		s.internalWg.Wait()
	}()
	for {
		select {
		case err := <-errChan:
			return errors.Wrapf(err, "server exit with error")
		// case <-s.cfgHandler.ChangedChan():
		//	s.handleConfigChanged()
		case <-s.internalCtx.Done():
			return nil
		}
	}
}

// run will run goroutines and add wait group for them
func (s *Server) run(fs ...func(errChan chan error)) chan error {
	s.internalWg = &sync.WaitGroup{}
	s.internalWg.Add(len(fs))
	errChan := make(chan error, len(fs))
	for i := range fs {
		go fs[i](errChan)
	}
	return errChan
}

// initPProf init pprof
func (s *Server) initPProf(mux *http.ServeMux) {
	if !s.op.Debug {
		blog.Infof("pprof is disabled")
		return
	}
	blog.Infof("pprof is enabled")
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// initMetric init metric
func (s *Server) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}

// init metric server and pprof
func (s *Server) initMetricServer() {
	metricMux := http.NewServeMux()
	s.initPProf(metricMux)
	s.initMetric(metricMux)
	extraServerEndpoint := s.op.Address + ":" + strconv.Itoa(int(s.op.MetricPort))
	s.metricServer = &http.Server{
		Addr:    extraServerEndpoint,
		Handler: metricMux,
	}
	go func() {
		var err error
		blog.Infof("start extra modules [pprof, metric] server %s", extraServerEndpoint)
		err = s.metricServer.ListenAndServe()
		if err != nil {
			blog.Errorf("extra modules server listen failed, err %s", err.Error())
			s.internalCancel()
		}
	}()
	blog.Infof("init metricServer success")
}

// initRpcServer init micro server
func (s *Server) initRpcServer() error {
	etcdRegistry := etcd.NewRegistry(
		registry.Addrs(strings.Split(s.op.Registry.Endpoints, ",")...),
		registry.TLSConfig(s.op.Registry.tlsConfig),
	)
	microServer := microsvc.NewService(
		microsvc.Server(microgrpcsvc.NewServer(microgrpcsvc.AuthTLS(s.op.serverTLS))),
		microsvc.Context(s.internalCtx),
		microsvc.Name(RpcServerName),
		microsvc.Metadata(map[string]string{
			"httpport": fmt.Sprintf("%d", s.op.HTTPPort),
		}),
		microsvc.Registry(etcdRegistry),
		// set address to fist ip, if ipv4/ipv6 all set
		microsvc.Address(fmt.Sprintf("%s:%d", s.op.Address, s.op.Port)),
		microsvc.RegisterTTL(RegisterTTL*time.Second),
		microsvc.RegisterInterval(RegisterInterval*time.Second),
		microsvc.Version(version.BcsVersion),
	)
	s.rpcServer = microServer
	fmt.Println(s.rpcServer.Name())

	if err := powertrading.RegisterPowerTradingHandler(s.rpcServer.Server(),
		handler.New(s.cmCli, s.rmCli, s.ccCli, s.storage)); err != nil {
		blog.Errorf("register rpc handler failed, err: %s", err.Error())
		return err
	}
	blog.Infof("init micro service success")
	return nil
}

// initHTTPServer init http server
func (s *Server) initHTTPServer() error {
	ctx, _ := context.WithCancel(s.internalCtx) // nolint
	// register grpc server information
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(grpccred.NewTLS(s.op.clientTLS))}
	err := powertrading.RegisterPowerTradingGwFromEndpoint(
		ctx, mux, fmt.Sprintf("%s:%d", s.op.Address, s.op.Port), opts)
	if err != nil {
		return fmt.Errorf("http gateway register grpc service failed, %s", err.Error())
	}
	s.httpServer = &http.Server{
		Addr:      fmt.Sprintf("%s:%d", s.op.Address, s.op.HTTPPort),
		Handler:   mux,
		TLSConfig: s.op.serverTLS,
	}
	blog.Infof("init http server success")
	return nil
}

// initStorage init db
func (s *Server) initStorage() error {
	// convert to mongo options
	storageOption := &mongo.Options{
		Hosts:                 strings.Split(s.op.Storage.Endpoints, ","),
		ConnectTimeoutSeconds: 3,
		Database:              s.op.Storage.Database,
		Username:              s.op.Storage.UserName,
		Password:              s.op.Storage.Password,
	}
	instance, err := mongo.NewDB(storageOption)
	if err != nil {
		return fmt.Errorf("storage create dm mongo instance failed, %s", err.Error())
	}
	if pingErr := instance.Ping(); pingErr != nil {
		return fmt.Errorf("dm storage connection test failed, %s", pingErr.Error())
	}
	s.storage = mongostore.NewServer(instance)
	blog.Infof("init storage success")
	return nil
}

// runMetricServer run metric server
func (s *Server) runMetricServer(errChan chan error) {
	defer s.internalWg.Done()
	errs := make(chan error, len(s.metricListeners))
	for _, listener := range s.metricListeners {
		go func(listener net.Listener) {
			if err := s.metricServer.Serve(listener); err != nil {
				errs <- errors.Wrapf(err, "serve metric listener '%s' failed", listener.Addr())
			}
		}(listener)
	}
	if err, ok := <-errs; ok {
		if err != nil {
			errChan <- err
		}
	}
}

// runHttpServer run http server
func (s *Server) runHttpServer(errChan chan error) {
	var err error
	defer func() {
		s.internalWg.Done()
		if r := recover(); r != nil {
			err = errors.Errorf("http_server panic, err: %v, stack:\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			blog.Errorf("http_server exited: %s", err.Error())
			errChan <- err
		} else {
			blog.Infof("http_server is stopped.")
		}
	}()
	if err = s.httpServer.ListenAndServeTLS("", ""); err != nil {
		err = errors.Wrapf(err, "serve http listener '%d' failed", s.op.HTTPPort)
	}
}

// runRpcServer run rpc server
func (s *Server) runRpcServer(errChan chan error) {
	var err error
	defer func() {
		s.internalWg.Done()
		if r := recover(); r != nil {
			err = errors.Errorf("rpc_server panic, err: %v, stack:\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			blog.Errorf("rpc_server exited: %s", err.Error())
			errChan <- err
		} else {
			blog.Infof("rpc_server is stopped.")
		}
	}()

	blog.Infof("rpc_server is started.")
	if err = s.rpcServer.Run(); err != nil {
		err = errors.Wrapf(err, "rpc_server exit with error")
	}
}

// runTaskController run controller
func (s *Server) runTaskController(errChan chan error) {
	var err error
	defer func() {
		s.internalWg.Done()
		if r := recover(); r != nil {
			err = errors.Errorf("task controller panic, err: %v, stack:\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			blog.Errorf("task controller exited: %s", err.Error())
			errChan <- err
		} else {
			blog.Infof("task controller is stopped.")
		}
	}()

	blog.Infof("task controller is started.")
	s.taskController.Run(s.internalCtx)
}

// runDataController run controller
func (s *Server) runDataController(errChan chan error) {
	var err error
	defer func() {
		s.internalWg.Done()
		if r := recover(); r != nil {
			err = errors.Errorf("task controller panic, err: %v, stack:\n%s", r, string(debug.Stack()))
		}
		if err != nil {
			blog.Errorf("data controller exited: %s", err.Error())
			errChan <- err
		} else {
			blog.Infof("data controller is stopped.")
		}
	}()

	blog.Infof("data controller is started.")
	s.dataController.Run(s.internalCtx)
}

// initClient init client
func (s *Server) initClient() error {
	cmOpts := &clustermgr.ClientOptions{
		Name:         s.op.ClusterManager,
		Etcd:         strings.Split(s.op.Registry.Endpoints, ","),
		EtcdConfig:   s.op.Registry.tlsConfig,
		ClientConfig: s.op.clientTLS,
		Token:        s.op.Gateway.Token,
	}
	blog.Infof("gateway:%s, token:%s", s.op.Gateway.Endpoint, s.op.Gateway.Token)
	clusterClient := &clustermgr.ClusterClient{
		Endpoint: s.op.Gateway.Endpoint,
		Token:    s.op.Gateway.Token,
		Sender:   requester2.NewRequester(),
	}
	cmCli := clustermgr.NewClient(cmOpts, s.op.Concurrency, clusterClient)
	s.cmCli = cmCli
	_, cmErr := s.cmCli.ListAllNodeGroups(s.internalCtx)
	if cmErr != nil {
		return fmt.Errorf("init cmCli error:%s", cmErr.Error())
	}
	rmOpts := &resourcemgr.ClientOptions{
		Name:         s.op.ResourceManager,
		Etcd:         strings.Split(s.op.Registry.Endpoints, ","),
		EtcdConfig:   s.op.Registry.tlsConfig,
		ClientConfig: s.op.clientTLS,
	}
	rmCli := resourcemgr.NewClient(rmOpts, s.op.Concurrency)
	s.rmCli = rmCli
	_, rmErr := s.rmCli.ListDevicePool(s.internalCtx, []string{"self"})
	if rmErr != nil {
		return fmt.Errorf("init rmCli error:%s", rmErr.Error())
	}
	if err := s.initCCCli(); err != nil {
		return err
	}
	if err := s.initCrCli(); err != nil {
		return err
	}
	if err := s.initJobCli(); err != nil {
		return err
	}
	if err := s.initBkSopsCli(); err != nil {
		return err
	}
	return nil
}

// initDataService init service
func (s *Server) initDataService() {
	dataService := data.NewDataService(s.cmCli, s.rmCli)
	s.dataService = dataService
}

// initCrCli init cr client
func (s *Server) initCrCli() error {
	requester := requester2.NewRequester()
	cli := cr.New(&apis.ClientOptions{
		Endpoint:    s.op.EndPoints.BKCR,
		UserName:    s.op.APPConfig.UserName,
		AppCode:     s.op.APPConfig.AppCode,
		AppSecret:   s.op.APPConfig.AppSecret,
		AccessToken: s.op.APPConfig.AccessToken,
	}, requester)
	if cli == nil {
		return fmt.Errorf("init cccli failed, nil")
	}
	s.crCli = cli
	return nil
}

// initCCCli init bkcc client
func (s *Server) initCCCli() error {
	requester := requester2.NewRequester()
	cli := bkcc.New(&apis.ClientOptions{
		Endpoint:    s.op.EndPoints.BKCC,
		UserName:    s.op.APPConfig.UserName,
		AppCode:     s.op.APPConfig.AppCode,
		AppSecret:   s.op.APPConfig.AppSecret,
		AccessToken: s.op.APPConfig.AccessToken,
	}, requester)
	if cli == nil {
		return fmt.Errorf("init cccli failed, nil")
	}
	s.ccCli = cli
	return nil
}

// initBkSopsCli init bksops client
func (s *Server) initBkSopsCli() error {
	requester := requester2.NewRequester()
	cli := bksops.New(&apis.ClientOptions{
		Endpoint:    s.op.EndPoints.BKSops,
		UserName:    s.op.APPConfig.UserName,
		AppCode:     s.op.APPConfig.AppCode,
		AppSecret:   s.op.APPConfig.AppSecret,
		AccessToken: s.op.APPConfig.AccessToken,
	}, requester)
	if cli == nil {
		return fmt.Errorf("init cccli failed, nil")
	}
	s.bksopsCli = cli
	return nil
}

// initJobCli init job client
func (s *Server) initJobCli() error {
	requester := requester2.NewRequester()
	cli := job.New(&apis.ClientOptions{
		Endpoint:    s.op.EndPoints.BKJob,
		UserName:    s.op.APPConfig.UserName,
		AppCode:     s.op.APPConfig.AppCode,
		AppSecret:   s.op.APPConfig.AppSecret,
		AccessToken: s.op.APPConfig.AccessToken,
	}, requester)
	if cli == nil {
		return fmt.Errorf("init cccli failed, nil")
	}
	s.jobCli = cli
	return nil
}

// initTaskController init controller
func (s *Server) initTaskController() error {
	taskController := supply.NewTaskController(&scenes.Options{
		Interval:       10,
		Concurrency:    100,
		BKsopsCli:      s.bksopsCli,
		JobCli:         s.jobCli,
		Storage:        s.storage,
		BkccCli:        s.ccCli,
		BkCrCli:        s.crCli,
		ClusterMgrCli:  s.cmCli,
		ResourceMgrCli: s.rmCli,
	})
	s.taskController = taskController
	return taskController.Init()
}

// initDataController init controller
func (s *Server) initDataController() error {
	dataController := data.NewDataController(&scenes.Options{
		Interval:       s.op.DataControllerInterval,
		Concurrency:    100,
		BKsopsCli:      s.bksopsCli,
		JobCli:         s.jobCli,
		Storage:        s.storage,
		BkccCli:        s.ccCli,
		BkCrCli:        s.crCli,
		ClusterMgrCli:  s.cmCli,
		ResourceMgrCli: s.rmCli,
	})
	s.dataController = dataController
	return dataController.Init()
}
