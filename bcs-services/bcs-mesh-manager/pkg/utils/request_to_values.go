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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	pointer "k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

const (
	trueValue  = "true"
	falseValue = "false"
)

// ConvertRequestToValues 从 IstioRequest 构建 IstiodInstallValues 部署配置
func ConvertRequestToValues(req *meshmanager.IstioRequest) *common.IstiodInstallValues {
	installValues := &common.IstiodInstallValues{}

	// 构建基础配置
	buildBasicConfig(req, installValues)

	// 构建Sidecar资源配置
	buildSidecarResourceConfig(req, installValues)

	// 构建高可用配置
	buildHighAvailabilityConfig(req, installValues)

	// 构建功能特性配置
	buildFeatureConfigs(req, installValues)

	// 构建可观测性配置
	buildObservabilityConfig(req, installValues)

	return installValues
}

// buildBasicConfig 构建基础配置
func buildBasicConfig(
	req *meshmanager.IstioRequest,
	installValues *common.IstiodInstallValues,
) {
	// 构建Global配置
	if installValues.Global == nil {
		installValues.Global = &common.IstiodGlobalConfig{}
	}

	// 构建MultiCluster配置
	if len(req.PrimaryClusters) > 0 {
		if installValues.MultiCluster == nil {
			installValues.MultiCluster = &common.IstiodMultiClusterConfig{}
		}
		clusterName := strings.ToLower(req.PrimaryClusters[0])
		installValues.MultiCluster.ClusterName = &clusterName
	}
}

// buildSidecarResourceConfig 构建Sidecar资源配置
func buildSidecarResourceConfig(req *meshmanager.IstioRequest, installValues *common.IstiodInstallValues) {
	if req.SidecarResourceConfig == nil {
		return
	}

	if installValues.Global == nil {
		installValues.Global = &common.IstiodGlobalConfig{}
	}
	if installValues.Global.Proxy == nil {
		installValues.Global.Proxy = &common.IstioProxyConfig{}
	}
	if installValues.Global.Proxy.Resources == nil {
		installValues.Global.Proxy.Resources = &v1.ResourceRequirements{
			Requests: make(v1.ResourceList),
			Limits:   make(v1.ResourceList),
		}
	}

	// 设置CPU请求
	if req.SidecarResourceConfig.CpuRequest != nil && req.SidecarResourceConfig.CpuRequest.GetValue() != "" {
		installValues.Global.Proxy.Resources.Requests[v1.ResourceCPU] =
			resource.MustParse(req.SidecarResourceConfig.CpuRequest.GetValue())
	}

	// 设置CPU限制
	if req.SidecarResourceConfig.CpuLimit != nil && req.SidecarResourceConfig.CpuLimit.GetValue() != "" {
		installValues.Global.Proxy.Resources.Limits[v1.ResourceCPU] =
			resource.MustParse(req.SidecarResourceConfig.CpuLimit.GetValue())
	}

	// 设置内存请求
	if req.SidecarResourceConfig.MemoryRequest != nil && req.SidecarResourceConfig.MemoryRequest.GetValue() != "" {
		installValues.Global.Proxy.Resources.Requests[v1.ResourceMemory] =
			resource.MustParse(req.SidecarResourceConfig.MemoryRequest.GetValue())
	}

	// 设置内存限制
	if req.SidecarResourceConfig.MemoryLimit != nil && req.SidecarResourceConfig.MemoryLimit.GetValue() != "" {
		installValues.Global.Proxy.Resources.Limits[v1.ResourceMemory] =
			resource.MustParse(req.SidecarResourceConfig.MemoryLimit.GetValue())
	}
}

