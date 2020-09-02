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
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/service"
	microgrpc "github.com/micro/go-micro/v2/service/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/bcslogmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/util"
)

const (
	// LogManagerServiceName service name
	LogManagerServiceName = "bcs-log-manager.bkbcs.tencent.com"
)

// Server bcslogmanager apiserver
type Server struct {
	conf         *config.APIServerConfig
	grpcEndpoint string
	httpEndpoint string
	ctx          context.Context
	lis          net.Listener
	grpcServer   *grpc.Server
	mux          *http.ServeMux
	gwmux        *runtime.ServeMux
	apiImpl      *LogManagerServerImpl
	microSvr     service.Service
}

// NewAPIServer creates Server instance
func NewAPIServer(ctx context.Context, conf *config.APIServerConfig, logManager *k8s.LogManager) *Server {
	return &Server{
		conf: conf,
		ctx:  ctx,
		apiImpl: &LogManagerServerImpl{
			logManager: logManager,
			apiHost:    conf.BKDataAPIHost,
		},
	}
}

// init tcp listener
func (s *Server) initListener() error {
	var err error
	s.grpcEndpoint = fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port)
	s.httpEndpoint = fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port-1)
	s.lis, err = net.Listen("tcp", s.httpEndpoint)
	if err != nil {
		blog.Infof("listen tcp failed: %s", err.Error())
		return err
	}
	return nil
}

// init http gateway(with TLS) of grpc server(with TLS)
func (s *Server) startGateway() error {
	var opts []grpc.DialOption
	var err error
	if s.conf.APICerts.ServerCertFile != "" {
		crds, err := credentials.NewClientTLSFromFile(s.conf.APICerts.ServerCertFile, "")
		if err != nil {
			blog.Errorf("Build credential from server cert(%s)/key(%s) file failed: %s", s.conf.APICerts.ServerCertFile, s.conf.APICerts.ServerKeyFile, err.Error())
			return err
		}
		opts = append(opts, grpc.WithTransportCredentials(crds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	// init http gateway
	s.gwmux = runtime.NewServeMux()
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
	if s.conf.APICerts.ServerCertFile != "" && s.conf.APICerts.ServerKeyFile != "" {
		// TODO
		tlsConfig, err := ssl.ServerTslConf(s.conf.APICerts.CAFile, s.conf.APICerts.ServerCertFile, s.conf.APICerts.ServerKeyFile, "")
		if err != nil {
			blog.Errorf("ServerTslConf of api gateway failed: %s", err.Error())
			return err
		}
		server = http.Server{
			Addr:      s.httpEndpoint,
			Handler:   mux,
			TLSConfig: tlsConfig,
		}
		workFunc = func() {
			blog.Info("starting logmanager gateway api server...")
			if err := server.ServeTLS(s.lis, s.conf.APICerts.ServerCertFile, s.conf.APICerts.ServerKeyFile); err != nil {
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
			if err := server.Serve(s.lis); err != nil {
				blog.Errorf("start grpc server with net listener failed, err %s", err.Error())
				util.SendTermSignal()
			}
		}
	}
	go workFunc()
	return nil
}

func (s *Server) startMicroService() error {
	var opts []service.Option
	var etcdTlsConf *tls.Config
	var err error
	if s.conf.EtcdCerts.CAFile != "" && s.conf.EtcdCerts.ClientCertFile != "" && s.conf.EtcdCerts.ClientKeyFile != "" {
		etcdTlsConf, err = ssl.ClientTslConfVerity(s.conf.EtcdCerts.CAFile, s.conf.EtcdCerts.ClientCertFile, s.conf.EtcdCerts.ClientKeyFile, "")
		if err != nil {
			blog.Errorf("Build etcd tlsconf failed: %s", err.Error())
			return err
		}
	} else {
		etcdTlsConf = nil
	}

	// etcd registry options
	regOption := func(e *registry.Options) {
		blog.Errorf("%+v", s.conf.EtcdHosts)
		e.Addrs = s.conf.EtcdHosts
		if etcdTlsConf != nil {
			e.TLSConfig = etcdTlsConf
			e.Secure = true
		} else {
			e.Secure = false
		}
	}
	opts = append(opts, service.Name(LogManagerServiceName))
	opts = append(opts, service.Version("1.19.x"))
	opts = append(opts, service.Context(s.ctx))
	opts = append(opts, service.Address(s.grpcEndpoint))
	opts = append(opts, service.RegisterInterval(time.Second*30))
	opts = append(opts, service.RegisterTTL(time.Second*30))
	opts = append(opts, service.Registry(etcd.NewRegistry(regOption)))
	// whether to use transport layer secuerity
	if s.conf.APICerts.ServerCertFile != "" && s.conf.APICerts.ServerKeyFile != "" {
		apiTLSConf, err := ssl.ServerTslConfVerity(s.conf.APICerts.ServerCertFile, s.conf.APICerts.ServerKeyFile, "")
		if err != nil {
			blog.Errorf("Build logmanager tlsconf failed: %s", err.Error())
			return err
		}
		opts = append(opts, microgrpc.WithTLS(apiTLSConf))
	}

	s.microSvr = microgrpc.NewService(opts...)
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
	err := s.initListener()
	if err != nil {
		return err
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
