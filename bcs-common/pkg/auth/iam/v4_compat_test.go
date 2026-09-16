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

package iam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"
)

type staticPath string

func (s staticPath) BuildIAMPath() string { return string(s) }

func newTestV4Client(t *testing.T, handler http.HandlerFunc) PermClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cli, err := NewIamClient(&Options{
		SystemID:      "bk_bcs",
		AppCode:       "bk_bcs",
		AppSecret:     "test-secret",
		V4GateWayHost: srv.URL,
		Version:       VersionV4,
	})
	if err != nil {
		t.Fatalf("NewIamClient v4: %v", err)
	}
	return cli
}

func writeAuthJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func TestNewIamClientV4UsesPermAuthClient(t *testing.T) {
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]bool{"allowed": true},
		})
	})
	allow, err := cli.IsAllowedWithoutResource("project_view", PermissionRequest{
		SystemID: "bk_bcs", UserName: "alice",
	}, false)
	if err != nil || !allow {
		t.Fatalf("IsAllowedWithoutResource: %v %v", allow, err)
	}
}

func TestV4IsAllowedWithResourcePath(t *testing.T) {
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/auth/") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		res := req["resource"].(map[string]interface{})
		if res["id"] != "c1" {
			t.Fatalf("resource.id = %v", res["id"])
		}
		attrs := res["attributes"].(map[string]interface{})
		if attrs["_bk_iam_path_"] != "/project,p1/" {
			t.Fatalf("path = %v", attrs["_bk_iam_path_"])
		}
		if strings.Contains(r.URL.RawQuery, "test-secret") {
			t.Fatal("secret must not appear in URL")
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]bool{"allowed": false},
		})
	})
	allow, err := cli.IsAllowedWithResource("cluster_view", PermissionRequest{
		SystemID: "bk_bcs", UserName: "alice",
	}, []ResourceNode{{
		System: "bk_bcs", RType: "cluster", RInstance: "c1", Rp: staticPath("/project,p1/"),
	}}, false)
	if err != nil || allow {
		t.Fatalf("IsAllowedWithResource: %v %v", allow, err)
	}
}

func TestV4BatchResourceIsAllowedKeys(t *testing.T) {
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "auth-by-resources") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{
			"data": []map[string]interface{}{
				{"resource_id": "c1", "allowed": true},
				{"resource_id": "c2", "allowed": false},
			},
		})
	})
	got, err := cli.BatchResourceIsAllowed("cluster_view", PermissionRequest{
		SystemID: "bk_bcs", UserName: "alice",
	}, [][]ResourceNode{
		{{RType: "cluster", RInstance: "c1", Rp: staticPath("")}},
		{{RType: "cluster", RInstance: "c2", Rp: staticPath("")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !got["c1"] || got["c2"] {
		t.Fatalf("got = %+v", got)
	}
}

func TestV4MultiActionsGroupsByResourceType(t *testing.T) {
	var batches [][]string
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "auth-by-actions") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		raw, _ := req["action_ids"].([]interface{})
		ids := make([]string, 0, len(raw))
		for _, v := range raw {
			ids = append(ids, v.(string))
		}
		batches = append(batches, ids)
		res := req["resource"].(map[string]interface{})
		allowed := make([]map[string]interface{}, 0, len(ids))
		for _, id := range ids {
			ok := (id == "cluster_view" && res["id"] == "c1") ||
				(id == "project_view" && res["id"] == "p1")
			allowed = append(allowed, map[string]interface{}{"action_id": id, "allowed": ok})
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{"data": allowed})
	})
	got, err := cli.ResourceMultiActionsAllowed([]string{"project_view", "cluster_view"}, PermissionRequest{
		SystemID: "bk_bcs", UserName: "alice",
	}, []ResourceNode{{RType: "cluster", RInstance: "c1", Rp: staticPath("/project,p1/")}})
	if err != nil {
		t.Fatal(err)
	}
	if !got["project_view"] || !got["cluster_view"] {
		t.Fatalf("got = %+v batches=%v", got, batches)
	}
	if len(batches) != 2 {
		t.Fatalf("want 2 same-type batches, got %v", batches)
	}
	for _, b := range batches {
		if len(b) != 1 {
			t.Fatalf("mixed batch %v", b)
		}
	}
}

