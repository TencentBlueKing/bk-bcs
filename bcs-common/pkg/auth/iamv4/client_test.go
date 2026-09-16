/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package iamv4

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cli, err := NewClient(&Options{
		SystemID:    "bk_bcs",
		AppCode:     "bk_bcs",
		AppSecret:   "test-secret",
		GateWayHost: srv.URL,
		TenantID:    "default",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return cli
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func TestNewClientValidate(t *testing.T) {
	_, err := NewClient(nil)
	if err == nil {
		t.Fatal("expected error for nil options")
	}
	_, err = NewClient(&Options{AppCode: "a", GateWayHost: "http://x"})
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
	cli, err := NewClient(&Options{AppCode: "a", AppSecret: "b", GateWayHost: "http://x/"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if cli.opt.GateWayHost != "http://x" {
		t.Fatalf("gateway host trim: %s", cli.opt.GateWayHost)
	}
	if cli.opt.TenantID != DefaultTenantID {
		t.Fatalf("default tenant: %s", cli.opt.TenantID)
	}
}

func TestRequireModelID(t *testing.T) {
	cases := []struct {
		id    string
		valid bool
	}{
		{"project", true},
		{"cloud_account", true},
		{"namespace_scoped_create", true},
		{"a", true},
		{"", false},
		{"Project", false},
		{"_project", false},
		{"project-", false},
		{strings.Repeat("a", 33), false},
		{strings.Repeat("a", 32), true},
	}
	for _, c := range cases {
		err := requireModelID("id", c.id)
		if c.valid && err != nil {
			t.Errorf("id %q should be valid: %v", c.id, err)
		}
		if !c.valid && err == nil {
			t.Errorf("id %q should be invalid", c.id)
		}
	}
}

func TestAPIErrorParse(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"error":      map[string]string{"code": "NOT_FOUND", "message": "system missing"},
			"request_id": "rid-1",
		})
	})
	_, err := cli.RetrieveSystem(context.Background(), "bk_bcs")
	if err == nil {
		t.Fatal("expected api error")
	}
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		t.Fatalf("want APIError, got %v", err)
	}
	if apiErr.Code != "NOT_FOUND" || apiErr.RequestID != "rid-1" || !isNotFound(err) {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
	if strings.Contains(err.Error(), "test-secret") {
		t.Fatal("error must not contain app secret")
	}
}

func TestAuthHeaderAndDirectAuth(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/open/rbac/authorization/systems/bk_bcs/auth/" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get(HeaderTenantID) != "default" {
			t.Errorf("tenant = %s", r.Header.Get(HeaderTenantID))
		}
		auth := r.Header.Get(HeaderBkAPIAuthorization)
		if !strings.Contains(auth, "bk_bcs") || !strings.Contains(auth, "test-secret") {
			t.Errorf("missing app identity header")
		}
		if strings.Contains(r.URL.RawQuery, "test-secret") || strings.Contains(r.URL.Path, "test-secret") {
			t.Error("secret must not appear in URL")
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data":       map[string]bool{"allowed": true},
			"request_id": "rid-auth",
		})
	})
	res, err := cli.DirectAuth(context.Background(), "", DirectAuthRequest{
		Subject:  Subject{Type: SubjectUser, ID: "alice"},
		ActionID: string(ProjectView),
		Resource: &AuthResource{ID: "p1"},
	})
	if err != nil {
		t.Fatalf("DirectAuth: %v", err)
	}
	if res == nil || !res.Allowed {
		t.Fatalf("allowed = %+v", res)
	}
}

func TestDirectAuthValidation(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call gateway")
	})
	_, err := cli.DirectAuth(context.Background(), "", DirectAuthRequest{
		Subject: Subject{Type: SubjectUser}, ActionID: "project_view",
	})
	if err == nil {
		t.Fatal("expected subject.id required")
	}
	_, err = cli.DirectAuthByActions(context.Background(), "", AuthByActionsRequest{
		Subject: Subject{Type: SubjectUser, ID: "u"}, ActionIDs: make([]string, 21),
	})
	if err == nil {
		t.Fatal("expected action_ids max 20")
	}
	_, err = cli.DirectAuthByResources(context.Background(), "", AuthByResourcesRequest{
		Subject: Subject{Type: SubjectUser, ID: "u"}, ActionID: "project_view",
	})
	if err == nil {
		t.Fatal("expected resources required")
	}
}

func TestOperatorRequired(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call gateway")
	})
	err := cli.AddAuthorization(context.Background(), "bk_bcs", "", []AuthorizationItem{{
		Subject: Subject{Type: SubjectUser, ID: "u"}, RoleID: "read_only",
	}})
	if err != ErrEmptyOperator {
		t.Fatalf("want ErrEmptyOperator, got %v", err)
	}
}

func TestShareRetrieveSystemPath(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/open/rabc/share/model/systems/bk_bcs/" {
			t.Fatalf("share retrieve path must keep gateway typo, got %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": System{ID: "bk_bcs", Name: "BCS"},
		})
	})
	sys, err := cli.ShareRetrieveSystem(context.Background(), "bk_bcs")
	if err != nil {
		t.Fatalf("ShareRetrieveSystem: %v", err)
	}
	if sys.ID != "bk_bcs" {
		t.Fatalf("system = %+v", sys)
	}
}

