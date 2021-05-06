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

	"bk-bscp/cmd/atomic-services/bscp-datamanager/modules/metrics"
	"bk-bscp/cmd/atomic-services/bscp-datamanager/modules/statistics"
	"bk-bscp/internal/dbsharding"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/framework/executor"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// DataManager is bscp datamanager server.
type DataManager struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// datamanager discovery instances.
	service *grpclb.Service

	// network listener.
	lis net.Listener

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// authserver gRPC connection/client.
	authSvrConn *grpclb.GRPCConn
	authSvrCli  pbauthserver.AuthClient

	// db sharding manager.
	smgr *dbsharding.ShardingManager

	// prometheus metrics collector.
	collector *metrics.Collector

	// internal business statistics.
	statistics *statistics.Collector

	// action executor.
	executor *executor.Executor
}

// NewDataManager creates new datamanager server instance.
func NewDataManager() *DataManager {
	return &DataManager{}
}

// Init initialize the settings.
func (dm *DataManager) Init(setting framework.Setting) {
	dm.setting = setting
}

// initialize config and check base content.
func (dm *DataManager) initConfig() {
	cfg := config{}
	viper, err := cfg.init(dm.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	dm.viper = viper
}

// initialize logger.
func (dm *DataManager) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          dm.viper.GetString("logger.directory"),
		LogMaxSize:      dm.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       dm.viper.GetInt("logger.maxnum"),
		ToStdErr:        dm.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    dm.viper.GetBool("logger.alsoStderr"),
		Verbosity:       dm.viper.GetInt32("logger.level"),
		StdErrThreshold: dm.viper.GetString("logger.stderrThreshold"),
		VModule:         dm.viper.GetString("logger.vmodule"),
		TraceLocation:   dm.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		dm.viper.GetString("logger.directory"), dm.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v %+v] metrics[%+v] etcdCluster[%+v] database[%+v]",
		dm.viper.Get("server.endpoint.ip"), dm.viper.Get("server.endpoint.port"), dm.viper.Get("metrics"),
		dm.viper.Get("etcdCluster"), dm.viper.Get("database"))
}

// create new service struct of datamanager, and register service later.
func (dm *DataManager) initServiceDiscovery() {
	dm.service = grpclb.NewService(dm.viper.GetString("server.serviceName"),
		common.Endpoint(dm.viper.GetString("server.endpoint.ip"), dm.viper.GetInt("server.endpoint.port")),
		dm.viper.GetString("server.metadata"),
		dm.viper.GetInt64("server.discoveryTTL"))

	caFile := dm.viper.GetString("etcdCluster.tls.caFile")
	certFile := dm.viper.GetString("etcdCluster.tls.certFile")
	keyFile := dm.viper.GetString("etcdCluster.tls.keyFile")
	certPassword := dm.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTLSConfVerify(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		dm.etcdCfg = clientv3.Config{
			Endpoints:   dm.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: dm.viper.GetDuration("etcdCluster.dialTimeout"),
			TLS:         tlsConf,
		}
	} else {
		dm.etcdCfg = clientv3.Config{
			Endpoints:   dm.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: dm.viper.GetDuration("etcdCluster.dialTimeout"),
		}
	}
	logger.Info("create service for discovery success.")
}

// create auth server gRPC client.
func (dm *DataManager) initAuthServerClient() {
	ctx := &grpclb.Context{
		Target:     dm.viper.GetString("authserver.serviceName"),
		EtcdConfig: dm.etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(dm.viper.GetDuration("authserver.callTimeout")),
	}

	// build gRPC client of authserver.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create authserver gRPC client, %+v", err)
	}
	dm.authSvrConn = conn
	dm.authSvrCli = pbauthserver.NewAuthClient(conn.Conn())
	logger.Info("create authserver gRPC client success.")
}

// create and initialize database sharding manager.
func (dm *DataManager) initShardingDB() {
	dm.smgr = dbsharding.NewShardingMgr(&dbsharding.Config{
		DBHost:        dm.viper.GetString("database.host"),
		DBPort:        dm.viper.GetInt("database.port"),
		DBUser:        dm.viper.GetString("database.user"),
		DBPasswd:      dm.viper.GetString("database.passwd"),
		Size:          dm.viper.GetInt("database.cacheSize"),
		PurgeInterval: dm.viper.GetDuration("database.purgeInterval"),
	}, &dbsharding.DBConfigTemplate{
		ConnTimeout:  dm.viper.GetDuration("database.connTimeout"),
		ReadTimeout:  dm.viper.GetDuration("database.readTimeout"),
		WriteTimeout: dm.viper.GetDuration("database.writeTimeout"),
		MaxOpenConns: dm.viper.GetInt("database.maxOpenConns"),
		MaxIdleConns: dm.viper.GetInt("database.maxIdleConns"),
		KeepAlive:    dm.viper.GetDuration("database.keepalive"),
	})

	if err := dm.smgr.Init(); err != nil {
		logger.Fatal("init db sharding, %+v", err)
	}
	logger.Info("init db sharding success.")
}

// initializes prometheus metrics collector.
func (dm *DataManager) initMetricsCollector() {
	dm.collector = metrics.NewCollector(dm.viper.GetString("metrics.endpoint"),
		dm.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := dm.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")

	dm.statistics = statistics.NewCollector(dm.viper, dm.collector, dm.smgr)
	dm.statistics.Start()
	logger.Info("internal statistics setup success.")
}

// initializes action executor.
func (dm *DataManager) initExecutor() {
	dm.executor = executor.NewExecutor()
	logger.Info("create action executor success.")
}

// initMods initialize the server modules.
func (dm *DataManager) initMods() {
	// initialize service discovery.
	dm.initServiceDiscovery()

	// initialize auth server gRPC client.
	dm.initAuthServerClient()

	// initialize db sharding manager.
	dm.initShardingDB()

	// initialize metrics collector.
	dm.initMetricsCollector()

	// initialize action executor.
	dm.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(dm.viper.GetString("server.endpoint.ip"), dm.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("listen on target endpoint, %+v", err)
	}
	dm.lis = lis
}

// Run runs datamanager server.
func (dm *DataManager) Run() {
	// initialize config.
	dm.initConfig()

	// initialize logger.
	dm.initLogger()
	defer dm.Stop()

	// initialize server modules.
	dm.initMods()

	// register datamanger service.
	go func() {
		if err := dm.service.Register(dm.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// run service.
	s := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pb.RegisterDataManagerServer(s, dm)
	logger.Info("Data Manager running now.")

	if err := s.Serve(dm.lis); err != nil {
		logger.Fatal("start datamanager gRPC service. %+v", err)
	}
}

// Stop stops the datamanager.
func (dm *DataManager) Stop() {
	// stop internal business statistics.
	dm.statistics.Stop()

	// close authserver gRPC connection when server exit.
	if dm.authSvrConn != nil {
		dm.authSvrConn.Close()
	}

	// unregister service.
	if dm.service != nil {
		dm.service.UnRegister()
	}

	// close sharding manager.
	if dm.smgr != nil {
		if err := dm.smgr.Close(); err != nil {
			logger.Info("Close sharding manager, %+v", err)
		}
	}

	// close logger.
	logger.CloseLogs()
}
