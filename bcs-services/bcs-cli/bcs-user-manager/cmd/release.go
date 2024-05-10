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

// newReleaseCmd create the release tke cidrs command
func newReleaseCmd() *cobra.Command {
	releaseCmd := &cobra.Command{
		Use:   "release",
		Short: "release tkecidrs",
		Long:  "release tkecidrs from bcs-user-manager",
	}
	releaseCmd.AddCommand(releaseTkeCidrCmd())
	return releaseCmd
}

func releaseTkeCidrCmd() *cobra.Command {
	var tkeCidrForm string
	subCmd := &cobra.Command{
		Use:     "tkecidrs",
		Aliases: []string{"tkecidrs"},
		Short:   "release tke cidrs",
		Long:    "release tke cidrs from user manager",
		Example: "kubectl-bcs-user-manager release tkecidrs --tkecidr_form '{\"vpc\":\"\",\"cidr\":\"\",\"cluster\":\"\"}' ",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.ReleaseTkeCidr(tkeCidrForm)
			if err != nil {
				klog.Fatalf("release tke cidrs failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("release tke cidrs response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintReleaseTkeCidrCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&tkeCidrForm, "tkecidr_form", "f", "",
		"the form used to release tke cidrs")
	return subCmd
}
