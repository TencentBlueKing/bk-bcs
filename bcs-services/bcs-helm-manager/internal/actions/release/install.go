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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
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
	chartName := i.req.GetChart()
	chartVersion := i.req.GetVersion()
	values := i.req.GetValues()
	username := auth.GetUserFromCtx(i.ctx)

	contents, err := i.getContent()
	if err != nil {
		blog.Errorf("install release, get contents failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error(), nil)
		return nil
	}

	// 执行install操作
	result, err := release.InstallRelease(i.releaseHandler, contextx.GetProjectIDFromCtx(i.ctx), projectID, clusterID,
		releaseName, releaseNamespace, chartName, chartVersion, username, username, i.req.GetArgs(),
		i.req.GetBcsSysVar(), contents, values, false, false, false)
	if err != nil {
		blog.Errorf("install release failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error(), nil)
		return nil
	}

	// 存储release信息到store中
	if err = i.saveDB(result.Revision); err != nil {
		blog.Warnf("install release, save release in store failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		// release 不依赖数据库，保存到数据库失败不视为失败
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

func (i *InstallReleaseAction) getContent() ([]byte, error) {
	// 获取对应的仓库信息
	repository, err := i.model.GetRepository(i.ctx, i.req.GetProjectID(), i.req.GetRepository())
	if err != nil {
		return nil, err
	}

	// 下载到具体的chart version信息
	contents, err := i.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(i.req.GetChart()).
		Download(i.ctx, i.req.GetVersion())
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (i *InstallReleaseAction) saveDB(revision int) error {
	// 首先先删掉所有revision的数据，因为是安装操作，所以以前的数据都不需要了
	if err := i.model.DeleteRelease(i.ctx, i.req.GetClusterID(), i.req.GetNamespace(), i.req.GetName()); err != nil {
		return err
	}
	if err := i.model.CreateRelease(i.ctx, &entity.Release{
		Name:         i.req.GetName(),
		Namespace:    i.req.GetNamespace(),
		ProjectCode:  i.req.GetProjectID(),
		ClusterID:    i.req.GetClusterID(),
		Repo:         i.req.GetRepository(),
		ChartName:    i.req.GetChart(),
		ChartVersion: i.req.GetVersion(),
		Revision:     revision,
		Values:       i.req.GetValues(),
		Args:         i.req.GetArgs(),
		CreateBy:     auth.GetUserFromCtx(i.ctx),
	}); err != nil {
		return err
	}
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
