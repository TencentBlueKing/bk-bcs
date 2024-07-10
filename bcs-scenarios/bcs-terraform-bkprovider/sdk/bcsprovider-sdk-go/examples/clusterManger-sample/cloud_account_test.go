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
	"context"
	"log"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/clusterManger"
)

func Test_CreateCloudAccount(t *testing.T) {
	req := &clusterManger.CreateCloudAccountRequest{
		CloudID:     clusterManger.TencentCloud,
		AccountName: "my-aksk",
		Account: &clusterManger.Account{
			SecretID:  "xxxxxxxxxxx-xxxxxxxx",
			SecretKey: "yyyyy-yyyyyyyyyyyyyy",
		},
		ProjectID: projectID,
		Desc:      "xxxxx",
	}

	resp, err := service.CreateCloudAccount(context.TODO(), req)
	if err != nil {
		t.Fatalf("craete cloud account failed, err: %s", err.Error())
	}

	log.Printf("create cloud account success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_DeleteCloudAccount(t *testing.T) {
	req := &clusterManger.DeleteCloudAccountRequest{
		CloudID:   clusterManger.TencentCloud,
		AccountID: accountID,
	}

	resp, err := service.DeleteCloudAccount(context.TODO(), req)
	if err != nil {
		t.Fatalf("delete cloud account failed, err: %s", err.Error())
	}

	log.Printf("delete cloud account success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_UpdateCloudAccount(t *testing.T) {
	req := &clusterManger.UpdateCloudAccountRequest{
		CloudID:     clusterManger.TencentCloud,
		AccountName: "my-11111-test",
		AccountID:   accountID,
		Account: &clusterManger.Account{
			SecretID:  "xxxxxxxxxxx-xxxxxxxx",
			SecretKey: "yyyyy-yyyyyyyyyyyyyy",
		},
		ProjectID: projectID,
		Desc:      "test account",
	}

	resp, err := service.UpdateCloudAccount(context.TODO(), req)
	if err != nil {
		t.Fatalf("update cloud account failed, err: %s", err.Error())
	}

	log.Printf("update cloud account success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListCloudAccount(t *testing.T) {
	req := &clusterManger.ListCloudAccountRequest{
		CloudID:   clusterManger.TencentCloud,
		ProjectID: projectID,
	}

	resp, err := service.ListCloudAccount(context.TODO(), req)
	if err != nil {
		// 常见错误：http: server gave HTTP response to HTTPS client
		t.Fatalf("list cloud account failed, err: %s", err.Error())
	}

	log.Printf("list cloud account success. resp: %s", utils.ObjToPrettyJson(resp))
}
