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

	"bk-bscp/cmd/bscp-accessserver/modules/metrics"
	"bk-bscp/internal/framework"
	"bk-bscp/internal/framework/executor"
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbintegrator "bk-bscp/internal/protocol/integrator"
	pbtemplateserver "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
)

// AccessServer is bscp access server.
type AccessServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// acessserver discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// business server gRPC connection/client.
	businessSvrConn *grpclb.GRPCConn
	businessSvrCli  pbbusinessserver.BusinessClient

	// template server gRPC connection/client.
	templateSvrConn *grpclb.GRPCConn
	templateSvrCli  pbtemplateserver.TemplateClient

	// integrator gRPC connection/client.
	itgConn *grpclb.GRPCConn
	itgCli  pbintegrator.IntegratorClient

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewAccessServer creates a new access server instance.
func NewAccessServer() *AccessServer {
	return &AccessServer{}
}

// Init initialize the settings.
func (as *AccessServer) Init(setting framework.Setting) {
	as.setting = setting
}

// initialize config and check base content.
func (as *AccessServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(as.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	as.viper = viper
}

// initialize logger.
func (as *AccessServer) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          as.viper.GetString("logger.directory"),
		LogMaxSize:      as.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       as.viper.GetInt("logger.maxnum"),
		ToStdErr:        as.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    as.viper.GetBool("logger.alsoStderr"),
		Verbosity:       as.viper.GetInt32("logger.level"),
		StdErrThreshold: as.viper.GetString("logger.stderrThreshold"),
		VModule:         as.viper.GetString("logger.vmodule"),
		TraceLocation:   as.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		as.viper.GetString("logger.directory"), as.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v, %+v, %+v, %+v] auth[%+v, %+v] metrics[%+v] businessserver[%+v, %+v] integrator[%+v %+v] templateserver[%+v, %+v] etcdCluster[%+v, %+v]",
		as.viper.Get("server.servicename"), as.viper.Get("server.endpoint.ip"), as.viper.Get("server.endpoint.port"), as.viper.Get("server.discoveryttl"), as.viper.Get("auth.open"),
		as.viper.Get("auth.admin"), as.viper.Get("metrics.endpoint"), as.viper.Get("businessserver.servicename"), as.viper.Get("businessserver.calltimeout"), as.viper.Get("integrator.servicename"),
		as.viper.Get("integrator.calltimeout"), as.viper.Get("templateserver.servicename"), as.viper.Get("templateserver.calltimeout"), as.viper.Get("etcdCluster.endpoints"), as.viper.Get("etcdCluster.dialtimeout"))
}

// create new service struct of accessserver, and register service later.
func (as *AccessServer) initServiceDiscovery() {
	as.service = grpclb.NewService(
		as.viper.GetString("server.servicename"),
		common.Endpoint(as.viper.GetString("server.endpoint.ip"), as.viper.GetInt("server.endpoint.port")),
		as.viper.GetString("server.metadata"),
		as.viper.GetInt64("server.discoveryttl"))

	caFile := as.viper.GetString("etcdCluster.tls.cafile")
	certFile := as.viper.GetString("etcdCluster.tls.certfile")
	keyFile := as.viper.GetString("etcdCluster.tls.keyfile")
	certPassword := as.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTslConfVerity(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		as.etcdCfg = clientv3.Config{
			Endpoints:   as.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: as.viper.GetDuration("etcdCluster.dialtimeout"),
			TLS:         tlsConf,
		}
	} else {
		as.etcdCfg = clientv3.Config{
			Endpoints:   as.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: as.viper.GetDuration("etcdCluster.dialtimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create business server gRPC client.
func (as *AccessServer) initBusinessClient() {
	ctx := &grpclb.Context{
		Target:     as.viper.GetString("businessserver.servicename"),
		EtcdConfig: as.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(as.viper.GetDuration("businessserver.calltimeout")),
	}

	// build gRPC client of businessserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create businessserver gRPC client, %+v", err)
	}
	as.businessSvrConn = conn
	as.businessSvrCli = pbbusinessserver.NewBusinessClient(conn.Conn())
	logger.Info("create businessserver gRPC client success.")
}

// create template server gRPC client.
func (as *AccessServer) initTemplateClient() {
	ctx := &grpclb.Context{
		Target:     as.viper.GetString("templateserver.servicename"),
		EtcdConfig: as.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(as.viper.GetDuration("templateserver.calltimeout")),
	}

	// build gRPC client of templateserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create templateserver gRPC client, %+v", err)
	}
	as.templateSvrConn = conn
	as.templateSvrCli = pbtemplateserver.NewTemplateClient(conn.Conn())
	logger.Info("create templateserver gRPC client success.")
}

// create integrator gRPC client.
func (as *AccessServer) initIntegratorClient() {
	ctx := &grpclb.Context{
		Target:     as.viper.GetString("integrator.servicename"),
		EtcdConfig: as.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(as.viper.GetDuration("integrator.calltimeout")),
	}

	// build gRPC client of integrator.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create integrator gRPC client, %+v", err)
	}
	as.itgConn = conn
	as.itgCli = pbintegrator.NewIntegratorClient(conn.Conn())
	logger.Info("create integrator gRPC client success.")
}

// initializes prometheus metrics collector.
func (as *AccessServer) initMetricsCollector() {
	as.collector = metrics.NewCollector(as.viper.GetString("metrics.endpoint"),
		as.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := as.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initializes action executor.
func (as *AccessServer) initExecutor() {
	as.executor = executor.NewRateLimitExecutor(as.viper.GetInt("server.executorLimitRate"))
	logger.Info("create action executor success.")
}

// initMods initializes the server modules.
func (as *AccessServer) initMods() {
	// initialize service discovery.
	as.initServiceDiscovery()

	// initialize business server gRPC client.
	as.initBusinessClient()

	// initialize template server gRPC client.
	as.initTemplateClient()

	// initialize integrator gRPC client.
	as.initIntegratorClient()

	// initialize metrics collector.
	as.initMetricsCollector()

	// initialize action executor.
	as.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(as.viper.GetString("server.endpoint.ip"), as.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("listen on target endpoint, %+v", err)
	}
	as.lis = lis
}

// Run runs access server
func (as *AccessServer) Run() {
	// initialize config.
	as.initConfig()

	// initialize logger.
	as.initLogger()
	defer as.Stop()

	// initialize server modules.
	as.initMods()

	// register accessserver service.
	go func() {
		if err := as.service.Register(as.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterAccessServer(s, as)
	logger.Info("Access Server running now.")

	if err := s.Serve(as.lis); err != nil {
		logger.Fatal("start accessserver gRPC service. %+v", err)
	}
}

// Stop stops the accessserver.
func (as *AccessServer) Stop() {
	// close integrator gRPC connection when server exit.
	if as.itgConn != nil {
		as.itgConn.Close()
	}

	// close template server gRPC connection when server exit.
	if as.templateSvrConn != nil {
		as.templateSvrConn.Close()
	}

	// close business server gRPC connection when server exit.
	if as.businessSvrConn != nil {
		as.businessSvrConn.Close()
	}

	// unregister service.
	if as.service != nil {
		as.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
