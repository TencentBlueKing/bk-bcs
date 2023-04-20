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
	flagOutput     string
	flagAll        bool
	flagNum        = 20
	flagProject    string
	flagRepository string
	flagNamespace  string
	flagCluster    string
	flagChart      string
	flagName       string
	flagSize       = uint32(20)
	flagPage       = uint32(1)

	outputTypeJson = "json"
	outputTypeWide = "wide"

	getCMD = &cobra.Command{
		Use:   "get",
		Short: "get",
		Long:  "get resources",
	}
	getRepositoryCMD = &cobra.Command{
		Use:     "repository",
		Aliases: []string{"repo", "rp"},
		Short:   "get repository",
		Long:    "get repository",
		Run:     GetRepository,
	}
	getChartDetailCMD = &cobra.Command{
		Use:     "chartdetail",
		Aliases: []string{"detail", "cd"},
		Short:   "get chart detail",
		Long:    "get chart detail",
		Run:     GetChartDetail,
	}
	getVersionDetailCMD = &cobra.Command{
		Use:     "versiondetail",
		Aliases: []string{"versiondetail", "vd"},
		Short:   "get version detail",
		Long:    "get version detail",
		Run:     GetVersionDetail,
	}
	getReleaseDetailCMD = &cobra.Command{
		Use:     "release",
		Aliases: []string{"rl"},
		Short:   "get release detail",
		Long:    "get release detail",
		Run:     GetReleaseDetail,
	}
	getReleaseHistoryCMD = &cobra.Command{
		Use:     "releasehistory",
		Aliases: []string{"rlh"},
		Short:   "get release history",
		Long:    "get release history",
		Run:     GetReleaseHistory,
	}
)

// GetRepository provide the actions to do getRepositoryCMD
func GetRepository(cmd *cobra.Command, args []string) {
	req := &helmmanager.GetRepositoryReq{}

	req.ProjectCode = &flagProject
	req.Name = &flagRepository

	c := newClientWithConfiguration()
	r, err := c.Repository().Get(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get repository failed, %s\n", err.Error())
		os.Exit(1)
	}
	printData := []*helmmanager.Repository{r}
	if flagOutput == outputTypeJson {
		printer.PrintRepositoryInJSON(printData)
		return
	}

	printer.PrintRepositoryInTable(flagOutput == outputTypeWide, printData)
}

// GetChartDetail provide the actions to do getChartDetailCMD
func GetChartDetail(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("get chart detail need specific chart name\n")
		os.Exit(1)
	}

	req := &helmmanager.GetChartDetailV1Req{}
	req.ProjectCode = &flagProject
	req.RepoName = &flagRepository
	req.Name = common.GetStringP(args[0])

	c := newClientWithConfiguration()
	r, err := c.Chart().GetChartDetailV1(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get chart detail failed, %s\n", err.Error())
		os.Exit(1)
	}
	printData := &helmmanager.ChartListData{
		Data: []*helmmanager.Chart{r},
	}
	if flagOutput == outputTypeJson {
		printer.PrintChartInJson(printData)
		return
	}

	printer.PrintChartInTable(flagOutput == outputTypeWide, printData)
}

// GetVersionDetail provide the actions to do getVersionDetailCMD
func GetVersionDetail(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("get chart detail need specific chart name\n")
		os.Exit(1)
	}
	if len(args) == 1 {
		fmt.Printf("get chart detail need specific chart version\n")
		os.Exit(1)
	}

	req := &helmmanager.GetVersionDetailV1Req{}
	req.ProjectCode = &flagProject
	req.RepoName = &flagRepository
	req.Name = common.GetStringP(args[0])
	req.Version = common.GetStringP(args[1])

	c := newClientWithConfiguration()
	r, err := c.Chart().GetVersionDetailV1(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get chart detail failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJson {
		printer.PrintChartDetailInJson(r)
		return
	}

	printer.PrintChartDetailInTable(flagOutput == outputTypeWide, r)
}

// GetReleaseDetail provide the action to do getReleaseDetailCMD
func GetReleaseDetail(cmd *cobra.Command, args []string) {
	req := &helmmanager.GetReleaseDetailV1Req{}

	req.ProjectCode = &flagProject
	req.ClusterID = &flagCluster
	req.Namespace = &flagNamespace
	req.Name = &flagName

	c := newClientWithConfiguration()
	r, err := c.Release().GetReleaseDetail(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get release detail failed, %s\n", err.Error())
		os.Exit(1)
	}
	printData := []*helmmanager.ReleaseDetail{r}
	if flagOutput == outputTypeJson {
		printer.PrintReleaseDetailInJson(printData)
		return
	}

	printer.PrintReleaseDetailInTable(flagOutput == outputTypeWide, printData)
}

// GetReleaseHistory provide the actions to do getReleaseHistoryCMD
func GetReleaseHistory(cmd *cobra.Command, args []string) {
	req := &helmmanager.GetReleaseHistoryReq{}

	if len(args) != 1 {
		fmt.Printf("get release history need release name, rlh [name] \n")
		os.Exit(1)
	}

	req.ProjectCode = &flagProject
	req.Name = common.GetStringP(args[0])
	req.Namespace = &flagNamespace
	req.ClusterID = &flagCluster

	c := newClientWithConfiguration()
	r, err := c.Release().GetReleaseHistory(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get release history failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJson {
		printer.PrintReleaseHistoryInJson(r)
		return
	}

	printer.PrintReleaseHistoryInTable(flagOutput == outputTypeWide, r)
}

func init() {
	getCMD.AddCommand(getRepositoryCMD)
	getCMD.AddCommand(getChartDetailCMD)
	getCMD.AddCommand(getVersionDetailCMD)
	getCMD.AddCommand(getReleaseDetailCMD)
	getCMD.AddCommand(getReleaseHistoryCMD)
	getCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagRepository, "repository", "r", "", "repository name for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagName, "name", "", "", "release name for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagOutput, "output", "o", "", "output format, one of json|wide")
}
