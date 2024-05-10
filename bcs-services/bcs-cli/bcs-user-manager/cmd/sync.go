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

package cmd

import (
	"context"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
)

// newSyncCmd create the sync tke cidrs command
func newSyncCmd() *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "sync tkecidrs",
		Long:  "sync tkecidrs from bcs-user-manager",
	}
	syncCmd.AddCommand(syncTkeClusterCredentialsCmd())
	return syncCmd
}

func syncTkeClusterCredentialsCmd() *cobra.Command {
	var clusterId string
	subCmd := &cobra.Command{
		Use:     "tkecidrs",
		Aliases: []string{"tkecidrs"},
		Short:   "sync tke cidrs",
		Long:    "sync tke cidrs from user manager",
		Example: "kubectl-bcs-user-manager sync tkecidrs --cluster_id [cluster_id]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.SyncTkeClusterCredentials(clusterId)
			if err != nil {
				klog.Fatalf("sync the tke cluster credentials from tke failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("sync the tke cluster credentials from tke response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintSyncCredentialsResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&clusterId, "cluster_id", "i", "",
		"the cluster_id used to sync the tke cluster credentials from tke")
	return subCmd
}
