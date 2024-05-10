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
 */

package get

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
)

var (
	getProjectVariableLong = templates.LongDesc(i18n.T(`
		Display one or many project variable.
		Prints a table of the most important information about the specified resources.`))

	getProjectVariableExample = templates.Examples(i18n.T(`
		# List a project all variable in ps output format
		kubectl-bcs-project-manager list variable`))
)

func listVariable() *cobra.Command {
	request := new(pkg.ListVariableDefinitionsRequest)
	cmd := &cobra.Command{
		Use:     "variable --key=key",
		Aliases: []string{"variable", "v"},
		Short:   i18n.T("Display one or many project variable"),
		Long:    getProjectVariableLong,
		Example: getProjectVariableExample,
		Run: func(cmd *cobra.Command, args []string) {
			projectCode := viper.GetString("bcs.project_code")
			if len(projectCode) == 0 {
				klog.Infoln("Project code (English abbreviation), global unique, the length cannot exceed 64 characters")
				return
			}
			resp, err := pkg.NewClientWithConfiguration(context.Background()).ListVariableDefinitions(request, projectCode)
			if err != nil {
				klog.Infof("list variable definitions failed: %v", err)
				return
			}
			printer.PrintProjectVariablesListInTable(flagOutput, resp)
		},
	}

	cmd.Flags().StringVarP(&request.SearchKey, "key", "", "",
		"Variable key, through this field fuzzy query item variable")
	cmd.Flags().StringVarP(&request.Scope, "scope", "", "",
		"Scope query, scope value: global, cluster, namespace")
	cmd.Flags().Int64VarP(&request.Limit, "limit", "", 10,
		"Number of queries")
	cmd.Flags().Int64VarP(&request.Offset, "offset", "", 0,
		"Start query from offset")
	cmd.Flags().BoolVarP(&request.All, "all", "", false,
		"Get all variables under the project, default: false")

	return cmd
}
