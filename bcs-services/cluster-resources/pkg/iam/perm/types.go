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

package perm

import bkiam "github.com/TencentBlueKing/iam-go-sdk"

// IAMRes ...
type IAMRes struct {
	ResType string
	ResID   string
}

// Ctx 权限校验 Context
type Ctx interface {
	// Validate 参数完整性校验
	Validate(actionIDs []string) error
	// GetProjID 获取项目 ID
	GetProjID() string
	// GetClusterID 获取集群 ID
	GetClusterID() string
	// GetResID 获取资源 ID
	GetResID() string
	// GetUsername 获取用户名
	GetUsername() string
	// GetParentChain 获取父节点信息
	GetParentChain() []IAMRes
	// SetForceRaise 标记强制无权限
	SetForceRaise()
	// ForceRaise 返回强制无权限标记
	ForceRaise() bool
	// ToMap 转换成 map 类型
	ToMap() map[string]interface{}
	// FromMap 根据 Map 数据更新 Context
	FromMap(m map[string]interface{}) Ctx
}

// Perm 权限校验接口定义
type Perm interface {
	// CanList 能否获取指定资源列表
	CanList(ctx Ctx) (bool, error)
	// CanView 能否查看指定（域）资源
	CanView(ctx Ctx) (bool, error)
	// CanCreate 能否创建指定（域）资源
	CanCreate(ctx Ctx) (bool, error)
	// CanUpdate 能否更新指定（域）资源
	CanUpdate(ctx Ctx) (bool, error)
	// CanDelete 能否删除指定（域）资源
	CanDelete(ctx Ctx) (bool, error)
	// CanUse 能否使用（CURD）指定（域）资源
	CanUse(ctx Ctx) (bool, error)
	// CanManage 能否管理资源（仅集群有效）
	CanManage(ctx Ctx) (bool, error)
}

// ResRequest 请求体接口定义
type ResRequest interface {
	// MakeResources 生成 ResourceNode 列表
	MakeResources(resIDs []string) []bkiam.ResourceNode
	// MakeAttribute 生成 ResourceNode.Attribute
	MakeAttribute(resID string) map[string]interface{}
	// FormMap 根据 map 数据更新
	FormMap(m map[string]interface{}) ResRequest
}
