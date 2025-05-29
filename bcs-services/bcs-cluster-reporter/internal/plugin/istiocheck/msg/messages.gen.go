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

// GENERATED FILE -- DO NOT EDIT
//

package msg

import (
	"istio.io/istio/pkg/config/analysis/diag"
	"istio.io/istio/pkg/config/resource"
)

// CodeToFriendlyName maps message code to friendlyName.
var CodeToFriendlyName = map[string]string{
	"IST0001": "内部错误",
	"IST0002": "已弃用",
	"IST0101": "引用的资源不存在",
	"IST0102": "未启用注入",
	"IST0103": "缺少代理",
	"IST0104": "网关端口未暴露",
	"IST0106": "Schema校验错误",
	"IST0107": "注解位置错误",
	"IST0108": "未知注解",
	"IST0109": "VirtualService主机冲突",
	"IST0110": "Sidecar选择冲突",
	"IST0111": "Sidecar未指定选择器",
	"IST0112": "VirtualService端口未指定",
	"IST0113": "mTLS策略冲突",
	"IST0116": "多服务端口协议冲突",
	"IST0117": "未关联服务",
	"IST0118": "端口名不规范",
	"IST0119": "JWT端口前缀无效",
	"IST0122": "正则表达式无效",
	"IST0123": "注入标签冲突",
	"IST0125": "注解无效",
	"IST0126": "未知服务注册表",
	"IST0127": "无匹配工作负载",
	"IST0128": "未验证服务器证书",
	"IST0129": "端口未验证服务器证书",
	"IST0130": "规则不可达",
	"IST0131": "规则匹配无效",
	"IST0132": "主机未在Gateway中",
	"IST0133": "Schema校验警告",
	"IST0134": "ServiceEntry缺少地址",
	"IST0135": "注解已弃用",
	"IST0136": "Alpha阶段注解",
	"IST0137": "端口冲突",
	"IST0138": "证书重复",
	"IST0139": "Webhook无效",
	"IST0140": "路由规则无效",
	"IST0141": "权限不足",
	"IST0142": "K8S版本不支持",
	"IST0143": "本地监听端口",
	"IST0144": "UID冲突",
	"IST0145": "Gateway冲突",
	"IST0146": "未注入image:auto",
	"IST0147": "未注入image:auto(错误)",
	"IST0148": "默认可注入",
	"IST0149": "JWT路由未认证",
	"IST0150": "ExternalName端口名无效",
	"IST0151": "EnvoyFilter相对操作",
	"IST0152": "EnvoyFilter REPLACE用法错误",
	"IST0153": "EnvoyFilter ADD用法错误",
	"IST0154": "EnvoyFilter REMOVE用法错误",
	"IST0155": "EnvoyFilter相对操作+proxyVersion",
	"IST0156": "GatewayAPI版本不支持",
	"IST0157": "Telemetry未设置provider",
	"IST0158": "代理镜像不一致",
	"IST0159": "Telemetry选择冲突",
	"IST0160": "Telemetry未指定选择器",
	"IST0161": "Gateway凭证无效",
}

