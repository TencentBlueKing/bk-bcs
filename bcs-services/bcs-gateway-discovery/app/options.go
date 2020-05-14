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
	"crypto/tls"
	"fmt"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-common/common/static"
)

//NewServerOptions create default ServerOptions
func NewServerOptions() *ServerOptions {
	return &ServerOptions{}
}

//Options command flags for gateway-discovery
type ServerOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig
	IPv6Mode bool `json:"ipv6_mode" value:"false" usage:"api-gateway connections information, splited by comma if multiple instances. http mode in default, explicit setting https if needed." mapstructure:"ipv6_mode"`
	//gateway admin api info
	AdminAPI string `json:"admin_api" value:"127.0.0.1:8001" usage:"api-gateway connections information, splited by comma if multiple instances. http mode in default, explicit setting https if needed. custom cert/key comes from client_cert_file/client_key_file" mapstructure:"admin_api" `
	//new standard modules
	Modules   []string `json:"modules" usage:"new standard moduels that discovery serve for" mapstructure:"modules" `
	AuthToken string   `json:"auth_token" usage:"token for request bcs-user-manager" mapstructure:"auth_token" `
}

//Valid check if necessary paramter is setting correctly
func (opt *ServerOptions) Valid() error {
	if len(opt.AdminAPI) == 0 {
		return fmt.Errorf("Lost admin api setting")
	}
	if len(opt.ZkConfig.BCSZk) == 0 {
		return fmt.Errorf("Lost bk-bcs zookeeper setting")
	}
	return nil
}

//GetClientTLS construct client tls configuration
func (opt *ServerOptions) GetClientTLS() (*tls.Config, error) {
	if len(opt.CertConfig.CAFile) != 0 && len(opt.CertConfig.ClientCertFile) == 0 {
		//work with CA, and verify Server certification
		config, err := ssl.ClientTslConfVerityServer(opt.CertConfig.CAFile)
		if err != nil {
			blog.Errorf("gateway-discovery tls with only CA failed, %s", err.Error())
			return nil, err
		}
		return config, nil
	}
	//tls with CA/ClientCert/ClientKey
	if len(opt.CertConfig.CAFile) != 0 && len(opt.CertConfig.ClientCertFile) != 0 &&
		len(opt.CertConfig.ClientKeyFile) != 0 {
		config, err := ssl.ClientTslConfVerity(opt.CertConfig.CAFile, opt.CertConfig.ClientCertFile,
			opt.CertConfig.ClientKeyFile, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("gateway-discovery tls with CA/Cert/Key failed, %s", err.Error())
			return nil, err
		}
		return config, nil
	}
	return nil, fmt.Errorf("tls config error, only setting CA or setting CA/ClientCert/ClientKey")
}
