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

package iam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iamv4 "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iamv4"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
)

func TestGrantManagerRejectsEmptyUser(t *testing.T) {
	err := GrantProjectManager(context.Background(), "default", "", "pid", "pname")
	if err == nil {
		t.Fatal("empty username should be rejected")
	}
}

func TestGrantProjectManagerV4(t *testing.T) {
	var got iamv4.AuthorizationItem
	var operator, tenant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		wantPath := "/api/v1/open/rbac/mgmt/systems/" + bcsIAM.SystemIDBKBCS + "/authorizations/"
		if r.URL.Path != wantPath {
			t.Errorf("path = %s", r.URL.Path)
		}
		operator = r.Header.Get(iamv4.HeaderIAMOperator)
		tenant = r.Header.Get(iamv4.HeaderTenantID)
		var items []iamv4.AuthorizationItem
		if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
			t.Errorf("decode body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(items) != 1 {
			t.Errorf("items = %d", len(items))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		got = items[0]
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	prev := config.GlobalConf
	config.GlobalConf = &config.ProjectConfig{
		App: config.AppConfig{Code: "bk-bcs", Secret: "test-secret"},
		IAM: config.IAMConfig{EnableV4: true, V4GatewayHost: srv.URL},
	}
	t.Cleanup(func() { config.GlobalConf = prev })

	err := GrantProjectManager(context.Background(), "tenant-a", "alice", "proj-1", "demo")
	if err != nil {
		t.Fatalf("GrantProjectManager: %v", err)
	}
	if operator != "alice" || tenant != "tenant-a" {
		t.Fatalf("operator=%s tenant=%s", operator, tenant)
	}
	if got.RoleID != iamv4.RoleOperator.String() || got.Subject.ID != "alice" ||
		got.Subject.Type != iamv4.SubjectUser {
		t.Fatalf("authorization = %+v", got)
	}
	if got.RelatedResourceTypeID != iamv4.SysProject.String() || len(got.Resources) != 1 ||
		got.Resources[0].ID != "proj-1" || got.Resources[0].Type != iamv4.SysProject.String() {
		t.Fatalf("resource = %+v", got)
	}
	maxExpire := time.Now().Add(365 * 24 * time.Hour).Unix()
	minExpire := time.Now().Add(364 * 24 * time.Hour).Unix()
	if got.ExpiredAt > maxExpire || got.ExpiredAt < minExpire {
		t.Fatalf("expired_at = %d, want within 365 days", got.ExpiredAt)
	}
}
