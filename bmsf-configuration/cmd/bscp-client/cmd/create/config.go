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

package create

import (
	"context"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/accessserver"
)

// createConfigSetCmd: client create configset.
func createConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "Create configset",
		Long:    "Create ConfigSet for application",
		Example: `
	bk-bscp-client create configset --path cfgPath --name cfgName
		`,
		RunE: handleCreateConfigSet,
	}
	// command line flags.
	cmd.Flags().StringP("name", "n", "", "settings new ConfigSet Name")
	cmd.Flags().StringP("app", "a", "", "settings app that ConfigSet belongs to")
	cmd.Flags().StringP("path", "p", "", "settings new ConfigSet Fpath")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("path")
	return cmd
}

func handleCreateConfigSet(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	// get global command info and create business operator.
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}

	// check all flags.
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	fpath, err := cmd.Flags().GetString("path")
	if err != nil {
		return err
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return err
	}

	// construct createRequest.
	request := &accessserver.CreateConfigSetReq{
		Bid:     business.Bid,
		Appid:   app.Appid,
		Name:    name,
		Fpath:   fpath,
		Creator: operator.User,
	}

	// create configset and check result.
	cfgSetID, err := operator.CreateConfigSet(context.TODO(), request)
	if err != nil {
		return err
	}

	cmd.Printf("Create ConfigSet successfully: %s\n\n", cfgSetID)
	return nil
}
