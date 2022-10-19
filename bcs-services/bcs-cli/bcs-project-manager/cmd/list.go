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

package cmd

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

func newListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "",
		Long:  "list infos from bcs-project-manager",
	}
	listCmd.AddCommand(listProject())
	return listCmd
}

func listProject() *cobra.Command {
	var all bool
	request := new(bcsproject.ListProjectsRequest)
	subCmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"project", "p"},
		Short:   "",
		Long:    "list projects info with full-data or paging support",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
			if err != nil {
				klog.Fatalf("init client failed: %v", err.Error())
			}
			resp, err := client.ListProjects(cliCtx, request)
			if err != nil {
				klog.Fatalf("list projects failed: %v", err)
			}
			if resp != nil && resp.Code != 0 {
				klog.Fatal("list projects response code not 0 but %d: %s", resp.Code, resp.Message)
			}
			printer.PrintProjectsListInTable(flagOutput, resp)
		},
	}

	subCmd.PersistentFlags().StringVarP(&request.ProjectIDs, "project_ids", "", "",
		"the project ids that query, multiple separated by commas")
	subCmd.PersistentFlags().StringVarP(&request.Names, "names", "", "",
		"the project chinese name, multiple separated by commas")
	subCmd.PersistentFlags().StringVarP(&request.ProjectCode, "project_code", "", "",
		"project code query")
	subCmd.PersistentFlags().StringVarP(&request.SearchName, "search_name", "", "",
		"project name used to fuzzy query")
	subCmd.PersistentFlags().StringVarP(&request.Kind, "kind", "", "",
		"the cluster kind")
	subCmd.PersistentFlags().Int64VarP(&request.Limit, "limit", "", 10,
		"number of queries")
	subCmd.PersistentFlags().Int64VarP(&request.Offset, "offset", "", 0,
		"start query from offset")
	subCmd.PersistentFlags().BoolVarP(&all, "all", "", false,
		"get all projects, default: false")
	return subCmd
}

//var (
//	listParam bcsproject.ListProjectsRequest
//	all       bool
//	listCmd   = &cobra.Command{
//		Use:   "list",
//		Short: "list infos from bcs-project-manager",
//		Long:  "list metrics",
//	}
//	listProjectCmd = &cobra.Command{
//		Use:     "project",
//		Aliases: []string{"project", "p"},
//		Short:   "list projects info with full-data or paging support",
//		Long:    "list project",
//		Run:     ListProject,
//	}
//)

//func init() {
//	listCmd.AddCommand(listProjectCmd)
//	listCmd.PersistentFlags().StringVarP(&listParam.ProjectIDs, "project_ids", "", "",
//		"the project ids that query, multiple separated by commas")
//	listCmd.PersistentFlags().StringVarP(&listParam.Names, "names", "", "",
//		"the project chinese name, multiple separated by commas")
//	listCmd.PersistentFlags().StringVarP(&listParam.ProjectCode, "project_code", "", "",
//		"project code query")
//	listCmd.PersistentFlags().StringVarP(&listParam.SearchName, "search_name", "", "",
//		"project name used to fuzzy query")
//	listCmd.PersistentFlags().StringVarP(&listParam.Kind, "kind", "", "",
//		"the cluster kind")
//	listCmd.PersistentFlags().Int64VarP(&listParam.Limit, "limit", "", 10,
//		"number of queries")
//	listCmd.PersistentFlags().Int64VarP(&listParam.Offset, "offset", "", 0,
//		"start query from offset")
//	listCmd.PersistentFlags().BoolVarP(&all, "all", "", false,
//		"get all projects, default: false")
//}
//
//func ListProject(cmd *cobra.Command, args []string) {
//	request := new(bcsproject.ListProjectsRequest)
//	request.ProjectIDs = listParam.ProjectIDs
//	request.Names = listParam.Names
//	request.Kind = listParam.Kind
//	request.ProjectCode = listParam.ProjectCode
//	request.SearchName = listParam.SearchName
//	isOfflineB := &wrappers.BoolValue{}
//	isOfflineB.Value = true
//	request.All = all
//	if !all {
//		request.Limit = listParam.Limit
//		request.Offset = listParam.Offset
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
//	if err != nil {
//		klog.Fatalf("init client failed: %v", err.Error())
//	}
//	resp, err := client.ListProjects(cliCtx, request)
//	if err != nil {
//		klog.Fatalf("list projects failed: %v", err)
//	}
//	if resp != nil && resp.Code != 0 {
//		klog.Fatal("list projects response code not 0 but %d: %s", resp.Code, resp.Message)
//	}
//	printer.PrintProjectsListInTable(flagOutput, resp)
//}
