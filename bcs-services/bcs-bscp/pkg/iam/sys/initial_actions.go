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
	resourceActionList = append(resourceActionList, genApplicationActions()...)
	resourceActionList = append(resourceActionList, genCredentialActions()...)

	return resourceActionList
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
		RelatedActions:       []client.ActionID{BusinessViewResource},
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   AppEdit,
		Name:                 ActionIDNameMap[AppEdit],
		NameEn:               "Edit App",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       []client.ActionID{BusinessViewResource, AppView},
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   AppDelete,
		Name:                 ActionIDNameMap[AppDelete],
		NameEn:               "Delete App",
		Type:                 Delete,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       []client.ActionID{BusinessViewResource, AppView},
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   ReleaseGenerate,
		Name:                 ActionIDNameMap[ReleaseGenerate],
		NameEn:               "Generate Release",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       []client.ActionID{BusinessViewResource, AppView},
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   ReleasePublish,
		Name:                 ActionIDNameMap[ReleasePublish],
		NameEn:               "Publish Release",
		Type:                 Edit,
		RelatedResourceTypes: relatedResource,
		RelatedActions:       []client.ActionID{BusinessViewResource, AppView},
		Version:              1,
	})

	return actions
}

// genCredentialActions 服务密钥
func genCredentialActions() []client.ResourceAction {
	actions := make([]client.ResourceAction, 0)

	// actions = append(actions, client.ResourceAction{
	// 	ID:                   CredentialCreate,
	// 	Name:                 ActionIDNameMap[CredentialCreate],
	// 	NameEn:               "Create App Credential",
	// 	Type:                 Create,
	// 	RelatedResourceTypes: businessResource,
	// 	RelatedActions:       []client.ActionID{BusinessViewResource},
	// 	Version:              1,
	// })

	// relatedResource := []client.RelateResourceType{{
	// 	SystemID:    SystemIDBSCP,
	// 	ID:          Application,
	// 	NameAlias:   "",
	// 	NameAliasEn: "",
	// 	Scope:       nil,
	// 	InstanceSelections: []client.RelatedInstanceSelection{{
	// 		SystemID: SystemIDCMDB,
	// 		ID:       BusinessSelection,
	// 	}},
	// }}

	// actions = append(actions, client.ResourceAction{
	// 	ID:                   CredentialView,
	// 	Name:                 ActionIDNameMap[CredentialView],
	// 	NameEn:               "View App Credential",
	// 	Type:                 View,
	// 	RelatedResourceTypes: relatedResource,
	// 	RelatedActions:       []client.ActionID{BusinessViewResource},
	// 	Version:              1,
	// })

	// actions = append(actions, client.ResourceAction{
	// 	ID:                   CredentialEdit,
	// 	Name:                 ActionIDNameMap[CredentialEdit],
	// 	NameEn:               "Edit APP Credential",
	// 	Type:                 Edit,
	// 	RelatedResourceTypes: relatedResource,
	// 	RelatedActions:       []client.ActionID{BusinessViewResource, CredentialView},
	// 	Version:              1,
	// })

	// actions = append(actions, client.ResourceAction{
	// 	ID:                   CredentialDelete,
	// 	Name:                 ActionIDNameMap[CredentialDelete],
	// 	NameEn:               "Delete App Credential",
	// 	Type:                 Delete,
	// 	RelatedResourceTypes: relatedResource,
	// 	RelatedActions:       []client.ActionID{BusinessViewResource, CredentialView},
	// 	Version:              1,
	// })

	actions = append(actions, client.ResourceAction{
		ID:                   CredentialView,
		Name:                 ActionIDNameMap[CredentialView],
		NameEn:               "View App Credential",
		Type:                 View,
		RelatedResourceTypes: businessResource,
		RelatedActions:       []client.ActionID{BusinessViewResource},
		Version:              1,
	})

	actions = append(actions, client.ResourceAction{
		ID:                   CredentialManage,
		Name:                 ActionIDNameMap[CredentialManage],
		NameEn:               "Manage App Credential",
		Type:                 Manage,
		RelatedResourceTypes: businessResource,
		RelatedActions:       []client.ActionID{BusinessViewResource, CredentialView},
		Version:              1,
	})

	return actions

}

// genGroupActions 应用分组
//
//nolint:unused
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
		NameEn:               "View Business",
		Type:                 View,
		RelatedResourceTypes: businessResource,
		RelatedActions:       nil,
		Version:              1,
	})

	return actions
}
