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

package addons

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/storage/driver"

	actionsrelease "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/actions/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewPreviewAddonsAction return a new PreviewAddonsAction instance
func NewPreviewAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform, releaseHandler release.Handler) *PreviewAddonsAction {
	return &PreviewAddonsAction{
		model:          model,
		addons:         addons,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// PreviewAddonsAction provides the action to do preview addons
type PreviewAddonsAction struct {
	model          store.HelmManagerModel
	addons         release.AddonsSlice
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.PreviewAddonsReq
	resp *helmmanager.ReleasePreviewResp

	createBy string
	updateBy string
}

// Handle the addons preview process
func (u *PreviewAddonsAction) Handle(ctx context.Context,
	req *helmmanager.PreviewAddonsReq, resp *helmmanager.ReleasePreviewResp) error {
	u.req = req
	u.resp = resp
	if err := req.Validate(); err != nil {
		blog.Errorf("preview addons failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	// get addons
	addons := u.addons.FindByName(req.GetName())
	if addons == nil {
		blog.Errorf("get addons detail failed, %s", errorAddonsNotFound.Error())
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, errorAddonsNotFound.Error(), nil)
		return nil
	}
	if !addons.CanUpgrade() {
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, "addons can't upgrade", nil)
		return nil
	}

	old, err := u.model.GetRelease(ctx, u.req.GetClusterID(), addons.Namespace, addons.GetReleaseName())
	if err != nil {
		blog.Errorf("get release failed, %s", err.Error())
		u.setResp(common.ErrHelmManagerGetActionFailed, "get release failed", nil)
		return nil
	}

	u.createBy = old.CreateBy
	u.updateBy = auth.GetUserFromCtx(ctx)

	preview, err := u.getReleasePreview(ctx, addons)
	if err != nil {
		blog.Errorf("get release preview failed, %s", err.Error())
		u.setResp(common.ErrHelmManagerGetActionFailed, "get release preview failed", nil)
		return nil
	}

	u.setResp(common.ErrHelmManagerSuccess, "ok", preview)
	return nil
}

func (u *PreviewAddonsAction) getReleasePreview(
	ctx context.Context, addon *release.Addons) (*helmmanager.ReleasePreview, error) {
	// get manifest from helm
	currentRelease, err := u.releaseHandler.Cluster(u.req.GetClusterID()).Get(ctx, release.GetOption{
		Namespace: addon.Namespace, Name: addon.GetReleaseName()})
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return nil, fmt.Errorf("get current releasefailed, err %s", err.Error())
	}

	projectCode := contextx.GetProjectCodeFromCtx(ctx)
	contents, err := getChartContent(u.model, u.platform, projectCode, common.PublicRepoName,
		addon.ChartName, u.req.GetVersion())
	if err != nil {
		return nil, fmt.Errorf("get release preview, get contents failed, %s", err.Error())
	}

	values := []string{addon.DefaultValues, u.req.GetValues()}
	if addon.IgnoreDefaultValues {
		values = []string{u.req.GetValues()}
	}

	// dispatch release
	options := &actions.ReleasePreviewActionOption{
		Model:          u.model,
		Platform:       u.platform,
		ReleaseHandler: u.releaseHandler,
		ProjectCode:    contextx.GetProjectCodeFromCtx(ctx),
		ProjectID:      contextx.GetProjectIDFromCtx(ctx),
		ClusterID:      u.req.GetClusterID(),
		Name:           addon.GetReleaseName(),
		Namespace:      addon.Namespace,
		RepoName:       common.PublicRepoName,
		ChartName:      addon.ChartName,
		Version:        u.req.GetVersion(),
		Values:         values,
		Args:           defaultArgs,
		CreateBy:       u.createBy,
		UpdateBy:       u.updateBy,
		Content:        contents,
	}
	action := actions.NewReleasePreviewAction(options)
	newRelease, err := action.UpgradeRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("upgrade addons failed, %s", err.Error())
	}
	rp := actionsrelease.NewReleasePreviewAction(u.model, u.platform, u.releaseHandler)
	preview, err := rp.GenerateReleasePreview(currentRelease.Transfer2Release(), newRelease)
	if err != nil {
		blog.Errorf("generate release preview failed, %s", err.Error())
		u.setResp(common.ErrHelmManagerPreviewActionFailed, "addons preview failed", nil)
	}
	return preview, nil
}

func getChartContent(model store.HelmManagerModel, platform repo.Platform,
	projectID, repoName, chart, version string) ([]byte, error) {
	// 获取对应的仓库信息
	repository, err := model.GetProjectRepository(context.Background(), projectID, repoName)
	if err != nil {
		return nil, err
	}

	// 下载到具体的chart version信息
	contents, err := platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(chart).
		Download(context.Background(), version)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (u *PreviewAddonsAction) setResp(err common.HelmManagerError, message string, rp *helmmanager.ReleasePreview) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
	u.resp.Data = rp
}
