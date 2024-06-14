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

// newUpdateCmd create update command
func newUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update resource,such as token..",
		Long:  "update resource from bcs-user-manager",
	}
	updateCmd.AddCommand(refreshSaasTokenCmd())
	updateCmd.AddCommand(refreshPlainTokenCmd())
	updateCmd.AddCommand(updateCredentialsCmd())
	updateCmd.AddCommand(updateTokenCmd())
	return updateCmd
}

func refreshSaasTokenCmd() *cobra.Command {
	var userName string
	subCmd := &cobra.Command{
		Use:     "saas-token",
		Aliases: []string{"st"},
		Short:   "refresh saas token",
		Long:    "refresh saas token from user manager",
		Example: "kubectl-bcs-user-manager update saas-token -u [username]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.RefreshSaasToken(userName)
			if err != nil {
				klog.Fatalf("refresh saas token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("refresh saas token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintRefreshSaasTokenCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that refresh saas user token for a saas user")
	return subCmd
}

func refreshPlainTokenCmd() *cobra.Command {
	var userName, expireTime string
	subCmd := &cobra.Command{
		Use:     "plain-token",
		Aliases: []string{"pt"},
		Short:   "refresh plain-token",
		Long:    "refresh plain user token from user manager",
		Example: "kubectl-bcs-user-manager update plain-token -u [user_name] -t [expire_time]",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.RefreshPlainToken(userName, expireTime)
			if err != nil {
				klog.Fatalf("refresh plain user token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("refresh plain user token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintRefreshPlainTokenCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&userName, "user_name", "u", "",
		"the user name that refresh user token for a plain user")
	subCmd.Flags().StringVarP(&expireTime, "expire_time", "t", "",
		"the expire time that refresh user token for a plain user")

	return subCmd
}

func updateCredentialsCmd() *cobra.Command {
	var clusterId, credentialsForm string
	subCmd := &cobra.Command{
		Use:     "credentials",
		Aliases: []string{"c"},
		Short:   "update credentials",
		Long:    "update cluster credentials according cluster ID",
		Example: "kubectl-bcs-user-manager update credentials --cluster_id [cluster_id] " +
			"--credentials_form '{\"register_token\":\"\"," +
			"\"server_addresses\":\"\",\"cacert_data\":\"\",\"user_token\":\"\"}'",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.UpdateCredentials(clusterId, credentialsForm)
			if err != nil {
				klog.Fatalf("update credential according cluster ID failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("update credential according cluster ID response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintUpdateCredentialsCmdResult(flagOutput, resp)
		},
	}

	subCmd.Flags().StringVarP(&clusterId, "cluster_id", "i", "",
		"the cluster_id for update cluster credential")
	subCmd.Flags().StringVarP(&credentialsForm, "credentials_form", "f", "",
		"the credentials form for update cluster credential")
	return subCmd
}

func updateTokenCmd() *cobra.Command {
	var token, tokenForm string
	subCmd := &cobra.Command{
		Use:     "token",
		Example: "kubectl-bcs-manager update token --token [token] --form '{\"expiration\":-1}'",
		Aliases: []string{},
		Args:    cobra.ExactArgs(2),
		Short:   "update token",
		Long:    "update token from user manager",
		Run: func(cmd *cobra.Command, args []string) {
			cobra.OnInitialize(ensureConfig)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := pkg.NewClientWithConfiguration(ctx)
			resp, err := client.UpdateToken(token, tokenForm)
			if err != nil {
				klog.Fatalf("update token failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatalf("update token response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintUpdateTokenCmdResult(flagOutput, resp)
		},
	}
	subCmd.Flags().StringVarP(&token, "token", "t", "",
		"the cluster_id to update token")
	subCmd.Flags().StringVarP(&tokenForm, "form", "f", "",
		"the form used to update token")
	subCmd.MarkFlagsRequiredTogether("token", "form")
	return subCmd
}
