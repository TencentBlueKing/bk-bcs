/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package iam

// PermCtx 权限校验 Context
type PermCtx struct {
	Username  string
	ProjectID string
	ClusterID string
	Namespace string
	// 命名空间唯一 ID
	NamespaceID string
}

// Perm 权限校验接口定义
type Perm interface {
	// CanView 能否查看指定域资源
	CanView(ctx PermCtx) (bool, error)
	// CanCreate 能否创建指定域资源
	CanCreate(ctx PermCtx) (bool, error)
	// CanUpdate 能否更新指定域资源
	CanUpdate(ctx PermCtx) (bool, error)
	// CanDelete 能否删除指定域资源
	CanDelete(ctx PermCtx) (bool, error)
	// CanUse 能否使用（CURD）指定域资源
	CanUse(ctx PermCtx) (bool, error)
	// permCtx 校验
	ValidateCtx(ctx PermCtx) error
}
