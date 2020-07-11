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

package app

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// Config detail configuration item
type Config struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.CertConfig
	conf.ZkConfig
	conf.LogConfig
	conf.ProcessConfig
	Scheme     string `json:"metric_scheme" value:"http" usage:"scheme for metric api"`
	Zookeeper  string `json:"zookeeper" value:"127.0.0.1:3181" usage:"data source for taskgroups and services"`
	Cluster    string `json:"cluster" value:"" usage:"cluster id or name"`
	KubeConfig string `json:"kubeconfig" value:"kubeconfig" usage:"configuration file for kube-apiserver"`
}

// Validate validate command line parameter
func (c *Config) Validate() error {
	if len(c.Cluster) == 0 {
		return fmt.Errorf("cluster cannot be empty")
	}
	return nil
}
