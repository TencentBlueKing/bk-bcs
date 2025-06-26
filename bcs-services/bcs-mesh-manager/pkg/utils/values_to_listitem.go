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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ConvertValuesToListItem 将 MeshIstio 实体和 IstiodInstallValues 配置转换为 IstioListItem 展示结构
func ConvertValuesToListItem(
	meshIstio *entity.MeshIstio,
	istiodValues *common.IstiodInstallValues,
) (*meshmanager.IstioListItem, error) {
	if meshIstio == nil {
		blog.Errorf("meshIstio is nil")
		return nil, fmt.Errorf("meshIstio is nil")
	}
	if istiodValues == nil {
		blog.Errorf("istiodValues is nil")
		return nil, fmt.Errorf("istiodValues is nil")
	}
	blog.Infof("istiodValues: %+v", istiodValues)

	// 使用 Transfer2Proto 方法进行基础转换
	result := meshIstio.Transfer2Proto()

	// 从实际的部署配置中提取资源配置
	if istiodValues.Global != nil &&
		istiodValues.Global.Proxy != nil &&
		istiodValues.Global.Proxy.Resources != nil {
		result.SidecarResourceConfig = convertResourceConfigValues(istiodValues.Global.Proxy.Resources)
	}

	// 从实际的部署配置中提取高可用配置
	if istiodValues.Pilot != nil {
		convertHighAvailabilityValues(istiodValues.Pilot, result)
	}

	// 从实际的部署配置中提取可观测性配置
	if istiodValues.MeshConfig != nil {
		convertObservabilityConfigValues(istiodValues.MeshConfig, result)
	}

	// 从实际的部署配置中提取功能特性配置
	convertFeatureConfigs(istiodValues, result)

	// TODO: 以下 istiodValues 字段目前未在 BuildIstioListItem 中使用：
	// - istiodValues.IstiodRemote (远程 istiod 配置)
	// - istiodValues.Telemetry (遥测配置)
	// - istiodValues.Global.MeshID (网格ID)
	// - istiodValues.Global.Network (网络)
	// - istiodValues.Global.ConfigCluster (配置集群)
	// - istiodValues.Global.OmitSidecarInjectorConfigMap (是否省略 sidecar 注入器配置映射)
	// - istiodValues.Global.RemotePilotAddress (远程 pilot 地址)
	// - istiodValues.Global.ExternalIstiod (外部 istiod)
	// - istiodValues.MultiCluster (多集群配置)
	// - istiodValues.Pilot.TraceSampling (追踪采样)
	// - istiodValues.Pilot.ConfigMap (配置映射)
	// - istiodValues.Pilot.Env 中的其他环境变量（除了 PILOT_HTTP10）

	return result, nil
}

// convertResourceConfigValues 从实际的资源配置构建 ResourceConfig
func convertResourceConfigValues(
	resources *v1.ResourceRequirements,
) *meshmanager.ResourceConfig {
	config := &meshmanager.ResourceConfig{}

	if resources.Requests != nil {
		if cpu, ok := resources.Requests[v1.ResourceCPU]; ok {
			config.CpuRequest = wrapperspb.String(cpu.String())
		}
		if memory, ok := resources.Requests[v1.ResourceMemory]; ok {
			config.MemoryRequest = wrapperspb.String(memory.String())
		}
	}

	if resources.Limits != nil {
		if cpu, ok := resources.Limits[v1.ResourceCPU]; ok {
			config.CpuLimit = wrapperspb.String(cpu.String())
		}
		if memory, ok := resources.Limits[v1.ResourceMemory]; ok {
			config.MemoryLimit = wrapperspb.String(memory.String())
		}
	}

	return config
}

// updateResourceConfigValues 从实际的资源配置更新现有的 ResourceConfig
func updateResourceConfigValues(
	resources *v1.ResourceRequirements,
	config *meshmanager.ResourceConfig,
) {
	if config == nil {
		return
	}

	if resources.Requests != nil {
		if cpu, ok := resources.Requests[v1.ResourceCPU]; ok {
			config.CpuRequest = wrapperspb.String(cpu.String())
		}
		if memory, ok := resources.Requests[v1.ResourceMemory]; ok {
			config.MemoryRequest = wrapperspb.String(memory.String())
		}
	}

	if resources.Limits != nil {
		if cpu, ok := resources.Limits[v1.ResourceCPU]; ok {
			config.CpuLimit = wrapperspb.String(cpu.String())
		}
		if memory, ok := resources.Limits[v1.ResourceMemory]; ok {
			config.MemoryLimit = wrapperspb.String(memory.String())
		}
	}
}

