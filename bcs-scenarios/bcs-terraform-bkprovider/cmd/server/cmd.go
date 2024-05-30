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
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
)

var (
	commandName = "bcs-terraform-bkprovider"
	configFile  = "./bcs-terraform-bkprovider.json"
)

// NewServerCommand create command line for server
func NewServerCommand() *cobra.Command {
	option := common.DefaultOptions()

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "bcs-terraform-bkprovider is a bcs-service for terraform controller to call",
		Long: `
		bcs services for terraform controller 
		`,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint
			// loading configuration file for options
			if err := loadConfigFile(configFile, option); err != nil {
				fmt.Fprintf(os.Stderr, "server load json config file failure, %s\n", err.Error())
				return err
			}
			blog.InitLogs(option.LogConfig)
			blog.Infof("configs: %s", common.ToJsonString(option))
			// config option verification
			if err := option.Complete(); err != nil {
				fmt.Fprintf(os.Stderr, "server option complete failed, %s\n", err.Error())
				return err
			}
			if err := option.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "server option validate failed, %s\n", err.Error())
				return err
			}

			// run server
			serv := NewServer(option)
			if err := serv.Init(); err != nil {
				fmt.Fprintf(os.Stderr, "server init failed, %s", err.Error())
				return err
			}
			return serv.Run()
		},
	}
	// setting server configuration flag
	cmd.Flags().StringVarP(&configFile, "config", "c", configFile, "server configuration json file")
	return cmd
}

// loadConfigFile loading json config file
func loadConfigFile(fileName string, opt *common.Options) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, opt)
}
