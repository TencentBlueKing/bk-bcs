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

	"github.com/spf13/cobra"


	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/cmd/argocd"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/cmd/terraform"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/version"
)

const (
	defaultCfgFile = "/root/.bcs/config.yaml"
)

var (
	cfgFile string
)

func ensureConfig() {
	if cfgFile == "" {
		cfgFile = defaultCfgFile
	}

	blog.InitLogs(conf.LogConfig{
		LogDir:          "",
		ToStdErr:        true,
		AlsoToStdErr:    true,
		Verbosity:       2,
		StdErrThreshold: "2",
	})
	options.Parse(cfgFile)
}

// NewRootCommand returns the rootCmd instance
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bcs-gitops",
		Short: "bcs-gitops controls gitops service",
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version detail info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetVersion())
		},
	}
	ensureConfig()

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(argocd.NewArgoCmd())
	rootCmd.AddCommand(terraform.NewTerraformCmd())
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultCfgFile,
		"Config file. Example: '{\"server\": \"bcs-api.gateway.com\", \"token\": \"\"}'")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().IntVarP(&options.LogV, "verbose", "v", 2,
		"Log level")
	return rootCmd
}
