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
	getChartCMD = &cobra.Command{
		Use:     "chart",
		Aliases: []string{"ct"},
		Short:   "get chart",
		Long:    "get chart",
		Run:     GetChart,
	}
	getChartVersionCMD = &cobra.Command{
		Use:     "chartversion",
		Aliases: []string{"version", "cv"},
		Short:   "get chart version",
		Long:    "get chart version",
		Run:     GetChartVersion,
	}
	getChartDetailCMD = &cobra.Command{
		Use:     "chartdetail",
		Aliases: []string{"detail", "cd"},
		Short:   "get chart detail",
		Long:    "get chart detail",
		Run:     GetChartDetail,
	}
	getReleaseCMD = &cobra.Command{
		Use:     "release",
		Aliases: []string{"rl"},
		Short:   "get release",
		Long:    "get release",
		Run:     GetRelease,
	}
)

// GetRepository provide the actions to do getRepositoryCMD
func GetRepository(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListRepositoryReq{}

	if !flagAll {
		req.Size = common.GetUint32P(uint32(flagNum))
	}
	if len(args) > 0 {
		req.Name = common.GetStringP(args[0])
		req.Size = common.GetUint32P(1)
	}
	req.ProjectID = &flagProject

	c := newClientWithConfiguration()
	r, err := c.Repository().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get repository failed, %s\n", err.Error())
		os.Exit(1)
	}

	if flagOutput == outputTypeJson {
		printer.PrintRepositoryInJson(r)
		return
	}

	printer.PrintRepositoryInTable(flagOutput == outputTypeWide, r)
}

// GetChart provide the actions to do getChartCMD
func GetChart(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListChartReq{}

	if !flagAll {
		req.Size = common.GetUint32P(uint32(flagNum))
	}
	if len(args) > 0 {
		req.Size = common.GetUint32P(1)
	}
	req.ProjectID = &flagProject
	req.Repository = &flagRepository

	c := newClientWithConfiguration()
	r, err := c.Chart().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get chart failed, %s\n", err.Error())
		os.Exit(1)
	}

	if flagOutput == outputTypeJson {
		printer.PrintChartInJson(r)
		return
	}

	printer.PrintChartInTable(flagOutput == outputTypeWide, r)
}

// GetChartVersion provide the actions to do getChartVersionCMD
func GetChartVersion(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListChartVersionReq{}

	if !flagAll {
		req.Size = common.GetUint32P(uint32(flagNum))
	}
	if len(args) == 0 {
		fmt.Printf("get chart version need specific chart name\n")
		os.Exit(1)
	}
	req.ProjectID = &flagProject
	req.Repository = &flagRepository
	req.Name = common.GetStringP(args[0])

	c := newClientWithConfiguration()
	r, err := c.Chart().Versions(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get chart version failed, %s\n", err.Error())
		os.Exit(1)
	}

	if flagOutput == outputTypeJson {
		printer.PrintChartVersionInJson(r)
		return
	}

	printer.PrintChartVersionInTable(flagOutput == outputTypeWide, r)
}

// GetChartDetail provide the actions to do getChartDetailCMD
func GetChartDetail(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("get chart detail need specific chart name\n")
		os.Exit(1)
	}
	if len(args) == 1 {
		fmt.Printf("get chart detail need specific chart version\n")
		os.Exit(1)
	}

	req := &helmmanager.GetChartDetailReq{}
	req.ProjectID = &flagProject
	req.Repository = &flagRepository
	req.Name = common.GetStringP(args[0])
	req.Version = common.GetStringP(args[1])

	c := newClientWithConfiguration()
	r, err := c.Chart().Detail(cmd.Context(), req)
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

// GetRelease provide the action to do getReleaseCMD
func GetRelease(cmd *cobra.Command, args []string) {
	req := &helmmanager.ListReleaseReq{}

	if !flagAll {
		req.Size = common.GetUint32P(uint32(flagNum))
	}
	if len(args) > 0 {
		req.Size = common.GetUint32P(1)
		req.Name = common.GetStringP(args[0])
	}
	req.ClusterID = &flagCluster
	req.Namespace = &flagNamespace

	c := newClientWithConfiguration()
	r, err := c.Release().List(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get release failed, %s\n", err.Error())
		os.Exit(1)
	}

	if flagOutput == outputTypeJson {
		printer.PrintReleaseInJson(r)
		return
	}

	printer.PrintReleaseInTable(flagOutput == outputTypeWide, r)
}

func init() {
	getCMD.AddCommand(getRepositoryCMD)
	getCMD.AddCommand(getChartCMD)
	getCMD.AddCommand(getChartVersionCMD)
	getCMD.AddCommand(getChartDetailCMD)
	getCMD.AddCommand(getReleaseCMD)
	getCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagRepository, "repository", "r", "", "repository name for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
	getCMD.PersistentFlags().StringVarP(
		&flagOutput, "output", "o", "", "output format, one of json|wide")
	getCMD.PersistentFlags().BoolVarP(&flagAll, "all", "A", false, "list all records")
	getCMD.PersistentFlags().IntVarP(&flagNum, "num", "", 20, "list records num")
}
