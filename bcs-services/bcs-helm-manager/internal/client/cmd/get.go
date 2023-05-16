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
	"context"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var (
	flagOutput       string
	flagProject      string
	flagRepository   string
	flagNamespace    string
	flagAllNamespace bool
	flagCluster      string
	flagChart        string
	flagSize         = uint32(20)
	flagPage         = uint32(1)

	outputTypeJSON = "json"
	outputTypeWide = "wide"

	getCMD = &cobra.Command{
		Use:              "get",
		Short:            "get resources(repo, chart, chart version, release)",
		TraverseChildren: true,
	}
	getRepositoryCMD = &cobra.Command{
		Use:     "repo",
		Aliases: []string{"r"},
		Short:   "get repository",
		Run:     GetRepository,
		Example: "helmctl get repo -p <project_code>",
	}
	getChartCMD = &cobra.Command{
		Use:     "chart",
		Short:   "get chart detail",
		Run:     GetChart,
		Example: "helmctl get chart -p <project_code>\nhelmctl get chart -p <project_code> -r public-repo",
	}
	getChartVersionCMD = &cobra.Command{
		Use:     "chartVersion",
		Aliases: []string{"cv"},
		Short:   "get chart version",
		Run:     GetChartVersion,
		Example: "helmctl get chartVersion -p <project_code> <chart_name>\n" +
			"helmctl get chartVersion -p <project_code> <chart_name> <version>",
	}
	getReleaseDetailCMD = &cobra.Command{
		Use:     "release",
		Aliases: []string{"rl"},
		Short:   "get release detail",
		Run:     GetRelease,
		Example: "default namespace is default\nhelmctl get release -p <project_code> -c <cluster_id>\n" +
			"helmctl get release -p <project_code> -c <cluster_id> -n <namespace>\n" +
			"helmctl get release -p <project_code> -c <cluster_id> -n <namespace> <name>\n" +
			"helmctl get release -p <project_code> -c <cluster_id> -A",
	}
)

// GetRepository provide the actions to do getRepositoryCMD
func GetRepository(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListRepositoryReq{}
	req.ProjectCode = &flagProject

	c := newClientWithConfiguration()
	r, err := c.Repository().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get repository failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(r)
		return
	}

	printer.PrintRepositoryInTable(flagOutput == outputTypeWide, r)
}

// GetChart provide the actions to list chart
func GetChart(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		GetChartDetail(cmd.Context(), args[0])
		return
	}

	req := &helmmanager.ListChartV1Req{}
	req.Page = common.GetUint32P(1)
	req.Size = common.GetUint32P(10000)
	req.ProjectCode = &flagProject
	if flagRepository == "" {
		flagRepository = *req.ProjectCode
	}
	req.RepoName = &flagRepository
	c := newClientWithConfiguration()
	r, err := c.Chart().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get chart failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(r)
		return
	}
	printer.PrintChartInTable(flagOutput == outputTypeWide, r)
}

// GetChartDetail provide the actions to get chart
func GetChartDetail(ctx context.Context, chartName string) {
	req := &helmmanager.GetChartDetailV1Req{}
	req.ProjectCode = &flagProject
	if flagRepository == "" {
		flagRepository = *req.ProjectCode
	}
	req.RepoName = &flagRepository
	req.Name = &chartName

	c := newClientWithConfiguration()
	r, err := c.Chart().GetChartDetail(ctx, req)
	if err != nil {
		fmt.Printf("get chart detail failed, %s\n", err.Error())
		os.Exit(1)
	}
	printData := &helmmanager.ChartListData{
		Data: []*helmmanager.Chart{r},
	}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(printData)
		return
	}

	printer.PrintChartInTable(flagOutput == outputTypeWide, printData)
}