// buildHighAvailabilityConfig 构建高可用配置
func buildHighAvailabilityConfig(req *meshmanager.IstioRequest, installValues *common.IstiodInstallValues) {
	if req.HighAvailability == nil {
		return
	}

	if installValues.Pilot == nil {
		installValues.Pilot = &common.IstiodPilotConfig{}
	}

	// 设置副本数
	if req.HighAvailability.ReplicaCount != nil {
		installValues.Pilot.ReplicaCount = pointer.Int32(req.HighAvailability.ReplicaCount.GetValue())
	}

	// 设置HPA配置
	if req.HighAvailability.AutoscaleEnabled != nil {
		if req.HighAvailability.AutoscaleEnabled.GetValue() {
			installValues.Pilot.AutoscaleEnabled = pointer.Bool(true)

			if req.HighAvailability.AutoscaleMin != nil {
				installValues.Pilot.AutoscaleMin = pointer.Int32(req.HighAvailability.AutoscaleMin.GetValue())
			}
			if req.HighAvailability.AutoscaleMax != nil {
				installValues.Pilot.AutoscaleMax = pointer.Int32(req.HighAvailability.AutoscaleMax.GetValue())
			}
			if req.HighAvailability.TargetCPUAverageUtilizationPercent != nil {
				installValues.Pilot.CPU = &common.HPACPUConfig{
					TargetAverageUtilization: pointer.Int32(req.HighAvailability.TargetCPUAverageUtilizationPercent.GetValue()),
				}
			}
		} else {
			installValues.Pilot.AutoscaleEnabled = pointer.Bool(false)
		}
	}

	// 设置Pilot资源配置
	if req.HighAvailability.ResourceConfig != nil {
		if installValues.Pilot.Resources == nil {
			installValues.Pilot.Resources = &v1.ResourceRequirements{
				Requests: make(v1.ResourceList),
				Limits:   make(v1.ResourceList),
			}
		}

		if req.HighAvailability.ResourceConfig.CpuRequest != nil &&
			req.HighAvailability.ResourceConfig.CpuRequest.GetValue() != "" {
			installValues.Pilot.Resources.Requests[v1.ResourceCPU] =
				resource.MustParse(req.HighAvailability.ResourceConfig.CpuRequest.GetValue())
		}
		if req.HighAvailability.ResourceConfig.CpuLimit != nil &&
			req.HighAvailability.ResourceConfig.CpuLimit.GetValue() != "" {
			installValues.Pilot.Resources.Limits[v1.ResourceCPU] =
				resource.MustParse(req.HighAvailability.ResourceConfig.CpuLimit.GetValue())
		}
		if req.HighAvailability.ResourceConfig.MemoryRequest != nil &&
			req.HighAvailability.ResourceConfig.MemoryRequest.GetValue() != "" {
			installValues.Pilot.Resources.Requests[v1.ResourceMemory] =
				resource.MustParse(req.HighAvailability.ResourceConfig.MemoryRequest.GetValue())
		}
		if req.HighAvailability.ResourceConfig.MemoryLimit != nil &&
			req.HighAvailability.ResourceConfig.MemoryLimit.GetValue() != "" {
			installValues.Pilot.Resources.Limits[v1.ResourceMemory] =
				resource.MustParse(req.HighAvailability.ResourceConfig.MemoryLimit.GetValue())
		}
	}

	// 设置专属节点配置
	if req.HighAvailability.DedicatedNode != nil && req.HighAvailability.DedicatedNode.Enabled != nil {
		if req.HighAvailability.DedicatedNode.Enabled.GetValue() &&
			req.HighAvailability.DedicatedNode.NodeLabels != nil {
			installValues.Pilot.NodeSelector = req.HighAvailability.DedicatedNode.NodeLabels
		}
	}
}

