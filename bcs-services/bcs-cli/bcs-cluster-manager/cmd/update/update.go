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
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	updateLong = templates.LongDesc(i18n.T(`
	update a resource from a file or from stdin.`))

	updateExample = templates.Examples(i18n.T(`
	# update a cluster variable
	kubectl-bcs-cluster-manager update`))

	desiredNode uint32
	desiredSize uint32
	filename    string
	status      string
	clusterID   string
	nodeGroupID string
	innerIPs    []string
)

// NewUpdateCmd 创建update子命令实例
func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update -f FILENAME",
		Short:   i18n.T("update a resource from a file or from stdin"),
		Long:    updateLong,
		Example: updateExample,
	}

	cmd.PersistentFlags().StringVarP(&filename, "filename", "f", "", "File address Support json file")

	// update subcommands
	cmd.AddCommand(newUpdateCloudAccountCmd())
	cmd.AddCommand(newUpdateCloudVPCCmd())
	cmd.AddCommand(newUpdateClusterCmd())
	cmd.AddCommand(newUpdateNodeCmd())
	cmd.AddCommand(newUpdateNodeGroupCmd())
	cmd.AddCommand(newUpdateGroupDesiredNodeCmd())
	cmd.AddCommand(newUpdateGroupDesiredSizeCmd())
	cmd.AddCommand(newUpdateTaskCmd())

	return cmd
}
