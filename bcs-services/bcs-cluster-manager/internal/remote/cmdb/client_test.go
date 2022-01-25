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

package cmdb

import (
	"testing"
)

func getNewClient() *Client {
	cli, err := NewCmdbClient(Options{
		Enable:     true,
		AppCode:    "xxx",
		AppSecret:  "xxx",
		Server:     "http://xxx.com",
		BKUserName: "bcs",
		Debug:      true,
	})
	if err != nil {
		return nil
	}

	return cli
}

func TestClient_QueryHostNumByBizID(t *testing.T) {
	cli := getNewClient()
	hosts, _, err := cli.QueryHostByBizID(0, Page{
		Start: 0,
		Limit: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hosts)
}

func TestClient_FetchAllHostsByBizID(t *testing.T) {
	cli := getNewClient()
	hosts, err := cli.FetchAllHostsByBizID(0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hosts)
}

func TestClient_GetBusinessMaintainer(t *testing.T) {
	cli := getNewClient()
	data, err := cli.GetBusinessMaintainer(1024)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(data.BKBizMaintainer)
}

func TestClient_GetBS2IDByBizID(t *testing.T) {
	cli := getNewClient()
	id, err := cli.GetBS2IDByBizID(0)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(id)
}
