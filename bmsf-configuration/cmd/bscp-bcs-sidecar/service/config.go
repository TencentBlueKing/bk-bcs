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
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/pkg/common"
)

const (
	// ENVPREFIX is prefix for env variables.
	ENVPREFIX = "BSCP_BCSSIDECAR"
)

// config for local module.
type config struct {
	viper *viper.Viper
}

// init initialize and check the module configs.
func (c *config) init(localConfigFile string) (*viper.Viper, error) {
	c.viper = viper.New()
	c.viper.SetConfigFile(localConfigFile)

	if err := c.viper.ReadInConfig(); err != nil {
		return nil, err
	}
	c.viper.WatchConfig()

	c.viper.SetEnvPrefix(ENVPREFIX)
	c.viper.AutomaticEnv()

	if err := c.check(); err != nil {
		return nil, err
	}
	return c.viper, nil
}

func (c *config) envName(bindKey string) string {
	return ENVPREFIX + "_" + bindKey
}

// check bind the env vars and checks base config content.
func (c *config) check() error {
	// sidecar base configs.
	c.viper.BindEnv("sidecar.pullConfigInterval", c.envName("PULL_CFG_INTERVAL"))
	c.viper.SetDefault("sidecar.pullConfigInterval", 10*time.Minute)

	c.viper.BindEnv("sidecar.syncConfigsetListInterval", c.envName("SYNC_CFGSETLIST_INTERVAL"))
	c.viper.SetDefault("sidecar.syncConfigsetListInterval", 10*time.Minute)

	c.viper.BindEnv("sidecar.reportInfoInterval", c.envName("REPORT_INFO_INTERVAL"))
	c.viper.SetDefault("sidecar.reportInfoInterval", 10*time.Minute)

	c.viper.BindEnv("sidecar.accessInterval", c.envName("ACCESS_INTERVAL"))
	c.viper.SetDefault("sidecar.accessInterval", 2*time.Second)

	c.viper.BindEnv("sidecar.sessionTimeout", c.envName("SESSION_TIMEOUT"))
	c.viper.SetDefault("sidecar.sessionTimeout", 5*time.Second)

	c.viper.BindEnv("sidecar.sessionCoefficient", c.envName("SESSION_COEFFICIENT"))
	c.viper.SetDefault("sidecar.sessionCoefficient", 2)

	c.viper.BindEnv("sidecar.configSetListSize", c.envName("CFGSETLIST_SIZE"))
	c.viper.SetDefault("sidecar.configSetListSize", 100)

	c.viper.BindEnv("sidecar.handlerChSize", c.envName("HANDLER_CH_SIZE"))
	c.viper.SetDefault("sidecar.handlerChSize", 10000)

	c.viper.BindEnv("sidecar.handlerChTimeout", c.envName("HANDLER_CH_TIMEOUT"))
	c.viper.SetDefault("sidecar.handlerChTimeout", time.Second)

	c.viper.BindEnv("sidecar.configHandlerChSize", c.envName("CFG_HANDLER_CH_SIZE"))
	c.viper.SetDefault("sidecar.configHandlerChSize", 10000)

	c.viper.BindEnv("sidecar.configHandlerChTimeout", c.envName("CFG_HANDLER_CH_TIMEOUT"))
	c.viper.SetDefault("sidecar.configHandlerChTimeout", time.Second)

	c.viper.BindEnv("sidecar.fileReloadMode", c.envName("FILE_RELOAD_MODE"))
	c.viper.SetDefault("sidecar.fileReloadMode", false)

	c.viper.BindEnv("sidecar.fileReloadFName", c.envName("FILE_RELOAD_FNAME"))
	c.viper.SetDefault("sidecar.fileReloadFName", "BSCP.reload")

	c.viper.BindEnv("sidecar.readyPullConfigs", c.envName("READY_PULL_CONFIGS"))
	c.viper.SetDefault("sidecar.readyPullConfigs", true)

	// connserver configs.
	c.viper.BindEnv("connserver.hostname", c.envName("CONNSERVER_HOSTNAME"))
	c.viper.SetDefault("connserver.hostname", "conn.bscp.bk.com")

	c.viper.BindEnv("connserver.port", c.envName("CONNSERVER_PORT"))
	c.viper.SetDefault("connserver.port", 9516)

	c.viper.BindEnv("connserver.dialtimeout", c.envName("CONNSERVER_DIAL_TIMEOUT"))
	c.viper.SetDefault("connserver.dialtimeout", 10*time.Second)

	c.viper.BindEnv("connserver.calltimeout", c.envName("CONNSERVER_CALL_TIMEOUT"))
	c.viper.SetDefault("connserver.calltimeout", 60*time.Second)

	c.viper.BindEnv("connserver.retry", c.envName("CONNSERVER_RETRY"))
	c.viper.SetDefault("connserver.retry", 5)

	// appinfo configs.
	c.viper.BindEnv("appinfo.ipeth", c.envName("APPINFO_IP_ETH"))
	c.viper.SetDefault("appinfo.ipeth", "eth1")

	ipnet, _ := common.GetEthAddr(c.viper.GetString("appinfo.ipeth"))
	podIP := common.GetenvCfg("BCS_CONTAINER_IP", ipnet)
	if len(podIP) == 0 {
		return errors.New("config check, missing podip / ipnet")
	}

	c.viper.BindEnv("appinfo.ip", c.envName("APPINFO_IP"))
	c.viper.SetDefault("appinfo.ip", podIP)

	// instance http server.
	c.viper.BindEnv("instance.open", c.envName("INS_OPEN"))
	c.viper.SetDefault("instance.open", false)

	c.viper.BindEnv("instance.reloadChanSize", c.envName("INS_RELOAD_CHAN_SIZE"))
	c.viper.SetDefault("instance.reloadChanSize", 10000)

	c.viper.BindEnv("instance.reloadChanTimeout", c.envName("INS_RELOAD_CHAN_TIMEOUT"))
	c.viper.SetDefault("instance.reloadChanTimeout", 3*time.Second)

	c.viper.BindEnv("instance.httpEndpoint.ip", c.envName("INS_HTTP_ENDPOINT_IP"))
	c.viper.SetDefault("instance.httpEndpoint.ip", "localhost")

	c.viper.BindEnv("instance.httpEndpoint.port", c.envName("INS_HTTP_ENDPOINT_PORT"))
	c.viper.SetDefault("instance.httpEndpoint.port", 39610)

	c.viper.BindEnv("instance.grpcEndpoint.ip", c.envName("INS_GRPC_ENDPOINT_IP"))
	c.viper.SetDefault("instance.grpcEndpoint.ip", "localhost")

	c.viper.BindEnv("instance.grpcEndpoint.port", c.envName("INS_GRPC_ENDPOINT_PORT"))
	c.viper.SetDefault("instance.grpcEndpoint.port", 39611)

	c.viper.BindEnv("instance.dialtimeout", c.envName("INS_DIALTIMEOUT"))
	c.viper.SetDefault("instance.dialtimeout", 3*time.Second)

	c.viper.BindEnv("instance.calltimeout", c.envName("INS_CALLTIMEOUT"))
	c.viper.SetDefault("instance.calltimeout", 3*time.Second)

	// check reload modes.
	if c.viper.GetBool("sidecar.fileReloadMode") && c.viper.GetBool("instance.open") {
		return errors.New("config check, can't open filereload mode and instance server in the same time")
	}

	// env settings compatibility.
	singleAppCfgPathVal := common.GetenvCfg(c.envName("APPCFG_PATH"), "")
	singleAppInfoBuVal := common.GetenvCfg(c.envName("APPINFO_BUSINESS"), "")
	singleAppInfoAppVal := common.GetenvCfg(c.envName("APPINFO_APP"), "")
	singleAppInfoClusterVal := common.GetenvCfg(c.envName("APPINFO_CLUSTER"), "")
	singleAppInfoZoneVal := common.GetenvCfg(c.envName("APPINFO_ZONE"), "")
	singleAppInfoDCVal := common.GetenvCfg(c.envName("APPINFO_DC"), "")
	singleAppInfoLabelsVal := common.GetenvCfg(c.envName("APPINFO_LABELS"), "{}")

	// normal env settings.
	appInfoModEnvVal := common.GetenvCfg(c.envName("APPINFO_MOD"), "")

	if len(appInfoModEnvVal) != 0 {
		// use appinfo mod config from env.
		c.viper.Set("appmods", appInfoModEnvVal)

	} else if len(singleAppCfgPathVal) != 0 && len(singleAppInfoBuVal) != 0 && len(singleAppInfoAppVal) != 0 {
		newMod := AppModInfo{
			BusinessName: singleAppInfoBuVal,
			AppName:      singleAppInfoAppVal,
			ClusterName:  singleAppInfoClusterVal,
			ZoneName:     singleAppInfoZoneVal,
			DC:           singleAppInfoDCVal,
			Path:         singleAppCfgPathVal,
			Labels:       make(map[string]string),
		}
		if err := json.Unmarshal([]byte(singleAppInfoLabelsVal), &newMod.Labels); err != nil {
			return fmt.Errorf("config check, invalid appinfo labels, %+v", err)
		}

		// marshal to common format as env.
		appModInfos := []AppModInfo{newMod}
		appInfoModCfgVal, err := json.Marshal(&appModInfos)
		if err != nil {
			return fmt.Errorf("config check, can't marshal appmod from single envs, %+v", err)
		}

		c.viper.Set("appmods", string(appInfoModCfgVal))

	} else {
		// use appinfo mod config from local file.
		if !c.viper.IsSet("appinfo.mod") {
			return errors.New("config check, missing 'appinfo.mod'")
		}

		modCfg := c.viper.Get("appinfo.mod")
		if modCfg == nil {
			return errors.New("config check, missing 'appinfo.mod' nil")
		}

		modSlice := modCfg.([]interface{})
		if len(modSlice) == 0 {
			return errors.New("config check, missing 'appinfo.mod', empty mods")
		}

		appModInfos := []AppModInfo{}

		for _, mod := range modSlice {
			if mod == nil {
				continue
			}
			m := mod.(map[interface{}]interface{})

			newMod := AppModInfo{
				BusinessName: m["business"].(string),
				AppName:      m["app"].(string),
				Path:         m["path"].(string),
				Labels:       make(map[string]string),
			}

			if m["cluster"] != nil {
				newMod.ClusterName = m["cluster"].(string)
			}
			if m["zone"] != nil {
				newMod.ZoneName = m["zone"].(string)
			}
			if m["dc"] != nil {
				newMod.DC = m["dc"].(string)
			}
			if m["labels"] != nil {
				labels := m["labels"].(map[interface{}]interface{})
				for labelk, labelv := range labels {
					newMod.Labels[labelk.(string)] = labelv.(string)
				}
			}

			appModInfos = append(appModInfos, newMod)
		}

		// marshal to common format as env.
		appInfoModCfgVal, err := json.Marshal(&appModInfos)
		if err != nil {
			return fmt.Errorf("config check, can't marshal appmod from local file, %+v", err)
		}

		c.viper.Set("appmods", string(appInfoModCfgVal))
	}

	// cache configs.
	c.viper.BindEnv("cache.fileCachePath", c.envName("FILE_CACHE_PATH"))
	c.viper.SetDefault("cache.fileCachePath", "./cache/fcache/")

	c.viper.BindEnv("cache.contentCachePath", c.envName("CONTENT_CACHE_PATH"))
	c.viper.SetDefault("cache.contentCachePath", "./cache/ccache/")

	c.viper.BindEnv("cache.contentExpiredCachePath", c.envName("CONTENT_EXPCACHE_PATH"))
	c.viper.SetDefault("cache.contentExpiredCachePath", "/tmp/")

	c.viper.BindEnv("cache.contentMCacheSize", c.envName("CONTENT_MCACHE_SIZE"))
	c.viper.SetDefault("cache.contentMCacheSize", 1000)

	c.viper.BindEnv("cache.mcacheExpiration", c.envName("CONTENT_MCACHE_EXPIRATION"))
	c.viper.SetDefault("cache.mcacheExpiration", 10*time.Minute)

	c.viper.BindEnv("cache.contentCacheExpiration", c.envName("CONTENT_CACHE_EXPIRATION"))
	c.viper.SetDefault("cache.contentCacheExpiration", 7*24*time.Hour)

	c.viper.BindEnv("cache.contentCachePurgeInterval", c.envName("CONTENT_CACHE_PURGE_INTERVAL"))
	c.viper.SetDefault("cache.contentCachePurgeInterval", 30*time.Minute)

	c.viper.BindEnv("logger.directory", c.envName("LOG_DIR"))
	c.viper.SetDefault("logger.directory", "./log")

	c.viper.BindEnv("logger.maxsize", c.envName("LOG_MAXSIZE"))
	c.viper.SetDefault("logger.maxsize", 200)

	c.viper.BindEnv("logger.maxnum", c.envName("LOG_MAXNUM"))
	c.viper.SetDefault("logger.maxnum", 5)

	c.viper.BindEnv("logger.stderr", c.envName("LOG_STDERR"))
	c.viper.SetDefault("logger.stderr", false)

	c.viper.BindEnv("logger.alsoStderr", c.envName("LOG_ALSOSTDERR"))
	c.viper.SetDefault("logger.alsoStderr", false)

	c.viper.BindEnv("logger.level", c.envName("LOG_LEVEL"))
	c.viper.SetDefault("logger.level", 3)

	c.viper.BindEnv("logger.stderrThreshold", c.envName("LOG_STDERR_THRESHOLD"))
	c.viper.SetDefault("logger.stderrThreshold", 2)

	return nil
}
