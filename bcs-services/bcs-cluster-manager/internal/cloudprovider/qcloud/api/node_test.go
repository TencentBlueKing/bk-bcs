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

package api

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

var nodeManager = &NodeManager{}

func TestNodeManager_GetCVMImageIDByImageName(t *testing.T) {
	imageName1 := "Tencent tlinux release 2.2 (Final)"
	imageID, err := nodeManager.GetCVMImageIDByImageName(imageName1, &cloudprovider.CommonOption{
		Secret: "xxx",
		Key:    "xxx",
		Region: regions.Nanjing,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf(imageID)
}

func TestNodeManager_GetRegionsInfo(t *testing.T) {
	regions, err := nodeManager.GetRegionsInfo(&cloudprovider.CommonOption{
		Secret: "xxx",
		Key:    "xxx",
		Region: regions.Nanjing,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(regions)
}

func TestNodeManager_GetZoneList(t *testing.T) {
	zones, err := nodeManager.GetZoneList(&cloudprovider.CommonOption{
		Secret: "xxx",
		Key:    "xxx",
		Region: regions.Nanjing,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(zones)
}

func TestNodeManager_GetNodeByIP(t *testing.T) {
	node, err := nodeManager.GetNodeByIP("127.0.0.1", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(node)
}

func TestNodeManager_ListNodesByIP(t *testing.T) {
	IPList := []string{"127.0.0.1", "127.0.0.2"}

	nodes, err := nodeManager.ListNodesByIP(IPList, &cloudprovider.ListNodesOption{
		Common: &cloudprovider.CommonOption{
			Secret: "xxx",
			Key:    "xxx",
			Region: regions.Shanghai,
		},
		ClusterVPCID: "vpc-xxx",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(nodes)
}

func TestGetZoneInfoByRegion(t *testing.T) {
	cli, err := GetCVMClient(&cloudprovider.CommonOption{
		Secret: "xxx",
		Key:    "xxx",
		Region: regions.Nanjing,
	})
	if err != nil {
		t.Fatal(err)
	}

	zoneInfo, err := GetZoneInfoByRegion(cli, regions.Nanjing)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(zoneInfo)
}
