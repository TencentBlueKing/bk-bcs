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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

const (
	defaultResourceManagerService = "resourcemanager.bkbcs.tencent.com"
	defaultClusterManagerService  = "clustermanager.bkbcs.tencent.com"
)

// Etcd configuration
type Etcd struct {
	Endpoints string `json:"endpoints"`
	CA        string `json:"ca"`
	Key       string `json:"key"`
	Cert      string `json:"cert"`
	tlsConfig *tls.Config
}

// Storage configuration
type Storage struct {
	Endpoints string `json:"endpoints"`
	Database  string `json:"database"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
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

// GatewayConfig bcs gateway config
type GatewayConfig struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

// AppConfig app config
type AppConfig struct {
	AppCode     string `json:"appCode"`
	AppSecret   string `json:"appSecret"`
	UserName    string `json:"userName"`
	AccessToken string `json:"accessToken"`
}

// DefaultOptions create default options for server
func DefaultOptions() *Options {
	return &Options{
		FileConfig: conf.FileConfig{},
		LogConfig: conf.LogConfig{
			LogDir:       "/data/bcs/logs/bcs",
			Verbosity:    3,
			AlsoToStdErr: true,
		},
		ServerConfig: ServerConfig{
			Address:    "127.0.0.1",
			Port:       8081,
			HTTPPort:   8080,
			MetricPort: 8082,
		},
		ClientConfig:           ClientConfig{},
		ResourceManager:        defaultResourceManagerService,
		ClusterManager:         defaultClusterManagerService,
		Storage:                &Storage{},
		Registry:               &Etcd{},
		Gateway:                &GatewayConfig{},
		Debug:                  false,
		APPConfig:              &AppConfig{},
		EndPoints:              &Endpoints{},
		Concurrency:            0,
		DataControllerInterval: 600,
		DBConfig:               &common.DBConfig{},
		MaxRecvMsgSize:         10 * 1024 * 1024,
	}
}

// Options for whole server
type Options struct {
	conf.FileConfig
	conf.LogConfig
	ServerConfig
	ClientConfig
	ResourceManager        string           `json:"resourcemanager"`
	ClusterManager         string           `json:"clusterManager"`
	Storage                *Storage         `json:"storage"`
	Registry               *Etcd            `json:"registry"`
	Gateway                *GatewayConfig   `json:"gateway"`
	Debug                  bool             `json:"debug"`
	APPConfig              *AppConfig       `json:"appConfig"`
	EndPoints              *Endpoints       `json:"endPoints"`
	Concurrency            int              `json:"concurrency"`
	DataControllerInterval int              `json:"dataControllerInterval"`
	AdminNamespace         string           `json:"adminNamespace"`
	DBConfig               *common.DBConfig `json:"dbConfig,omitempty"`
	MaxRecvMsgSize         int              `json:"maxRecvMsgSize"`
}

// Endpoints endpoints config
type Endpoints struct {
	TGit                string `json:"tGit"`
	TGitSubStr          string `json:"tGitSubStr"`
	PowerApp            string `json:"powerApp"`
	TGitOut             string `json:"tGitOut"`
	TGitOutSubStr       string `json:"tGitOutSubStr"`
	TGitCommunity       string `json:"tGitCommunity"`
	TGitCommunitySubStr string `json:"tGitCommunitySubStr"`
	Gitlab              string `json:"gitlab"`
	GitlabSubStr        string `json:"gitlabSubStr"`
}

// Complete all unsetting config items
func (opt *Options) Complete() error {
	if len(opt.ResourceManager) == 0 {
		opt.ResourceManager = defaultResourceManagerService
	}
	// loading registry tls configuration
	etcdConfig, err := ssl.ClientTslConfVerity(opt.Registry.CA, opt.Registry.Cert,
		opt.Registry.Key, "")
	if err != nil {
		return fmt.Errorf("loading etcd registry tls configuration failed, %s", err.Error())
	}
	opt.Registry.tlsConfig = etcdConfig
	// loading client tls configuration
	cliConfig, err := ssl.ClientTslConfVerity(opt.ClientCa, opt.ClientCert,
		opt.ClientKey, static.ClientCertPwd)
	if err != nil {
		return fmt.Errorf("loading client side tls configuration failed, %s", err.Error())
	}
	opt.clientTLS = cliConfig
	// loading server tls configuration
	svrConfig, err := ssl.ServerTslConfVerityClient(opt.ServerCa, opt.ServerCert,
		opt.ServerKey, static.ServerCertPwd)
	if err != nil {
		return fmt.Errorf("loading server side tls config failed, %s", err.Error())
	}
	opt.serverTLS = svrConfig
	return nil
}

// Validate all config items
func (opt *Options) Validate() error {
	if opt.clientTLS == nil {
		return fmt.Errorf("lost client side TLS config")
	}
	if opt.serverTLS == nil {
		return fmt.Errorf("lost server side TLS config")
	}
	if opt.Registry.tlsConfig == nil {
		return fmt.Errorf("lost registry TLS config")
	}
	// if len(opt.Storage.Password) == 0 {
	//	return fmt.Errorf("lost password")
	// }
	return nil
}
