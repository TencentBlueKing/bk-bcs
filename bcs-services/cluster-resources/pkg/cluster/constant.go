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

package cluster

import (
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
)

// 集群类型
const (
	// ClusterTypeSingle 独立集群
	ClusterTypeSingle = "Single"

	// ClusterTypeShared 共享集群
	ClusterTypeShared = "Shared"

	// ClusterTypeFederation 联邦集群
	ClusterTypeFederation = "Federation"

	// ClusterTypeFederationShared 共享联邦集群
	ClusterTypeFederationShared = "FederationShared"

	// ClusterStatusRunning 集群运行状态
	ClusterStatusRunning = "RUNNING"
)

// SharedClusterTypes 非独占的集群类型（包含普通共享集群，共享联邦集群）
var SharedClusterTypes = []string{ClusterTypeShared, ClusterTypeFederationShared}

// SharedClusterEnabledNativeKinds 共享集群支持的 k8s 原生资源
var SharedClusterEnabledNativeKinds = []string{
	resCsts.NS, resCsts.CJ, resCsts.Deploy, resCsts.Job, resCsts.Po, resCsts.STS, resCsts.HPA,
	resCsts.EP, resCsts.Ing, resCsts.SVC, resCsts.CM, resCsts.Secret, resCsts.PVC, resCsts.SA,
}

// SharedClusterBypassClusterScopedKinds 共享集群鉴权忽略的原生资源
var SharedClusterBypassClusterScopedKinds = []string{
	resCsts.SC,
	resCsts.CRD,
}
