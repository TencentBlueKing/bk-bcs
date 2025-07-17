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

	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/apimachinery/pkg/api/resource"

	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// NormalizeResourceConfigUnits 将资源配置转换为标准单位格式
// CPU: 统一为 m (毫核)
// Memory: 统一为 Mi (二进制兆字节)
func NormalizeResourceConfigUnits(config *meshmanager.ResourceConfig) error {
	if config == nil {
		return nil
	}

	// 转换 CPU 相关字段
	if config.CpuRequest != nil {
		convertedCPU := convertCPUToMilliCores(config.CpuRequest.GetValue(), true)
		config.CpuRequest = wrapperspb.String(convertedCPU)
	}

	if config.CpuLimit != nil {
		convertedCPU := convertCPUToMilliCores(config.CpuLimit.GetValue(), false)
		config.CpuLimit = wrapperspb.String(convertedCPU)
	}

	// 转换内存相关字段
	if config.MemoryRequest != nil {
		convertedMemory := convertMemoryToMi(config.MemoryRequest.GetValue(), true)
		config.MemoryRequest = wrapperspb.String(convertedMemory)
	}

	if config.MemoryLimit != nil {
		convertedMemory := convertMemoryToMi(config.MemoryLimit.GetValue(), false)
		config.MemoryLimit = wrapperspb.String(convertedMemory)
	}

	return nil
}

// NormalizeHighAvailabilityResource 标准化高可用配置中的资源配置
func NormalizeHighAvailabilityResource(ha *meshmanager.HighAvailability) error {
	if ha == nil || ha.ResourceConfig == nil {
		return nil
	}
	return NormalizeResourceConfigUnits(ha.ResourceConfig)
}

// NormalizeResourcesConfig 标准化 IstioDetailInfo 中所有资源配置的单位
func NormalizeResourcesConfig(detailInfo *meshmanager.IstioDetailInfo) error {
	if detailInfo == nil {
		return nil
	}

	// 标准化 Sidecar 资源配置
	if detailInfo.SidecarResourceConfig != nil {
		if err := NormalizeResourceConfigUnits(detailInfo.SidecarResourceConfig); err != nil {
			return err
		}
	}

	// 标准化高可用资源配置
	if err := NormalizeHighAvailabilityResource(detailInfo.HighAvailability); err != nil {
		return err
	}

	return nil
}

// convertCPUToMilliCores 将 CPU 值转换为毫核格式
// 输入可能是: "1", "0.5", "1000m", "500m"
// 输出统一为: "1000m", "500m", "1000m", "500m"
func convertCPUToMilliCores(cpuValue string, isRequest bool) string {
	if cpuValue == "" {
		return ""
	}

	// 解析数值
	quantity, err := resource.ParseQuantity(cpuValue)
	if err != nil {
		// 转换失败时返回原始值
		return cpuValue
	}

	// 检查是否为0
	if quantity.IsZero() {
		if !isRequest {
			return "" // limit 为0时返回空字符串
		}
	}

	// 检查是否为负数
	if quantity.Sign() < 0 {
		// 负数时直接返回原始值
		return cpuValue
	}

	milliValue := quantity.MilliValue()
	return fmt.Sprintf("%dm", milliValue)
}

// convertMemoryToMi 将内存值转换为 Mi 格式
// 输入可能是: "1Gi", "1000Mi", "1G", "1000M", "1024Ki"
// 输出统一为: "1024Mi", "1000Mi", "1000Mi", "1000Mi", "1Mi"
func convertMemoryToMi(memoryValue string, isRequest bool) string {
	if memoryValue == "" {
		return ""
	}

	// 解析数值
	quantity, err := resource.ParseQuantity(memoryValue)
	if err != nil {
		// 转换失败时直接返回原始值
		return memoryValue
	}

	// 检查是否为0
	if quantity.IsZero() {
		if !isRequest {
			return ""
		}
	}

	// 检查是否为负数
	if quantity.Sign() < 0 {
		// 负数时直接返回原始值
		return memoryValue
	}

	// 1 Mi = 1024 * 1024 bytes
	bytes := quantity.Value()
	miValue := bytes / (1024 * 1024)

	// 如果有余数，保留精度
	if bytes%(1024*1024) != 0 {
		miFloat := float64(bytes) / (1024 * 1024)
		return fmt.Sprintf("%.2fMi", miFloat)
	}

	return fmt.Sprintf("%dMi", miValue)
}
