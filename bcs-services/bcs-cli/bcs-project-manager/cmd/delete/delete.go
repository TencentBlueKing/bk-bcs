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
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	deleteLong = templates.LongDesc(i18n.T(`
		Delete resources by stdin, resources and names, or by resources and label selector.`))

	deleteExample = templates.Examples(i18n.T(`
		# Delete one or many resources
		kubectl-bcs-project-manager delete`))
)

// NewCmdDelete 新建命令删除
func NewCmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete ( variable )",
		Short:   i18n.T("Delete resources by stdin, resources and names, or by resources and label selector"),
		Long:    deleteLong,
		Example: deleteExample,
	}

	// delete subcommands
	cmd.AddCommand(deleteVariable())
	cmd.AddCommand(deleteClustersNamespace())

	return cmd
}
