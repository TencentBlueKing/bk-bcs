/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

//ConsoleManagerOption is option in flags
type ConsoleManagerOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	WebConsoleImage       string `json:"web-console-image" value:"ccr.ccs.tencentyun.com/bk-cmdb-lf/bcs-webconsole:v0.1" usage:"web-console images url"`
	KubeConfigFile        string `json:"kubeconfig" value:"" usage:"Path to kubeconfig file with authorization and master location information."`
	RedisAddress          string `json:"redis-address" value:"127.0.0.1:6379" usage:"Redis Server Address"`
	RedisPassword         string `json:"redis-password" value:"" usage:"Redis Password"`
	RedisDatabase         string `json:"redis-database" value:"0" usage:"Redis DB"`
	RedisMasterName       string `json:"redis-master-name" value:"" usage:"The master name."`
	RedisSentinelPassword string `json:"redis-sentinel-password" value:"3000" usage:"A seed list of host:port
addresses of
sentinel nodes."`
	RedisPoolSize int `json:"redis-poolSize" value:"" usage:"Redis Pool Size"`

	Conf  config.ConsoleConfig
	Redis RedisConfig
}

// RedisConfig define redis config
type RedisConfig struct {
	Address          string
	Password         string
	Database         string
	MasterName       string
	SentinelPassword string
	PoolSize         int
}

// ServerConfig option for server
type ServerConfig struct {
	Address         string `json:"address"`
	InsecureAddress string `json:"insecureaddress"`
	Port            uint   `json:"port"`
	HTTPPort        uint   `json:"httpport"`
	MetricPort      uint   `json:"metricport"`
	ServerCert      string `json:"servercert"`
	ServerKey       string `json:"serverkey"`
	ServerCa        string `json:"serverca"`
}

// ClientConfig option for bcs-cluster-manager as client
type ClientConfig struct {
	ClientCert string `json:"clientcert"`
	ClientKey  string `json:"clientkey"`
	ClientCa   string `json:"clientca"`
}

//NewConsoleOption create ConsoleManagerOption object
func NewConsoleOption() *ConsoleManagerOption {
	return &ConsoleManagerOption{
		Conf: config.NewConsoleConfig(),
	}
}
