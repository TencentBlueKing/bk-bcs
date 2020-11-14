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
	"fmt"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/internal/version"
)

// BSCP client for Configuration distribution
func main() {
	bscpCli := &cobra.Command{
		Use:   "bk-bscp-client",
		Short: "bk-bscp-client controls the BlueKing Service Configuration Platform.",
		Long: `bk-bscp-client controls the BlueKing Service Configuration Platform.

Publishing ConfigSet steps：
    bk-bscp-client init  -> bk-bscp-client add -> bk-bscp-client commit -> bk-bscp-client release -> bk-bscp-client publish

Explanation:
    First initialize the configuration file repository, and then add the configuration files to be submitted to the scanning area. Then, use the commit command to submit the scan area file (after submission, the content in the scan area will be cleared). Use the release command to select the commit record submitted and generate the corresponding release version. Finally, use the publish command to select the release version to be published for publication.
`,
		Version: version.GetVersion(),
		//parse global option with command line & environments
		PersistentPreRunE: option.ParseGlobalOption,
	}
	bscpCli.AddCommand(cmd.GetCommandList()...)
	//loading all global flags
	global := option.GlobalOptions
	bscpCli.PersistentFlags().StringVar(&global.Business, "business", "", "business Name to operate. Get parameter priority: command -> env -> .bscp/desc")
	bscpCli.PersistentFlags().StringVar(&global.Operator, "operator", "", "user name for operation.  Get parameter priority: command -> env -> .bscp/desc")
	bscpCli.PersistentFlags().StringVar(&global.Token, "token", "", "user token for operation. Get parameter priority: command -> env -> .bscp/desc")
	if err := bscpCli.Execute(); err != nil {
		fmt.Printf("bscp-client Error: %s\n", err.Error())
	}
}