var (
	// InternalError defines a diag.MessageType for message "InternalError".
	// Description: 工具链发生内部错误。这通常是实现中的 bug。
	InternalError = diag.NewMessageType(diag.Error, "IST0001", "内部错误: %v")

	// Deprecated defines a diag.MessageType for message "Deprecated".
	// Description: 配置依赖的某个功能已被弃用。
	Deprecated = diag.NewMessageType(diag.Warning, "IST0002", "已弃用: %s")

	// ReferencedResourceNotFound defines a diag.MessageType for message "ReferencedResourceNotFound".
	// Description: 被引用的资源不存在。
	ReferencedResourceNotFound = diag.NewMessageType(diag.Error, "IST0101", "引用的 %s 未找到: %q")

	// NamespaceNotInjected defines a diag.MessageType for message "NamespaceNotInjected".
	// Description: 命名空间未启用 Istio 注入。
	NamespaceNotInjected = diag.NewMessageType(diag.Info, "IST0102", "该命名空间未启用 Istio 注入。运行 'kubectl label namespace %s istio-injection=enabled' 以启用，或 'kubectl label namespace %s istio-injection=disabled' 明确标记为不需要注入。")

	// PodMissingProxy defines a diag.MessageType for message "PodMissingProxy".
	// Description: Pod 缺少 Istio 代理。
	PodMissingProxy = diag.NewMessageType(diag.Warning, "IST0103", "Pod %s 缺少 Istio 代理。通常可通过重启或重新部署工作负载解决。")

	// GatewayPortNotOnWorkload defines a diag.MessageType for message "GatewayPortNotOnWorkload".
	// Description: 未处理的网关端口
	GatewayPortNotOnWorkload = diag.NewMessageType(diag.Warning, "IST0104", "网关引用了未在工作负载（pod selector %s；端口 %d）上暴露的端口")

	// SchemaValidationError defines a diag.MessageType for message "SchemaValidationError".
	// Description: 资源存在 schema 校验错误。
	SchemaValidationError = diag.NewMessageType(diag.Error, "IST0106", "Schema 校验错误: %v")

	// MisplacedAnnotation defines a diag.MessageType for message "MisplacedAnnotation".
	// Description: Istio 注解应用在了错误的资源类型上。
	MisplacedAnnotation = diag.NewMessageType(diag.Warning, "IST0107", "注解位置错误: %s 只能应用于 %s")

	// UnknownAnnotation defines a diag.MessageType for message "UnknownAnnotation".
	// Description: Istio 注解无法识别，未适用于任何资源类型。
	UnknownAnnotation = diag.NewMessageType(diag.Warning, "IST0108", "未知注解: %s")

	// ConflictingMeshGatewayVirtualServiceHosts defines a diag.MessageType for message "ConflictingMeshGatewayVirtualServiceHosts".
	// Description: 与 mesh gateway 关联的 VirtualService 存在主机冲突。
	ConflictingMeshGatewayVirtualServiceHosts = diag.NewMessageType(diag.Error, "IST0109", "与 mesh gateway 关联的 VirtualService %s 定义了相同的主机 %s，可能导致未定义行为。可通过合并冲突的 VirtualService 资源解决。")

	// ConflictingSidecarWorkloadSelectors defines a diag.MessageType for message "ConflictingSidecarWorkloadSelectors".
	// Description: Sidecar 资源选择了与其他 Sidecar 资源相同的工作负载。
	ConflictingSidecarWorkloadSelectors = diag.NewMessageType(diag.Error, "IST0110", "命名空间 %q 中的 Sidecar %v 选择了相同的工作负载 pod %q，可能导致未定义行为。")

	// MultipleSidecarsWithoutWorkloadSelectors defines a diag.MessageType for message "MultipleSidecarsWithoutWorkloadSelectors".
	// Description: 一个命名空间中有多个 Sidecar 资源未设置 workload selector。
	MultipleSidecarsWithoutWorkloadSelectors = diag.NewMessageType(diag.Error, "IST0111", "命名空间 %q 中的 Sidecar %v 未设置 workload selector，可能导致未定义行为。")

	// VirtualServiceDestinationPortSelectorRequired defines a diag.MessageType for message "VirtualServiceDestinationPortSelectorRequired".
	// Description: VirtualService 路由到的服务暴露了多个端口，但未指定使用哪个端口。
	VirtualServiceDestinationPortSelectorRequired = diag.NewMessageType(diag.Error, "IST0112", "该 VirtualService 路由到的服务 %q 暴露了多个端口 %v。必须在 destination 中指定端口以消除歧义。")

	// MTLSPolicyConflict defines a diag.MessageType for message "MTLSPolicyConflict".
	// Description: DestinationRule 与 Policy 在 mTLS 配置上存在冲突。
	MTLSPolicyConflict = diag.NewMessageType(diag.Error, "IST0113", "DestinationRule %q 和 Policy %q 在主机 %s 的 mTLS 配置上存在冲突。DestinationRule 要求 mTLS 为 %t，而 Policy 对象要求为 %s。")

	// DeploymentAssociatedToMultipleServices defines a diag.MessageType for message "DeploymentAssociatedToMultipleServices".
	// Description: 服务网格部署的 pod 不能通过同一端口但不同协议关联到多个服务。
	DeploymentAssociatedToMultipleServices = diag.NewMessageType(diag.Warning, "IST0116", "该部署 %s 关联到多个服务，使用端口 %d 但协议不同: %v")

	// DeploymentRequiresServiceAssociated defines a diag.MessageType for message "DeploymentRequiresServiceAssociated".
	// Description: 服务网格部署的 pod 必须至少关联一个服务。
	DeploymentRequiresServiceAssociated = diag.NewMessageType(diag.Warning, "IST0117", "未关联任何服务。服务网格部署必须关联至少一个服务。")

	// PortNameIsNotUnderNamingConvention defines a diag.MessageType for message "PortNameIsNotUnderNamingConvention".
	// Description: 端口名不符合命名规范，将对该端口应用协议检测。
	PortNameIsNotUnderNamingConvention = diag.NewMessageType(diag.Info, "IST0118", "端口名 %s（端口: %d, targetPort: %s）不符合 Istio 端口命名规范。")

	// JwtFailureDueToInvalidServicePortPrefix defines a diag.MessageType for message "JwtFailureDueToInvalidServicePortPrefix".
	// Description: 带 JWT 的认证策略目标服务端口规范无效。
	JwtFailureDueToInvalidServicePortPrefix = diag.NewMessageType(diag.Warning, "IST0119", "带 JWT 的认证策略目标服务端口规范无效（端口: %d, 名称: %s, 协议: %s, targetPort: %s）。")

	// InvalidRegexp defines a diag.MessageType for message "InvalidRegexp".
	// Description: 无效的正则表达式
	InvalidRegexp = diag.NewMessageType(diag.Warning, "IST0122", "字段 %q 的正则表达式无效: %q (%s)")

	// NamespaceMultipleInjectionLabels defines a diag.MessageType for message "NamespaceMultipleInjectionLabels".
	// Description: 命名空间同时存在新旧注入标签。
	NamespaceMultipleInjectionLabels = diag.NewMessageType(diag.Warning, "IST0123", "该命名空间同时存在新旧注入标签。运行 'kubectl label namespace %s istio.io/rev-' 或 'kubectl label namespace %s istio-injection-'")

	// InvalidAnnotation defines a diag.MessageType for message "InvalidAnnotation".
	// Description: 无效的 Istio 注解
	InvalidAnnotation = diag.NewMessageType(diag.Warning, "IST0125", "无效注解 %s: %s")

	// UnknownMeshNetworksServiceRegistry defines a diag.MessageType for message "UnknownMeshNetworksServiceRegistry".
	// Description: Mesh Networks 中的服务注册表未知
	UnknownMeshNetworksServiceRegistry = diag.NewMessageType(diag.Error, "IST0126", "网络 %s 中的服务注册表 %s 未知")

	// NoMatchingWorkloadsFound defines a diag.MessageType for message "NoMatchingWorkloadsFound".
	// Description: 没有匹配资源标签的工作负载
	NoMatchingWorkloadsFound = diag.NewMessageType(diag.Warning, "IST0127", "该资源没有匹配以下标签的工作负载: %s")

	// NoServerCertificateVerificationDestinationLevel defines a diag.MessageType for message "NoServerCertificateVerificationDestinationLevel".
	// Description: DestinationRule 未设置 caCertificates，导致不会验证服务器证书。
	NoServerCertificateVerificationDestinationLevel = diag.NewMessageType(diag.Warning, "IST0128", "命名空间 %s 中的 DestinationRule %s TLS 模式为 %s，但未设置 caCertificates 验证主机 %s 的服务器身份。")

	// NoServerCertificateVerificationPortLevel defines a diag.MessageType for message "NoServerCertificateVerificationPortLevel".
	// Description: DestinationRule 未设置 caCertificates，导致不会验证指定端口的服务器证书。
	NoServerCertificateVerificationPortLevel = diag.NewMessageType(diag.Warning, "IST0129", "命名空间 %s 中的 DestinationRule %s TLS 模式为 %s，但未设置 caCertificates 验证主机 %s 的端口 %s 的服务器身份。")

	// VirtualServiceUnreachableRule defines a diag.MessageType for message "VirtualServiceUnreachableRule".
	// Description: VirtualService 某条规则因前面规则匹配相同，永远不会被使用。
	VirtualServiceUnreachableRule = diag.NewMessageType(diag.Warning, "IST0130", "VirtualService 规则 %v 未被使用（%s）。")

	// VirtualServiceIneffectiveMatch defines a diag.MessageType for message "VirtualServiceIneffectiveMatch".
	// Description: VirtualService 某条规则的 match 与前面规则重复。
	VirtualServiceIneffectiveMatch = diag.NewMessageType(diag.Info, "IST0131", "VirtualService 规则 %v 的 match %v 未被使用（在规则 %v 中重复/重叠）。")

	// VirtualServiceHostNotFoundInGateway defines a diag.MessageType for message "VirtualServiceHostNotFoundInGateway".
	// Description: VirtualService 中定义的主机未在 Gateway 中找到。
	VirtualServiceHostNotFoundInGateway = diag.NewMessageType(diag.Warning, "IST0132", "VirtualService %s 中定义的一个或多个主机 %v 未在 Gateway %s 中找到。")

	// SchemaWarning defines a diag.MessageType for message "SchemaWarning".
	// Description: 资源存在 schema 校验警告。
	SchemaWarning = diag.NewMessageType(diag.Warning, "IST0133", "Schema 校验警告: %v")

	// ServiceEntryAddressesRequired defines a diag.MessageType for message "ServiceEntryAddressesRequired".
	// Description: TCP（或未设置）协议的端口需要虚拟 IP 地址。
	ServiceEntryAddressesRequired = diag.NewMessageType(diag.Warning, "IST0134", "ServiceEntry 必须为该协议设置 addresses。")

	// DeprecatedAnnotation defines a diag.MessageType for message "DeprecatedAnnotation".
	// Description: 资源使用了已弃用的 Istio 注解。
	DeprecatedAnnotation = diag.NewMessageType(diag.Info, "IST0135", "注解 %q 已被弃用%s，未来 Istio 版本可能无法使用。")

	// AlphaAnnotation defines a diag.MessageType for message "AlphaAnnotation".
	// Description: Istio 注解属于 alpha 阶段，可能不适合生产环境。
	AlphaAnnotation = diag.NewMessageType(diag.Info, "IST0136", "注解 %q 属于 alpha 阶段功能，可能支持不完整。")

	// DeploymentConflictingPorts defines a diag.MessageType for message "DeploymentConflictingPorts".
	// Description: 选择同一工作负载且 targetPort 相同的两个服务，必须引用同一端口。
	DeploymentConflictingPorts = diag.NewMessageType(diag.Warning, "IST0137", "该部署 %s 关联到多个服务 %v，使用 targetPort %q 但端口不同: %v。")

	// GatewayDuplicateCertificate defines a diag.MessageType for message "GatewayDuplicateCertificate".
	// Description: 多个网关中重复的证书可能导致客户端复用 HTTP2 连接时出现 404。
	GatewayDuplicateCertificate = diag.NewMessageType(diag.Warning, "IST0138", "多个网关 %v 中重复的证书可能导致客户端复用 HTTP2 连接时出现 404。")

	// InvalidWebhook defines a diag.MessageType for message "InvalidWebhook".
	// Description: Webhook 无效或引用了不存在的控制面服务。
	InvalidWebhook = diag.NewMessageType(diag.Error, "IST0139", "%v")

	// IngressRouteRulesNotAffected defines a diag.MessageType for message "IngressRouteRulesNotAffected".
	// Description: 路由规则对 ingress gateway 请求无效。
	IngressRouteRulesNotAffected = diag.NewMessageType(diag.Warning, "IST0140", "virtual service %s 的 subset 对 ingress gateway %s 的请求无效")

	// InsufficientPermissions defines a diag.MessageType for message "InsufficientPermissions".
	// Description: 缺少安装 Istio 所需的权限。
	InsufficientPermissions = diag.NewMessageType(diag.Error, "IST0141", "缺少创建资源 %v 的权限（%v）")

	// UnsupportedKubernetesVersion defines a diag.MessageType for message "UnsupportedKubernetesVersion".
	// Description: Kubernetes 版本不受支持
	UnsupportedKubernetesVersion = diag.NewMessageType(diag.Error, "IST0142", "Kubernetes 版本 %q 低于最低要求版本: %v")

	// LocalhostListener defines a diag.MessageType for message "LocalhostListener".
	// Description: Service 暴露的端口绑定在本地地址。
	LocalhostListener = diag.NewMessageType(diag.Error, "IST0143", "端口 %v 在 Service 中暴露，但监听在 localhost，仅本地可访问。")

	// InvalidApplicationUID defines a diag.MessageType for message "InvalidApplicationUID".
	// Description: 应用 pod 不应以用户 ID (UID) 1337 运行。
	InvalidApplicationUID = diag.NewMessageType(diag.Warning, "IST0144", "用户 ID (UID) 1337 为 sidecar 代理保留。")

	// ConflictingGateways defines a diag.MessageType for message "ConflictingGateways".
	// Description: Gateway 不应具有相同的 selector、端口和主机。
	ConflictingGateways = diag.NewMessageType(diag.Error, "IST0145", "与 gateway %s 冲突（workload selector %s，端口 %s，主机 %v）。")

	// ImageAutoWithoutInjectionWarning defines a diag.MessageType for message "ImageAutoWithoutInjectionWarning".
	// Description: 带有 `image: auto` 的部署应启用注入。
	ImageAutoWithoutInjectionWarning = diag.NewMessageType(diag.Warning, "IST0146", "%s %s 包含 `image: auto` 但未匹配任何 Istio 注入 webhook selector。")

	// ImageAutoWithoutInjectionError defines a diag.MessageType for message "ImageAutoWithoutInjectionError".
	// Description: 带有 `image: auto` 的 pod 应启用注入。
	ImageAutoWithoutInjectionError = diag.NewMessageType(diag.Error, "IST0147", "%s %s 包含 `image: auto` 但未匹配任何 Istio 注入 webhook selector。")

	// NamespaceInjectionEnabledByDefault defines a diag.MessageType for message "NamespaceInjectionEnabledByDefault".
	// Description: 如果 Istio 安装时启用了 enableNamespacesByDefault 且未设置注入标签，则用户命名空间应可注入。
	NamespaceInjectionEnabledByDefault = diag.NewMessageType(diag.Info, "IST0148", "已启用 Istio 注入，因为 Istio 安装时 enableNamespacesByDefault 为 true。")

	// JwtClaimBasedRoutingWithoutRequestAuthN defines a diag.MessageType for message "JwtClaimBasedRoutingWithoutRequestAuthN".
	// Description: VirtualService 使用基于 JWT claim 的路由但未配置请求认证。
	JwtClaimBasedRoutingWithoutRequestAuthN = diag.NewMessageType(diag.Error, "IST0149", "该 virtual service 使用基于 JWT claim（key: %s）路由，但未为 gateway（%s）pod（%s）配置请求认证。必须先为 gateway pod 配置请求认证以校验 JWT 并使 claim 可用于路由。")

	// ExternalNameServiceTypeInvalidPortName defines a diag.MessageType for message "ExternalNameServiceTypeInvalidPortName".
	// Description: ExternalName 服务的端口名无效，可能导致代理无法正确转发 TCP 命名端口和未匹配流量。
	ExternalNameServiceTypeInvalidPortName = diag.NewMessageType(diag.Warning, "IST0150", "ExternalName 服务的端口名无效。代理可能无法正确转发 TCP 命名端口和未匹配流量。")

	// EnvoyFilterUsesRelativeOperation defines a diag.MessageType for message "EnvoyFilterUsesRelativeOperation".
	// Description: 该 EnvoyFilter 未设置优先级且使用了相对 patch 操作，可能导致未被应用。建议使用 INSERT_FIRST 或 ADD 选项，或设置优先级以确保正确应用。
	EnvoyFilterUsesRelativeOperation = diag.NewMessageType(diag.Warning, "IST0151", "该 EnvoyFilter 未设置优先级且使用了相对 patch 操作，可能导致未被应用。建议使用 INSERT_FIRST 或 ADD 选项，或设置优先级以确保正确应用。")

	// EnvoyFilterUsesReplaceOperationIncorrectly defines a diag.MessageType for message "EnvoyFilterUsesReplaceOperationIncorrectly".
	// Description: REPLACE 操作仅对 HTTP_FILTER 和 NETWORK_FILTER 有效。
	EnvoyFilterUsesReplaceOperationIncorrectly = diag.NewMessageType(diag.Error, "IST0152", "REPLACE 操作仅对 HTTP_FILTER 和 NETWORK_FILTER 有效。")

	// EnvoyFilterUsesAddOperationIncorrectly defines a diag.MessageType for message "EnvoyFilterUsesAddOperationIncorrectly".
	// Description: 当 applyTo 设置为 ROUTE_CONFIGURATION 或 HTTP_ROUTE 时，ADD 操作将被忽略。
	EnvoyFilterUsesAddOperationIncorrectly = diag.NewMessageType(diag.Error, "IST0153", "当 applyTo 设置为 ROUTE_CONFIGURATION 或 HTTP_ROUTE 时，ADD 操作将被忽略。")

	// EnvoyFilterUsesRemoveOperationIncorrectly defines a diag.MessageType for message "EnvoyFilterUsesRemoveOperationIncorrectly".
	// Description: 当 applyTo 设置为 ROUTE_CONFIGURATION 或 HTTP_ROUTE 时，REMOVE 操作将被忽略。
	EnvoyFilterUsesRemoveOperationIncorrectly = diag.NewMessageType(diag.Error, "IST0154", "当 applyTo 设置为 ROUTE_CONFIGURATION 或 HTTP_ROUTE 时，REMOVE 操作将被忽略。")

	// EnvoyFilterUsesRelativeOperationWithProxyVersion defines a diag.MessageType for message "EnvoyFilterUsesRelativeOperationWithProxyVersion".
	// Description: 该 EnvoyFilter 未设置优先级，且使用了相对 patch 操作（NSTERT_BEFORE/AFTER、REPLACE、MERGE、DELETE）和 proxyVersion，可能导致升级时未被应用。建议使用 INSERT_FIRST 或 ADD 选项，或设置优先级以确保正确应用。
	EnvoyFilterUsesRelativeOperationWithProxyVersion = diag.NewMessageType(diag.Warning, "IST0155", "该 EnvoyFilter 未设置优先级，且使用了相对 patch 操作（NSTERT_BEFORE/AFTER、REPLACE、MERGE、DELETE）和 proxyVersion，可能导致升级时未被应用。建议使用 INSERT_FIRST 或 ADD 选项，或设置优先级以确保正确应用。")

	// UnsupportedGatewayAPIVersion defines a diag.MessageType for message "UnsupportedGatewayAPIVersion".
	// Description: Gateway API CRD 版本不受支持
	UnsupportedGatewayAPIVersion = diag.NewMessageType(diag.Error, "IST0156", "Gateway API CRD 版本 %v 低于最低要求版本: %v")

	// InvalidTelemetryProvider defines a diag.MessageType for message "InvalidTelemetryProvider".
	// Description: Telemetry 资源未设置 provider，将被忽略。
	InvalidTelemetryProvider = diag.NewMessageType(diag.Warning, "IST0157", "命名空间 %q 中的 Telemetry %v 未设置 provider，将被忽略。")

	// PodsIstioProxyImageMismatchInNamespace defines a diag.MessageType for message "PodsIstioProxyImageMismatchInNamespace".
	// Description: 命名空间中 pod 的 Istio 代理镜像与注入配置中定义的不一致。
	PodsIstioProxyImageMismatchInNamespace = diag.NewMessageType(diag.Warning, "IST0158", "命名空间中 pod 的 Istio 代理镜像与注入配置中定义的不一致（pod 名称: %v）。通常在升级 Istio 控制面后出现，可通过重新部署 pod 解决。")

	// ConflictingTelemetryWorkloadSelectors defines a diag.MessageType for message "ConflictingTelemetryWorkloadSelectors".
	// Description: Telemetry 资源选择了与其他 Telemetry 资源相同的工作负载。
	ConflictingTelemetryWorkloadSelectors = diag.NewMessageType(diag.Error, "IST0159", "命名空间 %q 中的 Telemetry %v 选择了相同的工作负载 pod %q，可能导致未定义行为。")

	// MultipleTelemetriesWithoutWorkloadSelectors defines a diag.MessageType for message "MultipleTelemetriesWithoutWorkloadSelectors".
	// Description: 一个命名空间中有多个 Telemetry 资源未设置 workload selector。
	MultipleTelemetriesWithoutWorkloadSelectors = diag.NewMessageType(diag.Error, "IST0160", "命名空间 %q 中的 Telemetry %v 未设置 workload selector，可能导致未定义行为。")

	// InvalidGatewayCredential defines a diag.MessageType for message "InvalidGatewayCredential".
	// Description: Gateway 资源引用的凭证无效
	InvalidGatewayCredential = diag.NewMessageType(diag.Error, "IST0161", "命名空间 %s 中的 Gateway %s 引用的凭证无效，可能导致流量异常。")
)

