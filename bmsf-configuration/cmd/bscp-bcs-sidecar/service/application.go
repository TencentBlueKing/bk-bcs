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
	"fmt"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

const (
	// defaultCluster default cluster name.
	defaultCluster = "default"

	// defaultZone default zone name.
	defaultZone = "default"

	// defaultDC default dc name.
	defaultDC = "default"

	// defaultInitFailedModsWaitTime is duration for wait to init failed mods.
	defaultInitFailedModsWaitTime = time.Second

	// defaultInitInnerModsWaitTime is duration for wait to init inner mods.
	defaultInitInnerModsWaitTime = time.Second
)

// AppModInfo is multi app mode information.
type AppModInfo struct {
	// BusinessName business name.
	BusinessName string `json:"business"`

	// AppName app name.
	AppName string `json:"app"`

	// ClusterName cluster name.
	ClusterName string `json:"cluster"`

	// ZoneName zone name.
	ZoneName string `json:"zone"`

	// DC datacenter tag.
	DC string `json:"dc"`

	// Labels sidecar instance KV labels.
	Labels map[string]string `json:"labels"`

	// Path is sidecar mod app configs effect path.
	Path string `json:"path"`
}

// AppModManager is app mod manager.
type AppModManager struct {
	// configs handler.
	viper *viper.Viper

	// configs reloader.
	reloader *Reloader

	// app mod information.
	appMods []*AppModInfo

	// init failed app mods async queue.
	failedMods []*AppModInfo
}

// NewAppModManager creates a new AppModManager.
func NewAppModManager(viper *viper.Viper, reloader *Reloader) *AppModManager {
	return &AppModManager{viper: viper, reloader: reloader, appMods: []*AppModInfo{}, failedMods: []*AppModInfo{}}
}

// Init inits base app mod infos.
func (mgr *AppModManager) Init() {
	appmodCfgVal := mgr.viper.GetString("appmod")
	if len(appmodCfgVal) == 0 {
		logger.Fatal("can't init appinfo, empty appmod configs here")
	}

	if err := json.Unmarshal([]byte(appmodCfgVal), &mgr.appMods); err != nil {
		logger.Fatal("can't init appinfo, unmarshal %+v", err)
	}
	logger.Info("AppModManager| init and parse appmod count[%d] success.", len(mgr.appMods))

	for _, mod := range mgr.appMods {
		logger.Info("AppModManager| init base infos for mod[%+v]", mod)

		if len(mod.BusinessName) == 0 || len(mod.AppName) == 0 || len(mod.Path) == 0 {
			logger.Fatal("can't init the mod, appinfo missing, %+v", mod)
		}

		// set configs cache path.
		mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.path", mod.BusinessName, mod.AppName), mod.Path)

		// set app mod cluster name.
		if len(mod.ClusterName) == 0 {
			mod.ClusterName = defaultCluster
		}
		mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.cluster", mod.BusinessName, mod.AppName), mod.ClusterName)

		// set app mod zone name.
		if len(mod.ZoneName) == 0 {
			mod.ZoneName = defaultZone
		}
		mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.zone", mod.BusinessName, mod.AppName), mod.ZoneName)

		// set dc tag, maybe empty.
		if len(mod.DC) == 0 {
			mod.DC = defaultDC
		}
		mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.dc", mod.BusinessName, mod.AppName), mod.DC)

		// set labels, maybe empty.
		mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.labels", mod.BusinessName, mod.AppName), mod.Labels)
	}
}

// initialize single appinfo.
func (mgr *AppModManager) initSingleAppinfo(businessName, appName, clusterName, zoneName string) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(mgr.viper.GetDuration("connserver.dialtimeout")),
	}
	endpoint := mgr.viper.GetString("connserver.hostname") + ":" + mgr.viper.GetString("connserver.port")

	// connect to connserver.
	c, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return fmt.Errorf("dial connserver, %+v", err)
	}
	defer c.Close()
	client := pb.NewConnectionClient(c)

	// query app metadata now.
	r := &pb.QueryAppMetadataReq{
		Seq:          common.Sequence(),
		BusinessName: businessName,
		AppName:      appName,
		ClusterName:  clusterName,
		ZoneName:     zoneName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), mgr.viper.GetDuration("connserver.calltimeout"))
	defer cancel()

	logger.Info("AppModManager| query app metadata, %+v", r)

	resp, err := client.QueryAppMetadata(ctx, r)
	if err != nil {
		return fmt.Errorf("query metadata, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return fmt.Errorf("can't query app metadata, %+v", resp)
	}

	// query success, save metadata to local.
	mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.bid", businessName, appName), resp.Bid)
	mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.appid", businessName, appName), resp.Appid)
	mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.clusterid", businessName, appName), resp.Clusterid)
	mgr.viper.Set(fmt.Sprintf("appmod.%s_%s.zoneid", businessName, appName), resp.Zoneid)

	logger.Info("AppModManager| init appinfo success[%s %s %s %s], bid[%+v] appid[%+v] clusterid[%+v] zoneid[%+v]",
		businessName, appName, clusterName, zoneName, resp.Bid, resp.Appid, resp.Clusterid, resp.Zoneid)

	return nil
}

