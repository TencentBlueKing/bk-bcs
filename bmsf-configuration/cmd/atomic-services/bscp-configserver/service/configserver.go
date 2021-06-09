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

	"bk-bscp/cmd/atomic-services/bscp-configserver/modules/metrics"
	"bk-bscp/internal/audit"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/framework/executor"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// ConfigServer is bscp config server.
type ConfigServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// configserver discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// authserver gRPC connection/client.
	authSvrConn *grpclb.GRPCConn
	authSvrCli  pbauthserver.AuthClient

	// datamanager server gRPC connection/client.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// gse controller gRPC connection/client.
	gseControllerConn *grpclb.GRPCConn
	gseControllerCli  pbgsecontroller.GSEControllerClient

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewConfigServer creates new config server instance.
func NewConfigServer() *ConfigServer {
	return &ConfigServer{}
}

// Init initialize the settings.
func (cs *ConfigServer) Init(setting framework.Setting) {
	cs.setting = setting
}

// initialize config and check base content.
func (cs *ConfigServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(cs.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	cs.viper = viper
}

// initialize logger.
func (cs *ConfigServer) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          cs.viper.GetString("logger.directory"),
		LogMaxSize:      cs.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       cs.viper.GetInt("logger.maxnum"),
		ToStdErr:        cs.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    cs.viper.GetBool("logger.alsoStderr"),
		Verbosity:       cs.viper.GetInt32("logger.level"),
		StdErrThreshold: cs.viper.GetString("logger.stderrThreshold"),
		VModule:         cs.viper.GetString("logger.vmodule"),
		TraceLocation:   cs.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		cs.viper.GetString("logger.directory"), cs.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v %+v] metrics[%+v] datamanager[%+v] gsecontroller[%+v] etcdCluster[%+v]",
		cs.viper.Get("server.endpoint.ip"), cs.viper.Get("server.endpoint.port"), cs.viper.Get("metrics"),
		cs.viper.Get("datamanager"), cs.viper.Get("gsecontroller"), cs.viper.Get("etcdCluster"))
}

// create new service struct of configserver, and register service later.
func (cs *ConfigServer) initServiceDiscovery() {
	cs.service = grpclb.NewService(
		cs.viper.GetString("server.serviceName"),
		common.Endpoint(cs.viper.GetString("server.endpoint.ip"), cs.viper.GetInt("server.endpoint.port")),
		cs.viper.GetString("server.metadata"),
		cs.viper.GetInt64("server.discoveryTTL"))

	caFile := cs.viper.GetString("etcdCluster.tls.caFile")
	certFile := cs.viper.GetString("etcdCluster.tls.certFile")
	keyFile := cs.viper.GetString("etcdCluster.tls.keyFile")
	certPassword := cs.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTLSConfVerify(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		cs.etcdCfg = clientv3.Config{
			Endpoints:   cs.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: cs.viper.GetDuration("etcdCluster.dialTimeout"),
			TLS:         tlsConf,
		}
	} else {
		cs.etcdCfg = clientv3.Config{
			Endpoints:   cs.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: cs.viper.GetDuration("etcdCluster.dialTimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create auth server gRPC client.
func (cs *ConfigServer) initAuthServerClient() {
	ctx := &grpclb.Context{
		Target:     cs.viper.GetString("authserver.serviceName"),
		EtcdConfig: cs.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(cs.viper.GetDuration("authserver.callTimeout")),
	}

	// build gRPC client of authserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create authserver gRPC client, %+v", err)
	}
	cs.authSvrConn = conn
	cs.authSvrCli = pbauthserver.NewAuthClient(conn.Conn())
	logger.Info("create authserver gRPC client success.")
}

// create datamanager server gRPC client.
func (cs *ConfigServer) initDataManagerClient() {
	ctx := &grpclb.Context{
		Target:     cs.viper.GetString("datamanager.serviceName"),
		EtcdConfig: cs.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(cs.viper.GetDuration("datamanager.callTimeout")),
	}

	// build gRPC client of datamanager.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create datamanager gRPC client, %+v", err)
	}
	cs.dataMgrConn = conn
	cs.dataMgrCli = pbdatamanager.NewDataManagerClient(conn.Conn())
	logger.Info("create datamanager gRPC client success.")
}

// init audit handler.
func (cs *ConfigServer) initAuditHandler() {
	audit.InitAuditHandler(cs.viper.GetInt("audit.infoChanSize"), cs.viper.GetDuration("audit.infoChanTimeout"),
		cs.dataMgrCli, cs.viper.GetDuration("datamanager.callTimeout"))
	logger.Info("init audit handler success.")
}

// create gse controller gRPC client.
func (cs *ConfigServer) initGSEControllerClient() {
	ctx := &grpclb.Context{
		Target:     cs.viper.GetString("gsecontroller.serviceName"),
		EtcdConfig: cs.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(cs.viper.GetDuration("gsecontroller.callTimeout")),
	}

	// build gRPC client of gsecontroller.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create gsecontroller gRPC client, %+v", err)
	}
	cs.gseControllerConn = conn
	cs.gseControllerCli = pbgsecontroller.NewGSEControllerClient(conn.Conn())
	logger.Info("create gse-controller gRPC client success.")
}

// initializes prometheus metrics collector.
func (cs *ConfigServer) initMetricsCollector() {
	cs.collector = metrics.NewCollector(cs.viper.GetString("metrics.endpoint"), cs.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := cs.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initializes action executor.
func (cs *ConfigServer) initExecutor() {
	cs.executor = executor.NewRateLimitExecutor(cs.viper.GetInt("server.executorLimitRate"))
	logger.Info("create action executor success.")
}

// initMods initializes the server modules.
func (cs *ConfigServer) initMods() {
	// initialize service discovery.
	cs.initServiceDiscovery()

	// initialize auth server gRPC client.
	cs.initAuthServerClient()

	// initialize datamanager server gRPC client.
	cs.initDataManagerClient()

	// initialize gse controller gRPC client.
	cs.initGSEControllerClient()

	// initialize audit handler.
	cs.initAuditHandler()

	// initialize metrics collector.
	cs.initMetricsCollector()

	// initialize action executor.
	cs.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(cs.viper.GetString("server.endpoint.ip"), cs.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("listen on target endpoint, %+v", err)
	}
	cs.lis = lis
}

// Run runs config server.
func (cs *ConfigServer) Run() {
	// initialize config.
	cs.initConfig()

	// initialize logger.
	cs.initLogger()
	defer cs.Stop()

	// initialize server modules.
	cs.initMods()

	// register configserver service.
	go func() {
		if err := cs.service.Register(cs.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterConfigServer(s, cs)
	logger.Info("Config Server running now.")

	if err := s.Serve(cs.lis); err != nil {
		logger.Fatal("start configserver gRPC service. %+v", err)
	}
}

// Stop stops the configserver.
func (cs *ConfigServer) Stop() {
	// close authserver gRPC connection when server exit.
	if cs.authSvrConn != nil {
		cs.authSvrConn.Close()
	}

	// close datamanager server gRPC connection when server exit.
	if cs.dataMgrConn != nil {
		cs.dataMgrConn.Close()
	}

	// close gse controller gRPC connection when server exit.
	if cs.gseControllerConn != nil {
		cs.gseControllerConn.Close()
	}

	// unregister service.
	if cs.service != nil {
		cs.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
