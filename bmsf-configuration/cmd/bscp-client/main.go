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

package main

import (
	"bk-bscp/cmd/bscp-client/cmd"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/internal/version"
	"fmt"

	"github.com/spf13/cobra"
)

// BSCP client for Configuration distribution
func main() {
	bscpCli := &cobra.Command{
		Use:     "bscp-client",
		Long:    "bscp-client controls the BlueKing Service Configuration Platform.",
		Version: version.GetVersion(),
		//parse global option with command line & environments
		PersistentPreRun: option.ParseGlobalOption,
	}
	//loading all subcommands
	bscpCli.AddCommand(cmd.GetCommandList()...)
	//loading all global flags
	global := option.GlobalOptions
	bscpCli.PersistentFlags().StringVar(&global.ConfigFile, "configfile", global.ConfigFile, "BlueKing Service Configuration Platform CLI configuration.")
	bscpCli.PersistentFlags().StringVar(&global.Business, "business", "", "Business Name to operate. Also comes from ENV BSCP_BUSINESS")
	bscpCli.PersistentFlags().StringVar(&global.Operator, "operator", "", "user name for operation, use for audit, Also comes from ENV BSCP_OPERATOR")

	if err := bscpCli.Execute(); err != nil {
		fmt.Printf("bscp-client Error: %s\n", err.Error())
	}
}
