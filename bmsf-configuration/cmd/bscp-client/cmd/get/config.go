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

//getCommitCmd: client get commit
func getCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "get commit details",
		Long:    "get commit detail information",
		Example: `
	bscp-client get commit --Id xxxxxxx
	bscp-client get ci --Id cvbnm
		`,
		RunE: handleGetCommit,
	}
	// --Id is required
	cmd.Flags().String("Id", "", "the name of business")
	cmd.MarkFlagRequired("Id")
	return cmd
}

func handleGetCommit(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --Id option
	commitID, _ := cmd.Flags().GetString("Id")
	//create business and check result
	commit, err := operator.GetCommit(context.TODO(), commitID)
	if err != nil {
		return err
	}
	if commit == nil {
		cmd.Printf("Found no Commit resource.\n")
		return nil
	}
	//todo(DeveloperJim): get App and ConfigSet info for print
	//format output
	utils.PrintDetailsCommit(commit, operator.Business)
	return nil
}
