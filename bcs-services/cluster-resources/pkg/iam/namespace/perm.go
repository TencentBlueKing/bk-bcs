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

// NewPerm ...
func NewPerm() *Perm {
	return &Perm{}
}

// Perm ...
type Perm struct{}

// CanView ...
func (p *Perm) CanView(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanCreate ...
func (p *Perm) CanCreate(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanUpdate ...
func (p *Perm) CanUpdate(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanDelete ...
func (p *Perm) CanDelete(ctx crIAM.PermCtx) (bool, error) {
	return true, nil
}

// CanUse ...
func (p *Perm) CanUse(ctx crIAM.PermCtx) (bool, error) {
	panic("not implement")
}

// ValidateCtx 校验 PermCtx 是否缺失参数
func (p *Perm) ValidateCtx(ctx crIAM.PermCtx) error {
	return nil
}
