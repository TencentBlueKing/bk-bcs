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
	"bk-bscp/cmd/bscp-client/service"
)

var checkout *cobra.Command

// checkout file from the scan area
func checkoutCmd() *cobra.Command {
	checkout = &cobra.Command{
		Use:     "checkout",
		Aliases: []string{"co"},
		Short:   "Checkout the file from the scan area",
		Long:    "Checkout the file from the scan area, the file will not be submitted when commit",
		Example: `
	bk-bscp-client checkout .
	bk-bscp-client checkout etc/local.yaml etc/server.yaml
		`,
		RunE: handCheckoutConfigFile,
	}
	return checkout
}

func handCheckoutConfigFile(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Nothing specified, nothing checkout.\nMaybe you wanted to say 'bk-bscp-client checkout .?\n")
	}
	var coFiles []string
	recordConfigFiles := make(map[string]service.ConfigFile)

	// if . only need write empty struct to record file
	if args[0] == "." && len(args) == 1 {
		// pass
	} else {
		// get co files list
		for _, filePath := range args {
			// check path
			if !utils.IsExists(filePath) {
				return fmt.Errorf("Pathspec '%s' did not match any files\n", filePath)
			}

			// check is dir
			if utils.IsDir(filePath) {
				files := utils.GetCurrentDirAllFiles(filePath)
				for _, file := range files {
					coFiles = append(coFiles, file)
				}
			} else {
				coFiles = append(coFiles, filePath)
			}
		}

		// get .bscp/record
		recordConfigFiles, err := utils.ReadRCFileFromScanArea()
		if err != nil {
			return err
		}

		// delete deleted files from record
		recordConfigFiles = utils.DeleteEdFileFromRecord(recordConfigFiles)

		// delete checkout file list
		for _, roFile := range coFiles {
			delete(recordConfigFiles, roFile)
		}
	}

	err := utils.WriteRCFileToScanArea(recordConfigFiles)
	if err != nil {
		return err
	}
	return nil
}
