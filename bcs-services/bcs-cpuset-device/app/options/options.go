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

// Option is option in flags
type Option struct {
	conf.FileConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ZkConfig
	conf.CertConfig
	conf.ServiceConfig

	DockerSock      string `json:"docker_sock" value:"unix:///var/run/docker.sock" usage:"docker socket file"`
	PluginSocketDir string `json:"plugin_socket_dir" value:"/var/lib/kubelet/device-plugins" usage:"device-plugin socket directory"`
	ClusterID       string `json:"clusterid" value:"" usage:"mesos cluster id"`
	Engine          string `json:"engine" value:"k8s" usage:"enum: k8s、mesos; default: k8s"`
}

// NewOption create Option object
func NewOption() *Option {
	return &Option{}
}
