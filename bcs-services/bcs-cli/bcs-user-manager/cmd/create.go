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

// newCreateCmd create the resource create command
func newCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create resource",
		Long:  "create resource from bcs-user-manager",
	}
	createCmd.AddCommand(createAdminUserCmd())
	createCmd.AddCommand(createPlainUserCmd())
	createCmd.AddCommand(createTokenCmd())
	createCmd.AddCommand(createTempTokenCmd())
	createCmd.AddCommand(createClientTokenCmd())
	return createCmd
}

// createAdminUserCmd fro admin user
func createAdminUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "admin-user",
		Aliases: []string{"au"},
		Short:   "create admin user ",
		Long:    "create admin user from user manager",
		Example: "kubectl-bcs-user-manager create au -u [user_name to create]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreateAdminUser(userName)
			if err != nil {
				klog.Fatalf("create admin user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("create admin user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintCreateAdminUserCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that query admin user")
	return subCmd
}

// createPlainUserCmd for plain user
func createPlainUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "plain-user",
		Aliases: []string{"pu"},
		Short:   "create plain",
		Long:    "create plain user from user manager",
		Example: "kubectl-bcs-user-manager create pu -u [user_name to create]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreatePlainUser(userName)
			if err != nil {
				klog.Fatalf("create plain user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("create plain user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintCreatePlainUserCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that query plain user")
	return subCmd
}

// createTokenCmd create token
func createTokenCmd() *cobra.Command {
	var tokenForm string
	subCmd := &cobra.Command{
		Use:     "token",
		Aliases: []string{"t"},
		Short:   "create token",
		Long:    "create token from user manager",
		Example: "kubectl-bcs-user-manager create token " +
			"--token_form '{\"usertype\":1,\"username\":\"\", \"expiration\":-1}' ",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreateToken(tokenForm)
			if err != nil {
				klog.Fatalf("create token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("create token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintCreateTokenCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&tokenForm, "token_form", "f", "",
		"the form used to create token ")
	return subCmd
}

// createTempTokenCmd create temp token
func createTempTokenCmd() *cobra.Command {
	var tokenForm string
	subCmd := &cobra.Command{
		Use:     "temp-token",
		Aliases: []string{"temp-token"},
		Short:   "create temp token",
		Long:    "create temp token from user manager",
		Example: "kubectl-bcs-user-manager create temp-token " +
			"--token_form '{\"usertype\":,\"username\":\"\", \"expiration\":}' ",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreateTempToken(tokenForm)
			if err != nil {
				klog.Fatalf("create temp token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("create temp token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintCreateTempTokenCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&tokenForm, "token_form", "f", "",
		"the form used to create temp token")
	return subCmd
}

// createClientTokenCmd create client token
func createClientTokenCmd() *cobra.Command {
	var tokenForm string
	subCmd := &cobra.Command{
		Use:     "client-token",
		Aliases: []string{"client-token"},
		Short:   "create client token",
		Long:    "create client token from user manager",
		Example: "kubectl-bcs-user-manager create client-token --token_form '{\"clientName\":\"\"," +
			"\"clientSecret\":\"\", \"expiration\":}'",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreateClientToken(tokenForm)
			if err != nil {
				klog.Fatalf("create client token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("create client token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintCreateClientTokenCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&tokenForm, "token_form", "f", "",
		"the form used to create token ")
	return subCmd
}
