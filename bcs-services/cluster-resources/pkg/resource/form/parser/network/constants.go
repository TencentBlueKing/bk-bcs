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

package network

const (
	// IngClsAnnoKey ...
	IngClsAnnoKey = "kubernetes.io/ingress.class"

	// IngExistLBIDAnnoKey ...
	IngExistLBIDAnnoKey = "kubernetes.io/ingress.existLbId"

	// IngSubNetIDAnnoKey ...
	IngSubNetIDAnnoKey = "kubernetes.io/ingress.subnetId"
)

const (
	// IngClsQCloud QCloud 类型 Ingress 控制器
	// annotations: kubernetes.io/ingress.class: qcloud
	IngClsQCloud = "qcloud"

	// IngClsNginx Nginx 类型 Ingress 控制器（默认类型）
	IngClsNginx = "nginx"
)

const (
	// CLBUseTypeUseExists 使用已经存在的 clb 实例
	CLBUseTypeUseExists = "useExists"

	// CLBUseTypeAutoCreate 自动创建新的 clb 实例
	CLBUseTypeAutoCreate = "autoCreate"
)

const (
	// CLBTypeExternal 外网 CLB
	CLBTypeExternal = "external"

	// CLBTypeInternal 内网 CLB
	CLBTypeInternal = "internal"
)

const (
	// SessionAffinityTypeNone Session 亲和性 None 类型
	SessionAffinityTypeNone = "None"

	// SessionAffinityTypeClientIP Session 亲和性 ClusterIP 类型
	SessionAffinityTypeClientIP = "ClientIP"
)
