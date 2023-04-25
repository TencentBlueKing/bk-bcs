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

package cloudaccount

import (
	"context"
	"fmt"

	cloudAccountMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete cloud account from bcs-cluster-manager",
		Run:   delete,
	}

	cmd.Flags().StringVarP(&cloudID, "cloudID", "c", "", `cloud ID`)
	cmd.MarkFlagRequired("cloudID")
	cmd.Flags().StringVarP(&accountID, "accountID", "a", "", `account ID`)
	cmd.MarkFlagRequired("accountID")

	return cmd
}

func delete(cmd *cobra.Command, args []string) {
	err := cloudAccountMgr.New(context.Background()).Delete(types.DeleteCloudAccountReq{
		CloudID:   cloudID,
		AccountID: accountID,
	})
	if err != nil {
		klog.Fatalf("delete cloud account failed: %v", err)
	}

	fmt.Println("delete cloud account succeed")
}