// All 返回所有消息类型。
func All() []*diag.MessageType {
	return []*diag.MessageType{
		InternalError,
		Deprecated,
		ReferencedResourceNotFound,
		NamespaceNotInjected,
		PodMissingProxy,
		GatewayPortNotOnWorkload,
		SchemaValidationError,
		MisplacedAnnotation,
		UnknownAnnotation,
		ConflictingMeshGatewayVirtualServiceHosts,
		ConflictingSidecarWorkloadSelectors,
		MultipleSidecarsWithoutWorkloadSelectors,
		VirtualServiceDestinationPortSelectorRequired,
		MTLSPolicyConflict,
		DeploymentAssociatedToMultipleServices,
		DeploymentRequiresServiceAssociated,
		PortNameIsNotUnderNamingConvention,
		JwtFailureDueToInvalidServicePortPrefix,
		InvalidRegexp,
		NamespaceMultipleInjectionLabels,
		InvalidAnnotation,
		UnknownMeshNetworksServiceRegistry,
		NoMatchingWorkloadsFound,
		NoServerCertificateVerificationDestinationLevel,
		NoServerCertificateVerificationPortLevel,
		VirtualServiceUnreachableRule,
		VirtualServiceIneffectiveMatch,
		VirtualServiceHostNotFoundInGateway,
		SchemaWarning,
		ServiceEntryAddressesRequired,
		DeprecatedAnnotation,
		AlphaAnnotation,
		DeploymentConflictingPorts,
		GatewayDuplicateCertificate,
		InvalidWebhook,
		IngressRouteRulesNotAffected,
		InsufficientPermissions,
		UnsupportedKubernetesVersion,
		LocalhostListener,
		InvalidApplicationUID,
		ConflictingGateways,
		ImageAutoWithoutInjectionWarning,
		ImageAutoWithoutInjectionError,
		NamespaceInjectionEnabledByDefault,
		JwtClaimBasedRoutingWithoutRequestAuthN,
		ExternalNameServiceTypeInvalidPortName,
		EnvoyFilterUsesRelativeOperation,
		EnvoyFilterUsesReplaceOperationIncorrectly,
		EnvoyFilterUsesAddOperationIncorrectly,
		EnvoyFilterUsesRemoveOperationIncorrectly,
		EnvoyFilterUsesRelativeOperationWithProxyVersion,
		UnsupportedGatewayAPIVersion,
		InvalidTelemetryProvider,
		PodsIstioProxyImageMismatchInNamespace,
		ConflictingTelemetryWorkloadSelectors,
		MultipleTelemetriesWithoutWorkloadSelectors,
		InvalidGatewayCredential,
	}
}

