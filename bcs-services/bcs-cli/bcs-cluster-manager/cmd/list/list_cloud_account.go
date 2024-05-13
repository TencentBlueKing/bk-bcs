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
	"strings"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/printer"
	cloudAccountMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	listCloudAccountExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager list cloudAccount`))
)

func newListCloudAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloudAccount",
		Short:   "list cloud account from bcs-cluster-manager",
		Example: listCloudAccountExample,
		Run:     listCloudAccount,
	}

	return cmd
}

func listCloudAccount(cmd *cobra.Command, args []string) {
	resp, err := cloudAccountMgr.New(context.Background()).List(types.ListCloudAccountReq{})
	if err != nil {
		klog.Fatalf("list cloud account failed: %v", err)
	}

	header := []string{"ACCOUNT_ID", "ACCOUNT_NAME", "PROJECT_ID", "DESC", "Clusters"}
	data := make([][]string, len(resp.Data))
	for key, item := range resp.Data {
		data[key] = []string{
			item.AccountID,
			item.AccountName,
			item.ProjectID,
			item.Desc,
			strings.Join(item.Clusters, "\n"),
		}
	}

	err = printer.PrintList(flagOutput, resp.Data, header, data)
	if err != nil {
		klog.Fatalf("list cloud account to perm failed: %v", err)
	}
}
