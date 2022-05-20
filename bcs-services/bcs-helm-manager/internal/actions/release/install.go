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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewInstallReleaseAction return a new InstallReleaseAction instance
func NewInstallReleaseAction(
	model store.HelmManagerModel, platform repo.Platform, releaseHandler release.Handler) *InstallReleaseAction {
	return &InstallReleaseAction{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// InstallReleaseAction provides the action to do install release
type InstallReleaseAction struct {
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.InstallReleaseReq
	resp *helmmanager.InstallReleaseResp
}

// Handle the installing process
func (i *InstallReleaseAction) Handle(ctx context.Context,
	req *helmmanager.InstallReleaseReq, resp *helmmanager.InstallReleaseResp) error {

	if req == nil || resp == nil {
		blog.Errorf("install release failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	i.ctx = ctx
	i.req = req
	i.resp = resp

	if err := i.req.Validate(); err != nil {
		blog.Errorf("install release failed, invalid request, %s, param: %v", err.Error(), i.req)
		i.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return i.install()
}

func (i *InstallReleaseAction) install() error {
	releaseName := i.req.GetName()
	releaseNamespace := i.req.GetNamespace()
	clusterID := i.req.GetClusterID()
	projectID := i.req.GetProjectID()
	repoName := i.req.GetRepository()
	chartName := i.req.GetChart()
	chartVersion := i.req.GetVersion()
	values := i.req.GetValues()
	username := auth.GetUserFromCtx(i.ctx)

	// 获取对应的仓库信息
	repository, err := i.model.GetRepository(i.ctx, projectID, repoName)
	if err != nil {
		blog.Errorf("install release get repository failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error(), nil)
		return nil
	}

	// 下载到具体的chart version信息
	contents, err := i.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.ProjectID).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.Name,
		).
		Chart(chartName).
		Download(i.ctx, chartVersion)
	if err != nil {
		blog.Errorf("install release get chart detail failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error(), nil)
		return nil
	}

	vls := make([]*release.File, 0, len(values))
	for index, v := range values {
		vls = append(vls, &release.File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}

	// 执行install操作
	result, err := i.releaseHandler.Cluster(clusterID).Install(
		i.ctx,
		release.HelmInstallConfig{
			Name:      releaseName,
			Namespace: releaseNamespace,
			Chart: &release.File{
				Name:    chartName + "-" + chartVersion + ".tgz",
				Content: contents,
			},
			Args:   i.req.GetArgs(),
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: "",
				common.PTKClusterID: clusterID,
				common.PTKNamespace: releaseNamespace,
				common.PTKCreator:   username,
				common.PTKUpdator:   username,
				common.PTKVersion:   "",
				common.PTKName:      "",
			},
			VarTemplateValues: i.req.GetBcsSysVar(),
		})
	if err != nil {
		blog.Errorf("install release failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error(), nil)
		return nil
	}

	// 存储release信息到store中, 首先先删掉所有revision的数据
	if err = i.model.DeleteReleases(i.ctx, clusterID, releaseNamespace, releaseNamespace); err != nil {
		blog.Errorf("install release, delete release in store failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error(), nil)
		return nil
	}
	if err = i.model.CreateRelease(i.ctx, &entity.Release{
		Name:         releaseName,
		Namespace:    releaseNamespace,
		ClusterID:    clusterID,
		ChartName:    chartName,
		ChartVersion: chartVersion,
		Revision:     result.Revision,
		Values:       values,
	}); err != nil {
		blog.Errorf("install release, create release in store failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error(), nil)
		return nil
	}

	blog.Infof("install release successfully, with revision %d, "+
		"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
		result.Revision, projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
	i.setResp(common.ErrHelmManagerSuccess, "ok", (&release.Release{
		Name:         releaseName,
		Namespace:    releaseNamespace,
		Revision:     result.Revision,
		Status:       result.Status,
		Chart:        chartName,
		ChartVersion: chartVersion,
		AppVersion:   result.AppVersion,
		UpdateTime:   result.UpdateTime,
	}).Transfer2DetailProto())
	return nil
}

func (i *InstallReleaseAction) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseDetail) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	i.resp.Code = &code
	i.resp.Message = &msg
	i.resp.Result = err.OK()
	i.resp.Data = r
}