// NewInternalError 创建一个 InternalError 类型的消息。
func NewInternalError(r *resource.Instance, detail string) diag.Message {
	return diag.NewMessage(
		InternalError,
		r,
		detail,
	)
}

// NewDeprecated 创建一个 Deprecated 类型的消息。
func NewDeprecated(r *resource.Instance, detail string) diag.Message {
	return diag.NewMessage(
		Deprecated,
		r,
		detail,
	)
}

// NewReferencedResourceNotFound 创建一个 ReferencedResourceNotFound 类型的消息。
func NewReferencedResourceNotFound(r *resource.Instance, reftype string, refval string) diag.Message {
	return diag.NewMessage(
		ReferencedResourceNotFound,
		r,
		reftype,
		refval,
	)
}

// NewNamespaceNotInjected 创建一个 NamespaceNotInjected 类型的消息。
func NewNamespaceNotInjected(r *resource.Instance, namespace string, namespace2 string) diag.Message {
	return diag.NewMessage(
		NamespaceNotInjected,
		r,
		namespace,
		namespace2,
	)
}

// NewPodMissingProxy 创建一个 PodMissingProxy 类型的消息。
func NewPodMissingProxy(r *resource.Instance, podName string) diag.Message {
	return diag.NewMessage(
		PodMissingProxy,
		r,
		podName,
	)
}

