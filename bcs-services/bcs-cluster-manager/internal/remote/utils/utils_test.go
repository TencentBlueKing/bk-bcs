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

package utils

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
)

func ensureCMOptions() *options.ClusterManagerOptions {
	if options.GetGlobalCMOptions() == nil {
		options.SetGlobalCMOptions(&options.ClusterManagerOptions{})
	}
	return options.GetGlobalCMOptions()
}

func TestGetGatewayAuthAndTenantInfoSkipLookupWhenMultiTenantDisabled(t *testing.T) {
	opt := ensureCMOptions()
	opt.TenantConfig.EnableMultiTenantMode = false
	t.Cleanup(func() {
		opt.TenantConfig.EnableMultiTenantMode = false
	})

	auth := &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   "test-code",
			BkAppSecret: "test-secret",
		},
		BkUserName: "bcs-cluster-manager",
	}

	userAuth, tenantID, err := GetGatewayAuthAndTenantInfo(context.Background(), auth, "")
	if err != nil {
		t.Fatalf("GetGatewayAuthAndTenantInfo returned error: %v", err)
	}
	if auth.BkUserName != "bcs-cluster-manager" {
		t.Fatalf("BkUserName changed when multi-tenant disabled, got %s", auth.BkUserName)
	}
	if tenantID == "" {
		t.Fatal("tenantID should not be empty")
	}

	got := &types.AuthInfo{}
	if err := json.Unmarshal([]byte(userAuth), got); err != nil {
		t.Fatalf("unmarshal userAuth: %v", err)
	}
	if got.BkUserName != "bcs-cluster-manager" {
		t.Fatalf("auth json BkUserName = %s, want bcs-cluster-manager", got.BkUserName)
	}
}

func TestGetGatewayAuthAndTenantInfoExplicitUserSkipsLookup(t *testing.T) {
	opt := ensureCMOptions()
	opt.TenantConfig.EnableMultiTenantMode = true
	t.Cleanup(func() {
		opt.TenantConfig.EnableMultiTenantMode = false
	})

	auth := &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   "test-code",
			BkAppSecret: "test-secret",
		},
		BkUserName: "bcs-cluster-manager",
	}

	_, _, err := GetGatewayAuthAndTenantInfo(context.Background(), auth, "explicit-user")
	if err != nil {
		t.Fatalf("GetGatewayAuthAndTenantInfo returned error: %v", err)
	}
	if auth.BkUserName != "explicit-user" {
		t.Fatalf("BkUserName = %s, want explicit-user", auth.BkUserName)
	}
}

func TestGetGatewayAuthAndTenantInfoLookupWhenMultiTenantEnabled(t *testing.T) {
	cache.InitCache()
	opt := ensureCMOptions()
	opt.TenantConfig.EnableMultiTenantMode = true
	t.Cleanup(func() {
		opt.TenantConfig.EnableMultiTenantMode = false
	})

	auth := &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   "test-code",
			BkAppSecret: "test-secret",
		},
		BkUserName: "bcs-cluster-manager",
	}

	_, _, err := GetGatewayAuthAndTenantInfo(context.Background(), auth, "")
	if err == nil {
		t.Fatal("expected lookup error when multi-tenant enabled and bkUser client not init")
	}
	if !strings.Contains(err.Error(), "get bkUserName by tenant failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
