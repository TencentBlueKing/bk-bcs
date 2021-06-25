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
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing"
)

//CertConfig is configuration of Cert
type CertConfig struct {
	CAFile   string
	CertFile string
	KeyFile  string
	CertPwd  string
	IsSSL    bool
}

//StorageOptions is options in flags
type StorageOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.ServerOnlyCertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	ServerCert *CertConfig
	Etcd       registry.CMDOptions `json:"etcdRegistry"`
	Tracing    tracing.Options     `json:"tracing"`

	DBConfig     string `json:"database_config_file" value:"storage-database.conf" usage:"Config file for database."`
	QueueConfig  string `json:"queue_config_file" value:"queue.conf" usage:"Config file for database."`
	EventMaxTime int64  `json:"event_max_day" value:"15" usage:"Max day for holding events data."`
	EventMaxCap  int64  `json:"event_max_cap" value:"10000" usage:"Max num for holding events data of each cluster."`
	AlarmMaxTime int64  `json:"alarm_max_day" value:"15" usage:"Max day for holding alarms data."`
	AlarmMaxCap  int64  `json:"alarm_max_cap" value:"10000" usage:"Max num for holding alarms data of each cluster."`
	QueryMaxNum  int64  `json:"query_max_num" value:"100" usage:"Max num query to same url one time."`
	WatchTimeSep int64  `json:"watch_time_sep" value:"10" usage:"Request watch time sep."`
	PrintBody    bool   `json:"print_body" value:"false" usage:"Print body every request."`
	PrintManager bool   `json:"print_manager" value:"false" usage:"Print manager."`
	DebugMode    bool   `json:"debug_mode" value:"false" usage:"Debug mode, use pprof."`
}

//NewStorageOptions create StorageOptions object
func NewStorageOptions() *StorageOptions {
	return &StorageOptions{
		ServerCert: &CertConfig{
			CertPwd: static.ServerCertPwd,
			IsSSL:   false,
		},
		Etcd: registry.CMDOptions{},
	}
}