// NewGatewayPortNotOnWorkload 创建一个 GatewayPortNotOnWorkload 类型的消息。
func NewGatewayPortNotOnWorkload(r *resource.Instance, selector string, port int) diag.Message {
	return diag.NewMessage(
		GatewayPortNotOnWorkload,
		r,
		selector,
		port,
	)
}

// NewSchemaValidationError 创建一个 SchemaValidationError 类型的消息。
func NewSchemaValidationError(r *resource.Instance, err error) diag.Message {
	return diag.NewMessage(
		SchemaValidationError,
		r,
		err,
	)
}

// NewMisplacedAnnotation 创建一个 MisplacedAnnotation 类型的消息。
func NewMisplacedAnnotation(r *resource.Instance, annotation string, kind string) diag.Message {
	return diag.NewMessage(
		MisplacedAnnotation,
		r,
		annotation,
		kind,
	)
}

// NewUnknownAnnotation 创建一个 UnknownAnnotation 类型的消息。
func NewUnknownAnnotation(r *resource.Instance, annotation string) diag.Message {
	return diag.NewMessage(
		UnknownAnnotation,
		r,
		annotation,
	)
}

// NewConflictingMeshGatewayVirtualServiceHosts 创建一个 ConflictingMeshGatewayVirtualServiceHosts 类型的消息。
func NewConflictingMeshGatewayVirtualServiceHosts(r *resource.Instance, virtualServices string, host string) diag.Message {
	return diag.NewMessage(
		ConflictingMeshGatewayVirtualServiceHosts,
		r,
		virtualServices,
		host,
	)
}

