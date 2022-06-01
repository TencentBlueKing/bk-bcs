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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

var nodeManager = &NodeManager{}

var defaultCommonOption = &cloudprovider.CommonOption{
	Key:    "xxx",
	Secret: "xxx",
}

func TestNodeManager_GetCVMImageIDByImageName(t *testing.T) {
	imageName1 := "Tencent tlinux xxx"
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
	defaultCommonOption.Region = regions.Nanjing

	regions, err := nodeManager.GetCloudRegions(defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(regions)
}

func TestNodeManager_GetZoneList(t *testing.T) {
	defaultCommonOption.Region = regions.Nanjing

	zones, err := nodeManager.GetZoneList(defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(zones)
}

func TestNodeManager_GetNodeByIP(t *testing.T) {
	defaultCommonOption.Region = regions.Guangzhou

	node, err := nodeManager.GetNodeByIP("10.0.xx", &cloudprovider.GetNodeOption{
		Common:       defaultCommonOption,
		ClusterVPCID: "vpc-6jhti3nx",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(node)
}

func TestNodeManager_ListNodesByIP(t *testing.T) {
	IPList := []string{"xxx"}

	defaultCommonOption.Region = regions.Guangzhou
	nodes, err := nodeManager.ListNodesByIP(IPList, &cloudprovider.ListNodesOption{
		Common:       defaultCommonOption,
		ClusterVPCID: "vpc-xxx",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(nodes)
}

func TestNodeManager_ListNodesByInstance(t *testing.T) {
	instanceList := []string{"ins-xxx", "ins-xxx"}

	defaultCommonOption.Region = regions.Guangzhou
	nodes, err := nodeManager.ListNodesByInstanceID(instanceList, &cloudprovider.ListNodesOption{
		Common:       defaultCommonOption,
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

func TestDescribeInstanceTypeConfigs(t *testing.T) {
	filters := []*Filter{
		{Name: "zone", Values: []string{"ap-guangzhou-3"}},
	}
	instanceTypeConfigs, err := nodeManager.DescribeInstanceTypeConfigs(filters, &cloudprovider.CommonOption{
		Key:    "xxx",
		Secret: "xxx",
		Region: regions.Guangzhou,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(utils.ToJSONString(instanceTypeConfigs))
}
