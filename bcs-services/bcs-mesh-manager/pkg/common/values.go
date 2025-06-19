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

import (
	v1 "k8s.io/api/core/v1"

	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

const (
	// AccessLogFileStdout 访问日志文件
	AccessLogFileStdout = "/dev/stdout"
)

// IstioInstallOption istio安装操作选项
type IstioInstallOption struct {
	ChartValuesPath string
	ChartRepo       string

	ProjectID             string
	ProjectCode           string
	Name                  string
	Description           string
	Version               string
	ControlPlaneMode      string
	ClusterMode           string
	PrimaryClusters       []string
	RemoteClusters        []string
	SidecarResourceConfig *meshmanager.ResourceConfig
	HighAvailability      *meshmanager.HighAvailability
	ObservabilityConfig   *meshmanager.ObservabilityConfig
	FeatureConfigs        map[string]*meshmanager.FeatureConfig

	MeshID       string
	NetworkID    string
	ChartVersion string
}

// IstiodInstallValues istiod安装参数
type IstiodInstallValues struct {
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
	Resources        *v1.ResourceRequirements `yaml:"resources,omitempty"`
	AutoscaleEnabled *bool                    `yaml:"autoscaleEnabled,omitempty"`
	AutoscaleMin     *int32                   `yaml:"autoscaleMin,omitempty"`
	AutoscaleMax     *int32                   `yaml:"autoscaleMax,omitempty"`
	ReplicaCount     *int32                   `yaml:"replicaCount,omitempty"`
	TraceSampling    *float64                 `yaml:"traceSampling,omitempty"`
	ConfigMap        *bool                    `yaml:"configMap,omitempty"`
	CPU              *HPACPUConfig            `yaml:"cpu,omitempty"`
	Env              map[string]string        `yaml:"env,omitempty"`
	NodeSelector     map[string]string        `yaml:"nodeSelector,omitempty"`
}

// HPACPUConfig HPA cpu配置
type HPACPUConfig struct {
	TargetAverageUtilization *int32 `yaml:"targetAverageUtilization,omitempty"`
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
	ExcludeIPRanges *string                  `yaml:"excludeIPRanges,omitempty"`
	Resources       *v1.ResourceRequirements `yaml:"resources,omitempty"`
}

// IstiodMultiClusterConfig multiCluster配置
type IstiodMultiClusterConfig struct {
	ClusterName *string `yaml:"clusterName,omitempty"`
}

// IstiodMeshConfig mesh配置
type IstiodMeshConfig struct {
	OutboundTrafficPolicy *OutboundTrafficPolicy `yaml:"outboundTrafficPolicy,omitempty"`
	DefaultConfig         *DefaultConfig         `yaml:"defaultConfig,omitempty"`
	EnableTracing         *bool                  `yaml:"enableTracing,omitempty"`
	AccessLogFile         *string                `yaml:"accessLogFile,omitempty"`
	AccessLogFormat       *string                `yaml:"accessLogFormat,omitempty"`
	AccessLogEncoding     *string                `yaml:"accessLogEncoding,omitempty"`
}

// OutboundTrafficPolicy 出站流量策略
type OutboundTrafficPolicy struct {
	Mode *string `yaml:"mode,omitempty"`
}

// DefaultConfig 默认配置
type DefaultConfig struct {
	HoldApplicationUntilProxyStarts *bool          `yaml:"holdApplicationUntilProxyStarts,omitempty"`
	ProxyMetadata                   *ProxyMetadata `yaml:"proxyMetadata,omitempty"`
	TracingConfig                   *TracingConfig `yaml:"tracingConfig,omitempty"`
}

// TracingConfig 追踪配置
type TracingConfig struct {
	Zipkin *ZipkinConfig `yaml:"zipkin,omitempty"`
}

// ZipkinConfig zipkin配置
type ZipkinConfig struct {
	Address *string `yaml:"address,omitempty"`
}

// ProxyMetadata proxy metadata
type ProxyMetadata struct {
	ExitOnZeroActiveConnections *bool   `yaml:"EXIT_ON_ZERO_ACTIVE_CONNECTIONS,omitempty"`
	IstioMetaDnsCapture         *string `yaml:"ISTIO_META_DNS_CAPTURE,omitempty"`
	IstioMetaDnsAutoAllocate    *string `yaml:"ISTIO_META_DNS_AUTO_ALLOCATE,omitempty"`
}
