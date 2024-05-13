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

// NetAgentOption option for net agent
type NetAgentOption struct {
	conf.ServiceConfig
	conf.MetricConfig
	conf.LogConfig
	conf.FileConfig

	// nolint
	Cluster string `json:"cluster" value:"" usage:"cluster for bcs"`

	// nolint
	Kubeconfig string `json:"kubeconfig" value:"" usage:"kubeconfig for kube-apiserver, Only required if out-of-cluster."` // nolint

	// nolint
	KubeResyncPeriod int `json:"kube_resync_peried" value:"1200" usage:"resync interval for informer factory in seconds; (default 300)"` // nolint

	// nolint
	KubeCacheSyncTimeout int `json:"kube_cachesync_timeout" value:"10" usage:"wait for kube cache sync timeout in seconds; (default 10)"` // nolint

	// nolint
	FixedIPWorkloads string `json:"fixed_ip_workloads" value:"StatefulSet,GameStatefulSet" usage:"names of workloads that support fixed ip, split by comma, default[StatefulSet,GameStatefulSet]"` // nolint

	// nolint
	CloudNetserviceEndpoints string `json:"cloud_netservice_endpoints" value:"" usage:"cloud netservice endpoints, split by comma"` // nolint

	// nolint
	EniMTU int `json:"eni_mtu" value:"1500" usage:"the mtu of eni"`

	// nolint
	Ifaces string `json:"ifaces" value:"eth1" usage:"use ip of these network interfaces as node identity, split with comma or semicolon"` // nolint

	// nolint
	DockerSock string `json:"docker_sock" value:"unix:///var/run/docker.sock" usage:"unix socket file path for docker"` // nolint

	// nolint
	ReconcileInterval int `json:"reconcile_interval" value:"7200" usage:"reconcile interval for ready NIC in seconds; (default 7200)"` // nolint

	// nolint
	UseDevicePlugin bool `json:"use_device_plugin" value:"true" usage:"whether use device plugin server"`

	// nolint
	KubeletSockPath string `json:"kubelet_sock_path" value:"/var/lib/kubelet/device-plugins/kubelet.sock" usage:"kubelet sock file path for device plugin to register"` // nolint

	// nolint
	DevicePluginSockPath string `json:"device_plugin_sock_path" value:"/var/lib/kubelet/device-plugins/bcs-cloud-netagent.sock" usage:"device plugin unix socket file of bcs cloud agent"` // nolint

	// nolint
	DevicePluginResourceName string `json:"device_plugin_resource_name" value:"cloud.bkbcs.tencent.com/eip" usage:"device plugin resource name"` // nolint
}

// New new option
func New() *NetAgentOption {
	return &NetAgentOption{}
}

// Parse parse options
func Parse(opt *NetAgentOption) {
	conf.Parse(opt)

	if len(opt.Cluster) == 0 {
		blog.Fatal("cluster cannot be empty")
	}
	if len(opt.CloudNetserviceEndpoints) == 0 {
		blog.Fatal("cloud netservice endpoints cannot be empty")
	}

	blog.Infof("get option %+v", opt)
}
