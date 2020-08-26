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
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
	"bk-bscp/internal/strategy"
)

//getStrategyCmd: client get strategy
func getStrategyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "strategy",
		Aliases: []string{"str"},
		Short:   "Get strategy details",
		Long:    "Get strategy detail information",
		RunE:    handleGetStrategy,
	}
	// --name is required
	cmd.Flags().StringP("name", "n", "", "the name of strategy")
	cmd.Flags().StringP("id", "i", "", "the id of strategy")
	cmd.Flags().StringP("app", "a", "", "the name of application")
	return cmd
}

func handleGetStrategy(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}

	//check flag
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	sid, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	if len(name) == 0 && len(sid) == 0 {
		return fmt.Errorf("id or name is required")
	}
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}

	// query strategy
	strategyInfo := &common.Strategy{}
	if len(sid) != 0 {
		strategyInfo, err = operator.GetStrategyById(context.TODO(), sid)
	} else if len(name) != 0 {
		strategyInfo, err = operator.GetStrategyByName(context.TODO(), appName, name)
	}

	// check result
	if err != nil {
		return err
	}
	if strategyInfo == nil {
		cmd.Printf("Found no strategy resource.\n")
		return nil
	}

	// Convert resource id to name
	strategyContent := &strategy.Strategy{}
	if err := json.Unmarshal([]byte(strategyInfo.Content), strategyContent); err != nil {
		return err
	}
	strategyJson, err := operator.GetStrategyFromIdToName(context.TODO(), strategyContent)
	if err != nil {
		return err
	}
	strategyInfo.Content = string(strategyJson)

	app, err := operator.GetApp(context.TODO(), appName)
	if err != nil {
		return err
	}
	//format output
	utils.PrintStrategy(strategyInfo, app)
	cmd.Println()
	return nil
}

func handleGetRelease(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --Id option
	releaseID, _ := cmd.Flags().GetString("mid")
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
	application, err := operator.QueryApplication(context.TODO(), &common.App{Bid: release.Bid, Appid: release.Appid})
	cfgset, err := operator.QueryConfigSet(context.TODO(), &common.ConfigSet{Bid: release.Bid, Appid: release.Appid, Cfgsetid: release.Cfgsetid})
	strategyName := ""
	if len(release.Strategyid) != 0 {
		strategy, _ := operator.GetStrategyById(context.TODO(), release.Strategyid)
		strategyName = strategy.Name
	}
	utils.PrintRelease(release, operator.Business, application.Name, path.Clean(cfgset.Fpath+"/"+cfgset.Name), strategyName)
	cmd.Println()
	cmd.Printf("\t（use \"bk-bscp-client publish --id <releaseid>\" to confrim release to publish)\n\n")
	return nil
}

//getReleaseCmd: client get release
func getMultiReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rl"},
		Short:   "Get release detail",
		Long:    "Get release detail information",
		RunE:    handleGetReleaseInfo,
	}
	// --Id is required
	cmd.Flags().StringP("id", "", "", "the id of release")
	cmd.Flags().StringP("mid", "", "", "the id of release module")
	return cmd
}

func handleGetReleaseInfo(cmd *cobra.Command, args []string) error {
	multiReleaseID, _ := cmd.Flags().GetString("id")
	releaseID, _ := cmd.Flags().GetString("mid")
	if len(multiReleaseID) == 0 && len(releaseID) == 0 {
		return fmt.Errorf("id or mid is a required parameter")
	} else if len(multiReleaseID) != 0 && len(releaseID) != 0 {
		return fmt.Errorf("mid or id can only enter one as a parameter")
	}
	if len(multiReleaseID) != 0 {
		handleGetMultiRelease(cmd, args)
	}
	if len(releaseID) != 0 {
		handleGetRelease(cmd, args)
	}
	return nil
}

func handleGetMultiRelease(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --Id option
	multiReleaseId, _ := cmd.Flags().GetString("id")
	//create business and check result
	multiRelease, metaDatas, err := operator.GetMultiRelease(context.TODO(), multiReleaseId)
	if err != nil {
		return err
	}
	if multiRelease == nil {
		cmd.Printf("Found no Release resource.\n")
		return nil
	}
	//format output
	application, _ := operator.QueryApplication(context.TODO(), &common.App{Bid: multiRelease.Bid, Appid: multiRelease.Appid})
	strategyName := ""
	if len(multiRelease.Strategyid) != 0 {
		strategy, _ := operator.GetStrategyById(context.TODO(), multiRelease.Strategyid)
		strategyName = strategy.Name
	}
	utils.PrintMultiRelease(multiRelease, operator.Business, application.Name, strategyName)
	utils.PrintMultiReleaseMetadatas(operator, metaDatas, multiRelease.Bid, multiRelease.Appid)
	cmd.Println()
	cmd.Printf("\t（use \"bk-bscp-client get release --mid <moduleId>\" to get release detail）\n")
	cmd.Printf("\t（use \"bk-bscp-client publish --id <releaseid>\" to confrim release to publish）\n\n")
	return nil
}
