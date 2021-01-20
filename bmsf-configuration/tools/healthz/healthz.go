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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	pbapiserver "bk-bscp/internal/protocol/apiserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/json"
)

var (
	// host is mysql host.
	host string

	// port is mysql port.
	port int

	// timeout is database read/write timeout.
	timeout time.Duration
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "APIServer host.")
	flag.IntVar(&port, "port", 8080, "APIServer port.")
	flag.DurationVar(&timeout, "timeout", 3*time.Second, "Health check request timeout.")
}

func healthz() (*pbapiserver.HealthzResponse, error) {
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(fmt.Sprintf("http://%s:%d/healthz", host, port))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &pbapiserver.HealthzResponse{}
	if err := json.UnmarshalPB(string(body), response); err != nil {
		return nil, err
	}

	if !response.Result {
		return nil, fmt.Errorf("request to healthz api failed, %s", body)
	}
	if response.Code != pbcommon.ErrCode_E_OK {
		return nil, fmt.Errorf("healthz failed, errcode: %d, errmsg: %s", response.Code, response.Message)
	}

	return response, nil
}

// genSimpleHealthzCmd generates simple healthz command.
func genSimpleHealthzCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "SimpleHealthz",
		Short: "Simple healthz check command.",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := healthz()
			if err != nil {
				fmt.Printf("make healthz request failed, %+v\n", err)
				return
			}
			fmt.Println(response.Data.IsHealthy)
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("host"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("port"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))

	return cmd
}

// genHealthzCmd generates healthz command.
func genHealthzCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Healthz",
		Short: "Healthz check command.",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := healthz()
			if err != nil {
				fmt.Printf("make healthz request failed, %+v\n", err)
				return
			}
			fmt.Printf("healthz result[%+v]:\n", response.Data.IsHealthy)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"module", "version", "build_time", "git_hash", "is_healthy", "message"})

			for _, module := range response.Data.Modules {
				var line []string
				line = append(line, module.Module)
				line = append(line, module.Version)
				line = append(line, module.BuildTime)
				line = append(line, module.GitHash)
				line = append(line, fmt.Sprintf("%+v", module.IsHealthy))
				line = append(line, module.Message)
				table.Append(line)
			}
			table.Render()
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("host"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("port"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))

	return cmd
}

// genSubCmds returns sub commands.
func genSubCmds() []*cobra.Command {
	cmds := []*cobra.Command{}
	cmds = append(cmds, genSimpleHealthzCmd())
	cmds = append(cmds, genHealthzCmd())
	return cmds
}

// bscp healthz tool.
func main() {
	// root command.
	rootCmd := &cobra.Command{Use: "bk-bscp-healthz-tool"}

	// sub commands.
	subCmds := genSubCmds()

	// add sub commands.
	rootCmd.AddCommand(subCmds...)

	// run root command.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
