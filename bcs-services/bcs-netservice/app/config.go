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

package app

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

//Backend for mysql
type Backend struct {
	Host   string `json:"backend_host" value:"" usage:"netservice data storage backend ip address"`
	User   string `json:"backend_user" value:"" usage:"netservice data storage backend user info"`
	Passwd string `json:"backend_passwd" value:"" usage:"netservice data storage backend password"`
}

//NewConfig creat new Config for net-server
func NewConfig() *Config {
	cfg := new(Config)
	return cfg
}

//Config for bcs-netservice in conf/bcs.conf
type Config struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.ServerOnlyCertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig
}
