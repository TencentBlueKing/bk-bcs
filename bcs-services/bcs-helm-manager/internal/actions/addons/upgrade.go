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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
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

// NewUpgradeAddonsAction return a new UpgradeAddonsAction instance
func NewUpgradeAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform, releaseHandler release.Handler) *UpgradeAddonsAction {
	return &UpgradeAddonsAction{
		model:          model,
		addons:         addons,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// UpgradeAddonsAction provides the action to do upgrade addons
type UpgradeAddonsAction struct {
	model          store.HelmManagerModel
	addons         release.AddonsSlice
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.UpgradeAddonsReq
	resp *helmmanager.UpgradeAddonsResp

	createBy string
	updateBy string
}

// Handle the addons upgrade process
func (u *UpgradeAddonsAction) Handle(ctx context.Context,
	req *helmmanager.UpgradeAddonsReq, resp *helmmanager.UpgradeAddonsResp) error {
	u.req = req
	u.resp = resp
	if err := req.Validate(); err != nil {
		blog.Errorf("update addons failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	// get addons
	addons := u.addons.FindByName(req.GetName())
	if addons == nil {
		blog.Errorf("get addons detail failed, %s", errorAddonsNotFound.Error())
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, errorAddonsNotFound.Error())
		return nil
	}
	if !addons.CanUpgrade() {
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, "addons can't upgrade")
		return nil
	}

	// save db
	if err := u.saveDB(ctx, addons.Namespace, addons.ChartName, addons.GetReleaseName()); err != nil {
		blog.Errorf("save addons failed, %s", err.Error())
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error())
		return nil
	}

	values := []string{addons.DefaultValues, u.req.GetValues()}
	if addons.IgnoreDefaultValues {
		values = []string{u.req.GetValues()}
	}

	// dispatch release
	options := &actions.ReleaseUpgradeActionOption{
		Model:          u.model,
		Platform:       u.platform,
		ReleaseHandler: u.releaseHandler,
		ProjectCode:    contextx.GetProjectCodeFromCtx(ctx),
		ProjectID:      contextx.GetProjectIDFromCtx(ctx),
		ClusterID:      u.req.GetClusterID(),
		Name:           addons.GetReleaseName(),
		Namespace:      addons.Namespace,
		RepoName:       common.PublicRepoName,
		ChartName:      addons.ChartName,
		Version:        u.req.GetVersion(),
		Values:         values,
		Args:           defaultArgs,
		CreateBy:       u.createBy,
		UpdateBy:       u.updateBy,
	}
	action := actions.NewReleaseUpgradeAction(options)
	if _, err := operation.GlobalOperator.Dispatch(action, releaseDefaultTimeout); err != nil {
		u.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error())
		return fmt.Errorf("dispatch failed, %s", err.Error())
	}
	u.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (u *UpgradeAddonsAction) saveDB(ctx context.Context, ns, chartName, releaseName string) error {
	create := false
	old, err := u.model.GetRelease(ctx, u.req.GetClusterID(), ns, releaseName)
	if err != nil {
		if !errors.Is(err, drivers.ErrTableRecordNotFound) {
			return err
		}
		create = true
	}

	status := helmrelease.StatusPendingUpgrade.String()
	// 对于非 chart 类型的 addons，直接返回成功
	if isFakeChart(chartName) {
		status = helmrelease.StatusDeployed.String()
	}
	createBy := auth.GetUserFromCtx(ctx)
	if create {
		u.createBy = createBy
		u.updateBy = createBy
		if err := u.model.CreateRelease(ctx, &entity.Release{
			Name:         releaseName,
			ProjectCode:  contextx.GetProjectCodeFromCtx(ctx),
			Namespace:    ns,
			ClusterID:    u.req.GetClusterID(),
			Repo:         common.PublicRepoName,
			ChartName:    chartName,
			ChartVersion: u.req.GetVersion(),
			Values:       []string{u.req.GetValues()},
			Args:         defaultArgs,
			CreateBy:     createBy,
			Status:       status,
		}); err != nil {
			return err
		}
	} else {
		u.createBy = old.CreateBy
		u.updateBy = createBy
		if u.req.GetVersion() == "" {
			u.req.Version = &old.ChartVersion
		}
		if u.req.GetValues() == "" && len(old.Values) > 0 {
			u.req.Values = &old.Values[len(old.Values)-1]
		}
		rl := entity.M{
			entity.FieldKeyRepoName:     common.PublicRepoName,
			entity.FieldKeyChartName:    chartName,
			entity.FieldKeyChartVersion: u.req.GetVersion(),
			entity.FieldKeyValues:       []string{u.req.GetValues()},
			entity.FieldKeyArgs:         defaultArgs,
			entity.FieldKeyUpdateBy:     createBy,
			entity.FieldKeyStatus:       status,
			entity.FieldKeyMessage:      "",
		}
		if err := u.model.UpdateRelease(ctx, u.req.GetClusterID(), ns, releaseName, rl); err != nil {
			return err
		}
	}
	return nil
}

func (u *UpgradeAddonsAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
}
