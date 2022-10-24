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
		Short: "get resource info",
		Long:  "get infos from bcs-user-manager",
	}
	getCmd.AddCommand(getAdminUserCmd())
	getCmd.AddCommand(getSaasUserCmd())
	getCmd.AddCommand(getPlainUserCmd())
	getCmd.AddCommand(getRegisterTokenCmd())
	getCmd.AddCommand(getCredentialsCmd())
	getCmd.AddCommand(getPermissionCmd())
	getCmd.AddCommand(getTokenCmd())
	getCmd.AddCommand(getTokenByUserAndClusterIDCmd())
	return getCmd
}

func getAdminUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "admin-user",
		Aliases: []string{"au"},
		Short:   "get admin user from user manager",
		Long:    "",
		Example: "kubectl-bcs-user-manager get admin-user -u [username]",
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
			printer.PrintGetAdminUserCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that query admin user")
	return subCmd
}

func getSaasUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "saas-user",
		Aliases: []string{"su"},
		Short:   "get saas user from user manager",
		Long:    "",
		Example: "kubectl-bcs-user-manager get saas-user -u [username]",
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
			printer.PrintGetSaasUserCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that query sass user")
	return subCmd
}

func getPlainUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "plain-user",
		Aliases: []string{"pu"},
		Short:   "get plain user from user manager",
		Long:    "",
		Example: "kubectl-bcs-user-manager get plain-user -u [username]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetPlainUser(userName)
			if err != nil {
				klog.Fatalf("get plain user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get plain user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintGetPlainUserCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that query plain user")
	return subCmd
}

func getRegisterTokenCmd() *cobra.Command {
	var clusterId string
	subCmd := &cobra.Command{
		Use:     "register-token",
		Aliases: []string{"rt"},
		Short:   "register-token",
		Long:    "register specified cluster token from user manager",
		Example: "kubectl-bcs-user-manager get register-token --cluster_id [cluster_id]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetRegisterToken(clusterId)
			if err != nil {
				klog.Fatalf("search specified cluster token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("search specified cluster token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintGetRegisterTokenCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&clusterId, "cluster_id", "i", "",
		"the cluster_id for search specified cluster token")
	return subCmd
}

func getCredentialsCmd() *cobra.Command {
	var clusterId string
	subCmd := &cobra.Command{
		Use:     "credentials",
		Aliases: []string{"c"},
		Short:   "get credentials",
		Long:    "get credential according cluster ID",
		Example: "kubectl-bcs-user-manager get credentials --cluster_id [cluster_id]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetCredentials(clusterId)
			if err != nil {
				klog.Fatalf("get credential according cluster ID failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get credential according cluster ID response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintGetCredentialsCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&clusterId, "cluster_id", "i", "",
		"the cluster_id for get credential")
	return subCmd
}

func getPermissionCmd() *cobra.Command {
	var permissionForm string
	subCmd := &cobra.Command{
		Use:     "permission",
		Aliases: []string{"p"},
		Short:   "get permissions from user manager",
		Example: "kubectl-bcs-user-manager get permission -f '{\"user_name\":\"\",\"resource_type\":\"\"}' ",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetPermission(permissionForm)
			if err != nil {
				klog.Fatalf("get permissions failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get permissions response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintGetPermissionCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&permissionForm, "permission_form", "f", "",
		"the permission_form that query permissions")
	return subCmd
}

func getTokenCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "token",
		Aliases: []string{"t"},
		Short:   "get token from user manager",
		Example: "kubectl-bcs-user-manager get token -u [username]",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetToken(userName)
			if err != nil {
				klog.Fatalf("get token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintGetTokenCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that query token")
	return subCmd
}

func getTokenByUserAndClusterIDCmd() *cobra.Command {
	var userName, clusterID, businessID string
	subCmd := &cobra.Command{
		Use:     "extra-token",
		Aliases: []string{"et"},
		Args:    cobra.ExactArgs(3),
		Short:   "get token from user manager",
		Example: "kubectl-bcs-user-manager get extra-token -u [user_name] --cluster_id [cluster_id] --business_id [business_id]",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.GetTokenByUserAndClusterID(userName, clusterID, businessID)
			if err != nil {
				klog.Fatalf("get token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("get token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintGetTokenByUserAndClusterIDCmdResult(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "u", "",
		"the user name that query token")
	subCmd.PersistentFlags().StringVarP(&clusterID, "cluster_id", "", "",
		"the cluster_id that query token")
	subCmd.PersistentFlags().StringVarP(&businessID, "business_id", "", "",
		"the business_id that query token")
	subCmd.MarkFlagsRequiredTogether("user_name", "cluster_id", "business_id")
	return subCmd
}
