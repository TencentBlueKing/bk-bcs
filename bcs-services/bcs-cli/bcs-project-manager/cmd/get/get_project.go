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
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
)

var (
	getProjectLong = templates.LongDesc(i18n.T(`
		Display one or many project.
		Prints a table of the most important information about the specified resources.`))

	getProjectExample = templates.Examples(i18n.T(`
		# List all project in ps output format
		kubectl-bcs-project-manager list project
		# List my project in ps output format
		kubectl-bcs-project-manager list project my-project`))

	myProjectLong = templates.LongDesc(i18n.T(`
	Display one or many my project.`))

	myProjectExample = templates.Examples(i18n.T(`
	# Display one or many my project
	kubectl-bcs-project-manager list project my-project`))
)

func listProject() *cobra.Command {
	request := new(pkg.ListProjectsRequest)
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"project", "p"},
		Short:   i18n.T("Display one or many project"),
		Long:    getProjectLong,
		Example: getProjectExample,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := pkg.NewClientWithConfiguration(context.Background()).ListProjects(request)
			if err != nil {
				klog.Infof("list variable definitions failed: %v", err)
				return
			}
			printer.PrintProjectsListInTable(flagOutput, resp)
		},
	}

	cmd.AddCommand(MyProject())

	cmd.PersistentFlags().StringVarP(&request.ProjectIDs, "project_ids", "", "",
		"The project ids that query, multiple separated by commas")
	cmd.PersistentFlags().StringVarP(&request.SearchName, "name", "", "",
		"The project chinese name, multiple separated by commas")
	cmd.PersistentFlags().StringVarP(&request.ProjectCode, "project_code", "", "",
		"Project code query")
	cmd.PersistentFlags().StringVarP(&request.Kind, "kind", "", "",
		"The cluster kind")
	cmd.PersistentFlags().Int64VarP(&request.Limit, "limit", "", 10,
		"Number of queries")
	cmd.PersistentFlags().Int64VarP(&request.Offset, "offset", "", 0,
		"Start query from offset")
	cmd.PersistentFlags().BoolVarP(&request.All, "all", "", false,
		"Get all projects, default: false")
	return cmd
}

// MyProject 授权项目
func MyProject() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "my-project",
		Aliases: []string{"mp"},
		Short:   i18n.T("Present your project list"),
		Long:    myProjectLong,
		Example: myProjectExample,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := pkg.NewClientWithConfiguration(context.Background()).ListAuthorizedProjects()
			if err != nil {
				klog.Infof("list authorized projects failed: %v", err)
				return
			}
			printer.PrinListAuthorizedProjectsInTable(flagOutput, resp)
		},
	}
	return cmd
}
