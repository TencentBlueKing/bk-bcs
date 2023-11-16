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
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewUninstallAddonsAction return a new UninstallAddonsAction instance
func NewUninstallAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform, releaseHandler release.Handler) *UninstallAddonsAction {
	return &UninstallAddonsAction{
		model:          model,
		addons:         addons,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// UninstallAddonsAction provides the action to do uninstall addons
type UninstallAddonsAction struct {
	model          store.HelmManagerModel
	addons         release.AddonsSlice
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.UninstallAddonsReq
	resp *helmmanager.UninstallAddonsResp
}

// Handle the uninstalling process
func (u *UninstallAddonsAction) Handle(ctx context.Context,
	req *helmmanager.UninstallAddonsReq, resp *helmmanager.UninstallAddonsResp) error {
	u.req = req
	u.resp = resp
	if err := u.req.Validate(); err != nil {
		blog.Errorf("uninstall addons failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	// get addons
	addons := u.addons.FindByName(req.GetName())
	if addons == nil {
		blog.Errorf("get addons detail failed, %s", errorAddonsNotFound.Error())
		u.setResp(common.ErrHelmManagerUninstallActionFailed, errorAddonsNotFound.Error())
		return nil
	}

	if err := u.uninstall(ctx, addons.Namespace, addons.ChartName, addons.ReleaseName()); err != nil {
		blog.Errorf("uninstall addons %s failed, clusterID: %s, namespace: %s, error: %s",
			addons.ReleaseName(), u.req.GetClusterID(), addons.Namespace, err.Error())
		u.setResp(common.ErrHelmManagerUninstallActionFailed, err.Error())
		return nil
	}

	blog.Infof("dispatch release successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s, operator: %s",
		u.req.GetProjectCode(), u.req.GetClusterID(), addons.Namespace, addons.ReleaseName(), auth.GetUserFromCtx(ctx))
	u.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (u *UninstallAddonsAction) uninstall(ctx context.Context, ns, chartName, releaseName string) error {
	if err := u.saveDB(ctx, ns, chartName, releaseName); err != nil {
		return fmt.Errorf("db error, %s", err.Error())
	}

	// 对于非 chart 类型的 addons，直接返回成功
	if chartName == "" {
		return nil
	}

	// dispatch release
	options := &actions.ReleaseUninstallActionOption{
		Model:          u.model,
		ReleaseHandler: u.releaseHandler,
		ClusterID:      u.req.GetClusterID(),
		Name:           releaseName,
		Namespace:      ns,
		Username:       auth.GetUserFromCtx(ctx),
	}
	action := actions.NewReleaseUninstallAction(options)
	_, err := operation.GlobalOperator.Dispatch(action, releaseDefaultTimeout)
	if err != nil {
		return fmt.Errorf("dispatch failed, %s", err.Error())
	}
	return nil
}

func (u *UninstallAddonsAction) saveDB(ctx context.Context, ns, chartName, releaseName string) error {
	_, err := u.model.GetRelease(ctx, u.req.GetClusterID(), ns, releaseName)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			return nil
		}
		return err
	}

	if chartName == "" {
		if err := u.model.DeleteRelease(ctx, u.req.GetClusterID(), ns, releaseName); err != nil {
			return err
		}
		return nil
	}

	username := auth.GetUserFromCtx(ctx)
	rl := entity.M{
		entity.FieldKeyUpdateBy: username,
		entity.FieldKeyStatus:   helmrelease.StatusUninstalling.String(),
		entity.FieldKeyMessage:  "",
	}
	if err := u.model.UpdateRelease(ctx, u.req.GetClusterID(), ns, releaseName, rl); err != nil {
		return err
	}
	return nil
}

func (u *UninstallAddonsAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
}
