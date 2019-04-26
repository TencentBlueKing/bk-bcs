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
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/config"
	//"github.com/spf13/pflag"
)

//HealthCheckOption is option in flags
type HealthCheckOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	MesosZK string `json:"mesos_zookeeper" value:"" usage:"the address to register and discover scheduler"`

	Cluster string `json:"cluster" value:"" usage:"the cluster ID under bcs"`

	Conf config.HealthCheckConfig
}

//NewHealthCheckOption create HealthCheckOption object
func NewHealthCheckOption() *HealthCheckOption {
	return &HealthCheckOption{
		Conf: config.NewHealthCheckConfig(),
	}
}
