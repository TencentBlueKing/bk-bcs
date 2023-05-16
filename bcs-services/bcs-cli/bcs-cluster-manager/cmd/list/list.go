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
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	listLong = templates.LongDesc(i18n.T(`
	display many resources.
	prints a table of the most important information about the specified resources.`))

	listExample = templates.Examples(i18n.T(`
	# list a cluster variable
	kubectl-bcs-cluster-manager list`))

	offset      uint32
	limit       uint32
	cloudID     string
	networkType string
	clusterID   string
	nodeGroupID string
	projectID   string
	flagOutput  string
)

// NewListCmd 创建list子命令实例
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   i18n.T("display many resources"),
		Long:    listLong,
		Example: listExample,
	}

	cmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "wide",
		"optional parameter: json/wide/yaml, output json or yaml to stdout")

	// list subcommands
	cmd.AddCommand(newListCloudAccountCmd())
	cmd.AddCommand(newListCloudAccountToPermCmd())
	cmd.AddCommand(newListCloudRegionsCmd())
	cmd.AddCommand(newListCloudVPCCmd())
	cmd.AddCommand(newListCommonClusterCmd())
	cmd.AddCommand(newListNodesInClusterCmd())
	cmd.AddCommand(newListClusterCmd())
	cmd.AddCommand(newListNodesInGroupCmd())
	cmd.AddCommand(newListNodeGroupCmd())
	cmd.AddCommand(newListTaskCmd())

	return cmd
}
