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

const (
	// IngClsAnnoKey ...
	IngClsAnnoKey = "kubernetes.io/ingress.class"

	// IngAutoRewriteHTTPAnnoKey http 端口重定向到 https 端口
	IngAutoRewriteHTTPAnnoKey = "ingress.cloud.tencent.com/auto-rewrite"

	// IngExistLBIDAnnoKey ...
	IngExistLBIDAnnoKey = "kubernetes.io/ingress.existLbId"

	// IngQcloudCurLBIDAnnoKey 控制器为 qcloud 的 ingress 当前使用的 lb id
	IngQcloudCurLBIDAnnoKey = "kubernetes.io/ingress.qcloud-loadbalance-id"

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
	// SVCCurLBIDAnnoKey service 目前使用的 lb-id，可能来源于 existed-lbid，也可能来源于自动创建的
	SVCCurLBIDAnnoKey = "service.kubernetes.io/loadbalance-id"

	// SVCExistLBIDAnnoKey ...
	SVCExistLBIDAnnoKey = "service.kubernetes.io/tke-existed-lbid"

	// SVCSubNetIDAnnoKey ...
	SVCSubNetIDAnnoKey = "service.kubernetes.io/qcloud-loadbalancer-internal-subnetid"
)

const (
	// SVCTypeClusterIP ClusterIP 类型的 Service
	SVCTypeClusterIP = "ClusterIP"

	// SVCTypeNodePort ...
	SVCTypeNodePort = "NodePort"

	// SVCTypeLoadBalancer ...
	SVCTypeLoadBalancer = "LoadBalancer"
)

// IngTargetSVCEnabledServiceTypes 可以作为 ingress target service 的 service 类型
var IngTargetSVCEnabledServiceTypes = []string{SVCTypeNodePort, SVCTypeLoadBalancer}

const (
	// SessionAffinityTypeNone Session 亲和性 None 类型
	SessionAffinityTypeNone = "None"

	// SessionAffinityTypeClientIP Session 亲和性 ClusterIP 类型
	SessionAffinityTypeClientIP = "ClientIP"

	// DefaultSessionAffinityStickyTime 默认的会话保留时间：10800 秒
	DefaultSessionAffinityStickyTime = int64(10800)
)
