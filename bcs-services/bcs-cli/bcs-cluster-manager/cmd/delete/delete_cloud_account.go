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

	cloudAccountMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	deleteCloudAccountExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager delete cloudAccount --cloudID xxx --accountID xxx`))
)

func newDeleteCloudAccoundCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloudAccount",
		Short:   "delete cloud account from bcs-cluster-manager",
		Example: deleteCloudAccountExample,
		Run:     deleteCloudAccount,
	}

	cmd.Flags().StringVarP(&cloudID, "cloudID", "c", "", `cloud ID`)
	_ = cmd.MarkFlagRequired("cloudID")
	cmd.Flags().StringVarP(&accountID, "accountID", "a", "", `account ID`)
	_ = cmd.MarkFlagRequired("accountID")

	return cmd
}

func deleteCloudAccount(cmd *cobra.Command, args []string) {
	err := cloudAccountMgr.New(context.Background()).Delete(types.DeleteCloudAccountReq{
		CloudID:   cloudID,
		AccountID: accountID,
	})
	if err != nil {
		klog.Fatalf("delete cloud account failed: %v", err)
	}

	fmt.Println("delete cloud account succeed")
}
