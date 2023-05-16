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

// Package options xxx
package options

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// ServerOption is option in flags
type ServerOption struct {
	conf.FileConfig
	conf.MetricConfig
	conf.LogConfig
	conf.ProcessConfig

	Address        string `json:"address" short:"a" value:"0.0.0.0" usage:"IP address to listen on for this service"`
	Port           uint   `json:"port" short:"p" value:"443" usage:"Port to listen on for this service"`
	ServerCertFile string `json:"server_cert_file" value:"" usage:"Server public key file(*.crt). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`
	ServerKeyFile  string `json:"server_key_file" value:"" usage:"Server private key file(*.key). If both server_cert_file and server_key_file are set, it will set up an HTTPS server"`
	EngineType     string `json:"engine_type" value:"kubernetes" usage:"the platform that bcs-webhook-server runs in, kubernetes or mesos"`
	PluginDir      string `json:"plugin_dir" value:"./plugins" usage:"directory for bcs webhook plugins"`
	Plugins        string `json:"plugins" value:"" usage:"plugin names, call plugin Handle in this order"`
}

const (
	// EngineTypeKubernetes kubernetes engine type
	EngineTypeKubernetes = "kubernetes"
	// EngineTypeMesos mesos engine type
	EngineTypeMesos = "mesos"
)

// NewServerOption create a ServerOption object
func NewServerOption() *ServerOption {
	s := ServerOption{}
	return &s
}

// Parse parse server options
func Parse(ops *ServerOption) error {
	conf.Parse(ops)
	if ops.EngineType != EngineTypeKubernetes && ops.EngineType != EngineTypeMesos {
		return fmt.Errorf("unsupported engine type %s", ops.EngineType)
	}
	strings.Replace(ops.Plugins, ";", ",", -1)
	return nil
}