// initialize config cache.
func (mgr *AppModManager) initCache(businessName, appName string) (*EffectCache, *ContentCache) {
	// config release effect cache, "filepath(etc)/businessName/appName".
	effectCache := NewEffectCache(fmt.Sprintf("%s/%s/%s",
		mgr.viper.GetString("cache.fileCachePath"), businessName, appName), businessName, appName)

	// content cache.
	contentCache := NewContentCache(
		fmt.Sprintf("%s/%s/%s", mgr.viper.GetString("cache.contentCachePath"), businessName, appName),
		businessName,
		appName,
		mgr.viper.GetInt("cache.contentMCacheSize"),
		mgr.viper.GetString("cache.contentExpiredCachePath"),
		mgr.viper.GetDuration("cache.mcacheExpiration"),
		mgr.viper.GetDuration("cache.contentCacheExpiration"),
		mgr.viper.GetDuration("cache.contentCachePurgeInterval"),
	)
	go contentCache.Setup()

	logger.Info("AppModManager| init single[%s %s] effect/content cache success.", businessName, appName)
	return effectCache, contentCache
}

// initialize sidecar config handler.
func (mgr *AppModManager) initHandler(businessName, appName string, effectCache *EffectCache, contentCache *ContentCache) *Handler {
	// config handler.
	configHandler := NewConfigHandler(mgr.viper, businessName, appName, effectCache, contentCache, mgr.reloader)

	// handler.
	handler := NewHandler(mgr.viper, businessName, appName, configHandler)

	logger.Info("AppModManager| init single[%s %s] sidecar handler success.", businessName, appName)
	return handler
}

// initInnerMods initialize the inner modules for target app.
func (mgr *AppModManager) initInnerMods(businessName, appName string) {
	// wait for ready to pull configs.
	for {
		// main flag to control all mods.
		if mgr.viper.GetBool("sidecar.readyPullConfigs") {
			break
		}

		// app mod level flags.
		if mgr.viper.GetBool(fmt.Sprintf("sidecar.%s_%s.readyPullConfigs", businessName, appName)) {
			break
		}
		logger.Warn("AppModManager| [%s %s] waiting for pulling configs until flag mark ready", businessName, appName)

		time.Sleep(defaultInitInnerModsWaitTime)
	}
	logger.Info("AppModManager| init inner mods for [%s %s] right now!", businessName, appName)

	// initialize cache.
	effectCache, contentCache := mgr.initCache(businessName, appName)

	// initialize handler.
	handler := mgr.initHandler(businessName, appName, effectCache, contentCache)

	// initialize signalling channel.
	signalling := NewSignallingChannel(mgr.viper, businessName, appName, handler)

	// run sidecar config handler.
	handler.Run()
	logger.Info("AppModManager| sidecar config handler[%s %s] run success.", businessName, appName)

	// setup signalling channel.
	go signalling.Setup()
	logger.Info("AppModManager| sidecar signallingChannel[%s %s] run success.", businessName, appName)
}

// initialize failed mods.
func (mgr *AppModManager) initFailedMods() {
	for {
		time.Sleep(defaultInitFailedModsWaitTime)

		failedAgainMods := []*AppModInfo{}

		// init failed mods now.
		for _, mod := range mgr.failedMods {
			if err := mgr.initSingleAppinfo(mod.BusinessName, mod.AppName, mod.ClusterName, mod.ZoneName); err != nil {
				logger.Warn("AppModManager| init failed app mod[%+v], %+v", mod, err)
				failedAgainMods = append(failedAgainMods, mod)
				continue
			}

			// init failed mod success.
			logger.Info("AppModManager| init failed app mod[%+v] success!", mod)

			// init inner mods in another coroutine.
			go mgr.initInnerMods(mod.BusinessName, mod.AppName)
		}
		mgr.failedMods = failedAgainMods

		if len(failedAgainMods) == 0 {
			// all failed mods init success.
			logger.Info("AppModManager| init all failed app mods success!")
			return
		}
	}
}

// AppModInfos returns app mod infos.
func (mgr *AppModManager) AppModInfos() []*AppModInfo {
	return mgr.appMods
}

// Setup setups all app mods.
func (mgr *AppModManager) Setup() {
	for _, mod := range mgr.appMods {
		// connserver retry times.
		retryTimes := mgr.viper.GetInt("connserver.retry")
		if retryTimes < 0 {
			retryTimes = 0
		}

		// init current app mod success flag.
		isInitSucc := false

		// call connserver and init app mod info with retry action.
		for i := 0; i <= retryTimes; i++ {
			if err := mgr.initSingleAppinfo(mod.BusinessName, mod.AppName, mod.ClusterName, mod.ZoneName); err != nil {
				logger.Warn("init single app mod[%+v] failed[%d], %+v", mod, i, err)

				// retry.
				continue
			}

			// init inner modules for app.
			logger.Info("init app mod[%+v][%d] success!", mod, i)

			// init inner mods in another coroutine.
			go mgr.initInnerMods(mod.BusinessName, mod.AppName)

			isInitSucc = true
			break
		}

		if !isInitSucc {
			logger.Warn("finally[%+v %+v][%d], add async wait queue, can't init mod, %+v",
				mgr.viper.GetDuration("connserver.dialtimeout"), mgr.viper.GetDuration("connserver.calltimeout"),
				mgr.viper.GetInt("connserver.retry"), mod)

			// add async queue and init later.
			mgr.failedMods = append(mgr.failedMods, mod)
		}
	}

	if len(mgr.failedMods) == 0 {
		logger.Info("init all appmods success!")
	} else {
		logger.Info("init a part of appmods success, and try to init the failed[%d] mods later!", len(mgr.failedMods))
		go mgr.initFailedMods()
	}
}
