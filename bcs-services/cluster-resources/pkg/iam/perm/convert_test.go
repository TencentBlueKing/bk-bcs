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

package perm

import (
	"testing"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

func TestConvertResourceNodeWithIAMPath(t *testing.T) {
	src := bkiam.ResourceNode{
		System: "bk_bcs_app",
		Type:   "cluster",
		ID:     "BCS-K8S-40000",
		Attribute: map[string]interface{}{
			string(iam.BkIAMPath): "/project,p1/",
		},
	}
	got := convertResourceNode(src)
	if got.System != src.System || got.RType != src.Type || got.RInstance != src.ID {
		t.Fatalf("node fields mismatch: %+v", got)
	}
	if got.Rp.BuildIAMPath() != "/project,p1/" {
		t.Fatalf("unexpected iam path: %s", got.Rp.BuildIAMPath())
	}
	if got.Attr[string(iam.BkIAMPath)] != "/project,p1/" {
		t.Fatalf("attr path lost: %+v", got.Attr)
	}
	// V3 BuildResourceNode 依赖 Rp，转换后不能 panic
	_ = got.BuildResourceNode()
}

func TestConvertResourceNodeEmptyAttr(t *testing.T) {
	got := convertResourceNode(bkiam.ResourceNode{System: "bk_bcs_app", Type: "project", ID: "p1"})
	if got.Attr == nil {
		t.Fatal("attr should be empty map, not nil")
	}
	if got.Rp.BuildIAMPath() != "" {
		t.Fatalf("empty path expected, got %q", got.Rp.BuildIAMPath())
	}
	_ = got.BuildResourceNode()
}

func TestConvertApplicationActionParentChain(t *testing.T) {
	conf.G = &conf.GlobalConf{IAM: conf.IAMConf{SystemID: "bk_bcs_app"}}
	req := ActionResourcesRequest{
		ActionID: "namespace_view",
		ResType:  "namespace",
		ResIDs:   []string{"40000:abc"},
		ParentChain: []IAMRes{
			{ResType: "project", ResID: "p1"},
			{ResType: "cluster", ResID: "BCS-K8S-40000"},
		},
	}
	actions := convertApplicationActions([]ActionResourcesRequest{req})
	if len(actions) != 1 {
		t.Fatalf("actions len=%d", len(actions))
	}
	if actions[0].ActionID != "namespace_view" {
		t.Fatalf("action id=%s", actions[0].ActionID)
	}
	if len(actions[0].RelatedResources) != 1 {
		t.Fatalf("related types=%d", len(actions[0].RelatedResources))
	}
	rel := actions[0].RelatedResources[0]
	if rel.SystemID != "bk_bcs_app" || rel.Type != "namespace" {
		t.Fatalf("related type mismatch: %+v", rel)
	}
	if len(rel.Instances) != 1 || len(rel.Instances[0]) != 3 {
		t.Fatalf("parent chain lost: %+v", rel.Instances)
	}
	nodes := rel.Instances[0]
	if nodes[0].Type != "project" || nodes[0].ID != "p1" {
		t.Fatalf("parent project: %+v", nodes[0])
	}
	if nodes[1].Type != "cluster" || nodes[1].ID != "BCS-K8S-40000" {
		t.Fatalf("parent cluster: %+v", nodes[1])
	}
	if nodes[2].Type != "namespace" || nodes[2].ID != "40000:abc" {
		t.Fatalf("self node: %+v", nodes[2])
	}
}
