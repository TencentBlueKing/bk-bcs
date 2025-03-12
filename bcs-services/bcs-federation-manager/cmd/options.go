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

// Package cmd xxx
package cmd

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/types"
)

// FederationManagerOptions options for federation manager
type FederationManagerOptions struct {
	conf.LogConfig
	ServerConfig
	ClientConfig
	Etcd    *EtcdOption             `json:"etcd"`
	Mongo   *MongoOption            `json:"mongo"`
	Gateway *GatewayConfig          `json:"gateway"`
	Debug   bool                    `json:"debug"`
	Broker  *BrokerConfig           `json:"broker"`
	Deploy  *types.FederationCharts `json:"deploy"`
	IAM     IAMConfig               `json:"iam"`
	Auth    AuthConfig              `json:"auth"`
}

// Parse parse
func Parse(opt *FederationManagerOptions) error {
	//conf.Parse(opts)
	configPath := flag.String("f", "./bcs-federation-manager.json", "server configuration json file")
	flag.Parse()

	err := loadConfigFile(*configPath, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server load json config file failure, %s\n", err.Error())
	}
	return err
}

// NewFederationManagerOptions new federation manager options
func NewFederationManagerOptions() *FederationManagerOptions {
	return &FederationManagerOptions{
		LogConfig: conf.LogConfig{
			LogDir:       "./logs",
			AlsoToStdErr: true,
			Verbosity:    2,
		},
		ServerConfig: ServerConfig{
			Address:    "127.0.0.1",
			Port:       8081,
			HTTPPort:   8080,
			MetricPort: 8082,
		},
		ClientConfig: ClientConfig{},
		Etcd:         &EtcdOption{},
		Mongo:        &MongoOption{},
		Gateway:      &GatewayConfig{},
		Debug:        false,
	}
}

// ServerConfig config for server
type ServerConfig struct {
	Address         string `json:"address"`
	InsecureAddress string `json:"insecureaddress"`
	Port            uint   `json:"port"`
	HTTPPort        uint   `json:"httpport"`
	MetricPort      uint   `json:"metricport"`
	ServerCert      string `json:"servercert"`
	ServerKey       string `json:"serverkey"`
	ServerCa        string `json:"serverca"`
}

// ClientConfig config for client
type ClientConfig struct {
	ClientCert string `json:"clientcert"`
	ClientKey  string `json:"clientkey"`
	ClientCa   string `json:"clientca"`
}

// GatewayConfig bcs gateway config
type GatewayConfig struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

// EtcdOption options for ectd to registry
type EtcdOption struct {
	EtcdEndpoints string `json:"endpoints" value:"" usage:"endpoints of etcd"`
	EtcdCert      string `json:"cert" value:"" usage:"cert file of etcd"`
	EtcdKey       string `json:"key" value:"" usage:"key file for etcd"`
	EtcdCa        string `json:"ca" value:"" usage:"ca file for etcd"`
	tlsConfig     *tls.Config
}

// MongoOption option for mongo db
type MongoOption struct {
	// MongoEndpoints addr of mongodb
	MongoEndpoints string `json:"endpoints"`
	// MongoConnectTimeout connect timeout of mongodb
	MongoConnectTimeout int `json:"connecttimeout"`
	// MongoDatabaseName database of mongodb
	MongoDatabaseName string `json:"database"`
	// MongoUsername username of mongodb
	MongoUsername string `json:"username"`
	// MongoPassword password of mongodb
	MongoPassword string `json:"password"`
}

// BrokerConfig config for go-machinery broker
type BrokerConfig struct {
	QueueAddress string `json:"address"`
	Exchange     string `json:"exchange"`
}

// IAMConfig config for iam
type IAMConfig struct {
	SystemID         string `json:"systemID"`
	AppCode          string `json:"appCode"`
	AppSecret        string `json:"appSecret"`
	External         bool   `json:"external"`
	GatewayServer    string `json:"gateWayServer"`
	IAMServer        string `json:"iamServer"`
	BkiIAMServer     string `json:"bkiIamServer"`
	Metric           bool   `json:"metric"`
	Debug            bool   `json:"debug"`
	ApplyPermAddress string `json:"applyPermAddress"`
}

// AuthConfig config for auth
type AuthConfig struct {
	Enable bool `json:"enable"`
	// jwt key
	PublicKeyFile  string `json:"publicKeyFile"`
	PrivateKeyFile string `json:"privateKeyFile"`
	// client 类型用户权限，使用 json 格式，key 为 client 名称，values 为拥有的权限，'*' 表示所有
	// 如：`{"admin": ["*"], "client_a": ["FederationManager.ListTasks"]}`
	ClientPermissions string `json:"clientPermissions"`
	// 不鉴权接口，使用逗号分隔，格式 `FederationManager.Health,FederationManager.Health`
	NoAuthMethod string `json:"noAuthMethod"`
}

// loadConfigFile loading json config file
func loadConfigFile(fileName string, opt *FederationManagerOptions) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, opt)
}
