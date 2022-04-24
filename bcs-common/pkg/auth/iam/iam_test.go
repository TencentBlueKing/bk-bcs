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
 *
 */

package iam

import (
	"fmt"
	"testing"
)

const (
	AppCode   = "bk_bcs"
	AppSecret = "xxx"

	GateWayHost = "http://xxx/stage"
)

var opts = &Options{
	SystemID:    SystemIDBKBCS,
	AppCode:     AppCode,
	AppSecret:   AppSecret,
	External:    false,
	GateWayHost: GateWayHost,
	Metric:      false,
	Debug:       true,
}

func newIAMClient() (PermClient, error) {
	client, err := NewIamClient(opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func TestIamClient_GetToken(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	token, err := cli.GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	t.Log(token)
}

func TestIamClient_IsAllowedWithoutResource(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	actionID := "access_developer_center"
	req := PermissionRequest{
		SystemID: SystemIDBKBCS,
		UserName: "evanxinli",
	}

	allow, err := cli.IsAllowedWithoutResource(actionID, req, false)
	if err != nil {
		t.Fatalf("IsAllowedWithoutResource failed: %v", err)
	}

	t.Log(allow)
}

func TestIamClient_IsAllowedWithResource(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	user := "evanxinli"
	req := PermissionRequest{
		SystemID: SystemIDBKBCS,
		UserName: user,
	}
	actionID := ClusterCreate
	rn := []ResourceNode{
		{
			System:    SystemIDBKBCS,
			RType:     string(SysProject),
			RInstance: "b37778ec757544868a01e1f01f07037f",
			Rp: ClusterResourcePath{
				ClusterCreate: true,
			},
		},
	}

	allow, err := cli.IsAllowedWithResource(actionID.String(), req, rn, false)
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}
	t.Log(allow)

	actionID = ClusterView
	rn = []ResourceNode{
		{
			System:    SystemIDBKBCS,
			RType:     string(SysCluster),
			RInstance: "BCS-K8S-15201",
			Rp: ClusterResourcePath{
				ProjectID: "b37778ec757544868a01e1f01f07037f",
			},
		},
	}

	allow, err = cli.IsAllowedWithResource(actionID.String(), req, rn, false)
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}

	t.Log(allow)
}

func TestIamClient_BatchResourceIsAllowed(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	var (
		req = PermissionRequest{
			SystemID: SystemIDBKBCS,
			UserName: "evanxinli",
		}
		actionID = "cluster_view"
		rn1      = []ResourceNode{
			{
				System:    SystemIDBKBCS,
				RType:     string(SysCluster),
				RInstance: "BCS-K8S-15201",
				Rp: ClusterResourcePath{
					ProjectID: "b37778ec757544868a01e1f01f07037f",
				},
			},
		}

		rn2 = []ResourceNode{
			{
				System:    SystemIDBKBCS,
				RType:     string(SysCluster),
				RInstance: "BCS-K8S-15200",
				Rp: ClusterResourcePath{
					ProjectID: "b37778ec757544868a01e1f01f07037f",
				},
			},
		}
	)

	permission, err := cli.BatchResourceIsAllowed(actionID, req, [][]ResourceNode{rn1, rn2})
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}

	t.Log(permission) // map[BCS-K8S-15200:true BCS-K8S-15201:true]
}

func TestIamClient_ResourceMultiActionsAllowed(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	var (
		req = PermissionRequest{
			SystemID: SystemIDBKBCS,
			UserName: "evanxinli",
		}
		actions = []string{"cluster_view", "cluster_delete", "cluster_manage"}
		rn1     = ResourceNode{
			System:    SystemIDBKBCS,
			RType:     string(SysCluster),
			RInstance: "BCS-K8S-15201",
			Rp: ClusterResourcePath{
				ProjectID: "846e8195d9ca4097b354ed190acce4b1",
			},
		}
	)

	permission, err := cli.ResourceMultiActionsAllowed(actions, req, []ResourceNode{rn1})
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}

	t.Log(permission) // map[cluster_delete:true cluster_manage:true cluster_view:true]
}

type Resource struct {
	Action       string
	User         string
	ResourceType string
	ResourceID   string
}

