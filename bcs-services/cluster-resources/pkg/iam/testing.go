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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// NoPermErr ...
var NoPermErr = errorx.New(errcode.NoPerm, "no permission")

// MockPerm ...
type MockPerm struct{}

// CanView ...
func (p *MockPerm) CanView(permCtx PermCtx) (bool, error) {
	if p.forceNoPerm(permCtx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanCreate ...
func (p *MockPerm) CanCreate(permCtx PermCtx) (bool, error) {
	if p.forceNoPerm(permCtx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanUpdate ...
func (p *MockPerm) CanUpdate(permCtx PermCtx) (bool, error) {
	if p.forceNoPerm(permCtx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanDelete ...
func (p *MockPerm) CanDelete(permCtx PermCtx) (bool, error) {
	if p.forceNoPerm(permCtx) {
		return false, NoPermErr
	}
	return true, nil
}

// CanUse ...
func (p *MockPerm) CanUse(permCtx PermCtx) (bool, error) {
	if p.forceNoPerm(permCtx) {
		return false, NoPermErr
	}
	return true, nil
}

// ValidateCtx 校验 PermCtx 是否缺失参数
func (p *MockPerm) ValidateCtx(_ PermCtx) error {
	return nil
}

// 单测用，若指定参数符合条件，则强制无权限
func (p *MockPerm) forceNoPerm(permCtx PermCtx) bool {
	return permCtx.ClusterID == envs.TestNoPermClusterID || permCtx.Namespace == envs.TestNoPermNS
}
