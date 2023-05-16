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

package chart

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewDeleteChartAction return a new DeleteChartAction instance
func NewDeleteChartAction(model store.HelmManagerModel, platform repo.Platform) *DeleteChartAction {
	return &DeleteChartAction{
		model:    model,
		platform: platform,
	}
}

// DeleteChartAction provides the action to do delete chart
type DeleteChartAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.DeleteChartReq
	resp *helmmanager.DeleteChartResp
}

// Handle the chart deleting process
func (d *DeleteChartAction) Handle(ctx context.Context,
	req *helmmanager.DeleteChartReq, resp *helmmanager.DeleteChartResp) error {
	d.ctx = ctx
	d.req = req
	d.resp = resp

	if err := d.req.Validate(); err != nil {
		blog.Errorf("delete chart failed, invalid request, %s, param: %v", err.Error(), d.req)
		d.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	return d.deleteChart()
}

func (d *DeleteChartAction) deleteChart() error {
	projectCode := contextx.GetProjectCodeFromCtx(d.ctx)
	repoName := d.req.GetRepoName()
	chartName := d.req.GetName()
	username := auth.GetUserFromCtx(d.ctx)

	repository, err := d.model.GetRepository(d.ctx, projectCode, repoName)
	if err != nil {
		blog.Errorf("delete chart failed, %s, projectCode: %s, repository: %s, chartName: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, username)
		d.setResp(common.ErrHelmManagerListActionFailed, err.Error())
		return nil
	}

	err = d.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(chartName).
		Delete(d.ctx)
	if err != nil {
		blog.Errorf("delete chart failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, username)
		d.setResp(common.ErrHelmManagerGetActionFailed, err.Error())
		return nil
	}

	d.setResp(common.ErrHelmManagerSuccess, "ok")
	blog.Infof("delete chart successfully, projectCode: %s, repository: %s, chartName: %s, operator: %s",
		projectCode, repoName, chartName, username)
	return nil
}

func (d *DeleteChartAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	d.resp.Code = &code
	d.resp.Message = &msg
	d.resp.Result = err.OK()
}

// NewDeleteChartVersionAction return a new DeleteChartVersionAction instance
func NewDeleteChartVersionAction(model store.HelmManagerModel, platform repo.Platform) *DeleteChartVersionAction {
	return &DeleteChartVersionAction{
		model:    model,
		platform: platform,
	}
}

// DeleteChartVersionAction provides the action to do delete chart version
type DeleteChartVersionAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.DeleteChartVersionReq
	resp *helmmanager.DeleteChartVersionResp
}

// Handle the chart version deleting process
func (d *DeleteChartVersionAction) Handle(ctx context.Context,
	req *helmmanager.DeleteChartVersionReq, resp *helmmanager.DeleteChartVersionResp) error {
	d.ctx = ctx
	d.req = req
	d.resp = resp

	if err := d.req.Validate(); err != nil {
		blog.Errorf("delete chart version failed, invalid request, %s, param: %v", err.Error(), d.req)
		d.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}

	return d.deleteChartVersion()
}

func (d *DeleteChartVersionAction) deleteChartVersion() error {
	projectCode := contextx.GetProjectCodeFromCtx(d.ctx)
	repoName := d.req.GetRepoName()
	chartName := d.req.GetName()
	version := d.req.GetVersion()
	username := auth.GetUserFromCtx(d.ctx)

	repository, err := d.model.GetRepository(d.ctx, projectCode, repoName)
	if err != nil {
		blog.Errorf("delete chart version failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, version, username)
		d.setResp(common.ErrHelmManagerListActionFailed, err.Error())
		return nil
	}

	err = d.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(chartName).
		DeleteVersion(d.ctx, version)
	if err != nil {
		blog.Errorf("delete chart version failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, version, username)
		d.setResp(common.ErrHelmManagerGetActionFailed, err.Error())
		return nil
	}

	d.setResp(common.ErrHelmManagerSuccess, "ok")
	blog.Infof("delete chart version successfully, "+
		"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
		projectCode, repoName, chartName, version, username)
	return nil
}

func (d *DeleteChartVersionAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	d.resp.Code = &code
	d.resp.Message = &msg
	d.resp.Result = err.OK()
}
