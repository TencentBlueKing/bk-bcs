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

package list

import (
	"context"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

//listLogicClusterCmd: client list cluster
func listConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "List configset",
		Long:    "List all ConfigSet information under application",
		RunE:    handleListConfigSet,
	}
	cmd.Flags().StringP("app", "a", "", "application name that ConfigSet belongs to")
	return cmd
}

//listMultiCommitCmd: client list commit
func listMultiCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "List commit",
		Long:    "List all Commit information under specified Application",
		RunE:    handleListMultiCommits,
	}
	cmd.Flags().StringP("app", "a", "", "application name that commit belongs to")
	return cmd
}

func handleListMultiCommits(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return err
	}
	//list all datas if exists
	multiCommits, err := operator.ListMultiCommitsAllByAppID(context.TODO(), business.Bid, app.Appid)
	if err != nil {
		return err
	}
	if multiCommits == nil {
		cmd.Printf("Found no Commit resource.\n")
		return nil
	}
	//format output
	utils.PrintMultiCommits(multiCommits, business, app)
	cmd.Println()
	cmd.Printf("\t(use \"bk-bscp-client get commit --id <commitid>\" to get commit detail)\n")
	cmd.Printf("\t(use \"bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo \"this is a example\"\" to create release to publish)\n\n")
	return nil
}

func handleListConfigSet(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return err
	}

	//list all datas if exists
	configSets, err := operator.ListConfigSetByApp(context.TODO(), appName)
	if err != nil {
		return err
	}
	if configSets == nil {
		cmd.Printf("Found no ConfigSet resource.\n")
		return nil
	}
	//format output
	utils.PrintConfigSet(configSets, business, app)
	return nil
}
