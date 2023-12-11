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

package repository

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	defaultSize = 1000 // nolint
)

// NewListRepositoryAction return a new ListRepositoryAction instance
func NewListRepositoryAction(model store.HelmManagerModel) *ListRepositoryAction {
	return &ListRepositoryAction{
		model: model,
	}
}

// ListRepositoryAction provides the action to do list repositories
type ListRepositoryAction struct {
	ctx context.Context

	model store.HelmManagerModel

	req  *helmmanager.ListRepositoryReq
	resp *helmmanager.ListRepositoryResp
}

// Handle the listing process
func (l *ListRepositoryAction) Handle(ctx context.Context,
	req *helmmanager.ListRepositoryReq, resp *helmmanager.ListRepositoryResp) error {

	if req == nil || resp == nil {
		blog.Errorf("get repository failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list repository failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return l.list()
}

func (l *ListRepositoryAction) list() error {
	option := l.getOption()
	_, origin, err := l.model.ListRepository(l.ctx, l.getCondition(), option)
	if err != nil {
		blog.Errorf("list repository failed, %s", err.Error())
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	r := make([]*helmmanager.Repository, 0, len(origin))
	for _, item := range origin {
		r = append(r, item.Transfer2Proto(l.ctx))
	}

	l.setResp(common.ErrHelmManagerSuccess, "ok", r)
	blog.Infof("list repository successfully")
	return nil
}

func (l *ListRepositoryAction) getCondition() *operator.Condition {
	// get project repo
	projectCond := make(operator.M)
	projectCond.Update(entity.FieldKeyProjectID, contextx.GetProjectCodeFromCtx(l.ctx))
	projectCond.Update(entity.FieldKeyPersonal, false)
	projectCond2 := make(operator.M)
	projectCond2.Update(entity.FieldKeyProjectID, contextx.GetProjectCodeFromCtx(l.ctx))
	projectCond2.Update(entity.FieldKeyPersonal, nil)

	// get personal repo
	personalCond := make(operator.M)
	personalCond.Update(entity.FieldKeyProjectID, contextx.GetProjectCodeFromCtx(l.ctx))
	personalCond.Update(entity.FieldKeyPersonal, true)
	personalCond.Update(entity.FieldKeyCreateBy, auth.GetUserFromCtx(l.ctx))
	return operator.NewBranchCondition(operator.Or, operator.NewLeafCondition(operator.Eq, projectCond),
		operator.NewLeafCondition(operator.Eq, projectCond2),
		operator.NewLeafCondition(operator.Eq, personalCond))
}

func (l *ListRepositoryAction) getOption() *utils.ListOption {
	sortOpt := map[string]int{
		"public": 1,
	}

	return &utils.ListOption{
		Sort: sortOpt,
		Page: 0,
		Size: 0,
	}
}

func (l *ListRepositoryAction) setResp(err common.HelmManagerError, message string, r []*helmmanager.Repository) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