func TestV4ListNamespaceActionsDoNotMixTypes(t *testing.T) {
	var batches [][]string
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		raw, _ := req["action_ids"].([]interface{})
		ids := make([]string, 0, len(raw))
		for _, v := range raw {
			ids = append(ids, v.(string))
		}
		batches = append(batches, ids)
		allowed := make([]map[string]interface{}, 0, len(ids))
		for _, id := range ids {
			allowed = append(allowed, map[string]interface{}{"action_id": id, "allowed": true})
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{"data": allowed})
	})
	got, err := cli.ResourceMultiActionsAllowed(
		[]string{"project_view", "namespace_list", "cluster_view"},
		PermissionRequest{SystemID: "bk_bcs", UserName: "alice"},
		[]ResourceNode{{RType: "project", RInstance: "p1"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !got["project_view"] {
		t.Fatalf("project_view = %+v", got)
	}
	if got["namespace_list"] || got["cluster_view"] {
		t.Fatalf("cluster actions must stay false on project node: %+v", got)
	}
	if len(batches) != 1 || len(batches[0]) != 1 || batches[0][0] != "project_view" {
		t.Fatalf("project node must only auth project_view, batches=%v", batches)
	}
}

func TestV4MultiActionsFallbackSameType(t *testing.T) {
	calls := 0
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if strings.Contains(r.URL.Path, "auth-by-actions") {
			writeAuthJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error": map[string]string{"code": "INVALID_ARGUMENT", "message": "temporary"},
			})
			return
		}
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]bool{"allowed": req["action_id"] == "namespace_view"},
		})
	})
	got, err := cli.ResourceMultiActionsAllowed([]string{"namespace_view", "namespace_update"}, PermissionRequest{
		SystemID: "bk_bcs", UserName: "alice",
	}, []ResourceNode{{RType: "namespace", RInstance: "ns1", Rp: staticPath("/project,p1/cluster,c1/")}})
	if err != nil {
		t.Fatal(err)
	}
	if !got["namespace_view"] || got["namespace_update"] {
		t.Fatalf("got = %+v calls=%d", got, calls)
	}
	if calls < 3 {
		t.Fatalf("expected batch + 2 fallback, calls=%d", calls)
	}
}

func TestV4GetTokenAndBasicAuth(t *testing.T) {
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "auth-token") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]string{"auth_token": "tok-1"},
		})
	})
	token, err := cli.GetToken()
	if err != nil || token != "tok-1" {
		t.Fatalf("GetToken: %s %v", token, err)
	}
	if err := cli.IsBasicAuthAllowed(BkUser{BkUserName: "alice", BkToken: "tok-1"}); err != nil {
		t.Fatalf("basic auth: %v", err)
	}
	err = cli.IsBasicAuthAllowed(BkUser{BkUserName: "alice", BkToken: "bad"})
	if err == nil || strings.Contains(err.Error(), "tok-1") || strings.Contains(err.Error(), "bad") {
		t.Fatalf("failed auth must not echo token: %v", err)
	}
}

func TestV4GetApplyURL(t *testing.T) {
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/open/application/permission-apply-urls/" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		perms := req["permissions"].([]interface{})
		if len(perms) != 1 {
			t.Fatalf("permissions = %v", perms)
		}
		p0 := perms[0].(map[string]interface{})
		if p0["action_id"] != "cluster_view" {
			t.Fatalf("action = %v", p0["action_id"])
		}
		res := p0["resources"].([]interface{})[0].(map[string]interface{})
		if res["id"] != "c1" || res["type"] != "cluster" {
			t.Fatalf("resource = %v", res)
		}
		anc := res["ancestors"].([]interface{})[0].(map[string]interface{})
		if anc["id"] != "p1" || anc["type"] != "project" {
			t.Fatalf("ancestor = %v", anc)
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]string{"url": "https://iam.example/apply"},
		})
	})
	url, err := cli.GetApplyURL(ApplicationRequest{SystemID: "bk_bcs"}, []ApplicationAction{{
		ActionID: "cluster_view",
		RelatedResources: []bkiam.ApplicationRelatedResourceType{{
			SystemID: "bk_bcs",
			Type:     "cluster",
			Instances: []bkiam.ApplicationResourceInstance{{
				{Type: "project", ID: "p1"},
				{Type: "cluster", ID: "c1"},
			}},
		}},
	}}, BkUser{BkUserName: "alice"})
	if err != nil || url != "https://iam.example/apply" {
		t.Fatalf("GetApplyURL: %s %v", url, err)
	}
}

func TestV4AuthValidationAndUnsupported(t *testing.T) {
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call gateway")
	})
	if _, err := cli.IsAllowedWithoutResource("", PermissionRequest{SystemID: "bk_bcs", UserName: "u"}, false); err == nil {
		t.Fatal("expected action required")
	}
	if _, err := cli.IsAllowedWithoutResource("project_view", PermissionRequest{}, false); err == nil {
		t.Fatal("expected request required")
	}
	if _, err := cli.CreateGradeManagers(nil, GradeManagerRequest{}); err != ErrIAMV4Unsupported {
		t.Fatalf("want unsupported, got %v", err)
	}
}

