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

package delete

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
	deleteClustersNamespaceLong = templates.LongDesc(i18n.T(`
		Delete project namespace by project id cluster id and name.`))

	deleteClustersNamespaceExample = templates.Examples(i18n.T(`
		# Delete a variable with project code cluster id and name
		kubectl-bcs-project-manager delete namespace --cluster-id=clusterID --name=name`))
)

func deleteClustersNamespace() *cobra.Command {
	request := new(pkg.DeleteNamespaceRequest)
	cmd := &cobra.Command{
		Use:                   "namespace --cluster-id=clusterID --name=name",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"n"},
		Short:                 i18n.T("Delete project namespace by project id cluster id and name."),
		Long:                  deleteClustersNamespaceLong,
		Example:               deleteClustersNamespaceExample,
		Run: func(cmd *cobra.Command, args []string) {
			projectCode := viper.GetString("bcs.project_code")
			if len(projectCode) == 0 {
				klog.Infoln("Project code (English abbreviation), global unique, the length cannot exceed 64 characters")
				return
			}
			request.ProjectCode = projectCode
			resp, err := pkg.NewClientWithConfiguration(context.Background()).DeleteNamespace(request)
			if err != nil {
				klog.Infof("delete project namespace failed: %v", err)
				return
			}
			printer.PrintInJSON(resp)
		},
	}

	cmd.Flags().StringVarP(&request.ClusterID, "cluster-id", "", "",
		"Cluster ID, required")
	cmd.Flags().StringVarP(&request.Name, "name", "", "",
		"namespace name, required")

	return cmd
}
