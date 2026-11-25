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

import "testing"

func TestBaseConfGetTenantMode(t *testing.T) {
	cases := []struct {
		name  string
		conf  BaseConf
		mode  string
		multi bool
	}{
		{name: "empty", conf: BaseConf{}, mode: TenantModeOff},
		{name: "off", conf: BaseConf{Tenant: TenantModeOff}, mode: TenantModeOff},
		{name: "single", conf: BaseConf{Tenant: TenantModeSingle}, mode: TenantModeSingle},
		{name: "multi", conf: BaseConf{Tenant: TenantModeMulti}, mode: TenantModeMulti, multi: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.conf.GetTenantMode(); got != c.mode {
				t.Fatalf("GetTenantMode() = %q, want %q", got, c.mode)
			}
			if got := c.conf.IsMultiTenantEnabled(); got != c.multi {
				t.Fatalf("IsMultiTenantEnabled() = %v, want %v", got, c.multi)
			}
		})
	}
}

func TestBaseConfValidateTenant(t *testing.T) {
	for _, mode := range []string{"", TenantModeOff, TenantModeSingle, TenantModeMulti} {
		if err := (&BaseConf{Tenant: mode}).ValidateTenant(); err != nil {
			t.Fatalf("ValidateTenant(%q) unexpected error: %v", mode, err)
		}
	}
	if err := (&BaseConf{Tenant: "true"}).ValidateTenant(); err == nil {
		t.Fatal("ValidateTenant(\"true\") should return error")
	}
}
