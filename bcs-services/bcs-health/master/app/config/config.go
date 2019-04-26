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

package config

import (
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-services/bcs-health/util"
)

type Config struct {
	conf.FileConfig
	conf.CertConfig
	conf.ServiceConfig
	conf.LocalConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	Receivers
	KafkaConf          KafkaConf `json:"kafka_conf"`
	ETCD               EtcdConf  `json:"etcd"`
	HttpCheckList      string    `json:"http_check_list" value:"" usage:"bcs-health http check url list, comma separated."`
	EnableBsAlarm      bool      `json:"enable_bs_alarm" value:"false" usage:"enable blue shield alarm function"`
	EnableLogAlarm     bool      `json:"enable_log_alarm" value:"false" usage:"enable log alarm, all alarms will write to log"`
	EnableStorageAlarm bool      `json:"enable_storage_alarm" value:"false" usage:"enable storage alarm, all alarms will write to bcs-storage"`
	Silence            bool      `json:"silence" value:"false" usage:"silence all the alarm for test usage. *Attention*: only used for test."`
}

type Receivers struct {
	BcsReceivers      string `json:"bcs_receivers" value:"" usage:"bcs's alarm default receivers, comma separated."`
	LBReceivers       string `json:"lb_receivers" value:"" usage:"lb endpoints alarm receivers, comma separated."`
	KubeReceivers     string `json:"kube_receivers" value:"" usage:"kube endpoints alarm receivers, comma separated."`
	EndpintsReceivers string `json:"endpoints_receivers" value:"" usage:"bcs endpoints alarm receivers, comma separated."`
	AppReceivers      string `json:"app_receivers" value:"" usage:"bcs app alarm receivers, format like appid1:zhangsan,lisi;appid2:wangwu"`
}

type EtcdConf struct {
	EtcdEndpoints string `json:"etcd_endpoints" value:"" usage:"etcd cluster endpoints addr, comma separated."`
	EtcdRootPath  string `json:"etcd_root_path" value:"/bcshealth" usage:"etcd root path value."`
	CaFile        string `json:"etcd_ca_file" value:"" usage:"etcd ca file path."`
	CertFile      string `json:"etcd_cert_file" value:"" usage:"etcd cert file path."`
	KeyFile       string `json:"etcd_key_file" value:"" usage:"etcd key file path."`
	PassWord      string `json:"etcd_key_password" value:"" usage:"etcd key file decrypt password."`
}

type BcsConfig struct {
	BcsZkAddr string      `json:"bcsZkAddr"`
	Server    util.Server `json:"server"`
	ClientTLS util.TLS    `json:"clientTLS"`
}

type KafkaConf struct {
	DataID     string `json:"data_id" value:"9748" usage:"the data id that used to send event data to dataplatform"`
	PluginPath string `json:"plugin_path" value:"" usage:"the path of plugin binary"`
	ConfigFile string `json:"config_file" value:"" usage:"the config file path of kafka binary"`
}

func ParseConfig() Config {
	c := new(Config)
	conf.Parse(c)
	return *c
}
