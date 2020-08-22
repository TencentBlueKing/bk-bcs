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

//listStrategyCmd: client list commit
func listStrategyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "strategy",
		Aliases: []string{"str"},
		Short:   "List strategy",
		Long:    "List all strategy information under specified Application",
		RunE:    handleListStrategies,
	}
	cmd.Flags().StringP("app", "a", "", "application name that ConfigSet belongs to")
	return cmd
}

func handleListStrategies(cmd *cobra.Command, args []string) error {
	// level 3 read appName required
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	//list all datas if exists
	strategies, err := operator.ListStrategyByApp(context.TODO(), appName)
	if err != nil {
		return err
	}
	if strategies == nil {
		cmd.Printf("Found no Strategy resource.\n")
		return nil
	}
	app, _ := operator.GetApp(context.TODO(), appName)
	//format output
	utils.PrintStrategies(strategies, app)
	return nil
}

//listMultiReleaseCmd: client list release
func listMultiReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rl"},
		Short:   "List release",
		Long:    "List all release information specified Application",
		RunE:    handleListMultiRelease,
	}
	cmd.Flags().StringP("app", "a", "", "application name that Release belongs to")
	cmd.Flags().Int32P("type", "t", 0, "the status of release, 1 is init, 2 is published, 3 is canceled, 4 is rollbacked, default 0 is all")
	return cmd
}

func handleListMultiRelease(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	queryType, _ := cmd.Flags().GetInt32("type")
	//list all datas if exists
	multiReleases, err := operator.ListMultiReleaseByApp(context.TODO(), appName, queryType)
	if err != nil {
		return err
	}
	if multiReleases == nil {
		cmd.Printf("Found no Release resource.\n")
		return nil
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil
	}
	//format output
	utils.PrintMultiReleases(operator, multiReleases, business, app)
	cmd.Printf("\n\t(use \"bk-bscp-client get release --id <releaseid>\" to get release detail)\n")
	cmd.Printf("\t(use \"bk-bscp-client publish --id <releaseid>\" to confrim release to publish)\n\n")
	return nil
}
