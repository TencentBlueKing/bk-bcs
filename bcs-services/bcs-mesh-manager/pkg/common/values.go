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

// Field key constants for Istio configuration
const (
	FieldKeyEnableTracing             = "enableTracing"
	FieldKeyMeshConfig                = "meshConfig"
	FieldKeyPilot                     = "pilot"
	FieldKeyTelemetry                 = "telemetry"
	FieldKeyGlobal                    = "global"
	FieldKeyMultiCluster              = "multiCluster"
	FieldKeyExtensionProviders        = "extensionProviders"
	FieldKeyTraceSampling             = "traceSampling"
	FieldKeyDefaultConfig             = "defaultConfig"
	FieldKeyTracingConfig             = "tracing"
	FieldKeyAccessLogFile             = "accessLogFile"
	FieldKeyAccessLogFormat           = "accessLogFormat"
	FieldKeyAccessLogEncoding         = "accessLogEncoding"
	FieldKeyAutoscaleEnabled          = "autoscaleEnabled"
	FieldKeyAutoscaleMin              = "autoscaleMin"
	FieldKeyAutoscaleMax              = "autoscaleMax"
	FieldKeyCPU                       = "cpu"
	FieldKeyMemory                    = "memory"
	FieldKeyDedicatedNode             = "dedicatedNode"
	FieldKeyDedicatedNodeEnabled      = "enabled"
	FieldKeyDedicatedNodeNodeSelector = "nodeSelector"
	FieldKeyDedicatedNodeTolerations  = "tolerations"
	FieldKeyZipkin                    = "zipkin"
	FieldKeyZipkinAddress             = "address"
	FieldKeyResources                 = "resources"
	FieldKeyRequests                  = "requests"
	FieldKeyLimits                    = "limits"
	FieldKeyProxy                     = "proxy"
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
	Resources        *ResourceConfig   `yaml:"resources,omitempty"`
	AutoscaleEnabled *bool             `yaml:"autoscaleEnabled,omitempty"`
	AutoscaleMin     *int32            `yaml:"autoscaleMin,omitempty"`
	AutoscaleMax     *int32            `yaml:"autoscaleMax,omitempty"`
	ReplicaCount     *int32            `yaml:"replicaCount,omitempty"`
	TraceSampling    *float64          `yaml:"traceSampling,omitempty"`
	ConfigMap        *bool             `yaml:"configMap,omitempty"`
	CPU              *HPACPUConfig     `yaml:"cpu,omitempty"`
	Env              map[string]string `yaml:"env,omitempty"`
	NodeSelector     map[string]string `yaml:"nodeSelector,omitempty"`
	Tolerations      []v1.Toleration   `yaml:"tolerations,omitempty"`
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
	ExcludeIPRanges *string         `yaml:"excludeIPRanges,omitempty"`
	Resources       *ResourceConfig `yaml:"resources,omitempty"`
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
	ExtensionProviders    []*ExtensionProvider   `yaml:"extensionProviders,omitempty"`
	AccessLogFile         *string                `yaml:"accessLogFile,omitempty"`
	AccessLogFormat       *string                `yaml:"accessLogFormat,omitempty"`
	AccessLogEncoding     *string                `yaml:"accessLogEncoding,omitempty"`
}

// ExtensionProvider 扩展提供者
type ExtensionProvider struct {
	Name          *string              `yaml:"name,omitempty"`
	OpenTelemetry *OpenTelemetryConfig `yaml:"opentelemetry,omitempty"`
}

// OpenTelemetryConfig open telemetry配置
type OpenTelemetryConfig struct {
	Service *string                  `yaml:"service,omitempty"`
	Port    *int32                   `yaml:"port,omitempty"`
	Http    *OpenTelemetryHttpConfig `yaml:"http,omitempty"`
}

// OpenTelemetryHttpConfig http配置
type OpenTelemetryHttpConfig struct {
	Path    *string           `yaml:"path,omitempty"`
	Timeout *string           `yaml:"timeout,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// OutboundTrafficPolicy 出站流量策略
type OutboundTrafficPolicy struct {
	Mode *string `yaml:"mode,omitempty"`
}

// DefaultConfig 默认配置
type DefaultConfig struct {
	HoldApplicationUntilProxyStarts *bool          `yaml:"holdApplicationUntilProxyStarts,omitempty"`
	ProxyMetadata                   *ProxyMetadata `yaml:"proxyMetadata,omitempty"`
	TracingConfig                   *TracingConfig `yaml:"tracing,omitempty"`
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
	ExitOnZeroActiveConnections *string `yaml:"EXIT_ON_ZERO_ACTIVE_CONNECTIONS,omitempty"`
	IstioMetaDnsCapture         *string `yaml:"ISTIO_META_DNS_CAPTURE,omitempty"`
	IstioMetaDnsAutoAllocate    *string `yaml:"ISTIO_META_DNS_AUTO_ALLOCATE,omitempty"`
}

// ResourceConfig 自定义资源配置，用于正确的 YAML 序列化
type ResourceConfig struct {
	Limits   *ResourceLimits   `yaml:"limits,omitempty"`
	Requests *ResourceRequests `yaml:"requests,omitempty"`
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	CPU    *string `yaml:"cpu,omitempty"`
	Memory *string `yaml:"memory,omitempty"`
}

// ResourceRequests 资源请求
type ResourceRequests struct {
	CPU    *string `yaml:"cpu,omitempty"`
	Memory *string `yaml:"memory,omitempty"`
}
