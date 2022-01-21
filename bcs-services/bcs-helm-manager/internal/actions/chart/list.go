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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	defaultSize = 1000
)

// NewListChartAction return a new ListChartAction instance
func NewListChartAction(model store.HelmManagerModel, platform repo.Platform) *ListChartAction {
	return &ListChartAction{
		model:    model,
		platform: platform,
	}
}

// ListChartAction provides the action to do list charts
type ListChartAction struct {
	ctx context.Context

	model    store.HelmManagerModel
	platform repo.Platform

	req  *helmmanager.ListChartReq
	resp *helmmanager.ListChartResp
}

// Handle the listing process
func (l *ListChartAction) Handle(ctx context.Context,
	req *helmmanager.ListChartReq, resp *helmmanager.ListChartResp) error {

	if req == nil || resp == nil {
		blog.Errorf("list chart failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
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

func (l *ListChartAction) list() error {
	projectID := l.req.GetProjectID()
	repoName := l.req.GetRepository()
	opName := l.req.GetOperator()

	repository, err := l.model.GetRepository(l.ctx, projectID, repoName)
	if err != nil {
		blog.Errorf("list chart failed, %s, projectID: %s, repository: %s, operator: %s",
			err.Error(), projectID, repoName, opName)
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	origin, err := l.platform.
		User(repository.Username).
		Project(repository.ProjectID).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.Name,
		).
		ListChart(l.ctx, l.getOption())
	if err != nil {
		blog.Errorf("list chart failed, %s, projectID: %s, repository: %s, operator: %s",
			err.Error(), projectID, repoName, opName)
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	r := make([]*helmmanager.Chart, 0, len(origin.Charts))
	for _, item := range origin.Charts {
		chart := item.Transfer2Proto()
		chart.ProjectID = common.GetStringP(projectID)
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

func (l *ListChartAction) getCondition() *operator.Condition {
	cond := make(operator.M)
	if l.req.ProjectID != nil {
		cond.Update(entity.FieldKeyProjectID, l.req.GetProjectID())
	}
	if l.req.Repository != nil {
		cond.Update(entity.FieldKeyName, l.req.GetRepository())
	}

	return operator.NewLeafCondition(operator.Eq, cond)
}

func (l *ListChartAction) getOption() repo.ListOption {
	size := l.req.GetSize()
	if size == 0 {
		size = defaultSize
	}

	return repo.ListOption{
		Page: int64(l.req.GetPage()),
		Size: int64(size),
	}
}

func (l *ListChartAction) setResp(err common.HelmManagerError, message string, r *helmmanager.ChartListData) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
