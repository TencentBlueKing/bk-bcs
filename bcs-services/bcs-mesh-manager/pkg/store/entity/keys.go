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

package entity

// Field keys for mesh entity
const (
	// ===== 基础信息 =====
	FieldKeyMeshID        = "meshID"
	FieldKeyNetworkID     = "networkID"
	FieldKeyProjectID     = "projectID"
	FieldKeyProjectCode   = "projectCode"
	FieldKeyDescription   = "description"
	FieldKeyChartVersion  = "chartVersion"
	FieldKeyStatus        = "status"
	FieldKeyStatusMessage = "statusMessage"
	FieldKeyCreateTime    = "createTime"
	FieldKeyUpdateTime    = "updateTime"
	FieldKeyCreateBy      = "createBy"
	FieldKeyUpdateBy      = "updateBy"
	FieldKeyIsDeleted     = "isDeleted"
	FieldKeyVersion       = "version"

	// ===== Mesh 配置 =====
	FieldKeyControlPlaneMode = "controlPlaneMode"
	FieldKeyClusterMode      = "clusterMode"
	FieldKeyPrimaryClusters  = "primaryClusters"
	FieldKeyRemoteClusters   = "remoteClusters"
	FieldKeyDifferentNetwork = "differentNetwork"

	// ===== 特性配置 =====
	FieldKeyFeatureConfigs = "featureConfigs"

	// ===== Sidecar 资源配置 =====
	FieldKeySidecarResourceConfig = "sidecarResourceConfig"
	FieldKeyCPURequest            = "cpuRequest"
	FieldKeyCPULimit              = "cpuLimit"
	FieldKeyMemoryRequest         = "memoryRequest"
	FieldKeyMemoryLimit           = "memoryLimit"

	// ===== 高可用配置 =====
	FieldKeyHighAvailability                   = "highAvailability"
	FieldKeyAutoscaleEnabled                   = "autoscaleEnabled"
	FieldKeyAutoscaleMin                       = "autoscaleMin"
	FieldKeyAutoscaleMax                       = "autoscaleMax"
	FieldKeyReplicaCount                       = "replicaCount"
	FieldKeyTargetCPUAverageUtilizationPercent = "targetCPUAverageUtilizationPercent"
	FieldKeyResourceConfig                     = "resourceConfig"
	FieldKeyDedicatedNode                      = "dedicatedNode"
	FieldKeyEnabled                            = "enabled"
	FieldKeyNodeLabels                         = "nodeLabels"

	// ===== 可观测性配置 =====
	FieldKeyObservabilityConfig        = "observabilityConfig"
	FieldKeyMetricsConfig              = "metricsConfig"
	FieldKeyLogCollectorConfig         = "logCollectorConfig"
	FieldKeyTracingConfig              = "tracing"
	FieldKeyTraceSamplingPercent       = "traceSamplingPercent"
	FieldKeyMetricsEnabled             = "metricsEnabled"
	FieldKeyControlPlaneMetricsEnabled = "controlPlaneMetricsEnabled"
	FieldKeyDataPlaneMetricsEnabled    = "dataPlaneMetricsEnabled"

	// ===== 日志收集配置 =====
	FieldKeyAccessLogEncoding = "accessLogEncoding"
	FieldKeyAccessLogFormat   = "accessLogFormat"

	// ===== 链路追踪配置 =====
	FieldKeyEndpoint = "endpoint"
	FieldKeyBkToken  = "bkToken"

	// ===== 特性配置字段 =====
	FieldKeyName            = "name"
	FieldKeyDefaultValue    = "defaultValue"
	FieldKeyAvailableValues = "availableValues"
	FieldKeySupportVersions = "supportVersions"

	// ===== 集群自定义Release名称 =====
	FieldKeyCustomReleaseNames = "customReleaseNames"
	// ===== 集群Release名称映射 =====
	FieldKeyReleaseNames = "releaseNames"
)

// Dot notation field keys for granular updates
const (
	// ===== Sidecar 资源配置点操作符 =====
	DotKeySidecarCPURequest    = "sidecarResourceConfig.cpuRequest"
	DotKeySidecarCPULimit      = "sidecarResourceConfig.cpuLimit"
	DotKeySidecarMemoryRequest = "sidecarResourceConfig.memoryRequest"
	DotKeySidecarMemoryLimit   = "sidecarResourceConfig.memoryLimit"

	// ===== 高可用配置点操作符 =====
	DotKeyHAAutoscaleEnabled                   = "highAvailability.autoscaleEnabled"
	DotKeyHAAutoscaleMin                       = "highAvailability.autoscaleMin"
	DotKeyHAAutoscaleMax                       = "highAvailability.autoscaleMax"
	DotKeyHAReplicaCount                       = "highAvailability.replicaCount"
	DotKeyHATargetCPUAverageUtilizationPercent = "highAvailability.targetCPUAverageUtilizationPercent"

	// ===== 高可用资源配置点操作符 =====
	DotKeyHAResourceCPURequest    = "highAvailability.resourceConfig.cpuRequest"
	DotKeyHAResourceCPULimit      = "highAvailability.resourceConfig.cpuLimit"
	DotKeyHAResourceMemoryRequest = "highAvailability.resourceConfig.memoryRequest"
	DotKeyHAResourceMemoryLimit   = "highAvailability.resourceConfig.memoryLimit"

	// ===== 高可用专用节点配置点操作符 =====
	DotKeyHADedicatedNodeEnabled    = "highAvailability.dedicatedNode.enabled"
	DotKeyHADedicatedNodeNodeLabels = "highAvailability.dedicatedNode.nodeLabels"

	// ===== 可观测性指标配置点操作符 =====
	DotKeyObsMetricsEnabled             = "observabilityConfig.metricsConfig.metricsEnabled"
	DotKeyObsMetricsControlPlaneEnabled = "observabilityConfig.metricsConfig.controlPlaneMetricsEnabled"
	DotKeyObsMetricsDataPlaneEnabled    = "observabilityConfig.metricsConfig.dataPlaneMetricsEnabled"

	// ===== 可观测性日志收集配置点操作符 =====
	DotKeyObsLogEnabled  = "observabilityConfig.logCollectorConfig.enabled"
	DotKeyObsLogEncoding = "observabilityConfig.logCollectorConfig.accessLogEncoding"
	DotKeyObsLogFormat   = "observabilityConfig.logCollectorConfig.accessLogFormat"

	// ===== 可观测性链路追踪配置点操作符 =====
	DotKeyObsTracingEnabled              = "observabilityConfig.tracingConfig.enabled"
	DotKeyObsTracingEndpoint             = "observabilityConfig.tracingConfig.endpoint"
	DotKeyObsTracingBkToken              = "observabilityConfig.tracingConfig.bkToken" //nolint:gosec
	DotKeyObsTracingTraceSamplingPercent = "observabilityConfig.tracingConfig.traceSamplingPercent"
)
