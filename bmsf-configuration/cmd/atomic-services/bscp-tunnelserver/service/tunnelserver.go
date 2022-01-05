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
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/metrics"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/publish"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/session"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	pb "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/framework/executor"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// TunnelServer is bscp tunnel server.
type TunnelServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// tunnelserver discovery instances.
	service *grpclb.Service

	// tunnel server gRPC service network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// session manager, handles gse plugin sessions.
	sessionMgr *session.Manager

	// handle all publish events, push notification to gse plugin(sidecars).
	publishMgr *publish.Manager

	// strategy handler, check release strategies when publish event coming.
	strategyHandler *strategy.Handler

	// gse controller gRPC connection/client.
	gseControllerConn *grpclb.GRPCConn
	gseControllerCli  pbgsecontroller.GSEControllerClient

	// datamanager gRPC connection/client.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// gse client manager.
	gseCliMgr *GSEClientManager

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewTunnelServer creates new tunnel server instance.
func NewTunnelServer() *TunnelServer {
	return &TunnelServer{}
}

// Init initialize the settings.
func (ts *TunnelServer) Init(setting framework.Setting) {
	ts.setting = setting
}

// initialize config and check base content.
func (ts *TunnelServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(ts.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	ts.viper = viper
}

// initialize logger.
func (ts *TunnelServer) initLogger() {
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

	logger.Info("dump configs: server[%+v %+v] metrics[%+v] gsecontroller[%+v] datamanager[%+v] etcdCluster[%+v] "+
		"gseTaskServer[%+v]",
		ts.viper.Get("server.endpoint.ip"), ts.viper.Get("server.endpoint.port"), ts.viper.Get("metrics"),
		ts.viper.Get("gsecontroller"), ts.viper.Get("datamanager"), ts.viper.Get("etcdCluster"),
		ts.viper.Get("gseTaskServer"))
}

// create new service struct of tunnelserver, and register service later.
func (ts *TunnelServer) initServiceDiscovery() {
	ts.service = grpclb.NewService(
		ts.viper.GetString("server.serviceName"),
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

// create tunnel session manager, handle all gse plugin sessions.
func (ts *TunnelServer) initSessionManager() {
	ts.sessionMgr = session.NewManager(ts.viper)
	logger.Info("init session manager success.")
}

// create strategy handler, it would check release strategies when publish event coming.
func (ts *TunnelServer) initStrategyHandler() {
	ts.strategyHandler = strategy.NewHandler(nil)
	logger.Info("init strategy handler success.")
}

// create publish manager, receive notification from message queue.
func (ts *TunnelServer) initPublishManager() {
	ts.publishMgr = publish.NewManager(ts.viper, ts.sessionMgr, ts.strategyHandler, ts.collector)
	logger.Info("init publish manager success.")
}

// init gse-controller gRPC connection/client.
func (ts *TunnelServer) initGSEControllerClient() {
	ctx := &grpclb.Context{
		Target:     ts.viper.GetString("gsecontroller.serviceName"),
		EtcdConfig: ts.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(ts.viper.GetDuration("gsecontroller.callTimeout")),
	}

	// build gRPC client of gse controller.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("create gse controller gRPC client, %+v", err)
	}
	ts.gseControllerConn = conn
	ts.gseControllerCli = pbgsecontroller.NewGSEControllerClient(conn.Conn())
	logger.Info("create gse controller gRPC client success.")
}

// init datamanager server gRPC connection/client.
func (ts *TunnelServer) initDataManagerClient() {
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
		logger.Fatal("create datamanager gRPC client, %+v", err)
	}
	ts.dataMgrConn = conn
	ts.dataMgrCli = pbdatamanager.NewDataManagerClient(conn.Conn())
	logger.Info("create datamanager gRPC client success.")
}

// initializes gse client manager.
func (ts *TunnelServer) initGSEClientManager() {
	ts.gseCliMgr = NewGSEClientManager(ts.viper, ts)
	if err := ts.gseCliMgr.Init(); err != nil {
		logger.Fatal("init gse client manager, %+v", err)
	}
	logger.Info("init gse taskserver client success.")
}

// initializes prometheus metrics collector.
func (ts *TunnelServer) initMetricsCollector() {
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
func (ts *TunnelServer) initExecutor() {
	ts.executor = executor.NewRateLimitExecutor(ts.viper.GetInt("server.executorLimitRate"))
	logger.Info("create action executor success.")
}

// initMods initializes the server modules.
func (ts *TunnelServer) initMods() {
	// initialize service discovery.
	ts.initServiceDiscovery()

	// initialize session manager module.
	ts.initSessionManager()

	// initialize strategy handler.
	ts.initStrategyHandler()

	// initialize metrics collector.
	ts.initMetricsCollector()

	// initialize publish manager.
	ts.initPublishManager()

	// create gse controller gRPC client.
	ts.initGSEControllerClient()

	// create datamanager gRPC client.
	ts.initDataManagerClient()

	// create gse client manager.
	ts.initGSEClientManager()

	// initialize action executor.
	ts.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(ts.viper.GetString("server.endpoint.ip"), ts.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("can't listen on local endpoint, %+v", err)
	}
	ts.lis = lis
	logger.Info("listen on local endpoint success.")
}

// collect collects local normal metrics data.
func (ts *TunnelServer) collect() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// session count.
		if count, err := ts.sessionMgr.SessionCount(); err == nil {
			logger.Info("STAT | session manager, plugin sidecar session infos, count[%d]", count)
			ts.collector.StatSessionNum(count)
		}
	}
}

// Run runs tunnel server
func (ts *TunnelServer) Run() {
	// initialize config.
	ts.initConfig()

	// initialize logger.
	ts.initLogger()
	defer ts.Stop()

	// initialize server modules.
	ts.initMods()

	// collect normal metrics data.
	go ts.collect()

	// register tunnelserver service.
	go func() {
		if err := ts.service.Register(ts.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterTunnelServer(s, ts)
	logger.Info("Tunnel Server running now.")

	if err := s.Serve(ts.lis); err != nil {
		logger.Fatal("start tunnelserver gRPC service, %+v", err)
	}
}

// Stop stops the tunnelserver.
func (ts *TunnelServer) Stop() {
	// close gse controller gRPC connection.
	if ts.gseControllerConn != nil {
		ts.gseControllerConn.Close()
	}

	// close datamanager gRPC connection.
	if ts.dataMgrConn != nil {
		ts.dataMgrConn.Close()
	}

	// close gse clients.
	if ts.gseCliMgr != nil {
		ts.gseCliMgr.Close()
	}

	// unregister service.
	if ts.service != nil {
		ts.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
