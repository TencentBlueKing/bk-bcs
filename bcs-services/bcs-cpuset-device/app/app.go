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
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cpuset-device/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cpuset-device/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cpuset-device/cpuset-device"
)

// Run run the server
func Run(op *options.Option) error {

	conf := config.NewConfig()
	setConfig(conf, op)

	controller := cpuset_device.NewCpusetDevicePlugin(conf)
	err := controller.Start()
	if err != nil {
		blog.Errorf("CpusetDevicePlugin Start failed: %s", err.Error())
		os.Exit(1)
	}

	blog.Info("CpusetDevicePlugin server ... ")
	return nil
}

func setConfig(conf *config.Config, op *options.Option) {
	conf.DockerSocket = op.DockerSock
	conf.PluginSocketDir = op.PluginSocketDir
	conf.BcsZk = op.BCSZk
	conf.Engine = op.Engine
	conf.ClusterID = op.ClusterID
	conf.NodeIP = op.Address

	// client cert directoty
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.CAFile != "" &&
		op.CertConfig.ClientKeyFile != "" {

		conf.ClientCert.CertFile = op.CertConfig.ClientCertFile
		conf.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		conf.ClientCert.CAFile = op.CertConfig.CAFile
		conf.ClientCert.IsSSL = true
		conf.ClientCert.CertPasswd = static.ClientCertPwd
	}

	conf.ReservedCPUSet = make(map[string]struct{})
	// parse reserved cpuset list
	if len(op.ReservedCPUSetList) != 0 {
		cpuSetStrList := strings.Split(op.ReservedCPUSetList, ",")
		for _, cpuSetStr := range cpuSetStrList {
			conf.ReservedCPUSet[strings.TrimSpace(cpuSetStr)] = struct{}{}
		}
	}

}
