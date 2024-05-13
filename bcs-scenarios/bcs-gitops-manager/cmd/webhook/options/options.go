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

// Package options defines the options of webhook server
package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// GitopsWebhookOptions defines the option of gitops webhook
type GitopsWebhookOptions struct {
	conf.FileConfig
	conf.LogConfig
	conf.CertConfig

	Address string `json:"address,omitempty"`
	// IPv6Address string `json:"ipv6address,omitempty"`
	GRPCPort   uint `json:"grpcPort,omitempty"`
	HTTPPort   uint `json:"httpPort,omitempty"`
	MetricPort uint `json:"metricPort,omitempty"`

	Registry Registry `json:"registry,omitempty"`

	GitOpsWebhook string `json:"gitOpsWebhook"`
	GitOpsToken   string `json:"gitOpsToken"`

	TraceConfig TraceConfig `json:"traceConfig"`
}

// TraceConfig defines the config of trace
type TraceConfig struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

// Registry defines the registry of gitops webhook
type Registry struct {
	Endpoints string `json:"endpoints,omitempty"`
	CA        string `json:"ca,omitempty"`
	Key       string `json:"key,omitempty"`
	Cert      string `json:"cert,omitempty"`
}
