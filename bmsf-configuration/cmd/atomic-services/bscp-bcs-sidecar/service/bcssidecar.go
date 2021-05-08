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

	"bk-bscp/internal/safeviper"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/framework"
	"bk-bscp/pkg/logger"
)

// Sidecar is bscp bcs sidecar.
type Sidecar struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *safeviper.SafeViper

	// reloader.
	reloader *Reloader

	// app mods manager.
	appModMgr *AppModManager

	// instance server.
	insServer *InstanceServer
}

// NewSidecar creates new sidecar instance.
func NewSidecar() *Sidecar {
	return &Sidecar{}
}

// Init initialize the settings.
func (s *Sidecar) Init(setting framework.Setting) {
	s.setting = setting
}

// initialize config and check base content.
func (s *Sidecar) initConfig() {
	cfg := config{}
	viper, err := cfg.init(s.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	s.viper = viper
}

// initialize logger.
func (s *Sidecar) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          s.viper.GetString("logger.directory"),
		LogMaxSize:      s.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       s.viper.GetInt("logger.maxnum"),
		ToStdErr:        s.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    s.viper.GetBool("logger.alsoStderr"),
		Verbosity:       s.viper.GetInt32("logger.level"),
		StdErrThreshold: s.viper.GetString("logger.stderrThreshold"),
		VModule:         s.viper.GetString("logger.vmodule"),
		TraceLocation:   s.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		s.viper.GetString("logger.directory"), s.viper.GetInt32("logger.level"))

	logger.Info("dump configs: sidecar[%+v] gateway[%+v] connserver[%+v] appinfo[%+v] instance[%+v] cache[%+v]",
		s.viper.Get("sidecar"), s.viper.Get("gateway"), s.viper.Get("connserver"), s.viper.Get("appinfo"),
		s.viper.Get("instance"), s.viper.Get("cache"))
}

// init configs reloader.
func (s *Sidecar) initReloader() {
	s.reloader = NewReloader(s.viper)
	s.reloader.Init()
	logger.Info("Sidecar| init reloader success.")
}

// init app mods.
func (s *Sidecar) initAppMods() {
	s.appModMgr = NewAppModManager(s.viper, s.reloader)
	s.appModMgr.Setup()
	logger.Info("Sidecar| init app mod manager setup success.")
}

// init instance server.
func (s *Sidecar) initInstanceServer() {
	if !s.viper.GetBool("instance.open") {
		return
	}

	// create instance server.
	s.insServer = NewInstanceServer(s.viper,
		common.Endpoint(s.viper.GetString("instance.httpEndpoint.ip"), s.viper.GetInt("instance.httpEndpoint.port")),
		common.Endpoint(s.viper.GetString("instance.grpcEndpoint.ip"), s.viper.GetInt("instance.grpcEndpoint.port")),
		s.appModMgr, s.reloader)

	// init and run.
	if err := s.insServer.Init(); err != nil {
		logger.Warn("Sidecar| init instance server, %+v", err)
	} else {
		go s.insServer.Run()
	}
}

// Run runs sidecar.
func (s *Sidecar) Run() {
	// initialize config.
	s.initConfig()

	// initialize logger.
	s.initLogger()
	defer s.Stop()

	// initialize reloader.
	s.initReloader()

	// initialize app mods.
	s.initAppMods()

	// initialize instance server.
	s.initInstanceServer()

	// run success.
	logger.Info("Sidecar running now.")
	select {}
}

// Stop stops the sidecar.
func (s *Sidecar) Stop() {
	// close logger.
	logger.CloseLogs()
}
