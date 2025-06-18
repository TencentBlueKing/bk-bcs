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

// Package k8s 提供 Istio 相关常量和函数
package k8s

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

// istioGroups 定义 Istio 资源组集合
var istioGroups = map[string]struct{}{
	string(SecurityGroup):   {},
	string(NetworkingGroup): {},
	string(InstallGroup):    {},
	string(TelemetryGroup):  {},
	string(ExtensionsGroup): {},
}

// IsIstioGroup 判断是否为Istio资源组
func IsIstioGroup(group string) bool {
	_, ok := istioGroups[group]
	return ok
}