// NewConflictingSidecarWorkloadSelectors 创建一个 ConflictingSidecarWorkloadSelectors 类型的消息。
func NewConflictingSidecarWorkloadSelectors(r *resource.Instance, conflictingSidecars []string, namespace string, workloadPod string) diag.Message {
	return diag.NewMessage(
		ConflictingSidecarWorkloadSelectors,
		r,
		conflictingSidecars,
		namespace,
		workloadPod,
	)
}

// NewMultipleSidecarsWithoutWorkloadSelectors 创建一个 MultipleSidecarsWithoutWorkloadSelectors 类型的消息。
func NewMultipleSidecarsWithoutWorkloadSelectors(r *resource.Instance, conflictingSidecars []string, namespace string) diag.Message {
	return diag.NewMessage(
		MultipleSidecarsWithoutWorkloadSelectors,
		r,
		conflictingSidecars,
		namespace,
	)
}

// NewVirtualServiceDestinationPortSelectorRequired 创建一个 VirtualServiceDestinationPortSelectorRequired 类型的消息。
func NewVirtualServiceDestinationPortSelectorRequired(r *resource.Instance, destHost string, destPorts []int) diag.Message {
	return diag.NewMessage(
		VirtualServiceDestinationPortSelectorRequired,
		r,
		destHost,
		destPorts,
	)
}

// NewMTLSPolicyConflict 创建一个 MTLSPolicyConflict 类型的消息。
func NewMTLSPolicyConflict(r *resource.Instance, host string, destinationRuleName string, destinationRuleMTLSMode bool, policyName string, policyMTLSMode string) diag.Message {
	return diag.NewMessage(
		MTLSPolicyConflict,
		r,
		host,
		destinationRuleName,
		destinationRuleMTLSMode,
		policyName,
		policyMTLSMode,
	)
}

// NewDeploymentAssociatedToMultipleServices 创建一个 DeploymentAssociatedToMultipleServices 类型的消息。
func NewDeploymentAssociatedToMultipleServices(r *resource.Instance, deployment string, port int32, services []string) diag.Message {
	return diag.NewMessage(
		DeploymentAssociatedToMultipleServices,
		r,
		deployment,
		port,
		services,
	)
}

// NewDeploymentRequiresServiceAssociated 创建一个 DeploymentRequiresServiceAssociated 类型的消息。
func NewDeploymentRequiresServiceAssociated(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		DeploymentRequiresServiceAssociated,
		r,
	)
}

// NewPortNameIsNotUnderNamingConvention 创建一个 PortNameIsNotUnderNamingConvention 类型的消息。
func NewPortNameIsNotUnderNamingConvention(r *resource.Instance, portName string, port int, targetPort string) diag.Message {
	return diag.NewMessage(
		PortNameIsNotUnderNamingConvention,
		r,
		portName,
		port,
		targetPort,
	)
}

// NewJwtFailureDueToInvalidServicePortPrefix 创建一个 JwtFailureDueToInvalidServicePortPrefix 类型的消息。
func NewJwtFailureDueToInvalidServicePortPrefix(r *resource.Instance, port int, portName string, protocol string, targetPort string) diag.Message {
	return diag.NewMessage(
		JwtFailureDueToInvalidServicePortPrefix,
		r,
		port,
		portName,
		protocol,
		targetPort,
	)
}

// NewInvalidRegexp 创建一个 InvalidRegexp 类型的消息。
func NewInvalidRegexp(r *resource.Instance, where string, re string, problem string) diag.Message {
	return diag.NewMessage(
		InvalidRegexp,
		r,
		where,
		re,
		problem,
	)
}

// NewNamespaceMultipleInjectionLabels 创建一个 NamespaceMultipleInjectionLabels 类型的消息。
func NewNamespaceMultipleInjectionLabels(r *resource.Instance, namespace string, namespace2 string) diag.Message {
	return diag.NewMessage(
		NamespaceMultipleInjectionLabels,
		r,
		namespace,
		namespace2,
	)
}

// NewInvalidAnnotation 创建一个 InvalidAnnotation 类型的消息。
func NewInvalidAnnotation(r *resource.Instance, annotation string, problem string) diag.Message {
	return diag.NewMessage(
		InvalidAnnotation,
		r,
		annotation,
		problem,
	)
}

// NewUnknownMeshNetworksServiceRegistry 创建一个 UnknownMeshNetworksServiceRegistry 类型的消息。
func NewUnknownMeshNetworksServiceRegistry(r *resource.Instance, serviceregistry string, network string) diag.Message {
	return diag.NewMessage(
		UnknownMeshNetworksServiceRegistry,
		r,
		serviceregistry,
		network,
	)
}

