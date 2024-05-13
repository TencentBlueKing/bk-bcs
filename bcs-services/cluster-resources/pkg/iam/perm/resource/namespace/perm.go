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

package namespace

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// Perm xxx
type Perm struct {
	perm.IAMPerm
}

// NewPerm xxx
func NewPerm(projectID, clusterID string) *Perm {
	return &Perm{
		IAMPerm: perm.IAMPerm{
			Cli:           perm.NewIAMClient(),
			ResType:       perm.ResTypeNS,
			PermCtx:       &PermCtx{},
			ResReq:        NewResRequest(projectID, clusterID, ""),
			ParentResPerm: &cluster.NewPerm(projectID).IAMPerm,
		},
	}
}

// CanList xxx
func (p *Perm) CanList(ctx perm.Ctx) (bool, error) {
	allow, err := p.IAMPerm.CanAction(ctx, NamespaceList, false)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanView xxx
func (p *Perm) CanView(ctx perm.Ctx) (bool, error) {
	allow, err := p.IAMPerm.CanAction(ctx, NamespaceView, false)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanCreate xxx
func (p *Perm) CanCreate(ctx perm.Ctx) (bool, error) {
	allow, err := p.IAMPerm.CanAction(ctx, NamespaceCreate, false)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanUpdate xxx
func (p *Perm) CanUpdate(ctx perm.Ctx) (bool, error) {
	allow, err := p.IAMPerm.CanAction(ctx, NamespaceUpdate, false)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanDelete xxx
func (p *Perm) CanDelete(ctx perm.Ctx) (bool, error) {
	allow, err := p.IAMPerm.CanAction(ctx, NamespaceDelete, false)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanUse xxx
func (p *Perm) CanUse(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// CanManage xxx
func (p *Perm) CanManage(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}

// ScopedPerm xxx
type ScopedPerm struct {
	perm perm.IAMPerm
}

// NewScopedPerm xxx
func NewScopedPerm(projectID, clusterID, res string) *ScopedPerm {
	return &ScopedPerm{
		perm: perm.IAMPerm{
			Cli:           perm.NewIAMClient(),
			ResType:       perm.ResTypeNS,
			PermCtx:       &PermCtx{},
			ResReq:        NewResRequest(projectID, clusterID, res),
			ParentResPerm: &cluster.NewPerm(projectID).IAMPerm,
		},
	}
}

// CanList xxx
func (p *ScopedPerm) CanList(ctx perm.Ctx) (bool, error) {
	// 命名空间域资源 List 权限，与 View 权限相同
	return p.CanView(ctx)
}

// CanView xxx
func (p *ScopedPerm) CanView(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{NamespaceScopedView, NamespaceView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanCreate xxx
func (p *ScopedPerm) CanCreate(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{NamespaceScopedCreate, NamespaceView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanUpdate xxx
func (p *ScopedPerm) CanUpdate(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{NamespaceScopedUpdate, NamespaceView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanDelete xxx
func (p *ScopedPerm) CanDelete(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{NamespaceScopedDelete, NamespaceView}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanUse xxx
func (p *ScopedPerm) CanUse(ctx perm.Ctx) (bool, error) {
	actionIDs := []string{
		NamespaceScopedView,
		NamespaceScopedCreate,
		NamespaceScopedUpdate,
		NamespaceScopedDelete,
		NamespaceView,
	}
	allow, err := p.perm.CanMultiActions(ctx, actionIDs)
	return cluster.RelatedClusterCanViewPerm(ctx, allow, err)
}

// CanManage xxx
func (p *ScopedPerm) CanManage(_ perm.Ctx) (bool, error) {
	return false, errorx.New(errcode.Unsupported, "perm validate unsupported")
}
