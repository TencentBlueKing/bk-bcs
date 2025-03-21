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

package release

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewInstallReleaseV1Action return a new InstallReleaseAction instance
func NewInstallReleaseV1Action(
	model store.HelmManagerModel, platform repo.Platform, releaseHandler release.Handler) *InstallReleaseV1Action {
	return &InstallReleaseV1Action{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// InstallReleaseV1Action provides the action to do install release
type InstallReleaseV1Action struct {
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.InstallReleaseV1Req
	resp *helmmanager.InstallReleaseV1Resp
}

// Handle the installing process
func (i *InstallReleaseV1Action) Handle(ctx context.Context,
	req *helmmanager.InstallReleaseV1Req, resp *helmmanager.InstallReleaseV1Resp) error {
	i.ctx = ctx
	i.req = req
	i.resp = resp

	if err := i.req.Validate(); err != nil {
		blog.Errorf("install release failed, invalid request, %s, param: %v", err.Error(), i.req)
		i.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	if err := i.install(); err != nil {
		blog.Errorf("install release %s failed, clusterID: %s, namespace: %s, error: %s",
			i.req.GetName(), i.req.GetClusterID(), i.req.GetNamespace(), err.Error())
		i.setResp(common.ErrHelmManagerInstallActionFailed, err.Error())
		return nil
	}

	blog.Infof("dispatch release successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s, operator: %s",
		i.req.GetProjectCode(), i.req.GetClusterID(), i.req.GetNamespace(), i.req.GetName(), auth.GetUserFromCtx(i.ctx))
	i.setResp(common.ErrHelmManagerSuccess, "ok")
	return nil
}

func (i *InstallReleaseV1Action) install() error {
	// check release exist
	_, err := i.releaseHandler.Cluster(i.req.GetClusterID()).Get(i.ctx, release.GetOption{
		Namespace: i.req.GetNamespace(), Name: i.req.GetName(),
	})
	if err == nil {
		return fmt.Errorf("release is exist")
	}
	if !errors.Is(err, driver.ErrReleaseNotFound) {
		return fmt.Errorf("check release failed, %s", err.Error())
	}

	if err = i.saveDB(); err != nil {
		return fmt.Errorf("db error, %s", err.Error())
	}

	cls, err := clustermanager.GetCluster(i.ctx, i.req.GetClusterID())
	if err != nil {
		return err
	}

	// dispatch release
	options := &actions.ReleaseInstallActionOption{
		Model:          i.model,
		Platform:       i.platform,
		ReleaseHandler: i.releaseHandler,
		ProjectCode:    contextx.GetProjectCodeFromCtx(i.ctx),
		ProjectID:      contextx.GetProjectIDFromCtx(i.ctx),
		ClusterID:      i.req.GetClusterID(),
		Name:           i.req.GetName(),
		Namespace:      i.req.GetNamespace(),
		RepoName:       i.req.GetRepository(),
		ChartName:      i.req.GetChart(),
		Version:        i.req.GetVersion(),
		Values:         i.req.GetValues(),
		Args:           i.req.GetArgs(),
		Username:       auth.GetUserFromCtx(i.ctx),
		AuthUser:       auth.GetRealUserFromCtx(i.ctx),
		IsShardCluster: cls.IsShared,
	}
	action := actions.NewReleaseInstallAction(options)
	_, err = operation.GlobalOperator.Dispatch(action, releaseDefaultTimeout)
	if err != nil {
		return fmt.Errorf("dispatch failed, %s", err.Error())
	}
	return nil
}

func (i *InstallReleaseV1Action) saveDB() error {
	if err := i.model.DeleteRelease(i.ctx, i.req.GetClusterID(), i.req.GetNamespace(), i.req.GetName()); err != nil {
		return err
	}
	createBy := auth.GetUserFromCtx(i.ctx)
	if i.req.GetOperator() != "" {
		createBy = i.req.GetOperator()
	}
	if err := i.model.CreateRelease(i.ctx, &entity.Release{
		Name:         i.req.GetName(),
		ProjectCode:  contextx.GetProjectCodeFromCtx(i.ctx),
		Namespace:    i.req.GetNamespace(),
		ClusterID:    i.req.GetClusterID(),
		Repo:         i.req.GetRepository(),
		ChartName:    i.req.GetChart(),
		ChartVersion: i.req.GetVersion(),
		ValueFile:    i.req.GetValueFile(),
		Values:       i.req.GetValues(),
		Args:         i.req.GetArgs(),
		CreateBy:     createBy,
		Status:       helmrelease.StatusPendingInstall.String(),
		Env:          i.req.GetEnv(),
	}); err != nil {
		return err
	}
	return nil
}

func (i *InstallReleaseV1Action) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	i.resp.Code = &code
	i.resp.Message = &msg
	i.resp.Result = err.OK()
}
