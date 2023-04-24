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

// NewUpgradeReleaseAction return a new UpgradeReleaseAction instance
func NewUpgradeReleaseAction(
	model store.HelmManagerModel, platform repo.Platform, releaseHandler release.Handler) *UpgradeReleaseAction {
	return &UpgradeReleaseAction{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// UpgradeReleaseAction provides the actions to do upgrade release
type UpgradeReleaseAction struct {
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.UpgradeReleaseReq
	resp *helmmanager.UpgradeReleaseResp
}

// Handle the upgrading process
func (u *UpgradeReleaseAction) Handle(ctx context.Context,
	req *helmmanager.UpgradeReleaseReq, resp *helmmanager.UpgradeReleaseResp) error {

	if req == nil || resp == nil {
		blog.Errorf("upgrade release failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	u.ctx = ctx
	u.req = req
	u.resp = resp

	if err := u.req.Validate(); err != nil {
		blog.Errorf("upgrade release failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return u.upgrade()
}

func (u *UpgradeReleaseAction) upgrade() error {
	releaseName := u.req.GetName()
	releaseNamespace := u.req.GetNamespace()
	clusterID := u.req.GetClusterID()
	projectID := u.req.GetProjectID()
	chartName := u.req.GetChart()
	chartVersion := u.req.GetVersion()
	username := auth.GetUserFromCtx(u.ctx)
	values := u.req.GetValues()

	contents, err := u.getContent()
	if err != nil {
		blog.Errorf("upgrade release, get contents failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error(), nil)
		return nil
	}

	result, err := release.UpgradeRelease(u.releaseHandler, contextx.GetProjectIDFromCtx(u.ctx), projectID, clusterID,
		releaseName, releaseNamespace, chartName, chartVersion, username, username, u.req.GetArgs(),
		u.req.GetBcsSysVar(), contents, values, false)
	if err != nil {
		blog.Errorf("upgrade release failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error(), nil)
		return nil
	}

	// 存储release信息到store中
	if err = u.saveDB(result.Revision); err != nil {
		blog.Warnf("upgrade release, save release in store failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
		// 更新 release 不依赖 db，db 报错也视为 release 更新成功
	}

	blog.Infof("upgrade release successfully, with revision %d, "+
		"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
		result.Revision, projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, username)
	u.setResp(common.ErrHelmManagerSuccess, "ok", (&release.Release{
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

func (u *UpgradeReleaseAction) getContent() ([]byte, error) {
	// 获取对应的仓库信息
	repository, err := u.model.GetRepository(u.ctx, u.req.GetProjectID(), u.req.GetRepository())
	if err != nil {
		return nil, err
	}

	// 下载到具体的chart version信息
	contents, err := u.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(u.req.GetChart()).
		Download(u.ctx, u.req.GetVersion())
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (u *UpgradeReleaseAction) saveDB(revision int) error {
	if err := u.model.DeleteRelease(u.ctx, u.req.GetClusterID(), u.req.GetNamespace(), u.req.GetName()); err != nil {
		return err
	}
	if err := u.model.CreateRelease(u.ctx, &entity.Release{
		Name:         u.req.GetName(),
		ProjectCode:  u.req.GetProjectID(),
		Namespace:    u.req.GetNamespace(),
		ClusterID:    u.req.GetClusterID(),
		Repo:         u.req.GetRepository(),
		ChartName:    u.req.GetChart(),
		ChartVersion: u.req.GetVersion(),
		Revision:     revision,
		Values:       u.req.GetValues(),
		Args:         u.req.GetArgs(),
		CreateBy:     auth.GetUserFromCtx(u.ctx),
	}); err != nil {
		return err
	}
	return nil
}

func (u *UpgradeReleaseAction) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseDetail) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
	u.resp.Data = r
}
