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
	"encoding/json"
	"log"
	"math"
	"net"
	"time"

	"github.com/bluele/gcache"
	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"bk-bscp/cmd/bscp-connserver/modules/metrics"
	"bk-bscp/cmd/bscp-connserver/modules/publish"
	"bk-bscp/cmd/bscp-connserver/modules/resourcesche"
	"bk-bscp/cmd/bscp-connserver/modules/session"
	"bk-bscp/internal/framework"
	"bk-bscp/internal/framework/executor"
	pbbcscontroller "bk-bscp/internal/protocol/bcs-controller"
	pb "bk-bscp/internal/protocol/connserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/natsmq"
	"bk-bscp/pkg/rssche"
)

// ConnServer is bscp connection server.
type ConnServer struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// connection server gRPC service network listener.
	lis net.Listener

	// session manager, handles connection sessions.
	sessionMgr *session.Manager

	// load info reporter for current connserver.
	reporter *rssche.Reporter

	// config publish notification message subscriber.
	subscriber *mq.Subscriber

	// handle all publish events, push notification sidecars.
	publishMgr *publish.Manager

	// strategy handler, check release strategies when publish event coming.
	strategyHandler *strategy.Handler

	// access resources object for schedule.
	accessResource *resourcesche.ConnServerRes

	// access resources scheduler, handle connection servers
	// discovery and schedule endpoints for sidecar access.
	accessScheduler *rssche.Scheduler

	// configs content cache.
	configsCache gcache.Cache

	// bcs controller gRPC connection/client.
	bcsControllerConn *grpclb.GRPCConn
	bcsControllerCli  pbbcscontroller.BCSControllerClient

	// datamanager gRPC connection/client.
	dataMgrConn *grpclb.GRPCConn
	dataMgrCli  pbdatamanager.DataManagerClient

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor
}

// NewConnServer creates new connection server instance.
func NewConnServer() *ConnServer {
	return &ConnServer{}
}

// Init initialize the settings.
func (cs *ConnServer) Init(setting framework.Setting) {
	cs.setting = setting
}

