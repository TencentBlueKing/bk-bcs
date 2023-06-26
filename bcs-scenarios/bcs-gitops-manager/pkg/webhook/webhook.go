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

package webhook

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	grpccli "github.com/asim/go-micro/plugins/client/grpc/v4"
	"github.com/asim/go-micro/plugins/registry/etcd/v4"
	grpcsvr "github.com/asim/go-micro/plugins/server/grpc/v4"
	"go-micro.dev/v4"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc"
	grpccred "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/google/uuid"

	"github.com/pkg/errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/webhook/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/webhook/transfer"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/proto"
)

const (
	ServiceName = "gitopswebhook.bkbcs.tencent.com"

	RetryTimes        = 30
	RetryDuration     = 10
	RpcDialTimeout    = 20
	RpcRequestTimeout = 20

	RegisterTTL      = 20
	RegisterInterval = 10
)

// Server defines the server of webhook. It will handle the webhook event from git
type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	op     *options.GitopsWebhookOptions

	serverTlsConfig *tls.Config
	// tls config for cluster manager as microclient role
	clientTLSConfig *tls.Config
	// etcdTLSConfig etcd tls config
	etcdTLSConfig *tls.Config

	httpServer   *http.Server
	rpcServer    micro.Service
	metricServer *http.Server

	tgitHandler transfer.Interface
}

// NewServer create the instance of Server
func NewServer(ctx context.Context, op *options.GitopsWebhookOptions) *Server {
	ctxWrapper, cancel := context.WithCancel(ctx)
	return &Server{
		ctx:         ctxWrapper,
		cancel:      cancel,
		op:          op,
		tgitHandler: transfer.NewTGitHandler(),
	}
}

// Init the webhook server
func (s *Server) Init() error {
	s.initMetricServer()
	initializer := []func() error{
		s.initTLSConfig, s.initRpcServer, s.initHTTPServer,
	}
	for _, init := range initializer {
		if err := init(); err != nil {
			return errors.Wrapf(err, "init failed")
		}
	}
	blog.Infof("all init completed")
	return nil
}

// Run all the goroutines
func (s *Server) Run() error {
	errChan := make(chan error, 3)
	go s.runRpcServer(errChan)
	go s.runHTTPServer(errChan)

	defer func() {
		s.stop()
	}()
	select {
	case err := <-errChan:
		return err
	case <-s.ctx.Done():
		return nil
	}
}

// Stop will stop all the servers
func (s *Server) stop() {
	if err := s.httpServer.Shutdown(s.ctx); err != nil {
		blog.Errorf("shutdown http_server failed, err: %s", err.Error())
	}
	s.cancel()
}

func (s *Server) runRpcServer(errChan chan error) {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("rpc_server panic, stacktrace:\n%s", debug.Stack())
			errChan <- fmt.Errorf("rpc_server panic")
		}
	}()

	blog.Infof("rpc_server is started.")
	if err := s.rpcServer.Run(); err != nil {
		blog.Errorf("rpc_server exit with err: %s", err.Error())
		errChan <- err
		return
	}
	blog.Infof("rpc_server is stopped.")
}

func (s *Server) runHTTPServer(errChan chan error) {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("http_server panic, stacktrace:\n%s", debug.Stack())
			errChan <- fmt.Errorf("http_server panic")
		}
	}()

	blog.Infof("http_server is started.")
	var err error
	if s.serverTlsConfig != nil {
		s.httpServer.TLSConfig = s.serverTlsConfig
		err = s.httpServer.ListenAndServeTLS("", "")
	} else {
		err = s.httpServer.ListenAndServe()
	}
	if err != nil {
		blog.Errorf("http_server stopped with err: %s", err.Error())
		errChan <- err
		return
	}
	blog.Infof("http_server stopped.")
}

func (s *Server) initMetricServer() {
	metricMux := http.NewServeMux()
	s.initPProf(metricMux)
	s.initMetric(metricMux)
	extraServerEndpoint := net.JoinHostPort(s.op.Address, strconv.Itoa(int(s.op.MetricPort)))

	s.metricServer = &http.Server{
		Addr:    extraServerEndpoint,
		Handler: metricMux,
	}
	go func() {
		blog.Infof("start extra modules [pprof, metric] server %s", extraServerEndpoint)
		if err := s.metricServer.ListenAndServe(); err != nil {
			blog.Errorf("metric server listen failed, err %s", err.Error())
		}
	}()
}