// NewNoMatchingWorkloadsFound 创建一个 NoMatchingWorkloadsFound 类型的消息。
func NewNoMatchingWorkloadsFound(r *resource.Instance, labels string) diag.Message {
	return diag.NewMessage(
		NoMatchingWorkloadsFound,
		r,
		labels,
	)
}

// NewNoServerCertificateVerificationDestinationLevel 创建一个 NoServerCertificateVerificationDestinationLevel 类型的消息。
func NewNoServerCertificateVerificationDestinationLevel(r *resource.Instance, destinationrule string, namespace string, mode string, host string) diag.Message {
	return diag.NewMessage(
		NoServerCertificateVerificationDestinationLevel,
		r,
		destinationrule,
		namespace,
		mode,
		host,
	)
}

// NewNoServerCertificateVerificationPortLevel 创建一个 NoServerCertificateVerificationPortLevel 类型的消息。
func NewNoServerCertificateVerificationPortLevel(r *resource.Instance, destinationrule string, namespace string, mode string, host string, port string) diag.Message {
	return diag.NewMessage(
		NoServerCertificateVerificationPortLevel,
		r,
		destinationrule,
		namespace,
		mode,
		host,
		port,
	)
}

// NewVirtualServiceUnreachableRule 创建一个 VirtualServiceUnreachableRule 类型的消息。
func NewVirtualServiceUnreachableRule(r *resource.Instance, ruleno string, reason string) diag.Message {
	return diag.NewMessage(
		VirtualServiceUnreachableRule,
		r,
		ruleno,
		reason,
	)
}

// NewVirtualServiceIneffectiveMatch 创建一个 VirtualServiceIneffectiveMatch 类型的消息。
func NewVirtualServiceIneffectiveMatch(r *resource.Instance, ruleno string, matchno string, dupno string) diag.Message {
	return diag.NewMessage(
		VirtualServiceIneffectiveMatch,
		r,
		ruleno,
		matchno,
		dupno,
	)
}

// NewVirtualServiceHostNotFoundInGateway 创建一个 VirtualServiceHostNotFoundInGateway 类型的消息。
func NewVirtualServiceHostNotFoundInGateway(r *resource.Instance, host []string, virtualservice string, gateway string) diag.Message {
	return diag.NewMessage(
		VirtualServiceHostNotFoundInGateway,
		r,
		host,
		virtualservice,
		gateway,
	)
}

// NewSchemaWarning 创建一个 SchemaWarning 类型的消息。
func NewSchemaWarning(r *resource.Instance, err error) diag.Message {
	return diag.NewMessage(
		SchemaWarning,
		r,
		err,
	)
}

// NewServiceEntryAddressesRequired 创建一个 ServiceEntryAddressesRequired 类型的消息。
func NewServiceEntryAddressesRequired(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		ServiceEntryAddressesRequired,
		r,
	)
}

// NewDeprecatedAnnotation 创建一个 DeprecatedAnnotation 类型的消息。
func NewDeprecatedAnnotation(r *resource.Instance, annotation string, extra string) diag.Message {
	return diag.NewMessage(
		DeprecatedAnnotation,
		r,
		annotation,
		extra,
	)
}

// NewAlphaAnnotation 创建一个 AlphaAnnotation 类型的消息。
func NewAlphaAnnotation(r *resource.Instance, annotation string) diag.Message {
	return diag.NewMessage(
		AlphaAnnotation,
		r,
		annotation,
	)
}

// NewDeploymentConflictingPorts 创建一个 DeploymentConflictingPorts 类型的消息。
func NewDeploymentConflictingPorts(r *resource.Instance, deployment string, services []string, targetPort string, ports []int32) diag.Message {
	return diag.NewMessage(
		DeploymentConflictingPorts,
		r,
		deployment,
		services,
		targetPort,
		ports,
	)
}

// NewGatewayDuplicateCertificate 创建一个 GatewayDuplicateCertificate 类型的消息。
func NewGatewayDuplicateCertificate(r *resource.Instance, gateways []string) diag.Message {
	return diag.NewMessage(
		GatewayDuplicateCertificate,
		r,
		gateways,
	)
}

// NewInvalidWebhook 创建一个 InvalidWebhook 类型的消息。
func NewInvalidWebhook(r *resource.Instance, error string) diag.Message {
	return diag.NewMessage(
		InvalidWebhook,
		r,
		error,
	)
}

// NewIngressRouteRulesNotAffected 创建一个 IngressRouteRulesNotAffected 类型的消息。
func NewIngressRouteRulesNotAffected(r *resource.Instance, virtualservicesubset string, virtualservice string) diag.Message {
	return diag.NewMessage(
		IngressRouteRulesNotAffected,
		r,
		virtualservicesubset,
		virtualservice,
	)
}

// NewInsufficientPermissions 创建一个 InsufficientPermissions 类型的消息。
func NewInsufficientPermissions(r *resource.Instance, resource string, error string) diag.Message {
	return diag.NewMessage(
		InsufficientPermissions,
		r,
		resource,
		error,
	)
}

// NewUnsupportedKubernetesVersion 创建一个 UnsupportedKubernetesVersion 类型的消息。
func NewUnsupportedKubernetesVersion(r *resource.Instance, version string, minimumVersion string) diag.Message {
	return diag.NewMessage(
		UnsupportedKubernetesVersion,
		r,
		version,
		minimumVersion,
	)
}

// NewLocalhostListener 创建一个 LocalhostListener 类型的消息。
func NewLocalhostListener(r *resource.Instance, port string) diag.Message {
	return diag.NewMessage(
		LocalhostListener,
		r,
		port,
	)
}

// NewInvalidApplicationUID 创建一个 InvalidApplicationUID 类型的消息。
func NewInvalidApplicationUID(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		InvalidApplicationUID,
		r,
	)
}

