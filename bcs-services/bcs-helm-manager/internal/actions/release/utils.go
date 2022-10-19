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

package release

import (
	"context"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
)

func getChartContent(model store.HelmManagerModel, platform repo.Platform,
	projectID, repoName, chart, version string) ([]byte, error) {
	// 获取对应的仓库信息
	repository, err := model.GetRepository(context.Background(), projectID, repoName)
	if err != nil {
		return nil, err
	}

	// 下载到具体的chart version信息
	contents, err := platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.ProjectID).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.Name,
		).
		Chart(chart).
		Download(context.Background(), version)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func installRelease(releaseHandler release.Handler, projectID, clusterID, releaseName,
	releaseNamespace, chartName, version, username string, args []string, bcsSysVar map[string]string,
	contents []byte, values []string, dryRun bool) (*release.HelmInstallResult, error) {
	vls := make([]*release.File, 0, len(values))
	for index, v := range values {
		vls = append(vls, &release.File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}
	return releaseHandler.Cluster(clusterID).Install(
		context.Background(),
		release.HelmInstallConfig{
			DryRun:    dryRun,
			Name:      releaseName,
			Namespace: releaseNamespace,
			Chart: &release.File{
				Name:    chartName + "-" + version + ".tgz",
				Content: contents,
			},
			Args:   args,
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: projectID,
				common.PTKClusterID: clusterID,
				common.PTKNamespace: releaseNamespace,
				common.PTKCreator:   username,
				common.PTKUpdator:   username,
				common.PTKVersion:   version,
			},
			VarTemplateValues: bcsSysVar,
		})
}

func upgradeRelease(releaseHandler release.Handler, projectID, clusterID, releaseName,
	releaseNamespace, chartName, version, username string, args []string, bcsSysVar map[string]string,
	contents []byte, values []string, dryRun bool) (*release.HelmUpgradeResult, error) {
	vls := make([]*release.File, 0, len(values))
	for index, v := range values {
		vls = append(vls, &release.File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}
	return releaseHandler.Cluster(clusterID).Upgrade(
		context.Background(),
		release.HelmUpgradeConfig{
			DryRun:    dryRun,
			Name:      releaseName,
			Namespace: releaseNamespace,
			Chart: &release.File{
				Name:    chartName + "-" + version + ".tgz",
				Content: contents,
			},
			Args:   args,
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: projectID,
				common.PTKClusterID: clusterID,
				common.PTKNamespace: releaseNamespace,
				common.PTKUpdator:   username,
				common.PTKVersion:   version,
			},
			VarTemplateValues: bcsSysVar,
		})
}
