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

package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func newListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list infos from bcs-cluster-manager",
		Long:  "",
	}
	listCmd.AddCommand(listClusterCmd())
	return listCmd
}

func listClusterCmd() *cobra.Command {
	request := new(clustermanager.ListClusterReq)
	subCmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters", "c"},
		Short:   "list clusters info from bcs-cluster-manager",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
			if err != nil {
				klog.Fatalf("init client failed: %v", err.Error())
			}
			resp, err := client.ListCluster(cliCtx, request)
			if err != nil {
				klog.Fatalf("list projects failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatal("list projects response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintClustersListInTable(flagOutput, resp)
		},
	}
	subCmd.PersistentFlags().StringVarP(&request.ClusterName, "cluster_name", "n", "",
		"the name of cluster")
	return subCmd
}
