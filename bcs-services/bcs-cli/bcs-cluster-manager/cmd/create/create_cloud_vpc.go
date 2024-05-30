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

package create

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	cloudvpcMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_vpc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	createCloudVPCExample = templates.Examples(i18n.T(`create cloud vpc from json file.file template: 
	{"cloudID":"tencentCloud","networkType":"overlay","region":"ap-guangzhou","regionName":"广州",
	"vpcName":"vpc-xxxxxxx-1","vpcID":"vpc-123456789","available":"true","extra":"","creator":"bcs"}`))
)

func newCreateCloudVPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloudVPC",
		Short:   "create cloud vpc from bcs-cluster-manager",
		Example: createCloudVPCExample,
		Run:     createCloudVPC,
	}

	return cmd
}

func createCloudVPC(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.CreateCloudVPCReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = cloudvpcMgr.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create cloud vpc failed: %v", err)
	}

	fmt.Println("create cloud vpc succeed")
}
