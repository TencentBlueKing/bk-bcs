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
	flagOutput string
	getLong    = templates.LongDesc(i18n.T(`
		Display one or many from bcs-project-manager.
		Prints a table of the most important information about the specified resources.`))

	getExample = templates.Examples(i18n.T(`
		# List all project in ps output format
		kubectl-bcs-project-manager list project
		# List all project variable in ps output format
		kubectl-bcs-project-manager list variable
		# List all project clusters namespaces. in ps output format
		kubectl-bcs-project-manager list namespaces`))
)

// NewCmdGet 新建命令Get
func NewCmdGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   i18n.T("List infos from bcs-project-manager"),
		Long:    getLong,
		Example: getExample,
	}

	cmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "wide",
		"optional parameter: json/wide/yaml, output json or yaml to stdout")

	// get subcommands
	cmd.AddCommand(listProject())
	cmd.AddCommand(listVariable())
	cmd.AddCommand(listClustersNamespace())

	return cmd
}
