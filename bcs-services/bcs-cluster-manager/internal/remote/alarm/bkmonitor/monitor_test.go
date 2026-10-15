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

package bkmonitor

import (
	"context"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
)

func getClient() *Client {
	cli, err := NewClient(Options{
		AppCode:   "xxx",
		AppSecret: "xxx",
		Enable:    true,
		Server:    "xxx",
		Debug:     true,
	})
	if err != nil {
		return nil
	}

	return cli
}

func getPermServer() *auth.ClientAuth { // nolint
	cli := auth.NewAccessClient(auth.Options{
		Server: "xxx",
		Debug:  true,
	})

	return cli
}

func TestClient_ShieldHostAlarmConfig(t *testing.T) {
	cli := getClient()

	// user 必须是该业务的运维人员,通过该身份屏蔽主机告警
	err := cli.ShieldHostAlarmConfig(context.Background(), "", &alarm.ShieldHost{
		BizID: "xxx",
		HostList: []alarm.HostInfo{
			{
				IP:      "xxx",
				CloudID: 0,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("successful")
}
