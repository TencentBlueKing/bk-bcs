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

package delete

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	cloudvpcMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_vpc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	deleteCloudVPCExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager delete cloudVPC --cloudID xxx --vpcID xxx`))
)

func newDeleteCloudVPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloudVPC",
		Short:   "delete cloud vpc from bcs-cluster-manager",
		Example: deleteCloudVPCExample,
		Run:     deleteCloudVPC,
	}

	cmd.Flags().StringVarP(&cloudID, "cloudID", "c", "", `cloud ID (required)`)
	_ = cmd.MarkFlagRequired("cloudID")
	cmd.Flags().StringVarP(&vpcID, "vpcID", "", "", `VPC ID (required)`)
	_ = cmd.MarkFlagRequired("vpcID")

	return cmd
}

func deleteCloudVPC(cmd *cobra.Command, args []string) {
	err := cloudvpcMgr.New(context.Background()).Delete(types.DeleteCloudVPCReq{
		CloudID: cloudID,
		VPCID:   vpcID,
	})
	if err != nil {
		klog.Fatalf("delete cloud vpc failed: %v", err)
	}

	fmt.Println("delete cloud vpc succeed")
}