// buildFeatureConfigs 构建功能特性配置
func buildFeatureConfigs(req *meshmanager.IstioRequest, installValues *common.IstiodInstallValues) {
	if req.FeatureConfigs == nil {
		return
	}

	for featureName, featureConfig := range req.FeatureConfigs {
		switch featureName {
		case common.FeatureOutboundTrafficPolicy:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			installValues.MeshConfig.OutboundTrafficPolicy = &common.OutboundTrafficPolicy{
				Mode: pointer.String(featureConfig.Value),
			}

		case common.FeatureHoldApplicationUntilProxyStarts:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			installValues.MeshConfig.DefaultConfig.HoldApplicationUntilProxyStarts =
				pointer.Bool(featureConfig.Value == trueValue)

		case common.FeatureExitOnZeroActiveConnections:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installValues.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installValues.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.ExitOnZeroActiveConnections =
				pointer.Bool(featureConfig.Value == trueValue)

		case common.FeatureIstioMetaDnsCapture:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installValues.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installValues.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsCapture =
				pointer.String(featureConfig.Value)

		case common.FeatureIstioMetaDnsAutoAllocate:
			if installValues.MeshConfig == nil {
				installValues.MeshConfig = &common.IstiodMeshConfig{}
			}
			if installValues.MeshConfig.DefaultConfig == nil {
				installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
			}
			if installValues.MeshConfig.DefaultConfig.ProxyMetadata == nil {
				installValues.MeshConfig.DefaultConfig.ProxyMetadata = &common.ProxyMetadata{}
			}
			installValues.MeshConfig.DefaultConfig.ProxyMetadata.IstioMetaDnsAutoAllocate =
				pointer.String(featureConfig.Value)

		case common.FeatureIstioMetaHttp10:
			if installValues.Pilot == nil {
				installValues.Pilot = &common.IstiodPilotConfig{}
			}
			if installValues.Pilot.Env == nil {
				installValues.Pilot.Env = make(map[string]string)
			}
			installValues.Pilot.Env[common.EnvPilotHTTP10] = featureConfig.Value

		case common.FeatureExcludeIPRanges:
			if installValues.Global == nil {
				installValues.Global = &common.IstiodGlobalConfig{}
			}
			if installValues.Global.Proxy == nil {
				installValues.Global.Proxy = &common.IstioProxyConfig{}
			}
			installValues.Global.Proxy.ExcludeIPRanges = pointer.String(featureConfig.Value)

		default:
			blog.Warnf("unknown feature config: %s", featureName)
		}
	}
}

// buildObservabilityConfig 构建可观测性配置
func buildObservabilityConfig(req *meshmanager.IstioRequest, installValues *common.IstiodInstallValues) {
	if req.ObservabilityConfig == nil {
		return
	}

	if installValues.MeshConfig == nil {
		installValues.MeshConfig = &common.IstiodMeshConfig{}
	}

	// 构建日志采集配置
	if req.ObservabilityConfig.LogCollectorConfig != nil {
		if req.ObservabilityConfig.LogCollectorConfig.Enabled != nil {
			if req.ObservabilityConfig.LogCollectorConfig.Enabled.GetValue() {
				installValues.MeshConfig.AccessLogFile = pointer.String(common.AccessLogFileStdout)

				if req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat != nil {
					installValues.MeshConfig.AccessLogFormat =
						pointer.String(req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue())
				}
				if req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding != nil {
					installValues.MeshConfig.AccessLogEncoding =
						pointer.String(req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue())
				}
			} else {
				// 如果禁用日志采集，清空相关配置
				installValues.MeshConfig.AccessLogFile = nil
				installValues.MeshConfig.AccessLogFormat = nil
				installValues.MeshConfig.AccessLogEncoding = nil
			}
		}
	}

	// 构建全链路追踪配置
	if req.ObservabilityConfig.TracingConfig != nil {
		if req.ObservabilityConfig.TracingConfig.Enabled != nil {
			if req.ObservabilityConfig.TracingConfig.Enabled.GetValue() {
				installValues.MeshConfig.EnableTracing = pointer.Bool(true)

				// 构建Zipkin配置
				if req.ObservabilityConfig.TracingConfig.Endpoint != nil &&
					req.ObservabilityConfig.TracingConfig.Endpoint.GetValue() != "" {
					if installValues.MeshConfig.DefaultConfig == nil {
						installValues.MeshConfig.DefaultConfig = &common.DefaultConfig{}
					}
					installValues.MeshConfig.DefaultConfig.TracingConfig = &common.TracingConfig{
						Zipkin: &common.ZipkinConfig{
							Address: pointer.String(req.ObservabilityConfig.TracingConfig.Endpoint.GetValue()),
						},
					}
				}
			} else {
				installValues.MeshConfig.EnableTracing = pointer.Bool(false)
			}
		}
	}
}
