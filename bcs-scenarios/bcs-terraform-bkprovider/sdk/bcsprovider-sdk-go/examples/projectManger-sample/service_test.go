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

// Package helmMangerSample 测试
package helmMangerSample

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/projectManger"
)

const (
	// username rtx
	username = ""

	// token 个人token
	token = ""

	// gatewayApi xxx
	gatewayApi = "" // 此处结尾多一个"/"，少一个"/"不影响
)

var (
	gClient sdk.Client

	service projectManger.Service
)

func init() {
	config := &options.Config{
		Username:       username,
		Token:          token,
		BcsGatewayAddr: gatewayApi,
	}

	client, err := sdk.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("new sdk client err: %s", err.Error()))
	}
	gClient = client

	service = gClient.ProjectManger()

	log.Printf("config: %s", utils.ObjToPrettyJson(client.Config()))
}

// Test_Create 创建
func Test_Create(t *testing.T) {
	req := &projectManger.CreateProjectRequest{
		ProjectCode: "huiwen-test-project", // 项目英文名称
		Name:        "huiwen测试项目",
		Description: "测试项目",
		BusinessID:  "xxxx", // 必须是该业务下的业务运维
	}

	resp, err := service.CreateProject(context.TODO(), req)
	if err != nil {
		t.Fatalf("create project failed, err: %s", err.Error())
	}

	log.Printf("create project success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_Delete 删除
func Test_Delete(t *testing.T) {
	req := &projectManger.DeleteProjectRequest{
		ProjectID: "xxxxx", // 项目id
	}

	resp, err := service.DeleteProject(context.TODO(), req)
	if err != nil {
		t.Fatalf("delete project failed, err: %s", err.Error())
	}

	log.Printf("delete project success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_Update 修改
func Test_Update(t *testing.T) {
	req := &projectManger.UpdateProjectRequest{
		ProjectID:   "xxxxx", // 项目id
		Name:        "debug-test",
		Description: "debug-test",
		BusinessID:  "xxxx",
		Managers:    "huiwen1;huiwen2",
		Creator:     "huiwen",
	}

	resp, err := service.UpdateProject(context.TODO(), req)
	if err != nil {
		t.Fatalf("update project failed, err: %s", err.Error())
	}

	log.Printf("update project success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_Get 查询
func Test_Get(t *testing.T) {
	req := &projectManger.GetProjectRequest{
		ProjectID: "xxxx", // 项目id
	}

	resp, err := service.GetProject(context.TODO(), req)
	if err != nil {
		t.Fatalf("get project failed, err: %s", err.Error())
	}

	log.Printf("get project success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_List 查询
func Test_List(t *testing.T) {
	resp, err := service.ListProjects(context.TODO(), &projectManger.ListProjectsRequest{})
	if err != nil {
		t.Fatalf("list project failed, err: %s", err.Error())
	}

	log.Printf("list project success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_ProjectCodeRegexp 测试正则语法
func Test_ProjectCodeRegexp(t *testing.T) {
	regex, err := regexp.Compile(`^[a-z][a-z0-9-]*$`)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return
	}

	cases := []string{
		"test-string",
		"test123-string",
		"tEST123-string", // no
		"test123-STRING", // no
		"TEST123-STRING", // no
		"TEST123_STRING", // no
		"test123_string", // no
		"Test-string",    // no
		"test_string",    // no
		"123test-string", // no
	}

	for _, str := range cases {
		if regex.MatchString(str) {
			fmt.Printf("%s matches the regex\n", str)
		} else {
			fmt.Printf("%s doesn't match the regex\n", str)
		}
	}
}
