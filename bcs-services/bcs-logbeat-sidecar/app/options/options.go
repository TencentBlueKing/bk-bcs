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
)

//SidecarOption is option in flags
type SidecarOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	DockerSock   string `json:"docker_sock" value:"unix:///var/run/docker.sock" usage:"docker socket file"`
	LogbeatDir   string `json:"logbeat_dir" value:"" usage:"logbeat config directory"`
	TemplateFile string `json:"template_file" value:"./unifytlogc-template.conf" usage:"logbeat tempalte config file"`
	PrefixFile   string `json:"prefix_file" value:"" usage:"logbeat config file prefix name"`
	Kubeconfig   string `json:"kubeconfig" value:"" usage:"kubeconfig"`
}

//NewSidecarOption create SidecarOption object
func NewSidecarOption() *SidecarOption {
	return &SidecarOption{}
}
