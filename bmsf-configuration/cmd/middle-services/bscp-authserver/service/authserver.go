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

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth/bkiam"
	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth/local"
	"bk-bscp/cmd/middle-services/bscp-authserver/modules/metrics"
	pb "bk-bscp/internal/protocol/authserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/framework/executor"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// AuthServer is bscp auth server.
type AuthServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// authserver discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// prometheus metrics collector.
	collector *metrics.Collector

	// auth mode.
	authMode string

	// local casbin auth controller.
	localAuthController *local.Controller

	// bkiam auth controller.
	bkiamAuthController *bkiam.Controller

	// action executor.
	executor *executor.Executor
}

// NewAuthServer creates new auth server instance.
func NewAuthServer() *AuthServer {
	return &AuthServer{}
}

// Init initialize the settings.
func (as *AuthServer) Init(setting framework.Setting) {
	as.setting = setting
}

// initialize config and check base content.
func (as *AuthServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(as.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	as.viper = viper
}

// initialize logger.
func (as *AuthServer) initLogger() {
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

	logger.Info("dump configs: server[%+v %+v] metrics[%+v] etcdCluster[%+v] auth[%+v] bkiam[%+v] database[%+v]",
		as.viper.Get("server.endpoint.ip"), as.viper.Get("server.endpoint.port"), as.viper.Get("metrics"),
		as.viper.Get("etcdCluster"), as.viper.Get("auth"), as.viper.Get("bkiam"), as.viper.Get("database"))
}

// create new service struct of authserver, and register service later.
func (as *AuthServer) initServiceDiscovery() {
	as.service = grpclb.NewService(as.viper.GetString("server.serviceName"),
		common.Endpoint(as.viper.GetString("server.endpoint.ip"), as.viper.GetInt("server.endpoint.port")),
		as.viper.GetString("server.metadata"),
		as.viper.GetInt64("server.discoveryTTL"))

	caFile := as.viper.GetString("etcdCluster.tls.caFile")
	certFile := as.viper.GetString("etcdCluster.tls.certFile")
	keyFile := as.viper.GetString("etcdCluster.tls.keyFile")
	certPassword := as.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTLSConfVerify(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		as.etcdCfg = clientv3.Config{
			Endpoints:   as.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: as.viper.GetDuration("etcdCluster.dialTimeout"),
			TLS:         tlsConf,
		}
	} else {
		as.etcdCfg = clientv3.Config{
			Endpoints:   as.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: as.viper.GetDuration("etcdCluster.dialTimeout"),
		}
	}
	logger.Info("create service for discovery success.")
}

// initializes prometheus metrics collector.
func (as *AuthServer) initMetricsCollector() {
	as.collector = metrics.NewCollector(as.viper.GetString("metrics.endpoint"), as.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := as.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initializes auth controllers.
func (as *AuthServer) initAuthController() {
	as.authMode = as.viper.GetString("auth.mode")

	if as.authMode == auth.AuthModeLocal {
		controller, err := local.NewController(as.viper)
		if err != nil {
			logger.Fatal("init local auth controller, %+v", err)
		}
		as.localAuthController = controller
	} else {
		controller, err := bkiam.NewController(as.viper)
		if err != nil {
			logger.Fatal("init bkiam auth controller, %+v", err)
		}
		as.bkiamAuthController = controller
	}
	logger.Infof("auth controller[%s] init success.", as.authMode)
}

// initializes action executor.
func (as *AuthServer) initExecutor() {
	as.executor = executor.NewExecutor()
	logger.Info("create action executor success.")
}

// initMods initialize the server modules.
func (as *AuthServer) initMods() {
	// initialize service discovery.
	as.initServiceDiscovery()

	// initialize metrics collector.
	as.initMetricsCollector()

	// initialize auth controller.
	as.initAuthController()

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

// Run runs auth server.
func (as *AuthServer) Run() {
	// initialize config.
	as.initConfig()

	// initialize logger.
	as.initLogger()
	defer as.Stop()

	// initialize server modules.
	as.initMods()

	// register datamanger service.
	go func() {
		if err := as.service.Register(as.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterAuthServer(s, as)
	logger.Info("Auth server running now.")

	if err := s.Serve(as.lis); err != nil {
		logger.Fatal("start auth server gRPC service. %+v", err)
	}
}

// Stop stops the authserver.
func (as *AuthServer) Stop() {
	// unregister service.
	if as.service != nil {
		as.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
