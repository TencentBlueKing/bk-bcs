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

import "testing"

func TestTenantMode(t *testing.T) {
	bkUser := BcsClusterReporterOptions{BkUserServer: "http://bkuser", BkAppCode: "code", BkAppSecret: "secret"}
	tests := []struct {
		name      string
		opts      BcsClusterReporterOptions
		wantMode  string
		wantMulti bool
		wantErr   bool
	}{
		{name: "empty", wantMode: TenantModeOff},
		{name: "off", opts: BcsClusterReporterOptions{Tenant: "off"}, wantMode: TenantModeOff},
		{name: "single", opts: BcsClusterReporterOptions{Tenant: "single"}, wantMode: TenantModeSingle},
		{name: "multi", opts: func() BcsClusterReporterOptions { o := bkUser; o.Tenant = "multi"; return o }(),
			wantMode: TenantModeMulti, wantMulti: true},
		{name: "multi without bk-user", opts: BcsClusterReporterOptions{Tenant: "multi"},
			wantMode: TenantModeMulti, wantMulti: true, wantErr: true},
		{name: "invalid", opts: BcsClusterReporterOptions{Tenant: "true"}, wantMode: TenantModeOff, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.GetTenantMode(); got != tt.wantMode {
				t.Errorf("GetTenantMode() = %s, want %s", got, tt.wantMode)
			}
			if got := tt.opts.IsMultiTenantEnabled(); got != tt.wantMulti {
				t.Errorf("IsMultiTenantEnabled() = %v, want %v", got, tt.wantMulti)
			}
			if err := tt.opts.ValidateTenant(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTenant() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