// convertHighAvailabilityValues 从实际的高可用配置更新 HighAvailability
func convertHighAvailabilityValues(
	pilot *common.IstiodPilotConfig,
	result *meshmanager.IstioListItem,
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
	meshConfig *common.IstiodMeshConfig,
	result *meshmanager.IstioListItem,
) {
	// 确保 result.ObservabilityConfig 存在
	if result.ObservabilityConfig == nil {
		result.ObservabilityConfig = &meshmanager.ObservabilityConfig{}
	}

	// 更新追踪配置
	if meshConfig.EnableTracing != nil && *meshConfig.EnableTracing {
		if result.ObservabilityConfig.TracingConfig == nil {
			result.ObservabilityConfig.TracingConfig = &meshmanager.TracingConfig{}
		}
		result.ObservabilityConfig.TracingConfig.Enabled = wrapperspb.Bool(true)

		// 更新追踪端点
		if meshConfig.DefaultConfig != nil && meshConfig.DefaultConfig.TracingConfig != nil &&
			meshConfig.DefaultConfig.TracingConfig.Zipkin != nil &&
			meshConfig.DefaultConfig.TracingConfig.Zipkin.Address != nil {
			result.ObservabilityConfig.TracingConfig.Endpoint =
				wrapperspb.String(*meshConfig.DefaultConfig.TracingConfig.Zipkin.Address)
		}
	}

	// 更新日志配置
	if meshConfig.AccessLogFile != nil {
		if result.ObservabilityConfig.LogCollectorConfig == nil {
			result.ObservabilityConfig.LogCollectorConfig = &meshmanager.LogCollectorConfig{}
		}
		result.ObservabilityConfig.LogCollectorConfig.Enabled = wrapperspb.Bool(true)

		// 更新日志格式
		if meshConfig.AccessLogFormat != nil {
			result.ObservabilityConfig.LogCollectorConfig.AccessLogFormat = wrapperspb.String(*meshConfig.AccessLogFormat)
		}
		// 更新日志编码
		if meshConfig.AccessLogEncoding != nil {
			result.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding = wrapperspb.String(*meshConfig.AccessLogEncoding)
		}
	}
}

// convertFeatureConfigs 从实际的功能特性配置更新 FeatureConfigs
func convertFeatureConfigs(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioListItem,
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
	result *meshmanager.IstioListItem,
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
	result *meshmanager.IstioListItem,
) {
	if istiodValues.MeshConfig != nil && istiodValues.MeshConfig.OutboundTrafficPolicy != nil &&
		istiodValues.MeshConfig.OutboundTrafficPolicy.Mode != nil {
		setFeatureConfig(result, common.FeatureOutboundTrafficPolicy, *istiodValues.MeshConfig.OutboundTrafficPolicy.Mode)
	}
}

// convertHoldApplicationUntilProxyStarts 转换等待代理启动配置
func convertHoldApplicationUntilProxyStarts(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioListItem,
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
	result *meshmanager.IstioListItem,
) {
	if istiodValues.MeshConfig != nil && istiodValues.MeshConfig.DefaultConfig != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata != nil &&
		istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.ExitOnZeroActiveConnections != nil {
		setFeatureConfig(result, common.FeatureExitOnZeroActiveConnections,
			fmt.Sprintf("%t", *istiodValues.MeshConfig.DefaultConfig.ProxyMetadata.ExitOnZeroActiveConnections))
	}
}

// convertIstioMetaDnsCapture 转换 DNS 捕获配置
func convertIstioMetaDnsCapture(
	istiodValues *common.IstiodInstallValues,
	result *meshmanager.IstioListItem,
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
	result *meshmanager.IstioListItem,
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
	result *meshmanager.IstioListItem,
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
	result *meshmanager.IstioListItem,
) {
	if istiodValues.Global != nil && istiodValues.Global.Proxy != nil &&
		istiodValues.Global.Proxy.ExcludeIPRanges != nil {
		setFeatureConfig(result, common.FeatureExcludeIPRanges, *istiodValues.Global.Proxy.ExcludeIPRanges)
	}
}
