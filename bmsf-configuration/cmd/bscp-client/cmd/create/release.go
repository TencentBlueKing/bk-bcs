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
		Use:   "strategy",
		Short: "create strategy",
		Long:  "create strategy for application configuration release",
		Example: `
	bscp-client create strategy --app gamesvr --json ./somefile.json --name strategyName
	json template as followed:
	{
		"Appid":"appid",
		"Clusterids": ["clusterid01","clusterid02","clusterid03"],
		"Zoneids": ["zoneid01","zoneid02","zoneid03"],
		"Dcs": ["dc01","dc02","dc03"],
		"IPs": ["X.X.X.1","X.X.X.2","X.X.X.3"],
		"Labels": {
			"k1":"v1",
			"k2":"v2",
			"k3":"v3"
		}
	}
		`,
		RunE: handleCreateStrategy,
	}
	//command line flags
	cmd.Flags().StringP("app", "a", "", "settings application name that strategy belongs to")
	cmd.Flags().StringP("name", "n", "", "settings strategy name.")
	cmd.Flags().StringP("json", "j", "", "json details for strategy")
	cmd.MarkFlagRequired("app")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("json")
	return cmd
}

func handleCreateStrategy(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	name, _ := cmd.Flags().GetString("name")
	jsonFile, _ := cmd.Flags().GetString("json")
	//reading all details from json-file
	cfgBytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return err
	}
	//construct createRequest
	request := &service.StrategyOption{
		Name:    name,
		AppName: appName,
		Content: string(cfgBytes),
	}
	//create Commit and check result
	strategyID, err := operator.CreateStrategy(context.TODO(), request)
	if err != nil {
		return err
	}
	cmd.Printf("Create Strategy successfully: %s\n", strategyID)
	return nil
}

//createReleaseCmd: client create strategy
func createReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rel"},
		Short:   "create release",
		Long:    "create release for application configuration commit",
		Example: `
	bscp-client create release --app gamesvr --strategy bluestrategy --name relname --commitId xxxxxxxx
	bscp-client create rl -a game -s bluestrategy -n relname -i xxxxxxxx
		`,
		RunE: handleCreateRelease,
	}
	//command line flags
	cmd.Flags().StringP("app", "a", "", "settings application name that release belongs to.")
	cmd.MarkFlagRequired("app")
	cmd.Flags().StringP("name", "n", "", "settings release name.")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringP("strategy", "s", "", "settings release strategy name, optional")
	cmd.Flags().StringP("commitId", "i", "", "settings release relative CommitID")
	cmd.MarkFlagRequired("commitId")
	return cmd
}

func handleCreateRelease(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	name, _ := cmd.Flags().GetString("name")
	strategyName, _ := cmd.Flags().GetString("strategy")
	commitID, _ := cmd.Flags().GetString("commitId")
	//construct createRequest
	request := &service.ReleaseOption{
		Name:         name,
		AppName:      appName,
		StrategyName: strategyName,
		CommitID:     commitID,
	}
	//create and check result
	releaseID, err := operator.CreateRelease(context.TODO(), request)
	if err != nil {
		return err
	}
	cmd.Printf("Create Release successfully: %s\n", releaseID)
	return nil
}
