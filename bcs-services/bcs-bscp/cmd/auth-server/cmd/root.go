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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/app"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/options"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// SysOpt is the system option
var SysOpt *options.Option

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bk-bscp-authservice",
	Short: "BSCP AuthService",
	Run: func(cmd *cobra.Command, args []string) {

		if err := app.Run(SysOpt); err != nil {
			fmt.Fprintf(os.Stderr, "start auth server failed, err: %v", err)
			logs.CloseLogs()
			os.Exit(1)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	SysOpt = options.InitOptions()

	cc.InitService(cc.AuthServerName)
}
