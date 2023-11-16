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

package manager

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/secretstore"
)

const (
	gracefulPeriod = 3
)

// GitOps configuraiotn
type GitOps struct {
	Service        string `json:"service,omitempty"`
	User           string `json:"user,omitempty"`
	Pass           string `json:"pass,omitempty"`
	AdminNamespace string `json:"adminnamespace,omitempty"`
	RepoServer     string `json:"repoServer,omitempty"`
}

// Options for bcs-gitops-manager
type Options struct {
	conf.LogConfig
	common.ServerConfig
	common.ClientConfig
	Registry *common.Registry `json:"registry,omitempty"`
	// work mode, tunnel/service
	Mode string `json:"mode,omitempty"`
	// 用于存放 Cluster Server 地址，为空则使用 APIGateway 的值
	APIGatewayForCluster string                          `json:"apigatewayforcluster,omitempty"`
	APIGateway           string                          `json:"apigateway,omitempty"`
	APIGatewayToken      string                          `json:"apigatewaytoken,omitempty"`
	APIConnectToken      string                          `json:"apiconnecttoken,omitempty"`
	APIConnectURL        string                          `json:"apiconnecturl,omitempty"`
	ClusterSyncInterval  uint                            `json:"clustersyncinterval,omitempty"`
	GitOps               *GitOps                         `json:"gitops,omitempty"`
	PublicProjectsStr    string                          `json:"publicProjects"`
	PublicProjects       []string                        `json:"-"`
	SecretServer         *secretstore.SecretStoreOptions `json:"secretserver,omitempty"`
	Auth                 *common.AuthConfig              `json:"auth,omitempty"`
	TraceConfig          *common.TraceConfig             `json:"traceConfig,omitempty"`
	AuditConfig          *common.AuditConfig             `json:"auditConfig"`
}

// DefaultOptions for gitops-manager
func DefaultOptions() *Options {
	return &Options{
		LogConfig: conf.LogConfig{
			LogDir:       "/data/bcs/logs/bcs",
			Verbosity:    3,
			AlsoToStdErr: true,
		},
		ServerConfig: common.ServerConfig{
			Address:    "127.0.0.1",
			Port:       8081,
			HTTPPort:   8080,
			MetricPort: 8082,
			ServerCa:   "/data/bcs/cert/bcs/bcs-ca.crt",
			ServerCert: "/data/bcs/cert/bcs/bcs-server.crt",
			ServerKey:  "/data/bcs/cert/bcs/bcs-server.key",
		},
		ClientConfig: common.ClientConfig{
			ClientCa:   "/data/bcs/cert/bcs/bcs-ca.crt",
			ClientCert: "/data/bcs/cert/bcs/bcs-client.crt",
			ClientKey:  "/data/bcs/cert/bcs/bcs-client.key",
		},
		Registry: &common.Registry{
			Endpoints: "127.0.0.1",
			CA:        "/data/bcs/cert/etcd/etcd-ca.pem",
			Key:       "/data/bcs/cert/etcd/etcd-key.pem",
			Cert:      "/data/bcs/cert/etcd/etcd.pem",
		},
		Mode:          common.ModeTunnel,
		APIConnectURL: "",
		GitOps:        &GitOps{},
		Auth:          &common.AuthConfig{},
	}
}

// Complete all unsetting config items
func (opt *Options) Complete() error {
	if err := opt.ClientConfig.Complete(); err != nil {
		return err
	}
	if err := opt.ServerConfig.Complete(); err != nil {
		return err
	}
	if err := opt.Registry.Complete(); err != nil {
		return err
	}
	if len(opt.Mode) == 0 {
		opt.Mode = common.ModeTunnel
	}
	if len(opt.APIConnectURL) == 0 {
		opt.APIConnectURL = common.GatewayURL
	}
	if opt.ClusterSyncInterval == 0 {
		opt.ClusterSyncInterval = 300
	}
	if err := opt.Auth.Complete(); err != nil {
		return err
	}
	return nil
}

// Validate all config items
func (opt *Options) Validate() error {
	if opt.ServerTLS == nil {
		return fmt.Errorf("lost server side TLS config")
	}
	if opt.ClientTLS == nil {
		return fmt.Errorf("lost client side TLS config")
	}
	if opt.Registry.TLSConfig == nil {
		return fmt.Errorf("lost registry TLS config")
	}
	if opt.Mode != common.ModeTunnel && opt.Mode != common.ModeService {
		return fmt.Errorf("manager work mode error")
	}
	if opt.Mode == common.ModeTunnel {
		if len(opt.APIGateway) == 0 || len(opt.APIGatewayToken) == 0 {
			return fmt.Errorf("lost bcs-api-gateway config in tunnel mode")
		}
		if len(opt.APIGatewayForCluster) == 0 {
			opt.APIGatewayForCluster = opt.APIGateway
		}
		if len(opt.APIConnectToken) == 0 || len(opt.APIConnectURL) == 0 {
			return fmt.Errorf("lost bcs-api-gateway gitops proxy config in tunnel mode")
		}
	}
	if len(opt.GitOps.AdminNamespace) == 0 {
		return fmt.Errorf("lost gitops service admin namespace")
	}
	if opt.PublicProjectsStr != "" {
		opt.PublicProjects = strings.Split(opt.PublicProjectsStr, ",")
	}
	if opt.SecretServer == nil || opt.SecretServer.Address == "" || opt.SecretServer.Port == "" {
		return fmt.Errorf("lost secret service address or port")
	}
	if err := opt.Auth.Validate(); err != nil {
		return err
	}
	return nil
}
