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
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/client/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/client/pkg"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/spf13/cobra"
)

var (
	flagOutput       string
	flagProject      string
	flagCluster      string
	flagNamespace    string
	flagDimension    string
	flagWorkloadType string

	outputTypeJSON = "json"
	outputTypeWide = "wide"

	getCMD = &cobra.Command{
		Use:   "get",
		Short: "get",
		Long:  "get metrics",
	}
	getProjectCMD = &cobra.Command{
		Use:     "project",
		Aliases: []string{"project", "p"},
		Short:   "get project data",
		Long:    "get project",
		Run:     GetProject,
	}
	getClusterCMD = &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"cluster", "ct"},
		Short:   "get cluster",
		Long:    "get cluster",
		Run:     GetCluster,
	}
	getNamespaceCMD = &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"namespace", "ns"},
		Short:   "get namespace",
		Long:    "get namespace",
		Run:     GetNamespace,
	}
	getWorkloadCMD = &cobra.Command{
		Use:     "workload",
		Aliases: []string{"workload", "wl"},
		Short:   "get workload",
		Long:    "get workload",
		Run:     GetWorkload,
	}
)

// GetProject get project info
func GetProject(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetProjectInfoRequest{}
	if len(args) == 0 {
		fmt.Printf("get project data need specific projectid\n")
		os.Exit(1)
	}
	req.ProjectID = args[0]
	req.Dimension = flagDimension
	client := pkg.NewClientWithConfiguration()
	rsp, err := client.GetProjectInfo(req)
	if err != nil {
		fmt.Printf("get project data err:%v\n", err)
		os.Exit(1)
	}
	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}

	if flagOutput == outputTypeJSON {
		printer.PrintProjectInJSON(rsp.Data)
		return
	}
	printer.PrintProjectInTable(flagOutput == outputTypeWide, rsp.Data)
}

// GetCluster get cluster info
func GetCluster(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetClusterInfoRequest{}
	if len(args) == 0 {
		fmt.Printf("get cluster data need specific clusterid\n")
		os.Exit(1)
	}
	req.ClusterID = args[0]
	req.Dimension = flagDimension
	client := pkg.NewClientWithConfiguration()
	rsp, err := client.GetClusterInfo(req)
	if err != nil {
		fmt.Printf("get cluster data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}

	if flagOutput == outputTypeJSON {
		printer.PrintClusterInJSON(rsp.Data)
		return
	}
	printer.PrintClusterInTable(flagOutput == outputTypeWide, rsp.Data)

}

// GetNamespace get namespace info
func GetNamespace(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetNamespaceInfoRequest{}
	if len(args) == 0 {
		fmt.Printf("get namespace data need specific namespace\n")
		os.Exit(1)
	}
	req.Namespace = args[0]
	req.ClusterID = flagCluster
	req.Dimension = flagDimension
	client := pkg.NewClientWithConfiguration()
	rsp, err := client.GetNamespaceInfo(req)
	if err != nil {
		fmt.Printf("get namespace data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}

	if flagOutput == outputTypeJSON {
		printer.PrintNamespaceInJSON(rsp.Data)
		return
	}
	printer.PrintNamespaceInTable(flagOutput == outputTypeWide, rsp.Data)

}

// GetWorkload get workload info
func GetWorkload(cmd *cobra.Command, args []string) {
	req := &bcsdatamanager.GetWorkloadInfoRequest{}
	if len(args) == 0 {
		fmt.Printf("get workload data need specific workload\n")
		os.Exit(1)
	}
	req.Namespace = flagNamespace
	req.ClusterID = flagCluster
	req.Dimension = flagDimension
	if flagWorkloadType == "" {
		fmt.Printf("get workload data need specific workloadType, use -t {workloadType}\n")
		os.Exit(1)
	}
	req.WorkloadType = flagWorkloadType
	req.WorkloadName = args[0]
	client := pkg.NewClientWithConfiguration()
	rsp, err := client.GetWorkloadInfo(req)
	if err != nil {
		fmt.Printf("get workload data err:%v\n", err)
		os.Exit(1)
	}

	if rsp != nil && rsp.Code != 0 {
		fmt.Printf(rsp.Message)
		os.Exit(1)
	}

	if flagOutput == outputTypeJSON {
		printer.PrintWorkloadInJSON(rsp.Data)
		return
	}
	printer.PrintWorkloadInTable(flagOutput == outputTypeWide, rsp.Data)
}

func init() {
	getCMD.AddCommand(getProjectCMD)
	getCMD.AddCommand(getClusterCMD)
	getCMD.AddCommand(getNamespaceCMD)
	getCMD.AddCommand(getWorkloadCMD)
	getCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagDimension, "dimension", "d", "", "release time dimension for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagWorkloadType, "workloadType", "t", "", "release workload type for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagOutput, "output", "o", "", "output format, one of json|wide")
}
