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

	"bk-bscp/cmd/atomic-services/bscp-bcs-sidecar/service"
	"bk-bscp/pkg/framework"
)

func genRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run bk-bcs-sidecar",
		Run: func(cmd *cobra.Command, args []string) {
			sidecar := service.NewSidecar()
			framework.Run(sidecar)
		},
	}

	// subcommand flags.
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("httpprof"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("cpuprofile"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("memprofile"))
	runCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("configfile"))

	return runCmd
}