func TestIamClient_BatchResourceMultiActionsAllowed(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	resources := []Resource{
		{
			Action:       "cluster_view",
			ResourceType: string(SysCluster),
			ResourceID:   "BCS-K8S-15200",
		},
		{
			Action:       "project_view",
			ResourceType: string(SysProject),
			ResourceID:   "b37778ec757544868a01e1f01f07037d",
		},
	}

	var (
		req = PermissionRequest{
			SystemID: SystemIDBKBCS,
			UserName: "evanxinli",
		}
		actions = []string{"cluster_view", "project_view"}
		rn1     = []ResourceNode{
			{
				System:    SystemIDBKBCS,
				RType:     string(SysCluster),
				RInstance: "BCS-K8S-15200",
				Rp: ClusterResourcePath{
					ProjectID:     "b37778ec757544868a01e1f01f07037f",
					ClusterCreate: false,
				},
			},
		}

		rn2 = []ResourceNode{
			{
				System:    SystemIDBKBCS,
				RType:     string(SysProject),
				RInstance: "b37778ec757544868a01e1f01f07037d",
				Rp:        ProjectResourcePath{},
			},
		}
	)

	permission, err := cli.BatchResourceMultiActionsAllowed(actions, req, [][]ResourceNode{rn1, rn2})
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}

	t.Log(permission) // map[BCS-K8S-15200:map[cluster_delete:true cluster_manage:true cluster_view:true]]

	for _, r := range resources {
		perm, ok := permission[r.ResourceID]
		if ok {
			allow := perm[r.Action]
			fmt.Println(allow)
		}
	}
}

func TestIamClient_IsBasicAuthAllowed(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	token, err := cli.GetToken()
	if err != nil {
		t.Fatalf("GetToken failed: %v", token)
	}

	err = cli.IsBasicAuthAllowed(BkUser{
		BkToken:    token,
		BkUserName: "bk_iam",
	})
	if err != nil {
		t.Fatalf("IsBasicAuthAllowed failed: %v", err)
	}

	t.Log("IsBasicAuthAllowed successful")
}

func TestIamClient_GetApplyURL(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	req := ApplicationRequest{
		SystemID: SystemIDBKBCS,
	}

	actionApplication1 := ApplicationAction{
		ActionID:         "cluster_view",
		RelatedResources: make([]iam.ApplicationRelatedResourceType, 0),
	}
	resource1 := BuildRelatedResourceTypes(SystemIDBKBCS, string(SysCluster), []iam.ApplicationResourceInstance{
		BuildResourceInstance([]Instance{
			{
				ResourceType: string(SysProject),
				ResourceID:   "b37778ec757544868a01e1f01f07037f",
			},
			{
				ResourceType: string(SysCluster),
				ResourceID:   "BCS-K8S-15113",
			},
		}),
		BuildResourceInstance([]Instance{
			{
				ResourceType: string(SysProject),
				ResourceID:   "b37778ec757544868a01e1f01f07037f",
			},
			{
				ResourceType: string(SysCluster),
				ResourceID:   "BCS-K8S-15091",
			},
		}),
	})
	actionApplication1.RelatedResources = append(actionApplication1.RelatedResources, resource1)

	actionApplication2 := ApplicationAction{
		ActionID:         "cluster_create",
		RelatedResources: make([]iam.ApplicationRelatedResourceType, 0),
	}
	resource2 := BuildRelatedResourceTypes(SystemIDBKBCS, string(SysProject), []iam.ApplicationResourceInstance{
		BuildResourceInstance([]Instance{
			{
				ResourceType: string(SysProject),
				ResourceID:   "b37778ec757544868a01e1f01f07037f",
			},
		}),
		BuildResourceInstance([]Instance{
			{
				ResourceType: string(SysProject),
				ResourceID:   "846e8195d9ca4097b354ed190acce4b1",
			},
		}),
	})
	actionApplication2.RelatedResources = append(actionApplication2.RelatedResources, resource2)

	url, err := cli.GetApplyURL(req, []ApplicationAction{actionApplication1, actionApplication2}, BkUser{BkUserName: "bk_iam"})
	if err != nil {
		t.Fatalf("GetApplyURL withResource failed: %v", err)
	}

	t.Log(url)
}
