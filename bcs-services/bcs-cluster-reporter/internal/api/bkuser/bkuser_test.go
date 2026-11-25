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

package bkuser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListTenant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != listTenantPath {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("X-Bk-Tenant-Id"); got != systemTenantID {
			t.Errorf("unexpected tenant header %q", got)
		}
		auth := authInfo{}
		if err := json.Unmarshal([]byte(r.Header.Get("X-Bkapi-Authorization")), &auth); err != nil ||
			auth.BkAppCode != "code" || auth.BkAppSecret != "secret" {
			t.Errorf("unexpected auth header, err: %v", err)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"default","name":"Default"},{"id":"t-1"},{"id":"bad\r\nid"},{"id":""}]}`))
	}))
	defer server.Close()

	cli, err := NewClient(server.URL+"/", "code", "secret")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	tenants, err := cli.ListTenant(context.Background())
	if err != nil {
		t.Fatalf("ListTenant failed: %v", err)
	}
	ids := make([]string, 0, len(tenants))
	for _, tenant := range tenants {
		ids = append(ids, tenant.ID)
	}
	if strings.Join(ids, ",") != "default,t-1" {
		t.Fatalf("unexpected tenants %v", ids)
	}
}

func TestListTenantStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	cli, err := NewClient(server.URL, "code", "secret")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	_, err = cli.ListTenant(context.Background())
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("expect status error without secret, got %v", err)
	}
}
