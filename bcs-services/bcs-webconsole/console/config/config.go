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

import "github.com/Tencent/bk-bcs/bcs-common/common/static"

// CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// ConsoleConfig Config is a configuration
type ConsoleConfig struct {
	Address                string
	Port                   int
	ServCert               *CertConfig
	WebConsoleImage        string
	Privilege              bool
	Cmd                    []string
	Tty                    bool
	Ips                    []string
	IsAuth                 bool
	IsOneSession           bool
	DockerUser             string
	DockerPasswd           string
	Image                  string
	IndexPageTemplatesFile string
	MgrPageTemplatesFile   string
}

// NewConsoleConfig create a config object
func NewConsoleConfig() ConsoleConfig {
	return ConsoleConfig{
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
	}
}
