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
	"context"
	"errors"
	"testing"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

type fakePermClient struct {
	withoutRes func(actionID string, request iam.PermissionRequest, cache bool) (bool, error)
	withRes    func(actionID string, request iam.PermissionRequest, nodes []iam.ResourceNode, cache bool) (bool, error)
	multi      func(actions []string, request iam.PermissionRequest, nodes []iam.ResourceNode) (map[string]bool, error)
	batch      func(actions []string, request iam.PermissionRequest, nodes [][]iam.ResourceNode) (map[string]map[string]bool, error)
	applyURL   func(request iam.ApplicationRequest, related []iam.ApplicationAction, user iam.BkUser) (string, error)
}

func (f *fakePermClient) IsAllowedWithoutResource(actionID string, request iam.PermissionRequest, cache bool) (bool, error) {
	return f.withoutRes(actionID, request, cache)
}

func (f *fakePermClient) IsAllowedWithResource(actionID string, request iam.PermissionRequest, nodes []iam.ResourceNode,
	cache bool) (bool, error) {
	return f.withRes(actionID, request, nodes, cache)
}

func (f *fakePermClient) BatchResourceIsAllowed(string, iam.PermissionRequest, [][]iam.ResourceNode) (map[string]bool, error) {
	return nil, errors.New("unused")
}

func (f *fakePermClient) MultiActionsAllowedWithoutResource([]string, iam.PermissionRequest) (map[string]bool, error) {
	return nil, errors.New("unused")
}

func (f *fakePermClient) ResourceMultiActionsAllowed(actions []string, request iam.PermissionRequest,
	nodes []iam.ResourceNode) (map[string]bool, error) {
	return f.multi(actions, request, nodes)
}

func (f *fakePermClient) BatchResourceMultiActionsAllowed(actions []string, request iam.PermissionRequest,
	nodes [][]iam.ResourceNode) (map[string]map[string]bool, error) {
	return f.batch(actions, request, nodes)
}

func (f *fakePermClient) GetToken() (string, error) { return "", errors.New("unused") }

func (f *fakePermClient) IsBasicAuthAllowed(iam.BkUser) error { return errors.New("unused") }

func (f *fakePermClient) GetApplyURL(request iam.ApplicationRequest, related []iam.ApplicationAction,
	user iam.BkUser) (string, error) {
	return f.applyURL(request, related, user)
}

func (f *fakePermClient) CreateGradeManagers(context.Context, iam.GradeManagerRequest) (uint64, error) {
	return 0, errors.New("unused")
}

func (f *fakePermClient) CreateUserGroup(context.Context, uint64, iam.CreateUserGroupRequest) ([]uint64, error) {
	return nil, errors.New("unused")
}

func (f *fakePermClient) DeleteUserGroup(context.Context, uint64) error { return errors.New("unused") }

func (f *fakePermClient) AddUserGroupMembers(context.Context, uint64, iam.AddGroupMemberRequest) error {
	return errors.New("unused")
}

func (f *fakePermClient) DeleteUserGroupMembers(context.Context, uint64, iam.DeleteGroupMemberRequest) error {
	return errors.New("unused")
}

func (f *fakePermClient) CreateUserGroupPolicies(context.Context, uint64, iam.AuthorizationScope) error {
	return errors.New("unused")
}

func (f *fakePermClient) AuthResourceCreatorPerm(context.Context, iam.ResourceCreator, []iam.Ancestor) error {
	return errors.New("unused")
}

func setupFakeIAM(t *testing.T, fake *fakePermClient) *IAMClient {
	t.Helper()
	conf.G = &conf.GlobalConf{
		IAM: conf.IAMConf{
			SystemID: "bk_bcs_app",
			Cli:      func(string) iam.PermClient { return fake },
		},
	}
	return NewIAMClient()
}

func TestIAMClientResTypeAllowed(t *testing.T) {
	cli := setupFakeIAM(t, &fakePermClient{
		withoutRes: func(actionID string, request iam.PermissionRequest, cache bool) (bool, error) {
			if actionID != "project_create" || request.UserName != "alice" || request.SystemID != "bk_bcs_app" || !cache {
				t.Fatalf("unexpected args: %s %+v %v", actionID, request, cache)
			}
			return true, nil
		},
	})
	allow, err := cli.ResTypeAllowed("t1", "alice", "project_create", true)
	if err != nil || !allow {
		t.Fatalf("ResTypeAllowed: %v %v", allow, err)
	}
}

