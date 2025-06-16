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

// Package app contains the app for the mesh manager
package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	discovery "github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	trace "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/micro"
	grpccli "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	grpcsvr "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// Server mesh manager server
type Server struct {
	microService  micro.Service
	microRegistry registry.Registry

	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config

	httpServer *http.Server
	opt        *options.MeshManagerOptions

	discovery            *discovery.ModuleDiscovery
	helmManagerDiscovery *discovery.ModuleDiscovery

	helmManagerClient *helmmanager.HelmClientWrapper

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopChan      chan struct{}
}

// NewServer create mesh manager instance
func NewServer(opt *options.MeshManagerOptions) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		opt:           opt,
		ctx:           ctx,
		ctxCancelFunc: cancel,
		stopChan:      make(chan struct{}),
	}
}

// Init init modules of server
func (s *Server) Init() error {
	// initializers by sequence
	initializer := []func() error{
		s.initTLSConfig,
		s.initRegistry,
		// s.initModel,
		// s.initIAMClient,
		// s.initJWTClient,
		// s.initSkipClients,
		s.initDiscovery,
		// s.initSkipHandler,
		s.initMicro,
		s.initHTTPService,
		// TODO: add log wrapper
	}

	// init
	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}

	// init helm manager client
	helmManagerClient, _, err := helmmanager.GetClient(common.ServiceDomain)
	if err != nil {
		return err
	}
	s.helmManagerClient = helmManagerClient

	return nil
}

// Run run the server
func (s *Server) Run() error {

	eg, _ := errgroup.WithContext(s.ctx)

	eg.Go(func() error {
		return s.microService.Run()
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

// initTLSConfig init server and client tls config
func (s *Server) initTLSConfig() error {
	if len(s.opt.ServerCert) != 0 && len(s.opt.ServerKey) != 0 && len(s.opt.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(s.opt.ServerCa, s.opt.ServerCert,
			s.opt.ServerKey, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load mesh manager server tls config failed, err %s", err.Error())
			return err
		}
		s.tlsConfig = tlsConfig
		blog.Infof("load mesh manager server tls config successfully")
	}

	if len(s.opt.ClientCert) != 0 && len(s.opt.ClientKey) != 0 && len(s.opt.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(s.opt.ClientCa, s.opt.ClientCert,
			s.opt.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load mesh manager client tls config failed, err %s", err.Error())
			return err
		}
		s.clientTLSConfig = tlsConfig
		blog.Infof("load mesh manager client tls config successfully")
	}

	// init tls config success
	blog.Infof("init tls config successfully")
	return nil
}

// initRegistry init micro service registry
func (s *Server) initRegistry() error {
	if s.opt.Etcd.EtcdEndpoints == "" {
		blog.Warnf("etcd endpoints is empty, use default endpoints")
		return nil
	}
	// parse etcd endpoints
	address := strings.ReplaceAll(s.opt.Etcd.EtcdEndpoints, ";", ",")
	address = strings.ReplaceAll(address, " ", ",")
	etcdEndpoints := strings.Split(address, ",")
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if len(s.opt.Etcd.EtcdCa) != 0 && len(s.opt.Etcd.EtcdCert) != 0 && len(s.opt.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(s.opt.Etcd.EtcdCa, s.opt.Etcd.EtcdCert, s.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}
	s.microRegistry = etcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)
	if err := s.microRegistry.Init(); err != nil {
		return err
	}

	// init registry success
	blog.Infof("init registry successfully")
	return nil
}

// init micro service
func (s *Server) initMicro() error {
	opts := []micro.Option{
		micro.Server(grpcsvr.NewServer(
			grpcsvr.AuthTLS(s.tlsConfig),
		)),
		micro.Client(grpccli.NewClient(
			grpccli.AuthTLS(s.clientTLSConfig),
		)),
		micro.Name(common.ServiceDomain),
		micro.Context(s.ctx),
		micro.Metadata(map[string]string{common.MicroMetaKeyHTTPPort: strconv.Itoa(int(s.opt.HTTPPort))}),
		micro.Address(net.JoinHostPort(s.opt.Address, strconv.Itoa(int(s.opt.Port)))),
		micro.Version(version.BcsVersion),
		micro.RegisterTTL(30 * time.Second),
		micro.RegisterInterval(25 * time.Second),
		micro.Flags(&cli.StringFlag{
			Name:        "f",
			Usage:       "set config file path",
			DefaultText: "./bcs-mesh-manager.json",
		}),
		micro.WrapHandler(
			utils.ResponseWrapper,
			utils.RequestLogWarpper,
			// authWrapper.AuthenticationFunc,
			// authWrapper.AuthorizationFunc,
			trace.NewTracingWrapper(),
		),
	}
	if s.microRegistry != nil {
		opts = append(opts, micro.Registry(s.microRegistry))
	}
	s.microService = micro.NewService(
		opts...,
	)
	s.microService.Init()

	if err := meshmanager.RegisterMeshManagerHandler(
		s.microService.Server(),
		handler.NewMeshManager(&handler.MeshManagerOptions{IstioConfig: s.opt.IstioConfig}),
	); err != nil {
		blog.Errorf("failed to register mesh manager handler to micro, error: %s", err.Error())
		return err
	}

	blog.Infof("success to register mesh manager handler to micro")
	return nil
}

// initHTTPService init http service
func (s *Server) initHTTPService() error {
	// init http gateway
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			// 设置为true，表示输出未设置的值
			MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: true},
		}),
	)
	grpcDialOpts := []grpc.DialOption{}
	if s.tlsConfig != nil && s.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(credentials.NewTLS(s.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}
	err := meshmanager.RegisterMeshManagerGwFromEndpoint(
		context.TODO(),
		gwmux,
		net.JoinHostPort(s.opt.Address, strconv.Itoa(int(s.opt.Port))),
		grpcDialOpts)
	if err != nil {
		blog.Errorf("register http gateway failed, err %s", err.Error())
		return fmt.Errorf("register http gateway failed, err %s", err.Error())
	}
	router := mux.NewRouter()
	router.Handle("/{uri:.*}", handlers.LoggingHandler(os.Stdout, gwmux))
	blog.Info("register grpc gateway handler to path /")

	// init http server
	smux := http.NewServeMux()
	smux.Handle("/", router)

	httpAddress := net.JoinHostPort(s.opt.Address, strconv.Itoa(int(s.opt.HTTPPort)))

	s.httpServer = &http.Server{
		Addr:    httpAddress,
		Handler: smux,
	}

	// start http gateway server
	go func() {
		var err error
		blog.Infof("start http gateway server on address %s", httpAddress)
		if s.tlsConfig != nil {
			s.httpServer.TLSConfig = s.tlsConfig
			err = s.httpServer.ListenAndServeTLS("", "")
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil {
			blog.Errorf("start http gateway server failed, err %s", err.Error())
			s.ctxCancelFunc()
		}
	}()

	return nil
}

func (s *Server) initDiscovery() error {
	if !discovery.UseServiceDiscovery() {
		s.discovery = discovery.NewModuleDiscovery(common.ServiceDomain, s.microRegistry)
		blog.Info("init discovery for project manager successfully")
		// enable discovery cluster manager module
		s.helmManagerDiscovery = discovery.NewModuleDiscovery(common.HelmManagerServiceDomain, s.microRegistry)
		helmmanager.SetClientConfig(s.clientTLSConfig, s.helmManagerDiscovery)
	} else {
		helmmanager.SetClientConfig(s.clientTLSConfig, nil)
	}
	blog.Info("init discovery for cluster manager successfully")
	return nil
}
