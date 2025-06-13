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

package common

// IstiodInstallArgs istiod安装参数
type IstiodInstallArgs struct {
	IstiodRemote *IstiodRemoteConfig       `yaml:"istiodRemote,omitempty"`
	Pilot        *IstiodPilotConfig        `yaml:"pilot,omitempty"`
	Telemetry    *IstiodTelemetryConfig    `yaml:"telemetry,omitempty"`
	Global       *IstiodGlobalConfig       `yaml:"global,omitempty"`
	MultiCluster *IstiodMultiClusterConfig `yaml:"multiCluster,omitempty"`
	MeshConfig   *IstiodMeshConfig         `yaml:"meshConfig,omitempty"`
}

// IstiodRemoteConfig istiod远程配置
type IstiodRemoteConfig struct {
	Enabled       *bool   `yaml:"enabled,omitempty"`
	InjectionPath *string `yaml:"injectionPath,omitempty"`
}

// IstiodPilotConfig pilot配置
type IstiodPilotConfig struct {
	ConfigMap *bool             `yaml:"configMap,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
}

// IstiodTelemetryConfig telemetry配置
type IstiodTelemetryConfig struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// IstiodGlobalConfig global配置
type IstiodGlobalConfig struct {
	MeshID                       *string           `yaml:"meshID,omitempty"`
	Network                      *string           `yaml:"network,omitempty"`
	ConfigCluster                *bool             `yaml:"configCluster,omitempty"`
	OmitSidecarInjectorConfigMap *bool             `yaml:"omitSidecarInjectorConfigMap,omitempty"`
	RemotePilotAddress           *string           `yaml:"remotePilotAddress,omitempty"`
	ExternalIstiod               *bool             `yaml:"externalIstiod,omitempty"`
	Proxy                        *IstioProxyConfig `yaml:"proxy,omitempty"`
}

// IstioProxyConfig proxy配置
type IstioProxyConfig struct {
	ExcludeIPRanges *string `yaml:"excludeIPRanges,omitempty"`
}

// IstiodMultiClusterConfig multiCluster配置
type IstiodMultiClusterConfig struct {
	ClusterName *string `yaml:"clusterName,omitempty"`
}

// IstiodMeshConfig mesh配置
type IstiodMeshConfig struct {
	OutboundTrafficPolicy *OutboundTrafficPolicy `yaml:"outboundTrafficPolicy,omitempty"`
	DefaultConfig         *DefaultConfig         `yaml:"defaultConfig,omitempty"`
}

// OutboundTrafficPolicy 出站流量策略
type OutboundTrafficPolicy struct {
	Mode *string `yaml:"mode,omitempty"`
}

// DefaultConfig 默认配置
type DefaultConfig struct {
	HoldApplicationUntilProxyStarts *bool          `yaml:"holdApplicationUntilProxyStarts,omitempty"`
	ProxyMetadata                   *ProxyMetadata `yaml:"proxyMetadata,omitempty"`
}

// ProxyMetadata proxy metadata
type ProxyMetadata struct {
	ExitOnZeroActiveConnections *bool   `yaml:"EXIT_ON_ZERO_ACTIVE_CONNECTIONS,omitempty"`
	IstioMetaDnsCapture         *string `yaml:"ISTIO_META_DNS_CAPTURE,omitempty"`
	IstioMetaDnsAutoAllocate    *string `yaml:"ISTIO_META_DNS_AUTO_ALLOCATE,omitempty"`
}
