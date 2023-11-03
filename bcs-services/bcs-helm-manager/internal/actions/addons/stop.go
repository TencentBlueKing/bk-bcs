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

// NewStopAddonsAction return a new StopAddonsAction instance
func NewStopAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform, releaseHandler release.Handler) *StopAddonsAction {
	return &StopAddonsAction{
		model:          model,
		addons:         addons,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// StopAddonsAction provides the action to do stop addons
type StopAddonsAction struct {
	model          store.HelmManagerModel
	addons         release.AddonsSlice
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.StopAddonsReq
	resp *helmmanager.StopAddonsResp

	createBy string
	updateBy string
	version  string
}

// Handle the addons stop process
func (s *StopAddonsAction) Handle(ctx context.Context,
	req *helmmanager.StopAddonsReq, resp *helmmanager.StopAddonsResp) error {
	s.req = req
	s.resp = resp
	if err := req.Validate(); err != nil {
		blog.Errorf("stop addons failed, invalid request, %s, param: %v", err.Error(), s.req)
		s.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	// get addons
	addons := s.addons.FindByName(req.GetName())
	if addons == nil {
		blog.Errorf("get addons detail failed, %s", errorAddonsNotFound.Error())
		s.setResp(common.ErrHelmManagerUpgradeActionFailed, errorAddonsNotFound.Error())
		return nil
	}
	if !addons.CanStop() {
		s.setResp(common.ErrHelmManagerUpgradeActionFailed, "addons can't stop")
		return nil
	}

	// save db
	if err := s.saveDB(ctx, addons.Namespace, addons.ChartName, addons.ReleaseName()); err != nil {
		blog.Errorf("save addons failed, %s", err.Error())
		s.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error())
		return nil
	}

	// dispatch release
	options := &actions.ReleaseUpgradeActionOption{
		Model:          s.model,
		Platform:       s.platform,
		ReleaseHandler: s.releaseHandler,
		ProjectCode:    contextx.GetProjectCodeFromCtx(ctx),
		ProjectID:      contextx.GetProjectIDFromCtx(ctx),
		ClusterID:      s.req.GetClusterID(),
		Name:           addons.ReleaseName(),
		Namespace:      addons.Namespace,
		RepoName:       common.PublicRepoName,
		ChartName:      addons.ChartName,
		Version:        s.version,
		Values:         []string{addons.StopValues},
		CreateBy:       s.createBy,
		UpdateBy:       s.updateBy,
	}
	action := actions.NewReleaseUpgradeAction(options)
	if _, err := operation.GlobalOperator.Dispatch(action, releaseDefaultTimeout); err != nil {
		s.setResp(common.ErrHelmManagerUpgradeActionFailed, err.Error())
		return fmt.Errorf("dispatch failed, %s", err.Error())
	}
	s.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (s *StopAddonsAction) saveDB(ctx context.Context, ns, chartName, releaseName string) error { // nolint
	old, err := s.model.GetRelease(ctx, s.req.GetClusterID(), ns, releaseName)
	if err != nil {
		return err
	}

	s.createBy = old.CreateBy
	s.updateBy = auth.GetUserFromCtx(ctx)
	s.version = old.ChartVersion
	rl := entity.M{
		entity.FieldKeyUpdateBy: s.updateBy,
		entity.FieldKeyStatus:   helmrelease.StatusPendingUpgrade.String(),
		entity.FieldKeyMessage:  "",
	}
	if err := s.model.UpdateRelease(ctx, s.req.GetClusterID(), ns, releaseName, rl); err != nil {
		return err
	}
	return nil
}

func (s *StopAddonsAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	s.resp.Code = &code
	s.resp.Message = &msg
	s.resp.Result = err.OK()
}
