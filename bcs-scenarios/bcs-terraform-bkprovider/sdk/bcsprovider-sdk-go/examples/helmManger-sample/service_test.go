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

// Package helmMangerSample 测试
package helmMangerSample

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/helmManger"
)

const (
	// username rtx
	username = ""

	// token 个人token
	token = ""

	// gatewayApi xxx
	gatewayApi = "" // 此处结尾多一个"/"，少一个"/"不影响

	// clusterID 集群id
	clusterID = ""

	// projectID 项目id
	projectID = ""

	// bcsIngress ingress
	bcsIngress = "bcs-ingress-controller"
)

var (
	gClient sdk.Client

	service helmManger.Service
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

	service = gClient.HelmManger()

	log.Printf("config: %s", utils.ObjToPrettyJson(client.Config()))
}

// Test_Install 安装集群组件
func Test_Install(t *testing.T) {
	// values.yaml中si、sk需要自行补充，仅作为使用样例
	values := `
# tencentcloud, aws
cloud: tencentcloud

# tencent cloud setting
# tencent cloud clb API v3 domain
tencentcloudClbDomain: clb.internal.tencentcloudapi.com
# tencent cloud default region
tencentcloudRegion: ap-shanghai
# tencent cloud AccessID after Base64 encoding.
# echo -n "<SECRET_ID>" | base64
tencentcloudSecretID: xxxxxxxxxxxxxxxx
# tencent cloud AccessKeyafter Base64 encoding.
# echo -n "<SECRET_ID>" | base64
tencentcloudSecretKey: xxxxxxxxxxxxxxx

# aws setting
# aws default region
awsRegion: "us-west-1"
# aws AccessID after Base64 encoding.
# echo -n "<SECRET_ID>" | base64
awsSecretID: xxxxxxxxxxxxxxxx
# aws AccessKeyafter Base64 encoding.
# echo -n "<SECRET_ID>" | base64
awsSecretKey: xxxxxxxxxxxxxxx
serviceMonitor:
  enabled: false
`

	req := &helmManger.InstallAddonsRequest{
		ProjectID: projectID,
		ClusterID: clusterID,
		Name:      bcsIngress,
		Version:   "1.29.0-alpha.59",
		Values:    values,
	}

	resp, err := service.InstallAddons(context.TODO(), req)
	if err != nil {
		t.Fatalf("install failed, err: %s", err.Error())
	}

	log.Printf("install success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_Uninstall 卸载集群组件
func Test_Uninstall(t *testing.T) {
	req := &helmManger.UninstallAddonsRequest{
		ProjectID: projectID,
		ClusterID: clusterID,
		Name:      bcsIngress,
	}

	resp, err := service.UninstallAddons(context.TODO(), req)
	if err != nil {
		t.Fatalf("remove failed, err: %s", err.Error())
	}

	log.Printf("remove success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_Upgrade 升级或更新集群组件
func Test_Upgrade(t *testing.T) {
	// 参数更新
	values := `
# tencentcloud, aws
cloud: tencentcloud

# tencent cloud setting
# tencent cloud clb API v3 domain
tencentcloudClbDomain: clb.internal.tencentcloudapi.com
# tencent cloud default region
tencentcloudRegion: ap-shanghai
# tencent cloud AccessID after Base64 encoding.
# echo -n "<SECRET_ID>" | base64
tencentcloudSecretID: yyyyyyyxxxxxxx1111111111
# tencent cloud AccessKeyafter Base64 encoding.
# echo -n "<SECRET_ID>" | base64
tencentcloudSecretKey: yyyyyyyxxxxxxx1111111111

# aws setting
# aws default region
awsRegion: "us-west-1"
# aws AccessID after Base64 encoding.
# echo -n "<SECRET_ID>" | base64
awsSecretID: yyyyyyyxxxxxxx11111111
# aws AccessKeyafter Base64 encoding.
# echo -n "<SECRET_ID>" | base64
awsSecretKey: yyyyyyyxxxxxxx111111111
serviceMonitor:
  enabled: false
`

	req := &helmManger.UpgradeAddonsRequest{
		ProjectID: projectID,
		ClusterID: clusterID,
		Name:      bcsIngress,
		Version:   "1.29.0-alpha.181", // 升级版本
		Values:    values,
	}

	resp, err := service.UpgradeAddons(context.TODO(), req)
	if err != nil {
		t.Fatalf("update failed, err: %s", err.Error())
	}

	log.Printf("update success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_Get 查询集群组件详情
func Test_Get(t *testing.T) {
	req := &helmManger.GetAddonsDetailRequest{
		ProjectID: projectID,
		ClusterID: clusterID,
		Name:      "bcs-k8s-watch", // 查询bcs-ingress使用详情
	}

	resp, err := service.GetAddonsDetail(context.TODO(), req)
	if err != nil {
		t.Fatalf("get '%s' failed, err: %s", bcsIngress, err.Error())
	}

	log.Printf("get '%s' success. resp: %s", bcsIngress, utils.ObjToPrettyJson(resp))
}

// Test_List 查询集群组件列表
func Test_List(t *testing.T) {
	req := &helmManger.ListAddonsRequest{
		ProjectID: projectID,
		ClusterID: clusterID,
	}

	resp, err := service.ListAddons(context.TODO(), req)
	if err != nil {
		t.Fatalf("list '%s' failed, err: %s", clusterID, err.Error())
	}

	log.Printf("list '%s' success. resp: %s", clusterID, utils.ObjToPrettyJson(resp))
}
