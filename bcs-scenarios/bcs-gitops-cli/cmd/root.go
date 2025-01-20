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
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"

	"github.com/pterm/pterm"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/cmd/argocd"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/cmd/kubectl"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/cmd/secret"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/cmd/terraform"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/cmd/workflow"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/internal/clusterset"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/version"
)

var (
	defaultCfgFile = "./.bcs/config.yaml"
)

func ensureConfig() {
	if options.ConfigFile == "" {
		options.ConfigFile = defaultCfgFile
	}

	blog.InitLogs(conf.LogConfig{
		LogDir:          "",
		ToStdErr:        true,
		AlsoToStdErr:    true,
		Verbosity:       2,
		StdErrThreshold: "2",
	})
	if options.LogV != 0 {
		blog.SetV(int32(options.LogV))
	}
	options.Parse(options.ConfigFile)
}

// NewRootCommand returns the rootCmd instance
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "powerapp",
		Short: "powerapp controls gitops service",
		Long:  printLogo(),
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version detail info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetVersion())
		},
	}
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		ensureConfig()
		blog.SetV(int32(options.LogV))
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(terraform.NewTerraformCmd())
	rootCmd.AddCommand(workflow.NewWorkflowCmd())
	rootCmd.AddCommand(secret.NewSecretCmd())
	argoCmd := argocd.NewArgoCmd()
	argoCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		ensureConfig()
		_ = argoCmd.PersistentFlags().Set("header", "X-BCS-Client: bcs-gitops-manager")
		_ = argoCmd.PersistentFlags().Set("header", "bkUserName: admin")
		_ = argoCmd.PersistentFlags().Set("grpc-web-root-path", options.GlobalOption().ProxyPath)
		_ = argoCmd.PersistentFlags().Set("server", options.GlobalOption().Server)
		_ = argoCmd.PersistentFlags().Set("header", fmt.Sprintf("Authorization: Bearer %s",
			options.GlobalOption().Token))
	}
	rootCmd.AddCommand(argoCmd)
	kubeObj := kubectl.NewKubectlCmd()
	kubeCmd := kubeObj.GetCommand()
	kubectlPreRunOriginal := kubeCmd.PersistentPreRunE
	kubeCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := kubectlPreRunOriginal(cmd, args); err != nil {
			return nil
		}
		ensureConfig()
		kubeCfg := kubeObj.GetConfigs()
		serverUrl := options.GlobalOption().Server
		// NOTE: dirty transfer used to make user to prod api
		serverUrl = strings.ReplaceAll(serverUrl, "debug-bcs-api", "prod-bcs-api")

		setter := clusterset.ClusterSetter{}
		clusterID, err := setter.GetCurrentCluster()
		if err != nil {
			utils.ExitError(fmt.Sprintf("get current-cluster failed: %s", err.Error()))
		}
		apiServer := fmt.Sprintf("https://%s/clusters/%s/", serverUrl, clusterID)
		bearerToken := options.GlobalOption().Token
		kubeCfg.APIServer = &apiServer
		kubeCfg.BearerToken = &bearerToken
		return nil
	}
	rootCmd.AddCommand(kubeCmd)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(errors.Wrapf(err, "get user home directory failed"))
	} else {
		defaultCfgFile = path.Join(homeDir, defaultCfgFile)
	}
	rootCmd.PersistentFlags().StringVar(&options.ConfigFile, "bcscfg", defaultCfgFile,
		"Config file. Example: '{\"server\": \"bcs-api.gateway.com\", \"token\": \"\"}'")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().IntVarP(&options.LogV, "verbose", "v", 2,
		"Log level")
	return rootCmd
}

// NOCC:tosa/indent(设计如此)
func printLogo() string {
	panel := pterm.DefaultHeader.WithMargin(8).
		WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).Sprint("Manage your gitops more easily.")
	// nolint
	logo := pterm.FgLightGreen.Sprint(`
 ____                              _     ____   ____  
|  _ \  ___ __      __ ___  _ __  / \   |  _ \ |  _ \ 
| |_) |/ _ \\ \ /\ / // _ \| '__|/ _ \  | |_) || |_) |
|  __/| (_) |\ V  V /|  __/| |  / ___ \ |  __/ |  __/ 
|_|    \___/  \_/\_/  \___||_| /_/   \_\|_|    |_|    
`)
	pterm.Info.Prefix = pterm.Prefix{
		Text:  "Tips",
		Style: pterm.NewStyle(pterm.BgBlue, pterm.FgLightWhite),
	}
	return fmt.Sprintf(`
%s%s
`, panel, logo)
}