// NewConflictingGateways 创建一个 ConflictingGateways 类型的消息。
func NewConflictingGateways(r *resource.Instance, gateway string, selector string, portnumber string, hosts string) diag.Message {
	return diag.NewMessage(
		ConflictingGateways,
		r,
		gateway,
		selector,
		portnumber,
		hosts,
	)
}

// NewImageAutoWithoutInjectionWarning 创建一个 ImageAutoWithoutInjectionWarning 类型的消息。
func NewImageAutoWithoutInjectionWarning(r *resource.Instance, resourceType string, resourceName string) diag.Message {
	return diag.NewMessage(
		ImageAutoWithoutInjectionWarning,
		r,
		resourceType,
		resourceName,
	)
}

// NewImageAutoWithoutInjectionError 创建一个 ImageAutoWithoutInjectionError 类型的消息。
func NewImageAutoWithoutInjectionError(r *resource.Instance, resourceType string, resourceName string) diag.Message {
	return diag.NewMessage(
		ImageAutoWithoutInjectionError,
		r,
		resourceType,
		resourceName,
	)
}

// NewNamespaceInjectionEnabledByDefault 创建一个 NamespaceInjectionEnabledByDefault 类型的消息。
func NewNamespaceInjectionEnabledByDefault(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		NamespaceInjectionEnabledByDefault,
		r,
	)
}

// NewJwtClaimBasedRoutingWithoutRequestAuthN 创建一个 JwtClaimBasedRoutingWithoutRequestAuthN 类型的消息。
func NewJwtClaimBasedRoutingWithoutRequestAuthN(r *resource.Instance, key string, gateway string, pod string) diag.Message {
	return diag.NewMessage(
		JwtClaimBasedRoutingWithoutRequestAuthN,
		r,
		key,
		gateway,
		pod,
	)
}

// NewExternalNameServiceTypeInvalidPortName 创建一个 ExternalNameServiceTypeInvalidPortName 类型的消息。
func NewExternalNameServiceTypeInvalidPortName(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		ExternalNameServiceTypeInvalidPortName,
		r,
	)
}

// NewEnvoyFilterUsesRelativeOperation 创建一个 EnvoyFilterUsesRelativeOperation 类型的消息。
func NewEnvoyFilterUsesRelativeOperation(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		EnvoyFilterUsesRelativeOperation,
		r,
	)
}

// NewEnvoyFilterUsesReplaceOperationIncorrectly 创建一个 EnvoyFilterUsesReplaceOperationIncorrectly 类型的消息。
func NewEnvoyFilterUsesReplaceOperationIncorrectly(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		EnvoyFilterUsesReplaceOperationIncorrectly,
		r,
	)
}

// NewEnvoyFilterUsesAddOperationIncorrectly 创建一个 EnvoyFilterUsesAddOperationIncorrectly 类型的消息。
func NewEnvoyFilterUsesAddOperationIncorrectly(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		EnvoyFilterUsesAddOperationIncorrectly,
		r,
	)
}

// NewEnvoyFilterUsesRemoveOperationIncorrectly 创建一个 EnvoyFilterUsesRemoveOperationIncorrectly 类型的消息。
func NewEnvoyFilterUsesRemoveOperationIncorrectly(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		EnvoyFilterUsesRemoveOperationIncorrectly,
		r,
	)
}

// NewEnvoyFilterUsesRelativeOperationWithProxyVersion 创建一个 EnvoyFilterUsesRelativeOperationWithProxyVersion 类型的消息。
func NewEnvoyFilterUsesRelativeOperationWithProxyVersion(r *resource.Instance) diag.Message {
	return diag.NewMessage(
		EnvoyFilterUsesRelativeOperationWithProxyVersion,
		r,
	)
}

// NewUnsupportedGatewayAPIVersion 创建一个 UnsupportedGatewayAPIVersion 类型的消息。
func NewUnsupportedGatewayAPIVersion(r *resource.Instance, version string, minimumVersion string) diag.Message {
	return diag.NewMessage(
		UnsupportedGatewayAPIVersion,
		r,
		version,
		minimumVersion,
	)
}

// NewInvalidTelemetryProvider 创建一个 InvalidTelemetryProvider 类型的消息。
func NewInvalidTelemetryProvider(r *resource.Instance, name string, namespace string) diag.Message {
	return diag.NewMessage(
		InvalidTelemetryProvider,
		r,
		name,
		namespace,
	)
}

// NewPodsIstioProxyImageMismatchInNamespace 创建一个 PodsIstioProxyImageMismatchInNamespace 类型的消息。
func NewPodsIstioProxyImageMismatchInNamespace(r *resource.Instance, podNames []string) diag.Message {
	return diag.NewMessage(
		PodsIstioProxyImageMismatchInNamespace,
		r,
		podNames,
	)
}

// NewConflictingTelemetryWorkloadSelectors 创建一个 ConflictingTelemetryWorkloadSelectors 类型的消息。
func NewConflictingTelemetryWorkloadSelectors(r *resource.Instance, conflictingTelemetries []string, namespace string, workloadPod string) diag.Message {
	return diag.NewMessage(
		ConflictingTelemetryWorkloadSelectors,
		r,
		conflictingTelemetries,
		namespace,
		workloadPod,
	)
}

// NewMultipleTelemetriesWithoutWorkloadSelectors 创建一个 MultipleTelemetriesWithoutWorkloadSelectors 类型的消息。
func NewMultipleTelemetriesWithoutWorkloadSelectors(r *resource.Instance, conflictingTelemetries []string, namespace string) diag.Message {
	return diag.NewMessage(
		MultipleTelemetriesWithoutWorkloadSelectors,
		r,
		conflictingTelemetries,
		namespace,
	)
}

// NewInvalidGatewayCredential 创建一个 InvalidGatewayCredential 类型的消息。
func NewInvalidGatewayCredential(r *resource.Instance, gatewayName string, gatewayNamespace string) diag.Message {
	return diag.NewMessage(
		InvalidGatewayCredential,
		r,
		gatewayName,
		gatewayNamespace,
	)
}
