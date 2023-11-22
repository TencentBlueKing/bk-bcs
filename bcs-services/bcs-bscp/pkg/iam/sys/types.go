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

import "bscp.io/pkg/iam/client"

const (
	// SystemIDBSCP is bscp in iam's system id.
	SystemIDBSCP = "bk-bscp"
	// SystemNameBSCPEn is bscp in iam's system english name.
	SystemNameBSCPEn = "bscp"
	// SystemNameBSCP is bscp in iam's system name.
	SystemNameBSCP = "服务配置中心"

	// SystemIDCMDB is cmdb in iam's system id.
	SystemIDCMDB = "bk_cmdb"
	// SystemNameCMDB is cmdb system name in iam.
	SystemNameCMDB = "配置平台"
)

// SystemIDNameMap is system id to name map.
var SystemIDNameMap = map[string]string{
	SystemIDBSCP: SystemNameBSCP,
	SystemIDCMDB: SystemNameCMDB,
}

// TypeID resource type to register iam.
const (
	Business      client.TypeID = "biz"
	Application   client.TypeID = "app"
	AppCredential client.TypeID = "app_credential" //nolint:gosec
)

// ActionID action id to register iam.
const (
	// BusinessViewResource business view.
	BusinessViewResource client.ActionID = "find_business_resource"

	// AppCreate app create.
	AppCreate client.ActionID = "app_create"
	// AppView
	AppView client.ActionID = "app_view"
	// AppEdit app edit.
	AppEdit client.ActionID = "app_edit"
	// AppDelete app delete.
	AppDelete client.ActionID = "app_delete"
	// ReleaseGenerate generate release.
	ReleaseGenerate client.ActionID = "release_generate"
	// ReleasePublish release publish.
	ReleasePublish client.ActionID = "release_publish"
	// ConfigItemFinishPublish config item finish publish.
	ConfigItemFinishPublish client.ActionID = "config_item_finish_publish"

	// StrategySetCreate strategy set create.
	StrategySetCreate client.ActionID = "strategy_set_create"
	// StrategySetEdit strategy set edit.
	StrategySetEdit client.ActionID = "strategy_set_edit"
	// StrategySetDelete strategy set delete.
	StrategySetDelete client.ActionID = "strategy_set_delete"

	// StrategyCreate strategy create.
	StrategyCreate client.ActionID = "strategy_create"
	// StrategyEdit strategy edit.
	StrategyEdit client.ActionID = "strategy_edit"
	// StrategyDelete strategy delete.
	StrategyDelete client.ActionID = "strategy_delete"

	// TaskHistoryView task history view.
	TaskHistoryView client.ActionID = "history_view"

	// GroupCreate 分组创建
	GroupCreate client.ActionID = "group_create"
	// GroupDelete 分组删除
	GroupDelete client.ActionID = "group_delete"
	// GroupEdit 分组编辑
	GroupEdit client.ActionID = "group_edit"

	// Unsupported is an action that can not be recognized
	Unsupported client.ActionID = "unsupported"
	// Skip is an action that no need to auth
	Skip client.ActionID = "skip"

	// CredentialView 服务密钥查看
	CredentialView client.ActionID = "app_credential_view" //nolint:gosec
	// CredentialManage 服务密钥管理
	CredentialManage client.ActionID = "app_credential_manage" //nolint:gosec
)

// ActionIDNameMap is action id type map.
var ActionIDNameMap = map[client.ActionID]string{
	BusinessViewResource: "业务访问",

	AppCreate:               "服务创建",
	AppView:                 "服务查看",
	AppEdit:                 "服务编辑",
	AppDelete:               "服务删除",
	ReleaseGenerate:         "生成版本",
	ReleasePublish:          "上线版本",
	ConfigItemFinishPublish: "配置项结束发布",

	StrategySetCreate: "策略集创建",
	StrategySetEdit:   "策略集编辑",
	StrategySetDelete: "策略集删除",

	StrategyCreate: "策略创建",
	StrategyEdit:   "策略编辑",
	StrategyDelete: "策略删除",

	GroupCreate: "分组创建",
	GroupEdit:   "分组编辑",
	GroupDelete: "分组删除",

	TaskHistoryView: "任务历史",

	CredentialView:   "服务秘钥查看",
	CredentialManage: "服务秘钥管理",
}

// InstanceSelectionID selection id to register iam.
const (
	BusinessSelection    client.InstanceSelectionID = "business"
	ApplicationSelection client.InstanceSelectionID = "application"
)

// ActionType action type to register iam.
const (
	Create client.ActionType = "create"
	Delete client.ActionType = "delete"
	View   client.ActionType = "view"
	Edit   client.ActionType = "edit"
	Manage client.ActionType = "manage"
	List   client.ActionType = "list"
)

// ActionTypeIDNameMap is action type map.
var ActionTypeIDNameMap = map[client.ActionType]string{
	Create: "新建",
	Edit:   "编辑",
	Delete: "删除",
	View:   "查询",
	Manage: "管理",
	List:   "列表查询",
}
