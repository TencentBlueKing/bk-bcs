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

package check

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	checkLong = templates.LongDesc(i18n.T(`
	check whether a resource exists from a file or from stdin.`))

	checkExample = templates.Examples(i18n.T(`
	# check a cluster node variable
	kubectl-bcs-cluster-manager check`))

	filename string
	innerIPs []string
)

// NewCheckCmd 创建check子命令实例
func NewCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check",
		Short:   i18n.T("check whether a resource exists from a file or from stdin"),
		Long:    checkLong,
		Example: checkExample,
	}

	cmd.PersistentFlags().StringVarP(&filename, "filename", "f", "", "File address Support json file")

	// check subcommands
	cmd.AddCommand(newCheckCloudKubeConfigCmd())
	cmd.AddCommand(newCheckNodeInClusterCmd())

	return cmd
}
