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

package get

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getLong = templates.LongDesc(i18n.T(`
	display one resources.
	Prints a table of the most important information about the specified resources.`))

	getExample = templates.Examples(i18n.T(`
	# get a cluster variable
	kubectl-bcs-cluster-manager get`))

	vpcID       string
	clusterID   string
	innerIP     string
	nodeGroupID string
	taskID      string
)

// NewGetCmd 创建get子命令实例
func NewGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   i18n.T("display one resources."),
		Long:    getLong,
		Example: getExample,
	}

	// get subcommands
	cmd.AddCommand(newGetVPCCidrCmd())
	cmd.AddCommand(newGetClusterCmd())
	cmd.AddCommand(newGetNodeCmd())
	cmd.AddCommand(newGetNodeGroupCmd())
	cmd.AddCommand(newGetTaskCmd())

	return cmd
}