func TestIAMClientResInstAllowed(t *testing.T) {
	cli := setupFakeIAM(t, &fakePermClient{
		withRes: func(actionID string, request iam.PermissionRequest, nodes []iam.ResourceNode, cache bool) (bool, error) {
			if actionID != "cluster_view" || request.UserName != "bob" || cache {
				t.Fatalf("unexpected args: %s %+v %v", actionID, request, cache)
			}
			if len(nodes) != 1 || nodes[0].RInstance != "BCS-K8S-1" || nodes[0].RType != "cluster" {
				t.Fatalf("nodes=%+v", nodes)
			}
			if nodes[0].Attr[string(iam.BkIAMPath)] != "/project,p1/" {
				t.Fatalf("path lost: %+v", nodes[0].Attr)
			}
			return false, nil
		},
	})
	allow, err := cli.ResInstAllowed("t1", "bob", "cluster_view", []bkiam.ResourceNode{{
		System: "bk_bcs_app",
		Type:   "cluster",
		ID:     "BCS-K8S-1",
		Attribute: map[string]interface{}{
			string(iam.BkIAMPath): "/project,p1/",
		},
	}}, false)
	if err != nil || allow {
		t.Fatalf("ResInstAllowed: %v %v", allow, err)
	}
}

func TestIAMClientResInstMultiActionsAllowed(t *testing.T) {
	cli := setupFakeIAM(t, &fakePermClient{
		multi: func(actions []string, request iam.PermissionRequest, nodes []iam.ResourceNode) (map[string]bool, error) {
			if request.UserName != "carol" || len(actions) != 2 || len(nodes) != 1 {
				t.Fatalf("unexpected: %v %+v %+v", actions, request, nodes)
			}
			return map[string]bool{"cluster_view": true, "cluster_manage": false}, nil
		},
	})
	got, err := cli.ResInstMultiActionsAllowed("t1", "carol", []string{"cluster_view", "cluster_manage"},
		[]bkiam.ResourceNode{{Type: "cluster", ID: "c1"}})
	if err != nil || !got["cluster_view"] || got["cluster_manage"] {
		t.Fatalf("multi: %+v %v", got, err)
	}
}

func TestIAMClientBatchResMultiActionsAllowed(t *testing.T) {
	cli := setupFakeIAM(t, &fakePermClient{
		batch: func(actions []string, request iam.PermissionRequest, nodes [][]iam.ResourceNode) (map[string]map[string]bool, error) {
			if request.UserName != "dave" || len(nodes) != 2 || len(nodes[0]) != 1 || nodes[1][0].RInstance != "c2" {
				t.Fatalf("batch nodes=%+v user=%s", nodes, request.UserName)
			}
			return map[string]map[string]bool{
				"c1": {"cluster_view": true},
				"c2": {"cluster_view": false},
			}, nil
		},
	})
	got, err := cli.BatchResMultiActionsAllowed("t1", "dave", []string{"cluster_view"}, []bkiam.ResourceNode{
		{Type: "cluster", ID: "c1"},
		{Type: "cluster", ID: "c2"},
	})
	if err != nil || !got["c1"]["cluster_view"] || got["c2"]["cluster_view"] {
		t.Fatalf("batch: %+v %v", got, err)
	}
}

func TestApplyURLGeneratorGen(t *testing.T) {
	fake := &fakePermClient{
		applyURL: func(request iam.ApplicationRequest, related []iam.ApplicationAction, user iam.BkUser) (string, error) {
			if request.SystemID != "bk_bcs_app" || user.BkUserName != "erin" {
				t.Fatalf("apply args: %+v %+v", request, user)
			}
			if len(related) != 1 || related[0].ActionID != "project_view" {
				t.Fatalf("related=%+v", related)
			}
			return "https://iam.example/apply", nil
		},
	}
	conf.G = &conf.GlobalConf{
		IAM: conf.IAMConf{
			SystemID: "bk_bcs_app",
			Cli:      func(string) iam.PermClient { return fake },
		},
	}
	url, err := NewApplyURLGenerator().Gen("t1", "erin", []ActionResourcesRequest{{
		ActionID: "project_view",
		ResType:  "project",
		ResIDs:   []string{"p1"},
	}})
	if err != nil || url != "https://iam.example/apply" {
		t.Fatalf("Gen: %s %v", url, err)
	}
}
