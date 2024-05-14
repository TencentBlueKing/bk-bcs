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
	InCluster bool
	ClusterID string
	BizID     string
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
		"Set the kubeconfig path to load, kubeconfig file should have suffix of config")

	// incluster选项
	fs.BoolVarP(&bcro.InCluster, "inCluster", "", false, "Set true the reporter will work as in-cluster mode")
	fs.StringVarP(&bcro.ClusterID, "clusterID", "", "0", "Set clusterID")
	fs.StringVarP(&bcro.BizID, "bizID", "", "incluster", "Set cluster bizID")
}
