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
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	usermanager "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

// Run run app
func Run(op *options.UserManagerOptions) {
	conf, err := parseConfig(op)
	if err != nil {
		blog.Errorf("error parse config: %s", err.Error())
		os.Exit(1)
	}
	if conf.ClientCert.IsSSL {
		cliTLS, err := ssl.ClientTslConfVerity(conf.ClientCert.CAFile, conf.ClientCert.CertFile, conf.ClientCert.KeyFile, conf.ClientCert.CertPasswd)
		if err != nil {
			blog.Errorf("set client tls config error %s", err.Error())
		} else {
			config.CliTls = cliTLS
			blog.Infof("set client tls config success")
		}
	}
	userManager := usermanager.NewUserManager(conf)

	//start userManager, and http service
	err = userManager.Start()
	if err != nil {
		blog.Errorf("start processor error %s, and exit", err.Error())
		os.Exit(1)
	}

	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	metrics.RunMetric(conf)
}

// parseConfig parse the option to config
func parseConfig(op *options.UserManagerOptions) (*config.UserMgrConfig, error) {
	userMgrConfig := config.NewUserMgrConfig()

	userMgrConfig.Address = op.Address
	userMgrConfig.Port = op.Port
	userMgrConfig.InsecureAddress = op.InsecureAddress
	userMgrConfig.InsecurePort = op.InsecurePort
	userMgrConfig.LocalIp = op.LocalIP
	userMgrConfig.MetricPort = op.MetricPort
	userMgrConfig.BootStrapUsers = op.BootStrapUsers
	userMgrConfig.TKE = op.TKE
	userMgrConfig.PeerToken = op.PeerToken

	config.Tke = op.TKE
	secretID, err := encrypt.DesDecryptFromBase([]byte(config.Tke.SecretId))
	if err != nil {
		return nil, fmt.Errorf("error decrypting tke secretId and exit: %s", err.Error())
	}
	config.Tke.SecretId = string(secretID)
	secretKey, err := encrypt.DesDecryptFromBase([]byte(config.Tke.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("error decrypting tke secretKey and exit: %s", err.Error())
	}
	config.Tke.SecretKey = string(secretKey)

	dsn, err := encrypt.DesDecryptFromBase([]byte(op.DSN))
	if err != nil {
		return nil, fmt.Errorf("error decrypting db config and exit: %s", err.Error())
	}
	userMgrConfig.DSN = string(dsn)

	userMgrConfig.VerifyClientTLS = op.VerifyClientTLS

	//server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.ServerKeyFile != "" {
		userMgrConfig.ServCert.CertFile = op.CertConfig.ServerCertFile
		userMgrConfig.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		userMgrConfig.ServCert.CAFile = op.CertConfig.CAFile
		userMgrConfig.ServCert.IsSSL = true
	}

	//client cert directory
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.ClientKeyFile != "" {
		userMgrConfig.ClientCert.CertFile = op.CertConfig.ClientCertFile
		userMgrConfig.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		userMgrConfig.ClientCert.CAFile = op.CertConfig.CAFile
		userMgrConfig.ClientCert.IsSSL = true
	}

	return userMgrConfig, nil
}
