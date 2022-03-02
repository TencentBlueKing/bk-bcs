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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
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
)

var (
	// SharedClusterEnabledNativeKinds 共享集群支持的 k8s 原生资源
	SharedClusterEnabledNativeKinds = []string{
		res.NS, res.CJ, res.Deploy, res.Job, res.Po, res.STS, res.HPA,
		res.EP, res.Ing, res.SVC, res.CM, res.Secret, res.PVC, res.SA, res.CRD,
	}
	// SharedClusterAccessibleResKinds 共享集群支持的资源
	SharedClusterAccessibleResKinds = append(SharedClusterEnabledNativeKinds, envs.SharedClusterEnabledCObjKinds...)
)
