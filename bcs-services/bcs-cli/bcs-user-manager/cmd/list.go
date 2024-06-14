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

// newListCmd create the resource list command
func newListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list resource",
		Long:  "list resource from bcs-user-manager",
	}
	listCmd.AddCommand(listCredentialsCmd())
	listCmd.AddCommand(listTkeCidrCmd())
	return listCmd
}

func listCredentialsCmd() *cobra.Command {
	subCmd := &cobra.Command{
		Use:     "credentials",
		Aliases: []string{"c"},
		Short:   "list credentials",
		Long:    "list all cluster credentials",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.ListCredentials()
			if err != nil {
				klog.Fatalf("get credential according cluster ID failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get credential according cluster ID response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintListCredentialsCmdResult(flagOutput, resp)
		},
	}

	return subCmd
}

func listTkeCidrCmd() *cobra.Command {
	subCmd := &cobra.Command{
		Use:     "tkecidrs",
		Aliases: []string{"tkecidrs"},
		Short:   "list tke cidrs",
		Long:    "list tke cidrs from user manager",
		Example: "kubectl-bcs-user-manager list tkecidrs",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.ListTkeCidr()
			if err != nil {
				klog.Fatalf("list tke cidrs failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("list tke cidrs response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintListTkeCidrCmdResult(flagOutput, resp)
		},
	}

	return subCmd
}
