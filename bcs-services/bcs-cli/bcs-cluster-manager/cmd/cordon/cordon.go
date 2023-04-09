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

package cordon

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	cordonLong = templates.LongDesc(i18n.T(`
	mark node as unschedulable.`))

	cordonExample = templates.Examples(i18n.T(`
	# mark node "foo" as unschedulable
	kubectl-bcs-cluster-manager cordon`))

	clusterID string
	innerIPs  []string
)

// NewCordonCmd 创建cordon子命令实例
func NewCordonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cordon",
		Short:   i18n.T("mark node as unschedulable."),
		Long:    cordonLong,
		Example: cordonExample,
	}

	// cordon subcommands
	cmd.AddCommand(newCordonNodeCmd())

	return cmd
}
