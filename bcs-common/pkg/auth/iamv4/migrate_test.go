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
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

type migrateState struct {
	system        *System
	resourceTypes map[string]ResourceType
	actions       map[string]Action
	roles         map[string]Role
	createSystem  int
	updateSystem  int
	createRT      int
	updateRT      int
	createAction  int
	updateAction  int
	createRole    int
	updateRole    int
	addRoleAct    int
	delRoleAct    int
}

func newMigrateState() *migrateState {
	return &migrateState{
		resourceTypes: map[string]ResourceType{},
		actions:       map[string]Action{},
		roles:         map[string]Role{},
	}
}

func (s *migrateState) handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/auth-token/"):
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": map[string]string{"auth_token": "tok"}})
	case strings.Contains(path, "/resource-types/") && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": pageOf(values(s.resourceTypes))})
	case strings.Contains(path, "/resource-types/") && r.Method == http.MethodPost:
		s.createRT++
		var items []ResourceType
		_ = json.NewDecoder(r.Body).Decode(&items)
		for _, it := range items {
			s.resourceTypes[it.ID] = it
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": []string{"ok"}})
	case strings.Contains(path, "/resource-types/") && r.Method == http.MethodPut:
		s.updateRT++
		w.WriteHeader(http.StatusNoContent)
	case strings.Contains(path, "/actions/") && r.Method == http.MethodGet && !strings.Contains(path, "/roles/"):
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": pageOf(values(s.actions))})
	case strings.Contains(path, "/actions/") && r.Method == http.MethodPost && !strings.Contains(path, "/roles/"):
		s.createAction++
		var items []Action
		_ = json.NewDecoder(r.Body).Decode(&items)
		for _, it := range items {
			s.actions[it.ID] = it
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": []string{"ok"}})
	case strings.Contains(path, "/actions/") && r.Method == http.MethodPut && !strings.Contains(path, "/roles/"):
		s.updateAction++
		w.WriteHeader(http.StatusNoContent)
	case strings.Contains(path, "/roles/") && strings.Contains(path, "/actions/") && r.Method == http.MethodPost:
		s.addRoleAct++
		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": []string{"ok"}})
	case strings.Contains(path, "/roles/") && strings.Contains(path, "/actions/") && r.Method == http.MethodDelete:
		s.delRoleAct++
		w.WriteHeader(http.StatusNoContent)
	case strings.Contains(path, "/roles/") && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": pageOf(values(s.roles))})
	case strings.Contains(path, "/roles/") && r.Method == http.MethodPost:
		s.createRole++
		var items []Role
		_ = json.NewDecoder(r.Body).Decode(&items)
		for _, it := range items {
			s.roles[it.ID] = it
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": []string{"ok"}})
	case strings.Contains(path, "/roles/") && r.Method == http.MethodPut:
		s.updateRole++
		w.WriteHeader(http.StatusNoContent)
	case strings.Contains(path, "/model/systems/") && r.Method == http.MethodGet:
		if s.system == nil {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": map[string]string{"code": "NOT_FOUND", "message": "missing"},
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": s.system})
	case strings.HasSuffix(path, "/model/systems/") && r.Method == http.MethodPost:
		s.createSystem++
		var req CreateSystemRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		s.system = &System{ID: req.ID, Name: req.Name, Clients: req.Clients, CallbackURL: req.CallbackURL}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"data": map[string]string{"id": req.ID}})
	case strings.Contains(path, "/model/systems/") && r.Method == http.MethodPut:
		s.updateSystem++
		w.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(w, r)
	}
}

func pageOf[T any](items []T) PageResult[T] {
	return PageResult[T]{Count: len(items), Results: items}
}