func TestBatchDeleteRoleActionQuery(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Query().Get("ids") != "project_view,cluster_view" {
			t.Fatalf("ids = %s", r.URL.Query().Get("ids"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := cli.BatchDeleteRoleAction(context.Background(), "bk_bcs", "read_only",
		[]string{"project_view", "cluster_view"})
	if err != nil {
		t.Fatalf("BatchDeleteRoleAction: %v", err)
	}
}

func TestGeneratePermApplyURL(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/open/application/permission-apply-urls/" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]string{"url": "https://iam.example/apply"},
		})
	})
	data, err := cli.GeneratePermApplyURL(context.Background(), GenerateApplyURLRequest{
		Permissions: []ApplyPermission{{ActionID: "project_view"}},
	})
	if err != nil {
		t.Fatalf("GeneratePermApplyURL: %v", err)
	}
	if data.URL != "https://iam.example/apply" {
		t.Fatalf("url = %s", data.URL)
	}
}

func TestCreateSpaceAndGroup(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/spaces/"):
			writeJSON(w, http.StatusCreated, map[string]interface{}{"data": map[string]int64{"id": 7}})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/groups/"):
			if r.Header.Get(HeaderIAMOperator) != "admin" {
				t.Errorf("missing operator")
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"data": map[string]int64{"id": 9}})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	space, err := cli.CreateSpace(context.Background(), "bk_bcs", CreateSpaceRequest{
		Name: "s", Description: "d", Managers: []string{"admin"},
		PermissionScope: []PermissionScopeItem{{ID: "read_only"}},
	})
	if err != nil || space.ID != 7 {
		t.Fatalf("CreateSpace: %+v %v", space, err)
	}
	group, err := cli.CreateGroup(context.Background(), "bk_bcs", 7, "admin", CreateGroupRequest{
		Name: "g", PermissionExpiredAt: 1,
	})
	if err != nil || group.ID != 9 {
		t.Fatalf("CreateGroup: %+v %v", group, err)
	}
}

func TestUnwrappedListBody(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"count":   1,
			"results": []ResourceType{{ID: "project", Name: "项目"}},
		})
	})
	page, err := cli.ListResourceType(context.Background(), "bk_bcs", PageQuery{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatalf("ListResourceType: %v", err)
	}
	if page.Count != 1 || len(page.Results) != 1 || page.Results[0].ID != "project" {
		t.Fatalf("page = %+v", page)
	}
}

func TestDirectAuthByResourcesArrayData(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": []map[string]interface{}{
				{"resource_id": "ns-1", "allowed": true},
				{"resource_id": "ns-2", "allowed": false},
			},
		})
	})
	got, err := cli.DirectAuthByResources(context.Background(), "bk_bcs", AuthByResourcesRequest{
		Subject:   Subject{Type: SubjectUser, ID: "alice"},
		ActionID:  "namespace_view",
		Resources: []AuthResource{{ID: "ns-1"}, {ID: "ns-2"}},
	})
	if err != nil {
		t.Fatalf("array data: %v", err)
	}
	if len(got) != 2 || got[0].ResourceID != "ns-1" || !got[0].Allowed || got[1].Allowed {
		t.Fatalf("got = %+v", got)
	}
}

func TestDirectAuthByResourcesMapData(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "auth-by-resources") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]bool{"ns-1": true, "ns-2": false},
		})
	})
	got, err := cli.DirectAuthByResources(context.Background(), "bk_bcs", AuthByResourcesRequest{
		Subject:  Subject{Type: SubjectUser, ID: "alice"},
		ActionID: "namespace_view",
		Resources: []AuthResource{
			{ID: "ns-1"},
			{ID: "ns-2"},
		},
	})
	if err != nil {
		t.Fatalf("DirectAuthByResources: %v", err)
	}
	byID := map[string]bool{}
	for _, r := range got {
		byID[r.ResourceID] = r.Allowed
	}
	if !byID["ns-1"] || byID["ns-2"] {
		t.Fatalf("got = %+v", got)
	}
}

func TestDirectAuthByActionsMapData(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]bool{"namespace_view": true, "namespace_update": false},
		})
	})
	got, err := cli.DirectAuthByActions(context.Background(), "bk_bcs", AuthByActionsRequest{
		Subject:   Subject{Type: SubjectUser, ID: "alice"},
		ActionIDs: []string{"namespace_view", "namespace_update"},
		Resource:  &AuthResource{ID: "ns-1"},
	})
	if err != nil {
		t.Fatalf("DirectAuthByActions: %v", err)
	}
	byID := map[string]bool{}
	for _, r := range got {
		byID[r.ActionID] = r.Allowed
	}
	if !byID["namespace_view"] || byID["namespace_update"] {
		t.Fatalf("got = %+v", got)
	}
}

func TestIAMPathHelpers(t *testing.T) {
	if ClusterIAMPath("p1") != "/project,p1/" {
		t.Fatalf("cluster path")
	}
	if NamespaceIAMPath("p1", "c1") != "/project,p1/cluster,c1/" {
		t.Fatalf("namespace path")
	}
	res := AuthResourceWithPath("ns1", NamespaceIAMPath("p1", "c1"))
	if res.Attributes[IAMPathAttr] != "/project,p1/cluster,c1/" {
		t.Fatalf("attr = %+v", res.Attributes)
	}
}

func TestDoDoesNotLogSecret(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "test-secret") {
			t.Error("secret must not be in request body")
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]bool{"allowed": false},
		})
	})
	_, err := cli.DirectAuth(context.Background(), "bk_bcs", DirectAuthRequest{
		Subject: Subject{Type: SubjectUser, ID: "u"}, ActionID: "project_view",
	})
	if err != nil {
		t.Fatalf("DirectAuth: %v", err)
	}
}
