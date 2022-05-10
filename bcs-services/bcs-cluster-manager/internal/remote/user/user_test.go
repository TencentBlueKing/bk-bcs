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

package user

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func getUserManagerClient() UserManager {
	return NewUserManagerClient(&Options{
		Enable:  true,
		GateWay: "https://xxx/bcsapi/v4/",
		Token:   "xxx",
	})
}

var cli = getUserManagerClient()

func TestUser_CreateUserToken(t *testing.T) {
	token, err := cli.CreateUserToken(CreateTokenReq{
		Username:   "xxx",
		Expiration: -1,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(token)
}

func TestUser_GetUserToken(t *testing.T) {
	token, err := cli.GetUserToken("BCS-K8S-40025")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(token)
}

func TestUser_DeleteUserToken(t *testing.T) {
	err := cli.DeleteUserToken("xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestUser_GrantUserPermission(t *testing.T) {
	err := cli.GrantUserPermission([]types.Permission{
		types.Permission{
			UserName:     "xxx",
			ResourceType: ResourceTypeClusterManager,
			Resource:     "BCS-K8S-15202",
			Role:         PermissionViewerRole,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestUser_RevokeUserPermission(t *testing.T) {
	err := cli.RevokeUserPermission([]types.Permission{
		types.Permission{
			UserName:     "xxx",
			ResourceType: ResourceTypeClusterManager,
			Resource:     "BCS-K8S-15202",
			Role:         PermissionViewerRole,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestUser_VerifyUserPermission(t *testing.T) {

}