func (s *Server) initPProf(mux *http.ServeMux) {
	//blog.Infof("pprof is enabled")
	//mux.HandleFunc("/debug/pprof/", pprof.Index)
	//mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	//mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	//mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	//mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (s *Server) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}

func (s *Server) initRpcServer() error {
	blog.Infof("init rpc server")

	registryEps := strings.Split(s.op.Registry.Endpoints, ",")
	etcdRegistry := etcd.NewRegistry(
		registry.Addrs(registryEps...),
		registry.TLSConfig(s.etcdTLSConfig),
	)
	s.rpcServer = micro.NewService(
		micro.Server(grpcsvr.NewServer(grpcsvr.AuthTLS(s.serverTlsConfig))),
		micro.Client(grpccli.NewClient(grpccli.AuthTLS(s.clientTLSConfig))),
		micro.Name(ServiceName),
		micro.Metadata(map[string]string{
			"httpport": strconv.Itoa(int(s.op.HTTPPort)),
		}),
		micro.Registry(etcdRegistry),
		micro.Context(s.ctx),
		micro.RegisterTTL(RegisterTTL*time.Second),
		micro.RegisterInterval(RegisterInterval*time.Second),
		micro.Address(net.JoinHostPort(s.op.Address, strconv.Itoa(int(s.op.GRPCPort)))),
		micro.WrapHandler(func(handlerFunc server.HandlerFunc) server.HandlerFunc {
			return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
				md, ok := metadata.FromContext(ctx)
				if !ok {
					blog.Errorf("failed get metadata from micro context")
					return nil
				}
				var requestID string
				requestID, ok = md.Get(traceconst.RequestIDHeaderKey)
				if !ok {
					requestID = uuid.New().String()
				}
				ctx = context.WithValue(ctx, traceconst.RequestIDHeaderKey, requestID)
				//user := apis.GetAuthUserFromCtx(ctx)
				//blog.Infof("RequestID[%s] received user '%s' action '%s'", requestID, user.Username, req.Method())
				return handlerFunc(ctx, req, rsp)
			}
		}),
	)

	if err := pb.RegisterBcsGitopsWebhookHandler(s.rpcServer.Server(), s); err != nil {
		return errors.Wrapf(err, "register strategy manager handler failed")
	}
	blog.Infof("init rpc server done")
	return nil
}

func (s *Server) initHTTPServer() error {
	blog.Infof("init http server")
	grpcDialOpts := make([]grpc.DialOption, 0)
	gMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	if s.serverTlsConfig != nil && s.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpccred.NewTLS(s.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err := pb.RegisterBcsGitopsWebhookGwFromEndpoint(s.ctx, gMux, net.JoinHostPort(s.op.Address,
		strconv.Itoa(int(s.op.GRPCPort))), grpcDialOpts); err != nil {
		return errors.Wrapf(err, "register http gateway failed")
	}

	mux := http.NewServeMux()
	mux.Handle("/", gMux)
	s.httpServer = &http.Server{
		TLSConfig: s.serverTlsConfig,
		Handler:   mux,
		Addr:      net.JoinHostPort(s.op.Address, strconv.Itoa(int(s.op.HTTPPort))),
	}
	blog.Infof("init http server done")
	return nil
}

func (s *Server) initTLSConfig() error {
	blog.Infof("init tls config")

	if len(s.op.CAFile) != 0 && len(s.op.ServerCertFile) != 0 && len(s.op.ServerKeyFile) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(s.op.CAFile, s.op.ServerCertFile,
			s.op.ServerKeyFile, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load cluster manager server tls config failed, err %s", err.Error())
			return err
		}
		s.serverTlsConfig = tlsConfig
		blog.Infof("load cluster manager server tls config successfully")
	}

	if len(s.op.CAFile) != 0 && len(s.op.ClientCertFile) != 0 && len(s.op.ClientKeyFile) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(s.op.CAFile, s.op.ClientCertFile,
			s.op.ClientKeyFile, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load cluster manager client tls config failed, err %s", err.Error())
			return err
		}
		s.clientTLSConfig = tlsConfig
		blog.Infof("load cluster manager client tls config successfully")
	}

	if s.op.Registry.CA != "" && s.op.Registry.Endpoints != "" && s.op.Registry.Key != "" && s.op.Registry.Cert != "" {
		etcdTLSConfig, err := ssl.ClientTslConfVerity(s.op.Registry.CA, s.op.Registry.Cert, s.op.Registry.Key, "")
		if err != nil {
			return errors.Wrapf(err, "load registry tls config failed")
		}
		s.etcdTLSConfig = etcdTLSConfig
		blog.Infof("load etcd client tls config successfully")
	}

	blog.Infof("init tls config done")
	return nil
}
