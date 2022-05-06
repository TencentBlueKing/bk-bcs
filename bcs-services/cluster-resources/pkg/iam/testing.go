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

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// NoPermErr ...
var NoPermErr = errorx.New(errcode.NoPerm, "no permission")

// MockPerm ...
type MockPerm struct{}

// CanList ...
func (p *MockPerm) CanList(ctx perm.Ctx) (bool, error) {
	if p.forceNoPerm(ctx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanView ...
func (p *MockPerm) CanView(ctx perm.Ctx) (bool, error) {
	if p.forceNoPerm(ctx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanCreate ...
func (p *MockPerm) CanCreate(ctx perm.Ctx) (bool, error) {
	if p.forceNoPerm(ctx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanUpdate ...
func (p *MockPerm) CanUpdate(ctx perm.Ctx) (bool, error) {
	if p.forceNoPerm(ctx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanDelete ...
func (p *MockPerm) CanDelete(ctx perm.Ctx) (bool, error) {
	if p.forceNoPerm(ctx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanUse ...
func (p *MockPerm) CanUse(ctx perm.Ctx) (bool, error) {
	if p.forceNoPerm(ctx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanManage ...
func (p *MockPerm) CanManage(ctx perm.Ctx) (bool, error) {
	if p.forceNoPerm(ctx) {
		return false, NoPermErr
	}
	return true, nil
}

// 单测用，若指定参数符合条件，则强制无权限
func (p *MockPerm) forceNoPerm(ctx perm.Ctx) bool {
	return ctx.GetClusterID() == envs.TestNoPermClusterID || ctx.ForceRaise()
}
