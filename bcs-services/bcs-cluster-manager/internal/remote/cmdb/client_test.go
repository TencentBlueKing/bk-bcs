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

package cmdb

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func getNewClient() *Client {
	cli, err := NewCmdbClient(Options{
		Enable:     true,
		AppCode:    "xxx",
		AppSecret:  "xxx",
		Server:     "xxx",
		BKUserName: "xxx",
		Debug:      true,
	})
	if err != nil {
		return nil
	}

	return cli
}

func TestClient_QueryHostNumByBizID(t *testing.T) {
	cli := getNewClient()
	hosts, _, err := cli.QueryHostByBizID(1, Page{
		Start: 0,
		Limit: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hosts)
}

func TestClient_QueryHostInfoWithoutBiz(t *testing.T) {
	cli := getNewClient()

	ips := []string{"x"}
	hostList, err := cli.QueryHostInfoWithoutBiz(FieldHostIP, ips, Page{
		Start: 0,
		Limit: len(ips),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, h := range hostList {
		fmt.Println(strings.ToLower(h.CpuModule))
		t.Log(h, h.BkCloudID, h.NormalDeviceType, h.IDCCityName, h.BkAgentID, h.CpuModule, h.SubZoneID)
	}
}

func TestClient_FindHostBizRelations(t *testing.T) {
	cli := getNewClient()

	hostBizRelations, err := cli.FindHostBizRelations([]int{2, 3})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hostBizRelations[0].BkHostID, hostBizRelations[0].BkBizID, hostBizRelations[0].BkModuleID,
		hostBizRelations[0].BkSetID)
	t.Log(hostBizRelations[1].BkHostID, hostBizRelations[1].BkBizID, hostBizRelations[1].BkModuleID,
		hostBizRelations[1].BkSetID)
}

func TestClient_TransHostToRecycleModule(t *testing.T) {
	cli := getNewClient()

	err := cli.TransHostToRecycleModule(1, []int{2, 3})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestClient_GetBusinessMaintainer(t *testing.T) {
	cli := getNewClient()
	data, err := cli.GetBusinessMaintainer(1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(data.BKBizMaintainer, data.BKBizName)
}

func TestClient_GetBS2IDByBizID(t *testing.T) {
	cli := getNewClient()
	id, err := cli.GetBS2IDByBizID(1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(id)
}

func TestClient_SearchBizInstTopo(t *testing.T) {
	cli := getNewClient()

	topos, err := cli.SearchBizInstTopo(100148)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(topos[0])
}

func TestClient_GetBizInternalModule(t *testing.T) {
	cli := getNewClient()

	internalModule, err := cli.GetBizInternalModule(context.Background(), 100148)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(internalModule)
}

func TestClient_ListTopology(t *testing.T) {
	cli := getNewClient()

	topoData, err := cli.ListTopology(context.Background(), 100148, true, true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(topoData)
}

func TestClient_FindHostTopoRelation(t *testing.T) {
	cli := getNewClient()

	cnt, data, err := cli.FindHostTopoRelation(100148, Page{
		Start: 0,
		Limit: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(cnt)
	t.Log(data)
}

func TestFetchAllHostTopoRelByBizID(t *testing.T) {
	cli := getNewClient()

	data, err := cli.FetchAllHostTopoRelationsByBizID(100148)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(data)
}

func TestClient_SearchCloudAreaByCloudID(t *testing.T) {
	cli := getNewClient()

	data, err := cli.SearchCloudAreaByCloudID(0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)
}
