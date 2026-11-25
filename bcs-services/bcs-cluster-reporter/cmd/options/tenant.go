/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package options

import "fmt"

const (
	// TenantModeSingle 单租户模式
	TenantModeSingle = "single"
	// TenantModeMulti 多租户模式
	TenantModeMulti = "multi"
	// TenantModeOff 关闭租户
	TenantModeOff = "off"
)

// GetTenantMode return single / multi / off
func (bcro *BcsClusterReporterOptions) GetTenantMode() string {
	switch bcro.Tenant {
	case TenantModeSingle, TenantModeMulti:
		return bcro.Tenant
	default:
		return TenantModeOff
	}
}

// ValidateTenant validate tenant config
func (bcro *BcsClusterReporterOptions) ValidateTenant() error {
	switch bcro.Tenant {
	case "", TenantModeOff, TenantModeSingle, TenantModeMulti:
	default:
		return fmt.Errorf("invalid tenant %q, must be one of: single, multi, off or empty", bcro.Tenant)
	}
	if bcro.IsMultiTenantEnabled() &&
		(bcro.BkUserServer == "" || bcro.BkAppCode == "" || bcro.BkAppSecret == "") {
		return fmt.Errorf("bkUserServer, bkAppCode, bkAppSecret must be set in multi tenant mode")
	}
	return nil
}

// IsMultiTenantEnabled multi tenant mode enabled
func (bcro *BcsClusterReporterOptions) IsMultiTenantEnabled() bool {
	return bcro.GetTenantMode() == TenantModeMulti
}
