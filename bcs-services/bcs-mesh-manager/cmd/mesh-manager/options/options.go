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

// Package options contains the options for the mesh manager
package options

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// MeshManagerOptions options for mesh manager
type MeshManagerOptions struct {
	conf.LogConfig
	ServerConfig
	ClientConfig
	Etcd        *EtcdConfig    `json:"etcd"`
	Mongo       *MongoConfig   `json:"mongo"`
	Gateway     *GatewayConfig `json:"gateway"`
	IAM         IAMConfig      `json:"iam"`
	Auth        AuthConfig     `json:"auth"`
	IstioConfig *IstioConfig   `json:"istio"`
}

// Parse parse
func Parse(opt *MeshManagerOptions) error {
	configPath := flag.String("f", "./bcs-mesh-manager.json", "server configuration json file")
	flag.Parse()

	err := loadConfigFile(*configPath, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server load json config file failure, %s\n", err.Error())
	}
	return err
}

// NewMeshManagerOptions new mesh manager options
func NewMeshManagerOptions() *MeshManagerOptions {
	return &MeshManagerOptions{
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
		Etcd:         &EtcdConfig{},
		Mongo:        &MongoConfig{},
		Gateway:      &GatewayConfig{},
	}
}

// ServerConfig config for server
type ServerConfig struct {
	Address         string `json:"address"`
	InsecureAddress string `json:"insecureAddress"`
	Port            uint   `json:"port"`
	HTTPPort        uint   `json:"httpPort"`
	MetricPort      uint   `json:"metricPort"`
	ServerCert      string `json:"serverCert"`
	ServerKey       string `json:"serverKey"`
	ServerCa        string `json:"serverCa"`
}

// ClientConfig config for client
type ClientConfig struct {
	ClientCert string `json:"clientCert"`
	ClientKey  string `json:"clientKey"`
	ClientCa   string `json:"clientCa"`
}

// GatewayConfig bcs gateway config
type GatewayConfig struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

// EtcdConfig options for ectd to registry
type EtcdConfig struct {
	EtcdEndpoints string `json:"endpoints" value:"" usage:"endpoints of etcd"`
	EtcdCert      string `json:"cert" value:"" usage:"cert file of etcd"`
	EtcdKey       string `json:"key" value:"" usage:"key file for etcd"`
	EtcdCa        string `json:"ca" value:"" usage:"ca file for etcd"`
}

// MongoConfig option for mongo db
type MongoConfig struct {
	Address        string `json:"address" yaml:"address"`
	Replicaset     string `json:"replicaset" yaml:"replicaset"`
	ConnectTimeout uint   `json:"connectTimeout" yaml:"connectTimeout"`
	AuthDatabase   string `json:"authDatabase" yaml:"authDatabase"`
	Database       string `json:"database" yaml:"database"`
	Username       string `json:"username" yaml:"username"`
	Password       string `json:"password" yaml:"password"`
	MaxPoolSize    uint   `json:"maxPoolSize" yaml:"maxPoolSize"`
	MinPoolSize    uint   `json:"minPoolSize" yaml:"minPoolSize"`
	Encrypted      bool   `json:"encrypted" yaml:"encrypted"`
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

// CredentialScope define credentials scope for a single resource
type CredentialScope struct {
	ProjectCode string `json:"projectCode" yaml:"projectCode"`
	ClusterID   string `json:"clusterID" yaml:"clusterID"`
	ProjectID   string `json:"projectID" yaml:"projectID"`
	Namespace   string `json:"namespace" yaml:"namespace"`
}

// AuthConfig config for auth
type AuthConfig struct {
	Enable bool `json:"enable"`
	// jwt key
	PublicKeyFile  string `json:"publicKeyFile"`
	PrivateKeyFile string `json:"privateKeyFile"`
	PublicKey      string `json:"publicKey"`
	PrivateKey     string `json:"privateKey"`
	// 不鉴权接口，使用逗号分隔，格式 `MeshManager.Health,MeshManager.Health`
	NoAuthMethod string `json:"noAuthMethod"`
	// client 类型用户权限，使用 json 格式，key 为 client 名称，values 为拥有的权限，'*' 表示所有
	// 如：`{"admin": ["*"], "client_a": ["MeshManager.ListTasks"]}`
	ClientPermissions string `json:"clientPermissions"`
}

// loadConfigFile loading json config file
func loadConfigFile(fileName string, opt *MeshManagerOptions) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, opt)
}

// Validate validate options
func (o *MeshManagerOptions) Validate() error {
	if o.IstioConfig == nil {
		return fmt.Errorf("istio config is nil")
	}
	// validate istio config
	if err := o.IstioConfig.Validate(); err != nil {
		return err
	}
	return nil
}

// GlobalOptions global mesh manager options
var GlobalOptions *MeshManagerOptions
