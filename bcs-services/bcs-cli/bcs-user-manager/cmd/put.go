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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newPutCmd() *cobra.Command {
	putCmd := &cobra.Command{
		Use:   "put",
		Short: "put resource from bcs-user-manager",
		Long:  "",
	}
	putCmd.AddCommand(refreshSaasTokenCmd())
	putCmd.AddCommand(refreshPlainTokenCmd())
	putCmd.AddCommand(UpdateCredentialsCmd())
	return putCmd
}

func refreshSaasTokenCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "saas-user",
		Aliases: []string{"au"},
		Short:   "get saas user from user manager",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.RefreshSaasToken(userName)
			if err != nil {
				klog.Fatalf("get saas user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get saas user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			//printer.PrintSaasUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "n", "",
		"the user name that refresh user token for a plain user")
	return subCmd
}

func refreshPlainTokenCmd() *cobra.Command {
	var userName, expireTime string
	subCmd := &cobra.Command{
		Use:     "saas-user",
		Aliases: []string{"au"},
		Short:   "get saas user from user manager",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.RefreshPlainToken(userName, expireTime)
			if err != nil {
				klog.Fatalf("get saas user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get saas user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			//printer.PrintSaasUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "n", "",
		"the user name that refresh user token for a plain user")
	subCmd.PersistentFlags().StringVarP(&expireTime, "expire_time", "e", "",
		"the expire time that refresh user token for a plain user")

	return subCmd
}

func UpdateCredentialsCmd() *cobra.Command {
	var clusterId string
	subCmd := &cobra.Command{
		Use:     "credentials",
		Aliases: []string{"c"},
		Short:   "update credentials",
		Long:    "update cluster credentials according cluster ID",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetRegisterToken(clusterId)
			if err != nil {
				klog.Fatalf("update credential according cluster ID failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("update credential according cluster ID response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			//printer.PrintAdminUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&clusterId, "cluster_id", "c", "",
		"the cluster_id for update cluster credential")
	return subCmd
}
