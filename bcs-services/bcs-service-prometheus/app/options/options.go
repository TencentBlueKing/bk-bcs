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

//PrometheusControllerOption is option in flags
type PrometheusControllerOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	ClusterZk            string   `json:"cluster_zookeeper" value:"" usage:"mesos cluster zookeeper"`
	CadvisorPort         int      `json:"cadvisor_port" value:"" usage:"node cadvisor port"`
	NodeExporterPort     int      `json:"node_exporter_port" value:"" usage:"node exporter port"`
	ClusterID            string   `json:"clusterid" value:"" usage:"mesos clusterid"`
	PromFilePrefix       string   `json:"prom_file_prefix" value:"" usage:"prometheus service discovery file prefix"`
	EnableMesos          bool     `json:"enable_mesos" value:"true" usage:"enable mesos prometheus service discovery"`
	EnableService        bool     `json:"enable_service" value:"true" usage:"enable service prometheus service discovery"`
	EnableNode           bool     `json:"enable_node" value:"true" usage:"enable node prometheus service discovery"`
	EnableServiceMonitor bool     `json:"enable_service_monitor" value:"true" usage:"enable service monitor discovery"`
	Kubeconfig           string   `json:"kubeconfig" value:"" usage:"kubernetes kubeconfig"`
	ServiceModules       []string `json:"service_modules" value:"" usage:"service module list"`
	ClusterModules       []string `json:"cluster_modules" value:"" usage:"cluster module list"`
}

//NewPrometheusControllerOption create PrometheusControllerOption object
func NewPrometheusControllerOption() *PrometheusControllerOption {
	return &PrometheusControllerOption{}
}
