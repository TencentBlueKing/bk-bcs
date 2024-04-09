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

package cluster

import (
	"os"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

var (
	AppCode   = os.Getenv("APP_CODE")
	AppSecret = os.Getenv("APP_SECRET")

	GateWayHost = os.Getenv("IAM_HOST")
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

func newBcsClusterPermCli() (*BCSClusterPerm, error) {
	cli, err := iam.NewIamClient(opts)
	if err != nil {
		return nil, err
	}

	return NewBCSClusterPermClient(cli), nil
}

func TestBCSClusterPerm_CanCreateCluster(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	// nolint
	projectID := "b37778ec757544868a01e1f01f07037f"
	// projectID := "846e8195d9ca4097b354ed190acce4b1"
	allow, url, _, err := cli.CanCreateCluster("liming", projectID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestBCSClusterPerm_CanManageCluster(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	projectID := "b37778ec757544868a01e1f01f07037f"
	// projectID := "846e8195d9ca4097b354ed190acce4b1"
	allow, url, _, err := cli.CanManageCluster("liming", projectID, "BCS-K8S-15091")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestBCSClusterPerm_CanViewCluster(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	// projectID := "b37778ec757544868a01e1f01f07037f"
	projectID := "846e8195d9ca4097b354ed190acce4b1"
	allow, url, _, err := cli.CanViewCluster("liming", projectID, "BCS-K8S-15091")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestBCSClusterPerm_CanDeleteCluster(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	projectID := "b37778ec757544868a01e1f01f07037f"
	// projectID := "846e8195d9ca4097b354ed190acce4b1"
	allow, url, _, err := cli.CanDeleteCluster("liming", projectID, "BCS-K8S-15091")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestGetClusterMultiActionPermission(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	projectID := "b37778ec757544868a01e1f01f07037f"
	actionIDs := []string{ClusterView.String(), ClusterManage.String(), ClusterDelete.String()}
	allow, err := cli.GetClusterMultiActionPermission("liming", projectID, "BCS-K8S-15091", actionIDs)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow)
}

func TestGetMultiClusterMultiActionPerm(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	projectID := "b37778ec757544868a01e1f01f07037f"
	actionIDs := []string{ClusterView.String(), ClusterManage.String(), ClusterDelete.String()}
	clusterIDs := []string{"BCS-K8S-15091", "BCS-K8S-15092"}
	allow, err := cli.GetMultiClusterMultiActionPerm("liming", projectID, clusterIDs, actionIDs)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow)
}

func TestCanCreateClusterScopedResource(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	var (
		projectID = os.Getenv("PROJECT_ID")
		clusterID = os.Getenv("CLUSTER_ID")
		username  = os.Getenv("PERM_USERNAME")
	)
	allow, url, _, err := cli.CanCreateClusterScopedResource(username, projectID, clusterID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestCanViewClusterScopedResource(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	var (
		projectID = os.Getenv("PROJECT_ID")
		clusterID = os.Getenv("CLUSTER_ID")
		username  = os.Getenv("PERM_USERNAME")
	)
	allow, url, _, err := cli.CanViewClusterScopedResource(username, projectID, clusterID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestCanUpdateClusterScopedResource(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	var (
		projectID = os.Getenv("PROJECT_ID")
		clusterID = os.Getenv("CLUSTER_ID")
		username  = os.Getenv("PERM_USERNAME")
	)
	allow, url, _, err := cli.CanUpdateClusterScopedResource(username, projectID, clusterID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestCanDeleteClusterScopedResource(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	var (
		projectID = os.Getenv("PROJECT_ID")
		clusterID = os.Getenv("CLUSTER_ID")
		username  = os.Getenv("PERM_USERNAME")
	)
	allow, url, _, err := cli.CanDeleteClusterScopedResource(username, projectID, clusterID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}
