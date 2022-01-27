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

package project

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

const (
	AppCode   = "bk_bcs"
	AppSecret = "xxx"

	GateWayHost = "http://xxx/stage"
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

func newBcsClusterPermCli() (*BCSProjectPerm, error) {
	cli, err := iam.NewIamClient(opts)
	if err != nil {
		return nil, err
	}

	return NewBCSProjectPermClient(cli), nil
}

func TestBCSProjectPerm_CanCreateProject(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	allow, url, err := cli.CanCreateProject("liming")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestBCSProjectPerm_CanEditProject(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	// projectID := "b37778ec757544868a01e1f01f07037f"
	projectID := "846e8195d9ca4097b354ed190acce4b1"
	allow, url, err := cli.CanEditProject("liming", projectID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestBCSProjectPerm_CanDeleteProject(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	projectID := "b37778ec757544868a01e1f01f07037f"
	// projectID := "846e8195d9ca4097b354ed190acce4b1"
	allow, url, err := cli.CanEditProject("liming", projectID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}

func TestBCSProjectPerm_CanViewProject(t *testing.T) {
	cli, err := newBcsClusterPermCli()
	if err != nil {
		t.Fatal(err)
	}

	// projectID := "b37778ec757544868a01e1f01f07037f"
	projectID := "846e8195d9ca4097b354ed190acce4b1"
	allow, url, err := cli.CanViewProject("liming", projectID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(allow, url)
}
