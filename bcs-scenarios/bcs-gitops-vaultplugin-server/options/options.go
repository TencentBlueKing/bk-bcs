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

// Package options define the options of vaultplugin
package options

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// Options vaultplugin-server options
type Options struct {
	conf.FileConfig
	conf.LogConfig

	ServerConfig
	Debug  bool          `json:"debug"`
	Secret SecretOptions `json:"secret"`
	Argo   ArgoOption    `json:"argo"`
}

// ServerConfig option for server side
type ServerConfig struct {
	Address    string `json:"address,omitempty"`
	Port       uint   `json:"port,omitempty"`
	HTTPPort   uint   `json:"httpport,omitempty"`
	MetricPort uint   `json:"metricport,omitempty"`
}

// SecretOptions secret option
type SecretOptions struct {
	CA        string `json:"ca"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Endpoints string `json:"endpoints"`
	Token     string `json:"token"`
}

// ArgoOption were used by the vault and gitops system interaction
type ArgoOption struct {
	Service string `json:"service"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
}

// Validate server options
func (o *Options) Validate() error {
	if o.Secret.Token == "" || o.Secret.Endpoints == "" {
		return fmt.Errorf("lost secret service token or endpoint")
	}
	if o.Argo.Service == "" ||
		o.Argo.User == "" ||
		o.Argo.Pass == "" {
		return fmt.Errorf("lost gitops service options")
	}
	return nil
}
