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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// UpdateValuesOptions istio更新时可选配置，用于处理values.yaml中的字段
type UpdateValuesOptions struct {
	LogCollectorConfigEnabled *bool
	AutoscaleEnabled          *bool
	DedicatedNodeEnabled      *bool
	EnableTracing             *bool
	// Sidecar 资源配置删除标志
	DeleteSidecarCpuRequest    bool
	DeleteSidecarMemoryRequest bool
	DeleteSidecarCpuLimit      bool
	DeleteSidecarMemoryLimit   bool
	// HighAvailability 资源配置删除标志
	DeleteHACpuRequest    bool
	DeleteHAMemoryRequest bool
	DeleteHACpuLimit      bool
	DeleteHAMemoryLimit   bool
}

func getMapValue(m interface{}, key string) (interface{}, bool) {
	switch mm := m.(type) {
	case map[string]interface{}:
		val, exists := mm[key]
		return val, exists
	case map[interface{}]interface{}:
		for k, v := range mm {
			if ks, ok := k.(string); ok && ks == key {
				return v, true
			}
		}
	}
	return nil, false
}

func deleteMapKey(m interface{}, key string) {
	switch mm := m.(type) {
	case map[string]interface{}:
		delete(mm, key)
	case map[interface{}]interface{}:
		for k := range mm {
			if ks, ok := k.(string); ok && ks == key {
				delete(mm, k)
				break
			}
		}
	}
}

func ensureMapKeyExists(m interface{}, key string) {
	switch mm := m.(type) {
	case map[string]interface{}:
		if _, exists := mm[key]; !exists {
			mm[key] = nil
		}
	case map[interface{}]interface{}:
		if _, exists := mm[key]; !exists {
			mm[key] = nil
		}
	}
}

// ProcessValues 根据配置状态处理字段的删除或初始化
func ProcessValues(values string, options *UpdateValuesOptions) (string, error) {
	if options == nil {
		return values, nil
	}

	var defaultValuesMap map[string]interface{}

	if err := yaml.Unmarshal([]byte(values), &defaultValuesMap); err != nil {
		blog.Errorf("unmarshal default values failed, err: %s", err)
		return values, err
	}

	// 处理 pilot 配置
	processPilotConfig(defaultValuesMap, options)

	// 处理 meshConfig 配置
	processMeshConfig(defaultValuesMap, options)

	// 处理资源相关配置
	processResourceConfig(defaultValuesMap, options)

	resultBytes, err := yaml.Marshal(defaultValuesMap)
	if err != nil {
		blog.Errorf("marshal processed values failed, err: %s", err)
		return values, err
	}

	return string(resultBytes), nil
}

// processPilotConfig 处理 pilot 相关的配置
func processPilotConfig(defaultValuesMap map[string]interface{}, options *UpdateValuesOptions) {
	// 处理 AutoscaleEnabled
	if options != nil && options.AutoscaleEnabled != nil {
		if !*options.AutoscaleEnabled {
			// 如果HPA被禁用，从defaultValues中删除对应的字段
			deletePilotAutoscaleFields(defaultValuesMap)
		}
	}

	// 处理 dedicatedNode 配置
	if options != nil && options.DedicatedNodeEnabled != nil {
		if !*options.DedicatedNodeEnabled {
			// 如果dedicatedNode被禁用，从defaultValues中删除对应的字段
			deletePilotDedicatedNodeFields(defaultValuesMap)
		}
	}
}

// deletePilotAutoscaleFields 删除 pilot 自动扩缩容相关字段
func deletePilotAutoscaleFields(defaultValuesMap map[string]interface{}) {
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		deleteMapKey(defaultPilotConfig, common.FieldKeyAutoscaleMin)
		deleteMapKey(defaultPilotConfig, common.FieldKeyAutoscaleMax)
		deleteMapKey(defaultPilotConfig, common.FieldKeyCPU)
	}
}

// deletePilotDedicatedNodeFields 删除 pilot 专属节点相关字段
func deletePilotDedicatedNodeFields(defaultValuesMap map[string]interface{}) {
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		deleteMapKey(defaultPilotConfig, common.FieldKeyDedicatedNodeNodeSelector)
		deleteMapKey(defaultPilotConfig, common.FieldKeyDedicatedNodeTolerations)
	}
}

// processMeshConfig 处理 meshConfig 相关的配置
func processMeshConfig(defaultValuesMap map[string]interface{}, options *UpdateValuesOptions) {
	// 处理 LogCollectorConfigEnabled
	if options != nil && options.LogCollectorConfigEnabled != nil {
		if !*options.LogCollectorConfigEnabled {
			// 如果日志采集被禁用，从defaultValues中删除对应的字段
			deleteMeshLogCollectorFields(defaultValuesMap)
		}
	}

	// 处理 TracingConfigEnabled
	if options != nil && options.EnableTracing != nil {
		if !*options.EnableTracing {
			// 如果tracing被禁用，从defaultValues中删除对应的字段
			deleteMeshTracingFields(defaultValuesMap)
		}
	}
}

