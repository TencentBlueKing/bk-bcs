/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmdb

import (
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/gse"
)

func GetGseTestClient() *gse.Client {
	cli, err := gse.NewGseClient(gse.Options{
		Enable:     true,
		AppCode:    "xxx",
		AppSecret:  "xxx",
		Server:     "xxx",
		BKUserName: "admin",
		Debug:      true,
	})
	if err != nil {
		return nil
	}

	return cli
}

var cmdb = getNewClient()
var gseCli = GetGseTestClient()

func TestGetHostCountByObject(t *testing.T) {
	cli := getNewClient()

	cnt, err := GetHostCountByObject(cli, 2, Object{
		ObjectName: "biz",
		ObjectID:   2,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cnt)

	cnt, err = GetHostCountByObject(cli, 2, Object{
		ObjectName: "biz",
		ObjectID:   2,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cnt)

}

func TestGetBizModuleTopoData(t *testing.T) {
	cli := getNewClient()

	data, err := GetBizModuleTopoData(cli, 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(data)
}

func TestGetBizHostDetailedData(t *testing.T) {
	cmdb := getNewClient()
	gseCli := GetGseTestClient()

	current := time.Now()
	hostInfo, err := GetBizHostDetailedData(cmdb, gseCli, 2, []HostModuleInfo{
		{
			ObjectID:   "biz",
			InstanceID: 2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(time.Since(current))
	t.Log(hostInfo)
}

func TestIpSelectorClient_GetBizModuleTopoData(t *testing.T) {
	cli := NewIpSelector(cmdb, gseCli)

	topo, err := cli.GetBizModuleTopoData(2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(topo)
}

func TestIpSelectorClient_GetCustomSettingModuleList(t *testing.T) {
	cli := NewIpSelector(cmdb, gseCli)

	setting := cli.GetCustomSettingModuleList([]string{IpSelectorHostList.String()})
	t.Log(setting)
}

func TestGetBizTopoHostDataTopoNodeFilter(t *testing.T) {
	cli := NewIpSelector(cmdb, gseCli)

	var b = 1
	filter := &HostFilterTopoNodes{
		Alive:         &b,
		SearchContent: "253.227",
	}
	hosts, err := cli.GetBizTopoHostData(2, []HostModuleInfo{
		{
			ObjectID:   "biz",
			InstanceID: 2,
		},
	}, filter)
	if err != nil {
		t.Fatal(err)
	}

	ips := make([]string, 0)
	for i := range hosts {
		ips = append(ips, hosts[i].Ip)
	}

	t.Log(len(ips), ips)
}

func TestGetBizTopoHostDataCheckNodesFilter(t *testing.T) {
	cli := NewIpSelector(cmdb, gseCli)

	filter := &HostFilterCheckNodes{
		IpList: []string{"xx", "xx", "0:xx", "0:xx"},

		Ipv6List: nil,
		KeyList:  []string{"VM", "centos"},
	}
	hosts, err := cli.GetBizTopoHostData(2, []HostModuleInfo{
		{
			ObjectID:   "biz",
			InstanceID: 2,
		},
	}, filter)
	if err != nil {
		t.Fatal(err)
	}

	ips := make([]string, 0)
	for i := range hosts {
		ips = append(ips, hosts[i].Ip)
	}

	t.Log(len(ips), ips)
}
