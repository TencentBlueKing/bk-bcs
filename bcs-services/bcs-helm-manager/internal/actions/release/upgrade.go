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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
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
	repoName := u.req.GetRepository()
	chartName := u.req.GetChart()
	chartVersion := u.req.GetVersion()
	opName := u.req.GetOperator()
	values := u.req.GetValues()

	// 获取对应的仓库信息
	repository, err := u.model.GetRepository(u.ctx, projectID, repoName)
	if err != nil {
		blog.Errorf("upgrade release get repository failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, opName)
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error(), nil)
		return nil
	}

	// 下载到具体的chart version信息
	contents, err := u.platform.
		User(repository.Username).
		Project(repository.ProjectID).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.Name,
		).
		Chart(chartName).
		Download(u.ctx, chartVersion)
	if err != nil {
		blog.Errorf("upgrade release get chart detail failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, opName)
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error(), nil)
		return nil
	}

	vls := make([]*release.File, 0, len(values))
	for index, v := range values {
		vls = append(vls, &release.File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}

	// 执行upgrade操作
	result, err := u.releaseHandler.Cluster(clusterID).Upgrade(
		u.ctx,
		release.HelmUpgradeConfig{
			Name:      releaseName,
			Namespace: releaseNamespace,
			Chart: &release.File{
				Name:    chartName + "-" + chartVersion + ".tgz",
				Content: contents,
			},
			Args:   u.req.GetArgs(),
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: "",
				common.PTKClusterID: clusterID,
				common.PTKNamespace: releaseNamespace,
				common.PTKUpdator:   opName,
				common.PTKVersion:   "",
				common.PTKName:      "",
			},
			VarTemplateValues: u.req.GetBcsSysVar(),
		})
	if err != nil {
		blog.Errorf("upgrade release failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, opName)
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error(), nil)
		return nil
	}

	// 存储release信息到store中, 首先先删掉原来的同revision的数据
	if err = u.model.DeleteRelease(u.ctx, clusterID, releaseNamespace, releaseNamespace, result.Revision); err != nil {
		blog.Errorf("upgrade release, delete release in store failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, opName)
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error(), nil)
		return nil
	}
	if err = u.model.CreateRelease(u.ctx, &entity.Release{
		Name:         releaseName,
		Namespace:    releaseNamespace,
		ClusterID:    clusterID,
		ChartName:    chartName,
		ChartVersion: chartVersion,
		Revision:     result.Revision,
		Values:       values,
	}); err != nil {
		blog.Errorf("upgrade release, create release in store failed, %s, "+
			"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, opName)
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error(), nil)
		return nil
	}

	blog.Infof("upgrade release successfully, with revision %d, "+
		"projectID: %s, clusterID: %s, chartName: %s, chartVersion: %s, namespace: %s, name: %s, operator: %s",
		result.Revision, projectID, clusterID, chartName, chartVersion, releaseNamespace, releaseName, opName)
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

func (u *UpgradeReleaseAction) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseDetail) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
	u.resp.Data = r
}
