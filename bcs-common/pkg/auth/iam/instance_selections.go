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
 *
 */

package iam

var (
	projectChain = ResourceChain{
		SystemID: SystemIDBKBCS,
		ID:       SysProject,
	}
	clusterChain = ResourceChain{
		SystemID: SystemIDBKBCS,
		ID:       SysCluster,
	}
	namespaceChain = ResourceChain{
		SystemID: SystemIDBKBCS,
		ID:       SysNamespace,
	}
)

// GenerateInstanceSelections generate all instance selections registered to iam
func GenerateInstanceSelections() []InstanceSelection {
	return []InstanceSelection{
		{
			ID:                ClusterSelection,
			Name:              "集群列表",
			NameEn:            "Cluster List",
			ResourceTypeChain: []ResourceChain{projectChain, clusterChain},
		},
		{
			ID:     NamespaceSelection,
			Name:   "命名空间列表",
			NameEn: "Namespace List",
			ResourceTypeChain: []ResourceChain{
				projectChain,
				clusterChain,
				namespaceChain,
			},
		},
	}
}
