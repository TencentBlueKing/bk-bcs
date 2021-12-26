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

// ActionIDNameMap map ActionID to name
var ActionIDNameMap = map[ActionID]string{
	ClusterScopedCreate: "集群域资源创建",
	ClusterScopedUpdate: "集群域资源更新",
	ClusterScopedDelete: "集群域资源删除",
	ClusterScopedView:   "集群域资源查看",
}

// GenerateActions generate registered action for iam
func GenerateActions() []ResourceAction {
	resourceActionList := make([]ResourceAction, 0)
	// add cluster actions
	resourceActionList = append(resourceActionList, generateClusterActions()...)

	return resourceActionList
}

func generateClusterActions() []ResourceAction {
	clusterSelection := []RelatedInstanceSelection{
		{
			SystemID:       SystemIDBKBCS,
			ID:             ClusterSelection,
			IgnoreAuthPath: false,
		},
	}

	relatedResource := []RelateResourceType{
		{
			SystemID:           SystemIDBKBCS,
			ID:                 SysCluster,
			NameAlias:          "",
			NameAliasEn:        "",
			InstanceSelections: clusterSelection,
		},
	}

	actions := make([]ResourceAction, 0)
	// register cluster scoped resource actions
	actions = append(actions, ResourceAction{
		ID:                   ClusterScopedCreate,
		Name:                 ActionIDNameMap[ClusterScopedCreate],
		NameEn:               "cluster manager role",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}
