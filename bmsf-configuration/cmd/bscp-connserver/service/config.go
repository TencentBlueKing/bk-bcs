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
	"errors"
	"time"

	"github.com/spf13/viper"
)

const (
	// ENVPREFIX is prefix for env variables.
	ENVPREFIX = "BSCP_CONNS"
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
	c.viper.BindEnv("server.servicename", c.envName("SERVICE_NAME"))
	if !c.viper.IsSet("server.servicename") {
		return errors.New("config check, missing 'server.servicename'")
	}

	c.viper.BindEnv("server.endpoint.ip", c.envName("ENDPOINT_IP"))
	if !c.viper.IsSet("server.endpoint.ip") {
		return errors.New("config check, missing 'server.endpoint.ip'")
	}

	c.viper.BindEnv("server.endpoint.port", c.envName("ENDPOINT_PORT"))
	if !c.viper.IsSet("server.endpoint.port") {
		return errors.New("config check, missing 'server.endpoint.port'")
	}

	c.viper.BindEnv("server.discoveryttl", c.envName("DISCOVERY_TTL"))
	c.viper.SetDefault("server.discoveryttl", 60)

	c.viper.BindEnv("server.keepaliveinterval", c.envName("KEEPALIVE_INTERVAL"))
	c.viper.SetDefault("server.keepaliveinterval", 60*time.Second)

	c.viper.BindEnv("server.keepalivetimeout", c.envName("KEEPALIVE_TIMEOUT"))
	c.viper.SetDefault("server.keepalivetimeout", time.Second)

	c.viper.BindEnv("server.schedule.nodesLimit", c.envName("SCHEDULE_NODES_LIMIT"))
	c.viper.SetDefault("server.schedule.nodesLimit", 10)

	c.viper.BindEnv("server.reportinterval", c.envName("REPORT_INTERVAL"))
	c.viper.SetDefault("server.reportinterval", 3*time.Second)

	c.viper.BindEnv("server.pubChanTimeout", c.envName("PUB_CHAN_TIMEOUT"))
	c.viper.SetDefault("server.pubChanTimeout", 3*time.Second)

	c.viper.BindEnv("server.configsCacheSize", c.envName("CONFIGS_CACHE_SIZE"))
	c.viper.SetDefault("server.configsCacheSize", 1000)

	c.viper.BindEnv("server.publishStepCount", c.envName("PUB_STEP_COUNT"))
	c.viper.SetDefault("server.publishStepCount", 1000)
	c.viper.BindEnv("server.publishMinUnitSize", c.envName("PUB_MIN_UNIT_SIZE"))
	c.viper.SetDefault("server.publishMinUnitSize", 1)
	c.viper.BindEnv("server.publishStepWait", c.envName("PUB_STEP_WAIT"))
	c.viper.SetDefault("server.publishStepWait", time.Second)

	c.viper.BindEnv("server.executorLimitRate", c.envName("EXEC_LIMIT_RATE"))
	c.viper.SetDefault("server.executorLimitRate", 0)

	c.viper.BindEnv("metrics.endpoint", c.envName("METRICS_ENDPOINT"))
	c.viper.SetDefault("metrics.endpoint", ":9100")
	c.viper.BindEnv("metrics.path", c.envName("METRICS_PATH"))

	c.viper.BindEnv("etcdCluster.endpoints", c.envName("ETCD_ENDPOINTS"))
	if !c.viper.IsSet("etcdCluster.endpoints") {
		return errors.New("config check, missing 'etcdCluster.endpoints'")
	}

	c.viper.BindEnv("etcdCluster.dialtimeout", c.envName("ETCD_DIAL_TIMEOUT"))
	c.viper.SetDefault("etcdCluster.dialtimeout", 3*time.Second)

	c.viper.BindEnv("bcscontroller.servicename", c.envName("BCS_SERVICE_NAME"))
	if !c.viper.IsSet("bcscontroller.servicename") {
		return errors.New("config check, missing 'bcscontroller.servicename'")
	}

	c.viper.BindEnv("bcscontroller.calltimeout", c.envName("BCS_CALL_TIMEOUT"))
	c.viper.SetDefault("bcscontroller.calltimeout", 3*time.Second)

	c.viper.BindEnv("datamanager.servicename", c.envName("DM_SERVICE_NAME"))
	if !c.viper.IsSet("datamanager.servicename") {
		return errors.New("config check, missing 'datamanager.servicename'")
	}

	c.viper.BindEnv("datamanager.calltimeoutST", c.envName("DM_CALL_TIMEOUT_ST"))
	c.viper.SetDefault("datamanager.calltimeoutST", 3*time.Second)

	c.viper.BindEnv("datamanager.calltimeout", c.envName("DM_CALL_TIMEOUT"))
	c.viper.SetDefault("datamanager.calltimeout", 3*time.Second)

	c.viper.BindEnv("natsmqCluster.endpoints", c.envName("NATS_ENDPOINTS"))
	if !c.viper.IsSet("natsmqCluster.endpoints") {
		return errors.New("config check, missing 'natsmqCluster.endpoints'")
	}

	c.viper.BindEnv("natsmqCluster.timeout", c.envName("NATS_TIMEOUT"))
	c.viper.SetDefault("natsmqCluster.timeout", 3*time.Second)

	c.viper.BindEnv("natsmqCluster.reconwait", c.envName("NATS_RECONWAIT"))
	c.viper.SetDefault("natsmqCluster.reconwait", 3*time.Second)

	c.viper.BindEnv("natsmqCluster.maxrecons", c.envName("NATS_MAXRECONS"))
	c.viper.SetDefault("natsmqCluster.maxrecons", 10)

	c.viper.BindEnv("natsmqCluster.publishtopic", c.envName("NATS_PUB_TOPIC"))
	c.viper.SetDefault("natsmqCluster.publishtopic", "bk-bscp-bcs-publishtopic")

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
