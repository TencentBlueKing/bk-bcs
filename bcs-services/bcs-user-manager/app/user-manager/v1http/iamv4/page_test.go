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
	"reflect"
	"testing"
)

func TestPaginate(t *testing.T) {
	items := []Instance{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	total, got := Paginate(items, 2, 2)
	if total != 3 || len(got) != 1 || got[0].ID != "c" {
		t.Fatalf("page2 got total=%d items=%v", total, got)
	}
	total, got = Paginate(items, 5, 2)
	if total != 3 || len(got) != 0 {
		t.Fatalf("out of range got total=%d items=%v", total, got)
	}
}

func TestMatchKeyword(t *testing.T) {
	if !MatchKeyword("BCS-K8S-1", "prod(BCS-K8S-1)", "k8s") {
		t.Fatal("should match id")
	}
	if MatchKeyword("a", "b", "zzz") {
		t.Fatal("should not match")
	}
}

func TestApplyRequires(t *testing.T) {
	ins := Instance{ID: "p1", DisplayName: "n", IAMPath: "/project,p1/", Approvers: []string{"u"}}
	full := applyRequires(ins, nil)
	if full[attrID] != "p1" || full[attrDisplayName] != "n" {
		t.Fatalf("full=%v", full)
	}
	onlyName := applyRequires(ins, []string{attrDisplayName})
	if _, ok := onlyName[attrIAMPath]; ok {
		t.Fatalf("path should be filtered: %v", onlyName)
	}
	if onlyName[attrDisplayName] != "n" {
		t.Fatalf("display_name missing: %v", onlyName)
	}
}

func TestResolveParent(t *testing.T) {
	typ, id := ResolveParent(Filter{Parent: ResourceParent{Type: Cluster, ID: "c1"}})
	if typ != Cluster || id != "c1" {
		t.Fatalf("parent=%s %s", typ, id)
	}
	typ, id = ResolveParent(Filter{Ancestors: []Ancestor{{Type: Project, ID: "p"}, {Type: Cluster, ID: "c"}}})
	if typ != Cluster || id != "c" {
		t.Fatalf("ancestor=%s %s", typ, id)
	}
}

func TestUniqueApproversAndPath(t *testing.T) {
	got := UniqueApprovers("a", "", "a", "b")
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("approvers=%v", got)
	}
	if ProjectPath("p") != "/project,p/" {
		t.Fatal(ProjectPath("p"))
	}
	if NamespacePath("p", "c") != "/project,p/cluster,c/" {
		t.Fatal(NamespacePath("p", "c"))
	}
}

func TestParseNSID(t *testing.T) {
	got, err := ParseNSID("40000:abcdns")
	if err != nil || got != "BCS-K8S-40000" {
		t.Fatalf("got=%s err=%v", got, err)
	}
	if _, err := ParseNSID("bad"); err == nil {
		t.Fatal("want invalid ns id")
	}
}

func TestValidateCallback(t *testing.T) {
	err := validateCallback(CallbackRequest{Type: "unknown", Method: methodListInstance})
	if err == nil {
		t.Fatal("want type error")
	}
	err = validateCallback(CallbackRequest{Type: Project, Method: "search_instance"})
	if err == nil {
		t.Fatal("want method error")
	}
	err = validateCallback(CallbackRequest{Type: Project, Method: methodListInstance, Page: Page{PageSize: 2000}})
	if err == nil {
		t.Fatal("want page_size error")
	}
	err = validateCallback(CallbackRequest{Type: Project, Method: methodFetchInstanceInfo})
	if err == nil {
		t.Fatal("want empty ids error")
	}
}
