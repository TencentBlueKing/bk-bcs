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

// Package cmd xxx
package cmd

import (
	"context"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
)

// newApplyCmd create new apply command
func newApplyCmd() *cobra.Command {
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "apply tkecidrs",
		Long:  "apply tkecidrs from bcs-user-manager",
	}
	applyCmd.AddCommand(applyTkeCidrCmd())
	return applyCmd
}

func applyTkeCidrCmd() *cobra.Command {
	var tkeCidrForm string
	subCmd := &cobra.Command{
		Use:     "tkecidrs",
		Aliases: []string{"tkecidrs"},
		Short:   "apply tke cidrs",
		Long:    "apply tke cidrs from user manager",
		Example: "kubectl-bcs-user-manager apply tkecidrs --tkecidr_form '{\"vpc\":\"\",\"cluster\":\"\", \"ip_number\":}' ",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.ApplyTkeCidr(tkeCidrForm)
			if err != nil {
				klog.Fatalf("apply tke cidrs failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("apply tke cidrs response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintApplyTkeCidrCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&tkeCidrForm, "tkecidr_form", "f", "",
		"the form json used to apply tke cidrs")
	return subCmd
}
