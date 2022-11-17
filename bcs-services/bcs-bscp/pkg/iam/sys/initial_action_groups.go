/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sys

import "bscp.io/pkg/iam/client"

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
			Name:   "业务管理",
			NameEn: "Business Management",
			Actions: []client.ActionWithID{
				{ID: BusinessViewResource},
			},
		},
		{
			Name:   "配置管理",
			NameEn: "Configuration Management",
			SubGroups: []client.ActionGroup{
				{
					Name:   "应用管理",
					NameEn: "Application Management",
					Actions: []client.ActionWithID{
						{ID: AppCreate},
						{ID: AppEdit},
						{ID: AppDelete},
						{ID: ConfigItemPacking},
						{ID: ConfigItemPublish},
						{ID: ConfigItemFinishPublish},
					},
				},
			},
		},
		{
			Name:   "策略管理",
			NameEn: "Strategy Management",
			SubGroups: []client.ActionGroup{
				{
					Name:   "策略集管理",
					NameEn: "StrategySet Management",
					Actions: []client.ActionWithID{
						{ID: StrategySetCreate},
						{ID: StrategySetEdit},
						{ID: StrategySetDelete},
					},
				},
				{
					Name:   "策略管理",
					NameEn: "Strategy Management",
					Actions: []client.ActionWithID{
						{ID: StrategyCreate},
						{ID: StrategyEdit},
						{ID: StrategyDelete},
					},
				},
			},
		},
		{
			Name:   "任务历史",
			NameEn: "Task History",
			SubGroups: []client.ActionGroup{
				{
					Name:   "任务历史",
					NameEn: "Task History",
					Actions: []client.ActionWithID{
						{ID: TaskHistoryView},
					},
				},
			},
		},
	}
}
