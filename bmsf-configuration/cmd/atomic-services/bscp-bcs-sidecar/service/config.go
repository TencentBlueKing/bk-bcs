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

	"bk-bscp/internal/safeviper"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/common"
)

const (
	// ENVPREFIX is prefix for env variables.
	ENVPREFIX = "BSCP_BCSSIDECAR"
)

// config for local module.
type config struct {
	viper     *viper.Viper
	safeViper *safeviper.SafeViper
}

// init initialize and check the module configs.
func (c *config) init(localConfigFile string) (*safeviper.SafeViper, error) {
	c.viper = viper.GetViper()
	c.viper.SetConfigFile(localConfigFile)

	if err := c.viper.ReadInConfig(); err != nil {
		return nil, err
	}
	c.viper.WatchConfig()

	if err := c.check(); err != nil {
		return nil, err
	}
	c.safeViper = safeviper.NewSafeViper(c.viper)

	return c.safeViper, nil
}

func (c *config) envName(bindKey string) string {
	return ENVPREFIX + "_" + bindKey
}

// check bind the env vars and checks base config content.
func (c *config) check() error {
	// sidecar base configs.
	c.viper.BindEnv("sidecar.pullConfigInterval", c.envName("PULL_CFG_INTERVAL"))
	c.viper.SetDefault("sidecar.pullConfigInterval", 5*time.Minute)

	c.viper.BindEnv("sidecar.pullConfigRetry", c.envName("PULL_CFG_RETRY"))
	c.viper.SetDefault("sidecar.pullConfigRetry", 3)

	c.viper.BindEnv("sidecar.maxAutoPullInterval", c.envName("MAX_AUTO_PULL_INTERVAL"))
	c.viper.SetDefault("sidecar.maxAutoPullInterval", 10*time.Second)

	c.viper.BindEnv("sidecar.maxAutoPullTimes", c.envName("MAX_AUTO_PULL_TIMES"))
	c.viper.SetDefault("sidecar.maxAutoPullTimes", 3)

	c.viper.BindEnv("sidecar.firstReloadCheckInterval", c.envName("FIRST_RELOAD_CHECK_INTERVAL"))
	c.viper.SetDefault("sidecar.firstReloadCheckInterval", 10*time.Second)

	c.viper.BindEnv("sidecar.syncConfigListInterval", c.envName("SYNC_CFGLIST_INTERVAL"))
	c.viper.SetDefault("sidecar.syncConfigListInterval", 10*time.Minute)

	c.viper.BindEnv("sidecar.reportInfoInterval", c.envName("REPORT_INFO_INTERVAL"))
	c.viper.SetDefault("sidecar.reportInfoInterval", 10*time.Minute)

	c.viper.BindEnv("sidecar.reportInfoLimit", c.envName("REPORT_INFO_LIMIT"))
	c.viper.SetDefault("sidecar.reportInfoLimit", 10)

	c.viper.BindEnv("sidecar.accessInterval", c.envName("ACCESS_INTERVAL"))
	c.viper.SetDefault("sidecar.accessInterval", 3*time.Second)

	c.viper.BindEnv("sidecar.sessionTimeout", c.envName("SESSION_TIMEOUT"))
	c.viper.SetDefault("sidecar.sessionTimeout", 60*time.Second)

	c.viper.BindEnv("sidecar.sessionCoefficient", c.envName("SESSION_COEFFICIENT"))
	c.viper.SetDefault("sidecar.sessionCoefficient", 2)

	c.viper.BindEnv("sidecar.configListPageSize", c.envName("CFGLIST_PAGE_SIZE"))
	c.viper.SetDefault("sidecar.configListPageSize", 100)

	c.viper.BindEnv("sidecar.handlerChSize", c.envName("HANDLER_CH_SIZE"))
	c.viper.SetDefault("sidecar.handlerChSize", 10000)

	c.viper.BindEnv("sidecar.handlerChTimeout", c.envName("HANDLER_CH_TIMEOUT"))
	c.viper.SetDefault("sidecar.handlerChTimeout", time.Second)

	c.viper.BindEnv("sidecar.configHandlerChSize", c.envName("CFG_HANDLER_CH_SIZE"))
	c.viper.SetDefault("sidecar.configHandlerChSize", 10000)

	c.viper.BindEnv("sidecar.configHandlerChTimeout", c.envName("CFG_HANDLER_CH_TIMEOUT"))
	c.viper.SetDefault("sidecar.configHandlerChTimeout", time.Second)

	c.viper.BindEnv("sidecar.pullerChSize", c.envName("CFG_PULLER_CH_SIZE"))
	c.viper.SetDefault("sidecar.pullerChSize", 1)

	c.viper.BindEnv("sidecar.pullerChTimeout", c.envName("CFG_PULLER_CH_TIMEOUT"))
	c.viper.SetDefault("sidecar.pullerChTimeout", time.Second)

	c.viper.BindEnv("sidecar.enableDeleteConfig", c.envName("ENABLE_DELETE_CONFIG"))
	c.viper.SetDefault("sidecar.enableDeleteConfig", true)

	c.viper.BindEnv("sidecar.fileReloadMode", c.envName("FILE_RELOAD_MODE"))
	c.viper.SetDefault("sidecar.fileReloadMode", false)

	c.viper.BindEnv("sidecar.fileReloadName", c.envName("FILE_RELOAD_NAME"))
	c.viper.SetDefault("sidecar.fileReloadName", "BSCP.reload")

	c.viper.BindEnv("sidecar.readyPullConfigs", c.envName("READY_PULL_CONFIGS"))
	c.viper.SetDefault("sidecar.readyPullConfigs", true)

	// gateway configs.
	c.viper.BindEnv("gateway.hostName", c.envName("GW_HOSTNAME"))
	c.viper.SetDefault("gateway.hostName", "gw.bkbscp.bk.com")

	c.viper.BindEnv("gateway.port", c.envName("GW_PORT"))
	c.viper.SetDefault("gateway.port", 8080)

	c.viper.BindEnv("gateway.fileContetAPIPath", c.envName("GW_FILE_CONTENT_API_PATH"))
	c.viper.SetDefault("gateway.fileContetAPIPath", "api/v2/file/content/biz")

	c.viper.BindEnv("gateway.dialTimeout", c.envName("GW_DIAL_TIMEOUT"))
	c.viper.SetDefault("gateway.dialTimeout", 10*time.Second)

	// connserver configs.
	c.viper.BindEnv("connserver.hostName", c.envName("CONNSERVER_HOSTNAME"))
	nodeIP := common.GetenvCfg("BCS_NODE_IP", "")
	if len(nodeIP) != 0 {
		c.viper.Set("connserver.hostName", nodeIP)
	}
	if !c.viper.IsSet("connserver.hostName") {
		return errors.New("config check, missing 'connserver.hostName'")
	}

	c.viper.BindEnv("connserver.port", c.envName("CONNSERVER_PORT"))
	c.viper.SetDefault("connserver.port", 59516)

	c.viper.BindEnv("connserver.dialTimeout", c.envName("CONNSERVER_DIAL_TIMEOUT"))
	c.viper.SetDefault("connserver.dialTimeout", types.RPCShortTimeout)

	c.viper.BindEnv("connserver.callTimeout", c.envName("CONNSERVER_CALL_TIMEOUT"))
	c.viper.SetDefault("connserver.callTimeout", types.RPCNormalTimeout)

	c.viper.BindEnv("connserver.retry", c.envName("CONNSERVER_RETRY"))
	c.viper.SetDefault("connserver.retry", 3)

	// appinfo configs.
	c.viper.BindEnv("appinfo.ipeth", c.envName("APPINFO_IP_ETH"))
	c.viper.SetDefault("appinfo.ipeth", "eth1")

	ipnet, _ := common.GetEthAddr(c.viper.GetString("appinfo.ipeth"))
	podIP := common.GetenvCfg("BCS_CONTAINER_IP", ipnet)
	if len(podIP) == 0 {
		return errors.New("config check, missing podIP/ipNet")
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

	c.viper.BindEnv("instance.dialTimeout", c.envName("INS_DIALTIMEOUT"))
	c.viper.SetDefault("instance.dialTimeout", 3*time.Second)

	c.viper.BindEnv("instance.callTimeout", c.envName("INS_CALLTIMEOUT"))
	c.viper.SetDefault("instance.callTimeout", 10*time.Second)

	// check reload modes.
	if c.viper.GetBool("sidecar.fileReloadMode") && c.viper.GetBool("instance.open") {
		return errors.New("config check, can't open filereload mode and instance server in the same time")
	}

	// normal env settings.
	appInfoModEnvVal := common.GetenvCfg(c.envName("APPINFO_MOD"), "")

	if len(appInfoModEnvVal) != 0 {
		c.viper.Set("appmods", appInfoModEnvVal)
	} else {
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
				BizID:   GetAppModInfoValue(m["biz_id"]),
				AppID:   GetAppModInfoValue(m["app_id"]),
				Path:    GetAppModInfoValue(m["path"]),
				CloudID: GetAppModInfoValue(m["cloud_id"]),
				Labels:  make(map[string]string),
			}

			if m["labels"] != nil {
				labels := m["labels"].(map[interface{}]interface{})

				for labelK, labelV := range labels {
					k := GetAppModInfoValue(labelK)
					v := GetAppModInfoValue(labelV)
					if len(k) != 0 && len(v) != 0 {
						newMod.Labels[k] = v
					}
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
	c.viper.BindEnv("cache.contentCachePath", c.envName("CONTENT_CACHE_PATH"))
	c.viper.SetDefault("cache.contentCachePath", "./bscp-cache/ccache/")

	c.viper.BindEnv("cache.linkContentCachePath", c.envName("LINK_CONTENT_CACHE_PATH"))
	c.viper.SetDefault("cache.linkContentCachePath", "./bscp-cache/lcache/")

	c.viper.BindEnv("cache.contentExpiredPath", c.envName("CONTENT_EXPIRED_CACHE_PATH"))
	c.viper.SetDefault("cache.contentExpiredPath", "/tmp/")

	// content cache max disk usage rate, default is 10%.
	c.viper.BindEnv("cache.contentCacheMaxDiskUsageRate", c.envName("CONTENT_CACHE_MAX_DISK_USAGE_RATE"))
	c.viper.SetDefault("cache.contentCacheMaxDiskUsageRate", 10)

	c.viper.BindEnv("cache.contentCacheDiskUsageCheckInterval", c.envName("CONTENT_CACHE_DISK_USAGE_CHECK_INTERVAL"))
	c.viper.SetDefault("cache.contentCacheDiskUsageCheckInterval", time.Minute)

	c.viper.BindEnv("cache.contentCacheExpiration", c.envName("CONTENT_CACHE_EXPIRATION"))
	c.viper.SetDefault("cache.contentCacheExpiration", 7*24*time.Hour)

	// download gcroutine num for per config file, default is 5.
	c.viper.BindEnv("cache.downloadPerFileConcurrent", c.envName("DOWNLOAD_PER_FILE_CONCURRENT"))
	c.viper.SetDefault("cache.downloadPerFileConcurrent", 5)

	// download limit bytes in second for per config file, default is 1MB.
	c.viper.BindEnv("cache.downloadPerFileLimitBytesInSecond", c.envName("DOWNLOAD_PER_FILE_LIMIT_BYTES"))
	c.viper.SetDefault("cache.downloadPerFileLimitBytesInSecond", 1*1024*1024)

	c.viper.BindEnv("cache.downloadTimeout", c.envName("DOWNLOAD_TIMEOUT"))
	c.viper.SetDefault("cache.downloadTimeout", 30*time.Minute)

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
