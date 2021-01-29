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
	"log"
	"math"
	"net"

	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"bk-bscp/cmd/atomic-services/bscp-templateserver/modules/metrics"
	"bk-bscp/internal/audit"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/framework/executor"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// TemplateServer is bscp template server.
type TemplateServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// templateserver discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// authserver gRPC connection/client.
	authSvrConn *grpclb.GRPCConn
	authSvrCli  pbauthserver.AuthClient

	// datamanager gRPC connection/client.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewTemplateServer creates new template server instance.
func NewTemplateServer() *TemplateServer {
	return &TemplateServer{}
}

// Init initialize the settings.
func (ts *TemplateServer) Init(setting framework.Setting) {
	ts.setting = setting
}

// initialize config and check base content.
func (ts *TemplateServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(ts.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	ts.viper = viper
}

// initialize logger.
func (ts *TemplateServer) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          ts.viper.GetString("logger.directory"),
		LogMaxSize:      ts.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       ts.viper.GetInt("logger.maxnum"),
		ToStdErr:        ts.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    ts.viper.GetBool("logger.alsoStderr"),
		Verbosity:       ts.viper.GetInt32("logger.level"),
		StdErrThreshold: ts.viper.GetString("logger.stderrThreshold"),
		VModule:         ts.viper.GetString("logger.vmodule"),
		TraceLocation:   ts.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		ts.viper.GetString("logger.directory"), ts.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v %+v] metrics[%+v] bkrepo[%+v], datamanager[%+v] etcdCluster[%+v]",
		ts.viper.Get("server.endpoint.ip"), ts.viper.Get("server.endpoint.port"), ts.viper.Get("metrics"),
		ts.viper.Get("bkrepo"), ts.viper.Get("datamanager"), ts.viper.Get("etcdCluster"))
}

// create new service struct of templateserver, and register service later.
func (ts *TemplateServer) initServiceDiscovery() {
	ts.service = grpclb.NewService(ts.viper.GetString("server.serviceName"),
		common.Endpoint(ts.viper.GetString("server.endpoint.ip"), ts.viper.GetInt("server.endpoint.port")),
		ts.viper.GetString("server.metadata"),
		ts.viper.GetInt64("server.discoveryTTL"))

	caFile := ts.viper.GetString("etcdCluster.tls.caFile")
	certFile := ts.viper.GetString("etcdCluster.tls.certFile")
	keyFile := ts.viper.GetString("etcdCluster.tls.keyFile")
	certPassword := ts.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTLSConfVerify(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		ts.etcdCfg = clientv3.Config{
			Endpoints:   ts.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: ts.viper.GetDuration("etcdCluster.dialTimeout"),
			TLS:         tlsConf,
		}
	} else {
		ts.etcdCfg = clientv3.Config{
			Endpoints:   ts.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: ts.viper.GetDuration("etcdCluster.dialTimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create auth server gRPC client.
func (ts *TemplateServer) initAuthServerClient() {
	ctx := &grpclb.Context{
		Target:     ts.viper.GetString("authserver.serviceName"),
		EtcdConfig: ts.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(ts.viper.GetDuration("authserver.callTimeout")),
	}

	// build gRPC client of authserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create authserver gRPC client, %+v", err)
	}
	ts.authSvrConn = conn
	ts.authSvrCli = pbauthserver.NewAuthClient(conn.Conn())
	logger.Info("create authserver gRPC client success.")
}

// create datamanager gRPC client.
func (ts *TemplateServer) initDataManagerClient() {
	ctx := &grpclb.Context{
		Target:     ts.viper.GetString("datamanager.serviceName"),
		EtcdConfig: ts.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(ts.viper.GetDuration("datamanager.callTimeout")),
	}

	// build gRPC client of datamanager server.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("create DataManager gRPC client, %+v", err)
	}
	ts.dataMgrConn = conn
	ts.dataMgrCli = pbdatamanager.NewDataManagerClient(conn.Conn())
	logger.Info("create DataManager gRPC client success.")
}

// init audit handler.
func (ts *TemplateServer) initAuditHandler() {
	audit.InitAuditHandler(ts.viper.GetInt("audit.infoChanSize"), ts.viper.GetDuration("audit.infoChanTimeout"),
		ts.dataMgrCli, ts.viper.GetDuration("datamanager.callTimeout"))
	logger.Info("init audit handler success.")
}

// initializes prometheus metrics collector.
func (ts *TemplateServer) initMetricsCollector() {
	ts.collector = metrics.NewCollector(ts.viper.GetString("metrics.endpoint"),
		ts.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := ts.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initializes action executor.
func (ts *TemplateServer) initExecutor() {
	ts.executor = executor.NewExecutor()
	logger.Info("create action executor success.")
}

// initMods initialize the server modules
func (ts *TemplateServer) initMods() {
	// initialize service discovery.
	ts.initServiceDiscovery()

	// initialize auth server gRPC client.
	ts.initAuthServerClient()

	// initialize datamanager client.
	ts.initDataManagerClient()

	// initialize audit handler.
	ts.initAuditHandler()

	// initialize metrics collector.
	ts.initMetricsCollector()

	// initialize action executor.
	ts.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(ts.viper.GetString("server.endpoint.ip"), ts.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("listen on target endpoint, %+v", err)
	}
	ts.lis = lis
}

// Run runs template server
func (ts *TemplateServer) Run() {
	// initialize config.
	ts.initConfig()

	// initialize logger.
	ts.initLogger()
	defer ts.Stop()

	// initialize server modules.
	ts.initMods()

	// register templateserver service.
	go func() {
		if err := ts.service.Register(ts.etcdCfg); err != nil {
			logger.Fatal("register templateserver service, %+v", err)
		}
	}()
	logger.Info("register templateserver service success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterTemplateServer(s, ts)
	logger.Info("Template Server running now.")

	if err := s.Serve(ts.lis); err != nil {
		logger.Fatal("start templateserver gRPC service, %+v", err)
	}
}

// Stop stops the templateserver.
func (ts *TemplateServer) Stop() {
	// close authserver gRPC connection when server exit.
	if ts.authSvrConn != nil {
		ts.authSvrConn.Close()
	}

	// close datamanager gRPC connection when server exit.
	if ts.dataMgrConn != nil {
		ts.dataMgrConn.Close()
	}

	// unregister service.
	if ts.service != nil {
		ts.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
