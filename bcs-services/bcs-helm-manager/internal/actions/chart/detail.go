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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetChartDetailAction return a new GetChartDetailAction instance
func NewGetChartDetailAction(model store.HelmManagerModel, platform repo.Platform) *GetChartDetailAction {
	return &GetChartDetailAction{
		model:    model,
		platform: platform,
	}
}

// GetChartDetailAction provides the action to do get chart detail info
type GetChartDetailAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.GetChartDetailReq
	resp *helmmanager.GetChartDetailResp
}

// Handle the chart detail getting process
func (g *GetChartDetailAction) Handle(ctx context.Context,
	req *helmmanager.GetChartDetailReq, resp *helmmanager.GetChartDetailResp) error {

	if req == nil || resp == nil {
		blog.Errorf("get chart detail failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get chart detail failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return g.getDetail()
}

func (g *GetChartDetailAction) getDetail() error {
	projectID := g.req.GetProjectID()
	repoName := g.req.GetRepository()
	chartName := g.req.GetName()
	version := g.req.GetVersion()
	opName := g.req.GetOperator()

	repository, err := g.model.GetRepository(g.ctx, projectID, repoName)
	if err != nil {
		blog.Errorf("get chart detail failed, %s, "+
			"projectID: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectID, repoName, chartName, version, opName)
		g.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	origin, err := g.platform.
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
		Detail(g.ctx, version)
	if err != nil {
		blog.Errorf("get chart detail failed, %s, "+
			"projectID: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectID, repoName, chartName, version, opName)
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	valuesFile := make([]string, 0, 0)
	readmeFile := ""
	for _, item := range origin.Contents {
		if isValuesFile(item) {
			valuesFile = append(valuesFile, item.Path)
		}
		if isReadMeFile(item) {
			readmeFile = item.Path
		}

		if len(item.Content) > 1024*100 {
			item.Content = nil
		}
	}

	r := origin.Transfer2Proto()
	r.Readme = common.GetStringP(readmeFile)
	r.ValuesFile = valuesFile
	g.setResp(common.ErrHelmManagerSuccess, "ok", r)
	blog.Infof("get chart detail successfully, "+
		"projectID: %s, repository: %s, chartName: %s, version: %s, operator: %s",
		projectID, repoName, chartName, version, opName)
	return nil
}

func (g *GetChartDetailAction) setResp(err common.HelmManagerError, message string, r *helmmanager.ChartDetail) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}

func isValuesFile(f *repo.FileContent) bool {
	// 允许所有以values.yaml结尾的文件, 如
	// values.yaml
	// game-values.yaml
	// my-values.yaml
	if strings.HasSuffix(f.Name, "values.yaml") {
		return true
	}

	// 允许所有在bcs-values文件夹下的文件, 如
	// bcs-values/values.yaml
	// templates/bcs-values/my.yaml
	if strings.HasSuffix(strings.TrimSuffix(f.Path, f.Name), "bcs-values/") {
		return true
	}

	return false
}

func isReadMeFile(f *repo.FileContent) bool {
	// README.md 一般以最外层为准, 目录层级不能超过1
	if f.Name == "README.md" && strings.Count(f.Path, "/") <= 1 {
		return true
	}

	return false
}
