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
	"github.com/spf13/cobra"
)

var updateCmd *cobra.Command

//init all resource create sub command
func init() {
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update resource",
		Long:  "Subcommand for update resources",
	}
}

//InitCommands init all update commands
func InitCommands() []*cobra.Command {
	updateCmd.AddCommand(updateAppCmd())
	updateCmd.AddCommand(updateZoneCmd())
	updateCmd.AddCommand(updateClusterCmd())
	updateCmd.AddCommand(updateSharingDBCmd())
	updateCmd.AddCommand(updateSharingCmd())
	return []*cobra.Command{updateCmd}
}
