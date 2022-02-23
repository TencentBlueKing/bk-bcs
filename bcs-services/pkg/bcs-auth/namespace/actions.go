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

package namespace

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

const (
	// NameSpaceCreate xxx
	NameSpaceCreate iam.ActionID = "namespace_create"
	// NameSpaceView xxx
	NameSpaceView iam.ActionID = "namespace_view"
	// NameSpaceUpdate xxx
	NameSpaceUpdate iam.ActionID = "namespace_update"
	// NameSpaceDelete xxx
	NameSpaceDelete iam.ActionID = "namespace_delete"
	// NameSpaceList xxx
	NameSpaceList iam.ActionID = "namespace_list"
	// NameSpaceUse xxx
	NameSpaceUse iam.ActionID = "namespace_use"
)

const (
	// NameSpaceScopedCreate xxx
	NameSpaceScopedCreate iam.ActionID = "namespace_scoped_create"
	// NameSpaceScopedView xxx
	NameSpaceScopedView iam.ActionID = "namespace_scoped_view"
	// NameSpaceScopedUpdate xxx
	NameSpaceScopedUpdate iam.ActionID = "namespace_scoped_update"
	// NameSpaceScopedDelete xxx
	NameSpaceScopedDelete iam.ActionID = "namespace_scoped_delete"
)

// ActionIDNameMap map ActionID to name
var ActionIDNameMap = map[iam.ActionID]string{
	NameSpaceCreate: "命名空间创建",
	NameSpaceView:   "命名空间查看",
	NameSpaceList:   "命令空间列举",
	NameSpaceUpdate: "命名空间更新",
	NameSpaceDelete: "命名空间删除",

	NameSpaceScopedCreate: "资源创建(命名空间域)",
	NameSpaceScopedUpdate: "资源更新(命名空间域)",
	NameSpaceScopedDelete: "资源删除(命名空间域)",
	NameSpaceScopedView:   "资源查看(命名空间域)",
}
