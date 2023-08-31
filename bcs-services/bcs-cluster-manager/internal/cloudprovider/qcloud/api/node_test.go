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
	"os"
	"testing"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

var nodeManager = &NodeManager{}

var defaultCommonOption = &cloudprovider.CommonOption{
	Account: &cmproto.Account{
		SecretID:  os.Getenv(TencentCloudSecretIDClusterEnv),
		SecretKey: os.Getenv(TencentCloudSecretKeyClusterEnv),
	},
	CommonConf: cloudprovider.CloudConf{
		CloudInternalEnable: true,
		MachineDomain:       "cvm.internal.tencentcloudapi.com",
	},
}

func TestGetImageInfoByImageID(t *testing.T) {
	imageName1 := "img-xxx"
	defaultCommonOption.Region = regions.Nanjing
	image, err := nodeManager.GetImageInfoByImageID(imageName1, defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", *image.OsName)
}

func TestGetCVMImageIDByImageName(t *testing.T) {
	imageName1 := "TencentOS Server 2.6 (TK4)"
	defaultCommonOption.Region = regions.Nanjing
	imageID, err := nodeManager.GetCVMImageIDByImageName(imageName1, defaultCommonOption)
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
	defaultCommonOption.Region = "ap-nanjing"
	node, err := nodeManager.GetNodeByIP("xxx", &cloudprovider.GetNodeOption{
		Common:       defaultCommonOption,
		ClusterVPCID: "vpc-xxx",
		ClusterID:    "BCS-K8S-xxx",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(node)
}

func TestListNodeInstancesByInstanceID(t *testing.T) {
	idList := []string{"ins-xxx", "ins-xxx"}

	defaultCommonOption.Region = regions.Nanjing
	instances, err := nodeManager.ListNodeInstancesByInstanceID(idList, defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	for i := range instances {
		t.Log(*instances[i].InstanceId)
		for _, address := range instances[i].PrivateIpAddresses {
			t.Log(*address)
		}
		t.Log(*instances[i].SystemDisk.DiskType, *instances[i].SystemDisk.DiskSize)
		for _, disk := range instances[i].DataDisks {
			t.Log(*disk.DiskType, *disk.DiskSize)
		}
	}
}

func TestListNodeInstancesByIP(t *testing.T) {
	IPList := []string{"xxx", "xxx", "xxx"}

	defaultCommonOption.Region = regions.Nanjing
	instances, err := nodeManager.ListNodeInstancesByIP(IPList, defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	for i := range instances {
		t.Log(*instances[i].InstanceId)
		for _, address := range instances[i].PrivateIpAddresses {
			t.Log(*address)
		}
		t.Log(*instances[i].SystemDisk.DiskType, *instances[i].SystemDisk.DiskSize)
		for _, disk := range instances[i].DataDisks {
			t.Log(*disk.DiskType, *disk.DiskSize)
		}
	}
}

func TestNodeManager_ListNodesByIP(t *testing.T) {
	IPList := []string{"xxx"}

	defaultCommonOption.Region = regions.Nanjing
	nodes, err := nodeManager.ListNodesByIP(IPList, &cloudprovider.ListNodesOption{
		Common:       defaultCommonOption,
		ClusterVPCID: "vpc-xxx",
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := range nodes {
		t.Log(nodes[i].InnerIP, nodes[i].NodeID)
	}
}

func TestNodeManager_ListNodesByInstance(t *testing.T) {
	instanceList := []string{"ins-xxx", "ins-xxx"}

	defaultCommonOption.Region = regions.Nanjing
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
	defaultCommonOption.Region = regions.Nanjing
	cli, err := GetCVMClient(defaultCommonOption)
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
		{Name: "zone", Values: []string{"ap-xxx-3"}},
	}
	instanceTypeConfigs, err := nodeManager.DescribeInstanceTypeConfigs(filters, &cloudprovider.CommonOption{
		Account: &cmproto.Account{
			SecretID:  os.Getenv(TencentCloudSecretIDClusterEnv),
			SecretKey: os.Getenv(TencentCloudSecretKeyClusterEnv),
		},
		Region: regions.Guangzhou,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(utils.ToJSONString(instanceTypeConfigs))
}

func TestListNodeInstanceType(t *testing.T) {
	instanceTypeConfigs, err := nodeManager.ListNodeInstanceType(cloudprovider.InstanceInfo{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(utils.ToJSONString(instanceTypeConfigs))
}

func TestNodeManager_DescribeImages(t *testing.T) {
	defaultCommonOption.Region = regions.Nanjing
	images, err := nodeManager.DescribeImages("PUBLIC_IMAGE", defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	for i := range images {
		t.Log(images[i].ImageID, images[i].OsName, images[i].Provider, images[i].Status, images[i].Alias)
	}
}

func TestListKeyPairs(t *testing.T) {
	defaultCommonOption.Region = regions.Nanjing

	pairs, err := nodeManager.ListKeyPairs(defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	for i := range pairs {
		t.Log(pairs[i].GetKeyID(), pairs[i].GetKeyName(), pairs[i].GetDescription())
	}
}
