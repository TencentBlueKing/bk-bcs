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
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
)

var statusCmd *cobra.Command

// init all resource create sub command.
func statusFileCmd() *cobra.Command {
	statusCmd = &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "Show the working tree status",
		Long:    "Display the difference between the scan area file and the local warehouse",
		RunE:    handStatus,
	}
	return statusCmd
}

func handStatus(cmd *cobra.Command, args []string) error {
	oldConfigFiles, err := utils.ReadRCFileFromScanArea()
	if err != nil {
		return err
	}
	// delete deleted files from record
	oldConfigFiles = utils.DeleteEdFileFromRecord(oldConfigFiles)
	scanAreaFiles := make([]string, 0)
	unAddscanAreaFiles := make([]string, 0)
	configFiles := utils.GetCurrentDirAllFiles("./")

	// get the file from the scan area and the local
	for _, newConfigFile := range configFiles {
		_, isExist := oldConfigFiles[newConfigFile]
		if !isExist {
			unAddscanAreaFiles = append(unAddscanAreaFiles, newConfigFile)
		} else {
			scanAreaFiles = append(scanAreaFiles, newConfigFile)
		}
	}

	// print format
	if len(scanAreaFiles) != 0 {
		cmd.Println("Scan area file list:")
		cmd.Println("  (use \"bk-bscp-client checkout <file>...\" to remove file from scan area)")
		cmd.Println("  (use \"bk-bscp-client commit\" to submit the configuration files in the scan area)")
		for _, addEdFile := range scanAreaFiles {
			color.Cyan("\tnew file:\t" + addEdFile)
		}
		cmd.Println()
	}

	if len(unAddscanAreaFiles) != 0 {
		cmd.Println("The local repository is not added to the scan area file list:")
		cmd.Println("  (use \"bk-bscp-client add <file>...\" to add file to scan area)")
		for _, newFile := range unAddscanAreaFiles {
			color.Red("\t" + newFile)
		}
		cmd.Println()
	}

	return nil
}
