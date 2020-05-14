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

package update

import (
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"

	"github.com/spf13/cobra"
)

//updateReleaseCmd: client update commit
func updateReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rel"},
		Short:   "update release refference",
		Long:    "update release information: name & strategy relation",
		Example: `
	bscp-client update release --Id xxxxx --app appname --new-name newname
		 `,
		RunE: handleUpdateRelease,
	}
	//command line flags
	cmd.Flags().StringP("new-name", "n", "", "new release name")
	cmd.Flags().String("Id", "", "release ID for index")
	cmd.Flags().String("app", "", "app name for index")
	cmd.MarkFlagRequired("Id")
	cmd.MarkFlagRequired("app")
	return cmd
}

func handleUpdateRelease(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	releaseID, _ := cmd.Flags().GetString("Id")
	appName, _ := cmd.Flags().GetString("app")
	name, _ := cmd.Flags().GetString("new-name")
	//construct createRequest
	request := &service.ReleaseOption{
		//StrategyName release for specified Strategy, can be empty
		ReleaseID: releaseID,
		Name:      name,
		AppName:   appName,
	}
	//update and check result
	if err := operator.UpdateRelease(context.TODO(), request); err != nil {
		return err
	}
	cmd.Printf("Update Release successfully: %s\n", releaseID)
	return nil
}
