/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"bscp.io/cmd/data-service/app"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/runtime/flags"
)

// SysOpt is the system option
var SysOpt *cc.SysOption

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bk-bscp-dataservice",
	Short: "BSCP DataService",
	Run: func(cmd *cobra.Command, args []string) {
		app.RunServer(SysOpt)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// add system flags
	fs := pflag.CommandLine
	SysOpt = flags.SysFlags(fs)
	rootCmd.Flags().AddFlagSet(fs)

	fs.IntVar(&SysOpt.GRPCPort, "grpc-port", 9511, "grpc service port")
	fs.IntVar(&SysOpt.Port, "port", 9611, "http/metrics port")

	cc.InitService(cc.DataServiceName)
}
