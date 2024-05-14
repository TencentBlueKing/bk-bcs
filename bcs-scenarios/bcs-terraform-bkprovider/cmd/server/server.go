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
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/middleware"
	grpccli "github.com/asim/go-micro/plugins/client/grpc/v4"
	"github.com/asim/go-micro/plugins/registry/etcd/v4"
	grpcsvr "github.com/asim/go-micro/plugins/server/grpc/v4"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/handler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/pkg/middleware/auth"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/proto"
)

const (
	defaultServiceName = "terraform-bkprovider.bkbcs.tencent.com"
	gracefulPeriod     = 3
)

// Server devspace manager server
type Server struct {
	// server options
	opt *Options
	// context control for exit
	svrContext context.Context
	svrCancel  context.CancelFunc
	// gatewayServer for grpc gateway
	gatewayServer http.Server
	// go-micro v4 grpc server
	microService micro.Service

	JwtAuth *auth.JWTAuth
}

// NewServer create server
func NewServer(opt *Options) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		opt:        opt,
		svrContext: ctx,
		svrCancel:  cancel,
	}
}

// Init init data manager server
func (s *Server) Init() error {
	initializer := []func() error{
		s.initAuth, s.initMicroService, s.initGatewayServer,
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
		s.startMicroService, s.startGatewayServer, s.startSignalHandler,
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

// initMicroService init grpc service
func (s *Server) initMicroService() error {
	blog.Infof("init go-micro service")

	// context for microservice
	ctx, _ := context.WithCancel(s.svrContext) // nolint
	// registry init
	hosts := strings.Split(s.opt.Registry.Endpoints, ",")
	globalRegistry := etcd.NewRegistry(
		registry.Addrs(hosts...),
		registry.TLSConfig(s.opt.Registry.tlsConfig),
	)
	authWrapper := middleware.NewGoMicroAuth(s.JwtAuth.GetJWTClient()).
		EnableSkipHandler(s.JwtAuth.SkipHandler)
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
		}),
		micro.Address(fmt.Sprintf("%s:%d", s.opt.Address, s.opt.Port)),
		micro.Version(version.BcsVersion),
		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(25*time.Second),
		micro.Registry(globalRegistry),
		micro.WrapHandler(authWrapper.AuthenticationFunc(), s.JwtAuth.AuthorizationFunc),
	)
	// nothing to init
	// s.microService.Init()
	// Register handler
	if err := pb.RegisterBcsTerraformBkProviderHandler(s.microService.Server(),
		handler.NewBcsApiHandler(s.opt.BkSystem.BkAppCode, s.opt.BkSystem.BkAppSecret, s.opt.BkSystem.BkEnv)); err != nil {
		return fmt.Errorf("micro service registry handle failed, %s", err.Error())
	}
	blog.Infof("init go-micro service success")
	return nil
}

// initGatewayServer init http gateway for grpc service proxy
func (s *Server) initGatewayServer() error {
	blog.Info("init gateway server")
	ctx, _ := context.WithCancel(s.svrContext) // nolint
	// register grpc server information
	gMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true,
		}),
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			if strings.HasPrefix(s, "X-") {
				return s, true
			}
			return runtime.DefaultHeaderMatcher(s)
		}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(s.opt.clientTLS))}
	err := pb.RegisterBcsTerraformBkProviderGwFromEndpoint(
		ctx, gMux, fmt.Sprintf("%s:%d", s.opt.Address, s.opt.Port), opts)
	if err != nil {
		return fmt.Errorf("http gateway register grpc service failed, %s", err.Error())
	}
	mux := http.NewServeMux()
	s.initSwagger(mux)
	mux.Handle("/", gMux)
	s.gatewayServer = http.Server{
		Addr:      fmt.Sprintf("%s:%d", s.opt.Address, s.opt.HTTPPort),
		Handler:   mux,
		TLSConfig: s.opt.serverTLS,
	}
	blog.Infof("init gateway server success")
	return nil
}

// initSwagger init swagger
func (s *Server) initSwagger(mux *http.ServeMux) {
	if len(s.opt.Swagger.Dir) != 0 {
		blog.Infof("swagger doc is enabled")
		mux.HandleFunc("/terraform-bkprovider/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(s.opt.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/terraform-bkprovider/swagger/")))
		})
	}
}

func (s *Server) initAuth() error {
	blog.Infof("init auth")
	var err error
	// init jwt auth
	s.JwtAuth, err = auth.NewJWTAuth(s.opt.Auth.PublicKeyFile, s.opt.Auth.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("init jwt auth failed, %s", err.Error())
	}
	blog.Infof("init auth success")
	return nil
}
