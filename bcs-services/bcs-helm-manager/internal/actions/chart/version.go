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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewListChartVersionAction return a new ListChartVersionAction instance
func NewListChartVersionAction(model store.HelmManagerModel, platform repo.Platform) *ListChartVersionAction {
	return &ListChartVersionAction{
		model:    model,
		platform: platform,
	}
}

// ListChartVersionAction provides the action to do list chart versions
type ListChartVersionAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.ListChartVersionReq
	resp *helmmanager.ListChartVersionResp
}

// Handle the version listing process
func (l *ListChartVersionAction) Handle(ctx context.Context,
	req *helmmanager.ListChartVersionReq, resp *helmmanager.ListChartVersionResp) error {

	if req == nil || resp == nil {
		blog.Errorf("list chart version failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list chart version failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return l.list()
}

func (l *ListChartVersionAction) list() error {
	projectID := l.req.GetProjectID()
	repoName := l.req.GetRepository()
	chartName := l.req.GetName()
	opName := l.req.GetOperator()

	repository, err := l.model.GetRepository(l.ctx, projectID, repoName)
	if err != nil {
		blog.Errorf(
			"list chart version failed, %s, projectID: %s, repository: %s, chartName: %s, operator: %s",
			err.Error(), projectID, repoName, chartName, opName)
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	origin, err := l.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.ProjectID).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.Name,
		).
		Chart(chartName).
		ListVersion(l.ctx, l.getOption())
	if err != nil {
		blog.Errorf("list chart version failed, %s, projectID: %s, repository: %s, chartName: %s, operator: %s",
			err.Error(), projectID, repoName, chartName, opName)
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	r := make([]*helmmanager.ChartVersion, 0, len(origin.Versions))
	for _, item := range origin.Versions {
		chart := item.Transfer2Proto()
		r = append(r, chart)
	}
	l.setResp(common.ErrHelmManagerSuccess, "ok", &helmmanager.ChartVersionListData{
		Page:  common.GetUint32P(uint32(origin.Page)),
		Size:  common.GetUint32P(uint32(origin.Size)),
		Total: common.GetUint32P(uint32(origin.Total)),
		Data:  r,
	})
	blog.Infof("list chart version successfully")
	return nil
}

func (l *ListChartVersionAction) getOption() repo.ListOption {
	size := l.req.GetSize()
	if size == 0 {
		size = defaultSize
	}

	return repo.ListOption{
		Page: int64(l.req.GetPage()),
		Size: int64(size),
	}
}

func (l *ListChartVersionAction) setResp(
	err common.HelmManagerError, message string, r *helmmanager.ChartVersionListData) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
