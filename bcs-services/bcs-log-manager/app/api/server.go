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
 *
 */

package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/service"
	microgrpc "github.com/micro/go-micro/v2/service/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/util"
)

const (
	// LogManagerServiceName service name
	LogManagerServiceName = "logmanager.bkbcs.tencent.com"
)

// Server bcslogmanager apiserver
type Server struct {
	conf         *config.APIServerConfig
	grpcEndpoint string
	httpEndpoint string
	ctx          context.Context
	grpcServer   *grpc.Server
	mux          *http.ServeMux
	gwmux        *runtime.ServeMux
	apiImpl      *LogManagerServerImpl
	microSvr     service.Service
	etcdTLS      *tls.Config
	serverTLS    *tls.Config
	clientTLS    *tls.Config
}

// NewAPIServer creates Server instance
func NewAPIServer(ctx context.Context, conf *config.APIServerConfig, logManager k8s.LogManagerInterface) *Server {
	return &Server{
		conf: conf,
		ctx:  ctx,
		apiImpl: &LogManagerServerImpl{
			logManager:          logManager,
			apiHost:             conf.BKDataAPIHost,
			bkdataClientCreator: bkdata.NewClientCreator(),
		},
	}
}

// init http gateway(with TLS) of grpc server(with TLS)
func (s *Server) startGateway() error {
	var opts []grpc.DialOption
	var err error
	if s.clientTLS != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(s.clientTLS)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	// init http gateway
	s.gwmux = runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	err = proto.RegisterLogManagerGwFromEndpoint(s.ctx, s.gwmux, s.grpcEndpoint, opts)
	if err != nil {
		blog.Errorf("register logmanager gateway failed, err %s", err.Error())
		return err
	}
	blog.Infof("register logmanager gateway succ")

	// start http server
	mux := http.NewServeMux()
	mux.Handle("/", s.gwmux)

	// http serve function
	var workFunc func()
	var server http.Server
	// whether to use transport layer secuerity
	if s.serverTLS != nil {
		server = http.Server{
			Addr:      s.httpEndpoint,
			Handler:   mux,
			TLSConfig: s.serverTLS,
		}
		workFunc = func() {
			blog.Info("starting logmanager gateway api server...")
			if err := server.ListenAndServeTLS("", ""); err != nil {
				blog.Errorf("start grpc server with net listener failed, err %s", err.Error())
				util.SendTermSignal()
			}
		}
	} else {
		server = http.Server{
			Addr:    s.httpEndpoint,
			Handler: mux,
		}
		workFunc = func() {
			blog.Info("starting logmanager gateway api server...")
			if err := server.ListenAndServe(); err != nil {
				blog.Errorf("start grpc server with net listener failed, err %s", err.Error())
				util.SendTermSignal()
			}
		}
	}
	go workFunc()
	return nil
}

func (s *Server) startMicroService() error {
	var err error

	// etcd registry options
	regOption := func(e *registry.Options) {
		e.Addrs = s.conf.EtcdHosts
		if s.etcdTLS != nil {
			e.TLSConfig = s.etcdTLS
			e.Secure = true
		} else {
			e.Secure = false
		}
	}
	sevOption := func(o *server.Options) {
		if s.serverTLS != nil {
			o.TLSConfig = s.serverTLS
		}
		o.Name = LogManagerServiceName
		o.Version = version.BcsVersion
		o.Context = s.ctx
		o.Address = s.grpcEndpoint
		o.RegisterInterval = time.Second * 30
		o.RegisterTTL = time.Second * 40
		o.Registry = etcd.NewRegistry(regOption)
	}

	s.microSvr = microgrpc.NewService()
	s.microSvr.Server().Init(sevOption)
	s.microSvr.Init()
	err = proto.RegisterLogManagerHandler(s.microSvr.Server(), s.apiImpl)
	if err != nil {
		blog.Errorf("RegisterLogManagerHandler failed: %s", err.Error())
		return err
	}
	blog.Infof("register logmanager grpc micro service succ")
	// start micro service
	go func() {
		blog.Info("starting logmanager grpc micro service...")
		if err := s.microSvr.Run(); err != nil {
			blog.Errorf("run micro grpc service failed: %s", err.Error())
			util.SendTermSignal()
		}
	}()
	return nil
}

// Run runs the api server
func (s *Server) Run() error {
	var err error
	s.grpcEndpoint = fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port)
	s.httpEndpoint = fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port-1)
	s.clientTLS, err = ssl.ClientTslConfVerity(s.conf.APICerts.CAFile, s.conf.APICerts.ClientCertFile, s.conf.APICerts.ClientKeyFile, static.ClientCertPwd)
	if err != nil {
		blog.Warnf("parse client TLS config failed: %s", err.Error())
		s.clientTLS = nil
	}
	s.serverTLS, err = ssl.ServerTslConf(s.conf.APICerts.CAFile, s.conf.APICerts.ServerCertFile, s.conf.APICerts.ServerKeyFile, static.ServerCertPwd)
	if err != nil {
		blog.Warnf("parse server TLS config failed: %s", err.Error())
		s.serverTLS = nil
	}
	s.etcdTLS, err = ssl.ClientTslConfVerity(s.conf.EtcdCerts.CAFile, s.conf.EtcdCerts.ClientCertFile, s.conf.EtcdCerts.ClientKeyFile, "")
	if err != nil {
		blog.Warnf("parse etcd client TLS config failed: %s", err.Error())
		s.etcdTLS = nil
	}
	// s.initGRPCServer()
	err = s.startGateway()
	if err != nil {
		return err
	}
	err = s.startMicroService()
	if err != nil {
		return err
	}
	return nil
}
