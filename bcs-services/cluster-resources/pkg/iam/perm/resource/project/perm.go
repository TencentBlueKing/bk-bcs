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

package project

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// Perm NOTE ClusterResources 不对项目进行管理，因此只实现 CanView 方法
type Perm struct {
	perm.IAMPerm
}

// NewPerm ...
func NewPerm() *Perm {
	return &Perm{
		IAMPerm: perm.IAMPerm{
			Cli:     perm.NewIAMClient(),
			ResType: perm.ResTypeProj,
			PermCtx: &PermCtx{},
			ResReq:  NewResRequest(),
		},
	}
}

// CanList ...
func (p *Perm) CanList(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// CanView ...
func (p *Perm) CanView(ctx perm.Ctx) (bool, error) {
	return p.IAMPerm.CanAction(ctx, ProjectView, true)
}

// CanCreate ...
func (p *Perm) CanCreate(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// CanUpdate ...
func (p *Perm) CanUpdate(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// CanDelete ...
func (p *Perm) CanDelete(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// CanUse ...
func (p *Perm) CanUse(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// CanManage ...
func (p *Perm) CanManage(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}
