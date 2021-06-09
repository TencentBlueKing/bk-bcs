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

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

const (
	// defaultCloudID default cloudid.
	defaultCloudID = "default"

	// defaultInitInnerModsWaitTime is duration for wait to init inner mods.
	defaultInitInnerModsWaitTime = time.Second

	// defaultDynamicLoadModsWaitTime is duration for wait to load app mods.
	defaultDynamicLoadModsWaitTime = 3 * time.Second
)

// AppModInfo is multi app mode information.
type AppModInfo struct {
	// BizID business id.
	BizID string `json:"biz_id"`

	// AppID app id.
	AppID string `json:"app_id"`

	// CloudID data center tag.
	CloudID string `json:"cloud_id"`

	// Path is sidecar mod app configs effect path.
	Path string `json:"path"`

	// Labels sidecar instance KV labels.
	Labels map[string]string `json:"labels"`
}

// AppModManager is app mod manager.
type AppModManager struct {
	// configs handler.
	viper *safeviper.SafeViper

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
func NewAppModManager(viper *safeviper.SafeViper, reloader *Reloader) *AppModManager {
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
			modKey := ModKey(mod.BizID, mod.AppID, mod.Path)

			// record current app mod.
			currentAppModMap[modKey] = mod

			// NOTE: can't re-init the mod that is already inited before, the way to update
			// target app mod is just delete it and re-init it.
			mgr.duplicateCheckMapMu.RLock()
			if _, isExist := mgr.duplicateCheckMap[modKey]; isExist {
				mgr.signallingsMu.RLock()
				signalling := mgr.signallings[modKey]
				mgr.signallingsMu.RUnlock()

				// update instance labels.
				signalling.Reset(mod.Labels)

				mgr.duplicateCheckMapMu.RUnlock()
				continue
			}
			mgr.duplicateCheckMapMu.RUnlock()

			logger.Warnf("AppModManager| setup new app mod[%+v]", mod)

			// base infos check.
			if len(mod.BizID) == 0 || len(mod.AppID) == 0 || len(mod.Path) == 0 {
				logger.Errorf("AppModManager| can't setup the new app mod, appinfo missing, %+v", mod)
				continue
			}

			// set configs cache path.
			mgr.viper.Set(fmt.Sprintf("appmod.%s.path", modKey), mod.Path)

			// set cloudid tag, maybe empty.
			if len(mod.CloudID) == 0 {
				mod.CloudID = defaultCloudID
			}
			mgr.viper.Set(fmt.Sprintf("appmod.%s.cloudid", modKey), mod.CloudID)

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
func (mgr *AppModManager) initSingleAppinfo(bizID, appID, path string) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(mgr.viper.GetDuration("connserver.dialTimeout")),
	}
	endpoint := mgr.viper.GetString("connserver.hostName") + ":" + mgr.viper.GetString("connserver.port")

	// connect to connserver.
	c, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return fmt.Errorf("dial connserver, %+v", err)
	}
	defer c.Close()
	client := pb.NewConnectionClient(c)

	// query app metadata now.
	r := &pb.QueryAppMetadataReq{
		Seq:   common.Sequence(),
		BizId: bizID,
		AppId: appID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), mgr.viper.GetDuration("connserver.callTimeout"))
	defer cancel()

	logger.Info("AppModManager| query app metadata, %+v", r)

	resp, err := client.QueryAppMetadata(ctx, r)
	if err != nil {
		return fmt.Errorf("query metadata, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return fmt.Errorf("can't query app metadata, %+v", resp)
	}

	// query success, save metadata to local.
	mgr.viper.Set(fmt.Sprintf("appmod.%s.bizid", ModKey(bizID, appID, path)), bizID)
	mgr.viper.Set(fmt.Sprintf("appmod.%s.appid", ModKey(bizID, appID, path)), appID)

	logger.Info("AppModManager| init new appinfo success, bizid[%+v] appid[%+v] path[%+v]", bizID, appID, path)

	return nil
}

// initialize config cache.
func (mgr *AppModManager) initCache(bizID, appID, path string) (*EffectCache, *ContentCache) {
	// config release effect cache, "filepath(etc)/bizid/appid/path".
	effectCache := NewEffectCache(bizID, appID, path)

	// content cache.
	contentCache := NewContentCache(
		mgr.viper,
		bizID,
		appID,
		path,
		mgr.viper.GetString("cache.contentCachePath"),
		mgr.viper.GetString("cache.linkContentCachePath"),
	)
	logger.Info("AppModManager| init single[%s %s %s] effect/content cache success.", bizID, appID, path)

	return effectCache, contentCache
}