// initialize config and check base content.
func (cs *ConnServer) initConfig() {
	cfg := config{}
	viper, err := cfg.init(cs.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	cs.viper = viper
}

// initialize logger.
func (cs *ConnServer) initLogger() {
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

	logger.Info("dump configs: server[%+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v, %+v] metrics[%+v] bcscontroller[%+v, %+v] datamanager[%+v, %+v] nats[%+v, %+v, %+v, %+v, %+v] etcdCluster[%+v, %+v]",
		cs.viper.Get("server.servicename"), cs.viper.Get("server.endpoint.ip"), cs.viper.Get("server.endpoint.port"), cs.viper.Get("server.discoveryttl"),
		cs.viper.Get("server.schedule.nodesLimit"), cs.viper.Get("server.reportinterval"), cs.viper.Get("server.pubChanTimeout"), cs.viper.Get("server.configsCacheSize"),
		cs.viper.Get("server.keepaliveinterval"), cs.viper.Get("server.keepalivetimeout"), cs.viper.Get("metrics.endpoint"), cs.viper.Get("bcscontroller.servicename"),
		cs.viper.Get("bcscontroller.calltimeout"), cs.viper.Get("datamanager.servicename"), cs.viper.Get("datamanager.calltimeout"), cs.viper.Get("natsmqCluster.endpoints"),
		cs.viper.Get("natsmqCluster.timeout"), cs.viper.Get("natsmqCluster.reconwait"), cs.viper.Get("natsmqCluster.maxrecons"), cs.viper.Get("natsmqCluster.publishtopic"),
		cs.viper.Get("etcdCluster.endpoints"), cs.viper.Get("etcdCluster.dialtimeout"))
}

// create resource reporter instance to keep reporting
// sidecar connection counts as a load-reporter.
func (cs *ConnServer) initReporter() {
	cs.reporter = rssche.NewReporter(cs.viper.GetString("server.servicename"),
		cs.viper.GetInt64("server.discoveryttl"))

	if err := cs.reporter.Init(clientv3.Config{
		Endpoints:   cs.viper.GetStringSlice("etcdCluster.endpoints"),
		DialTimeout: cs.viper.GetDuration("etcdCluster.dialtimeout"),
	}); err != nil {
		logger.Fatal("create new connserver resource reporter, %+v", err)
	}
	logger.Info("create new connserver resource reporter success.")
}

// build connection server access resource scheduler.
func (cs *ConnServer) initAccessScheduler() {
	cs.accessResource = resourcesche.NewConnServerRes(cs.viper)
	cs.accessScheduler = rssche.NewScheduler(cs.viper.GetString("server.servicename"), cs.accessResource)

	if err := cs.accessScheduler.Init(clientv3.Config{
		Endpoints:   cs.viper.GetStringSlice("etcdCluster.endpoints"),
		DialTimeout: cs.viper.GetDuration("etcdCluster.dialtimeout"),
	}); err != nil {
		logger.Fatal("create connserver resource scheduler, %+v", err)
	}

	if err := cs.accessScheduler.Start(); err != nil {
		logger.Fatal("start connserver resource scheduler, %+v", err)
	}
	logger.Info("start connserver resource scheduler success.")
}

// create connection session manager, handle all connections from bcs sidecar.
func (cs *ConnServer) initSessionManager() {
	cs.sessionMgr = session.NewManager(cs.viper)
	logger.Info("init session manager success.")
}

// create strategy handler, it would check release strategies when publish event coming.
func (cs *ConnServer) initStrategyHandler() {
	cs.strategyHandler = strategy.NewHandler(nil)
	logger.Info("init strategy handler success.")
}

// create subscriber used for receiving publish notification from message queue.
func (cs *ConnServer) initSubscriber() {
	cs.subscriber = mq.NewSubscriber(cs.viper.GetStringSlice("natsmqCluster.endpoints"))

	if err := cs.subscriber.Init(
		cs.viper.GetDuration("natsmqCluster.timeout"),
		cs.viper.GetDuration("natsmqCluster.reconwait"),
		cs.viper.GetInt("natsmqCluster.maxrecons"),
	); err != nil {
		logger.Fatal("init publish subscriber, %+v", err)
	}
	logger.Info("init publish subscriber success.")
}

// create publish manager, receive notification from message queue.
func (cs *ConnServer) initPublishManager() {
	cs.publishMgr = publish.NewManager(cs.viper, cs.subscriber, cs.sessionMgr, cs.strategyHandler,
		cs.collector, cs.configsCache, cs.dataMgrCli)

	if err := cs.publishMgr.Init(); err != nil {
		logger.Fatal("init publish manager, %+v", err)
	}
	logger.Info("init publish manager success.")
}

// init local configs content cache.
func (cs *ConnServer) initConfigsCache() {
	cs.configsCache = gcache.New(cs.viper.GetInt("server.configsCacheSize")).EvictType(gcache.TYPE_LRU).Build()
	logger.Info("init local configs content cache success.")
}

// init bcs-controller gRPC connection/client.
func (cs *ConnServer) initBCSControllerClient() {
	ctx := &grpclb.Context{
		Target: cs.viper.GetString("bcscontroller.servicename"),
		EtcdConfig: clientv3.Config{
			Endpoints:   cs.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: cs.viper.GetDuration("etcdCluster.dialtimeout"),
		},
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(cs.viper.GetDuration("bcscontroller.calltimeout")),
	}

	// build gRPC client of bcs controller.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("create bcs controller gRPC client, %+v", err)
	}
	cs.bcsControllerConn = conn
	cs.bcsControllerCli = pbbcscontroller.NewBCSControllerClient(conn.Conn())
	logger.Info("create bcs controller gRPC client success.")
}

// init datamanager server gRPC connection/client.
func (cs *ConnServer) initDataManagerClient() {
	ctx := &grpclb.Context{
		Target: cs.viper.GetString("datamanager.servicename"),
		EtcdConfig: clientv3.Config{
			Endpoints:   cs.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: cs.viper.GetDuration("etcdCluster.dialtimeout"),
		},
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(cs.viper.GetDuration("datamanager.calltimeout")),
	}

	// build gRPC client of datamanager server.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("create datamanager gRPC client, %+v", err)
	}
	cs.dataMgrConn = conn
	cs.dataMgrCli = pbdatamanager.NewDataManagerClient(conn.Conn())
	logger.Info("create datamanager gRPC client success.")
}

