/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"log"
	"net/http"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"bk-bscp/cmd/middle-services/bscp-apiserver/modules/metrics"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbconfigserver "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	pbtemplateserver "bk-bscp/internal/protocol/templateserver"
	pbtunnelserver "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// APIServer is bscp apiserver.
type APIServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// templateserver gRPC gateway server mux.
	tplGWMux *runtime.ServeMux

	// configserver gRPC gateway server mux.
	cfgGWMux *runtime.ServeMux

	// bk-repo http proxy.
	bkRepoProxy *BKRepoReverseProxy

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// config server gRPC connection/client.
	configSvrConn *grpclb.GRPCConn
	configSvrCli  pbconfigserver.ConfigClient

	// templateserver gRPC connection/client.
	templateSvrConn *grpclb.GRPCConn
	templateSvrCli  pbtemplateserver.TemplateClient

	// NOTE: authserver gRPC connection/client, only for healthz.
	authSvrConn *grpclb.GRPCConn
	authSvrCli  pbauthserver.AuthClient

	// NOTE: gse controller gRPC connection/client, only for healthz.
	gseControllerConn *grpclb.GRPCConn
	gseControllerCli  pbgsecontroller.GSEControllerClient

	// NOTE: tunnelserver gRPC connection/client, only for healthz.
	tunnelSvrConn *grpclb.GRPCConn
	tunnelSvrCli  pbtunnelserver.TunnelClient

	// NOTE: datamanager server gRPC connection/client, only for healthz.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// prometheus metrics collector.
	collector *metrics.Collector
}

// NewAPIServer creates new apiserver instance.
func NewAPIServer() *APIServer {
	return &APIServer{}
}

// Init initialize the settings.
func (s *APIServer) Init(setting framework.Setting) {
	s.setting = setting
}

// initialize config and check base content.
func (s *APIServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(s.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	s.viper = viper
}

// initialize logger.
func (s *APIServer) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          s.viper.GetString("logger.directory"),
		LogMaxSize:      s.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       s.viper.GetInt("logger.maxnum"),
		ToStdErr:        s.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    s.viper.GetBool("logger.alsoStderr"),
		Verbosity:       s.viper.GetInt32("logger.level"),
		StdErrThreshold: s.viper.GetString("logger.stderrThreshold"),
		VModule:         s.viper.GetString("logger.vmodule"),
		TraceLocation:   s.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		s.viper.GetString("logger.directory"), s.viper.GetInt32("logger.level"))

	logger.Info("dump configs: endpoint[%+v %+v], insecureEndpoint[%+v %+v] server[%+v] bkrepo[%+v] metrics[%+v] "+
		"etcdCluster[%+v]",
		s.viper.Get("server.endpoint.ip"), s.viper.Get("server.endpoint.port"),
		s.viper.Get("server.insecureEndpoint.ip"), s.viper.Get("server.insecureEndpoint.port"),
		s.viper.Get("server"), s.viper.Get("bkrepo"), s.viper.Get("metrics"), s.viper.Get("etcdCluster"))
}

// create gateway mux for backend servers.
func (s *APIServer) initGWMuxs() {
	opt := runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{EnumsAsInts: true, EmitDefaults: true, OrigName: true})

	headerOpt := runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
		return metadata.Pairs(
			common.RidHeaderKey, req.Header.Get(common.RidHeaderKey),
			common.UserHeaderKey, req.Header.Get(common.UserHeaderKey),
			common.AppCodeHeaderKey, req.Header.Get(common.AppCodeHeaderKey),
			common.AuthorizationHeaderKey, req.Header.Get(common.AuthorizationHeaderKey))
	})

	s.tplGWMux = runtime.NewServeMux(opt, headerOpt)
	s.cfgGWMux = runtime.NewServeMux(opt, headerOpt)
}

// initialize bkrepo http proxy.
func (s *APIServer) initBKRepoProxy() {
	bkrepoDirector := NewBKRepoDirector(s.viper.GetString("bkrepo.host"), s.viper.GetString("bkrepo.token"))
	s.bkRepoProxy = NewBKRepoReverseProxy(s.viper, bkrepoDirector, s.collector)
}