// initialize sidecar config handler.
func (mgr *AppModManager) initHandler(bizID, appID, path string,
	effectCache *EffectCache, contentCache *ContentCache) *Handler {

	// config handler.
	configHandler := NewConfigHandler(mgr.viper, bizID, appID, path, effectCache, contentCache, mgr.reloader)

	// handler.
	handler := NewHandler(mgr.viper, bizID, appID, path, configHandler)

	logger.Info("AppModManager| init single[%s %s %s] sidecar handler success.", bizID, appID, path)
	return handler
}

// initInnerMods initialize the inner modules for target app.
func (mgr *AppModManager) initInnerMods(bizID, appID, path string) {
	modKey := ModKey(bizID, appID, path)

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
		logger.Warn("AppModManager| [%s %s %s] waiting for pulling configs until FLAG marked ready", bizID, appID, path)
		time.Sleep(defaultInitInnerModsWaitTime)
	}

	logger.Info("AppModManager| init inner mods for [%s %s %s] now!", bizID, appID, path)

	// eliminate summit.
	common.DelayRandomMS(2500)

	// update app mod stop flag to false.
	mgr.viper.Set(fmt.Sprintf("appmod.%s.stop", modKey), false)

	// initialize cache.
	effectCache, contentCache := mgr.initCache(bizID, appID, path)

	// initialize handler.
	handler := mgr.initHandler(bizID, appID, path, effectCache, contentCache)

	// initialize signalling channel.
	signalling := NewSignallingChannel(mgr.viper, bizID, appID, path, handler)

	// run sidecar config handler.
	handler.Run()
	logger.Info("AppModManager| new sidecar config handler[%s %s %s] run success.", bizID, appID, path)

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

		logger.Info("AppModManager| new sidecar signallingChannel[%s %s %s] run success.", bizID, appID, path)
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

// AppSignallings returns app signalling channels.
func (mgr *AppModManager) AppSignallings() map[string]*SignallingChannel {
	mgr.signallingsMu.RLock()
	defer mgr.signallingsMu.RUnlock()

	signallings := make(map[string]*SignallingChannel, 0)

	for key, value := range mgr.signallings {
		signallings[key] = value
	}

	return signallings
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
		if err := mgr.initSingleAppinfo(mod.BizID, mod.AppID, mod.Path); err != nil {
			logger.Warn("AppModManager| init app mod[%+v] failed[%d], %+v", mod, i, err)

			// retry.
			continue
		}

		// init inner modules for app.
		logger.Info("AppModManager| init app mod[%+v][%d] success!", mod, i)

		mgr.duplicateCheckMapMu.Lock()
		mgr.duplicateCheckMap[ModKey(mod.BizID, mod.AppID, mod.Path)] = mod
		mgr.duplicateCheckMapMu.Unlock()

		// init inner mods in another coroutine.
		go mgr.initInnerMods(mod.BizID, mod.AppID, mod.Path)

		isInitSucc = true
		break
	}

	if !isInitSucc {
		logger.Warn("AppModManager| finally[%+v %+v][%d], can't init mod, %+v, try again later!",
			mgr.viper.GetDuration("connserver.dialTimeout"), mgr.viper.GetDuration("connserver.callTimeout"),
			mgr.viper.GetInt("connserver.retry"), mod)
	}
}

// stop and delete old app mods.
func (mgr *AppModManager) stopAppMods(currentAppModMap map[string]*AppModInfo) {
	mgr.duplicateCheckMapMu.Lock()
	defer mgr.duplicateCheckMapMu.Unlock()

	for modKey := range mgr.duplicateCheckMap {
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
	// load app mods in dynamic mode.
	go mgr.dynamicLoad()

	// run content cache cleaner.
	contentCacheCleaner := NewContentCacheCleaner(
		mgr.viper,
		mgr.viper.GetString("cache.contentCachePath"),
		mgr.viper.GetString("cache.contentExpiredPath"),
		mgr.viper.GetInt("cache.contentCacheMaxDiskUsageRate"),
		mgr.viper.GetDuration("cache.contentCacheExpiration"),
		mgr.viper.GetDuration("cache.contentCacheDiskUsageCheckInterval"),
	)
	go contentCacheCleaner.Run()

	// app mod debug.
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
		logger.V(2).Infof("AppModManager| debug, current appmods mod[%d] signallings[%d], %+v",
			modsCount, signallingsCount, mgr.viper.GetString("appmods"))
	}
}
