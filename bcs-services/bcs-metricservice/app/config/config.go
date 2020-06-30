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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
)

//CertConfig  configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// Config bcs-metricservice configure
type Config struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	TempDir           string   `json:"temp_dir" value:"./templates" usage:"the collector application templates directory"`                                                                                            // no lint
	EndpointWatchPath []string `json:"endpoint_watch_path" value:"" usage:"this tells endpoint watch path on bcs_zookeeper like: \"storage:/endpoints/storage\" or \"check:/endpoints/check/*\", * means all nodes."` // no lint
	ApiToken          string   `json:"api_token" value:"" usage:"api token for authority check"`

	StorageIP       string // ip address of the storage
	StoragePort     uint   // port address of the storage
	ZKServerAddress string // discovery

	ServCertDir string      // server cert directory of the server
	ServCert    *CertConfig // cert of the server

	StorageClientCertDir string      // client cert directory of the storage server
	StorageClientCert    *CertConfig // cert of the storage server
	RouteClientCertDir   string      // client cert directory of the route server
	RouteClientCert      *CertConfig // cert of the route server
}

// ParseConfig parse the command parameters into the config
func ParseConfig() *Config {
	c := &Config{
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
		RouteClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
		StorageClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
	}

	conf.Parse(c)
	c.ZKServerAddress = c.BCSZk

	c.ServCert.CertFile = c.ServerCertFile
	c.ServCert.KeyFile = c.ServerKeyFile
	c.ServCert.CAFile = c.CAFile

	if c.ServCert.CertFile != "" && c.ServCert.KeyFile != "" {
		c.ServCert.IsSSL = true
	}

	c.RouteClientCert.CertFile = c.ClientCertFile
	c.RouteClientCert.KeyFile = c.ClientKeyFile
	c.RouteClientCert.CAFile = c.CAFile

	if c.RouteClientCert.CertFile != "" && c.RouteClientCert.KeyFile != "" {
		c.RouteClientCert.IsSSL = true
	}

	c.StorageClientCert.CertFile = c.ClientCertFile
	c.StorageClientCert.KeyFile = c.ClientKeyFile
	c.StorageClientCert.CAFile = c.CAFile

	if c.StorageClientCert.CertFile != "" && c.StorageClientCert.KeyFile != "" {
		c.StorageClientCert.IsSSL = true
	}

	if 0 == len(c.TempDir) {
		fmt.Println("template dir no set")
		os.Exit(1)
	}

	blog.Info("Configuration info: %+#v", c)
	return c
}
