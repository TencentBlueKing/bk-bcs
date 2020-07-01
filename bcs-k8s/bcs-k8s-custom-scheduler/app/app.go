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
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/app/custom-scheduler"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/options"
	"os"
)

//Run the ipscheduler
func Run(op *options.ServerOption) {

	conf := parseConfig(op)

	customSched := custom_scheduler.NewCustomScheduler(conf)
	//start customSched, and http service
	err := customSched.Start()
	if err != nil {
		blog.Errorf("start processor error %s, and exit", err.Error())
		os.Exit(1)
	}

	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	return
}

func parseConfig(op *options.ServerOption) *config.IpschedulerConfig {
	ipschedulerConfig := config.NewIpschedulerConfig()

	ipschedulerConfig.Address = op.Address
	ipschedulerConfig.Port = op.Port
	ipschedulerConfig.InsecureAddress = op.InsecureAddress
	ipschedulerConfig.InsecurePort = op.InsecurePort
	ipschedulerConfig.ZkHosts = op.BCSZk
	ipschedulerConfig.VerifyClientTLS = op.VerifyClientTLS

	config.ZkHosts = op.BCSZk
	config.Cluster = op.Cluster
	config.Kubeconfig = op.Kubeconfig
	config.KubeMaster = op.KubeMaster
	config.UpdatePeriod = op.UpdatePeriod

	//server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.ServerKeyFile != "" {
		ipschedulerConfig.ServCert.CertFile = op.CertConfig.ServerCertFile
		ipschedulerConfig.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		ipschedulerConfig.ServCert.CAFile = op.CertConfig.CAFile
		ipschedulerConfig.ServCert.IsSSL = true
	}

	//client cert directory
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.ClientKeyFile != "" {
		ipschedulerConfig.ClientCert.CertFile = op.CertConfig.ClientCertFile
		ipschedulerConfig.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		ipschedulerConfig.ClientCert.CAFile = op.CertConfig.CAFile
		ipschedulerConfig.ClientCert.IsSSL = true
	}

	config.ClientCert = ipschedulerConfig.ClientCert

	return ipschedulerConfig
}
