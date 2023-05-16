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

package nodegroup

import (
	"github.com/spf13/cobra"
)

var (
	file        string
	clusterID   string
	nodeGroupID string
	nodes       []string
	desiredNode uint32
	desiredSize uint32
)

// NewNodeGroupCmd 创建节点池子命令实例
func NewNodeGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodeGroup",
		Short: "node group-related operations",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newMoveNodesCmd())
	cmd.AddCommand(newRemoveNodesCmd())
	cmd.AddCommand(newCleanNodesCmd())
	cmd.AddCommand(newCleanNodesV2Cmd())
	cmd.AddCommand(newListNodesCmd())
	cmd.AddCommand(newUpdateDesiredNodeCmd())
	cmd.AddCommand(newUpdateDesiredSizeCmd())
	cmd.AddCommand(newEnableAutoScaleCmd())
	cmd.AddCommand(newDisableAutoScaleCmd())

	return cmd
}
