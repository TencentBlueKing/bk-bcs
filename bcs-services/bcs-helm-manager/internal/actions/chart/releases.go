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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetChartReleaseAction return a new GetChartReleaseAction instance
func NewGetChartReleaseAction(model store.HelmManagerModel) *GetChartReleaseAction {
	return &GetChartReleaseAction{
		model: model,
	}
}

// GetChartReleaseAction provides the action to do list chart releases
type GetChartReleaseAction struct {
	ctx context.Context

	model store.HelmManagerModel

	req  *helmmanager.GetChartReleaseReq
	resp *helmmanager.GetChartReleaseResp
}

// Handle the listing process
func (l *GetChartReleaseAction) Handle(ctx context.Context,
	req *helmmanager.GetChartReleaseReq, resp *helmmanager.GetChartReleaseResp) error {
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list chart releases failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	result, err := l.list()
	if err != nil {
		blog.Errorf("list chart release failed, %s, projectCode: %s, repository: %s, chart: %s",
			err.Error(), l.req.GetProjectCode(), l.req.GetRepoName(), l.req.GetName())
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}
	blog.Errorf("list chart release successfully, projectCode: %s, repository: %s, chart: %s",
		l.req.GetProjectCode(), l.req.GetRepoName(), l.req.GetName())
	l.setResp(common.ErrHelmManagerSuccess, "ok", result)
	return nil
}

func (l *GetChartReleaseAction) list() ([]*helmmanager.Release, error) {
	_, result, err := l.model.ListRelease(l.ctx, l.getCondition(), &utils.ListOption{})
	if err != nil {
		return nil, err
	}

	releases := make([]*helmmanager.Release, 0)
	for _, item := range result {
		releases = append(releases, item.Transfer2Proto())
	}
	return releases, nil
}

func (l *GetChartReleaseAction) getCondition() *operator.Condition {
	cond := make(operator.M)
	if l.req.ProjectCode != nil {
		cond.Update(entity.FieldKeyProjectCode, contextx.GetProjectCodeFromCtx(l.ctx))
	}
	if l.req.RepoName != nil {
		cond.Update(entity.FieldKeyRepoName, l.req.GetRepoName())
	}
	if l.req.Name != nil {
		cond.Update(entity.FieldKeyChartName, l.req.GetName())
	}

	condIN := make(operator.M)
	if len(l.req.GetVersions()) != 0 {
		condIN.Update(entity.FieldKeyChartVersion, l.req.GetVersions())
	}

	condEq := operator.NewLeafCondition(operator.Eq, cond)
	condIn := operator.NewLeafCondition(operator.In, condIN)
	return operator.NewBranchCondition(operator.And, condEq, condIn)
}

func (l *GetChartReleaseAction) setResp(err common.HelmManagerError, message string, r []*helmmanager.Release) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
