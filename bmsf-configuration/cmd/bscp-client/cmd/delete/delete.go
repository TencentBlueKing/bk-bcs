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

package delete

import (
	"github.com/spf13/cobra"
)

var deleteCmd *cobra.Command

//init all resource create sub command
func init() {
	deleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del"},
		Short:   "Delete resource",
		Long:    "Subcommand for delete all resources, including strategy",
	}
}

//business/app/cluster/zone is too simple
//no update needed

//InitCommands init all update commands
func InitCommands() []*cobra.Command {
	deleteCmd.AddCommand(deleteStrategyCmd())
	return []*cobra.Command{deleteCmd}
}
