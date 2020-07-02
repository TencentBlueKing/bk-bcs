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
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/config"
)

//HpaControllerOption is option in flags
type HpaControllerOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	ClusterZkAddr string `json:"cluster_zookeeper" value:"" usage:"bcs mesos cluster zk address"`
	KubeConfig    string `json:"kubeconfig" value:"" usage:"kubeconfig for kube-apiserver"`
	CadvisorPort  int    `json:"cadvisor_port" value:"" usage:"container cadvisor port"`
	ClusterID     string `json:"clusterid" value:"" usage:"bcs mesos cluster id"`

	Conf *config.Config
}

//NewHpaControllerOption create HpaControllerOption object
func NewHpaControllerOption() *HpaControllerOption {
	return &HpaControllerOption{
		Conf: config.NewConfig(),
	}
}
