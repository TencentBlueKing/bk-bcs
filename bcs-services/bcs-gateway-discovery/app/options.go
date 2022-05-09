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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
)

//NewServerOptions create default ServerOptions
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Etcd: EtcdRegistry{
			Feature: false,
		},
	}
}

// EtcdRegistry config item for etcd discovery
type EtcdRegistry struct {
	Feature     bool   `json:"etcd_feature" value:"false" usage:"switch that turn on etcd registry feature"`
	GrpcModules string `json:"etcd_grpc_modules" value:"MeshManager,LogManager" usage:"modules that support grpc interface"`
	HTTPModules string `json:"etcd_http_modules" value:"MeshManager,LogManager" usage:"modules that support http interface"`
	Address     string `json:"etcd_address" value:"127.0.0.1:2379" usage:"etcd registry feature, multiple ip addresses splited by comma"`
	CA          string `json:"etcd_ca" value:"" usage:"etcd registry CA"`
	Cert        string `json:"etcd_cert" value:"" usage:"etcd registry tls cert file"`
	Key         string `json:"etcd_key" value:"" usage:"etcd registry tls key file"`
}

//ServerOptions command flags for gateway-discovery
type ServerOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LogConfig
	conf.ProcessConfig
	IPv6Mode bool `json:"ipv6_mode" value:"false" usage:"api-gateway connections information, splited by comma if multiple instances. http mode in default, explicit setting https if needed." mapstructure:"ipv6_mode"`
	//gateway admin api info
	AdminAPI              string `json:"admin_api" value:"127.0.0.1:8001" usage:"api-gateway connections information, splited by comma if multiple instances. http mode in default, explicit setting https if needed. custom cert/key comes from client_cert_file/client_key_file" mapstructure:"admin_api" `
	AdminToken            string `json:"amdin_token" value:"" usage:"api-gateway admin api token"`
	AdminType             string `json:"admin_type" value:"apisix" usage:"select apisix or kong as gateway"`
	Modules               string `json:"modules" value:"storage,mesosdriver,detection,usermanager,kubeagent" usage:"new standard moduels that discovery serve for" mapstructure:"modules"`
	AuthToken             string `json:"auth_token" usage:"token for request bcs-user-manager" mapstructure:"auth_token" `
	GatewayMetricsEnabled bool   `json:"gateway_metrics_enabled" value:"true" usage:"gateway(apisix) routes metrics plugins option"`

	Etcd EtcdRegistry `json:"etcdRegistry"`
}

//Valid check if necessary parameter is setting correctly
func (opt *ServerOptions) Valid() error {
	if len(opt.AdminAPI) == 0 {
		return fmt.Errorf("Lost admin api setting")
	}
	if opt.AdminType == "apisix" && len(opt.AdminToken) == 0 {
		return fmt.Errorf("lost apisix admin token")
	}
	if len(opt.ZkConfig.BCSZk) == 0 {
		return fmt.Errorf("Lost bk-bcs zookeeper setting")
	}
	if opt.Etcd.Feature {
		if len(opt.Etcd.Address) == 0 {
			return fmt.Errorf("Lost etcd address information")
		}
		//enable etcd registry feature, we have to ensure tls config
		if len(opt.Etcd.Cert) == 0 || len(opt.Etcd.CA) == 0 ||
			len(opt.Etcd.Key) == 0 {
			return fmt.Errorf("Lost etcd tls config when enable etcd registry feature")
		}
		if len(opt.Etcd.GrpcModules) == 0 || len(opt.Etcd.HTTPModules) == 0 {
			return fmt.Errorf("lost etcd watch module info")
		}
	}
	return nil
}

// GetEtcdRegistryTLS get specified etcd registry tls config
func (opt *ServerOptions) GetEtcdRegistryTLS() (*tls.Config, error) {
	config, err := ssl.ClientTslConfVerity(opt.Etcd.CA, opt.Etcd.Cert,
		opt.Etcd.Key, "")
	if err != nil {
		blog.Errorf("gateway-discovery etcd TLSConfig with CA/Cert/Key failed, %s", err.Error())
		return nil, err
	}
	return config, nil
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
