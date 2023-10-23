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

// Package templateset xxx
package templateset

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

const (
	// TemplateSetCreate xxx
	TemplateSetCreate iam.ActionID = "templateset_create"
	// TemplateSetView xxx
	TemplateSetView iam.ActionID = "templateset_view"
	// TemplateSetUpdate xxx
	TemplateSetUpdate iam.ActionID = "templateset_update"
	// TemplateSetDelete xxx
	TemplateSetDelete iam.ActionID = "templateset_delete"
	// TemplateSetCopy xxx
	TemplateSetCopy iam.ActionID = "templateset_copy"
	// TemplateSetInstantiate xxx
	TemplateSetInstantiate iam.ActionID = "templateset_instantiate"
)

// ActionIDNameMap map ActionID to name
var ActionIDNameMap = map[iam.ActionID]string{
	TemplateSetCreate:      "模板集创建",
	TemplateSetView:        "模板集查看",
	TemplateSetCopy:        "模板集复制",
	TemplateSetUpdate:      "模板集更新",
	TemplateSetDelete:      "模板集删除",
	TemplateSetInstantiate: "模板集实例化",
}
