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

package config

import "fmt"

const (
	// TenantModeSingle 单租户模式，仅使用默认租户，但 bk_username 需通过 bk-user 转换
	TenantModeSingle = "single"
	// TenantModeMulti 多租户模式
	TenantModeMulti = "multi"
	// TenantModeOff 关闭租户
	TenantModeOff = "off"
)

// GetTenantMode 返回租户模式 single / multi，关闭租户时返回 off
func (c *ProjectConfig) GetTenantMode() string {
	switch c.Tenant {
	case TenantModeSingle, TenantModeMulti:
		return c.Tenant
	}
	return TenantModeOff
}

// ValidateTenant 校验租户模式取值
func (c *ProjectConfig) ValidateTenant() error {
	switch c.Tenant {
	case "", TenantModeOff, TenantModeSingle, TenantModeMulti:
		return nil
	}
	return fmt.Errorf("invalid tenant mode %q, must be one of: single, multi, off or empty", c.Tenant)
}

// IsMultiTenantEnabled 是否开启多租户模式
func (c *ProjectConfig) IsMultiTenantEnabled() bool {
	return c.GetTenantMode() == TenantModeMulti
}

// IsTenantEnabled 是否启用了租户功能（单租户或多租户）
func (c *ProjectConfig) IsTenantEnabled() bool {
	return c.GetTenantMode() != TenantModeOff
}
