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
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/go-micro/plugins/v4/registry/etcd"
	microgrpcsvc "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	microsvc "go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/handler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/apis/git"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/storage"
	precheck "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/proto"
)

const (
	// RpcServerName the name of PredictApproval server
	RpcServerName = "gitopsprecheck.bkbcs.tencent.com"
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
	publicFunc     common.PublicFunc
	db             storage.Interface
}

// NewServer new server
func NewServer(ctx context.Context, cancel context.CancelFunc, opt *Options) *Server {
	return &Server{
		op:             opt,
		internalCtx:    ctx,
		internalCancel: cancel,
	}
}

// Init will parse options, and init grpc_server/http_server/metric_server
func (s *Server) Init() error {
	s.initMetricServer()
	if err := s.initDB(); err != nil {
		return errors.Wrapf(err, "init db failed")
	}
	if err := s.initPublicFunc(); err != nil {
		return errors.Wrapf(err, "init publicFunc failed")
	}
	if err := s.initHTTPServer(); err != nil {
		return errors.Wrapf(err, "init http server failed")
	}
	if err := s.initRpcServer(); err != nil {
		return errors.Wrapf(err, "init rpc server failed")
	}
	return nil
}

// Run all the server that had init
func (s *Server) Run() error {
	errChan := s.run(s.runHttpServer, s.runMetricServer, s.runRpcServer)

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
		// microsvc.Server(microgrpcsvc.NewServer(microgrpcsvc.AuthTLS(s.op.serverTLS))),
		microsvc.Server(microgrpcsvc.NewServer(microgrpcsvc.MaxMsgSize(s.op.MaxRecvMsgSize))),
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

	if err := precheck.RegisterGitOpsPreCheckHandler(s.rpcServer.Server(),
		handler.New(&handler.Opts{PublicFunc: s.publicFunc})); err != nil {
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
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := precheck.RegisterGitOpsPreCheckGwFromEndpoint(
		ctx, mux, fmt.Sprintf("%s:%d", s.op.Address, s.op.Port), opts)
	if err != nil {
		return fmt.Errorf("http gateway register grpc service failed, %s", err.Error())
	}
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.op.Address, s.op.HTTPPort),
		Handler: mux,
		// TLSConfig: s.op.serverTLS,
	}
	blog.Infof("init http server success")
	return nil
}

// initStorage init db
// func (s *Server) initStorage() error {
//	// convert to mongo options
//	storageOption := &mongo.Options{
//		Hosts:                 strings.Split(s.op.Storage.Endpoints, ","),
//		ConnectTimeoutSeconds: 3,
//		Database:              s.op.Storage.Database,
//		Username:              s.op.Storage.UserName,
//		Password:              s.op.Storage.Password,
//	}
//	instance, err := mongo.NewDB(storageOption)
//	if err != nil {
//		return fmt.Errorf("storage create dm mongo instance failed, %s", err.Error())
//	}
//	if pingErr := instance.Ping(); pingErr != nil {
//		return fmt.Errorf("dm storage connection test failed, %s", pingErr.Error())
//	}
//	s.storage = mongostore.NewServer(instance)
//	blog.Infof("init storage success")
//	return nil
//}

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
	if err = s.httpServer.ListenAndServe(); err != nil {
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

func (s *Server) initDB() error {
	db, err := storage.NewDriver(s.op.DBConfig)
	if err != nil {
		return errors.Wrapf(err, "create db failed")
	}
	if err = db.Init(); err != nil {
		return errors.Wrapf(err, "init db failed")
	}
	s.db = db
	blog.Infof("init db success.")
	return nil
}

func (s *Server) initPublicFunc() error {
	argoDB, _, err := store.NewArgoDB(s.internalCtx, s.op.AdminNamespace)
	if err != nil {
		return fmt.Errorf("new argoDB failed:%s", err.Error())
	}
	gitFactory := git.NewFactory(&git.FactoryOpts{
		TGitEndpoint:          s.op.EndPoints.TGit,
		TGitSubStr:            s.op.EndPoints.TGitSubStr,
		TGitOutEndpoint:       s.op.EndPoints.TGitOut,
		TGitOutSubStr:         s.op.EndPoints.TGitOutSubStr,
		TGitCommunityEndpoint: s.op.EndPoints.TGitCommunity,
		TGitCommunitySubStr:   s.op.EndPoints.TGitCommunitySubStr,
		GitlabEndpoint:        s.op.EndPoints.Gitlab,
		GitlabSubStr:          s.op.EndPoints.GitlabSubStr,
	})
	gitFactory.Init()
	opts := &common.PublicFuncOpts{PowerAppEp: s.op.EndPoints.PowerApp}
	s.publicFunc = common.NewPublicFunc(argoDB, gitFactory, s.db, opts)
	return nil
}
