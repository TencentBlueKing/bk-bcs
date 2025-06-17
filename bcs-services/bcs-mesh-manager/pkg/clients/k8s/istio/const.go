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

// Package istio 提供 Istio 相关常量和函数
package istio

// Group represents an Istio API group
type Group string

const (
	// SecurityGroup Istio 安全相关资源组
	SecurityGroup Group = "security.istio.io"
	// NetworkingGroup Istio 网络相关资源组
	NetworkingGroup Group = "networking.istio.io"
	// InstallGroup Istio 安装相关资源组
	InstallGroup Group = "install.istio.io"
	// TelemetryGroup Istio 遥测相关资源组
	TelemetryGroup Group = "telemetry.istio.io"
	// ExtensionsGroup Istio 扩展相关资源组
	ExtensionsGroup Group = "extensions.istio.io"
)

// Kind represents an Istio resource kind
type Kind string

const (
	// AuthorizationPolicyKind 授权策略资源类型
	AuthorizationPolicyKind Kind = "authorizationpolicies"
	// DestinationRuleKind 目标规则资源类型
	DestinationRuleKind Kind = "destinationrules"
	// EnvoyFilterKind Envoy 过滤器资源类型
	EnvoyFilterKind Kind = "envoyfilters"
	// GatewayKind 网关资源类型
	GatewayKind Kind = "gateways"
	// IstioOperatorKind Istio 操作符资源类型
	IstioOperatorKind Kind = "istiooperators"
	// PeerAuthenticationKind 对等认证资源类型
	PeerAuthenticationKind Kind = "peerauthentications"
	// ProxyConfigKind 代理配置资源类型
	ProxyConfigKind Kind = "proxyconfigs"
	// RequestAuthenticationKind 请求认证资源类型
	RequestAuthenticationKind Kind = "requestauthentications"
	// ServiceEntryKind 服务入口资源类型
	ServiceEntryKind Kind = "serviceentries"
	// SidecarKind Sidecar 资源类型
	SidecarKind Kind = "sidecars"
	// TelemetryKind 遥测资源类型
	TelemetryKind Kind = "telemetries"
	// VirtualServiceKind 虚拟服务资源类型
	VirtualServiceKind Kind = "virtualservices"
	// WasmPluginKind Wasm 插件资源类型
	WasmPluginKind Kind = "wasmplugins"
	// WorkloadEntryKind 工作负载入口资源类型
	WorkloadEntryKind Kind = "workloadentries"
	// WorkloadGroupKind 工作负载组资源类型
	WorkloadGroupKind Kind = "workloadgroups"
)

// istioGroups 定义 Istio 资源组集合
var istioGroups = map[string]struct{}{
	string(SecurityGroup):   {},
	string(NetworkingGroup): {},
	string(InstallGroup):    {},
	string(TelemetryGroup):  {},
	string(ExtensionsGroup): {},
}

// istioKinds 定义 Istio 资源类型集合
var istioKinds = map[string]struct{}{
	string(AuthorizationPolicyKind):   {},
	string(DestinationRuleKind):       {},
	string(EnvoyFilterKind):           {},
	string(GatewayKind):               {},
	string(IstioOperatorKind):         {},
	string(PeerAuthenticationKind):    {},
	string(ProxyConfigKind):           {},
	string(RequestAuthenticationKind): {},
	string(ServiceEntryKind):          {},
	string(SidecarKind):               {},
	string(TelemetryKind):             {},
	string(VirtualServiceKind):        {},
	string(WasmPluginKind):            {},
	string(WorkloadEntryKind):         {},
	string(WorkloadGroupKind):         {},
}

// IsIstioGroup 判断是否是 Istio 组
func IsIstioGroup(group string) bool {
	_, ok := istioGroups[group]
	return ok
}

// IsIstioKind 判断是否是 Istio 资源类型
func IsIstioKind(kind string) bool {
	_, ok := istioKinds[kind]
	return ok
}
