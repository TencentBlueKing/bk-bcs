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

import "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"

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

func getBoolValue(m interface{}, key string) (bool, bool) {
	if val, exists := getMapValue(m, key); exists {
		if b, ok := val.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// processFieldKey 根据配置状态处理字段的删除或初始化
func processFieldKey(defaultValuesMap, customValuesMap map[string]interface{}) {
	// 处理 pilot 配置
	processPilotConfig(defaultValuesMap, customValuesMap)

	// 处理 meshConfig 配置
	processMeshConfig(defaultValuesMap, customValuesMap)
}

// processPilotConfig 处理 pilot 相关的配置
func processPilotConfig(defaultValuesMap, customValuesMap map[string]interface{}) {
	if customPilotConfig, ok := customValuesMap[common.FieldKeyPilot]; ok {
		// 处理 AutoscaleEnabled
		processPilotAutoscaleConfig(defaultValuesMap, customPilotConfig)

		// 处理 dedicatedNode 配置
		processPilotDedicatedNodeConfig(defaultValuesMap, customPilotConfig)
	}
}

// processPilotAutoscaleConfig 处理 pilot 的自动扩缩容配置
func processPilotAutoscaleConfig(defaultValuesMap map[string]interface{}, customPilotConfig interface{}) {
	if autoscaleEnabled, exists := getBoolValue(customPilotConfig, common.FieldKeyAutoscaleEnabled); exists {
		if !autoscaleEnabled {
			// 如果HPA被禁用，从defaultValues中删除对应的字段
			deletePilotAutoscaleFields(defaultValuesMap)
		} else {
			// 如果HPA被启用，确保相关字段在defaultValuesMap中存在
			ensurePilotAutoscaleFields(defaultValuesMap)
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

// ensurePilotAutoscaleFields 确保 pilot 自动扩缩容相关字段存在
func ensurePilotAutoscaleFields(defaultValuesMap map[string]interface{}) {
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		autoscaleFields := []string{
			common.FieldKeyAutoscaleMin,
			common.FieldKeyAutoscaleMax,
			common.FieldKeyCPU,
		}
		for _, field := range autoscaleFields {
			if _, exists := getMapValue(defaultPilotConfig, field); !exists {
				ensureMapKeyExists(defaultPilotConfig, field)
			}
		}
	}
}

// processPilotDedicatedNodeConfig 处理 pilot 的专属节点配置
func processPilotDedicatedNodeConfig(defaultValuesMap map[string]interface{}, customPilotConfig interface{}) {
	if dedicatedNode, ok := getMapValue(customPilotConfig, common.FieldKeyDedicatedNode); ok {
		if enabled, exists := getBoolValue(dedicatedNode, common.FieldKeyDedicatedNodeEnabled); exists {
			if !enabled {
				// 如果dedicatedNode被禁用，从defaultValues中删除对应的字段
				deletePilotDedicatedNodeFields(defaultValuesMap)
			} else {
				// 如果dedicatedNode被启用，确保相关字段在defaultValuesMap中存在
				ensurePilotDedicatedNodeFields(defaultValuesMap)
			}
		}
	}
}

// deletePilotDedicatedNodeFields 删除 pilot 专属节点相关字段
func deletePilotDedicatedNodeFields(defaultValuesMap map[string]interface{}) {
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		deleteMapKey(defaultPilotConfig, common.FieldKeyDedicatedNodeNodeSelector)
		deleteMapKey(defaultPilotConfig, common.FieldKeyDedicatedNodeTolerations)
	}
}

// ensurePilotDedicatedNodeFields 确保 pilot 专属节点相关字段存在
func ensurePilotDedicatedNodeFields(defaultValuesMap map[string]interface{}) {
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		dedicatedNodeFields := []string{
			common.FieldKeyDedicatedNodeNodeSelector,
			common.FieldKeyDedicatedNodeTolerations,
		}
		for _, field := range dedicatedNodeFields {
			if _, exists := getMapValue(defaultPilotConfig, field); !exists {
				ensureMapKeyExists(defaultPilotConfig, field)
			}
		}
	}
}

// processMeshConfig 处理 meshConfig 相关的配置
func processMeshConfig(defaultValuesMap, customValuesMap map[string]interface{}) {
	if customMeshConfig, ok := customValuesMap[common.FieldKeyMeshConfig]; ok {
		// 处理 LogCollectorConfigEnabled
		processMeshLogCollectorConfig(defaultValuesMap, customMeshConfig)

		// 处理 TracingConfigEnabled
		processMeshTracingConfig(defaultValuesMap, customMeshConfig)
	}
}

// processMeshLogCollectorConfig 处理 meshConfig 的日志采集配置
func processMeshLogCollectorConfig(defaultValuesMap map[string]interface{}, customMeshConfig interface{}) {
	logCollectorConfigEnabled, exists := getBoolValue(customMeshConfig, common.FieldKeyLogCollectorConfigEnabled)
	if exists {
		if !logCollectorConfigEnabled {
			// 如果日志采集被禁用，从defaultValues中删除对应的字段
			deleteMeshLogCollectorFields(defaultValuesMap)
		} else {
			// 如果日志采集被启用，确保相关字段在defaultValuesMap中存在
			ensureMeshLogCollectorFields(defaultValuesMap)
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

// ensureMeshLogCollectorFields 确保 meshConfig 日志采集相关字段存在
func ensureMeshLogCollectorFields(defaultValuesMap map[string]interface{}) {
	if defaultMeshConfig, ok := defaultValuesMap[common.FieldKeyMeshConfig]; ok {
		logCollectorFields := []string{
			common.FieldKeyAccessLogFile,
			common.FieldKeyAccessLogFormat,
			common.FieldKeyAccessLogEncoding,
		}
		for _, field := range logCollectorFields {
			if _, exists := getMapValue(defaultMeshConfig, field); !exists {
				ensureMapKeyExists(defaultMeshConfig, field)
			}
		}
	}
}

// processMeshTracingConfig 处理 meshConfig 的追踪配置
func processMeshTracingConfig(defaultValuesMap map[string]interface{}, customMeshConfig interface{}) {
	if enableTracing, exists := getBoolValue(customMeshConfig, common.FieldKeyEnableTracing); exists {
		if !enableTracing {
			// 如果tracing被禁用，从defaultValues中删除对应的字段
			deleteMeshTracingFields(defaultValuesMap)
		} else {
			// 如果tracing被启用，确保相关字段在defaultValuesMap中存在
			ensureMeshTracingFields(defaultValuesMap)
		}
	}
}

// deleteMeshTracingFields 删除 meshConfig 追踪相关字段
func deleteMeshTracingFields(defaultValuesMap map[string]interface{}) {
	// 删除 meshConfig 中的追踪字段
	if defaultMeshConfig, ok := defaultValuesMap[common.FieldKeyMeshConfig]; ok {
		// 删除 extensionProviders 字段（用于 OpenTelemetry）
		deleteMapKey(defaultMeshConfig, common.FieldKeyExtensionProviders)

		// 删除 defaultConfig.tracingConfig.zipkin 字段
		if defaultConfig, ok := getMapValue(defaultMeshConfig, common.FieldKeyDefaultConfig); ok {
			if tracingConfig, ok := getMapValue(defaultConfig, common.FieldKeyTracingConfig); ok {
				deleteMapKey(tracingConfig, common.FieldKeyZipkin)
			}
		}
	}

	// 删除 pilot 中的 traceSampling 字段
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		deleteMapKey(defaultPilotConfig, common.FieldKeyTraceSampling)
	}
}

// ensureMeshTracingFields 确保 meshConfig 追踪相关字段存在
func ensureMeshTracingFields(defaultValuesMap map[string]interface{}) {
	// 确保 meshConfig 中的追踪字段存在
	if defaultMeshConfig, ok := defaultValuesMap[common.FieldKeyMeshConfig]; ok {
		// 确保 extensionProviders 字段存在
		if _, exists := getMapValue(defaultMeshConfig, common.FieldKeyExtensionProviders); !exists {
			ensureMapKeyExists(defaultMeshConfig, common.FieldKeyExtensionProviders)
		}

		// 确保 defaultConfig.tracingConfig.zipkin 字段存在
		if defaultConfig, ok := getMapValue(defaultMeshConfig, common.FieldKeyDefaultConfig); ok {
			if tracingConfig, ok := getMapValue(defaultConfig, common.FieldKeyTracingConfig); ok {
				if _, exists := getMapValue(tracingConfig, common.FieldKeyZipkin); !exists {
					ensureMapKeyExists(tracingConfig, common.FieldKeyZipkin)
				}
			} else {
				// 如果 tracingConfig 不存在，则创建
				ensureMapKeyExists(defaultConfig, common.FieldKeyTracingConfig)
				if tracingConfig, ok := getMapValue(defaultConfig, common.FieldKeyTracingConfig); ok {
					ensureMapKeyExists(tracingConfig, common.FieldKeyZipkin)
				}
			}
		} else {
			// 如果 defaultConfig 不存在，创建整个路径
			ensureMapKeyExists(defaultMeshConfig, common.FieldKeyDefaultConfig)
			if defaultConfig, ok := getMapValue(defaultMeshConfig, common.FieldKeyDefaultConfig); ok {
				ensureMapKeyExists(defaultConfig, common.FieldKeyTracingConfig)
				if tracingConfig, ok := getMapValue(defaultConfig, common.FieldKeyTracingConfig); ok {
					ensureMapKeyExists(tracingConfig, common.FieldKeyZipkin)
				}
			}
		}
	}

	// 确保 pilot 中的 traceSampling 字段存在
	if defaultPilotConfig, ok := defaultValuesMap[common.FieldKeyPilot]; ok {
		if _, exists := getMapValue(defaultPilotConfig, common.FieldKeyTraceSampling); !exists {
			ensureMapKeyExists(defaultPilotConfig, common.FieldKeyTraceSampling)
		}
	}
}
