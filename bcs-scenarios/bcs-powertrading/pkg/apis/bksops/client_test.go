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

package bksops

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

func Test_CreateTask(t *testing.T) {
	testCli := New(&apis.ClientOptions{
		Endpoint:    "",
		UserName:    "",
		AccessToken: "",
		AppCode:     "",
		AppSecret:   "",
	}, requester.NewRequester())
	constants := make(map[string]string)
	constants["${node_ip_list}"] = "11.187.113.19"
	constants["${biz_cc_id}"] = "100148"
	rsp, err := testCli.CreateTask("10179", "100148", "TKExIEG混部集群第三方节点上架前检测", constants)
	fmt.Println(err)
	fmt.Println(rsp)
	fmt.Println(rsp.Data.TaskID)
	fmt.Println(strconv.Itoa(int(rsp.Data.TaskID)))
}

func TestClient_GetTaskStatus(t *testing.T) {
	testCli := New(&apis.ClientOptions{
		Endpoint:    "",
		UserName:    "",
		AccessToken: "",
		AppCode:     "",
		AppSecret:   "",
	}, requester.NewRequester())
	constants := make(map[string]string)
	constants["${node_ip_list}"] = "11.187.113.19"
	constants["${biz_cc_id}"] = "100148"
	rsp, err := testCli.GetTaskStatus("38749998", "100148")
	fmt.Println(err)
	fmt.Println(rsp)
}

func TestClient_StartTask(t *testing.T) {
	testCli := New(&apis.ClientOptions{
		Endpoint:    "",
		UserName:    "",
		AccessToken: "",
		AppCode:     "",
		AppSecret:   "",
	}, requester.NewRequester())
	rsp, err := testCli.StartTask("38749998", "100148")
	fmt.Println(err)
	fmt.Println(rsp)
}

func TestClient_GetTaskNodeDetail(t *testing.T) {
	testCli := New(&apis.ClientOptions{
		Endpoint:    "",
		UserName:    "",
		AccessToken: "",
		AppCode:     "",
		AppSecret:   "",
	}, requester.NewRequester())
	rsp, err := testCli.GetTaskNodeDetail("38821919", "100148", "nef7ec2a1d2d3f1ebee6f7a373daf653")
	fmt.Println(err)
	fmt.Println(rsp)
	for _, output := range rsp.Data.Outputs {
		if output.Key == "job_inst_id" {
			fmt.Println(output.Value)
			switch x := output.Value.(type) {
			case int:
				id, _ := output.Value.(int)
				fmt.Println(strconv.Itoa(id))
			case string:
				id, _ := output.Value.(string)
				fmt.Println(id)
			case float64:
				id, _ := output.Value.(float64)
				fmt.Println(strconv.FormatFloat(id, 'f', -1, 64))
			default:
				id, ok := output.Value.(int64)
				if ok {
					fmt.Println(id)
				} else {
					fmt.Printf("not supported type:%s\n", x)
				}
			}
		}
	}
}
