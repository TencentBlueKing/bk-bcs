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

// GenerateActionGroups build action groups
func GenerateActionGroups() []ActionGroup {
	actionGroupList := make([]ActionGroup, 0)

	// generate bcs cluster actions group
	actionGroupList = append(actionGroupList, generateBcsClusterActionsGroup()...)

	return actionGroupList
}

func generateBcsClusterActionsGroup() []ActionGroup {
	return []ActionGroup{
		{
			Name:      "集群域权限",
			NameEn:    "Cluster Permissions Apply",
			SubGroups: nil,
			Actions:   []ActionWithID{ },
		},
		{
			Name:      "命名空间域权限",
			NameEn:    "Cluster Namespace Permissions Apply",
			SubGroups: nil,
			Actions:   []ActionWithID{},
		},
	}
}
