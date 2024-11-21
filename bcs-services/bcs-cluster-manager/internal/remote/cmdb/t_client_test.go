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

package cmdb

import (
	"fmt"
	"testing"
)

var (
	appId  = "xxx"
	appKey = "xxx"
	server = "xxx"
)

func newClient() *TClient {
	client, err := NewTCmdbClient(TOptions{
		Enable: true,
		AppId:  appId,
		AppKey: appKey,
		Server: server,
		Debug:  true,
	})

	if err != nil {
		return nil
	}

	return client
}

func TestQueryBusinessLevel2DetailInfo(t *testing.T) {
	client := newClient()
	if client == nil {
		t.Errorf("new tcmdb client failed")
		return
	}

	data, err := client.QueryBusinessLevel2DetailInfo(725545)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(data)
}

func TestQueryServerInfoByIps(t *testing.T) {
	client := newClient()
	if client == nil {
		t.Errorf("new tcmdb client failed")
		return
	}

	ips := []string{""}

	data, err := client.queryServerInfoByIps(ips)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(data)
}

func TestGetAssetIdsByIps(t *testing.T) {
	client := newClient()
	if client == nil {
		t.Errorf("new tcmdb client failed")
		return
	}

	ips := []string{""}
	data, err := client.GetAssetIdsByIps(ips)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(data)
	fmt.Println(len(data))
}
