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

package tenant

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

func ensureCMOptions() *options.ClusterManagerOptions {
	if options.GetGlobalCMOptions() == nil {
		options.SetGlobalCMOptions(&options.ClusterManagerOptions{})
	}
	return options.GetGlobalCMOptions()
}

func TestIsMultiTenantEnabled(t *testing.T) {
	opt := ensureCMOptions()
	opt.TenantConfig.EnableMultiTenantMode = false
	if IsMultiTenantEnabled() {
		t.Fatal("IsMultiTenantEnabled should be false by default")
	}

	opt.TenantConfig.EnableMultiTenantMode = true
	t.Cleanup(func() {
		opt.TenantConfig.EnableMultiTenantMode = false
	})
	if !IsMultiTenantEnabled() {
		t.Fatal("IsMultiTenantEnabled should be true when TenantConfig is enabled")
	}
}
