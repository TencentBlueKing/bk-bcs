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

package create

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
	createVariableLong = templates.LongDesc(i18n.T(`
		Create a project variable with the specified project code.`))

	createVariableExample = templates.Examples(i18n.T(`
		# Create a project variable with a command
		kubectl-bcs-project-manager create variable`))
)

func createVariable() *cobra.Command {
	request := new(pkg.CreateVariableRequest)
	cmd := &cobra.Command{
		Use:                   "variable --name=name",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create a project variable with the specified project code"),
		Long:                  createVariableLong,
		Example:               createVariableExample,
		Run: func(cmd *cobra.Command, args []string) {
			projectCode := viper.GetString("bcs.project_code")
			if len(projectCode) == 0 {
				klog.Infoln("Project code (English abbreviation), global unique, the length cannot exceed 64 characters")
				return
			}
			request.ProjectCode = projectCode
			resp, err := pkg.NewClientWithConfiguration(context.Background()).CreateVariable(request)
			if err != nil {
				klog.Infoln(err)
				return
			}
			printer.PrintInJSON(resp)
		},
	}
	cmd.Flags().StringVarP(&request.Name, "name", "n", "",
		"Variable name, length cannot exceed 32 characters, required")
	cmd.Flags().StringVarP(&request.Key, "key", "k", "",
		"The variable key, unique within the project, cannot exceed 64 characters in length, required")
	cmd.Flags().StringVarP(&request.Scope, "scope", "s", "",
		"Variable scope, value range: global, cluster, namespace")
	cmd.Flags().StringVarP(&request.Desc, "desc", "d", "",
		"Variable description and description, limited to 100 characters")
	cmd.Flags().StringVarP(&request.Default, "default", "", "",
		"Variable default value")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}
