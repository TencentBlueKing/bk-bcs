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

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iamv4 "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iamv4"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
)

func TestV4UnrestrictedProjectView(t *testing.T) {
	var authCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/open/rbac/mgmt/systems/"+bcsIAM.SystemIDBKBCS+"/authorizations/query-subject/" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get(iamv4.HeaderTenantID) != "tenant-a" {
			t.Errorf("tenant = %s", r.Header.Get(iamv4.HeaderTenantID))
		}
		var req iamv4.ListAuthorizationSubjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Resource == nil || req.Resource.ID != unrestrictedResourceID || req.Resource.Type != "project" {
			t.Errorf("resource = %+v", req.Resource)
		}
		results := []iamv4.AuthorizationSubject{}
		if req.RoleID == iamv4.RoleOperator.String() {
			results = append(results, iamv4.AuthorizationSubject{
				Subject: iamv4.Subject{Type: iamv4.SubjectUser, ID: "alice"},
			})
		}
		writeIAMData(w, map[string]interface{}{"count": len(results), "results": results})
		authCalls++
	}))
	defer srv.Close()
	useIAMV4(t, srv.URL)
	restoreLister := stubProjectLister(t, func(context.Context, string) ([]string, error) {
		t.Fatal("unrestricted grant should not list local projects")
		return nil, nil
	})
	defer restoreLister()

	ids, any, err := ListAuthorizedProjectIDs(context.Background(), "alice", "tenant-a")
	if err != nil {
		t.Fatalf("ListAuthorizedProjectIDs: %v", err)
	}
	if !any || len(ids) != 0 {
		t.Fatalf("ids=%v any=%v", ids, any)
	}
	if authCalls == 0 {
		t.Fatal("expected query-subject calls")
	}
}

func TestV4FilterProjectView(t *testing.T) {
	var batches [][]string
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/open/rbac/mgmt/systems/" + bcsIAM.SystemIDBKBCS + "/authorizations/query-subject/":
			writeIAMData(w, map[string]interface{}{"count": 0, "results": []iamv4.AuthorizationSubject{}})
		case "/api/v1/open/rbac/authorization/systems/" + bcsIAM.SystemIDBKBCS + "/auth-by-resources/":
			var req iamv4.AuthByResourcesRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if req.ActionID != ProjectView || req.Subject.ID != "alice" {
				t.Errorf("auth request = %+v", req)
			}
			ids := make([]string, 0, len(req.Resources))
			results := make([]map[string]interface{}, 0, len(req.Resources))
			for _, res := range req.Resources {
				ids = append(ids, res.ID)
				results = append(results, map[string]interface{}{
					"resource_id": res.ID,
					"allowed":     res.ID == "p1" || res.ID == "p3",
				})
			}
			mu.Lock()
			batches = append(batches, ids)
			mu.Unlock()
			writeIAMData(w, results)
		default:
			t.Errorf("path = %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	useIAMV4(t, srv.URL)
	restoreLister := stubProjectLister(t, func(_ context.Context, tenantID string) ([]string, error) {
		if tenantID != "tenant-a" {
			t.Errorf("tenant = %s", tenantID)
		}
		ids := []string{"p1", "", "p2", "p1"}
		for i := 0; i < 18; i++ {
			ids = append(ids, fmt.Sprintf("n%d", i))
		}
		ids = append(ids, "p3")
		return ids, nil
	})
	defer restoreLister()

	ids, any, err := ListAuthorizedProjectIDs(context.Background(), "alice", "tenant-a")
	if err != nil {
		t.Fatalf("ListAuthorizedProjectIDs: %v", err)
	}
	if any {
		t.Fatal("instance grant should not be unrestricted")
	}
	if len(ids) != 2 || ids[0] != "p1" || ids[1] != "p3" {
		t.Fatalf("ids = %v", ids)
	}
	if len(batches) != 2 {
		t.Fatalf("batches = %d, want 2", len(batches))
	}
}

func TestV4MissingGatewayHost(t *testing.T) {
	prev := config.GlobalConf
	config.GlobalConf = &config.ProjectConfig{
		App: config.AppConfig{Code: "bk-bcs", Secret: "test-secret"},
		IAM: config.IAMConfig{EnableV4: true},
	}
	t.Cleanup(func() { config.GlobalConf = prev })

	_, _, err := ListAuthorizedProjectIDs(context.Background(), "alice", "tenant-a")
	if err == nil {
		t.Fatal("missing v4 gateway host should fail")
	}
}

func useIAMV4(t *testing.T, host string) {
	t.Helper()
	prev := config.GlobalConf
	config.GlobalConf = &config.ProjectConfig{
		App: config.AppConfig{Code: "bk-bcs", Secret: "test-secret"},
		IAM: config.IAMConfig{EnableV4: true, V4GatewayHost: host},
	}
	t.Cleanup(func() { config.GlobalConf = prev })
}

func stubProjectLister(t *testing.T, fn func(context.Context, string) ([]string, error)) func() {
	t.Helper()
	prev := projectIDLister
	projectIDLister = fn
	return func() { projectIDLister = prev }
}

func writeIAMData(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":       data,
		"request_id": "rid-test",
	})
}
