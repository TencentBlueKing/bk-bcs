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
	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
)

//getBusinessCmd: client create business
func infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get repository init info",
		Long:  "Get current repository initialization information",
		RunE:  handleGetInfo,
	}
	return cmd
}

func handleGetInfo(cmd *cobra.Command, args []string) error {
	info, err := option.GetInitConfInfo()
	if err != nil {
		return err
	}
	utils.PrintInitInfo(info)
	cmd.Println()
	return nil
}
