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
	ENVPREFIX = "BSCP_GW"
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
	c.viper.BindEnv("gateway.endpoint.ip", c.envName("ENDPOINT_IP"))
	c.viper.BindEnv("gateway.endpoint.port", c.envName("ENDPOINT_PORT"))
	c.viper.BindEnv("gateway.insecureEndpoint.ip", c.envName("INSECURE_ENDPOINT_IP"))
	c.viper.BindEnv("gateway.insecureEndpoint.port", c.envName("INSECURE_ENDPOINT_PORT"))

	c.viper.BindEnv("gateway.api.open", c.envName("API_OPEN"))
	c.viper.SetDefault("gateway.api.open", false)

	c.viper.BindEnv("gateway.api.dir", c.envName("API_DIR"))
	c.viper.SetDefault("gateway.api.dir", "swagger")

	c.viper.BindEnv("gateway.tls.certPassword", c.envName("TLS_CERT_PASSWORD"))
	c.viper.SetDefault("gateway.tls.certPassword", "")

	c.viper.BindEnv("gateway.tls.cafile", c.envName("TLS_CAFILE"))
	c.viper.SetDefault("gateway.tls.cafile", "./bscp-ca.crt")

	c.viper.BindEnv("gateway.tls.certfile", c.envName("TLS_CERTFILE"))
	c.viper.SetDefault("gateway.tls.certfile", "./bscp-server.crt")

	c.viper.BindEnv("gateway.tls.keyfile", c.envName("TLS_KEYFILE"))
	c.viper.SetDefault("gateway.tls.keyfile", "./bscp-server.key")

	c.viper.BindEnv("accessserver.servicename", c.envName("AS_SERVICE_NAME"))
	if !c.viper.IsSet("accessserver.servicename") {
		return errors.New("config check, missing 'accessserver.servicename'")
	}

	c.viper.BindEnv("accessserver.calltimeout", c.envName("AS_CALL_TIMEOUT"))
	c.viper.SetDefault("accessserver.calltimeout", 120*time.Second)

	c.viper.BindEnv("etcdCluster.endpoints", c.envName("ETCD_ENDPOINTS"))
	if !c.viper.IsSet("etcdCluster.endpoints") {
		return errors.New("config check, missing 'etcdCluster.endpoints'")
	}

	c.viper.BindEnv("etcdCluster.tls.certPassword", c.envName("ETCD_TLS_CERT_PASSWORD"))
	c.viper.SetDefault("etcdCluster.tls.certPassword", "")

	c.viper.BindEnv("etcdCluster.tls.cafile", c.envName("ETCD_TLS_CAFILE"))
	c.viper.SetDefault("etcdCluster.tls.cafile", "")

	c.viper.BindEnv("etcdCluster.tls.certfile", c.envName("ETCD_TLS_CERTFILE"))
	c.viper.SetDefault("etcdCluster.tls.certfile", "")

	c.viper.BindEnv("etcdCluster.tls.keyfile", c.envName("ETCD_TLS_KEYFILE"))
	c.viper.SetDefault("etcdCluster.tls.keyfile", "")

	c.viper.BindEnv("etcdCluster.dialtimeout", c.envName("ETCD_DIAL_TIMEOUT"))
	c.viper.SetDefault("etcdCluster.dialtimeout", 3*time.Second)

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
