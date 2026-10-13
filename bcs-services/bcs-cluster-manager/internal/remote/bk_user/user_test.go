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

package bkuser

import (
	"context"
	"testing"
)

var opts = Options{
	AppCode:   "xxx",
	AppSecret: "xxx",
	Server:    "https://bkapi.example.com/api/bk-user/prod",
	Debug:     true,
}

func getBkUserClient() (*Client, error) {
	cli, err := NewBkUserClient(opts)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func TestQueryUserInfoByTenantLoginName(t *testing.T) {
	cli, err := getBkUserClient()
	if err != nil {
		t.Errorf("get bkUser client failed, err: %v", err)
		return
	}

	var (
		tenantId  = "putongoa"
		loginName = "evanxinli"
	)

	data, err := cli.QueryUserInfoByTenantLoginName(context.Background(), tenantId, loginName)
	if err != nil {
		t.Errorf("QueryUserInfoByTenantLoginName failed, err: %v", err)
		return
	}

	t.Logf("QueryUserInfoByTenantLoginName rsp: %v", data)
}
