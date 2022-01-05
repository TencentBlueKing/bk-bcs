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
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/gseagent"
	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/metrics"
	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/publish"
	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/session"
	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/tunnel"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/framework/executor"
	"bk-bscp/pkg/logger"
)

// GSEPlugin is bscp gse plugin.
type GSEPlugin struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *safeviper.SafeViper

	// plugin gRPC service network listener.
	lis net.Listener

	// session manager, handles sidecar connection sessions.
	sessionMgr *session.Manager

	// handle all publish events, push notification sidecars.
	publishMgr *publish.Manager

	// strategy handler, check release strategies when publish event coming.
	strategyHandler *strategy.Handler

	// gse agent tunnel client.
	gseTunnel *tunnel.Tunnel

	// gse agent base wrapper.
	gseAgent *gseagent.GSEAgent

	// prometheus metrics collector.
	collector *metrics.Collector

	// action executor.
	executor *executor.Executor

	// local siedecar instance.
	sidecar *Sidecar

	// running signal chan.
	running chan struct{}
}

// NewGSEPlugin creates new gse plugin instance.
func NewGSEPlugin() *GSEPlugin {
	return &GSEPlugin{running: make(chan struct{})}
}

// Init initialize the settings.
func (p *GSEPlugin) Init(setting framework.Setting) {
	p.setting = setting
}

// initialize config and check base content.
func (p *GSEPlugin) initConfig() {
	cfg := config{}
	viper, err := cfg.init(p.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	p.viper = viper
}

// initialize logger.
func (p *GSEPlugin) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          p.viper.GetString("logger.directory"),
		LogMaxSize:      p.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       p.viper.GetInt("logger.maxnum"),
		ToStdErr:        p.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    p.viper.GetBool("logger.alsoStderr"),
		Verbosity:       p.viper.GetInt32("logger.level"),
		StdErrThreshold: p.viper.GetString("logger.stderrThreshold"),
		VModule:         p.viper.GetString("logger.vmodule"),
		TraceLocation:   p.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		p.viper.GetString("logger.directory"), p.viper.GetInt32("logger.level"))

	logger.Info("dump configs: server[%+v] metrics[%+v] sidecar[%+v] connserver[%+v] "+
		"appinfo[%+v] instance[%+v] cache[%+v]",
		p.viper.Get("server"), p.viper.Get("metrics"), p.viper.Get("sidecar"), p.viper.Get("connserver"),
		p.viper.Get("appinfo"), p.viper.Get("instance"), p.viper.Get("cache"))
}

// create connection session manager, handle all connections from bcs sidecar.
func (p *GSEPlugin) initSessionManager() {
	p.sessionMgr = session.NewManager(p.viper)
	logger.Info("GSEPlugin| init session manager success.")
}

// create strategy handler, it would check release strategies when publish event coming.
func (p *GSEPlugin) initStrategyHandler() {
	p.strategyHandler = strategy.NewHandler(nil)
	logger.Info("GSEPlugin| init strategy handler success.")
}

// init gse agent handler wrapper.
func (p *GSEPlugin) initGSEAgent() {
	p.gseAgent = gseagent.NewGSEAgent(p.viper)

	p.gseTunnel = tunnel.NewTunnel(p.viper, p.gseAgent, p.publishMgr.Process)
	logger.Info("GSEPlugin| init gse tunnel success.")

	if err := p.gseAgent.Run(p.gseTunnel.RecvMessage); err != nil {
		logger.Fatal("init gse agent, %+v", err)
	}
	logger.Info("GSEPlugin| init gse agent success.")
}

// create publish manager, receive notification from message queue.
func (p *GSEPlugin) initPublishManager() {
	p.publishMgr = publish.NewManager(p.viper, p.sessionMgr, p.strategyHandler)
	logger.Info("GSEPlugin| init publish manager success.")
}

// initializes prometheus metrics collector.
func (p *GSEPlugin) initMetricsCollector() {
	p.collector = metrics.NewCollector(p.viper.GetString("metrics.endpoint"), p.viper.GetString("metrics.path"))

	// setup metrics collector.
	go func() {
		if err := p.collector.Setup(); err != nil {
			logger.Error("metrics collector setup/runtime, %+v", err)
		}
	}()
	logger.Info("GSEPlugin| metrics collector setup success.")
}

// initializes action executor.
func (p *GSEPlugin) initExecutor() {
	p.executor = executor.NewExecutor()
	logger.Info("GSEPlugin| create action executor success.")
}

// initializes local sidecar.
func (p *GSEPlugin) initSidecar() {
	p.sidecar = NewSidecar(p.viper, p.gseTunnel)
	go p.sidecar.Run()
	logger.Info("GSEPlugin| create local sidecar instance success.")
}

// initMods initializes the server modules.
func (p *GSEPlugin) initMods() {
	// initialize session manager module.
	p.initSessionManager()

	// initialize strategy handler.
	p.initStrategyHandler()

	// initialize metrics collector.
	p.initMetricsCollector()

	// initialize publish manager.
	p.initPublishManager()

	// initialize gse agent.
	p.initGSEAgent()

	// initialize action executor.
	p.initExecutor()

	// initialize local sidecar.
	p.initSidecar()

	// listen announces on the local network address, setup rpc server later.
	lis, err := net.Listen("tcp",
		common.Endpoint(p.viper.GetString("server.endpoint.ip"), p.viper.GetInt("server.endpoint.port")))
	if err != nil {
		logger.Fatal("can't listen on local endpoint, %+v", err)
	}
	p.lis = lis
	logger.Info("GSEPlugin| listen on local endpoint success.")
}

// collect collects local normal metrics data.
func (p *GSEPlugin) collect() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// session count.
		if count, err := p.sessionMgr.ConnCount(); err == nil {
			p.collector.StatConnNum(count)
		}
	}
}

// GSEAgent returns gse agent wrapper object.
func (p *GSEPlugin) GSEAgent() *gseagent.GSEAgent {
	return p.gseAgent
}

func (p *GSEPlugin) realRun() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	// initialize config.
	p.initConfig()

	// initialize logger.
	p.initLogger()

	// initialize server modules.
	p.initMods()

	// collect normal metrics data.
	go p.collect()

	// run service.
	s := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		Time:    p.viper.GetDuration("server.keepaliveInterval"),
		Timeout: p.viper.GetDuration("server.keepaliveTimeout"),
	}), grpc.MaxRecvMsgSize(math.MaxInt32))

	pb.RegisterConnectionServer(s, p)
	logger.Info("GSE Plugin running now.")

	p.running <- struct{}{}

	if err := s.Serve(p.lis); err != nil {
		logger.Fatal("start gse plugin connection gRPC service, %+v", err)
	}
}

// Run runs gse plugin.
func (p *GSEPlugin) Run() {
	go p.realRun()
	<-p.running
}

// Stop stops gse plugin.
func (p *GSEPlugin) Stop() {
	// close gse agent tunnel.
	p.gseTunnel.Close()

	// stop gse agent.
	p.gseAgent.Stop()

	// close logger.
	logger.CloseLogs()

	// print stack.
	debug.PrintStack()
}
