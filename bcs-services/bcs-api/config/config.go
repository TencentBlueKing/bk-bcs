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
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/options"
)

//CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

//ApiServConfig is a configuration of apiserver
type ApiServConfig struct {
	Address         string
	Port            uint
	InsecureAddress string
	InsecurePort    uint
	Sock            string
	CIHost          string
	CCHost          string
	CCPhpHost       string
	BcsRoute        string
	BcsDataHost     string
	HubHost         string
	TDHost          string
	AuthHost        string
	RegDiscvSrv     string
	JfrogAccount    map[string]string
	Filter          bool
	LocalIp         string
	MetricPort      uint
	VerifyClientTLS bool

	BKIamAuth options.AuthOption

	ServCert   *CertConfig
	ClientCert *CertConfig

	BKE                      options.BKEOptions
	TKE                      options.TKEOptions
	Edition                  string
	MesosWebconsoleProxyPort uint
	PeerToken                string
}

var (
	Edition                    = ""
	TurnOnRBAC                 = false
	BKIamAuth                  options.AuthOption
	ClusterCredentialsFixtures options.CredentialsFixturesOptions
	MesosWebconsoleProxyPort   uint
	TkeConf                    options.TKEOptions
)

//NewApiServConfig create a config object
func NewApiServConfig() *ApiServConfig {
	return &ApiServConfig{
		Address: "127.0.0.1",
		Port:    50001,
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
		ClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
		Filter: false,
	}
}
