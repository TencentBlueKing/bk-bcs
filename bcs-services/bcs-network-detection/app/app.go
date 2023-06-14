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

// Package app xxx
package app

import (
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network-detection/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network-detection/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network-detection/networkdetection"
)

// Run app entrypoint
func Run(op *options.Option) error {
	conf := &config.Config{}
	setConfig(conf, op)
	// pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	controller := networkdetection.NewNetworkDetection(conf)
	err := controller.Start()
	if err != nil {
		blog.Errorf("NetworkDetection start failed: %s", err.Error())
		os.Exit(1)
	}
	blog.Info("NetworkDetection start working ... ")
	return nil
}

func setConfig(conf *config.Config, op *options.Option) {
	conf.Address = op.Address
	conf.Port = op.Port
	conf.Clusters = op.Clusters
	conf.BcsZk = op.BCSZk
	conf.AppId = op.AppId
	conf.EsbUrl = op.EsbUrl
	conf.Operator = op.Operator
	conf.AppSecret = op.AppSecret
	conf.AppCode = op.AppCode
	conf.Template = op.Template
	conf.ServerCert = &types.CertConfig{
		CertPasswd: static.ServerCertPwd,
	}
	conf.ClientCert = &types.CertConfig{
		CertPasswd: static.ClientCertPwd,
	}
	// server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.CAFile != "" &&
		op.CertConfig.ServerKeyFile != "" {

		conf.ServerCert.CertFile = op.CertConfig.ServerCertFile
		conf.ServerCert.KeyFile = op.CertConfig.ServerKeyFile
		conf.ServerCert.CAFile = op.CertConfig.CAFile
		conf.ServerCert.IsSSL = true
	}

	// client cert directory
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.CAFile != "" &&
		op.CertConfig.ClientKeyFile != "" {

		conf.ClientCert.CertFile = op.CertConfig.ClientCertFile
		conf.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		conf.ClientCert.CAFile = op.CertConfig.CAFile
		conf.ClientCert.IsSSL = true
	}
}
