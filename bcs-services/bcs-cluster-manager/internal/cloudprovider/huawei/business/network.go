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

// Package business xxx
package business

import (
	"strings"

	modelv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
)

// GetCloudSubnetsByVpc get cloud subnets by vpc
func GetCloudSubnetsByVpc(vpcId string, common cloudprovider.CommonOption) ([]modelv2.Subnet, error) {
	client, err := api.NewVpcClient(&common)
	if err != nil {
		return nil, err
	}

	// huaweiCloud 子网无可用区属性
	return client.ListSubnets(vpcId)
}

// GetSubnetAvailableIpNum 获取子网可用ip数
func GetSubnetAvailableIpNum(subnetId string, common cloudprovider.CommonOption) (int32, error) {
	client, err := api.NewVpcClient(&common)
	if err != nil {
		return 0, err
	}

	rsp, err := client.ShowNetworkIpAvailabilities(subnetId)
	if err != nil {
		return 0, err
	}

	return rsp.TotalIps - rsp.UsedIps, nil
}

// GetZoneNameByZoneId 通过zoneId获取可用区名称
func GetZoneNameByZoneId(region, zoneId string) int {
	z := strings.TrimPrefix(zoneId, region)
	if len(z) != 1 {
		return -1
	}

	zoneNum := letterToNum(z)
	return zoneNum[0]
}

func letterToNum(s string) []int {
	var result []int
	for _, char := range s {
		if char >= 'a' && char <= 'z' {
			num := int(char - 'a' + 1)
			result = append(result, num)
		}
	}
	return result
}
