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

// Package options provides configuration structures and default values for the bcs-push-manager service.
package options

import (
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// ServiceOptions defines the options for the service.
type ServiceOptions struct {
	conf.LogConfig `json:"log_config"`
	ServerConfig   `json:"server_config"`
	ClientConfig   `json:"client_config"`
	Gateway        *GatewayConfig    `json:"gateway"`
	Mongo          *MongoOption      `json:"mongodb"`
	Etcd           *EtcdOption       `json:"etcd"`
	Thirdparty     *ThirdpartyOption `json:"thirdparty"`
	RabbitMQ       *RabbitMQOption   `json:"rabbitmq"`
	BkAppCode      string            `json:"bk_app_code"`
	BkAppSecret    string            `json:"bk_app_secret"`
	BkUserName     string            `json:"bk_username"`
}

// ServerConfig defines the config for the server.
type ServerConfig struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Address    string `json:"address"`
	Port       uint   `json:"port"`
	HTTPPort   uint   `json:"httpport"`
	ServerCert string `json:"servercert,omitempty"`
	ServerKey  string `json:"serverkey,omitempty"`
	ServerCa   string `json:"serverca,omitempty"`
}

// MongoOption defines the options for mongodb.
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

// EtcdOption defines the options for etcd.
type EtcdOption struct {
	EtcdEndpoints string `json:"endpoints" value:"" usage:"endpoints of etcd"`
	EtcdCert      string `json:"cert" value:"" usage:"cert file of etcd"`
	EtcdKey       string `json:"key" value:"" usage:"key file for etcd"`
	EtcdCa        string `json:"ca" value:"" usage:"ca file for etcd"`
	TlsConfig     *tls.Config
}

// RabbitMQOption defines the options for RabbitMQ.
type RabbitMQOption struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Vhost          string `json:"vhost"`
	SourceExchange string `json:"source_exchange"`
}

// ThirdpartyOption defines the options for thirdparty services.
type ThirdpartyOption struct {
	Endpoint      string      `json:"endpoint"`
	Token         string      `json:"token"`
	ClientTLS     *tls.Config `json:"-"`
	EtcdEndpoints string      `json:"etcdEndpoints"`
	EtcdTLS       *tls.Config `json:"-"`
}

// ClientConfig config for client
type ClientConfig struct {
	ClientCert string `json:"clientcert"`
	ClientKey  string `json:"clientkey"`
	ClientCa   string `json:"clientca"`
}

// NewServiceOptions creates a new ServiceOptions object.
func NewServiceOptions() *ServiceOptions {
	return &ServiceOptions{
		LogConfig:    conf.LogConfig{},
		ServerConfig: ServerConfig{},
		Mongo:        &MongoOption{},
		Etcd:         &EtcdOption{},
		Thirdparty:   &ThirdpartyOption{},
		RabbitMQ:     &RabbitMQOption{},
		Gateway:      &GatewayConfig{},
	}
}

// Credential can be used to provide authentication options when configuring a Client.
type Credential struct {
	AuthMechanism           string
	AuthMechanismProperties map[string]string
	AuthSource              string
	Username                string
	Password                string
	PasswordSet             bool
}

// GatewayConfig bcs gateway config
type GatewayConfig struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}