// deleteMeshLogCollectorFields 删除 meshConfig 日志采集相关字段
func deleteMeshLogCollectorFields(defaultValuesMap map[string]interface{}) {
	if defaultMeshConfig, ok := defaultValuesMap[common.FieldKeyMeshConfig]; ok {
		logCollectorFields := []string{
			common.FieldKeyAccessLogFile,
			common.FieldKeyAccessLogFormat,
			common.FieldKeyAccessLogEncoding,
		}
		for _, field := range logCollectorFields {
			deleteMapKey(defaultMeshConfig, field)
		}
	}
}

// deleteMeshTracingFields 删除 meshConfig 追踪相关字段
func deleteMeshTracingFields(defaultValuesMap map[string]interface{}) {
	// 删除 meshConfig 中的追踪字段
	if defaultMeshConfig, ok := defaultValuesMap[common.FieldKeyMeshConfig]; ok {
		// 删除 extensionProviders 字段（用于 OpenTelemetry）
		deleteMapKey(defaultMeshConfig, common.FieldKeyExtensionProviders)

		// 删除 defaultConfig.tracingConfig 整个字段
		if defaultConfig, ok := getMapValue(defaultMeshConfig, common.FieldKeyDefaultConfig); ok {
			deleteMapKey(defaultConfig, common.FieldKeyTracingConfig)
		}
	}

	// 删除 pilot 中的 traceSampling 字段
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		deleteMapKey(defaultPilotConfig, common.FieldKeyTraceSampling)
	}
}

// processResourceConfig 处理资源相关的配置
func processResourceConfig(defaultValuesMap map[string]interface{}, options *UpdateValuesOptions) {
	if options == nil {
		return
	}

	// 处理 Sidecar 资源配置删除
	if options.DeleteSidecarCpuRequest {
		deleteSidecarResourceField(defaultValuesMap, common.FieldKeyCPU, common.FieldKeyRequests)
	}
	if options.DeleteSidecarMemoryRequest {
		deleteSidecarResourceField(defaultValuesMap, common.FieldKeyMemory, common.FieldKeyRequests)
	}
	if options.DeleteSidecarCpuLimit {
		deleteSidecarResourceField(defaultValuesMap, common.FieldKeyCPU, common.FieldKeyLimits)
	}
	if options.DeleteSidecarMemoryLimit {
		deleteSidecarResourceField(defaultValuesMap, common.FieldKeyMemory, common.FieldKeyLimits)
	}

	// 处理 HighAvailability 资源配置删除
	if options.DeleteHACpuRequest {
		deleteHAResourceField(defaultValuesMap, common.FieldKeyCPU, common.FieldKeyRequests)
	}
	if options.DeleteHAMemoryRequest {
		deleteHAResourceField(defaultValuesMap, common.FieldKeyMemory, common.FieldKeyRequests)
	}
	if options.DeleteHACpuLimit {
		deleteHAResourceField(defaultValuesMap, common.FieldKeyCPU, common.FieldKeyLimits)
	}
	if options.DeleteHAMemoryLimit {
		deleteHAResourceField(defaultValuesMap, common.FieldKeyMemory, common.FieldKeyLimits)
	}
}

// deleteSidecarResourceField 删除 Sidecar 指定的资源字段
func deleteSidecarResourceField(defaultValuesMap map[string]interface{}, resourceType, fieldType string) {
	// Sidecar 资源配置路径: global.proxy.resources.{requests|limits}.{cpu|memory}
	if globalConfig, ok := defaultValuesMap[common.FieldKeyGlobal]; ok {
		if proxyConfig, ok := getMapValue(globalConfig, common.FieldKeyProxy); ok {
			if resources, ok := getMapValue(proxyConfig, common.FieldKeyResources); ok {
				if fieldType == common.FieldKeyRequests {
					if requests, ok := getMapValue(resources, common.FieldKeyRequests); ok {
						deleteMapKey(requests, resourceType)
					}
				} else if fieldType == common.FieldKeyLimits {
					if limits, ok := getMapValue(resources, common.FieldKeyLimits); ok {
						deleteMapKey(limits, resourceType)
					}
				}
			}
		}
	}
}

// deleteHAResourceField 删除 HighAvailability 指定的资源字段
func deleteHAResourceField(defaultValuesMap map[string]interface{}, resourceType, fieldType string) {
	if pilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		if resources, ok := getMapValue(pilotConfig, common.FieldKeyResources); ok {
			if fieldType == common.FieldKeyRequests {
				if requests, ok := getMapValue(resources, common.FieldKeyRequests); ok {
					deleteMapKey(requests, resourceType)
				}
			} else if fieldType == common.FieldKeyLimits {
				if limits, ok := getMapValue(resources, common.FieldKeyLimits); ok {
					deleteMapKey(limits, resourceType)
				}
			}
		}
	}
}