func values[K comparable, V any](m map[K]V) []V {
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func TestMigrateCreateThenUpdate(t *testing.T) {
	st := newMigrateState()
	cli := testClient(t, st.handler)
	fsys := fstest.MapFS{
		"0000_init.up.json": {Data: []byte(`{
			"system_id": "{{ .BK_IAM_SYSTEM_ID }}",
			"enabled": true,
			"operations": [
				{"operation":"upsert_system","data":{
					"id":"{{ .BK_IAM_SYSTEM_ID }}",
					"name":"容器管理平台",
					"clients":["{{ .APP_CODE }}"],
					"callback_url":"{{ .BCS_HOST }}/callback/"
				}},
				{"operation":"upsert_resource_type","data":{"id":"project","name":"项目","ancestors":[]}},
				{"operation":"upsert_action","data":{"id":"project_view","name":"项目查看","resource_type_id":"project"}},
				{"operation":"upsert_role","data":{
					"id":"read_only","name":"业务只读","actions":[{"id":"project_view","resource_type_id":"project"}]
				}}
			]
		}`)},
		"readme.txt": {Data: []byte("ignore")},
	}
	vars := map[string]string{
		"BK_IAM_SYSTEM_ID": "bk_bcs",
		"APP_CODE":         "bk_bcs",
		"BCS_HOST":         "https://bcs.example.com",
	}
	if err := cli.Migrate(context.Background(), fsys, vars); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if st.createSystem != 1 || st.createRT != 1 || st.createAction != 1 || st.createRole != 1 {
		t.Fatalf("create counts sys=%d rt=%d act=%d role=%d",
			st.createSystem, st.createRT, st.createAction, st.createRole)
	}
	if st.system == nil || st.system.CallbackURL != "https://bcs.example.com/callback/" {
		t.Fatalf("system = %+v", st.system)
	}

	if err := cli.Migrate(context.Background(), fsys, vars); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if st.createSystem != 1 || st.updateSystem != 1 {
		t.Fatalf("upsert system create=%d update=%d", st.createSystem, st.updateSystem)
	}
	if st.createRT != 1 || st.updateRT != 1 {
		t.Fatalf("upsert rt create=%d update=%d", st.createRT, st.updateRT)
	}
	if st.createRole != 1 || st.updateRole != 1 {
		t.Fatalf("upsert role create=%d update=%d", st.createRole, st.updateRole)
	}
}

func TestMigrateSkipDisabledAndUnknownOp(t *testing.T) {
	st := newMigrateState()
	cli := testClient(t, st.handler)
	disabled := fstest.MapFS{
		"0000_skip.up.json": {Data: []byte(`{"system_id":"bk_bcs","enabled":false,"operations":[
			{"operation":"upsert_system","data":{"id":"bk_bcs","name":"x","clients":["a"]}}
		]}`)},
	}
	if err := cli.Migrate(context.Background(), disabled, nil); err != nil {
		t.Fatalf("disabled migrate: %v", err)
	}
	if st.createSystem != 0 {
		t.Fatal("disabled file must not apply")
	}

	unknown := fstest.MapFS{
		"0001_bad.up.json": {Data: []byte(`{"system_id":"bk_bcs","enabled":true,"operations":[
			{"operation":"upsert_instance_selection","data":{}}
		]}`)},
	}
	err := cli.Migrate(context.Background(), unknown, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported operation") {
		t.Fatalf("want unsupported operation, got %v", err)
	}
}

func TestMigrateMissingTemplateKey(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not call gateway")
	})
	fsys := fstest.MapFS{
		"0000_init.up.json": {Data: []byte(`{"system_id":"{{ .BK_IAM_SYSTEM_ID }}","enabled":true,"operations":[]}`)},
	}
	err := cli.Migrate(context.Background(), fsys, map[string]string{"APP_CODE": "x"})
	if err == nil {
		t.Fatal("expected missing template key")
	}
}

func TestMigrateRoleActionSync(t *testing.T) {
	st := newMigrateState()
	st.roles["read_only"] = Role{
		ID: "read_only", Name: "旧",
		Actions: []RoleAction{{ID: "project_view", ResourceTypeID: "project"}, {ID: "obsolete"}},
	}
	cli := testClient(t, st.handler)
	fsys := fstest.MapFS{
		"0003_roles.up.json": {Data: []byte(`{
			"system_id":"bk_bcs","enabled":true,"operations":[
				{"operation":"upsert_role","data":{
					"id":"read_only","name":"业务只读",
					"actions":[{"id":"project_view","resource_type_id":"project"},{"id":"cluster_view","resource_type_id":"project"}]
				}}
			]
		}`)},
	}
	if err := cli.Migrate(context.Background(), fsys, nil); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if st.updateRole != 1 || st.addRoleAct != 1 || st.delRoleAct != 1 {
		t.Fatalf("sync update=%d add=%d del=%d", st.updateRole, st.addRoleAct, st.delRoleAct)
	}
}

func TestListMigrateFilesOrder(t *testing.T) {
	fsys := fstest.MapFS{
		"0002_b.up.json": {},
		"0001_a.up.json": {},
		"0010_c.up.json": {},
		"note.md":        {},
	}
	files, err := listMigrateFiles(fsys)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(files))
	for _, f := range files {
		got = append(got, f.name)
	}
	want := []string{"0001_a.up.json", "0002_b.up.json", "0010_c.up.json"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("order = %v", got)
	}
}

