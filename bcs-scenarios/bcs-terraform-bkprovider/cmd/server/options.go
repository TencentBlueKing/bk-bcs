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
 */

package server

import (
	"crypto/tls"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
)

// Etcd configuration
type Etcd struct {
	Endpoints string `json:"endpoints"`
	CA        string `json:"ca"`
	Key       string `json:"key"`
	Cert      string `json:"cert"`
	tlsConfig *tls.Config
}

// SwaggerConfig option for swagger
type SwaggerConfig struct {
	Dir string `json:"dir"`
}

// ServerConfig option for server
// nolint
type ServerConfig struct {
	Address     string      `json:"address"`
	IPv6Address string      `json:"ipv6address"`
	Port        uint        `json:"port"`
	HTTPPort    uint        `json:"httpport"`
	MetricPort  uint        `json:"metricport"`
	ServerCert  string      `json:"servercert"`
	ServerKey   string      `json:"serverkey"`
	ServerCa    string      `json:"serverca"`
	serverTLS   *tls.Config // nolint
}

// ClientConfig option for as client
type ClientConfig struct {
	ClientCert string      `json:"clientcert"`
	ClientKey  string      `json:"clientkey"`
	ClientCa   string      `json:"clientca"`
	clientTLS  *tls.Config // nolint
}

// AuthConfig config for auth
type AuthConfig struct {
	// jwt key
	PublicKeyFile  string `json:"publicKeyFile"`
	PrivateKeyFile string `json:"privateKeyFile"`
}

// BkSystemConfig dependency config
type BkSystemConfig struct {
	BkAppCode   string `json:"bkAppCode"`
	BkAppSecret string `json:"bkAppSecret"`
	BkEnv       string `json:bkEnv`
}

// Options is the options for server
type Options struct {
	conf.LogConfig
	ServerConfig
	ClientConfig
	Registry *Etcd           `json:"registry"`
	Swagger  *SwaggerConfig  `json:"swagger"`
	Auth     *AuthConfig     `json:"auth"`
	BkSystem *BkSystemConfig `json:"bkSystem"`
}

// DefaultOptions create default options for server
func DefaultOptions() *Options {
	return &Options{
		ServerConfig: ServerConfig{
			Address:    "127.0.0.1",
			Port:       8081,
			HTTPPort:   8080,
			MetricPort: 8082,
		},
		ClientConfig: ClientConfig{},
		Registry:     &Etcd{},
		LogConfig: conf.LogConfig{
			LogDir:       "/data/bcs/logs/bcs",
			Verbosity:    3,
			AlsoToStdErr: true,
		},
	}
}

// Validate all config items
func (opt *Options) Validate() error {

	if opt.serverTLS == nil {
		return fmt.Errorf("lost server side TLS config")
	}
	if opt.Registry.tlsConfig == nil {
		return fmt.Errorf("lost registry TLS config")
	}

	return nil
}

// Complete all unsetting config items
func (opt *Options) Complete() error {

	// loading registry tls configuration
	etcdConfig, err := ssl.ClientTslConfVerity(opt.Registry.CA, opt.Registry.Cert,
		opt.Registry.Key, "")
	if err != nil {
		return fmt.Errorf("loading etcd registry tls configuration failed, %s", err.Error())
	}
	opt.Registry.tlsConfig = etcdConfig

	// loading server tls configuration
	svrConfig, err := ssl.ServerTslConfVerityClient(opt.ServerCa, opt.ServerCert,
		opt.ServerKey, static.ServerCertPwd)
	if err != nil {
		return fmt.Errorf("loading server side tls config failed, %s", err.Error())
	}
	opt.serverTLS = svrConfig

	// loading client tls configuration
	cliConfig, err := ssl.ClientTslConfVerity(opt.ClientCa, opt.ClientCert,
		opt.ClientKey, static.ClientCertPwd)
	if err != nil {
		return fmt.Errorf("loading client side tls configuration failed, %s", err.Error())
	}
	opt.clientTLS = cliConfig

	return nil
}
