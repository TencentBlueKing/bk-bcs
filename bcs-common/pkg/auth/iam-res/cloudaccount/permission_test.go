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

package cloudaccount

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/project"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam-res/utils"
)

const (
	AppCode   = "xxx"
	AppSecret = "xxx"

	GateWayHost = "http://xxx/prod"
)

var opts = &iam.Options{
	SystemID:    iam.SystemIDBKBCS,
	AppCode:     AppCode,
	AppSecret:   AppSecret,
	External:    false,
	GateWayHost: GateWayHost,
	Metric:      false,
	Debug:       true,
}

func newBcsClusterPermCli() (*BCSCloudAccountPerm, error) {
	cli, err := iam.NewIamClient(opts)
	if err != nil {
		return nil, err
	}

	return NewBCSAccountPermClient(cli), nil
}

func TestPerm_CanManageCloudAccount(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	var (
		projectID = "b37778ec757544868a01e1f01f07037f"
		accountID = ""
	)
	allow, url, err := cli.CanManageCloudAccount("liming", projectID, accountID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestPerm_CanUseCloudAccount(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	var (
		projectID = "b37778ec757544868a01e1f01f07037f"
		accountID = ""
	)
	allow, url, err := cli.CanUseCloudAccount("liming", projectID, accountID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestPerm_GetMultiActionPermission(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	projectID := "b37778ec757544868a01e1f01f07037f"
	actionIDs := []string{AccountManage.String(), AccountUse.String(), project.ProjectView.String()}
	accountIDs := []string{"BCS-K8S-15091", "BCS-K8S-15092"}
	allow, err := cli.GetMultiAccountMultiActionPerm("liming", projectID, accountIDs, actionIDs)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow)
}

func TestCanCreateCloudAccount(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	allow, url, err := cli.CanCreateCloudAccount("xxx", "xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestAuthorizeResourceCreatorPerm(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	err = cli.AuthorizeResourceCreatorPerm("xxx", utils.ResourceInfo{
		Type: "project",
		ID:   "xxx",
		Name: "xxx",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}
