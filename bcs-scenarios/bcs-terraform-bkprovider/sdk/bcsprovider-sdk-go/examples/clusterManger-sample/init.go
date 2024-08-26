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

// Package clusterMangerSample 测试
package clusterMangerSample

import (
	"fmt"
	"log"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/clusterManger"
)

const (
	// username rtx
	username = ""

	// token 个人token
	token = ""

	// gatewayApi xxx
	gatewayApi = "" // 此处结尾多一个"/"，少一个"/"不影响

	// projectID 项目id
	projectID = ""

	// region 地域
	region = ""

	// accountID 地域
	accountID = ""

	// vpcID vpc
	vpcID = ""

	// clusterID cluster ID
	clusterID = ""

	// nodeGroupID node group id
	nodeGroupID = ""

	// ngPasswordCase password case
	ngPasswordCase = ""

	// securityGroupID security group id
	securityGroupID = ""

	// ngNameCase test case
	ngNameCase = ""

	// clsIdCase test case
	clsIdCase = ""
)

var (
	gClient sdk.Client

	service clusterManger.Service
)

func init() {
	config := &options.Config{
		Username:       username,
		Token:          token,
		BcsGatewayAddr: gatewayApi,
	}

	client, err := sdk.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("new sdk client err: %s", err.Error()))
	}
	gClient = client

	service = gClient.ClusterManger()

	log.Printf("config: %s", utils.ObjToPrettyJson(client.Config()))
}
