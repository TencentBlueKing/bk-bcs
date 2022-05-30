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
)

func getASClient(region string) *ASClient {
	cli, err := NewASClient(&cloudprovider.CommonOption{
		Key:    "xxx",
		Secret: "xxx",
		Region: region,
	})
	if err != nil {
		panic(err)
	}
	return cli
}

func TestDescribeAutoScalingInstances(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	ins, err := cli.DescribeAutoScalingInstances("asg-xxx")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ins: %s, count: %d", utils.ToJSONString(ins), len(ins))
}

func TestRemoveInstances(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	err := cli.RemoveInstances("asg-xxx", []string{"ins-xxx"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDetachInstances(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	err := cli.DetachInstances("asg-xxx", []string{"ins-xxx"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestModifyDesiredCapacity(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	err := cli.ModifyDesiredCapacity("asg-xxx", 3)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDescribeLaunchConfigurations(t *testing.T) {
	cli := getASClient("ap-guangzhou")
	asc, err := cli.DescribeLaunchConfigurations([]string{"asc-xxx"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.ToJSONString(asc))
}
