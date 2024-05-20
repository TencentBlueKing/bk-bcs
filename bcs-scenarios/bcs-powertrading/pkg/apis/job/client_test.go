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

package job

import (
	"fmt"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

func TestClient_GetBatchJobLog(t *testing.T) {
	testCli := New(&apis.ClientOptions{
		Endpoint:    "",
		UserName:    "",
		AccessToken: "",
		AppCode:     "",
		AppSecret:   "",
	}, requester.NewRequester())
	ips := make([]BatchLogIPRequest, 0)
	ips = append(ips, BatchLogIPRequest{
		BkCloudID: 0,
		IP:        "",
	})
	rsp, err := testCli.GetBatchJobLog("biz", "", "",
		"", ips)
	fmt.Println(err)
	fmt.Println(rsp)
}
