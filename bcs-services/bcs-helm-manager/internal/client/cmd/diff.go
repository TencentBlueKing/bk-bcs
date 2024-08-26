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

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/databus23/helm-diff/v3/diff"
	"github.com/databus23/helm-diff/v3/manifest"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// nolint goconst
var (
	diffCMD = &cobra.Command{
		Use:   "diff",
		Short: "diff",
		Long:  "helmctl diff",
	}
	diffRevisionCMD = &cobra.Command{
		Use:   "revision",
		Short: "helmctl diff revision",
		Long:  "helmctl diff revision",
		Run:   diffRevision,
		Example: "helmctl diff revision -p <project_code> -c <cluster_id> -n <namespace> " +
			"<release_name> <revision1> <revision2>",
	}
	diffRollbackCMD = &cobra.Command{
		Use:     "rollback",
		Short:   "helmctl diff rollback",
		Long:    "helmctl diff rollback",
		Run:     diffRollback,
		Example: "helmctl diff rollback -p <project_code> -c <cluster_id> -n <namespace> <release_name> <revision>",
	}
	diffUpgradeCMD = &cobra.Command{
		Use:   "upgrade",
		Short: "helmctl diff upgrade",
		Long:  "helmctl diff upgrade",
		Run:   diffUpgrade,
		Example: "helmctl diff upgrade -p <project_code> -c <cluster_id> -n <namespace> <release_name> <chart_name> " +
			"<version> -f values.yaml",
	}
)

// diffRevision provide the actions to do diffRevision
func diffRevision(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Printf("diff revision only need specific release_name revision1 revision2\n")
		os.Exit(1)
	}

	revision1, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		fmt.Printf("diff revision failed, parse revision1 %s, failed: %s\n", args[1], err.Error())
		os.Exit(1)
	}

	// 没有为0的revision
	if revision1 == 0 {
		return
	}

	revision2, err := strconv.ParseUint(args[2], 10, 32)
	if err != nil {
		fmt.Printf("diff revision failed, parse revision2 %s, failed: %s\n", args[2], err.Error())
		os.Exit(1)
	}

	if revision2 == 0 {
		return
	}

	uint32Revision1 := uint32(revision1)
	uint32Revision2 := uint32(revision2)
	req := helmmanager.GetReleaseManifestReq{
		ProjectCode: &flagProject,
		ClusterID:   &flagCluster,
		Namespace:   &flagNamespace,
		Name:        common.GetStringP(args[0]),
		Revision:    &uint32Revision1,
	}

	c := newClientWithConfiguration()
	releasePreview1, err := c.Release().GetReleaseManifest(cmd.Context(), &req)
	if err != nil {
		fmt.Printf("get release manifest failed, %s\n", err.Error())
		os.Exit(1)
	}

	req.Revision = &uint32Revision2

	releasePreview2, err := c.Release().GetReleaseManifest(cmd.Context(), &req)
	if err != nil {
		fmt.Printf("get release manifest failed, %s\n", err.Error())
		os.Exit(1)
	}

	diffRevisionContent(releasePreview1, releasePreview2)
}

// diffRollback provide the actions to do diffRollback
func diffRollback(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Printf("diff rollback only need specific release_name revision\n")
		os.Exit(1)
	}

	revision, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		fmt.Printf("diff rollback failed, parse revision %s, failed: %s\n", args[1], err.Error())
		os.Exit(1)
	}

	if revision == 0 {
		return
	}

	uint32Revision := uint32(revision)
	req := helmmanager.ReleasePreviewReq{
		ProjectCode: &flagProject,
		ClusterID:   &flagCluster,
		Namespace:   &flagNamespace,
		Name:        common.GetStringP(args[0]),
		Revision:    &uint32Revision,
	}

	c := newClientWithConfiguration()
	releasePreview, err := c.Release().ReleasePreview(cmd.Context(), &req)
	if err != nil {
		fmt.Printf("release preview failed, %s\n", err.Error())
		os.Exit(1)
	}

	diffReleaseContent(*releasePreview.OldContent, *releasePreview.NewContent)
}

