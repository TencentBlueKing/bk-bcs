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

// Package server xxx
package server

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

var (
	commandName = "bcs-gitops-proxy"
	configFile  = "./bcs-gitops-proxy.json"
)

// NewCommand create command line for gitops manager
func NewCommand() *cobra.Command {
	option := DefaultOptions()

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "bcs-gitops-proxy is a proxy for gitops-manager",
		Long:  `bcs-gitops-proxy integrates BKIAM and websocket tunnel for gitops-manager.`,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint
			// loading configuration file for options
			if err := common.LoadConfigFile(configFile, option); err != nil {
				fmt.Fprintf(os.Stderr, "proxy load json config file failure, %s\n", err.Error())
				return err
			}
			blog.InitLogs(option.LogConfig)
			defer blog.CloseLogs()
			// config option verification
			if err := option.Complete(); err != nil {
				fmt.Fprintf(os.Stderr, "proxy option complete failed, %s\n", err.Error())
				return err
			}
			if err := option.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "proxy option validate failed, %s\n", err.Error())
				return err
			}

			// run server
			proxy := NewProxy(option)
			if err := proxy.Init(); err != nil {
				fmt.Fprintf(os.Stderr, "proxy init failed, %s", err.Error())
				return err
			}
			return proxy.Run()
		},
	}
	// setting server configuration flag
	cmd.Flags().StringVarP(&configFile, "config", "c", configFile, "proxy configuration json file")
	return cmd
}
