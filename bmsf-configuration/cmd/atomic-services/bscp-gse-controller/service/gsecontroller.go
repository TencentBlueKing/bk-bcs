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

	"bk-bscp/cmd/atomic-services/bscp-gse-controller/modules/metrics"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/gse-controller"
	pbtunnelserver "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/framework/executor"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// GSEController is bscp gse controller.
type GSEController struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// gse controller discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// datamanager gRPC connection/client.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// tunnelserver gRPC connection/client.
	tunnelServerConn *grpclb.GRPCConn
	tunnelServerCli  pbtunnelserver.TunnelClient

	// strategy handler, check release strategies.
	strategyHandler *strategy.Handler

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewGSEController creates new gse controller instance.
func NewGSEController() *GSEController {
	return &GSEController{}
}

// Init initialize the settings.
func (c *GSEController) Init(setting framework.Setting) {
	c.setting = setting
}

// initialize config and check base content.
func (c *GSEController) initConfig() {
	cfg := config{}
	viper, err := cfg.init(c.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	c.viper = viper
}

// initialize logger.
func (c *GSEController) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          c.viper.GetString("logger.directory"),
		LogMaxSize:      c.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       c.viper.GetInt("logger.maxnum"),
		ToStdErr:        c.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    c.viper.GetBool("logger.alsoStderr"),
		Verbosity:       c.viper.GetInt32("logger.level"),
		StdErrThreshold: c.viper.GetString("logger.stderrThreshold"),
		VModule:         c.viper.GetString("logger.vmodule"),
		TraceLocation:   c.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		c.viper.GetString("logger.directory"), c.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v %+v] metrics[%+v] datamanager[%+v] tunnelserver[%+v] etcdCluster[%+v]",
		c.viper.Get("server.endpoint.ip"), c.viper.Get("server.endpoint.port"), c.viper.Get("metrics"),
		c.viper.Get("datamanager"), c.viper.Get("tunnelserver"), c.viper.Get("etcdCluster"))
}

// create new service struct of gse controller, and register service later.
func (c *GSEController) initServiceDiscovery() {
	c.service = grpclb.NewService(
		c.viper.GetString("server.serviceName"),
		common.Endpoint(c.viper.GetString("server.endpoint.ip"), c.viper.GetInt("server.endpoint.port")),
		c.viper.GetString("server.metadata"),
		c.viper.GetInt64("server.discoveryTTL"))

	caFile := c.viper.GetString("etcdCluster.tls.caFile")
	certFile := c.viper.GetString("etcdCluster.tls.certFile")
	keyFile := c.viper.GetString("etcdCluster.tls.keyFile")
	certPassword := c.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTLSConfVerify(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		c.etcdCfg = clientv3.Config{
			Endpoints:   c.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: c.viper.GetDuration("etcdCluster.dialTimeout"),
			TLS:         tlsConf,
		}

	} else {
		c.etcdCfg = clientv3.Config{
			Endpoints:   c.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: c.viper.GetDuration("etcdCluster.dialTimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create datamanager gRPC client.
func (c *GSEController) initDataManagerClient() {
	ctx := &grpclb.Context{
		Target:     c.viper.GetString("datamanager.serviceName"),
		EtcdConfig: c.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(c.viper.GetDuration("datamanager.callTimeout")),
	}

	// build gRPC client of datamanager.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("create datamanager gRPC client, %+v", err)
	}
	c.dataMgrConn = conn
	c.dataMgrCli = pbdatamanager.NewDataManagerClient(conn.Conn())
	logger.Info("create datamanager gRPC client success.")
}

// create tunnelserver gRPC client.
func (c *GSEController) initTunnelServerClient() {
	ctx := &grpclb.Context{
		Target:     c.viper.GetString("tunnelserver.serviceName"),
		EtcdConfig: c.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(c.viper.GetDuration("tunnelserver.callTimeout")),
	}

	// build gRPC client of tunnelserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("create tunnelserver gRPC client, %+v", err)
	}
	c.tunnelServerConn = conn
	c.tunnelServerCli = pbtunnelserver.NewTunnelClient(conn.Conn())
	logger.Info("create tunnelserver gRPC client success.")
}

// create strategy handler, it would check release strategies.
func (c *GSEController) initStrategyHandler() {
	c.strategyHandler = strategy.NewHandler(nil)
	logger.Info("init strategy handler success.")
}

// initializes prometheus metrics collector.
func (c *GSEController) initMetricsCollector() {
	c.collector = metrics.NewCollector(c.viper.GetString("metrics.endpoint"),
		c.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := c.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initializes action executor.
func (c *GSEController) initExecutor() {
	c.executor = executor.NewExecutor()
	logger.Info("create action executor success.")
}

// initMods initialize the server modules.
func (c *GSEController) initMods() {
	// initialize service discovery.
	c.initServiceDiscovery()

	// initialize datamanager client.
	c.initDataManagerClient()

	// initialize tunnelserver client.
	c.initTunnelServerClient()

	// initialize strategy handler.
	c.initStrategyHandler()

	// initialize metrics collector.
	c.initMetricsCollector()

	// initialize action executor.
	c.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(c.viper.GetString("server.endpoint.ip"), c.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("listen on target endpoint, %+v", err)
	}
	c.lis = lis
}

// Run runs gse controller.
func (c *GSEController) Run() {
	// initialize config.
	c.initConfig()

	// initialize logger.
	c.initLogger()
	defer c.Stop()

	// initialize server modules.
	c.initMods()

	// register gse controller service.
	go func() {
		if err := c.service.Register(c.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterGSEControllerServer(s, c)
	logger.Info("GSE Controller running now.")

	if err := s.Serve(c.lis); err != nil {
		logger.Fatal("start gse controller gRPC service. %+v", err)
	}
}

// Stop stops the gse controller.
func (c *GSEController) Stop() {
	// close tunnel server gRPC connection when server exit.
	if c.tunnelServerConn != nil {
		c.tunnelServerConn.Close()
	}

	// close datamanager gRPC connection when server exit.
	if c.dataMgrConn != nil {
		c.dataMgrConn.Close()
	}

	// unregister service.
	if c.service != nil {
		c.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
