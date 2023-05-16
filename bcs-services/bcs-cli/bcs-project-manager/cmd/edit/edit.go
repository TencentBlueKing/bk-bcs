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

package edit

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	editLong = templates.LongDesc(i18n.T(`
		Edit a resource from the default editor.
		The edit command allows you to directly edit any API resource you can retrieve via the
		command-line tools. It will open the editor 'vi' for Linux or 'notepad' for Windows.`))

	editExample = templates.Examples(i18n.T(`
		# Edit a project or project variable
		kubectl-bcs-project-manager edit project/variable`))
)

// NewCmdEdit 新建命令编辑
func NewCmdEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit",
		Short:   i18n.T("Edit a resource on the server"),
		Long:    editLong,
		Example: editExample,
	}

	// edit subcommands
	cmd.AddCommand(editProject())
	cmd.AddCommand(editVariable())
	cmd.AddCommand(editClustersNamespace())

	return cmd
}
