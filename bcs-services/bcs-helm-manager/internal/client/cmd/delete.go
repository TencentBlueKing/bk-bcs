/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var (
	deleteCMD = &cobra.Command{
		Use:   "delete",
		Short: "delete",
		Long:  "delete resource",
	}
	deleteChartCMD = &cobra.Command{
		Use:     "chart",
		Aliases: []string{"chart", "ch"},
		Short:   "delete chart",
		Long:    "delete chart",
		Run:     DeleteChart,
		Example: "helmctl delete chart -p <project_code> <chart_name>",
	}
	deleteChartVersionCMD = &cobra.Command{
		Use:     "chartVersion",
		Aliases: []string{"chart-version", "chv"},
		Short:   "delete chart version",
		Long:    "delete chart version",
		Run:     DeleteChartVersion,
		Example: "helmctl delete chartVersion -p <project_code> <chart_name> <version>",
	}
)

// DeleteRepository provide the actions to do deleteRepositoryCMD
func DeleteRepository(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("delete repository need specific repo name\n")
		os.Exit(1)
	}

	req := &helmmanager.DeleteRepositoryReq{
		ProjectCode: &flagProject,
		Name:        common.GetStringP(args[0]),
	}

	c := newClientWithConfiguration()
	if err := c.Repository().Delete(cmd.Context(), req); err != nil {
		fmt.Printf("delete repository failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to delete repository %s under project %s\n", req.GetName(), req.GetProjectCode())
}

func init() {
	deleteCMD.AddCommand(deleteChartCMD)
	deleteCMD.AddCommand(deleteChartVersionCMD)
	deleteChartCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	deleteChartVersionCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	deleteChartCMD.MarkPersistentFlagRequired("project")
	deleteChartCMD.MarkPersistentFlagRequired("repo")
	deleteChartVersionCMD.MarkPersistentFlagRequired("project")
	deleteChartVersionCMD.MarkPersistentFlagRequired("repo")
}

// DeleteChart provide the actions to do deleteChartCMD
func DeleteChart(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("delete chart need specific chart name\n")
		os.Exit(1)
	}

	if flagRepository == "" {
		flagRepository = flagProject
	}
	req := &helmmanager.DeleteChartReq{
		ProjectCode: &flagProject,
		RepoName:    &flagRepository,
		Name:        common.GetStringP(args[0]),
	}

	c := newClientWithConfiguration()
	if err := c.Chart().DeleteChart(cmd.Context(), req); err != nil {
		fmt.Printf("delete chart failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to delete chart %s\n", req.GetName())
}

// DeleteChartVersion provide the actions to do deleteChartVersionCMD
func DeleteChartVersion(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("delete chart version need specific chart name\n")
		os.Exit(1)
	}
	if len(args) == 1 {
		fmt.Printf("delete chart version need specific chart version\n")
		os.Exit(1)
	}
	if flagRepository == "" {
		flagRepository = flagProject
	}
	req := &helmmanager.DeleteChartVersionReq{
		ProjectCode: &flagProject,
		RepoName:    &flagRepository,
		Name:        common.GetStringP(args[0]),
		Version:     common.GetStringP(args[1]),
	}

	c := newClientWithConfiguration()
	if err := c.Chart().DeleteChartVersion(cmd.Context(), req); err != nil {
		fmt.Printf("delete chart version failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to delete chart version %s\n", req.GetVersion())
}
