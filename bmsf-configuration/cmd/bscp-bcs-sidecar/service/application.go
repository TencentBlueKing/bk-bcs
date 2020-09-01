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
	"path/filepath"
	"sync"
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

	// defaultDynamicLoadModsWaitTime is duration for wait to load app mods.
	defaultDynamicLoadModsWaitTime = time.Second
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

	// duplicate checking.
	duplicateCheckMap map[string]*AppModInfo

	// duplicate checking map mutex.
	duplicateCheckMapMu sync.RWMutex

	// app mod signallings.
	signallings map[string]*SignallingChannel

	// app mod signallings map mutex.
	signallingsMu sync.RWMutex
}

// NewAppModManager creates a new AppModManager.
func NewAppModManager(viper *viper.Viper, reloader *Reloader) *AppModManager {
	return &AppModManager{
		viper:             viper,
		reloader:          reloader,
		duplicateCheckMap: make(map[string]*AppModInfo),
		signallings:       make(map[string]*SignallingChannel),
	}
}

func (mgr *AppModManager) dynamicLoad() {
	isFirstCheck := true

	for {
		// do not wait at first time.
		if !isFirstCheck {
			time.Sleep(defaultDynamicLoadModsWaitTime)
		}
		isFirstCheck = false

		// load appmod value, and handle dynamic app mod load.
		appmodCfgVal := mgr.viper.GetString("appmods")

		dynamicAppMods := []*AppModInfo{}
		currentAppModMap := make(map[string]*AppModInfo)

		if len(appmodCfgVal) != 0 {
			if err := json.Unmarshal([]byte(appmodCfgVal), &dynamicAppMods); err != nil {
				logger.Errorf("AppModManager| can't dynamic init appinfo, unmarshal %+v", err)
				continue
			}
		}

		// init new app mods.
		for _, mod := range dynamicAppMods {
			// must clean path at first.
			mod.Path = filepath.Clean(mod.Path)

			// mod key.
			modKey := ModKey(mod.BusinessName, mod.AppName, mod.Path)

			// record current app mod.
			currentAppModMap[modKey] = mod

			// can't re-init the mod that is already inited before, the way to update
			// target app mod is just delete it and re-init it.
			mgr.duplicateCheckMapMu.RLock()
			if _, isExist := mgr.duplicateCheckMap[modKey]; isExist {
				mgr.duplicateCheckMapMu.RUnlock()
				continue
			}
			mgr.duplicateCheckMapMu.RUnlock()

			logger.Info("AppModManager| setup new app mod[%+v]", mod)

			// base infos check.
			if len(mod.BusinessName) == 0 || len(mod.AppName) == 0 || len(mod.Path) == 0 {
				logger.Errorf("AppModManager| can't setup the new app mod, appinfo missing, %+v", mod)
				continue
			}

			// set configs cache path.
			mgr.viper.Set(fmt.Sprintf("appmod.%s.path", modKey), mod.Path)

			// set app mod cluster name.
			if len(mod.ClusterName) == 0 {
				mod.ClusterName = defaultCluster
			}
			clusterLabels, clusterName := common.ParseClusterLabels(mod.ClusterName)
			mgr.viper.Set(fmt.Sprintf("appmod.%s.cluster", modKey), clusterName)
			mgr.viper.Set(fmt.Sprintf("appmod.%s.clusterLabels", modKey), clusterLabels)

			// set app mod zone name.
			if len(mod.ZoneName) == 0 {
				mod.ZoneName = defaultZone
			}
			mgr.viper.Set(fmt.Sprintf("appmod.%s.zone", modKey), mod.ZoneName)

			// set dc tag, maybe empty.
			if len(mod.DC) == 0 {
				mod.DC = defaultDC
			}
			mgr.viper.Set(fmt.Sprintf("appmod.%s.dc", modKey), mod.DC)

			// set labels, maybe empty.
			mgr.viper.Set(fmt.Sprintf("appmod.%s.labels", modKey), mod.Labels)

			// setup the new app mod.
			mgr.setupAppMod(mod)
		}

		// stop deleted mods.
		mgr.stopAppMods(currentAppModMap)
	}
}

