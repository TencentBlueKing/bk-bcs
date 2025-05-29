/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package options xxx
package options

import (
	"github.com/spf13/pflag"
)

// BcsClusterReporterOptions component options
type BcsClusterReporterOptions struct {
	// 需要开启的plugin
	Plugins string
	// plugin配置文件路径
	PluginConfDir string

	// 需要巡检的集群类型
	BcsClusterType string
	// 需要巡检的集群ID列表
	BcsClusterList []string

	// 集群配置来源
	// 如果从bcsClusterManager获取kubeconfig则需要配置
	BcsClusterManagerToken     string
	BcsClusterManagerApiserver string
	BcsGatewayToken            string
	BcsGatewayApiserver        string

	// 也可以单独配置kubeconfig
	KubeConfigDir string
	// 也可以配置incluster模式
	InCluster     bool
	ClusterID     string
	BizID         string
	RunMode       string
	LabelSelector string
}

// NewBcsClusterReporterOptions init options
func NewBcsClusterReporterOptions() *BcsClusterReporterOptions {
	return &BcsClusterReporterOptions{}
}

// AddFlags flags
func (bcro *BcsClusterReporterOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&bcro.Plugins, "plugins", "", bcro.Plugins,
		"Set the plugins to use, use ',' as separator. Available plugins:1.masterpodcheck")
	fs.StringVarP(&bcro.PluginConfDir, "pluginConfDir", "", "/data/bcs/plugins",
		"Set the pluginDir to get plugin config fie, default: /data/bcs/plugins")

	// bcs cluster 配置
	fs.StringVarP(&bcro.BcsClusterManagerToken, "bcsClusterManagerToken", "", "",
		"Set the pluginDir to get plugin config fie, default: /data/bcs/plugins")
	fs.StringVarP(&bcro.BcsClusterManagerApiserver, "bcsClusterManagerApiserver", "", "", "Set the bcs clustermanager api")
	fs.StringVarP(&bcro.BcsGatewayToken, "bcsGatewayToken", "", "", "Set the bcs apigateway token")
	fs.StringVarP(&bcro.BcsGatewayApiserver, "bcsGatewayApiserver", "", "", "Set the bcs apigateway url")

	fs.StringVarP(&bcro.BcsClusterType, "bcsClusterType", "", "",
		"Set the clusters type to check and report: debug, prod, all")
	fs.StringSliceVar(&bcro.BcsClusterList, "bcsClusterList", []string{}, "Set the clusters id to check and report")

	fs.StringVarP(&bcro.KubeConfigDir, "kubeConfigDir", "", "",
		"Set the kubeconfig path to load kubeconfig files, and the kubeconfig files’ name should end with \"config\"")

	// incluster选项
	fs.BoolVarP(&bcro.InCluster, "inCluster", "", false, "Set true the reporter will work as in-cluster mode")
	fs.StringVarP(&bcro.ClusterID, "clusterID", "", "0", "Set clusterID")
	fs.StringVarP(&bcro.BizID, "bizID", "", "incluster", "Set cluster bizID")
	fs.StringVarP(&bcro.LabelSelector, "labelSelector", "", "", "Label to select clusters")
	fs.StringVar(&bcro.RunMode, "runMode", "daemon", "daemon, once")
}

// NodeAgentOptions component options
type NodeAgentOptions struct {
	HostPath       string
	Upstream       string
	ConfigPath     string
	Plugins        string
	RunMode        string
	PluginDir      string
	CMNamespace    string
	Addr           string
	KubeConfigPath string
}

// NewNodeAgentOptions return NodeAgentOptions
func NewNodeAgentOptions() *NodeAgentOptions {
	return &NodeAgentOptions{}
}

// AddFlags xxx
func (brro *NodeAgentOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&brro.Plugins, "plugins", "dnscheck,configfilecheck,hwcheck,processcheck,diskcheck,netcheck,timecheck,nodeinfocheck,containercheck,uploader", "plugins")
	fs.StringVar(&brro.KubeConfigPath, "kubeconfigPath", "/root/.kube/config", "if incluster failed, use this kubeconfig path")

	fs.StringVar(&brro.HostPath, "hostPath", "/", "set here or set HOST_PATH env")
	fs.StringVar(&brro.Upstream, "upstream", "cluster", "cluster, mysql")
	fs.StringVar(&brro.RunMode, "runMode", "once", "daemon, once")
	fs.StringVar(&brro.PluginDir, "pluginDir", "/data/bcs/nodeagent", "/data/bcs/nodeagent")
	fs.StringVar(&brro.ConfigPath, "configPath", "/data/bcs/nodeagent/", "/data/bcs/nodeagent/")
	fs.StringVar(&brro.CMNamespace, "cmNamespace", "nodeagent", "namespace to store nodeagent checkresult configmap")
	fs.StringVar(&brro.Addr, "addr", "0.0.0.0:6216", "addr to bind listen")
}
