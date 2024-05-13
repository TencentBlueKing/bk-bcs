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
 */

// Package option is options for cloud netservice
package option

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// NewConfig create new config object
func NewConfig() *Config {
	cfg := new(Config)
	return cfg
}

// Config config for bcs cloud netservice
type Config struct {
	// Kubeconfig kubeconfig for kubernetes store
	Kubeconfig string `json:"kubeconfig" value:"" usage:"kubeconfig for kubernetes apiserver"`

	// Debug debug flag
	Debug bool `json:"debug" value:"false" usage:"debug flag, open pprof"`

	// SwaggerDir
	SwaggerDir string `json:"swagger_dir" value:"" usage:"swagger dir"`

	// CloudMode cloud mod
	CloudMode string `json:"cloud_mode" value:"" usage:"cloud mode, option [tencentcloud, aws]"`

	// IPMaxIdleMinute ip max idle time, max time for available ip before return to cloud
	IPMaxIdleMinute int `json:"ip_max_idle_minute" value:"1600" usage:"max time before return to cloud; unit[minute]"`
	// IPCleanIntervalMinute ip clean interval
	IPCleanIntervalMinute int `json:"ip_clean_interval_minute" value:"10" usage:"minute for ip cleaner check interval"`
	// FixedIPCleanIntervalMinute fixed clean interval
	FixedIPCleanIntervalMinute int `json:"fixed_ip_clean_interval_minute" value:"20" usage:"interval minute for ip cleaner check fixed ip"` // nolint

	// EtcdEndpoints endpoints of etcd
	EtcdEndpoints string `json:"etcd_endpoints" value:"" usage:"endpoints of etcd"`
	// EtcdCert cert file path of etcd
	EtcdCert string `json:"etcd_cert" value:"" usage:"cert file of etcd"`
	// EtcdKey key file path of etcd
	EtcdKey string `json:"etcd_key" value:"" usage:"key file of etcd"`
	// EtcdCa ca file path of etcd
	EtcdCa string `json:"etcd_ca" value:"" usage:"ca file of etcd"`

	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.LogConfig
}
