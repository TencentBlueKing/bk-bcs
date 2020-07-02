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
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/rdiscover"
)

//CertConfig  configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// RunType collector running mode
type RunType string

const (
	//ContainerType for Container mode
	ContainerType RunType = "container"
	//TraditionalType for process mode
	TraditionalType RunType = "traditional"
)

// Config Exporter服务器配置
type Config struct {
	conf.FileConfig
	conf.CertConfig
	conf.MetricConfig
	conf.LogConfig
	conf.ProcessConfig
	conf.ZkConfig
	conf.LocalConfig
	RunMode               RunType              `json:"run_mode" value:"container" usage:"should be one of container or traditional. container for containerized app in mesos/k8s, traditional for traditional app."` //no lint
	ExporterType          int                  `json:"exporter_type" value:"3" usage:"the type of exporter"`
	ZKServerAddress       string               // discovery
	MetricClientCertDir   string               // client cert directory of the metric service client
	MetricClientCert      *CertConfig          // cert of the  metric service client
	ExporterClientCertDir string               // client cert directory of the exporter client
	ExporterClientCert    *CertConfig          // cert of the exporter client
	Rd                    *rdiscover.RDiscover // rd
}

// ParseConfig 解析配置
func ParseConfig() *Config {

	c := &Config{
		ExporterClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
		MetricClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
		ZKServerAddress: os.Getenv("ZkAddress"),
	}

	conf.Parse(c)

	if c.RunMode == TraditionalType {
		c.ZKServerAddress = c.BCSZk
	}

	c.MetricClientCert.CAFile = c.CAFile
	c.MetricClientCert.CertFile = c.ClientCertFile
	c.MetricClientCert.KeyFile = c.ClientKeyFile

	if c.MetricClientCert.CertFile != "" && c.MetricClientCert.KeyFile != "" {
		c.MetricClientCert.IsSSL = true
	}

	c.ExporterClientCert.CAFile = c.CAFile
	c.ExporterClientCert.CertFile = c.ClientCertFile
	c.ExporterClientCert.KeyFile = c.ClientKeyFile

	if c.ExporterClientCert.CertFile != "" && c.ExporterClientCert.KeyFile != "" {
		c.ExporterClientCert.IsSSL = true
	}

	blog.Info("Configuration info: %+v", c)
	return c
}