// initializes prometheus metrics collector.
func (cs *ConnServer) initMetricsCollector() {
	cs.collector = metrics.NewCollector(cs.viper.GetString("metrics.endpoint"),
		cs.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := cs.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("metrics collector setup success.")
}

// initializes action executor.
func (cs *ConnServer) initExecutor() {
	cs.executor = executor.NewRateLimitExecutor(cs.viper.GetInt("server.executorLimitRate"))
	logger.Info("create action executor success.")
}

// initMods initializes the server modules.
func (cs *ConnServer) initMods() {
	// initialize reporter module.
	cs.initReporter()

	// initialize access scheduler module.
	cs.initAccessScheduler()

	// initialize session manager module.
	cs.initSessionManager()

	// initialize strategy handler.
	cs.initStrategyHandler()

	// initialize publish subscriber.
	cs.initSubscriber()

	// initialize metrics collector.
	cs.initMetricsCollector()

	// initialize configs cache.
	cs.initConfigsCache()

	// create bcs controller gRPC client.
	cs.initBCSControllerClient()

	// create datamanager gRPC client.
	cs.initDataManagerClient()

	// initialize publish manager.
	cs.initPublishManager()

	// initialize action executor.
	cs.initExecutor()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(cs.viper.GetString("server.endpoint.ip"), cs.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("can't listen on local endpoint, %+v", err)
	}
	cs.lis = lis
	logger.Info("listen on local endpoint success.")
}

// report reports connection server load info.
func (cs *ConnServer) report() {
	server := structs.ConnServer{
		IP:   cs.viper.GetString("server.endpoint.ip"),
		Port: cs.viper.GetInt("server.endpoint.port"),
	}

	// get connection counts hold in local connserver.
	ticker := time.NewTicker(cs.viper.GetDuration("server.reportinterval"))
	defer ticker.Stop()

	for {
		<-ticker.C

		// get connection counts hold in local connserver.
		count, err := cs.sessionMgr.ConnCount()
		if err != nil {
			logger.Error("update connserver access resource information, get local conn-count, %+v", err)
			continue
		}
		server.ConnCount = count

		bytes, err := json.Marshal(&server)
		if err != nil {
			logger.Error("update connserver access resource information, can't marshal %+v", err)
			continue
		}
		if err := cs.reporter.UpdateRes(structs.Metadata{Metadata: string(bytes)}); err != nil {
			logger.Error("update connserver access resource information, %+v", err)
			continue
		}
		logger.Info("update connserver access resource information success, %+v", server)
	}
}

// collect collects local normal metrics data.
func (cs *ConnServer) collect() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// session count.
		if count, err := cs.sessionMgr.ConnCount(); err == nil {
			cs.collector.StatConnNum(count)
		}

		// access node count.
		cs.collector.StatAccessNodeNum(cs.accessResource.NodeCount())
	}
}

// Run runs connection server
func (cs *ConnServer) Run() {
	// initialize config.
	cs.initConfig()

	// initialize logger.
	cs.initLogger()
	defer cs.Stop()

	// initialize server modules.
	cs.initMods()

	// add connserver node resource object.
	if err := cs.reporter.AddRes(structs.Metadata{}); err != nil {
		logger.Fatal("add connserver access resource, %+v", err)
	}
	logger.Info("add connserver access resource success.")

	// report connserver access resource information.
	go cs.report()

	// collect normal metrics data.
	go cs.collect()

	// run service.
	s := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		Time:    cs.viper.GetDuration("server.keepaliveinterval"),
		Timeout: cs.viper.GetDuration("server.keepalivetimeout"),
	}), grpc.MaxRecvMsgSize(math.MaxInt32))

	pb.RegisterConnectionServer(s, cs)
	logger.Info("Connection Server running now.")

	if err := s.Serve(cs.lis); err != nil {
		logger.Fatal("start connserver gRPC service, %+v", err)
	}
}

// Stop stops the connserver.
func (cs *ConnServer) Stop() {
	// delete resource.
	if cs.reporter != nil {
		cs.reporter.DeleteRes()
	}

	if cs.accessScheduler != nil {
		cs.accessScheduler.Stop()
	}

	// unsubscribe release topic and close subscriber.
	if cs.subscriber != nil {
		cs.subscriber.UnSubscribe()
		cs.subscriber.Close()
	}

	// close bcs controller gRPC connection.
	if cs.bcsControllerConn != nil {
		cs.bcsControllerConn.Close()
	}

	// close datamanager gRPC connection.
	if cs.dataMgrConn != nil {
		cs.dataMgrConn.Close()
	}

	// close logger.
	logger.CloseLogs()
}
