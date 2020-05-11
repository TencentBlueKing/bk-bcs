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

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/cmd/bscp-businessserver/modules/metrics"
	"bk-bscp/internal/framework"
	"bk-bscp/internal/framework/executor"
	pbbcscontroller "bk-bscp/internal/protocol/bcs-controller"
	pb "bk-bscp/internal/protocol/businessserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	pbtemplateserver "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
)

// BusinessServer is bscp business server.
type BusinessServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// businessserver discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// datamanager server gRPC connection/client.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// template server gRPC connection/client.
	templateSvrConn *grpclb.GRPCConn
	templateSvrCli  pbtemplateserver.TemplateClient

	// bcs controller gRPC connection/client.
	bcsControllerConn *grpclb.GRPCConn
	bcsControllerCli  pbbcscontroller.BCSControllerClient

	// gse controller gRPC connection/client.
	gseControllerConn *grpclb.GRPCConn
	gseControllerCli  pbgsecontroller.GSEControllerClient

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewBusinessServer creates new business server instance.
func NewBusinessServer() *BusinessServer {
	return &BusinessServer{}
}

// Init initialize the settings.
func (bs *BusinessServer) Init(setting framework.Setting) {
	bs.setting = setting
}

// initialize config and check base content.
func (bs *BusinessServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(bs.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	bs.viper = viper
}

// initialize logger.
func (bs *BusinessServer) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          bs.viper.GetString("logger.directory"),
		LogMaxSize:      bs.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       bs.viper.GetInt("logger.maxnum"),
		ToStdErr:        bs.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    bs.viper.GetBool("logger.alsoStderr"),
		Verbosity:       bs.viper.GetInt32("logger.level"),
		StdErrThreshold: bs.viper.GetString("logger.stderrThreshold"),
		VModule:         bs.viper.GetString("logger.vmodule"),
		TraceLocation:   bs.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		bs.viper.GetString("logger.directory"), bs.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v, %+v, %+v, %+v] metrics[%+v] templateserver[%+v, %+v] datamanager[%+v, %+v] bcscontroller[%+v, %+v] gsecontroller[%+v, %+v] etcdCluster[%+v, %+v]",
		bs.viper.Get("server.servicename"), bs.viper.Get("server.endpoint.ip"), bs.viper.Get("server.endpoint.port"), bs.viper.Get("server.discoveryttl"), bs.viper.Get("metrics.endpoint"),
		bs.viper.Get("templateserver.servicename"), bs.viper.Get("templateserver.calltimeout"), bs.viper.Get("datamanager.servicename"), bs.viper.Get("datamanager.calltimeout"),
		bs.viper.Get("bcscontroller.servicename"), bs.viper.Get("bcscontroller.calltimeout"), bs.viper.Get("gsecontroller.servicename"), bs.viper.Get("gsecontroller.calltimeout"),
		bs.viper.Get("etcdCluster.endpoints"), bs.viper.Get("etcdCluster.dialtimeout"))
}

