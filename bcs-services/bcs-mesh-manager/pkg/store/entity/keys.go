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
	FieldKeyMeshName      = "meshName"
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
	FieldKeyObservabilityConfig  = "observabilityConfig"
	FieldKeyMetricsConfig        = "metricsConfig"
	FieldKeyLogCollectorConfig   = "logCollectorConfig"
	FieldKeyTracingConfig        = "tracingConfig"
	FieldKeyTraceSamplingPercent = "traceSamplingPercent"

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
)
