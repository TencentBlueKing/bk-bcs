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

// NewUninstallReleaseV1Action return a new UninstallReleaseAction instance
func NewUninstallReleaseV1Action(
	model store.HelmManagerModel, platform repo.Platform, releaseHandler release.Handler) *UninstallReleaseV1Action {
	return &UninstallReleaseV1Action{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// UninstallReleaseV1Action provides the action to do uninstall release
type UninstallReleaseV1Action struct {
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.UninstallReleaseV1Req
	resp *helmmanager.UninstallReleaseV1Resp
}

// Handle the uninstalling process
func (u *UninstallReleaseV1Action) Handle(ctx context.Context,
	req *helmmanager.UninstallReleaseV1Req, resp *helmmanager.UninstallReleaseV1Resp) error {
	u.ctx = ctx
	u.req = req
	u.resp = resp

	if err := u.req.Validate(); err != nil {
		blog.Errorf("uninstall release failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	if err := u.uninstall(); err != nil {
		blog.Errorf("uninstall release %s failed, clusterID: %s, namespace: %s, error: %s",
			u.req.GetName(), u.req.GetClusterID(), u.req.GetNamespace(), err.Error())
		u.setResp(common.ErrHelmManagerUninstallActionFailed, err.Error())
		return nil
	}

	blog.Infof("dispatch release successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s, operator: %s",
		u.req.GetProjectCode(), u.req.GetClusterID(), u.req.GetNamespace(), u.req.GetName(), auth.GetUserFromCtx(u.ctx))
	u.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (u *UninstallReleaseV1Action) uninstall() error {
	if err := u.saveDB(); err != nil {
		return fmt.Errorf("db error, %s", err.Error())
	}

	// dispatch release
	options := &actions.ReleaseUninstallActionOption{
		Model:          u.model,
		ReleaseHandler: u.releaseHandler,
		ClusterID:      u.req.GetClusterID(),
		Name:           u.req.GetName(),
		Namespace:      u.req.GetNamespace(),
		Username:       auth.GetUserFromCtx(u.ctx),
	}
	action := actions.NewReleaseUninstallAction(options)
	_, err := operation.GlobalOperator.Dispatch(action, releaseDefaultTimeout)
	if err != nil {
		return fmt.Errorf("dispatch failed, %s", err.Error())
	}
	return nil
}

func (u *UninstallReleaseV1Action) saveDB() error {
	create := false
	_, err := u.model.GetRelease(u.ctx, u.req.GetClusterID(), u.req.GetNamespace(), u.req.GetName())
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			create = true
		} else {
			return err
		}
	}

	username := auth.GetUserFromCtx(u.ctx)
	if create {
		if err := u.model.CreateRelease(u.ctx, &entity.Release{
			ProjectCode: contextx.GetProjectCodeFromCtx(u.ctx),
			Name:        u.req.GetName(),
			Namespace:   u.req.GetNamespace(),
			ClusterID:   u.req.GetClusterID(),
			CreateBy:    username,
			Status:      helmrelease.StatusUninstalling.String(),
		}); err != nil {
			return err
		}
	} else {
		rl := entity.M{
			entity.FieldKeyUpdateBy: username,
			entity.FieldKeyStatus:   helmrelease.StatusUninstalling.String(),
			entity.FieldKeyMessage:  "",
		}
		if err := u.model.UpdateRelease(u.ctx, u.req.GetClusterID(), u.req.GetNamespace(),
			u.req.GetName(), rl); err != nil {

		}
	}
	return nil
}

func (u *UninstallReleaseV1Action) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
}
