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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewInstallAddonsAction return a new InstallAddonsAction instance
func NewInstallAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform, releaseHandler release.Handler) *InstallAddonsAction {
	return &InstallAddonsAction{
		model:          model,
		addons:         addons,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// InstallAddonsAction provides the action to do install addons
type InstallAddonsAction struct {
	model          store.HelmManagerModel
	addons         release.AddonsSlice
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.InstallAddonsReq
	resp *helmmanager.InstallAddonsResp
}

// Handle the install addons process
func (i *InstallAddonsAction) Handle(ctx context.Context,
	req *helmmanager.InstallAddonsReq, resp *helmmanager.InstallAddonsResp) error {
	i.req = req
	i.resp = resp
	if err := req.Validate(); err != nil {
		blog.Errorf("install addons failed, invalid request, %s, param: %v", err.Error(), i.req)
		i.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	// get addons
	addons := i.addons.FindByName(req.GetName())
	if addons == nil {
		blog.Errorf("get addons detail failed, %s", errorAddonsNotFound.Error())
		i.setResp(common.ErrHelmManagerInstallActionFailed, errorAddonsNotFound.Error())
		return nil
	}

	// save db
	if err := i.saveDB(ctx, addons.Namespace, addons.ChartName, addons.GetReleaseName()); err != nil {
		blog.Errorf("save addons failed, %s", err.Error())
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error())
		return nil
	}

	// 对于非 chart 类型的 addons，直接返回成功
	if isFakeChart(addons.ChartName) {
		i.setResp(common.ErrHelmManagerSuccess, "ok")
		return nil
	}

	// dispatch release
	options := &actions.ReleaseInstallActionOption{
		Model:          i.model,
		Platform:       i.platform,
		ReleaseHandler: i.releaseHandler,
		ProjectCode:    contextx.GetProjectCodeFromCtx(ctx),
		ProjectID:      contextx.GetProjectIDFromCtx(ctx),
		ClusterID:      i.req.GetClusterID(),
		Name:           addons.GetReleaseName(),
		Namespace:      addons.Namespace,
		RepoName:       common.PublicRepoName,
		ChartName:      addons.ChartName,
		Version:        i.req.GetVersion(),
		Values:         []string{i.req.GetValues()},
		Args:           defaultArgs,
		Username:       auth.GetUserFromCtx(ctx),
	}
	action := actions.NewReleaseInstallAction(options)
	if _, err := operation.GlobalOperator.Dispatch(action, releaseDefaultTimeout); err != nil {
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error())
		return fmt.Errorf("dispatch failed, %s", err.Error())
	}
	i.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (i *InstallAddonsAction) saveDB(ctx context.Context, ns, chartName, releaseName string) error {
	if err := i.model.DeleteRelease(ctx, i.req.GetClusterID(), ns, releaseName); err != nil {
		return err
	}
	createBy := auth.GetUserFromCtx(ctx)
	status := helmrelease.StatusPendingInstall.String()
	// 对于非 chart 类型的 addons，直接返回成功
	if isFakeChart(chartName) {
		status = helmrelease.StatusDeployed.String()
	}
	if err := i.model.CreateRelease(ctx, &entity.Release{
		Name:         releaseName,
		ProjectCode:  contextx.GetProjectCodeFromCtx(ctx),
		Namespace:    ns,
		ClusterID:    i.req.GetClusterID(),
		Repo:         common.PublicRepoName,
		ChartName:    chartName,
		ChartVersion: i.req.GetVersion(),
		Values:       []string{i.req.GetValues()},
		Args:         defaultArgs,
		CreateBy:     createBy,
		Status:       status,
	}); err != nil {
		return err
	}
	return nil
}

func (i *InstallAddonsAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	i.resp.Code = &code
	i.resp.Message = &msg
	i.resp.Result = err.OK()
}
