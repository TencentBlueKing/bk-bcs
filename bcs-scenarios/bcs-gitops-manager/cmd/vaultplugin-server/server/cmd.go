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

package server

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

var (
	commandName = "bcs-gitops-vaultplugin-server"
	configFile  = "./bcs-gitops-vaultplugin-server.json"
)

// NewCommand create command line for gitops vaultplugin-server
func NewCommand() *cobra.Command {
	option := DefaultOptions()

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "bcs-gitops-vaultplugin-server is a vaultplugin-server for gitops-manager",
		RunE: func(cmd *cobra.Command, args []string) error { // nolint
			//loading configuration file for options
			if err := common.LoadConfigFile(configFile, option); err != nil {
				fmt.Fprintf(os.Stderr, "vaultplugin-server load json config file failure, %s\n", err.Error())
				return err
			}
			blog.InitLogs(option.LogConfig)
			defer blog.CloseLogs()
			// config option verification
			if err := option.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "vaultplugin-server option validate failed, %s\n", err.Error())
				return err
			}

			// run server
			server := NewVaultPlugin(option)
			if err := server.Init(); err != nil {
				fmt.Fprintf(os.Stderr, "vaultplugin-server init failed, %s", err.Error())
				return err
			}
			return server.Run()
		},
	}
	cmd.Flags().StringVarP(&configFile, "config", "c", configFile, "vaultplugin-server configuration json file")
	return cmd
}
