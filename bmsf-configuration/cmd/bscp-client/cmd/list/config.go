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
	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"

	"github.com/spf13/cobra"
)

//listLogicClusterCmd: client list cluster
func listConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "list configset",
		Long:    "list all ConfigSet information under application",
		Example: `
	bscp-client list configset --business somegame --app gameserver
	bscp-client list cfgset --business somegame --app gameserver
		 `,
		RunE: handleListConfigSet,
	}
	cmd.Flags().StringP("app", "a", "", "application name that ConfigSet belongs to")
	cmd.MarkFlagRequired("app")
	return cmd
}

//listCommitCmd: client list commit
func listCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "list commit",
		Long:    "list all Commit information under specified ConfigSet",
		Example: `
	bscp-client list commit --business somegame --app gameserver --cfgset some
	bscp-client list ci --business somegame -a gameserver -c some
		 `,
		RunE: handleListCommits,
	}
	cmd.Flags().StringP("app", "a", "", "application name that ConfigSet belongs to")
	cmd.Flags().StringP("cfgset", "c", "", "ConfigSet name that Commits belongs to")
	cmd.MarkFlagRequired("app")
	cmd.MarkFlagRequired("cfgset")
	return cmd
}

func handleListCommits(cmd *cobra.Command, args []string) error {
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	cfgSetName, err := cmd.Flags().GetString("cfgset")
	if err != nil {
		return err
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return err
	}
	configSet, err := operator.GetConfigSet(context.TODO(), appName, cfgSetName)
	if err != nil {
		return err
	}
	if configSet == nil {
		cmd.Printf("Found no ConfigSet resource.\n")
		return nil
	}

	//list all datas if exists
	commits, err := operator.ListCommitsAllByID(context.TODO(), business.Bid, app.Appid, cfgSetName)
	if err != nil {
		return err
	}
	if commits == nil {
		cmd.Printf("Found no Commit resource.\n")
		return nil
	}
	//format output
	utils.PrintCommits(commits, business, app, configSet)
	return nil
}

func handleListConfigSet(cmd *cobra.Command, args []string) error {
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
