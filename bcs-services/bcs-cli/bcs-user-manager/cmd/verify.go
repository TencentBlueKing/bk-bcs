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

func newVerifyCmd() *cobra.Command {
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "verify permission",
		Long:  "verify permissions from bcs-user-manager",
	}
	verifyCmd.AddCommand(verifyPermissionCmd())
	verifyCmd.AddCommand(verifyPermissionV2Cmd())
	return verifyCmd
}

func verifyPermissionCmd() *cobra.Command {
	var verifyPermissionForm string
	subCmd := &cobra.Command{
		Use:     "permissions",
		Aliases: []string{"permissions", "ps"},
		Short:   "verify permission",
		Long:    "verify permissions from user manager",
		Example: "kubectl-bcs-user-manager verify permissions --form '{\"user_token\":\"\",\"resource_type\":\"\",\"resource\":\"\",\"action\":\"\"}' ",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.VerifyPermission(verifyPermissionForm)
			if err != nil {
				klog.Fatalf("verify permissions failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("verify permissions response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintVerifyPermissionCmdResult(flagOutput, resp)
		},
	}
	subCmd.Flags().StringVarP(&verifyPermissionForm, "form", "f", "",
		"the form used to verfiy permissions")

	return subCmd
}

func verifyPermissionV2Cmd() *cobra.Command {
	var verifyPermissionForm string
	subCmd := &cobra.Command{
		Use:     "permissionsv2",
		Aliases: []string{"permissionsv2", "psv2"},
		Short:   "verify permission v2",
		Long:    "verify permissions v2 from user manager",
		Example: "kubectl-bcs-user-manager verify permissionsv2 --form {\"user_token\":\"\",\"resource_type\":\"\",\"cluster_type\":\"\",\"cluster_type\":\"\",\"project_id\":\"\",\"cluster_id\":\"\",\"request_url\":\"\",\"resource\":\"\",\"action\":\"\"}",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.VerifyPermissionV2(verifyPermissionForm)
			if err != nil {
				klog.Fatalf("verify permissions failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("verify permissions response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintVerifyPermissionV2CmdResult(flagOutput, resp)
		},
	}
	subCmd.Flags().StringVarP(&verifyPermissionForm, "form", "f", "",
		"the form used to verfiy permissions with version 2")

	return subCmd
}
