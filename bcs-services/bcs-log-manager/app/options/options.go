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
	logmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
)

//LogManagerOption is option in flags
type LogManagerOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	CollectionConfigs []logmanager.CollectionConfig `json:"collection_configs" usage:"Custom configs of log collections"`
	BKDataAPIHost     string                        `json:"bkdata_api_host" value:"" usage:"bk-data api gateway host"`
	BcsAPIHost        string                        `json:"bcs_api_host" value:"" usage:"BcsApi Host"`

	EtcdHosts    string `json:"etcd_hosts" value:"" usage:"etcd host"`
	EtcdCertFile string `json:"etcd_cert_file" value:"" usage:"etcd cert file"`
	EtcdKeyFile  string `json:"etcd_key_file" value:"" usage:"etcd key file"`
	EtcdCAFile   string `json:"etcd_ca_file" value:"" usage:"etcd ca file"`

	AuthToken    string `json:"api_auth_token" value:"" usage:"BcsApi authentication token"`
	Gateway      bool   `json:"use_gateway" value:"true" usage:"whether use api gateway"`
	KubeConfig   string `json:"kubeconfig" value:"" usage:"k8s config file path"`
	SystemDataID string `json:"system_dataid" value:"" usage:"DataID used to upload logs of k8s and bcs system modules with standard output"`
	BkUsername   string `json:"bk_username" value:"" usage:"User to request bkdata api"`
	BkAppCode    string `json:"bk_appcode" value:"" usage:"BK app code"`
	BkAppSecret  string `json:"bk_appsecret" value:"" usage:"BK app secret"`
	BkBizID      int    `json:"bk_bizid" value:"-1" usage:"BK business id"`
}

// NewLogManagerOption create new manager operation object
func NewLogManagerOption() *LogManagerOption {
	return &LogManagerOption{}
}
