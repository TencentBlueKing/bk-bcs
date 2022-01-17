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
	ProjectCreate: "项目创建",
	ProjectView: "项目查看",
	ProjectEdit: "项目编辑",
	ProjectDelete: "项目删除",

	ClusterCreate: "集群创建",
	ClusterView: "集群查看",
	ClusterManage: "集群管理",
	ClusterUse: "集群使用",
	ClusterDelete: "集群删除",

	NameSpaceCreate: "命名空间创建",
	NameSpaceView: "命名空间查看",
	NameSpaceList: "命令空间列举",
	NameSpaceUpdate: "命名空间更新",
	NameSpaceDelete: "命名空间删除",

	ClusterScopedCreate: "资源创建",
	ClusterScopedUpdate: "资源更新",
	ClusterScopedDelete: "资源删除",
	ClusterScopedView:   "资源查看",

	NameSpaceScopedCreate: "资源创建(命名空间域)",
	NameSpaceScopedUpdate: "资源更新(命名空间域)",
	NameSpaceScopedDelete: "资源删除(命名空间域)",
	NameSpaceScopedView:   "资源查看(命名空间域)",
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
