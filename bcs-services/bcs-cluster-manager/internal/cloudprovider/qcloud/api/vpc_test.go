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
	"fmt"
	"os"
	"testing"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func getVpcClient(region string) *VPCClient {
	cli, err := NewVPCClient(&cloudprovider.CommonOption{
		Account: &cmproto.Account{
			SecretID:  os.Getenv(TencentCloudSecretIDClusterEnv),
			SecretKey: os.Getenv(TencentCloudSecretKeyClusterEnv),
		},
		Region: region,
		CommonConf: cloudprovider.CloudConf{
			CloudInternalEnable: true,
			VpcDomain:           "vpc.internal.tencentcloudapi.com",
		},
	})
	if err != nil {
		panic(err)
	}
	return cli
}

func TestDescribeSecurityGroups(t *testing.T) {
	cli := getVpcClient("ap-nanjing")
	sg, err := cli.DescribeSecurityGroups(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("sg: %s, count: %d", utils.ToJSONString(sg), len(sg))
}

func TestDescribeSubnets(t *testing.T) {
	cli := getVpcClient("ap-nanjing")

	filter := make([]*Filter, 0)
	filter = append(filter, &Filter{Name: "vpc-id", Values: []string{"vpc-xxx"}})

	subnets, err := cli.DescribeSubnets(nil, filter)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("subnets: %s, count: %d", utils.ToJSONString(subnets), len(subnets))

	fmt.Println()

	for i := range subnets {
		if utils.MatchSubnet(*subnets[i].SubnetName, "ap-nanjing") {
			t.Log(*subnets[i].VpcID, *subnets[i].SubnetID, *subnets[i].SubnetName, *subnets[i].AvailableIPAddressCount)
		}
	}
}

func TestVPCClient_DescribeNetworkAccountType(t *testing.T) {
	cli := getVpcClient("ap-nanjing")
	account, err := cli.GetCloudNetworkAccountType(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(account)
}
