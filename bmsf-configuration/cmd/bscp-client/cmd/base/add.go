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

package base

import (
	"fmt"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

var addCmd *cobra.Command

// subcommand add file to the scan area
func addFileCmd() *cobra.Command {
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add the configuration file to the scan area",
		Long:  "Add the configuration file to the scan area, which is the content submitted by the commit",
		Example: `
	bk-bscp-client add .
	bk-bscp-client add etc/local.yaml etc/server.yaml
		`,
		RunE: handAddConfigFile,
	}
	return addCmd
}

func handAddConfigFile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Nothing specified, nothing added. Maybe you wanted to say 'bk-bscp-client add .?\n")
	}
	recordConfigFiles, err := utils.ReadRCFileFromScanArea()
	if err != nil {
		return err
	}
	// Clear deleted files of recordConfigFiles
	recordConfigFiles = utils.DeleteEdFileFromRecord(recordConfigFiles)

	// get addFile and judge file is exist 三种输入内容 （1. '.' 全部 2. 文件夹 3. 文件）
	var addFiles []string
	if args[0] == "." && len(args) == 1 {
		addFiles = utils.GetCurrentDirAllFiles("./")
	} else {
		for _, filePath := range args {
			// check path
			if !utils.IsExists(filePath) {
				return fmt.Errorf("Pathspec '%s' did not match any files\n", filePath)
			}

			// check is dir
			if utils.IsDir(filePath) {
				files := utils.GetCurrentDirAllFiles(filePath)
				for _, file := range files {
					addFiles = append(addFiles, file)
				}
			} else {
				addFiles = append(addFiles, filePath)
			}
		}
	}

	// add to the scan area
	for _, addFile := range addFiles {
		addConfig := service.ConfigFile{State: 1}
		recordConfigFiles[addFile] = addConfig
	}

	// write to ./.bscp/record
	err = utils.WriteRCFileToScanArea(recordConfigFiles)
	if err != nil {
		return fmt.Errorf("%s - %s", option.ErrMsg_FILE_WRITEFAIL, err)
	}
	return nil
}
