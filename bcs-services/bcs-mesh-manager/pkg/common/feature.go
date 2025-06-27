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
	"google.golang.org/protobuf/types/known/wrapperspb"

	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

const (
	// FeatureOutboundTrafficPolicy 出站流量策略
	FeatureOutboundTrafficPolicy = "outboundTrafficPolicy"
	// FeatureHoldApplicationUntilProxyStarts 应用等待 sidecar 启动
	FeatureHoldApplicationUntilProxyStarts = "holdApplicationUntilProxyStarts"
	// FeatureExitOnZeroActiveConnections 无活动连接时退出
	FeatureExitOnZeroActiveConnections = "exitOnZeroActiveConnections"
	// FeatureExcludeIPRanges 排除IP范围
	FeatureExcludeIPRanges = "excludeIPRanges"
	// FeatureIstioMetaDnsCapture DNS转发
	FeatureIstioMetaDnsCapture = "istioMetaDnsCapture"
	// FeatureIstioMetaDnsAutoAllocate 自动分配IP
	FeatureIstioMetaDnsAutoAllocate = "istioMetaDnsAutoAllocate"
	// FeatureIstioMetaHttp10 是否支持HTTP/1.0
	FeatureIstioMetaHttp10 = "istioMetaHttp10"
)

// SupportedFeatures 支持的功能列表
var SupportedFeatures = []string{
	FeatureOutboundTrafficPolicy,
	FeatureHoldApplicationUntilProxyStarts,
	FeatureExitOnZeroActiveConnections,
	FeatureExcludeIPRanges,
	FeatureIstioMetaDnsCapture,
	FeatureIstioMetaDnsAutoAllocate,
	FeatureIstioMetaHttp10,
}

// DefaultFeatureConfigTemplate 默认特性配置模板
type DefaultFeatureConfigTemplate struct {
	Name                string
	Description         string
	DefaultValue        string
	AvailableValues     []string
	SupportIstioVersion string // semver
}

// GetDefaultFeatureConfigs 获取默认特性配置模板
func GetDefaultFeatureConfigs() map[string]*DefaultFeatureConfigTemplate {
	return map[string]*DefaultFeatureConfigTemplate{
		FeatureOutboundTrafficPolicy: {
			Name:            FeatureOutboundTrafficPolicy,
			Description:     "出站流量策略配置",
			DefaultValue:    "ALLOW_ANY",
			AvailableValues: []string{"ALLOW_ANY", "REGISTRY_ONLY"},
		},
		FeatureHoldApplicationUntilProxyStarts: {
			Name:            FeatureHoldApplicationUntilProxyStarts,
			Description:     "Sidecar 就绪保障",
			DefaultValue:    "false",
			AvailableValues: []string{"true", "false"},
		},
		FeatureExitOnZeroActiveConnections: {
			Name:                FeatureExitOnZeroActiveConnections,
			Description:         "Sidecar 停止保障",
			DefaultValue:        "false",
			AvailableValues:     []string{"true", "false"},
			SupportIstioVersion: ">=1.12",
		},
		FeatureExcludeIPRanges: {
			Name:            FeatureExcludeIPRanges,
			Description:     "排除IP范围配置",
			DefaultValue:    "",
			AvailableValues: []string{},
		},
		FeatureIstioMetaDnsCapture: {
			Name:            FeatureIstioMetaDnsCapture,
			Description:     "DNS转发",
			DefaultValue:    "false",
			AvailableValues: []string{"true", "false"},
		},
		FeatureIstioMetaDnsAutoAllocate: {
			Name:            FeatureIstioMetaDnsAutoAllocate,
			Description:     "自动分配IP配置",
			DefaultValue:    "false",
			AvailableValues: []string{"true", "false"},
		},
		FeatureIstioMetaHttp10: {
			Name:            FeatureIstioMetaHttp10,
			Description:     "是否支持HTTP/1.0",
			DefaultValue:    "false",
			AvailableValues: []string{"true", "false"},
		},
	}
}

// GetDefaultSidecarResourceConfig 获取默认Sidecar资源配置
func GetDefaultSidecarResourceConfig() *meshmanager.ResourceConfig {
	return &meshmanager.ResourceConfig{
		CpuRequest:    wrapperspb.String("100m"),
		CpuLimit:      wrapperspb.String("2000m"),
		MemoryRequest: wrapperspb.String("128Mi"),
		MemoryLimit:   wrapperspb.String("1024Mi"),
	}
}

// GetDefaultHighAvailabilityConfig 获取默认高可用配置
func GetDefaultHighAvailabilityConfig() *meshmanager.HighAvailability {
	return &meshmanager.HighAvailability{
		AutoscaleEnabled:                   wrapperspb.Bool(false),
		AutoscaleMin:                       wrapperspb.Int32(1),
		AutoscaleMax:                       wrapperspb.Int32(5),
		ReplicaCount:                       wrapperspb.Int32(2),
		TargetCPUAverageUtilizationPercent: wrapperspb.Int32(80),
		ResourceConfig: &meshmanager.ResourceConfig{
			CpuRequest:    wrapperspb.String("500m"),
			MemoryRequest: wrapperspb.String("2048Mi"),
		},
		DedicatedNode: &meshmanager.DedicatedNode{
			Enabled:    wrapperspb.Bool(false),
			NodeLabels: map[string]string{},
		},
	}
}