// initialize middle service discovery.
func (s *APIServer) initServiceDiscovery() {
	caFile := s.viper.GetString("etcdCluster.tls.caFile")
	certFile := s.viper.GetString("etcdCluster.tls.certFile")
	keyFile := s.viper.GetString("etcdCluster.tls.keyFile")
	certPassword := s.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTLSConfVerify(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		s.etcdCfg = clientv3.Config{
			Endpoints:   s.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: s.viper.GetDuration("etcdCluster.dialTimeout"),
			TLS:         tlsConf,
		}
	} else {
		s.etcdCfg = clientv3.Config{
			Endpoints:   s.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: s.viper.GetDuration("etcdCluster.dialTimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create config server gRPC client.
func (s *APIServer) initConfigServerClient() {
	ctx := &grpclb.Context{
		Target:     s.viper.GetString("configserver.serviceName"),
		EtcdConfig: s.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(s.viper.GetDuration("configserver.callTimeout")),
	}

	// build gRPC client of configserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create configserver gRPC client, %+v", err)
	}
	s.configSvrConn = conn
	s.configSvrCli = pbconfigserver.NewConfigClient(conn.Conn())
	logger.Info("create configserver gRPC client success.")
}

// create templateserver gRPC client.
func (s *APIServer) initTemplateServerClient() {
	ctx := &grpclb.Context{
		Target:     s.viper.GetString("templateserver.serviceName"),
		EtcdConfig: s.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(s.viper.GetDuration("templateserver.callTimeout")),
	}

	// build gRPC client of templateserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create templateserver gRPC client, %+v", err)
	}
	s.templateSvrConn = conn
	s.templateSvrCli = pbtemplateserver.NewTemplateClient(conn.Conn())
	logger.Info("create templateserver gRPC client success.")
}

// create auth server gRPC client.
func (s *APIServer) initAuthServerClient() {
	ctx := &grpclb.Context{
		Target:     s.viper.GetString("authserver.serviceName"),
		EtcdConfig: s.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(s.viper.GetDuration("authserver.callTimeout")),
	}

	// build gRPC client of authserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create authserver gRPC client, %+v", err)
	}
	s.authSvrConn = conn
	s.authSvrCli = pbauthserver.NewAuthClient(conn.Conn())
	logger.Info("create authserver gRPC client success.")
}

// create gse controller gRPC client.
func (s *APIServer) initGSEControllerClient() {
	ctx := &grpclb.Context{
		Target:     s.viper.GetString("gsecontroller.serviceName"),
		EtcdConfig: s.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(s.viper.GetDuration("gsecontroller.callTimeout")),
	}

	// build gRPC client of gse controller.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create gse controller gRPC client, %+v", err)
	}
	s.gseControllerConn = conn
	s.gseControllerCli = pbgsecontroller.NewGSEControllerClient(conn.Conn())
	logger.Info("create gse controller gRPC client success.")
}

// create tunnelserver gRPC client.
func (s *APIServer) initTunnelServerClient() {
	ctx := &grpclb.Context{
		Target:     s.viper.GetString("tunnelserver.serviceName"),
		EtcdConfig: s.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(s.viper.GetDuration("tunnelserver.callTimeout")),
	}

	// build gRPC client of tunnelserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create tunnelserver gRPC client, %+v", err)
	}
	s.tunnelSvrConn = conn
	s.tunnelSvrCli = pbtunnelserver.NewTunnelClient(conn.Conn())
	logger.Info("create tunnelserver gRPC client success.")
}

// create datamanager gRPC client.
func (s *APIServer) initDataManagerClient() {
	ctx := &grpclb.Context{
		Target:     s.viper.GetString("datamanager.serviceName"),
		EtcdConfig: s.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(s.viper.GetDuration("datamanager.callTimeout")),
	}

	// build gRPC client of datamanager.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create datamanager gRPC client, %+v", err)
	}
	s.dataMgrConn = conn
	s.dataMgrCli = pbdatamanager.NewDataManagerClient(conn.Conn())
	logger.Info("create datamanager gRPC client success.")
}

