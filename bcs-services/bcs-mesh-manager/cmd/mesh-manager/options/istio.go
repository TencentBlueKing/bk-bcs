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

package options

// IstioConfig istio config
type IstioConfig struct {
	IstioVersion  []*IstioVersion `json:"istioVersion"`
	FeatureConfig *FeatureConfig  `json:"featureConfig"`
}

// IstioVersion istio version
type IstioVersion struct {
	Name         string `json:"name"`
	ChartVersion string `json:"chartVersion"`
	KubeVersion  string `json:"kubeVersion"`
	Enabled      bool   `json:"enabled"`
}

// FeatureConfig feature config
type FeatureConfig struct {
	OutboundTrafficPolicy           OutboundTrafficPolicy           `json:"outboundTrafficPolicy"`
	HoldApplicationUntilProxyStarts HoldApplicationUntilProxyStarts `json:"holdApplicationUntilProxyStarts"`
	ExitOnZeroActiveConnections     ExitOnZeroActiveConnections     `json:"exitOnZeroActiveConnections"`
}

// OutboundTrafficPolicy outbound traffic policy
type OutboundTrafficPolicy struct {
	Default        string `json:"default"`
	Enabled        bool   `json:"enabled"`
	SupportVersion string `json:"supportVersion,omitempty"`
}

// HoldApplicationUntilProxyStarts hold application until proxy starts
type HoldApplicationUntilProxyStarts struct {
	Default        string `json:"default"`
	Enabled        bool   `json:"enabled"`
	SupportVersion string `json:"supportVersion,omitempty"`
}

// ExitOnZeroActiveConnections exit on zero active connections
type ExitOnZeroActiveConnections struct {
	Default        string `json:"default"`
	Enabled        bool   `json:"enabled"`
	SupportVersion string `json:"supportVersion,omitempty"`
}
