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

package config

// TLSConf : config for tls
type TLSConf struct {
	ServerCert string `json:"server_cert" yaml:"server_cert"`
	ServerKey  string `json:"server_key" yaml:"server_key"`
	ServerCa   string `json:"server_ca" yaml:"server_ca"`
	ClientCert string `json:"client_cert" yaml:"client_cert"`
	ClientKey  string `json:"client_key" yaml:"client_key"`
	ClientCa   string `json:"client_ca" yaml:"client_ca"`
}

// defaultTLSConf :
func defaultTLSConf() *TLSConf {
	// only for development
	return &TLSConf{
		ServerCert: "",
		ServerKey:  "",
		ServerCa:   "",
		ClientCert: "",
		ClientKey:  "",
		ClientCa:   "",
	}
}
