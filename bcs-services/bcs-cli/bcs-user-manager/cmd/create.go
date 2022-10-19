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
		Short: "create",
		Long:  "create resource from bcs-user-manager",
	}
	listCmd.AddCommand(createClusterCmd())
	listCmd.AddCommand(createSaasUserCmd())
	listCmd.AddCommand(createAdminUserCmd())
	listCmd.AddCommand(createPlainUserCmd())
	listCmd.AddCommand(createRegisterTokenCmd())
	listCmd.AddCommand(grantPermissionCmd())
	return listCmd
}

func createAdminUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "admin-user",
		Aliases: []string{"au"},
		Short:   "create admin user from user manager",
		Long:    "",
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
			//printer.PrintAdminUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "n", "",
		"the user name that query admin user")
	return subCmd
}

func createSaasUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "saas-user",
		Aliases: []string{"su"},
		Short:   "create saas user from user manager",
		Long:    "",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreateSaasUser(userName)
			if err != nil {
				klog.Fatalf("create saas user failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("create saas user response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			//printer.PrintAdminUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "n", "",
		"the user name that query saas user")
	return subCmd
}

func createPlainUserCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "plain-user",
		Aliases: []string{"pu"},
		Short:   "create plain user from user manager",
		Long:    "",
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
			//printer.PrintAdminUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&userName, "user_name", "n", "",
		"the user name that query plain user")
	return subCmd
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

func createRegisterTokenCmd() *cobra.Command {
	var clusterId string
	subCmd := &cobra.Command{
		Use:     "register-token",
		Aliases: []string{"rk"},
		Short:   "register-token",
		Long:    "register specified cluster token from user manager",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.CreateRegisterToken(clusterId)
			if err != nil {
				klog.Fatalf("register specified cluster token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("register specified cluster token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			//printer.PrintAdminUserListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&clusterId, "cluster_id", "c", "",
		"the id which cluser will register token ")
	return subCmd
}

func grantPermissionCmd() *cobra.Command {
	var reqBody string
	subCmd := &cobra.Command{
		Use:     "permissions",
		Example: "kubectl-bcs-manager create ps -p '{name=yxw}'",
		Aliases: []string{"permissions", "ps"},
		//Short:   "revoke permissions from user manager",
		Short: "revoke permission",
		Long:  "revoke permissions from user manager",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.RevokePermission(reqBody)
			if err != nil {
				klog.Fatalf("revoke permissions failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("revoke permissions response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintPermissionListInTable(flagOutput, resp)
		},
	}
	subCmd.PersistentFlags().StringVarP(&reqBody, "permissions", "p", "",
		"the permissions which will be revoked")

	return subCmd
}
