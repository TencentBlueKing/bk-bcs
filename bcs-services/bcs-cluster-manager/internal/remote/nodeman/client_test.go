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

package nodeman

import (
	"context"
	"testing"
)

/*
AppCode:   "bk_bcs_app",
	AppSecret: "95349c90-dfb3-4e4e-8710-a61cbd176630",
	Server:    "https://bkapi.bk-tenant-dev.woa.com/api/bk-user/prod",
*/

// NewClient xxx
func NewClient() *Client {
	cli, _ := NewNodeManClient(Options{
		AppCode:    "bk_bcs_app",
		AppSecret:  "95349c90-dfb3-4e4e-8710-a61cbd176630",
		BKUserName: "xxx",
		Server:     "https://bkapi.bk-tenant-dev.woa.com/api/bk-nodeman/prod",
		Debug:      true,
	})

	return cli
}

func TestClient_CloudList(t *testing.T) {
	cli := NewClient()
	cloudAreas, err := cli.CloudList(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(cloudAreas)
}
