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

package manager

import (
	"context"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

const (
	AppCode   = "xxx"
	AppSecret = "xxx"

	GateWayHost = "xxx"
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

func newPermManagerClient() (*PermManager, error) {
	cli, err := iam.NewIamClient(opts)
	if err != nil {
		return nil, err
	}

	return &PermManager{
		iamClient: cli,
	}, nil
}

func TestCreateProjectGradeManager(t *testing.T) {
	cli, err := newPermManagerClient()
	if err != nil {
		t.Fatal(err)
	}

	managerID, err := cli.CreateProjectGradeManager(context.Background(), []string{"xxx"},
		&GradeManagerInfo{
			Name: "",
			Desc: "",
			Project: &Project{
				ProjectID:   "xxx",
				ProjectCode: "xxx",
				Name:        "xxx",
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(managerID) // 445
}

func TestCreateProjectUserGroup(t *testing.T) {
	cli, err := newPermManagerClient()
	if err != nil {
		t.Fatal(err)
	}

	err = cli.CreateProjectUserGroup(context.Background(), 445, UserGroupInfo{
		Name:  "xxx",
		Desc:  "xxx",
		Users: []string{""},
		Project: &Project{
			ProjectID:   "",
			ProjectCode: "",
			Name:        "",
		},
		Policy: Manager,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = cli.CreateProjectUserGroup(context.Background(), 445, UserGroupInfo{
		Name:  "",
		Desc:  "",
		Users: []string{""},
		Project: &Project{
			ProjectID:   "",
			ProjectCode: "",
			Name:        "",
		},
		Policy: Viewer,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}