// GetChartVersion provide the actions to do getVersionDetailCMD
func GetChartVersion(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("get chart version need specific chart name\n")
		os.Exit(1)
	}
	if len(args) == 2 {
		GetVersionDetail(cmd.Context(), args[0], args[1])
		return
	}
	req := &helmmanager.ListChartVersionV1Req{}
	req.Page = common.GetUint32P(1)
	req.Size = common.GetUint32P(10000)
	req.ProjectCode = &flagProject
	if flagRepository == "" {
		flagRepository = *req.ProjectCode
	}
	req.RepoName = &flagRepository
	req.Name = common.GetStringP(args[0])

	c := newClientWithConfiguration()
	r, err := c.Chart().Versions(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get chart version failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(r)
		return
	}

	printer.PrintChartVersionInTable(flagOutput == outputTypeWide, r)
}

// GetVersionDetail provide the actions to do getVersionDetailCMD
func GetVersionDetail(ctx context.Context, chartName, version string) {
	req := &helmmanager.GetVersionDetailV1Req{}
	req.ProjectCode = &flagProject
	if flagRepository == "" {
		flagRepository = *req.ProjectCode
	}
	req.RepoName = &flagRepository
	req.Name = &chartName
	req.Version = &version

	c := newClientWithConfiguration()
	r, err := c.Chart().GetVersionDetail(ctx, req)
	if err != nil {
		fmt.Printf("get chart version failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(r)
		return
	}

	printer.PrintChartDetailInTable(flagOutput == outputTypeWide, r)
}

// GetRelease provide the action to do getReleaseDetailCMD
func GetRelease(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		GetReleaseDetail(cmd.Context(), args[0])
		return
	}
	req := &helmmanager.ListReleaseV1Req{
		ProjectCode: &flagProject,
		ClusterID:   &flagCluster,
		Page:        common.GetUint32P(1),
		Size:        common.GetUint32P(10000),
	}
	if flagNamespace == "" {
		flagNamespace = "default"
	}
	if flagAllNamespace {
		req.Namespace = common.GetStringP("")
	} else {
		req.Namespace = &flagNamespace
	}

	c := newClientWithConfiguration()
	r, err := c.Release().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get release detail failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(r)
		return
	}

	printer.PrintReleaseInTable(flagOutput == outputTypeWide, r)
}

// GetReleaseDetail provide the action to do getReleaseDetailCMD
func GetReleaseDetail(ctx context.Context, name string) {
	req := &helmmanager.GetReleaseDetailV1Req{}
	req.ProjectCode = &flagProject
	req.ClusterID = &flagCluster
	req.Namespace = &flagNamespace
	req.Name = &name

	c := newClientWithConfiguration()
	r, err := c.Release().GetReleaseDetail(ctx, req)
	if err != nil {
		fmt.Printf("get release detail failed, %s\n", err.Error())
		os.Exit(1)
	}
	printData := []*helmmanager.ReleaseDetail{r}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(printData)
		return
	}

	printer.PrintReleaseDetailInTable(flagOutput == outputTypeWide, printData)
}

func init() {
	initGetRepoCMD()
	initGetChartCMD()
	initGetChartVersionCMD()
	inintGetReleaseCMD()
}

func initGetRepoCMD() {
	getRepositoryCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	getRepositoryCMD.MarkPersistentFlagRequired("project")
	getCMD.AddCommand(getRepositoryCMD)
}

func initGetChartCMD() {
	getChartCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	getChartCMD.PersistentFlags().StringVarP(
		&flagRepository, "repo", "r", "", "repo name")
	getChartCMD.MarkPersistentFlagRequired("project")
	getCMD.AddCommand(getChartCMD)
}

func initGetChartVersionCMD() {
	getChartVersionCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	getChartVersionCMD.PersistentFlags().StringVarP(
		&flagRepository, "repo", "r", "", "repo name")
	getChartVersionCMD.MarkPersistentFlagRequired("project")
	getCMD.AddCommand(getChartVersionCMD)
}

func inintGetReleaseCMD() {
	getCMD.AddCommand(getReleaseDetailCMD)
	getReleaseDetailCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	getReleaseDetailCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "c", "", "release cluster id")
	getReleaseDetailCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace")
	getReleaseDetailCMD.PersistentFlags().BoolVarP(
		&flagAllNamespace, "all-namespace", "A", false, "list all namespace")
	getReleaseDetailCMD.MarkPersistentFlagRequired("project")
	getReleaseDetailCMD.MarkPersistentFlagRequired("cluster")
}
