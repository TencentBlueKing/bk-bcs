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

package cloudaccount

import "github.com/spf13/cobra"

var (
	file      string
	cloudID   string
	accountID string
)

// NewCloudAccountCmd 创建云凭证子命令实例
func NewCloudAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloudAccount",
		Short: "cloud account-related operations",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newListToPermCmd())

	return cmd
}
