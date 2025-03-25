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

package common

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

func getClient() *Client {
	cli, err := NewClient(Options{
		EsbServer:  "xxx",
		Server:     "xxx",
		AppCode:    "xxx",
		BKUserName: "xxx",
		AppSecret:  "xxx",
		Debug:      true,
	})
	if err != nil {
		return nil
	}

	return cli
}

func TestClient_CreateBkOpsTask(t *testing.T) {
	cli := getClient()
	if cli == nil {
		t.Fatal("client nil")
	}

	response, err := cli.CreateBkOpsTask(&CreateTaskPathParas{
		BkBizID:    "3",
		TemplateID: "19",
		Operator:   "admin",
	}, &CreateTaskRequest{
		// 模板来源
		TemplateSource: string(BusinessTpl),
		TaskName:       fmt.Sprintf("add node:%s", "BCS-K8S-123"),
		Constants: map[string]string{
			"${ip}": "1.2.3.4",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(response.Data.TaskID)
}

func TestClient_StartBkOpsTask(t *testing.T) {
	cli := getClient()
	if cli == nil {
		t.Fatal("client nil")
	}

	req := &TaskPathParas{
		BkBizID:  "3",
		TaskID:   "26",
		Operator: "xxx",
	}
	result, err := cli.StartBkOpsTask(req, &StartTaskRequest{
		Scope: "cmdb_biz",
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result.Result)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select { // nolint
		case <-ticker.C:
		}

		data, err := cli.GetTaskStatus(req, &StartTaskRequest{
			Scope: "cmdb_biz",
		})
		if err != nil {
			fmt.Printf("RunBKsopsJob GetTaskStatus failed: %v", err)
			continue
		}

		fmt.Printf("RunBKsopsJob GetTaskStatus %s status %s", req.TaskID, data.Data.State)
		if data.Data.State == FINISHED.String() {
			break
		}
		if data.Data.State == FAILED.String() || data.Data.State == REVOKED.String() {
			fmt.Printf("RunBKsopsJob GetTaskStatus task[%s] failed; %v", req.TaskID, err)
			return
		}
	}

	t.Log("successful")
}

func TestClient_GetBusinessTemplateList(t *testing.T) {
	cli := getClient()
	if cli == nil {
		t.Fatal("client nil")
	}

	path := &TemplateListPathPara{
		BkBizID:  "100148",
		Operator: "",
	}
	req := &TemplateRequest{}

	templateList, err := cli.GetBusinessTemplateList(path, req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(len(templateList))
	for _, tmp := range templateList {
		t.Log(tmp.BkBizID, tmp.ID, tmp.Name, tmp.BkBizName, tmp.Creator)
	}
}

func TestClient_GetUserProjectDetailInfo(t *testing.T) {
	cli := getClient()
	if cli == nil {
		t.Fatal("client nil")
	}

	p, err := cli.GetUserProjectDetailInfo("106")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(p)
}

func TestClient_GetBusinessTemplateInfo(t *testing.T) {
	cli := getClient()
	if cli == nil {
		t.Fatal("client nil")
	}

	path := &TemplateDetailPathPara{
		BkBizID:    "xx",
		TemplateID: "xx",
		Operator:   "xx",
	}
	req := &TemplateRequest{}

	globalConstantValue, err := cli.GetBusinessTemplateInfo(path, req)
	if err != nil {
		t.Fatal(err)
	}

	for _, tmp := range globalConstantValue {
		t.Log(tmp.Key, tmp.Name, tmp.Desc, tmp.SourceType)
		t.Log(extractValue(tmp.Key))
	}
}

func extractValue(value string) string {
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return ""
	}

	holderRex := regexp.MustCompile(`^\$\{(.*?)\}`)
	subMatch := holderRex.FindAllStringSubmatch(value, -1)
	if len(subMatch) > 0 && len(subMatch[0]) >= 1 {
		return subMatch[0][1]
	}

	return ""
}
