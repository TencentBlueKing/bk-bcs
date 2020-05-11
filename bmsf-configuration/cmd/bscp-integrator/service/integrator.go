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

	"bk-bscp/cmd/bscp-integrator/modules/metrics"
	"bk-bscp/internal/framework"
	"bk-bscp/internal/framework/executor"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pb "bk-bscp/internal/protocol/integrator"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
)

// Integrator is bscp integrator.
type Integrator struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// integrator discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// business server gRPC connection/client.
	businessSvrConn *grpclb.GRPCConn
	businessSvrCli  pbbusinessserver.BusinessClient

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewIntegrator creates new integrator instance.
func NewIntegrator() *Integrator {
	return &Integrator{}
}

// Init initialize the settings.
func (itg *Integrator) Init(setting framework.Setting) {
	itg.setting = setting
}

// initialize config and check base content.
func (itg *Integrator) initConfig() {
	cfg := config{}
	viper, err := cfg.init(itg.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	itg.viper = viper
}

// initialize logger.
func (itg *Integrator) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          itg.viper.GetString("logger.directory"),
		LogMaxSize:      itg.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       itg.viper.GetInt("logger.maxnum"),
		ToStdErr:        itg.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    itg.viper.GetBool("logger.alsoStderr"),
		Verbosity:       itg.viper.GetInt32("logger.level"),
		StdErrThreshold: itg.viper.GetString("logger.stderrThreshold"),
		VModule:         itg.viper.GetString("logger.vmodule"),
		TraceLocation:   itg.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		itg.viper.GetString("logger.directory"), itg.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v, %+v, %+v, %+v] metrics[%+v] businessserver[%+v, %+v] etcdCluster[%+v, %+v]",
		itg.viper.Get("server.servicename"), itg.viper.Get("server.endpoint.ip"), itg.viper.Get("server.endpoint.port"),
		itg.viper.Get("server.discoveryttl"), itg.viper.Get("metrics.endpoint"), itg.viper.Get("businessserver.servicename"),
		itg.viper.Get("businessserver.calltimeout"), itg.viper.Get("etcdCluster.endpoints"), itg.viper.Get("etcdCluster.dialtimeout"))
}

// create new service struct of integrator, and register service later.
func (itg *Integrator) initServiceDiscovery() {
	itg.service = grpclb.NewService(
		itg.viper.GetString("server.servicename"),
		common.Endpoint(itg.viper.GetString("server.endpoint.ip"), itg.viper.GetInt("server.endpoint.port")),
		itg.viper.GetString("server.metadata"),
		itg.viper.GetInt64("server.discoveryttl"))

	itg.etcdCfg = clientv3.Config{
		Endpoints:   itg.viper.GetStringSlice("etcdCluster.endpoints"),
		DialTimeout: itg.viper.GetDuration("etcdCluster.dialtimeout"),
	}
	logger.Info("init service discovery success.")
}

// create businessserver gRPC client.
func (itg *Integrator) initBusinessClient() {
	ctx := &grpclb.Context{
		Target: itg.viper.GetString("businessserver.servicename"),
		EtcdConfig: clientv3.Config{
			Endpoints:   itg.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: itg.viper.GetDuration("etcdCluster.dialtimeout"),
		},
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(itg.viper.GetDuration("businessserver.calltimeout")),
	}

	// build gRPC client of businessserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("create businessserver gRPC client, %+v", err)
	}
	itg.businessSvrConn = conn
	itg.businessSvrCli = pbbusinessserver.NewBusinessClient(conn.Conn())
	logger.Info("create businessserver gRPC client success.")
}

// initializes prometheus metrics collector.
func (itg *Integrator) initMetricsCollector() {
	itg.collector = metrics.NewCollector(itg.viper.GetString("metrics.endpoint"),
		itg.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := itg.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initMods initialize the server modules.
func (itg *Integrator) initMods() {
	// initialize service discovery.
	itg.initServiceDiscovery()

	// initialize businessserver client.
	itg.initBusinessClient()

	// initialize metrics collector.
	itg.initMetricsCollector()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(itg.viper.GetString("server.endpoint.ip"), itg.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("listen on target endpoint, %+v", err)
	}
	itg.lis = lis
}

// Run runs integrator.
func (itg *Integrator) Run() {
	// initialize config.
	itg.initConfig()

	// initialize logger.
	itg.initLogger()
	defer itg.Stop()

	// initialize server modules.
	itg.initMods()

	// register integrator service.
	go func() {
		if err := itg.service.Register(itg.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterIntegratorServer(s, itg)
	logger.Info("Integrator running now.")

	if err := s.Serve(itg.lis); err != nil {
		logger.Fatal("start integrator gRPC service. %+v", err)
	}
}

// Stop stops the Integrator.
func (itg *Integrator) Stop() {
	// close businessserver gRPC connection when server exit.
	if itg.businessSvrConn != nil {
		itg.businessSvrConn.Close()
	}

	// unregister service.
	if itg.service != nil {
		itg.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
