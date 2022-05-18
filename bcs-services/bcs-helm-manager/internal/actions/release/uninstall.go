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
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewUninstallReleaseAction return a new UninstallReleaseAction instance
func NewUninstallReleaseAction(
	model store.HelmManagerModel, platform repo.Platform, releaseHandler release.Handler) *UninstallReleaseAction {
	return &UninstallReleaseAction{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// UninstallReleaseAction provides the action to do uninstall release
type UninstallReleaseAction struct {
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.UninstallReleaseReq
	resp *helmmanager.UninstallReleaseResp
}

// Handle the uninstalling process
func (u *UninstallReleaseAction) Handle(ctx context.Context,
	req *helmmanager.UninstallReleaseReq, resp *helmmanager.UninstallReleaseResp) error {

	if req == nil || resp == nil {
		blog.Errorf("uninstall release failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	u.ctx = ctx
	u.req = req
	u.resp = resp

	if err := u.req.Validate(); err != nil {
		blog.Errorf("uninstall release failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	return u.uninstall()
}

func (u *UninstallReleaseAction) uninstall() error {
	releaseName := u.req.GetName()
	releaseNamespace := u.req.GetNamespace()
	clusterID := u.req.GetClusterID()
	username := auth.GetUserFromCtx(u.ctx)

	_, err := u.releaseHandler.Cluster(clusterID).Uninstall(
		u.ctx,
		release.HelmUninstallConfig{
			Name:      releaseName,
			Namespace: releaseNamespace,
		})
	if err != nil {
		blog.Errorf("uninstall release failed, %s, "+
			"clusterID: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), clusterID, releaseNamespace, releaseName, username)
		u.setResp(common.ErrHelmManagerUninstallActionFailed, err.Error())
		return nil
	}

	// 删掉所有revision的数据
	if err = u.model.DeleteReleases(u.ctx, clusterID, releaseNamespace, releaseName); err != nil {
		blog.Errorf("uninstall release, delete releases in store failed, %s, "+
			"clusterID: %s, namespace: %s, name: %s, operator: %s",
			err.Error(), clusterID, releaseNamespace, releaseName, username)
		u.setResp(common.ErrHelmManagerUninstallActionFailed, err.Error())
		return nil
	}

	blog.Infof("uninstall release successfully, clusterID: %s, namespace: %s, name: %s, operator: %s",
		clusterID, releaseNamespace, releaseName, username)
	u.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (u *UninstallReleaseAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
}