func TestListAllPagination(t *testing.T) {
	page := 0
	items, err := listAll(func(q PageQuery) (*PageResult[int], error) {
		page++
		if q.PageSize != listPageSize {
			t.Fatalf("page size = %d", q.PageSize)
		}
		if page == 1 {
			out := make([]int, listPageSize)
			return &PageResult[int]{Count: listPageSize + 1, Results: out}, nil
		}
		return &PageResult[int]{Count: listPageSize + 1, Results: []int{1}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != listPageSize+1 || page != 2 {
		t.Fatalf("items=%d pages=%d", len(items), page)
	}
}

func TestMigrateNilFS(t *testing.T) {
	cli := testClient(t, func(w http.ResponseWriter, r *http.Request) {})
	var fsys fs.FS
	if err := cli.Migrate(context.Background(), fsys, nil); err == nil {
		t.Fatal("expected fs required")
	}
}

func TestRealMigrationsV4Render(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "bcs-services", "bcs-user-manager", "migrations-v4")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("migrations-v4 not in this checkout")
	}
	fsys := os.DirFS(dir)
	files, err := listMigrateFiles(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 4 {
		t.Fatalf("expected >=4 migrate files, got %d", len(files))
	}
	vars := map[string]string{
		"BK_IAM_SYSTEM_ID": "bk_bcs",
		"APP_CODE":         "bk_bcs",
		"BCS_HOST":         "https://bcs.example.com",
	}
	ops := map[string]int{}
	for _, file := range files {
		body, err := renderMigrateFile(fsys, file.name, vars)
		if err != nil {
			t.Fatalf("render %s: %v", file.name, err)
		}
		var mf MigrationFile
		if err := json.Unmarshal(body, &mf); err != nil {
			t.Fatalf("decode %s: %v", file.name, err)
		}
		if !mf.Enabled || mf.SystemID != "bk_bcs" {
			t.Fatalf("%s enabled/system invalid", file.name)
		}
		for i, op := range mf.Operations {
			ops[op.Operation]++
			switch op.Operation {
			case OpUpsertSystem:
				var data CreateSystemRequest
				if err := json.Unmarshal(op.Data, &data); err != nil {
					t.Fatalf("%s[%d]: %v", file.name, i, err)
				}
				if err := requireModelID("id", data.ID); err != nil || len(data.Clients) == 0 {
					t.Fatalf("%s[%d] system invalid: %+v %v", file.name, i, data, err)
				}
			case OpUpsertResourceType:
				var data ResourceType
				if err := json.Unmarshal(op.Data, &data); err != nil {
					t.Fatalf("%s[%d]: %v", file.name, i, err)
				}
				if err := requireModelID("id", data.ID); err != nil {
					t.Fatalf("%s[%d] rt: %v", file.name, i, err)
				}
			case OpUpsertAction:
				var data Action
				if err := json.Unmarshal(op.Data, &data); err != nil {
					t.Fatalf("%s[%d]: %v", file.name, i, err)
				}
				if err := requireModelID("id", data.ID); err != nil {
					t.Fatalf("%s[%d] action: %v", file.name, i, err)
				}
			case OpUpsertRole:
				var data Role
				if err := json.Unmarshal(op.Data, &data); err != nil {
					t.Fatalf("%s[%d]: %v", file.name, i, err)
				}
				if err := requireModelID("id", data.ID); err != nil || len(data.Actions) == 0 {
					t.Fatalf("%s[%d] role invalid: %+v %v", file.name, i, data, err)
				}
			default:
				t.Fatalf("%s[%d] unsupported %s", file.name, i, op.Operation)
			}
		}
	}
	if ops[OpUpsertSystem] < 1 || ops[OpUpsertResourceType] < 5 ||
		ops[OpUpsertAction] < 30 || ops[OpUpsertRole] < 10 {
		t.Fatalf("operation counts = %+v", ops)
	}

	st := newMigrateState()
	cli := testClient(t, st.handler)
	if err := cli.Migrate(context.Background(), fsys, vars); err != nil {
		t.Fatalf("apply real migrations-v4: %v", err)
	}
	if st.system == nil || len(st.resourceTypes) < 5 || len(st.actions) < 30 || len(st.roles) < 10 {
		t.Fatalf("applied model incomplete: sys=%v rt=%d act=%d role=%d",
			st.system, len(st.resourceTypes), len(st.actions), len(st.roles))
	}
}
