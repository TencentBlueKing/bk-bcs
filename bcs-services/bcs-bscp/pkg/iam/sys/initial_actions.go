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

var (
	// 业务资源, 自动拉取 cmdb 业务列表
	businessResource = []client.RelateResourceType{
		{
			SystemID: SystemIDCMDB,
			ID:       Business,
			InstanceSelections: []client.RelatedInstanceSelection{
				{
					SystemID: SystemIDCMDB,
					ID:       BusinessSelection,
				},
			},
		},
	}
)

// GenerateStaticActions return need to register action.
func GenerateStaticActions() []client.ResourceAction {
	resourceActionList := make([]client.ResourceAction, 0)

	resourceActionList = append(resourceActionList, genBusinessActions()...)
	// resourceActionList = append(resourceActionList, genApplicationActions()...)
	// resourceActionList = append(resourceActionList, genGroupActions()...)
	// resourceActionList = append(resourceActionList, genStrategySetActions()...)
	// resourceActionList = append(resourceActionList, genStrategyActions()...)
	// resourceActionList = append(resourceActionList, genHistoryActions()...)

	return resourceActionList
}

func genHistoryActions() []client.ResourceAction {
	actions := make([]client.ResourceAction, 0)

	actions = append(actions, client.ResourceAction{
		ID:                   TaskHistoryView,
		Name:                 ActionIDNameMap[TaskHistoryView],
		NameEn:               "Task History",
		Type:                 View,
		RelatedResourceTypes: businessResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}

func genStrategyActions() []client.ResourceAction {
	actions := make([]client.ResourceAction, 0)

	actions = append(actions, client.ResourceAction{
		ID:     StrategyCreate,
		Name:   ActionIDNameMap[StrategyCreate],
		NameEn: "Create Strategy",
		Type:   Create,
		RelatedResourceTypes: []client.RelateResourceType{{
			SystemID:    SystemIDBSCP,
			ID:          Application,
			NameAlias:   "",
			NameAliasEn: "",
			Scope:       nil,
			InstanceSelections: []client.RelatedInstanceSelection{{
				SystemID: SystemIDBSCP,
				ID:       ApplicationSelection,
			}},
		}},
		RelatedActions: []client.ActionID{BusinessViewResource},
		Version:        1,
	})

	relatedResource := []client.RelateResourceType{{
		SystemID:    SystemIDBSCP,
		ID:          Application,
		NameAlias:   "",
		NameAliasEn: "",
		Scope:       nil,
		InstanceSelections: []client.RelatedInstanceSelection{{
			SystemID: SystemIDBSCP,
			ID:       ApplicationSelection,
		}},
	}}

	actions = append(actions, client.ResourceAction{
		ID:                   StrategyEdit,
		Name:                 ActionIDNameMap[StrategyEdit],
		NameEn:               "Edit Strategy",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   StrategyDelete,
		Name:                 ActionIDNameMap[StrategyDelete],
		NameEn:               "Delete Strategy",
		Type:                 Delete,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}

func genStrategySetActions() []client.ResourceAction {
	actions := make([]client.ResourceAction, 0)

	actions = append(actions, client.ResourceAction{
		ID:     StrategySetCreate,
		Name:   ActionIDNameMap[StrategySetCreate],
		NameEn: "Create Strategy Set",
		Type:   Create,
		RelatedResourceTypes: []client.RelateResourceType{{
			SystemID:    SystemIDBSCP,
			ID:          Application,
			NameAlias:   "",
			NameAliasEn: "",
			Scope:       nil,
			InstanceSelections: []client.RelatedInstanceSelection{{
				SystemID: SystemIDBSCP,
				ID:       ApplicationSelection,
			}},
		}},
		RelatedActions: []client.ActionID{BusinessViewResource},
		Version:        1,
	})

	relatedResource := []client.RelateResourceType{{
		SystemID:    SystemIDBSCP,
		ID:          Application,
		NameAlias:   "",
		NameAliasEn: "",
		Scope:       nil,
		InstanceSelections: []client.RelatedInstanceSelection{{
			SystemID: SystemIDBSCP,
			ID:       ApplicationSelection,
		}},
	}}

	actions = append(actions, client.ResourceAction{
		ID:                   StrategySetEdit,
		Name:                 ActionIDNameMap[StrategySetEdit],
		NameEn:               "Edit Strategy Set",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   StrategySetDelete,
		Name:                 ActionIDNameMap[StrategySetDelete],
		NameEn:               "Delete Strategy Set",
		Type:                 Delete,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}

func genApplicationActions() []client.ResourceAction {
	actions := make([]client.ResourceAction, 0)

	actions = append(actions, client.ResourceAction{
		ID:                   AppCreate,
		Name:                 ActionIDNameMap[AppCreate],
		NameEn:               "Create App",
		Type:                 Create,
		RelatedResourceTypes: businessResource,
		RelatedActions:       []client.ActionID{BusinessViewResource},
		Version:              1,
	})

	relatedResource := []client.RelateResourceType{{
		SystemID:    SystemIDBSCP,
		ID:          Application,
		NameAlias:   "",
		NameAliasEn: "",
		Scope:       nil,
		InstanceSelections: []client.RelatedInstanceSelection{{
			SystemID: SystemIDBSCP,
			ID:       ApplicationSelection,
		}},
	}}

	actions = append(actions, client.ResourceAction{
		ID:                   AppView,
		Name:                 ActionIDNameMap[AppView],
		NameEn:               "View App",
		Type:                 View,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   AppEdit,
		Name:                 ActionIDNameMap[AppEdit],
		NameEn:               "Edit App",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   AppDelete,
		Name:                 ActionIDNameMap[AppDelete],
		NameEn:               "Delete App",
		Type:                 Delete,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   ConfigItemPacking,
		Name:                 ActionIDNameMap[ConfigItemPacking],
		NameEn:               "Packing ConfigItem",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   ConfigItemPublish,
		Name:                 ActionIDNameMap[ConfigItemPublish],
		NameEn:               "Publish ConfigItem",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}

// genGroupActions 应用分组
func genGroupActions() []client.ResourceAction {
	actions := make([]client.ResourceAction, 0)

	actions = append(actions, client.ResourceAction{
		ID:     GroupCreate,
		Name:   ActionIDNameMap[GroupCreate],
		NameEn: "Create Group",
		Type:   Create,
		RelatedResourceTypes: []client.RelateResourceType{{
			SystemID:    SystemIDBSCP,
			ID:          Application,
			NameAlias:   "",
			NameAliasEn: "",
			Scope:       nil,
			InstanceSelections: []client.RelatedInstanceSelection{{
				SystemID: SystemIDBSCP,
				ID:       ApplicationSelection,
			}},
		}},
		RelatedActions: []client.ActionID{BusinessViewResource},
		Version:        1,
	})

	relatedResource := []client.RelateResourceType{{
		SystemID:    SystemIDBSCP,
		ID:          Application,
		NameAlias:   "",
		NameAliasEn: "",
		Scope:       nil,
		InstanceSelections: []client.RelatedInstanceSelection{{
			SystemID: SystemIDBSCP,
			ID:       ApplicationSelection,
		}},
	}}

	actions = append(actions, client.ResourceAction{
		ID:                   GroupEdit,
		Name:                 ActionIDNameMap[GroupEdit],
		NameEn:               "Edit Group",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   GroupDelete,
		Name:                 ActionIDNameMap[GroupDelete],
		NameEn:               "Delete Group",
		Type:                 Delete,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}

func genBusinessActions() []client.ResourceAction {
	actions := make([]client.ResourceAction, 0)

	actions = append(actions, client.ResourceAction{
		ID:                   BusinessViewResource,
		Name:                 ActionIDNameMap[BusinessViewResource],
		NameEn:               "View Business Resource",
		Type:                 View,
		RelatedResourceTypes: businessResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}
