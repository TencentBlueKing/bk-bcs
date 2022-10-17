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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
)

func newCreateCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "create",
		Short: "create resource from bcs-user-manager",
		Long:  "",
	}
	listCmd.AddCommand(createClusterCmd())
	return listCmd
}

func createClusterCmd() *cobra.Command {
	var clusterCreateBody string
	subCmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"c"},
		Short:   "create cluster from user manager",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreateCluster(clusterCreateBody)
			if err != nil {
				klog.Fatalf("create cluster failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("create cluster response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintClusterListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&clusterCreateBody, "cluster-body", "b", "",
		"the cluster body that create cluster")
	return subCmd
}
