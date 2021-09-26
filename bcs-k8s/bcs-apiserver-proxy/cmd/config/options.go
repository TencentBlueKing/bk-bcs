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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// ProxyAPIServerOptions cmd option for bcs-apiserver-proxy
type ProxyAPIServerOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.LogConfig
	conf.ProcessConfig

	DebugMode      bool               `json:"debugMode"`
	ProxyLvs       ProxyLvs           `json:"proxyLvs"`
	HealthCheck    HealthCheckOptions `json:"healthCheck"`
	K8sConfig      K8sConfig          `json:"k8sConfig"`
	SystemInterval SystemInterval     `json:"systemInterval"`
	RealServer     RealServer         `json:"realServer"`
}

// Validate check ProxyAPIServerOptions paras
func (opt ProxyAPIServerOptions) Validate() bool {
	if len(opt.ProxyLvs.VirtualAddress) == 0 {
		blog.Errorf("proxy virtual address is empty")
		return false
	}

	return true
}

// ProxyLvs virtual server
type ProxyLvs struct {
	VirtualAddress string `json:"virtualAddress" value:"127.0.0.1:6443" usage:"Proxy lvs address:port"`
}

// HealthCheckOptions health check scheme&path
type HealthCheckOptions struct {
	HealthScheme string `json:"healthScheme" usage:"health check request scheme"`
	HealthPath   string `json:"healthPath" usage:"health check path"`
}

// K8sConfig master & KubeConfig
type K8sConfig struct {
	Master     string `json:"master" usage:"kubernetes cluster master"`
	KubeConfig string `json:"kubeConfig" value:"" usage:""`
}

// SystemInterval ticker interval
type SystemInterval struct {
	EndpointInterval int64 `json:"endpointInterval" value:"5" usage:"dynamic update cluster endpointsIP interval"`
	ManagerInterval  int64 `json:"managerInterval" value:"10" usage:"dynamic refresh ipvs rules interval"`
}

// RealServer vs backend rs
type RealServer struct {
	RealAddress []string `json:"realServers" usage:"realServers init lvs address"`
}

// NewProxyAPIServerOptions init ProxyAPIServerOptions
func NewProxyAPIServerOptions() *ProxyAPIServerOptions {
	return &ProxyAPIServerOptions{}
}
