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
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"
	"io/ioutil"

	"github.com/spf13/cobra"
)

//createStrategyCmd: client create strategy
func createStrategyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "strategy",
		Aliases: []string{"str"},
		Short:   "Create strategy",
		Long:    "Create strategy for application configuration release",
		Example: `
	bk-bscp-client create strategy --json ./somefile.json --name strategyName
	json template as followed:
	{
		"App":"appName",
		"Clusters": ["cluster01","cluster02","cluster03"],
		"Zones": ["zone01","zone02","zone03"],
		"Dcs": ["dc01","dc02","dc03"],
		"IPs": ["X.X.X.1","X.X.X.2","X.X.X.3"],
		"Labels": {
			"k1":"v1",
			"k2":"v2",
			"k3":"v3"
		},
		"LabelsAnd": {
			"k3":"1",	 
			"k4":"1,2,3"
		}
	}
		`,
		RunE: handleCreateStrategy,
	}
	//command line flags
	cmd.Flags().StringP("app", "a", "", "settings application name that strategy belongs to")
	cmd.Flags().StringP("name", "n", "", "settings strategy name.")
	cmd.Flags().StringP("json", "j", "", "json details for strategy")
	cmd.Flags().StringP("memo", "m", "", "settings memo for strategy")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("json")
	return cmd
}

func handleCreateStrategy(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	jsonFile, err := cmd.Flags().GetString("json")
	if err != nil {
		return err
	}
	memo, err := cmd.Flags().GetString("memo")
	if err != nil {
		return err
	}
	//reading all details from json-file
	jsonBytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return err
	}
	//construct createRequest
	request := &service.StrategyOption{
		Name:    name,
		AppName: appName,
		Content: string(jsonBytes),
		Memo:    memo,
	}
	//create Commit and check result
	strategyID, err := operator.CreateStrategy(context.TODO(), request)
	if err != nil {
		return err
	}
	cmd.Printf("Create Strategy successfully: %s\n\n", strategyID)
	return nil
}
