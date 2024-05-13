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

package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/printer"
	cloudAccountMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	listCloudAccountToPermExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager list cloudAccountToPerm`))
)

func newListCloudAccountToPermCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloudAccountToPerm",
		Short:   "list cloud account to perm from bcs-cluster-manager",
		Example: listCloudAccountToPermExample,
		Run:     listCloudAccountToPerm,
	}

	return cmd
}

func listCloudAccountToPerm(cmd *cobra.Command, args []string) {
	resp, err := cloudAccountMgr.New(context.Background()).ListToPerm(types.ListCloudAccountToPermReq{})
	if err != nil {
		klog.Fatalf("list cloud account to perm failed: %v", err)
	}

	header := []string{"COUND_ID", "PROJECT_ID", "ACCOUNT_ID", "ACCOUNT_NAME", "ENABLE", "CREATOR", "CREAT_TIME"}
	data := make([][]string, len(resp.Data))
	for key, item := range resp.Data {
		data[key] = []string{
			item.CloudID,
			item.ProjectID,
			item.AccountID,
			item.AccountName,
			fmt.Sprintf("%t", item.Enable),
			item.Creator,
			item.CreatTime,
		}
	}

	err = printer.PrintList(flagOutput, resp.Data, header, data)
	if err != nil {
		klog.Fatalf("list cloud account to perm failed: %v", err)
	}
}
