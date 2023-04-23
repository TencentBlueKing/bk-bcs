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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var (
	listCMD = &cobra.Command{
		Use:   "list",
		Short: "list",
		Long:  "list resources",
	}

	listRepositoryCMD = &cobra.Command{
		Use:     "repository",
		Aliases: []string{"repo", "rp"},
		Short:   "list repository",
		Long:    "list repository",
		Run:     ListRepository,
	}

	listChartVersionCMD = &cobra.Command{
		Use:     "chartversion",
		Aliases: []string{"version", "cv"},
		Short:   "list chart version",
		Long:    "list chart version",
		Run:     ListChartVersion,
	}

	listChartCMD = &cobra.Command{
		Use:     "chartv1",
		Aliases: []string{"ch"},
		Short:   "list chart",
		Long:    "list chart",
		Run:     ListChart,
	}

	listReleaseCMD = &cobra.Command{
		Use:     "release",
		Aliases: []string{"rl"},
		Short:   "list release",
		Long:    "list release",
		Run:     ListRelease,
	}
)

func init() {
	listCMD.AddCommand(listRepositoryCMD)
	listCMD.AddCommand(listChartVersionCMD)
	listCMD.AddCommand(listChartCMD)
	listCMD.AddCommand(listReleaseCMD)

	listCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagRepository, "repository", "r", "", "repository name for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagName, "name", "", "", "release name for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
	listCMD.PersistentFlags().StringVarP(
		&flagOutput, "output", "o", "", "output format, one of json|wide")
	listCMD.PersistentFlags().BoolVarP(&flagAll, "all", "A", false, "list all records")
	listCMD.PersistentFlags().IntVarP(&flagNum, "num", "", 20, "list records num")

}

// ListRepository provide the actions to do listRepositoryCMD
func ListRepository(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListRepositoryReq{}

	req.ProjectCode = &flagProject

	c := newClientWithConfiguration()
	r, err := c.Repository().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("list repository failed, %s\n", err.Error())
		os.Exit(1)
	}

	if flagOutput == outputTypeJson {
		printer.PrintRepositoryInJSON(r)
		return
	}

	printer.PrintRepositoryInTable(flagOutput == outputTypeWide, r)
}

// ListChartVersion provide the actions to do listChartVersionCMD
func ListChartVersion(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListChartVersionReq{}

	if !flagAll {
		req.Size = common.GetUint32P(uint32(flagNum))
	}
	if len(args) == 0 {
		fmt.Printf("list chart version need specific chart name\n")
		os.Exit(1)
	}
	req.ProjectID = &flagProject
	req.Repository = &flagRepository
	req.Name = common.GetStringP(args[0])

	c := newClientWithConfiguration()
	r, err := c.Chart().Versions(cmd.Context(), req)
	if err != nil {
		fmt.Printf("list chart version failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJson {
		printer.PrintChartVersionInJson(r)
		return
	}
	printer.PrintChartVersionInTable(flagOutput == outputTypeWide, r)
}

// ListChart provide the actions to do listChartCMD
func ListChart(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListChartV1Req{}
	req.Page = common.GetUint32P(1)
	req.Size = common.GetUint32P(10000)
	req.ProjectCode = &flagProject
	req.RepoName = &flagRepository
	req.Name = &flagName
	c := newClientWithConfiguration()
	r, err := c.Chart().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get chartv1 failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJson {
		printer.PrintChartInJson(r)
		return
	}
	printer.PrintChartInTable(flagOutput == outputTypeWide, r)
}

// ListRelease provide the action to do listReleaseCMD
func ListRelease(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListReleaseV1Req{}

	req.ProjectCode = &flagProject
	req.ClusterID = &flagCluster
	req.Namespace = &flagNamespace
	req.Page = common.GetUint32P(1)
	req.Size = common.GetUint32P(10000)

	c := newClientWithConfiguration()
	r, err := c.Release().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("list release v1 failed, %s\n", err.Error())
		os.Exit(1)
	}

	if flagOutput == outputTypeJson {
		printer.PrintReleaseInJson(r)
		return
	}

	printer.PrintReleaseInTable(flagOutput == outputTypeWide, r)
}