func TestV4NewIamMigrateClientRejected(t *testing.T) {
	_, err := NewIamMigrateClient(&Options{
		SystemID: "bk_bcs", AppCode: "a", AppSecret: "b", V4GateWayHost: "http://x", Version: VersionV4,
	})
	if err == nil {
		t.Fatal("expected migrate client rejected")
	}
}

func TestV4GateWayHostRequired(t *testing.T) {
	_, err := NewIamClient(&Options{
		SystemID: "bk_bcs", AppCode: "a", AppSecret: "b", GateWayHost: "http://v3", Version: VersionV4,
	})
	if err == nil || !strings.Contains(err.Error(), "V4GateWayHost") {
		t.Fatalf("want V4GateWayHost required, got %v", err)
	}
}

func TestApplyV4Config(t *testing.T) {
	opt := &Options{GateWayHost: "http://v3"}
	ApplyV4Config(opt, false, " http://v4 ")
	if opt.Version != "" || opt.V4GateWayHost != "http://v4" {
		t.Fatalf("disabled: %+v", opt)
	}
	ApplyV4Config(opt, true, "http://v4-gw")
	if opt.Version != VersionV4 || opt.V4GateWayHost != "http://v4-gw" {
		t.Fatalf("enabled: %+v", opt)
	}
}

func TestResourceIDFromNodes(t *testing.T) {
	if resourceIDFromNodes([]ResourceNode{{RInstance: "c1"}}) != "c1" {
		t.Fatal("single")
	}
	got := resourceIDFromNodes([]ResourceNode{
		{RType: "project", RInstance: "p1"},
		{RType: "cluster", RInstance: "c1"},
	})
	if got != "project:p1/cluster:c1" {
		t.Fatalf("multi = %s", got)
	}
}

func TestBuildIAMPathFromAncestors(t *testing.T) {
	p := buildIAMPathFromAncestors([]ResourceNode{
		{RType: "project", RInstance: "p1"},
		{RType: "cluster", RInstance: "c1"},
	})
	if p != "/project,p1/cluster,c1/" {
		t.Fatalf("path = %s", p)
	}
}

func TestParseIAMPath(t *testing.T) {
	got := parseIAMPath("/project,p1/cluster,c1/")
	if len(got) != 2 || got[0].RType != "project" || got[0].RInstance != "p1" ||
		got[1].RType != "cluster" || got[1].RInstance != "c1" {
		t.Fatalf("got = %+v", got)
	}
	if len(parseIAMPath("")) != 0 {
		t.Fatal("empty")
	}
}

func TestV4BatchResourceMultiActionsCanListNamespace(t *testing.T) {
	var actions []string
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "auth-by-resources") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		action := req["action_id"].(string)
		actions = append(actions, action)
		raw, _ := req["resources"].([]interface{})
		allowed := make([]map[string]interface{}, 0, len(raw))
		for _, v := range raw {
			id := v.(map[string]interface{})["id"].(string)
			ok := (action == "project_view" && id == "p1") ||
				((action == "namespace_list" || action == "cluster_view") && id == "c1")
			allowed = append(allowed, map[string]interface{}{"resource_id": id, "allowed": ok})
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{"data": allowed})
	})
	got, err := cli.BatchResourceMultiActionsAllowed(
		[]string{"project_view", "namespace_list", "cluster_view"},
		PermissionRequest{SystemID: "bk_bcs", UserName: "alice"},
		[][]ResourceNode{
			{{RType: "project", RInstance: "p1"}},
			{{RType: "cluster", RInstance: "c1", Rp: staticPath("/project,p1/")}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !got["p1"]["project_view"] {
		t.Fatalf("project node = %+v", got["p1"])
	}
	if !got["c1"]["namespace_list"] || !got["c1"]["cluster_view"] {
		t.Fatalf("cluster node = %+v", got["c1"])
	}
	if len(actions) != 3 {
		t.Fatalf("want 3 by-action calls, got %v", actions)
	}
}

func TestV4BatchMultiActionsSplitsByActionAndDedups(t *testing.T) {
	var mu sync.Mutex
	type call struct {
		action string
		ids    []string
	}
	var calls []call
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		action := req["action_id"].(string)
		raw, _ := req["resources"].([]interface{})
		ids := make([]string, 0, len(raw))
		allowed := make([]map[string]interface{}, 0, len(raw))
		for _, v := range raw {
			id := v.(map[string]interface{})["id"].(string)
			ids = append(ids, id)
			allowed = append(allowed, map[string]interface{}{"resource_id": id, "allowed": action == "namespace_view"})
		}
		mu.Lock()
		calls = append(calls, call{action: action, ids: ids})
		mu.Unlock()
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{"data": allowed})
	})
	nodes := make([][]ResourceNode, 0, 3)
	for _, ns := range []string{"n1", "n2", "n3"} {
		nodes = append(nodes, []ResourceNode{{
			RType: "namespace", RInstance: ns, Rp: staticPath("/project,p1/cluster,c1/"),
		}})
	}
	got, err := cli.BatchResourceMultiActionsAllowed(
		[]string{"namespace_view", "namespace_create"},
		PermissionRequest{SystemID: "bk_bcs", UserName: "alice"},
		nodes,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !got["n1"]["namespace_view"] || got["n1"]["namespace_create"] {
		t.Fatalf("n1 = %+v", got["n1"])
	}
	var createIDs, viewIDs []string
	for _, c := range calls {
		switch c.action {
		case "namespace_create":
			createIDs = append(createIDs, c.ids...)
		case "namespace_view":
			viewIDs = append(viewIDs, c.ids...)
		default:
			t.Fatalf("unexpected action %s", c.action)
		}
	}
	if len(createIDs) != 1 || createIDs[0] != "c1" {
		t.Fatalf("namespace_create should dedup to cluster, got %v", createIDs)
	}
	if len(viewIDs) != 3 {
		t.Fatalf("namespace_view resources = %v", viewIDs)
	}
}

