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

package workload

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
	// EnvVarTypeKeyVal ...
	EnvVarTypeKeyVal = "keyValue"
	// EnvVarTypePodField ...
	EnvVarTypePodField = "podField"
	// EnvVarTypeResource ...
	EnvVarTypeResource = "resource"
	// EnvVarTypeCMKey ...
	EnvVarTypeCMKey = "configMapKey"
	// EnvVarTypeSecretKey ...
	EnvVarTypeSecretKey = "secretKey"
	// EnvVarTypeCM ...
	EnvVarTypeCM = "configMap"
	// EnvVarTypeSecret ...
	EnvVarTypeSecret = "secret"
)

const (
	// ProbeTypeHTTPGet ...
	ProbeTypeHTTPGet = "httpGet"
	// ProbeTypeTCPSocket ...
	ProbeTypeTCPSocket = "tcpSocket"
	// ProbeTypeExec ...
	ProbeTypeExec = "exec"
)