// initializes prometheus metrics collector.
func (s *APIServer) initMetricsCollector() {
	s.collector = metrics.NewCollector(s.viper.GetString("metrics.endpoint"), s.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := s.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initMods initializes the server modules.
func (s *APIServer) initMods() {
	// initialize service discovery.
	s.initServiceDiscovery()

	// initialize config server gRPC client.
	s.initConfigServerClient()

	// initialize templateserver gRPC client.
	s.initTemplateServerClient()

	// initialize auth server gRPC client.
	s.initAuthServerClient()

	// initialize gse controller gRPC client.
	s.initGSEControllerClient()

	// initialize tunnelserver gRPC client.
	s.initTunnelServerClient()

	// initialize datamanager gRPC client.
	s.initDataManagerClient()

	// initialize metrics collector.
	s.initMetricsCollector()
}

// initialize service.
func (s *APIServer) initService() {
	// init gateway muxs.
	s.initGWMuxs()

	// init bkrepo proxy.
	s.initBKRepoProxy()

	// http handler.
	httpMux := http.NewServeMux()

	// new router handler.
	rtr := mux.NewRouter()

	// setup routers.
	s.setupRouters(rtr)
	httpMux.Handle("/", rtr)

	// setup filters, all requests would cross in the filter.
	apiServerMux := s.setupFilters(httpMux)

	// register grpc clients.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pbconfigserver.RegisterConfigHandlerClient(ctx, s.cfgGWMux, s.configSvrCli); err != nil {
		logger.Fatal("register configserver handler, %+v", err)
	}
	if err := pbtemplateserver.RegisterTemplateHandlerClient(ctx, s.tplGWMux, s.templateSvrCli); err != nil {
		logger.Fatal("register template server handler, %+v", err)
	}

	// service listen and serve.
	if !(s.viper.IsSet("server.endpoint.ip") && s.viper.IsSet("server.endpoint.port")) &&
		!(s.viper.IsSet("server.insecureEndpoint.ip") && s.viper.IsSet("server.insecureEndpoint.port")) {
		logger.Fatal("no available http server here, check local config")
	}

	// secure http server.
	if s.viper.IsSet("server.endpoint.ip") && s.viper.IsSet("server.endpoint.port") {
		gwEndpoint := common.Endpoint(s.viper.GetString("server.endpoint.ip"), s.viper.GetInt("server.endpoint.port"))
		httpServer := &http.Server{Addr: gwEndpoint, Handler: apiServerMux}

		// http server with TLS must verify client in gateway.
		tlsConf, err := ssl.ServerTLSConfVerify(
			s.viper.GetString("server.tls.caFile"), s.viper.GetString("server.tls.certFile"),
			s.viper.GetString("server.tls.keyFile"), s.viper.GetString("server.tls.certPassword"))
		if err != nil {
			logger.Fatal("gateway setup https server failed, %+v", err)
		}
		httpServer.TLSConfig = tlsConf

		// listen and serve with TLS.
		go func() {
			logger.Info("APIServer with TLS running now.")
			if err := httpServer.ListenAndServeTLS("", ""); err != nil {
				logger.Fatal("https server listen and serve, %+v", err)
			}
		}()
	}

	// insecure http server.
	if s.viper.IsSet("server.insecureEndpoint.ip") && s.viper.IsSet("server.insecureEndpoint.port") {
		gwInsecureEndpoint := common.Endpoint(s.viper.GetString("server.insecureEndpoint.ip"),
			s.viper.GetInt("server.insecureEndpoint.port"))
		httpInsecureServer := &http.Server{Addr: gwInsecureEndpoint, Handler: apiServerMux}

		// listen and serve without TLS.
		go func() {
			logger.Info("APIServer without TLS running now.")
			if err := httpInsecureServer.ListenAndServe(); err != nil {
				logger.Fatal("http server listen and serve, %+v", err)
			}
		}()
	}
}

// Run runs config server.
func (s *APIServer) Run() {
	// initialize config.
	s.initConfig()

	// initialize logger.
	s.initLogger()
	defer s.Stop()

	// initialize server modules.
	s.initMods()

	// init api http server.
	s.initService()

	// hanging here.
	select {}
}

// Stop stops the apiserver.
func (s *APIServer) Stop() {
	// close datamanager gRPC connection when server exit.
	if s.dataMgrConn != nil {
		s.dataMgrConn.Close()
	}

	// close tunnelserver gRPC connection when server exit.
	if s.tunnelSvrConn != nil {
		s.tunnelSvrConn.Close()
	}

	// close gse controller gRPC connection when server exit.
	if s.gseControllerConn != nil {
		s.gseControllerConn.Close()
	}

	// close authserver gRPC connection when server exit.
	if s.authSvrConn != nil {
		s.authSvrConn.Close()
	}

	// close templateserver gRPC connection when server exit.
	if s.templateSvrConn != nil {
		s.templateSvrConn.Close()
	}

	// close configserver gRPC connection when server exit.
	if s.configSvrConn != nil {
		s.configSvrConn.Close()
	}

	// close logger.
	logger.CloseLogs()
}
