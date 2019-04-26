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

package config

import (
	"fmt"
	"os"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-common/common/encrypt"
	"bk-bcs/bcs-common/common/static"
)

//CertConfig  configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// Config Exporter服务器配置
type Config struct {
	conf.FileConfig
	conf.MetricConfig
	conf.ServiceConfig
	conf.ZkConfig
	conf.ServerOnlyCertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	OutputPlugins    []string `json:"output_plugins" value:"" usage:"the list of plugins path"`
	OutputAddress    string `json:"output_address" value:"" usage:"the address where the output data sent to"`
	OutputClientCA   string `json:"output_ca_file" value:"" usage:"Output CA file"`
	OutputClientCert string `json:"output_client_cert_file" value:"" usage:"Output client public key file(*.crt)"`
	OutputClientKey  string `json:"output_client_key_file" value:"" usage:"Output client private key file(*.key)"`
	OutputClientPwd  string `json:"output_client_key_pwd" value:"" usage:"Output client private key password"`

	ListenIP        string
	ListenPort      uint
	ZKServerAddress string
	ServCertDir     string      // server cert directory of the server
	ServCert        *CertConfig // cert of the server
	ClientCert      *CertConfig
}

// ParseConfig 解析配置
func ParseConfig() *Config {

	c := &Config{
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
		ClientCert: &CertConfig{
			IsSSL: false,
		},
	}

	conf.Parse(c)

	c.ListenIP = c.Address
	c.ListenPort = c.Port
	c.ZKServerAddress = c.BCSZk

	c.ServCert.CAFile = c.CAFile
	c.ServCert.CertFile = c.ServerCertFile
	c.ServCert.KeyFile = c.ServerKeyFile

	if c.ServCert.CertFile != "" && c.ServCert.KeyFile != "" {
		c.ServCert.IsSSL = true
	}

	c.ClientCert.CAFile = c.OutputClientCA
	c.ClientCert.CertFile = c.OutputClientCert
	c.ClientCert.KeyFile = c.OutputClientKey
	c.ClientCert.CertPasswd = c.OutputClientPwd
	if c.ClientCert.CertPasswd != "" {
		pwd, err := encrypt.DesDecryptFromBase([]byte(c.ClientCert.CertPasswd))
		if err != nil {
			fmt.Println("decode client cert password failed!")
			os.Exit(1)
		}
		c.ClientCert.CertPasswd = string(pwd)
	}

	if c.ClientCert.CertFile != "" && c.ClientCert.KeyFile != "" {
		c.ClientCert.IsSSL = true
	}

	blog.Info("Configuration info: %+v", c)
	return c
}
