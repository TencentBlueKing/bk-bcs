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

package commit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

//InitCommands init all create commands
func InitCommands() []*cobra.Command {
	return []*cobra.Command{commitCmd()}
}

//getBusinessCmd: client create business
func commitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "Submit the files in the scan area",
		Long:    "Submit the files in the scan area and generate a commit record",
		Example: `
	bk-bscp-client commit --memo "create commit"
		`,
		RunE: handleCommit,
	}
	//command line flags
	cmd.Flags().StringP("app", "a", "", "settings application name that multi-commit belongs to")
	cmd.Flags().StringP("memo", "m", "", "settings memo that multi-commit belongs to")
	return cmd
}

func handleCommit(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// CreateMultiCommit
	memo, _ := cmd.Flags().GetString("memo")
	appName, _ := cmd.Flags().GetString("app")
	// query delete files of record
	recordConfigFiles, err := utils.ReadRCFileFromScanArea()
	if err != nil {
		return err
	}
	recordConfigFiles = utils.DeleteEdFileFromRecord(recordConfigFiles)
	request := &service.CreateMultiCommitOption{
		AppName:           appName,
		Memo:              memo,
		RecordConfigFiles: recordConfigFiles,
	}
	multiCommitID, err := operator.CreateMultiCommit(context.TODO(), request)
	if err != nil {
		return fmt.Errorf("commit fail, please check the file list of the scan area")
	}
	err = operator.ConfirmMultiCommit(context.TODO(), multiCommitID)
	if err != nil {
		return err
	}

	// clear repo record
	configMap := make(map[string]service.ConfigFile)
	configJson, _ := json.Marshal(configMap)
	configBase64 := base64.StdEncoding.EncodeToString(configJson)
	ioutil.WriteFile("./.bscp/record", []byte(configBase64), 0644)
	cmd.Printf("Commit successfully! commitid: %s\n\n", multiCommitID)
	cmd.Printf("\t(use \"bk-bscp-client get commit --id <commitid>\" to get commit detail)\n")
	cmd.Printf("\t(use \"bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo \"this is a example\"\" to create release to publish)\n\n")
	return nil
}
