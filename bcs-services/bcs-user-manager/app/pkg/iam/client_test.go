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
	"os"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"

	"github.com/TencentBlueKing/iam-go-sdk"
	"github.com/TencentBlueKing/iam-go-sdk/logger"
	"github.com/sirupsen/logrus"
)

const (
	SystemID  = ""
	AppCode   = ""
	AppSecret = ""

	GateWayHost = ""
	IAMHost     = ""
	BkiIAMHost  = ""
)

var opts *Options = &Options{
	SystemID:    SystemID,
	AppCode:     AppCode,
	AppSecret:   AppSecret,
	External:    false,
	GateWayHost: GateWayHost,
	IAMHost:     IAMHost,
	BkiIAMHost:  BkiIAMHost,
	Metric:      false,
}

func newIAMClient() (PermIAMClient, error) {
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
		SystemID: SystemID,
		UserName: "xxxx",
	}

	allow, err := cli.IsAllowedWithoutResource(actionID, req)
	if err != nil {
		t.Fatalf("IsAllowedWithoutResource failed: %v", err)
	}

	t.Log(allow)
}

func TestIamClient_IsAllowedWithResource(t *testing.T) {
	// create a logger
	log := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	// do set logger
	logger.SetLogger(log)

	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	ns, _ := utils.CalIAMNamespaceID("BCS-K8S-xxxxx", "default")
	fmt.Println(ns)

	var (
		req = PermissionRequest{
			SystemID: SystemID,
			UserName: "xxxx",
		}

		actionID = "namespace_scoped_update"
		rn       = []ResourceNode{
			{
				System:    SystemID,
				RType:     "namespace",
				RInstance: ns,
				Rp: NamespaceResourcePath{
					ProjectID: "xxxxx",
					ClusterID: "BCS-K8S-xxxx",
				},
			},
		}
	)

	allow, err := cli.IsAllowedWithResource(actionID, req, rn)
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}
	fmt.Println()

	t.Log(allow)
}

func TestIamClient_BatchIsAllowed(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	var (
		req = PermissionRequest{
			SystemID: AppCode,
			UserName: "xxxx",
		}
		actionID = "manager_namespace"
		rn1      = []ResourceNode{
			{
				System:    AppCode,
				RType:     "namespace",
				RInstance: "namespace1",
				Rp: ClusterScopedResourcePath{
					ProjectID: "namespace",
				},
			},
		}

		rn2 = []ResourceNode{
			{
				System:    AppCode,
				RType:     "namespace",
				RInstance: "namespace100",
				Rp: ClusterScopedResourcePath{
					ProjectID: "namespace",
				},
			},
		}
	)

	permission, err := cli.BatchIsAllowed(actionID, req, [][]ResourceNode{rn1, rn2})
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}

	t.Log(permission)
}

func TestIamClient_ResourceMultiActionsAllowed(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	var (
		req = PermissionRequest{
			SystemID: AppCode,
			UserName: "xxxx",
		}
		actions = []string{"manager_namespace", "plain_namespace", "senior_namespace"}
		rn1     = ResourceNode{
			System:    AppCode,
			RType:     "namespace",
			RInstance: "namespace1",
			Rp: ClusterScopedResourcePath{
				ProjectID: "namespace",
			},
		}
	)

	permission, err := cli.ResourceMultiActionsAllowed(actions, req, []ResourceNode{rn1})
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}

	t.Log(permission)
}

func TestIamClient_BatchResourceMultiActionsAllowed(t *testing.T) {
	cli, err := newIAMClient()
	if err != nil {
		t.Fatalf("newIAMClient failed: %v", err)
	}

	var (
		req = PermissionRequest{
			SystemID: AppCode,
			UserName: "xxxx",
		}
		actions = []string{"manager_namespace", "plain_namespace", "senior_namespace"}
		rn1     = []ResourceNode{
			{
				System:    AppCode,
				RType:     "namespace",
				RInstance: "namespace1",
				Rp: ClusterScopedResourcePath{
					ProjectID: "namespace",
				},
			},
		}

		rn2 = []ResourceNode{
			{
				System:    AppCode,
				RType:     "namespace",
				RInstance: "namespace100",
				Rp: ClusterScopedResourcePath{
					ProjectID: "namespace",
				},
			},
		}
	)

	permission, err := cli.BatchResourceMultiActionsAllowed(actions, req, [][]ResourceNode{rn1, rn2})
	if err != nil {
		t.Fatalf("IsAllowedWithResource failed: %v", err)
	}

	t.Log(permission)
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
		BkUserName: "xxxx",
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
		SystemID: AppCode,
		ActionID: "access_developer_center",
	}

	url, err := cli.GetApplyURL(req, []iam.ApplicationRelatedResourceType{}, BkUser{
		BkToken:    "xxxx",
		BkUserName: "",
	})
	if err != nil {
		t.Fatalf("GetApplyURL without resource failed: %v", err)
	}
	t.Log(url)

	relatedResource := []iam.ApplicationRelatedResourceType{}

	req = ApplicationRequest{
		SystemID: AppCode,
		ActionID: "manager_namespace",
	}

	// first instance
	nodes := []RelatedResourceNode{
		{
			Type: "cluster",
			ID:   "bcs-k8s-yyyyy",
		},
		{
			Type: "namespace",
			ID:   "default",
		},
	}

	// second instance
	nodes1 := []RelatedResourceNode{
		{
			Type: "cluster",
			ID:   "bcs-k8s-xxxxx",
		},
		{
			Type: "namespace",
			ID:   "zzz",
		},
	}
	instance1 := RelatedResourceLevel{Nodes: nodes}.BuildInstance()
	instance2 := RelatedResourceLevel{Nodes: nodes1}.BuildInstance()

	relatedResource1 := RelatedResourceType{
		SystemID: AppCode,
		RType:    "namespace",
	}.BuildRelatedResource([]iam.ApplicationResourceInstance{instance1, instance2})

	relatedResource = append(relatedResource, relatedResource1)

	url, err = cli.GetApplyURL(req, relatedResource, BkUser{
		BkToken:    "xxxx",
		BkUserName: "",
	})
	if err != nil {
		t.Fatalf("GetApplyURL withResource failed: %v", err)
	}

	t.Log(url)
}