func TestV4BatchMultiActionsConcurrentWhenOver20(t *testing.T) {
	var arrived int32
	unblock := make(chan struct{})
	cli := newTestV4Client(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&arrived, 1)
		if n == 1 {
			select {
			case <-unblock:
			case <-time.After(2 * time.Second):
				t.Error("expected a concurrent second chunk")
			}
		} else {
			select {
			case <-unblock:
			default:
				close(unblock)
			}
		}
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		raw, _ := req["resources"].([]interface{})
		allowed := make([]map[string]interface{}, 0, len(raw))
		for _, v := range raw {
			id := v.(map[string]interface{})["id"].(string)
			allowed = append(allowed, map[string]interface{}{"resource_id": id, "allowed": true})
		}
		writeAuthJSON(w, http.StatusOK, map[string]interface{}{"data": allowed})
	})
	nodes := make([][]ResourceNode, 0, 25)
	for i := 0; i < 25; i++ {
		nodes = append(nodes, []ResourceNode{{
			RType: "namespace", RInstance: fmt.Sprintf("ns-%d", i), Rp: staticPath("/project,p1/cluster,c1/"),
		}})
	}
	got, err := cli.BatchResourceMultiActionsAllowed(
		[]string{"namespace_view"},
		PermissionRequest{SystemID: "bk_bcs", UserName: "alice"},
		nodes,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 25 || !got["ns-0"]["namespace_view"] || !got["ns-24"]["namespace_view"] {
		t.Fatalf("got = %d entries %+v", len(got), got["ns-0"])
	}
	if atomic.LoadInt32(&arrived) != 2 {
		t.Fatalf("want 2 chunks, arrived=%d", arrived)
	}
}

func TestGroupActionsByResource(t *testing.T) {
	groups := groupActionsByResource(
		[]string{"project_view", "namespace_list", "cluster_view", "namespace_create"},
		[]ResourceNode{{RType: "cluster", RInstance: "c1", Rp: staticPath("/project,p1/")}},
	)
	if len(groups) != 2 {
		t.Fatalf("groups = %+v", groups)
	}
	byID := map[string][]string{}
	for _, g := range groups {
		if g.resource == nil {
			t.Fatal("resource required")
		}
		byID[g.resource.ID] = g.actions
	}
	if len(byID["c1"]) != 3 {
		t.Fatalf("cluster group = %v", byID["c1"])
	}
	if len(byID["p1"]) != 1 || byID["p1"][0] != "project_view" {
		t.Fatalf("project group = %v", byID["p1"])
	}

	nsGroups := groupActionsByResource(
		[]string{"namespace_view", "namespace_create"},
		[]ResourceNode{{RType: "namespace", RInstance: "40000:abcdde", Rp: staticPath("/project,p1/cluster,c1/")}},
	)
	if len(nsGroups) != 2 {
		t.Fatalf("ns groups = %+v", nsGroups)
	}
}
