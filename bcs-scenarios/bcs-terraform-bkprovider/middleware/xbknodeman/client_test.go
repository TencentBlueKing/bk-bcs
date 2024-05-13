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

package xbknodeman

import (
	"context"
	"log"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
)

func NewTestClient() *Client {
	bkAppCode := ""
	bkAppSecret := ""
	bkUserName := ""
	return NewClient(0, "", bkAppCode, bkAppSecret, "", bkUserName)
}

func TestListCloud(t *testing.T) {
	client := NewTestClient()
	resp, err := client.ListCloud(context.Background(), &ListCloudRequest{})
	if err != nil {
		log.Fatal(err)
	}

	println(common.JsonMarshal(resp))
}

func TestCreateCloud(t *testing.T) {
	client := NewTestClient()
	resp, err := client.CreateCloud(context.TODO(), &CreateCloudRequest{
		BkCloudName: "porterlin-test-2",
		Isp:         "Tencent",
		ApID:        1,
	})
	if err != nil {
		log.Fatal(err)
	}

	println(common.JsonMarshal(resp))

	if _, err := client.DeleteCloud(context.TODO(), &DeleteCloudRequest{BkCloudID: resp.Data.BkCloudID}); err != nil {
		log.Fatal(err)
	}
}

func TestGetProxy(t *testing.T) {
	client := NewTestClient()
	resp, err := client.GetProxyHost(context.TODO(), &GetProxyHostRequest{BkCloudId: 400})
	if err != nil {
		log.Fatal(err.Error())
	}

	println(common.JsonMarshal(resp))
}

func TestListHost(t *testing.T) {
	client := NewTestClient()
	resp, err := client.ListHosts(context.TODO(), &ListHostRequest{
		Page:     1,
		PageSize: 50,
		// Conditions: []Condition{
		// 	{
		// 		Key:   "bk_cloud_id",
		// 		Value: []int{30000322},
		// 	},
		// },
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	println(common.JsonMarshal(resp))
}
