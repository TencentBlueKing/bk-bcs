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
	"github.com/bluele/gcache"
	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"bk-bscp/cmd/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/dbsharding"
	"bk-bscp/internal/framework"
	"bk-bscp/internal/framework/executor"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
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

	// db sharding manager.
	smgr *dbsharding.ShardingManager

	// commit cache.
	commitCache gcache.Cache

	// release cache.
	releaseCache gcache.Cache

	// configs cache.
	configsCache gcache.Cache

	// prometheus metrics collector.
	collector *metrics.Collector

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

	logger.Info("dump configs: server[%+v, %+v, %+v, %+v, %+v, %+v] metrics[%+v] etcdCluster[%+v, %+v] database[%+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v]",
		dm.viper.Get("server.servicename"), dm.viper.Get("server.endpoint.ip"), dm.viper.Get("server.endpoint.port"), dm.viper.Get("server.discoveryttl"),
		dm.viper.Get("server.releaseCacheSize"), dm.viper.Get("server.configsCacheSize"), dm.viper.Get("metrics.endpoint"), dm.viper.Get("etcdCluster.endpoints"),
		dm.viper.Get("etcdCluster.dialtimeout"), dm.viper.Get("database.dialect"), dm.viper.Get("database.host"), dm.viper.Get("database.port"), dm.viper.Get("database.user"),
		dm.viper.Get("database.conntimeout"), dm.viper.Get("database.readtimeout"), dm.viper.Get("database.writetimeout"), dm.viper.Get("database.maxopenconns"),
		dm.viper.Get("database.maxidleconns"), dm.viper.Get("database.keepalive"), dm.viper.Get("database.cachesize"), dm.viper.Get("database.purgeInterval"))
}

// create new service struct of datamanager, and register service later.
func (dm *DataManager) initServiceDiscovery() {
	dm.service = grpclb.NewService(dm.viper.GetString("server.servicename"),
		common.Endpoint(dm.viper.GetString("server.endpoint.ip"), dm.viper.GetInt("server.endpoint.port")),
		dm.viper.GetString("server.metadata"),
		dm.viper.GetInt64("server.discoveryttl"))

	caFile := dm.viper.GetString("etcdCluster.tls.cafile")
	certFile := dm.viper.GetString("etcdCluster.tls.certfile")
	keyFile := dm.viper.GetString("etcdCluster.tls.keyfile")
	certPassword := dm.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTslConfVerity(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		dm.etcdCfg = clientv3.Config{
			Endpoints:   dm.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: dm.viper.GetDuration("etcdCluster.dialtimeout"),
			TLS:         tlsConf,
		}
	} else {
		dm.etcdCfg = clientv3.Config{
			Endpoints:   dm.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: dm.viper.GetDuration("etcdCluster.dialtimeout"),
		}
	}
	logger.Info("create service for discovery success.")
}

// create and initialize database sharding manager.
func (dm *DataManager) initShardingDB() {
	dm.smgr = dbsharding.NewShardingMgr(&dbsharding.Config{
		Dialect:       dm.viper.GetString("database.dialect"),
		DBHost:        dm.viper.GetString("database.host"),
		DBPort:        dm.viper.GetInt("database.port"),
		DBUser:        dm.viper.GetString("database.user"),
		DBPasswd:      dm.viper.GetString("database.passwd"),
		Size:          dm.viper.GetInt("database.cachesize"),
		PurgeInterval: dm.viper.GetDuration("database.purgeInterval"),
	}, &dbsharding.DBConfigTemplate{
		ConnTimeout:  dm.viper.GetDuration("database.conntimeout"),
		ReadTimeout:  dm.viper.GetDuration("database.readtimeout"),
		WriteTimeout: dm.viper.GetDuration("database.writetimeout"),
		MaxOpenConns: dm.viper.GetInt("database.maxopenconns"),
		MaxIdleConns: dm.viper.GetInt("database.maxidleconns"),
		KeepAlive:    dm.viper.GetDuration("database.keepalive"),
	})

	if err := dm.smgr.Init(); err != nil {
		logger.Fatal("init db sharding, %+v", err)
	}
	logger.Info("init db sharding success.")
}

// init commit cache.
func (dm *DataManager) initCommitCache() {
	dm.commitCache = gcache.New(dm.viper.GetInt("server.commitCacheSize")).EvictType(gcache.TYPE_LRU).Build()
	logger.Info("init local commit cache success.")
}

// init release cache.
func (dm *DataManager) initReleaseCache() {
	dm.releaseCache = gcache.New(dm.viper.GetInt("server.releaseCacheSize")).EvictType(gcache.TYPE_LRU).Build()
	logger.Info("init local release cache success.")
}

// init configs cache.
func (dm *DataManager) initConfigsCache() {
	dm.configsCache = gcache.New(dm.viper.GetInt("server.configsCacheSize")).EvictType(gcache.TYPE_LRU).Build()
	logger.Info("init local configs cache success.")
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

	// initialize db sharding manager.
	dm.initShardingDB()

	// initialize commit cache.
	dm.initCommitCache()

	// initialize release cache.
	dm.initReleaseCache()

	// initialize configs cache.
	dm.initConfigsCache()

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
