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
	"net/http"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-patcher/crons"
	"bk-bscp/cmd/middle-services/bscp-patcher/modules/ctm"
	"bk-bscp/cmd/middle-services/bscp-patcher/modules/hpm"
	"bk-bscp/cmd/middle-services/bscp-patcher/patchs"
	"bk-bscp/internal/dbsharding"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

// Patcher is bscp patcher.
type Patcher struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// patcher discovery instances for master election.
	service *grpclb.Service

	// etcd cluster configs.
	etcdCfg clientv3.Config

	// db sharding manager.
	smgr *dbsharding.ShardingManager

	// patch manager.
	hpm *hpm.PatchManager

	// crontab manager.
	ctm *ctm.CrontabManager
}

// NewPatcher creates new patcher instance.
func NewPatcher() *Patcher {
	return &Patcher{}
}

// Init initialize the settings.
func (p *Patcher) Init(setting framework.Setting) {
	p.setting = setting
}

// initialize config and check base content.
func (p *Patcher) initConfig() {
	cfg := config{}
	viper, err := cfg.init(p.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	p.viper = viper
}

// initialize logger.
func (p *Patcher) initLogger() {
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

	logger.Info("dump configs: server[%+v %+v %+v] database[%+v]",
		p.viper.Get("server.endpoint.ip"), p.viper.Get("server.endpoint.port"),
		p.viper.Get("server"), p.viper.Get("database"))
}

// create new service struct of patcher, and register service later.
func (p *Patcher) initServiceDiscovery() {
	p.service = grpclb.NewService(
		p.viper.GetString("server.serviceName"),
		common.Endpoint(p.viper.GetString("server.endpoint.ip"), p.viper.GetInt("server.endpoint.port")),
		p.viper.GetString("server.metadata"),
		p.viper.GetInt64("server.discoveryTTL"))

	caFile := p.viper.GetString("etcdCluster.tls.caFile")
	certFile := p.viper.GetString("etcdCluster.tls.certFile")
	keyFile := p.viper.GetString("etcdCluster.tls.keyFile")
	certPassword := p.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTLSConfVerify(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		p.etcdCfg = clientv3.Config{
			Endpoints:   p.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: p.viper.GetDuration("etcdCluster.dialTimeout"),
			TLS:         tlsConf,
		}
	} else {
		p.etcdCfg = clientv3.Config{
			Endpoints:   p.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: p.viper.GetDuration("etcdCluster.dialTimeout"),
		}
	}
	logger.Info("init service discovery success.")
}

// create and initialize database sharding manager.
func (p *Patcher) initShardingDB() {
	p.smgr = dbsharding.NewShardingMgr(&dbsharding.Config{
		DBHost:        p.viper.GetString("database.host"),
		DBPort:        p.viper.GetInt("database.port"),
		DBUser:        p.viper.GetString("database.user"),
		DBPasswd:      p.viper.GetString("database.passwd"),
		Size:          p.viper.GetInt("database.cacheSize"),
		PurgeInterval: p.viper.GetDuration("database.purgeInterval"),
	}, &dbsharding.DBConfigTemplate{
		ConnTimeout:  p.viper.GetDuration("database.connTimeout"),
		ReadTimeout:  p.viper.GetDuration("database.readTimeout"),
		WriteTimeout: p.viper.GetDuration("database.writeTimeout"),
		MaxOpenConns: p.viper.GetInt("database.maxOpenConns"),
		MaxIdleConns: p.viper.GetInt("database.maxIdleConns"),
		KeepAlive:    p.viper.GetDuration("database.keepalive"),
	})
	if err := p.smgr.Init(); err != nil {
		logger.Fatal("init db sharding, %+v", err)
	}
	logger.Info("init db sharding success.")
}

func (p *Patcher) initPatchManager() {
	p.hpm = hpm.NewPatchManager(p.viper, p.smgr)
	p.hpm.Load(patchs.Patchs())

	logger.Info("init hpm success.")
}

func (p *Patcher) initCrontab() {
	// create crontab job manager.
	p.ctm = ctm.NewCrontabManager(p.viper, p.smgr, p.service)

	// start ctm.
	if err := p.ctm.Start(crons.CronJobs()); err != nil {
		logger.Warnf("init crontab failed, %+v", err)
	} else {
		logger.Info("init crontab success.")
	}
}

// initMods initializes the patcher modules.
func (p *Patcher) initMods() {
	// initialize service discovery.
	p.initServiceDiscovery()

	// initialize db sharding manager.
	p.initShardingDB()

	// initialize patch manager.
	p.initPatchManager()

	// initialize crontab.
	p.initCrontab()
}

// initialize service.
func (p *Patcher) initService() {
	// http handler.
	httpMux := http.NewServeMux()

	// new router handler.
	rtr := mux.NewRouter()

	// setup routers.
	p.setupRouters(rtr)
	httpMux.Handle("/", rtr)

	// setup filters, all requests would cross in the filter.
	serverMux := p.setupFilters(httpMux)

	// listen and serve without TLS.
	endpoint := common.Endpoint(p.viper.GetString("server.endpoint.ip"), p.viper.GetInt("server.endpoint.port"))
	httpServer := &http.Server{Addr: endpoint, Handler: serverMux}

	logger.Info("Patcher running now.")
	if err := httpServer.ListenAndServe(); err != nil {
		logger.Fatal("http server listen and serve, %+v", err)
	}
}

// Run runs config server.
func (p *Patcher) Run() {
	// initialize config.
	p.initConfig()

	// initialize logger.
	p.initLogger()
	defer p.Stop()

	// initialize server modules.
	p.initMods()

	// register patcher service.
	go func() {
		if err := p.service.Register(p.etcdCfg); err != nil {
			logger.Fatal("register service for discovery, %+v", err)
		}
	}()
	logger.Info("register service for discovery success.")

	// init api http server.
	p.initService()
}

// Stop stops the apiserver.
func (p *Patcher) Stop() {
	// stop crontab jobs.
	// NOTE: maybe hanging here.
	if p.ctm != nil {
		p.ctm.Stop()
	}

	// unregister service.
	if p.service != nil {
		p.service.UnRegister()
	}

	// close logger.
	logger.CloseLogs()
}
