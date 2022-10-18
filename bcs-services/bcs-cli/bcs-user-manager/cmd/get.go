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

func newGetCmd() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "get infos from bcs-user-manager",
		Long:  "",
	}
	getCmd.AddCommand(getAdminUserCmd())
	getCmd.AddCommand(getSaasUserCmd())
	return getCmd
}

func getAdminUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "admin-user",
		Aliases: []string{"au"},
		Short:   "get admin user from user manager",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetAdminUser(userName)
			if err != nil {
				klog.Fatalf("get admin user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get admin user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintAdminUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "n", "",
		"the user name that query admin user")
	return subCmd
}

func getSaasUserCmd() *cobra.Command {
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
			resp, err := client.GetSaasUser(userName)
			if err != nil {
				klog.Fatalf("get saas user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get saas user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintSaasUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "n", "",
		"the user name that query user user")
	return subCmd
}