// initialize single appinfo.
func (mgr *AppModManager) initSingleAppinfo(businessName, appName, clusterName, zoneName, path string) error {
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
		Seq:           common.Sequence(),
		BusinessName:  businessName,
		AppName:       appName,
		ClusterName:   clusterName,
		ZoneName:      zoneName,
		ClusterLabels: mgr.viper.GetString(fmt.Sprintf("appmod.%s.clusterLabels", ModKey(businessName, appName, path))),
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
	mgr.viper.Set(fmt.Sprintf("appmod.%s.bid", ModKey(businessName, appName, path)), resp.Bid)
	mgr.viper.Set(fmt.Sprintf("appmod.%s.appid", ModKey(businessName, appName, path)), resp.Appid)
	mgr.viper.Set(fmt.Sprintf("appmod.%s.clusterid", ModKey(businessName, appName, path)), resp.Clusterid)
	mgr.viper.Set(fmt.Sprintf("appmod.%s.zoneid", ModKey(businessName, appName, path)), resp.Zoneid)

	logger.Info("AppModManager| init new appinfo success[%s %s %s %s %s], bid[%+v] appid[%+v] clusterid[%+v] zoneid[%+v]",
		businessName, appName, clusterName, zoneName, path, resp.Bid, resp.Appid, resp.Clusterid, resp.Zoneid)

	return nil
}

// initialize config cache.
func (mgr *AppModManager) initCache(businessName, appName, path string) (*EffectCache, *ContentCache) {
	// config release effect cache, "filepath(etc)/businessName/appName/path".
	effectCache := NewEffectCache(fmt.Sprintf("%s/%s/%s/%s",
		mgr.viper.GetString("cache.fileCachePath"), businessName, appName, path), businessName, appName, path)

	// content cache.
	contentCache := NewContentCache(
		mgr.viper,
		mgr.viper.GetString("cache.contentCachePath"),
		businessName,
		appName,
		path,
		mgr.viper.GetInt("cache.contentMCacheSize"),
		mgr.viper.GetString("cache.contentExpiredCachePath"),
		mgr.viper.GetDuration("cache.mcacheExpiration"),
		mgr.viper.GetDuration("cache.contentCacheExpiration"),
		mgr.viper.GetDuration("cache.contentCachePurgeInterval"),
	)
	go contentCache.Setup()

	logger.Info("AppModManager| init single[%s %s %s] effect/content cache success.", businessName, appName, path)
	return effectCache, contentCache
}

// initialize sidecar config handler.
func (mgr *AppModManager) initHandler(businessName, appName, path string, effectCache *EffectCache, contentCache *ContentCache) *Handler {
	// config handler.
	configHandler := NewConfigHandler(mgr.viper, businessName, appName, path, effectCache, contentCache, mgr.reloader)

	// handler.
	handler := NewHandler(mgr.viper, businessName, appName, path, configHandler)

	logger.Info("AppModManager| init single[%s %s %s] sidecar handler success.", businessName, appName, path)
	return handler
}

// initInnerMods initialize the inner modules for target app.
func (mgr *AppModManager) initInnerMods(businessName, appName, path string) {
	modKey := ModKey(businessName, appName, path)

	// wait for ready to pull configs.
	for {
		// main flag to control all mods.
		if mgr.viper.GetBool("sidecar.readyPullConfigs") {
			break
		}

		// app mod level flags.
		if mgr.viper.GetBool(fmt.Sprintf("sidecar.%s.readyPullConfigs", modKey)) {
			break
		}
		logger.Warn("AppModManager| [%s %s %s] waiting for pulling configs until flag mark ready", businessName, appName, path)

		time.Sleep(defaultInitInnerModsWaitTime)
	}
	logger.Info("AppModManager| init inner mods for [%s %s %s] right now!", businessName, appName, path)

	// eliminate summit.
	common.DelayRandomMS(2500)

	// update app mod stop flag to false.
	mgr.viper.Set(fmt.Sprintf("appmod.%s.stop", modKey), false)

	// initialize cache.
	effectCache, contentCache := mgr.initCache(businessName, appName, path)

	// initialize handler.
	handler := mgr.initHandler(businessName, appName, path, effectCache, contentCache)

	// initialize signalling channel.
	signalling := NewSignallingChannel(mgr.viper, businessName, appName, path, handler)

	// run sidecar config handler.
	handler.Run()
	logger.Info("AppModManager| new sidecar config handler[%s %s %s] run success.", businessName, appName, path)

	// setup signalling channel.
	go signalling.Setup()

	// check target app mod exists or not this moment, and add new signalling for it.
	mgr.duplicateCheckMapMu.RLock()
	if _, isExist := mgr.duplicateCheckMap[modKey]; isExist {
		// mark and save new signalling, only lock for safely, the signalling of target
		// mod would not be duplicated here.
		mgr.signallingsMu.Lock()
		mgr.signallings[modKey] = signalling
		mgr.signallingsMu.Unlock()

		logger.Info("AppModManager| new sidecar signallingChannel[%s %s %s] run success.", businessName, appName, path)
	} else {
		signalling.Close()
	}
	mgr.duplicateCheckMapMu.RUnlock()
}

