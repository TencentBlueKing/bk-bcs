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
	"bk-bcs/bcs-common/common/conf"
)

//Option is option in flags
type Option struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	Clusters  string `json:"clusters" value:"" usage:"deploy detection node clusters; example: BCS-MESOS-10000,BCS-MESOS-10001,BCS-MESOS-10002..."`
	AppCode   string `json:"app_code" value:"" usage:"esb app_code"`
	AppSecret string `json:"app_secret" value:"" usage:"esb app_secret"`
	Operator  string `json:"operator" value:"" usage:"esb operator"`
	EsbUrl    string `json:"esb_url" value:"" usage:"esb url"`
	AppId     int    `json:"app_id" value:"" usage:"cmdb app id"`
	Template  string `json:"template" value:"./template/deployment.json" usage:"deployment template json file path"`
}

//NewOption create Option object
func NewOption() *Option {
	return &Option{}
}
