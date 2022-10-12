/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-data-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-data-manager/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"

	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/spf13/cobra"
)

var (
	flagPage    uint32
	flagSize    uint32
	allClusters bool

	listCMD = &cobra.Command{
		Use:   "list",
		Short: "list",
		Long:  "list metrics",
	}
	listProjectCMD = &cobra.Command{
		Use:     "project",
		Aliases: []string{"project", "p"},
		Short:   "list project",
		Long:    "list project",
		Run:     ListProject,
	}
	listClusterCMD = &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"cluster", "ct"},
		Short:   "list cluster",
		Long:    "list cluster",
		Run:     ListCluster,
	}
	listNamespaceCMD = &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"namespace", "ns"},
		Short:   "list namespace",
		Long:    "list namespace",
		Run:     ListNamespace,
	}
	listWorkloadCMD = &cobra.Command{
		Use:     "workload",
		Aliases: []string{"workload", "wl"},
		Short:   "list workload",
		Long:    "list workload",
		Run:     ListWorkload,
	}
	listPodAutoscalerCMD = &cobra.Command{
		Use:     "podAutoscaler",
		Aliases: []string{"podAutoscaler", "pa"},
		Short:   "list podAutoscaler",
		Long:    "list podAutoscaler",
		Run:     ListPodAutoscaler,
	}
)

// ListProject list projects
func ListProject(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetAllProjectListRequest{}
	req.Dimension = flagDimension
	req.Page = flagPage
	req.Size = flagSize
	ctx := context.Background()
	client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
	if err != nil {
		fmt.Printf("init datamanger conn error:%v\n", err)
		os.Exit(1)
	}
	rsp, err := client.GetAllProjectList(cliCtx, req)
	if err != nil {
		fmt.Printf("get project list data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}
	fmt.Printf("total: %d\n", rsp.Total)
	if flagOutput == outputTypeJSON {
		printer.PrintProjectListInJSON(rsp.Data)
		return
	}
	printer.PrintProjectListInTable(flagOutput == outputTypeWide, rsp.Data)
}

// ListCluster list cluster info
func ListCluster(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetClusterListRequest{}
	req.Project = flagProject
	req.Business = flagBusinessID
	req.Dimension = flagDimension
	req.ProjectCode = flagProjectCode
	req.Page = flagPage
	req.Size = flagSize
	ctx := context.Background()
	client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
	if err != nil {
		fmt.Printf("init datamanger conn error:%v\n", err)
		os.Exit(1)
	}
	rsp := &bcsdatamanager.GetClusterListResponse{}
	if allClusters {
		rsp, err = client.GetAllClusterList(cliCtx, req)
	} else {
		rsp, err = client.GetClusterListByProject(cliCtx, req)
	}
	if err != nil {
		fmt.Printf("get cluster list data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}
	fmt.Printf("total: %d\n", rsp.Total)
	if flagOutput == outputTypeJSON {
		printer.PrintClusterListInJSON(rsp.Data)
		return
	}
	printer.PrintClusterListInTable(flagOutput == outputTypeWide, rsp.Data)
}

// ListNamespace list namespace info
func ListNamespace(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetNamespaceInfoListRequest{}
	req.ClusterID = flagCluster
	req.Dimension = flagDimension
	req.Page = flagPage
	req.Size = flagSize
	ctx := context.Background()
	client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
	if err != nil {
		fmt.Printf("init datamanger conn error:%v\n", err)
		os.Exit(1)
	}
	rsp, err := client.GetNamespaceInfoList(cliCtx, req)
	if err != nil {
		fmt.Printf("get namespace list data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}
	fmt.Printf("total: %d\n", rsp.Total)
	if flagOutput == outputTypeJSON {
		printer.PrintNamespaceListInJSON(rsp.Data)
		return
	}
	printer.PrintNamespaceListInTable(flagOutput == outputTypeWide, rsp.Data)
}

// ListWorkload list workload list info
func ListWorkload(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetWorkloadInfoListRequest{}
	req.ClusterID = flagCluster
	req.Dimension = flagDimension
	req.Page = flagPage
	req.Size = flagSize
	req.Namespace = flagNamespace
	req.WorkloadType = flagWorkloadType
	ctx := context.Background()
	client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
	if err != nil {
		fmt.Printf("init datamanger conn error:%v\n", err)
		os.Exit(1)
	}
	rsp, err := client.GetWorkloadInfoList(cliCtx, req)
	if err != nil {
		fmt.Printf("get workload list data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}
	fmt.Printf("total: %d\n", rsp.Total)
	if flagOutput == outputTypeJSON {
		printer.PrintWorkloadListInJSON(rsp.Data)
		return
	}
	printer.PrintWorkloadListInTable(flagOutput == outputTypeWide, rsp.Data)
}

// ListPodAutoscaler list pod autoscaler list info
func ListPodAutoscaler(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetPodAutoscalerListRequest{}
	req.Business = flagBusinessID
	req.Project = flagProject
	req.ClusterID = flagCluster
	req.Dimension = flagDimension
	req.Page = flagPage
	req.Size = flagSize
	req.Namespace = flagNamespace
	if flagAutoscalerType != "" {
		switch flagAutoscalerType {
		case "hpa":
			req.PodAutoscalerType = types.HPAType
		case "gpa":
			req.PodAutoscalerType = types.GPAType
		default:
			fmt.Printf("wrong autoscaler type, use hpa/gpa")
			os.Exit(1)
		}
	}
	ctx := context.Background()
	client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
	if err != nil {
		fmt.Printf("init datamanger conn error:%v\n", err)
		os.Exit(1)
	}
	rsp, err := client.GetPodAutoscalerList(cliCtx, req)
	if err != nil {
		fmt.Printf("get pod autoscaler list data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}
	fmt.Printf("total: %d\n", rsp.Total)
	if flagOutput == outputTypeJSON {
		printer.PrintAutoscalerListInJSON(rsp.Data)
		return
	}
	printer.PrintAutoscalerListInTable(flagOutput == outputTypeWide, rsp.Data)
}

func init() {
	listCMD.AddCommand(listProjectCMD)
	listCMD.AddCommand(listClusterCMD)
	listCMD.AddCommand(listNamespaceCMD)
	listCMD.AddCommand(listWorkloadCMD)
	listCMD.AddCommand(listPodAutoscalerCMD)
	listCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagProjectCode, "projectCode", "", "", "project code for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagBusinessID, "business", "", "", "business id for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
	listCMD.PersistentFlags().BoolVarP(
		&allClusters, "all-clusters", "", false, "If true, get all clusters")
	listCMD.PersistentFlags().StringVarP(
		&flagWorkloadType, "workloadType", "t", "", "release workload type for operation, Deployment, ")
	listCMD.PersistentFlags().StringVarP(
		&flagOutput, "output", "o", "", "output format, one of json|wide")
	listCMD.PersistentFlags().Uint32VarP(
		&flagPage, "page", "", 0, "list page")
	listCMD.PersistentFlags().Uint32VarP(
		&flagSize, "size", "", 0, "list size")
	listCMD.PersistentFlags().StringVarP(
		&flagDimension, "dimension", "d", "", "release dimension for operation")
}
