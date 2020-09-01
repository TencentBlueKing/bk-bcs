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

package release

import (
	"context"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

//InitCommands init all create commands
func InitCommands() []*cobra.Command {
	return []*cobra.Command{createMultiReleaseCmd()}
}

//createReleaseCmd: client create strategy
func createMultiReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rl"},
		Short:   "Create release",
		Long:    "Create release for application configuration commit",
		Example: `
	bk-bscp-client release --name releaseName --commitid xxxxxxxx --strategy strategyName --memo "this is a example"
		`, RunE: handleCreateRelease,
	}
	//command line flags
	cmd.Flags().StringP("app", "a", "", "settings application name that release belongs to")
	cmd.Flags().StringP("name", "n", "", "settings release name")
	cmd.Flags().StringP("strategy", "s", "", "settings release strategy name")
	cmd.Flags().StringP("commitid", "i", "", "settings release relative commitid")
	cmd.Flags().StringP("memo", "m", "", "settings release relative memo")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("commitid")
	return cmd
}

func handleCreateRelease(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	name, _ := cmd.Flags().GetString("name")
	strategyName, _ := cmd.Flags().GetString("strategy")
	multiCommitID, _ := cmd.Flags().GetString("commitid")
	memo, _ := cmd.Flags().GetString("memo")
	//construct createRequest
	request := &service.ReleaseOption{
		Name:          name,
		AppName:       appName,
		StrategyName:  strategyName,
		MultiCommitID: multiCommitID,
		Memo:          memo,
	}
	//create and check result
	multiReleaseID, err := operator.CreateMultiRelease(context.TODO(), request)
	if err != nil {
		return err
	}
	cmd.Printf("Create Release successfully: %s\n\n", multiReleaseID)
	cmd.Printf("\t (use \"bk-bscp-client get release --id <releaseid>\" to get release detail)\n")
	cmd.Printf("\t（use \"bk-bscp-client publish --id <releaseid>\" to confrim release to publish)\n\n")
	return nil
}
