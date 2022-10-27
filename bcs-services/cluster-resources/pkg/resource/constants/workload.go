/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constants

// Volume2ResNameKeyMap Pod Volume 字段中，关联的资源类型与 NameKey 映射表
var Volume2ResNameKeyMap = map[string]string{
	PVC:    "claimName",
	Secret: "secretName",
	CM:     "name",
}

const (
	// DefaultUpdateStrategy 默认更新策略
	DefaultUpdateStrategy = "RollingUpdate"
	// DefaultMaxSurge 默认最大调度 Pod 数量
	DefaultMaxSurge = 25
	// DefaultMaxUnavailable 默认最大不可用数量
	DefaultMaxUnavailable = 25
)

const (
	// NodeSelectTypeAnyAvailable 节点选择类型 - 任意节点
	NodeSelectTypeAnyAvailable = "anyAvailable"
	// NodeSelectTypeSpecificNode 节点选择类型 - 指定节点
	NodeSelectTypeSpecificNode = "specificNode"
	// NodeSelectTypeSchedulingRule 节点选择类型 - 调度规则
	NodeSelectTypeSchedulingRule = "schedulingRule"
)

const (
	// AffinityTypeAffinity 亲和性类型 - 亲和性
	AffinityTypeAffinity = "affinity"
	// AffinityTypeAntiAffinity 亲和性类型 - 反亲和性
	AffinityTypeAntiAffinity = "antiAffinity"
	// AffinityPriorityRequired 亲和性优先级 - 必须
	AffinityPriorityRequired = "required"
	// AffinityPriorityPreferred 亲和性优先级 - 优先
	AffinityPriorityPreferred = "preferred"
)

const (
	// EnvVarTypeKeyVal Key-Value 类型
	EnvVarTypeKeyVal = "keyValue"
	// EnvVarTypePodField PodField 类型
	EnvVarTypePodField = "podField"
	// EnvVarTypeResource Resource 类型
	EnvVarTypeResource = "resource"
	// EnvVarTypeCMKey ConfigMap Key 类型
	EnvVarTypeCMKey = "configMapKey"
	// EnvVarTypeSecretKey Secret Key 类型
	EnvVarTypeSecretKey = "secretKey"
	// EnvVarTypeCM ConfigMap 类型
	EnvVarTypeCM = "configMap"
	// EnvVarTypeSecret Secret 类型
	EnvVarTypeSecret = "secret"
)

const (
	// ProbeTypeHTTPGet HTTP 探针
	ProbeTypeHTTPGet = "httpGet"
	// ProbeTypeTCPSocket TCP 探针
	ProbeTypeTCPSocket = "tcpSocket"
	// ProbeTypeExec 命令执行探针
	ProbeTypeExec = "exec"
)
