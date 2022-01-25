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

package common

import (
	"fmt"
	"testing"
	"time"
)

func getClient() *Client {
	cli, err := NewClient(Options{
		AppCode:   "xxx",
		AppSecret: "xxx",
		Debug:     true,
		External:  true,

		CreateTaskURL: "xxx",
		TaskStatusURL: "xxx",
		StartTaskURL:  "xxx",
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

	response, err := cli.CreateBkOpsTask("", &CreateTaskPathParas{
		BkBizID:    "2",
		TemplateID: "10004",
		Operator:   "xxx",
	}, &CreateTaskRequest{
		TaskName: fmt.Sprintf("add node:%s", "BCS-K8S-xxxxx"),
		Constants: map[string]string{
			"${ctrl_ip_list}": "1.2.3.4",
			"${cluster_id}":   "BCS-K8S-15000",
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
		BkBizID:  "2",
		TaskID:   "17977",
		Operator: "xxx",
	}
	result, err := cli.StartBkOpsTask("", req, &StartTaskRequest{})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result.Result)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		}

		data, err := cli.GetTaskStatus("", req, &StartTaskRequest{
			//Scope: "cmdb_biz",
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
