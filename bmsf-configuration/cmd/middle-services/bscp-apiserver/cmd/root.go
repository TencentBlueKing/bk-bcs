/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// APIServerRootCmd handles all application command line interfaces for apiserver.
type APIServerRootCmd struct {
	cobra.Command
	RunCmd     *cobra.Command
	VersionCmd *cobra.Command
}

// RootCmd is root command for apiserver.
var RootCmd = genRootCmd()

func genRootCmd() *APIServerRootCmd {
	rootCmd := &APIServerRootCmd{}

	// basic information.
	rootCmd.Use = "bk-bscp-apiserver"

	// all application commands.
	rootCmd.RunCmd = genRunCmd()
	rootCmd.VersionCmd = genVersionCmd()

	// add sub commands.
	rootCmd.AddCommand(rootCmd.RunCmd)
	rootCmd.AddCommand(rootCmd.VersionCmd)

	return rootCmd
}
