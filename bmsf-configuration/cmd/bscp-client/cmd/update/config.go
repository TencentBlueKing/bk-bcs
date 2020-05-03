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
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

//updateCommitCmd: client update commit
func updateCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "update commit",
		Long:    "update commit configuration content",
		Example: `
	bscp-client update commit --Id xxxx --config-file ./somefile.json
	bscp-client update ci --Id xxxx -f ./somefile.json
		`,
		RunE: handleUpdateCommit,
	}
	//command line flags
	cmd.Flags().String("Id", "", "settings application name that commit belongs to")
	cmd.Flags().StringP("config-file", "f", "", "configuration detail information that this commit works for.")
	cmd.Flags().StringP("template", "t", "", "configuration template that this commit use for generating configuration.")
	cmd.MarkFlagRequired("Id")
	return cmd
}

func handleUpdateCommit(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	templateFile, err := cmd.Flags().GetString("template")
	if err != nil {
		return err
	}
	if len(configFile) == 0 && len(templateFile) == 0 {
		return fmt.Errorf("--config-file or --template is required")
	}
	//reading all details from config-file
	var cfgContent []byte
	if len(configFile) != 0 {
		cfgBytes, readErr := ioutil.ReadFile(configFile)
		if readErr != nil {
			return err
		}
		cfgContent = cfgBytes
	}
	commitID, _ := cmd.Flags().GetString("Id")
	//construct createRequest
	request := &service.UpdateCommitOption{
		CommitID: commitID,
		Configs:  cfgContent,
	}
	//update Commit and check result
	if err = operator.UpdateCommit(context.TODO(), request); err != nil {
		return err
	}
	cmd.Printf("Update Commit successfully: %s\n", commitID)
	return nil
}