// create new service struct of businessserver, and register service later.
func (bs *BusinessServer) initServiceDiscovery() {
	bs.service = grpclb.NewService(
		bs.viper.GetString("server.servicename"),
		common.Endpoint(bs.viper.GetString("server.endpoint.ip"), bs.viper.GetInt("server.endpoint.port")),
		bs.viper.GetString("server.metadata"),
		bs.viper.GetInt64("server.discoveryttl"))

	caFile := bs.viper.GetString("etcdCluster.tls.cafile")
	certFile := bs.viper.GetString("etcdCluster.tls.certfile")
	keyFile := bs.viper.GetString("etcdCluster.tls.keyfile")
	certPassword := bs.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTslConfVerity(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		bs.etcdCfg = clientv3.Config{
			Endpoints:   bs.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: bs.viper.GetDuration("etcdCluster.dialtimeout"),
			TLS:         tlsConf,
		}
	} else {
		bs.etcdCfg = clientv3.Config{
			Endpoints:   bs.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: bs.viper.GetDuration("etcdCluster.dialtimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create datamanager server gRPC client.
func (bs *BusinessServer) initDataManagerClient() {
	ctx := &grpclb.Context{
		Target:     bs.viper.GetString("datamanager.servicename"),
		EtcdConfig: bs.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(bs.viper.GetDuration("datamanager.calltimeout")),
	}

	// build gRPC client of datamanager.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create datamanager gRPC client, %+v", err)
	}
	bs.dataMgrConn = conn
	bs.dataMgrCli = pbdatamanager.NewDataManagerClient(conn.Conn())
	logger.Info("create datamanager gRPC client success.")
}

// init audit handler.
func (bs *BusinessServer) initAuditHandler() {
	audit.InitAuditHandler(bs.viper, bs.dataMgrCli)
	logger.Info("init audit handler success.")
}

// create template server gRPC client.
func (bs *BusinessServer) initTemplateServerClient() {
	ctx := &grpclb.Context{
		Target:     bs.viper.GetString("templateserver.servicename"),
		EtcdConfig: bs.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(bs.viper.GetDuration("templateserver.calltimeout")),
	}

	// build gRPC client of templateserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create templateserver gRPC client, %+v", err)
	}
	bs.templateSvrConn = conn
	bs.templateSvrCli = pbtemplateserver.NewTemplateClient(conn.Conn())
	logger.Info("create templateserver gRPC client success.")
}

// create bcs controller gRPC client.
func (bs *BusinessServer) initBCSControllerClient() {
	ctx := &grpclb.Context{
		Target:     bs.viper.GetString("bcscontroller.servicename"),
		EtcdConfig: bs.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(bs.viper.GetDuration("bcscontroller.calltimeout")),
	}

	// build gRPC client of bcscontroller.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create bcscontroller gRPC client, %+v", err)
	}
	bs.bcsControllerConn = conn
	bs.bcsControllerCli = pbbcscontroller.NewBCSControllerClient(conn.Conn())
	logger.Info("create bcs-controller gRPC client success.")
}

// create gse controller gRPC client.
func (bs *BusinessServer) initGSEControllerClient() {
	ctx := &grpclb.Context{
		Target:     bs.viper.GetString("gsecontroller.servicename"),
		EtcdConfig: bs.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(bs.viper.GetDuration("gsecontroller.calltimeout")),
	}

	// build gRPC client of gsecontroller.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create gsecontroller gRPC client, %+v", err)
	}
	bs.gseControllerConn = conn
	bs.gseControllerCli = pbgsecontroller.NewGSEControllerClient(conn.Conn())
	logger.Info("create gse-controller gRPC client success.")
}

// initializes prometheus metrics collector.
func (bs *BusinessServer) initMetricsCollector() {
	bs.collector = metrics.NewCollector(bs.viper.GetString("metrics.endpoint"),
		bs.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := bs.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initializes action executor.
func (bs *BusinessServer) initExecutor() {
	bs.executor = executor.NewExecutor()
	logger.Info("create action executor success.")
}

// initMods initializes the server modules.
func (bs *BusinessServer) initMods() {
	// initialize service discovery.
	bs.initServiceDiscovery()

	// initialize datamanager server gRPC client.
	bs.initDataManagerClient()

	// initialize templateserver gRPC client.
	bs.initTemplateServerClient()

	// initialize bcs controller gRPC client.
	bs.initBCSControllerClient()

	// initialize gse controller gRPC client.
	bs.initGSEControllerClient()

	// initialize audit handler.
	bs.initAuditHandler()

	// initialize metrics collector.
	bs.initMetricsCollector()

	// initialize action executor.
	bs.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(bs.viper.GetString("server.endpoint.ip"), bs.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("listen on target endpoint, %+v", err)
	}
	bs.lis = lis
}

// Run runs business server.
func (bs *BusinessServer) Run() {
	// initialize config.
	bs.initConfig()

	// initialize logger.
	bs.initLogger()
	defer bs.Stop()

	// initialize server modules.
	bs.initMods()

	// register businessserver service.
	go func() {
		if err := bs.service.Register(bs.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterBusinessServer(s, bs)
	logger.Info("Business Server running now.")

	if err := s.Serve(bs.lis); err != nil {
		logger.Fatal("start businessserver gRPC service. %+v", err)
	}
}

// Stop stops the businessserver.
func (bs *BusinessServer) Stop() {
	// close datamanager server gRPC connection when server exit.
	if bs.dataMgrConn != nil {
		bs.dataMgrConn.Close()
	}

	// close templateserver gRPC connection when server exit.
	if bs.templateSvrConn != nil {
		bs.templateSvrConn.Close()
	}

	// close bcs controller gRPC connection when server exit.
	if bs.bcsControllerConn != nil {
		bs.bcsControllerConn.Close()
	}

	// close gse controller gRPC connection when server exit.
	if bs.gseControllerConn != nil {
		bs.gseControllerConn.Close()
	}

	// unregister service.
	if bs.service != nil {
		bs.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
