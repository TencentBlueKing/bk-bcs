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

	"bk-bscp/cmd/bscp-bcs-controller/modules/metrics"
	"bk-bscp/internal/framework"
	"bk-bscp/internal/framework/executor"
	pb "bk-bscp/internal/protocol/bcs-controller"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	mq "bk-bscp/pkg/natsmq"
)

// BCSController is bscp bcs controller.
type BCSController struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// bcs controller discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// datamanager gRPC connection/client.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// release publisher.
	publisher *mq.Publisher

	// release publish topic.
	pubTopic string

	// strategy handler, check release strategies.
	strategyHandler *strategy.Handler

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewBCSController creates new bcs controller instance.
func NewBCSController() *BCSController {
	return &BCSController{}
}

// Init initialize the settings.
func (c *BCSController) Init(setting framework.Setting) {
	c.setting = setting
}

// initialize config and check base content.
func (c *BCSController) initConfig() {
	cfg := config{}
	viper, err := cfg.init(c.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	c.viper = viper
}

// initialize logger.
func (c *BCSController) initLogger() {
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

	logger.Info("dump configs: server[%+v, %+v, %+v, %+v, %+v] metrics[%+v] datamanager[%+v, %+v] nats[%+v, %+v, %+v, %+v, %+v] etcdCluster[%+v, %+v]",
		c.viper.Get("server.servicename"), c.viper.Get("server.endpoint.ip"), c.viper.Get("server.endpoint.port"), c.viper.Get("server.discoveryttl"),
		c.viper.Get("server.queryNewestLimit"), c.viper.Get("metrics.endpoint"), c.viper.Get("datamanager.servicename"), c.viper.Get("datamanager.calltimeout"),
		c.viper.Get("natsmqCluster.endpoints"), c.viper.Get("natsmqCluster.timeout"), c.viper.Get("natsmqCluster.reconwait"), c.viper.Get("natsmqCluster.maxrecons"),
		c.viper.Get("natsmqCluster.publishtopic"), c.viper.Get("etcdCluster.endpoints"), c.viper.Get("etcdCluster.dialtimeout"))
}

// create new service struct of bcs controller, and register service later.
func (c *BCSController) initServiceDiscovery() {
	c.service = grpclb.NewService(
		c.viper.GetString("server.servicename"),
		common.Endpoint(c.viper.GetString("server.endpoint.ip"), c.viper.GetInt("server.endpoint.port")),
		c.viper.GetString("server.metadata"),
		c.viper.GetInt64("server.discoveryttl"))

	caFile := c.viper.GetString("etcdCluster.tls.cafile")
	certFile := c.viper.GetString("etcdCluster.tls.certfile")
	keyFile := c.viper.GetString("etcdCluster.tls.keyfile")
	certPassword := c.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTslConfVerity(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		c.etcdCfg = clientv3.Config{
			Endpoints:   c.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: c.viper.GetDuration("etcdCluster.dialtimeout"),
			TLS:         tlsConf,
		}
	} else {
		c.etcdCfg = clientv3.Config{
			Endpoints:   c.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: c.viper.GetDuration("etcdCluster.dialtimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create datamanager gRPC client.
func (c *BCSController) initDataManagerClient() {
	ctx := &grpclb.Context{
		Target:     c.viper.GetString("datamanager.servicename"),
		EtcdConfig: c.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(c.viper.GetDuration("datamanager.calltimeout")),
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

// create config release publisher.
func (c *BCSController) initPublisher() {
	timeout := c.viper.GetDuration("natsmqCluster.timeout")
	reconWait := c.viper.GetDuration("natsmqCluster.reconwait")
	maxRecons := c.viper.GetInt("natsmqCluster.maxrecons")

	publisher := mq.NewPublisher(c.viper.GetStringSlice("natsmqCluster.endpoints"))
	if err := publisher.Init(timeout, reconWait, maxRecons); err != nil {
		logger.Fatal("init publisher base on natsmq, %+v", err)
	}

	c.publisher = publisher
	c.pubTopic = c.viper.GetString("natsmqCluster.publishtopic")
	logger.Info("init publisher base on natsmq success.")
}

// create strategy handler, it would check release strategies.
func (c *BCSController) initStrategyHandler() {
	c.strategyHandler = strategy.NewHandler(nil)
	logger.Info("init strategy handler success.")
}

// initializes prometheus metrics collector.
func (c *BCSController) initMetricsCollector() {
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
func (c *BCSController) initExecutor() {
	c.executor = executor.NewExecutor()
	logger.Info("create action executor success.")
}

// initMods initialize the server modules.
func (c *BCSController) initMods() {
	// initialize service discovery.
	c.initServiceDiscovery()

	// initialize datamanager client.
	c.initDataManagerClient()

	// initialize publisher.
	c.initPublisher()

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

// Run runs bcs controller.
func (c *BCSController) Run() {
	// initialize config.
	c.initConfig()

	// initialize logger.
	c.initLogger()
	defer c.Stop()

	// initialize server modules.
	c.initMods()

	// register bcs controller service.
	go func() {
		if err := c.service.Register(c.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterBCSControllerServer(s, c)
	logger.Info("BCS Controller running now.")

	if err := s.Serve(c.lis); err != nil {
		logger.Fatal("start bcs controller gRPC service. %+v", err)
	}
}

// Stop stops the bcs controller.
func (c *BCSController) Stop() {
	// close datamanager gRPC connection when server exit.
	if c.dataMgrConn != nil {
		c.dataMgrConn.Close()
	}

	// close release message queue when server exit.
	if c.publisher != nil {
		c.publisher.Close()
	}

	// unregister service.
	if c.service != nil {
		c.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
