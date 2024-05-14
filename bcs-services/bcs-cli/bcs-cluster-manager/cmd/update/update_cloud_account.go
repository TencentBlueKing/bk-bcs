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

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	cloudAccountMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	updateCloudAccountExample = templates.Examples(i18n.T(`update cloud account from json file. file template:
	{"accountID":"BCS-tencentCloud-xxx","cloudID":"tencentCloud","accountName":"test001","desc":"腾讯云测试账号001",
	"account":{"secretID":"xxxxxxxxxxxxxxxxxx","secretKey":"xxxxxxxxxxxxxxxxxx"},"enable":true,"creator":"bcs",
	"projectID":"b363e23b1b354928a0f3xxxxxx"}`))
)

func newUpdateCloudAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloudAccount",
		Short:   "update cloud account from bcs-cluster-manager",
		Example: updateCloudAccountExample,
		Run:     updateCloudAccount,
	}

	return cmd
}

func updateCloudAccount(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.UpdateCloudAccountReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = cloudAccountMgr.New(context.Background()).Update(req)
	if err != nil {
		klog.Fatalf("update cloud account failed: %v", err)
	}

	fmt.Println("update cloud account succeed")
}
