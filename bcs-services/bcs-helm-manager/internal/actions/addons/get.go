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

// Package addons xxx
package addons

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetAddonsDetailAction return a new GetAddonsDetailAction instance
func NewGetAddonsDetailAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform, releaseHandler release.Handler) *GetAddonsDetailAction {
	return &GetAddonsDetailAction{
		model:          model,
		addons:         addons,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// GetAddonsDetailAction provides the action to do get addons
type GetAddonsDetailAction struct {
	model          store.HelmManagerModel
	addons         release.AddonsSlice
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.GetAddonsDetailReq
	resp *helmmanager.GetAddonsDetailResp
}

// Handle the get addons process
func (g *GetAddonsDetailAction) Handle(ctx context.Context,
	req *helmmanager.GetAddonsDetailReq, resp *helmmanager.GetAddonsDetailResp) error {
	g.req = req
	g.resp = resp
	if err := req.Validate(); err != nil {
		blog.Errorf("get addons detail failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	// get addons
	addons := g.addons.FindByName(req.GetName())
	if addons == nil {
		blog.Errorf("get addons detail failed, %s", errorAddonsNotFound.Error())
		g.setResp(common.ErrHelmManagerGetActionFailed, errorAddonsNotFound.Error(), nil)
		return nil
	}

	clusterAddons := addons.ToAddonsProto()

	// get latest version
	version, err := g.getLatestVersion(ctx, addons.ChartName)
	if err != nil {
		blog.Errorf("get addons latest version failed, %s", err.Error())
	}
	clusterAddons.Version = &version

	// get current status
	rl, err := g.model.GetRelease(ctx, g.req.GetClusterID(), addons.Namespace, addons.ReleaseName())
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			g.setResp(common.ErrHelmManagerSuccess, "ok", clusterAddons)
			return nil
		}
		blog.Errorf("get addons status failed, %s", err.Error())
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}
	clusterAddons.CurrentVersion = &rl.ChartVersion
	clusterAddons.Status = &rl.Status
	clusterAddons.Message = &rl.Message
	clusterAddons.ReleaseName = &rl.Name
	if len(rl.Values) > 0 {
		clusterAddons.CurrentValues = &rl.Values[len(rl.Values)-1]
	}

	g.setResp(common.ErrHelmManagerSuccess, "ok", clusterAddons)
	return nil
}

func (g *GetAddonsDetailAction) getLatestVersion(ctx context.Context, chartName string) (string, error) {
	repository, err := g.model.GetProjectRepository(ctx, g.req.GetProjectCode(), common.PublicRepoName)
	if err != nil {
		return "", err
	}

	detail, err := g.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		GetChartDetail(ctx, chartName)
	if err != nil {
		return "", err
	}
	return detail.Version, nil
}

func (g *GetAddonsDetailAction) setResp(err common.HelmManagerError, message string, r *helmmanager.Addons) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
