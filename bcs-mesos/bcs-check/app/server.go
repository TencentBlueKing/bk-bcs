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
	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/bcs-check/app/options"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck"
)

//Run the health check
func Run(op *options.HealthCheckOption) error {

	setConfig(op)

	server := bcscheck.NewHealthCheckServer(op.Conf)

	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	blog.Info("app begin to start health check server ... ")

	server.Run()

	return nil
}

func setConfig(op *options.HealthCheckOption) {
	blog.Infof("op mesos zk %s bcs zk %s", op.MesosZK, op.BCSZk)

	op.Conf.RegDiscvSvr = op.MesosZK
	op.Conf.SchedDiscvSvr = op.MesosZK
	op.Conf.BcsDiscvSvr = op.BCSZk
	op.Conf.Address = op.Address
	op.Conf.MetricPort = op.MetricPort

	op.Conf.Cluster = op.Cluster

	//server cert directoty
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.CAFile != "" &&
		op.CertConfig.ServerKeyFile != "" {

		op.Conf.ServCert.CertFile = op.CertConfig.ServerCertFile
		op.Conf.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		op.Conf.ServCert.CAFile = op.CertConfig.CAFile
		op.Conf.ServCert.IsSSL = true
	}

	//client cert directoty
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.CAFile != "" &&
		op.CertConfig.ClientKeyFile != "" {

		op.Conf.ClientCert.CertFile = op.CertConfig.ClientCertFile
		op.Conf.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		op.Conf.ClientCert.CAFile = op.CertConfig.CAFile
		op.Conf.ClientCert.IsSSL = true
	}
}
