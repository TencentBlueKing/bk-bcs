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

package namespace

import (
	crIAM "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
)

// NewScopedPerm ...
func NewScopedPerm() *ScopedPerm {
	return &ScopedPerm{}
}

// ScopedPerm ...
type ScopedPerm struct{}

// CanView ...
func (p *ScopedPerm) CanView(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanCreate ...
func (p *ScopedPerm) CanCreate(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanUpdate ...
func (p *ScopedPerm) CanUpdate(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanDelete ...
func (p *ScopedPerm) CanDelete(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanUse ...
func (p *ScopedPerm) CanUse(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// ValidateCtx 校验 PermCtx 是否缺失参数
func (p *ScopedPerm) ValidateCtx(ctx crIAM.PermCtx) error {
	return nil
}
