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
	// Basic mesh information
	FieldKeyMeshID     = "meshID"
	FieldKeyMeshName   = "meshName"
	FieldKeyStatus     = "status"
	FieldKeyCreateTime = "createTime"
	FieldKeyUpdateTime = "updateTime"
	FieldKeyCreateBy   = "createBy"
	FieldKeyUpdateBy   = "updateBy"

	// Mesh metadata
	FieldKeyChartVersion     = "chartVersion"
	FieldKeyClusters         = "clusters"
	FieldKeyFeatureConfigs   = "featureConfigs"
	FieldKeyServiceDiscovery = "serviceDiscovery"

	// Feature configuration keys
	FieldKeyMeshMode                = "meshMode"
	FieldKeyEgressTrafficMode       = "egressTrafficMode"
	FieldKeySidecarAutoInjection    = "sidecarAutoInjection"
	FieldKeySidecarBypassIPs        = "sidecarBypassIPs"
	FieldKeySidecarReadinessProbe   = "sidecarReadinessProbe"
	FieldKeySidecarTerminationGrace = "sidecarTerminationGrace"
	FieldKeyIngressGateway          = "ingressGateway"
	FieldKeyEgressGateway           = "egressGateway"
	FieldKeyMonitoring              = "monitoring"
	FieldKeyTracing                 = "tracing"
	FieldKeyLogging                 = "logging"

	// Service discovery fields
	FieldKeyAutoInjectionNamespaces = "autoInjectionNamespaces"
	FieldKeyDisabledInjectionPods   = "disabledInjectionPods"

	// Sidecar configuration fields
	FieldKeySidecarEnabled      = "enabled"
	FieldKeySidecarResources    = "resources"
	FieldKeySidecarBypassIPList = "ipList"
	FieldKeySidecarGracePeriod  = "gracePeriod"

	// Common resource fields
	FieldKeyCPURequest    = "cpuRequest"
	FieldKeyCPULimit      = "cpuLimit"
	FieldKeyMemoryRequest = "memoryRequest"
	FieldKeyMemoryLimit   = "memoryLimit"

	// Common service fields
	FieldKeyServiceType           = "type"
	FieldKeyServicePorts          = "ports"
	FieldKeyServicePortName       = "name"
	FieldKeyServicePortPort       = "port"
	FieldKeyServicePortTargetPort = "targetPort"
	FieldKeyServicePortProtocol   = "protocol"

	// Common deployment fields
	FieldKeyDeploymentReplicas  = "replicas"
	FieldKeyDeploymentImage     = "image"
	FieldKeyDeploymentResources = "resources"
)
