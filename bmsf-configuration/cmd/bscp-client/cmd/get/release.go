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

package get

import (
	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"

	"github.com/spf13/cobra"
)

//getStrategyCmd: client get strategy
func getStrategyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "strategy",
		Short: "get strategy details",
		Long:  "get strategy detail information",
		Example: `
	bscp-client get strategy --app gameserver --name xxxxxxx
		`,
		RunE: handleGetStrategy,
	}
	// --name is required
	cmd.Flags().String("name", "", "the name of strategy")
	cmd.Flags().String("app", "", "the name of application")
	cmd.MarkFlagRequired("app")
	cmd.MarkFlagRequired("name")
	return cmd
}

func handleGetStrategy(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --name option
	name, _ := cmd.Flags().GetString("name")
	appName, _ := cmd.Flags().GetString("app")
	//create business and check result
	strategy, err := operator.GetStrategy(context.TODO(), appName, name)
	if err != nil {
		return err
	}
	if strategy == nil {
		cmd.Printf("Found no strategy resource.\n")
		return nil
	}
	app, _ := operator.GetApp(context.TODO(), appName)
	//format output
	utils.PrintStrategy(strategy, app)
	return nil
}

//getReleaseCmd: client get release
func getReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rel"},
		Short:   "get release",
		Long:    "get release detail information",
		Example: `
	bscp-client get release --Id xxxxxxxxxxxx
	bscp-client get rel -i xxxxxxxxxxxxx
		`,
		RunE: handleGetRelease,
	}
	// --Id is required
	cmd.Flags().StringP("Id", "i", "", "the ID of release")
	cmd.MarkFlagRequired("Id")
	return cmd
}

func handleGetRelease(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --Id option
	releaseID, _ := cmd.Flags().GetString("Id")
	//create business and check result
	release, err := operator.GetRelease(context.TODO(), releaseID)
	if err != nil {
		return err
	}
	if release == nil {
		cmd.Printf("Found no Release resource.\n")
		return nil
	}
	//format output
	utils.PrintRelease(release)
	return nil
}