// AppModInfos returns app mod infos.
func (mgr *AppModManager) AppModInfos() []*AppModInfo {
	mgr.duplicateCheckMapMu.RLock()
	defer mgr.duplicateCheckMapMu.RUnlock()

	mods := []*AppModInfo{}

	for _, mod := range mgr.duplicateCheckMap {
		mods = append(mods, mod)
	}
	return mods
}

// setupAppMod setups target app mod.
func (mgr *AppModManager) setupAppMod(mod *AppModInfo) {
	// connserver retry times.
	retryTimes := mgr.viper.GetInt("connserver.retry")
	if retryTimes < 0 {
		retryTimes = 0
	}

	// init current app mod success flag.
	isInitSucc := false

	// call connserver and init app mod info with retry action.
	for i := 0; i <= retryTimes; i++ {
		if err := mgr.initSingleAppinfo(mod.BusinessName, mod.AppName, mod.ClusterName, mod.ZoneName, mod.Path); err != nil {
			logger.Warn("AppModManager| init app mod[%+v] failed[%d], %+v", mod, i, err)

			// retry.
			continue
		}

		// init inner modules for app.
		logger.Info("AppModManager| init app mod[%+v][%d] success!", mod, i)

		mgr.duplicateCheckMapMu.Lock()
		mgr.duplicateCheckMap[ModKey(mod.BusinessName, mod.AppName, mod.Path)] = mod
		mgr.duplicateCheckMapMu.Unlock()

		// init inner mods in another coroutine.
		go mgr.initInnerMods(mod.BusinessName, mod.AppName, mod.Path)

		isInitSucc = true
		break
	}

	if !isInitSucc {
		logger.Warn("AppModManager| finally[%+v %+v][%d], can't init mod, %+v, try again later!",
			mgr.viper.GetDuration("connserver.dialtimeout"), mgr.viper.GetDuration("connserver.calltimeout"),
			mgr.viper.GetInt("connserver.retry"), mod)
	}
}

// stop and delete old app mods.
func (mgr *AppModManager) stopAppMods(currentAppModMap map[string]*AppModInfo) {
	mgr.duplicateCheckMapMu.Lock()
	defer mgr.duplicateCheckMapMu.Unlock()

	for modKey, _ := range mgr.duplicateCheckMap {
		if _, isExist := currentAppModMap[modKey]; isExist {
			continue
		}

		// stop and delete the old app mod.
		mgr.signallingsMu.Lock()
		if signalling := mgr.signallings[modKey]; signalling != nil {
			signalling.Close()
			delete(mgr.signallings, modKey)
		}
		mgr.signallingsMu.Unlock()

		delete(mgr.duplicateCheckMap, modKey)
	}
}

// Setup setups app mod manager.
func (mgr *AppModManager) Setup() {
	go mgr.dynamicLoad()
	go mgr.debug()
}

// add debug infos here, do not add more interval logs.
func (mgr *AppModManager) debug() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// current inited app mods count.
		mgr.duplicateCheckMapMu.RLock()
		modsCount := len(mgr.duplicateCheckMap)
		mgr.duplicateCheckMapMu.RUnlock()

		// current setuped signallings count.
		mgr.signallingsMu.RLock()
		signallingsCount := len(mgr.signallings)
		mgr.signallingsMu.RUnlock()

		// debug app mod infos.
		logger.Info("AppModManager| debug, current appmods mod[%d] signallings[%d], %+v",
			modsCount, signallingsCount, mgr.viper.GetString("appmods"))
	}
}
