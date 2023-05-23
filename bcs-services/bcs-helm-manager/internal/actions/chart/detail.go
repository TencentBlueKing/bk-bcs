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
	"fmt"
	"sort"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetVersionDetailV1Action return a new GetChartDetailV1Action instance
func NewGetVersionDetailV1Action(model store.HelmManagerModel, platform repo.Platform) *GetVersionDetailV1Action {
	return &GetVersionDetailV1Action{
		model:    model,
		platform: platform,
	}
}

// GetVersionDetailV1Action provides the action to do get chart version detail info
type GetVersionDetailV1Action struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.GetVersionDetailV1Req
	resp *helmmanager.GetVersionDetailV1Resp
}

// Handle the chart detail getting process
func (g *GetVersionDetailV1Action) Handle(ctx context.Context,
	req *helmmanager.GetVersionDetailV1Req, resp *helmmanager.GetVersionDetailV1Resp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get chart version detail failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	r, err := g.getDetail()
	if err != nil {
		g.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}
	g.setResp(common.ErrHelmManagerSuccess, "ok", r)
	return nil
}

func (g *GetVersionDetailV1Action) getDetail() (*helmmanager.ChartDetail, error) {
	projectCode := contextx.GetProjectCodeFromCtx(g.ctx)
	repoName := g.req.GetRepoName()
	chartName := g.req.GetName()
	version := g.req.GetVersion()
	username := auth.GetUserFromCtx(g.ctx)

	repository, err := g.model.GetProjectRepository(g.ctx, projectCode, repoName)
	if err != nil {
		blog.Errorf("get chart version detail failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, version, username)
		return nil, err
	}

	origin, err := g.platform.
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
		Detail(g.ctx, version)
	if err != nil {
		blog.Errorf("get chart version detail failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, version, username)
		return nil, err
	}

	valuesFile := getValuesFiles(origin.Contents, chartName)
	readmeFile := ""
	for _, item := range origin.Contents {
		if isReadMeFile(item) {
			readmeFile = item.Path
		}
	}

	r := origin.Transfer2Proto(repository.RepoURL)
	r.Readme = common.GetStringP(readmeFile)
	r.ValuesFile = valuesFile
	blog.Infof("get chart version detail successfully, "+
		"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
		projectCode, repoName, chartName, version, username)
	return r, nil
}

func (g *GetVersionDetailV1Action) setResp(err common.HelmManagerError, message string, r *helmmanager.ChartDetail) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}

// NewGetChartDetailV1Action return a new GetChartDetailV1Action instance
func NewGetChartDetailV1Action(model store.HelmManagerModel, platform repo.Platform) *GetChartDetailV1Action {
	return &GetChartDetailV1Action{
		model:    model,
		platform: platform,
	}
}

// GetChartDetailV1Action provides the action to do get chart detail info
type GetChartDetailV1Action struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.GetChartDetailV1Req
	resp *helmmanager.GetChartDetailV1Resp
}

// Handle the chart detail getting process
func (g *GetChartDetailV1Action) Handle(ctx context.Context,
	req *helmmanager.GetChartDetailV1Req, resp *helmmanager.GetChartDetailV1Resp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get chart detail failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	chart, err := g.getDetail()
	if err != nil {
		g.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}
	g.setResp(common.ErrHelmManagerSuccess, "ok", chart)
	return nil
}

func (g *GetChartDetailV1Action) getDetail() (*helmmanager.Chart, error) {
	projectCode := contextx.GetProjectCodeFromCtx(g.ctx)
	repoName := g.req.GetRepoName()
	chartName := g.req.GetName()
	username := auth.GetUserFromCtx(g.ctx)

	repository, err := g.model.GetProjectRepository(g.ctx, projectCode, repoName)
	if err != nil {
		blog.Errorf("get chart detail failed, %s, projectCode: %s, repository: %s, chartName: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, username)
		return nil, err
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
		GetChartDetail(g.ctx, chartName)
	if err != nil {
		blog.Errorf("get chart detail failed, %s, projectCode: %s, repository: %s, chartName: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, username)
		return nil, err
	}
	chart := detail.Transfer2Proto()
	chart.ProjectID = common.GetStringP(projectCode)
	chart.ProjectCode = common.GetStringP(projectCode)
	chart.Repository = common.GetStringP(repoName)

	blog.Infof("get chart detail successfully, projectCode: %s, repository: %s, chartName: %s, operator: %s",
		projectCode, repoName, chartName, username)
	return chart, nil
}

func (g *GetChartDetailV1Action) setResp(err common.HelmManagerError, message string, r *helmmanager.Chart) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}

func getValuesFiles(files map[string]*repo.FileContent, chartName string) []string {
	valuesFile := make([]string, 0, 0)
	defaultValuesFile := fmt.Sprintf("%s/%s", chartName, "values.yaml")
	hasDefaultValuesFile := false
	for _, item := range files {
		if item.Path == defaultValuesFile {
			hasDefaultValuesFile = true
			continue
		}
		if isValuesFile(item) {
			valuesFile = append(valuesFile, item.Path)
		}
	}

	sort.Sort(sort.StringSlice(valuesFile))
	if hasDefaultValuesFile {
		valuesFile = append([]string{defaultValuesFile}, valuesFile...)
	}
	return valuesFile
}

func isValuesFile(f *repo.FileContent) bool {
	// 允许根目录所有以values.yaml结尾的文件, 如
	// values.yaml
	// game-values.yaml
	// my-values.yaml
	if (strings.HasSuffix(f.Name, "values.yaml") || strings.HasSuffix(f.Name, "values.yml")) &&
		strings.Count(f.Path, "/") <= 1 {
		return true
	}

	// 允许所有在bcs-values文件夹下的.yml或.yaml文件, 如
	// bcs-values/values.yaml
	// templates/bcs-values/my.yaml
	if strings.HasSuffix(strings.TrimSuffix(f.Path, f.Name), "bcs-values/") &&
		(strings.HasSuffix(f.Path, ".yaml") || strings.HasSuffix(f.Path, ".yml")) {
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
