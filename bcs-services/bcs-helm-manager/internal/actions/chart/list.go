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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	defaultSize = 5000
)

// NewListChartV1Action return a new ListChartActionV1 instance
func NewListChartV1Action(model store.HelmManagerModel, platform repo.Platform) *ListChartActionV1 {
	return &ListChartActionV1{
		model:    model,
		platform: platform,
	}
}

// ListChartActionV1 provides the action to do list charts
type ListChartActionV1 struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.ListChartV1Req
	resp *helmmanager.ListChartV1Resp
}

// Handle the listing process
func (l *ListChartActionV1) Handle(ctx context.Context,
	req *helmmanager.ListChartV1Req, resp *helmmanager.ListChartV1Resp) error {
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list chart failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return l.list()
}

func (l *ListChartActionV1) list() error {
	projectCode := contextx.GetProjectCodeFromCtx(l.ctx)
	repoName := l.req.GetRepoName()
	username := auth.GetUserFromCtx(l.ctx)

	repository, err := l.model.GetProjectRepository(l.ctx, projectCode, repoName)
	if err != nil {
		blog.Errorf("list chart failed, %s, projectCode: %s, repository: %s, operator: %s",
			err.Error(), projectCode, repoName, username)
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	origin, err := l.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		SearchChart(l.ctx, l.getOption())
	if err != nil {
		blog.Errorf("list chart failed, %s, projectCode: %s, repository: %s, operator: %s",
			err.Error(), projectCode, repoName, username)
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	r := make([]*helmmanager.Chart, 0, len(origin.Charts))
	for _, item := range origin.Charts {
		chart := item.Transfer2Proto()
		chart.ProjectID = common.GetStringP(projectCode)
		chart.ProjectCode = common.GetStringP(projectCode)
		chart.Repository = common.GetStringP(repoName)
		r = append(r, chart)
	}

	l.setResp(common.ErrHelmManagerSuccess, "ok", &helmmanager.ChartListData{
		Page:  common.GetUint32P(uint32(origin.Page)),
		Size:  common.GetUint32P(uint32(origin.Size)),
		Total: common.GetUint32P(uint32(origin.Total)),
		Data:  r,
	})
	blog.Infof("list chart successfully")
	return nil
}

func (l *ListChartActionV1) getCondition() *operator.Condition {
	cond := make(operator.M)
	if l.req.ProjectCode != nil {
		cond.Update(entity.FieldKeyProjectID, contextx.GetProjectCodeFromCtx(l.ctx))
	}
	if l.req.RepoName != nil {
		cond.Update(entity.FieldKeyName, l.req.GetRepoName())
	}

	return operator.NewLeafCondition(operator.Eq, cond)
}

func (l *ListChartActionV1) getOption() repo.ListOption {
	size := l.req.GetSize()
	if size == 0 {
		size = defaultSize
	}

	return repo.ListOption{
		PackageName: strings.ToLower(l.req.GetName()),
		Page:        int64(l.req.GetPage()),
		Size:        int64(size),
	}
}

func (l *ListChartActionV1) setResp(err common.HelmManagerError, message string, r *helmmanager.ChartListData) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