// diffUpgrade provide the actions to do diffUpgrade
func diffUpgrade(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Printf("diff upgrade only need specific release_name chart_name version\n")
		os.Exit(1)
	}

	values, err := getValues()
	if err != nil {
		fmt.Printf("read values file failed, %s\n", err.Error())
		os.Exit(1)
	}

	req := helmmanager.ReleasePreviewReq{
		ProjectCode: &flagProject,
		ClusterID:   &flagCluster,
		Namespace:   &flagNamespace,
		Name:        common.GetStringP(args[0]),
		Repository:  new(string),
		Chart:       common.GetStringP(args[1]),
		Version:     common.GetStringP(args[2]),
		Values:      values,
		Args:        flagArgs,
	}

	if flagRepository == "" {
		flagRepository = flagProject
	}
	req.Repository = &flagRepository

	c := newClientWithConfiguration()
	releasePreview, err := c.Release().ReleasePreview(cmd.Context(), &req)
	if err != nil {
		fmt.Printf("release preview failed, %s\n", err.Error())
		os.Exit(1)
	}

	diffReleaseContent(*releasePreview.OldContent, *releasePreview.NewContent)
}

// diffRevisionContent provide the actions to do diff revision content
func diffRevisionContent(releaseResponse1, releaseResponse2 map[string]*helmmanager.FileContent) {
	oldIndex := getManifestMappingResult(releaseResponse1)
	newIndex := getManifestMappingResult(releaseResponse2)
	options := diff.Options{
		OutputFormat:  "diff",
		OutputContext: -1,
	}
	_ = diff.Manifests(oldIndex, newIndex, &options, os.Stdout)
}

// diffReleaseContent provide the actions to do diff release content
func diffReleaseContent(releaseResponse1, releaseResponse2 string) {
	excludes := []string{"test", "test-success"}
	options := diff.Options{
		OutputFormat:  "diff",
		OutputContext: -1,
	}
	_ = diff.Manifests(
		manifest.Parse(releaseResponse1, "", false, excludes...),
		manifest.Parse(releaseResponse2, "", false, excludes...),
		&options,
		os.Stdout)
}

// helmmanager.FileContent transform to manifest.MappingResult
func getManifestMappingResult(releaseResponse map[string]*helmmanager.FileContent) map[string]*manifest.MappingResult {
	oldIndex := make(map[string]*manifest.MappingResult, len(releaseResponse))
	for key, data := range releaseResponse {
		mappingResult := manifest.MappingResult{
			Name:    *data.Name,
			Content: *data.Content,
		}
		// 分离出kind eg: Deployment/small
		paths := strings.Split(*data.Path, "/")
		if len(paths) > 0 {
			mappingResult.Kind = paths[0]
		}
		oldIndex[key] = &mappingResult
	}
	return oldIndex
}

func init() {
	diffCMD.AddCommand(diffUpgradeCMD)
	diffCMD.AddCommand(diffRevisionCMD)
	diffCMD.AddCommand(diffRollbackCMD)

	diffUpgradeCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	diffUpgradeCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "c", "", "release cluster id")
	diffUpgradeCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace")
	diffUpgradeCMD.PersistentFlags().StringSliceVarP(
		&flagValueFile, "file", "f", nil, "value file for installation, -f values.yaml")
	diffUpgradeCMD.PersistentFlags().StringSliceVarP(
		&flagArgs, "args", "", nil, "--args=--wait=true --args=--timeout=600s")
	diffUpgradeCMD.PersistentFlags().StringVarP(
		&flagRepository, "repo", "r", "", "repository name")

	_ = diffUpgradeCMD.MarkPersistentFlagRequired("project")
	_ = diffUpgradeCMD.MarkPersistentFlagRequired("cluster")
	_ = diffUpgradeCMD.MarkPersistentFlagRequired("namespace")

	diffRevisionCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	diffRevisionCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "c", "", "release cluster id")
	diffRevisionCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace")

	_ = diffRevisionCMD.MarkPersistentFlagRequired("project")
	_ = diffRevisionCMD.MarkPersistentFlagRequired("cluster")
	_ = diffRevisionCMD.MarkPersistentFlagRequired("namespace")

	diffRollbackCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	diffRollbackCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "c", "", "release cluster id")
	diffRollbackCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace")

	_ = diffRollbackCMD.MarkPersistentFlagRequired("project")
	_ = diffRollbackCMD.MarkPersistentFlagRequired("cluster")
	_ = diffRollbackCMD.MarkPersistentFlagRequired("namespace")
}
