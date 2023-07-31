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
 *
 */

package chart

import (
	"context"
	"fmt"
	"os"

	"github.com/chartmuseum/helm-push/pkg/helm"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewUploadChartAction return a new UploadChartAction instance
func NewUploadChartAction(model store.HelmManagerModel, platform repo.Platform) *UploadChartAction {
	return &UploadChartAction{
		model:    model,
		platform: platform,
	}
}

// UploadChartAction provides the action to do upload chart
type UploadChartAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.UploadChartReq
	resp *helmmanager.UploadChartResp
}

// Handle the chart upload process
func (d *UploadChartAction) Handle(ctx context.Context,
	req *helmmanager.UploadChartReq, resp *helmmanager.UploadChartResp) error {
	d.ctx = ctx
	d.req = req
	d.resp = resp
	username := auth.GetUserFromCtx(d.ctx)

	if err := d.req.Validate(); err != nil {
		blog.Errorf("upload chart failed, invalid request, %s, param: %v", err.Error(), d.req)
		return err
	}
	chartName, err := d.uploadChart()
	if err != nil {
		blog.Errorf("upload chart failed, %s, "+
			"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
			err.Error(), d.req.GetProjectCode(), d.req.GetRepoName(), chartName, d.req.GetVersion(), username)
		d.setResp(common.ErrHelmManagerUploadChartFailed, err.Error())
		return nil
	}
	d.setResp(common.ErrHelmManagerSuccess, "ok")
	blog.Infof("upload chart successfully, "+
		"projectCode: %s, repository: %s, chartName: %s, version: %s, operator: %s",
		d.req.GetProjectCode(), d.req.GetRepoName(), chartName, d.req.GetVersion(), username)
	return nil
}

func (d *UploadChartAction) uploadChart() (string, error) {
	projectCode := contextx.GetProjectCodeFromCtx(d.ctx)
	repoName := d.req.GetRepoName()
	version := d.req.GetVersion()
	data := d.req.GetFile()
	force := d.req.GetForce()
	username := auth.GetUserFromCtx(d.ctx)
	tmp, err := generateChartPackage(data)
	defer func(path string) {
		// 删除临时文件
		err := os.RemoveAll(path)
		if err != nil {
			blog.Errorf("failed to remove temporary file, %s: %s",
				path, err.Error())
		}
	}(tmp.Name())
	if err != nil {
		return "", err
	}
	// 获取chat信息
	chart, err := helm.GetChartByName(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("failed get chart by name, %s", err)
	}

	// 如果repoName不等于public-repo 则项目仓库
	// 获取仓库上传地址和账号密码
	repository, err := d.model.GetProjectRepository(d.ctx, projectCode, repoName)
	if err != nil {
		blog.Errorf("upload chart failed, %s, "+
			"projectCode: %s, repository: %s, version: %s, operator: %s",
			err.Error(), projectCode, repoName, version, username)
		return chart.Name(), err
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
		Chart(chart.Name()).
		Upload(d.ctx, repo.UploadOption{
			ProjectCode: projectCode,
			RepoName:    repoName,
			Version:     version,
			Force:       force,
			ChartPath:   tmp.Name(),
		})
	return chart.Name(), err
}

func (d *UploadChartAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	d.resp.Code = &code
	d.resp.Message = &msg
	d.resp.Result = err.OK()
}

// 生成临时chart包
func generateChartPackage(content []byte) (tmp *os.File, err error) {
	// 创建在系统默认临时目录下的临时文件
	tmp, err = os.CreateTemp("", "helm-chart-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file, %s", err)
	}
	_, err = tmp.Write(content)
	defer func(tmp *os.File) {
		// 关闭文件
		err := tmp.Close()
		if err != nil {
			blog.Errorf("failed closes the File, %s: %s",
				tmp.Name(), err.Error())
		}
	}(tmp)
	if err != nil {
		return nil, fmt.Errorf("error writing to temp file, %s", err)
	}
	return tmp, nil
}
