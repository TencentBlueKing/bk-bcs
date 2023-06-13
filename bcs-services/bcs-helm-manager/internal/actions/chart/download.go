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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewDownloadChartAction return a new DownloadChartAction instance
func NewDownloadChartAction(model store.HelmManagerModel, platform repo.Platform) *DownloadChartAction {
	return &DownloadChartAction{
		model:    model,
		platform: platform,
	}
}

// DownloadChartAction provides the action to do download chart
type DownloadChartAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.DownloadChartReq
	resp *httpbody.HttpBody
}

// Handle the chart download process
func (d *DownloadChartAction) Handle(ctx context.Context,
	req *helmmanager.DownloadChartReq, resp *httpbody.HttpBody) error {
	d.ctx = ctx
	d.req = req
	d.resp = resp

	if err := d.req.Validate(); err != nil {
		blog.Errorf("download chart failed, invalid request, %s, param: %v", err.Error(), d.req)
		return err
	}

	return d.downloadChart()
}

func (d *DownloadChartAction) downloadChart() error {
	projectCode := contextx.GetProjectCodeFromCtx(d.ctx)
	repoName := d.req.GetRepoName()
	chartName := d.req.GetName()
	version := d.req.GetVersion()
	username := auth.GetUserFromCtx(d.ctx)

	repository, err := d.model.GetProjectRepository(d.ctx, projectCode, repoName)
	if err != nil {
		blog.Errorf("download chart failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, version, username)
		return err
	}

	content, err := d.platform.
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
		Download(d.ctx, version)
	if err != nil {
		blog.Errorf("download chart failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), projectCode, repoName, chartName, version, username)
		return err
	}

	d.resp.ContentType = "application/octet-stream"
	d.resp.Data = content
	grpc.SendHeader(d.ctx, metadata.New(map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=%s-%s.tgz", chartName, version),
	}))
	blog.Infof("download chart successfully, "+
		"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
		projectCode, repoName, chartName, version, username)
	return nil
}
