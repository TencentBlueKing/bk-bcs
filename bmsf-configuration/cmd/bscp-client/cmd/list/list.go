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

package list

import (
	"bk-bscp/cmd/bscp-client/option"

	"github.com/spf13/cobra"
)

var listCmd *cobra.Command

//init all resource create sub command
func init() {
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List resources",
		Long:  "Subcommand for list all resources",
	}
	listCmd.PersistentFlags().Int32Var(&option.GlobalOptions.Index, "index", 0, "index for list ")
	listCmd.PersistentFlags().Int32Var(&option.GlobalOptions.Limit, "limit", 100, "limit for one list command")
}

//InitCommands init all create commands
func InitCommands() []*cobra.Command {
	//init all sub resource command
	listCmd.AddCommand(listBusinessCmd())
	listCmd.AddCommand(listShardingDBCmd())
	listCmd.AddCommand(listAppCmd())
	listCmd.AddCommand(listClusterCmd())
	listCmd.AddCommand(listZoneCmd())
	listCmd.AddCommand(listConfigSetCmd())
	listCmd.AddCommand(listMultiCommitCmd())
	listCmd.AddCommand(listStrategyCmd())
	listCmd.AddCommand(listMultiReleaseCmd())
	listCmd.AddCommand(listAppInstCmd())
	return []*cobra.Command{listCmd}
}
