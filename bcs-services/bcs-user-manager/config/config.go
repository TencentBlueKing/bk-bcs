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
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

//CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

//UserMgrConfig is a configuration of bcs-user-manager
type UserMgrConfig struct {
	Address         string
	Port            uint
	InsecureAddress string
	InsecurePort    uint
	LocalIp         string
	Sock            string
	MetricPort      uint
	ServCert        *CertConfig
	ClientCert      *CertConfig
	// server http tls authentication
	TlsServerConfig *tls.Config
	// client http tls authentication
	TlsClientConfig *tls.Config

	VerifyClientTLS bool

	DSN            string
	RedisDSN       string
	BootStrapUsers []options.BootStrapUser
	TKE            options.TKEOptions
	PeerToken      string

	IAMConfig     options.IAMConfig
	ClusterConfig options.ClusterManagerConfig
	EtcdConfig    registry.CMDOptions

	PermissionSwitch bool
}

var (
	//Tke option for sync tke cluster credentials
	Tke options.TKEOptions
	//CliTls for
	CliTls *tls.Config
)

//NewUserMgrConfig create a config object
func NewUserMgrConfig() *UserMgrConfig {
	return &UserMgrConfig{
		Address: "127.0.0.1",
		Port:    80,
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
		ClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
	}
}
