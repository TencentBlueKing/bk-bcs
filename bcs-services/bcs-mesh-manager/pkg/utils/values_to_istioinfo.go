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

package utils

import (
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ConvertValuesToIstioDetailInfo 将 MeshIstio 实体和 IstiodInstallValues 配置转换为 IstioListItem 展示结构
func ConvertValuesToIstioDetailInfo(
	meshIstio *entity.MeshIstio,
	istiodValues *common.IstiodInstallValues,
) (*meshmanager.IstioDetailInfo, error) {
	if meshIstio == nil {
		blog.Errorf("meshIstio is nil")
		return nil, fmt.Errorf("meshIstio is nil")
	}
	if istiodValues == nil {
		blog.Errorf("istiodValues is nil")
		return nil, fmt.Errorf("istiodValues is nil")
	}
	// 使用 Transfer2Proto 方法进行基础转换
	result := meshIstio.Transfer2ProtoForDetail()

	// 同步values中的revision
	if istiodValues.Revision != nil {
		result.Revision = *istiodValues.Revision
	}

	// 同步values中的资源配置
	if istiodValues.Global != nil &&
		istiodValues.Global.Proxy != nil &&
		istiodValues.Global.Proxy.Resources != nil {
		result.SidecarResourceConfig = convertResourceConfigValues(istiodValues.Global.Proxy.Resources)
	}

	// 同步values中的高可用配置
	if istiodValues.Pilot != nil {
		convertHighAvailabilityValues(istiodValues.Pilot, result)
	}

	// 同步values中的可观测性配置
	if istiodValues.MeshConfig != nil {
		convertObservabilityConfigValues(istiodValues, result)
	}

	// 同步values中的功能特性配置
	convertFeatureConfigs(istiodValues, result)

	return result, nil
}

// convertResourceConfigValues 从实际的资源配置构建 ResourceConfig
func convertResourceConfigValues(
	resources *common.ResourceConfig,
) *meshmanager.ResourceConfig {
	config := &meshmanager.ResourceConfig{}

	if resources.Requests != nil {
		if resources.Requests.CPU != nil {
			config.CpuRequest = wrapperspb.String(*resources.Requests.CPU)
		}
		if resources.Requests.Memory != nil {
			config.MemoryRequest = wrapperspb.String(*resources.Requests.Memory)
		}
	}

	if resources.Limits != nil {
		if resources.Limits.CPU != nil {
			config.CpuLimit = wrapperspb.String(*resources.Limits.CPU)
		}
		if resources.Limits.Memory != nil {
			config.MemoryLimit = wrapperspb.String(*resources.Limits.Memory)
		}
	}

	return config
}

// updateResourceConfigValues 从实际的资源配置更新现有的 ResourceConfig
func updateResourceConfigValues(
	resources *common.ResourceConfig,
	config *meshmanager.ResourceConfig,
) {
	if config == nil {
		return
	}

	if resources.Requests != nil {
		if resources.Requests.CPU != nil {
			config.CpuRequest = wrapperspb.String(*resources.Requests.CPU)
		}
		if resources.Requests.Memory != nil {
			config.MemoryRequest = wrapperspb.String(*resources.Requests.Memory)
		}
	}

	if resources.Limits != nil {
		if resources.Limits.CPU != nil {
			config.CpuLimit = wrapperspb.String(*resources.Limits.CPU)
		}
		if resources.Limits.Memory != nil {
			config.MemoryLimit = wrapperspb.String(*resources.Limits.Memory)
		}
	}
}

// convertHighAvailabilityValues 从实际的高可用配置更新 HighAvailability
func convertHighAvailabilityValues(
	pilot *common.IstiodPilotConfig,
	result *meshmanager.IstioDetailInfo,
) {
	if result.HighAvailability == nil {
		result.HighAvailability = &meshmanager.HighAvailability{}
	}
	// 更新副本数
	if pilot.ReplicaCount != nil {
		result.HighAvailability.ReplicaCount = wrapperspb.Int32(*pilot.ReplicaCount)
	}
	// 更新自动扩缩容配置
	if pilot.AutoscaleEnabled != nil && *pilot.AutoscaleEnabled {
		result.HighAvailability.AutoscaleEnabled = wrapperspb.Bool(true)
		if pilot.AutoscaleMin != nil {
			result.HighAvailability.AutoscaleMin = wrapperspb.Int32(*pilot.AutoscaleMin)
		}
		if pilot.AutoscaleMax != nil {
			result.HighAvailability.AutoscaleMax = wrapperspb.Int32(*pilot.AutoscaleMax)
		}
		if pilot.CPU != nil && pilot.CPU.TargetAverageUtilization != nil {
			result.HighAvailability.TargetCPUAverageUtilizationPercent =
				wrapperspb.Int32(*pilot.CPU.TargetAverageUtilization)
		}
	} else {
		result.HighAvailability.AutoscaleEnabled = wrapperspb.Bool(false)
	}

	// 更新资源配置
	if pilot.Resources != nil {
		if result.HighAvailability.ResourceConfig == nil {
			result.HighAvailability.ResourceConfig = &meshmanager.ResourceConfig{}
		}
		updateResourceConfigValues(pilot.Resources, result.HighAvailability.ResourceConfig)
	}

	// 更新专属节点配置
	if len(pilot.NodeSelector) > 0 {
		if result.HighAvailability.DedicatedNode == nil {
			result.HighAvailability.DedicatedNode = &meshmanager.DedicatedNode{}
		}
		result.HighAvailability.DedicatedNode.Enabled = wrapperspb.Bool(true)
		result.HighAvailability.DedicatedNode.NodeLabels = pilot.NodeSelector
	}
}

// convertObservabilityConfigValues 从实际的可观测性配置更新 ObservabilityConfig
func convertObservabilityConfigValues(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	meshConfig := istiodValues.MeshConfig
	// 确保 result.ObservabilityConfig 存在
	if result.ObservabilityConfig == nil {
		result.ObservabilityConfig = &meshmanager.ObservabilityConfig{}
	}

	// 更新追踪配置
	if result.ObservabilityConfig.TracingConfig == nil {
		result.ObservabilityConfig.TracingConfig = &meshmanager.TracingConfig{}
	}
	result.ObservabilityConfig.TracingConfig.Enabled =
		wrapperspb.Bool(meshConfig.EnableTracing != nil && *meshConfig.EnableTracing)

	// 先看istio版本
	if IsVersionSupported(result.Version, ">=1.21") {
		// 高于1.21的版本，使用otel
		if meshConfig.ExtensionProviders != nil {
			for _, provider := range meshConfig.ExtensionProviders {
				if provider.Name != nil && *provider.Name != OtelTracingName {
					continue
				}
				// 匹配到 otel-tracing
				endpoint := ""
				if provider.OpenTelemetry != nil && provider.OpenTelemetry.Port != nil && provider.OpenTelemetry.Service != nil {
					endpoint = *provider.OpenTelemetry.Service + ":" + strconv.Itoa(int(*provider.OpenTelemetry.Port))
				}
				if provider.OpenTelemetry != nil && provider.OpenTelemetry.Http != nil && provider.OpenTelemetry.Http.Path != nil {
					endpoint += *provider.OpenTelemetry.Http.Path
				}
				result.ObservabilityConfig.TracingConfig.Endpoint =
					wrapperspb.String(endpoint)

				// 获取token
				if provider.OpenTelemetry != nil &&
					provider.OpenTelemetry.Http != nil &&
					provider.OpenTelemetry.Http.Headers != nil {
					if token, ok := provider.OpenTelemetry.Http.Headers[OtelTracingHeader]; ok {
						result.ObservabilityConfig.TracingConfig.BkToken = wrapperspb.String(token)
					}
				}
			}
		}

	} else {
		// 低于1.21的版本，使用zipkin
		if meshConfig.DefaultConfig != nil && meshConfig.DefaultConfig.TracingConfig != nil &&
			meshConfig.DefaultConfig.TracingConfig.Zipkin != nil &&
			meshConfig.DefaultConfig.TracingConfig.Zipkin.Address != nil {
			result.ObservabilityConfig.TracingConfig.Endpoint =
				wrapperspb.String(*meshConfig.DefaultConfig.TracingConfig.Zipkin.Address)
		}
	}
	// 获取采样率
	if istiodValues.Pilot != nil && istiodValues.Pilot.TraceSampling != nil {
		result.ObservabilityConfig.TracingConfig.TraceSamplingPercent =
			wrapperspb.Int32(int32(*istiodValues.Pilot.TraceSampling * 100))
	}

	// 更新日志配置
	if result.ObservabilityConfig.LogCollectorConfig == nil {
		result.ObservabilityConfig.LogCollectorConfig = &meshmanager.LogCollectorConfig{}
	}
	result.ObservabilityConfig.LogCollectorConfig.Enabled =
		wrapperspb.Bool(meshConfig.AccessLogFile != nil && *meshConfig.AccessLogFile != "")
	// 更新日志格式
	if meshConfig.AccessLogFormat != nil {
		result.ObservabilityConfig.LogCollectorConfig.AccessLogFormat = wrapperspb.String(*meshConfig.AccessLogFormat)
	}
	// 更新日志编码
	if meshConfig.AccessLogEncoding != nil {
		result.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding = wrapperspb.String(*meshConfig.AccessLogEncoding)
	}
}

// convertFeatureConfigs 从实际的功能特性配置更新 FeatureConfigs
func convertFeatureConfigs(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	// 确保 result.FeatureConfigs 存在
	if result.FeatureConfigs == nil {
		result.FeatureConfigs = make(map[string]*meshmanager.FeatureConfig)
	}

	// 转换各个特性配置
	convertOutboundTrafficPolicy(istiodValues, result)
	convertHoldApplicationUntilProxyStarts(istiodValues, result)
	convertExitOnZeroActiveConnections(istiodValues, result)
	convertIstioMetaDnsCapture(istiodValues, result)
	convertIstioMetaDnsAutoAllocate(istiodValues, result)
	convertIstioMetaHttp10(istiodValues, result)
	convertExcludeIPRanges(istiodValues, result)
}

// setFeatureConfig 设置特性配置的通用辅助函数
func setFeatureConfig(
	result *meshmanager.IstioDetailInfo,
	featureName string,
	value string,
) {
	if result.FeatureConfigs[featureName] == nil {
		result.FeatureConfigs[featureName] = &meshmanager.FeatureConfig{}
	}
	result.FeatureConfigs[featureName].Value = value
}

// convertOutboundTrafficPolicy 转换出站流量策略配置
func convertOutboundTrafficPolicy(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	if istiodValues.MeshConfig != nil && istiodValues.MeshConfig.OutboundTrafficPolicy != nil &&
		istiodValues.MeshConfig.OutboundTrafficPolicy.Mode != nil {
		setFeatureConfig(result, common.FeatureOutboundTrafficPolicy, *istiodValues.MeshConfig.OutboundTrafficPolicy.Mode)
	}
}

// convertHoldApplicationUntilProxyStarts 转换等待代理启动配置
func convertHoldApplicationUntilProxyStarts(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	if istiodValues.MeshConfig != nil && istiodValues.MeshConfig.DefaultConfig != nil &&
		istiodValues.MeshConfig.DefaultConfig.HoldApplicationUntilProxyStarts != nil {
		setFeatureConfig(result, common.FeatureHoldApplicationUntilProxyStarts,
			fmt.Sprintf("%t", *istiodValues.MeshConfig.DefaultConfig.HoldApplicationUntilProxyStarts))
	}
}

// convertExitOnZeroActiveConnections 转换零连接时退出配置
func convertExitOnZeroActiveConnections(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	if istiodValues.MeshConfig != nil && istiodValues.MeshConfig.DefaultConfig != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.ExitOnZeroActiveConnections != nil {
		setFeatureConfig(result, common.FeatureExitOnZeroActiveConnections,
			*istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.ExitOnZeroActiveConnections)
	}
}

// convertIstioMetaDnsCapture 转换 DNS 捕获配置
func convertIstioMetaDnsCapture(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	if istiodValues.MeshConfig != nil && istiodValues.MeshConfig.DefaultConfig != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsCapture != nil {
		setFeatureConfig(result, common.FeatureIstioMetaDnsCapture,
			*istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsCapture)
	}
}

// convertIstioMetaDnsAutoAllocate 转换 DNS 自动分配配置
func convertIstioMetaDnsAutoAllocate(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	if istiodValues.MeshConfig != nil && istiodValues.MeshConfig.DefaultConfig != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsAutoAllocate != nil {
		setFeatureConfig(result, common.FeatureIstioMetaDnsAutoAllocate,
			*istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsAutoAllocate)
	}
}

// convertIstioMetaHttp10 转换 HTTP 1.0 支持配置
func convertIstioMetaHttp10(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	if istiodValues.Pilot != nil && istiodValues.Pilot.Env != nil {
		if http10, ok := istiodValues.Pilot.Env[common.EnvPilotHTTP10]; ok {
			setFeatureConfig(result, common.FeatureIstioMetaHttp10, http10)
		}
	}
}

// convertExcludeIPRanges 转换排除 IP 范围配置
func convertExcludeIPRanges(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioDetailInfo,
) {
	if istiodValues.Global != nil && istiodValues.Global.Proxy != nil &&
		istiodValues.Global.Proxy.ExcludeIPRanges != nil {
		setFeatureConfig(result, common.FeatureExcludeIPRanges, *istiodValues.Global.Proxy.ExcludeIPRanges)
	}
}
