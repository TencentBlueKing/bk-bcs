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

package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

const (
	// ServiceRegistryKubernetes service discovery for k8s
	ServiceRegistryKubernetes = "kubernetes"
	// ServiceRegistryMesos service discovery for mesos
	ServiceRegistryMesos = "mesos"

	// WorkModeGlobal define the global work mode
	WorkModeGlobal = "global"
	// WorkModePod define the pod work mode
	WorkModePod = "pod"
)

// NetworkPolicyOption the option of bcs network policy controller
type NetworkPolicyOption struct {
	conf.ServiceConfig
	conf.ServerOnlyCertConfig
	conf.FileConfig
	conf.MetricConfig
	conf.LogConfig
	conf.ProcessConfig

	ServiceRegistry      string `json:"serviceRegistry" value:"kubernetes" usage:"registry for service discovery; [kubernetes, mesos]"`
	KubeMaster           string `json:"kubeMaster" value:"" usage:"kube-apiserver url"`
	Kubeconfig           string `json:"kubeconfig" value:"" usage:"kubeconfig for kube-apiserver, Only required if out-of-cluster."`
	KubeReSyncPeriod     uint   `json:"kubeResyncPeried" value:"300" usage:"resync interval for informer factory in seconds; (default 300)"`
	KubeCacheSyncTimeout uint   `json:"kubeCacheSyncTimeout" value:"10" usage:"wait for kube cache sync timeout in seconds; (default 10)"`
	IPTableSyncPeriod    uint   `json:"iptablesSyncPeriod" value:"300" usage:"interval for sync iptables rules in seconds; (default 300)"`
	NetworkInterface     string `json:"iface" value:"eth1" usage:"network interface to get ip"`
	WorkMode             string `json:"workMode" value:"global" usage:"workmode for controller, available [global]/[pod]"`
	DockerSock           string `json:"dockerSock" value:"unix:///var/run/docker.sock" usage:"docker socket file"`
	Debug                bool   `json:"debug" value:"false" usage:"open pprof"`
}

// New new NetworkPolicyOption object
func New() *NetworkPolicyOption {
	return &NetworkPolicyOption{}
}

// Parse parse options
func Parse(opt *NetworkPolicyOption) {
	conf.Parse(opt)

	// validation config
	if opt.ServiceRegistry != ServiceRegistryKubernetes && opt.ServiceRegistry != ServiceRegistryMesos {
		blog.Fatal("registry for service discovery, available values [kubernetes, mesos]")
	}
	if len(opt.Kubeconfig) == 0 {
		blog.Fatal("kubeconfig cannot be empty")
	}
}
