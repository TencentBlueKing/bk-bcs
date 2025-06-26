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
