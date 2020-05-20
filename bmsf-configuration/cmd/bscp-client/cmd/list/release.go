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
		Use:   "strategy",
		Short: "list strategy",
		Long:  "list all strategy information under specified Application",
		Example: `
	bscp-client list strategy --business somegame --app gameserver
		 `,
		RunE: handleListStrategies,
	}
	cmd.Flags().StringP("app", "a", "", "application name that ConfigSet belongs to")
	cmd.MarkFlagRequired("app")
	return cmd
}

func handleListStrategies(cmd *cobra.Command, args []string) error {
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

//listReleaseCmd: client list release
func listReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rel"},
		Short:   "list release",
		Long:    "list all release information specified ConfigSet",
		Example: `
	bscp-client list release --business somegame --app gameserver --cfgset name
		 `,
		RunE: handleListRelease,
	}
	cmd.Flags().String("app", "a", "application name that Release belongs to")
	cmd.Flags().String("cfgset", "c", "application name that release belongs to")
	cmd.MarkFlagRequired("app")
	cmd.MarkFlagRequired("cfgset")
	return cmd
}

func handleListRelease(cmd *cobra.Command, args []string) error {
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	cfgSetName, _ := cmd.Flags().GetString("cfgset")
	//list all datas if exists
	releases, err := operator.ListReleaseByApp(context.TODO(), appName, cfgSetName)
	if err != nil {
		return err
	}
	if releases == nil {
		cmd.Printf("Found no Release resource.\n")
		return nil
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil
	}
	//format output
	utils.PrintReleases(releases, business, app)
	return nil
}
