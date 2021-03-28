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
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-patcher/service"
	"bk-bscp/pkg/framework"
)

func genRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run bk-bscp-patcher",
		Run: func(cmd *cobra.Command, args []string) {
			patcher := service.NewPatcher()
			framework.Run(patcher)
		},
	}

	// subcommand flags.
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("httpprof"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("cpuprofile"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("memprofile"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("configfile"))

	// service flags.
	runCmd.Flags().String("endpoint-ip", "", "IP to run application server on")
	viper.BindPFlag("server.endpoint.ip", runCmd.Flags().Lookup("endpoint-ip"))
	runCmd.Flags().Int("endpoint-port", 0, "Port to run application server on")
	viper.BindPFlag("server.endpoint.port", runCmd.Flags().Lookup("endpoint-port"))
	runCmd.Flags().String("cronfile", "", "Crontab job spec config file")
	viper.BindPFlag("server.cronFile", runCmd.Flags().Lookup("cronfile"))

	return runCmd
}
