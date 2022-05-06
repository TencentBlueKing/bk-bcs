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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// Perm ...
type Perm struct {
	perm.IAMPerm
}

// NewPerm ...
func NewPerm(projectID string) *Perm {
	return &Perm{
		IAMPerm: perm.IAMPerm{
			Cli:           perm.NewIAMClient(),
			ResType:       perm.ResTypeCluster,
			PermCtx:       &PermCtx{},
			ResReq:        NewResRequest(projectID),
			ParentResPerm: &project.NewPerm().IAMPerm,
		},
	}
}

// CanList ...
func (p *Perm) CanList(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// CanView ...
func (p *Perm) CanView(ctx perm.Ctx) (bool, error) {
	allow, err := p.IAMPerm.CanAction(ctx, ClusterView, true)
	return project.RelatedProjectCanViewPerm(ctx, allow, err)
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
func (p *Perm) CanManage(ctx perm.Ctx) (bool, error) {
	allow, err := p.IAMPerm.CanAction(ctx, ClusterManage, false)
	return project.RelatedProjectCanViewPerm(ctx, allow, err)
}

// ScopedPerm ...
type ScopedPerm struct {
	perm perm.IAMPerm
}

// NewScopedPerm ...
func NewScopedPerm(projectID string) *ScopedPerm {
	return &ScopedPerm{
		perm: perm.IAMPerm{
			Cli:           perm.NewIAMClient(),
			ResType:       perm.ResTypeCluster,
			PermCtx:       &PermCtx{},
			ResReq:        NewResRequest(projectID),
			ParentResPerm: &project.NewPerm().IAMPerm,
		},
	}
}

// CanList ...
func (p *ScopedPerm) CanList(ctx perm.Ctx) (bool, error) {
	// 集群域资源 List 权限，与 View 权限相同
	return p.CanView(ctx)
}

// CanView ...
func (p *ScopedPerm) CanView(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{ClusterScopedView, ClusterView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return project.RelatedProjectCanViewPerm(ctx, allow, err)
}

// CanCreate ...
func (p *ScopedPerm) CanCreate(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{ClusterScopedCreate, ClusterView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return project.RelatedProjectCanViewPerm(ctx, allow, err)
}

// CanUpdate ...
func (p *ScopedPerm) CanUpdate(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{ClusterScopedUpdate, ClusterView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return project.RelatedProjectCanViewPerm(ctx, allow, err)
}

// CanDelete ...
func (p *ScopedPerm) CanDelete(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{ClusterScopedDelete, ClusterView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return project.RelatedProjectCanViewPerm(ctx, allow, err)
}

// CanUse ...
func (p *ScopedPerm) CanUse(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{
		ClusterScopedView,
		ClusterScopedCreate,
		ClusterScopedUpdate,
		ClusterScopedDelete,
		ClusterView,
	}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return project.RelatedProjectCanViewPerm(ctx, allow, err)
}

// CanManage ...
func (p *ScopedPerm) CanManage(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}
