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

package sys

import "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"

// GenerateStaticActionGroups generate all the static resource action groups.
func GenerateStaticActionGroups() []client.ActionGroup {
	ActionGroups := make([]client.ActionGroup, 0)

	// generate business Management action groups, contains business related actions
	ActionGroups = append(ActionGroups, genBusinessManagementActionGroups()...)

	return ActionGroups
}

func genBusinessManagementActionGroups() []client.ActionGroup {
	return []client.ActionGroup{
		{
			Name:   "业务",
			NameEn: "Business",
			Actions: []client.ActionWithID{
				{ID: BusinessViewResource},
			},
		},
		{
			Name:   "服务密钥",
			NameEn: "App Credential",
			Actions: []client.ActionWithID{
				// {ID: CredentialCreate},
				// {ID: CredentialView},
				// {ID: CredentialEdit},
				// {ID: CredentialDelete},
				{ID: CredentialView},
				{ID: CredentialManage},
			},
		},
		{
			Name:   "服务",
			NameEn: "Application",
			Actions: []client.ActionWithID{
				{ID: AppCreate},
				{ID: AppView},
				{ID: AppEdit},
				{ID: AppDelete},
				{ID: ReleaseGenerate},
				{ID: ReleasePublish},
			},
			// {
			// 	Name:   "分组管理",
			// 	NameEn: "Group Management",
			// 	Actions: []client.ActionWithID{ // 有应用编辑权限即可查看
			// 		{ID: GroupCreate},
			// 		{ID: GroupEdit},
			// 		{ID: GroupDelete},
			// 	},
			// },
		},
	}
}
