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
	// CloudAWS aws cloud service
	CloudAWS = "aws"
	// CloudTencent tencent cloud service
	CloudTencent = "qcloud"
)

// NetworkOption the option of bcs elastic network interface controller
type NetworkOption struct {
	conf.ServiceConfig
	conf.ServerOnlyCertConfig
	conf.FileConfig
	conf.MetricConfig
	conf.LogConfig
	conf.ProcessConfig

	Cluster              string `json:"cluster" value:"" usage:"cluster for bcs"`
	Cloud                string `json:"cloud" value:"" usage:"name of cloud service, [aws, qcloud]"`
	Kubeconfig           string `json:"kubeconfig" value:"" usage:"kubeconfig for kube-apiserver, Only required if out-of-cluster."`
	KubeResyncPeriod     int    `json:"kubeResyncPeried" value:"300" usage:"resync interval for informer factory in seconds; (default 300)"`
	KubeCacheSyncTimeout int    `json:"kubeCacheSyncTimeout" value:"10" usage:"wait for kube cache sync timeout in seconds; (default 10)"`
	CheckInterval        int    `json:"checkInterval" value:"300" usage:"interval for checking ip rules and route tables"`

	NetServiceZookeeper string `json:"netserviceZookeeper" value:"" usage:"zookeeper to discovery netservice"`
	NetServiceCa        string `json:"netserviceCa" value:"" usage:"ca for netservice"`
	NetServiceKey       string `json:"netserviceKey" value:"" usage:"key for netservice"`
	NetServiceCert      string `json:"netserviceCert" value:"" usage:"cert for netservice"`

	EniNum      int    `json:"eniNum" value:"0" usage:"the number of elastic network interface for each node; default is 0, means apply for as many eni as possible"`
	IPNumPerEni int    `json:"ipNumPerEni" value:"0" usage:"the number of ip for each eni; default is 0, means apply for as many ip as possible"`
	EniMTU      int    `json:"eniMTU" value:"1500" usage:"the mtu of eni"`
	Ifaces      string `json:"ifaces" value:"eth1" usage:"use ip of these network interfaces as node identity, split with comma or semicolon"`
}

// New new option
func New() *NetworkOption {
	return &NetworkOption{}
}

// Parse parse options
func Parse(opt *NetworkOption) {
	conf.Parse(opt)

	if len(opt.Cloud) == 0 {
		blog.Fatal("cloud cannot be empty")
	}
	if len(opt.Kubeconfig) == 0 {
		blog.Fatal("kubeconfig cannot be empty")
	}
	if len(opt.NetServiceZookeeper) == 0 {
		blog.Fatal("netservice zookeeper cannot be empty")
	}
	if opt.EniMTU < 68 || opt.EniMTU > 65535 {
		blog.Fatal("invalid eni mtu")
	}

	blog.Infof("get option %+v", opt)
}
