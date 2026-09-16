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

package iamv4

import "testing"

func TestLookupActionResourceType(t *testing.T) {
	typ, known := LookupActionResourceType(string(ProjectCreate))
	if !known || typ != "" {
		t.Fatalf("project_create: typ=%q known=%v", typ, known)
	}
	typ, known = LookupActionResourceType(string(ProjectView))
	if !known || typ != string(SysProject) {
		t.Fatalf("project_view: typ=%q known=%v", typ, known)
	}
	typ, known = LookupActionResourceType(string(NamespaceList))
	if !known || typ != string(SysCluster) {
		t.Fatalf("namespace_list: typ=%q known=%v", typ, known)
	}
	typ, known = LookupActionResourceType(string(NamespaceView))
	if !known || typ != string(SysNamespace) {
		t.Fatalf("namespace_view: typ=%q known=%v", typ, known)
	}
	if _, known = LookupActionResourceType("not_an_action"); known {
		t.Fatal("unknown action must not be known")
	}
}
