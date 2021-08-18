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

import "github.com/Tencent/bk-bcs/bcs-common/common/types"

// Config parsed config for bcs-cpuset-device
type Config struct {
	// PluginSocketDir device plugin socket dir, examples: /var/lib/kubelet/device-plugins
	PluginSocketDir string
	// DockerSocket docker socket
	DockerSocket string
	// CgroupCpusetRoot root path of cpuset cgroup
	CgroupCpusetRoot string
	// ClientCert client https certs
	ClientCert *types.CertConfig `json:"-"`
	// BcsZk cluster zk address
	BcsZk string
	// ClusterID clusterid
	ClusterID string
	// Engine engine type enum: k8s、mesos
	Engine string
	// NodeIP node IP
	NodeIP string
	// ReservedCPUSet set of reserved cpu set
	ReservedCPUSet map[string]struct{}
}

// NewConfig create a config object
func NewConfig() *Config {
	return &Config{
		ClientCert: &types.CertConfig{},
	}
}
