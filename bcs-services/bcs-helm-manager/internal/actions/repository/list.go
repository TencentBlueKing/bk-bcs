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

package repository

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	defaultSize = 1000
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
	total, origin, err := l.model.ListRepository(l.ctx, l.getCondition(), option)
	if err != nil {
		blog.Errorf("list repository failed, %s", err.Error())
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	r := make([]*helmmanager.Repository, 0, len(origin))
	for _, item := range origin {
		r = append(r, item.Transfer2Proto())
	}

	l.setResp(common.ErrHelmManagerSuccess, "ok", &helmmanager.RepositoryListData{
		Page:  common.GetUint32P(uint32(option.Page)),
		Size:  common.GetUint32P(uint32(option.Size)),
		Total: common.GetUint32P(uint32(total)),
		Data:  r,
	})
	blog.Infof("list repository successfully")
	return nil
}

func (l *ListRepositoryAction) getCondition() *operator.Condition {
	cond := make(operator.M)
	if l.req.ProjectID != nil {
		cond.Update(entity.FieldKeyProjectID, l.req.GetProjectID())
	}
	if l.req.Name != nil {
		cond.Update(entity.FieldKeyName, l.req.GetName())
	}
	if l.req.Type != nil {
		cond.Update(entity.FieldKeyType, l.req.GetType())
	}

	return operator.NewLeafCondition(operator.Eq, cond)
}

func (l *ListRepositoryAction) getOption() *utils.ListOption {
	var sortOpt map[string]int
	if sort := l.req.GetSort(); len(sort) > 0 {
		v := 1
		if !l.req.GetDesc() {
			v = -1
		}
		sortOpt = map[string]int{sort: v}
	}

	size := l.req.GetSize()
	if size == 0 {
		size = defaultSize
	}

	return &utils.ListOption{
		Sort: sortOpt,
		Page: int64(l.req.GetPage()),
		Size: int64(size),
	}
}

func (l *ListRepositoryAction) setResp(err common.HelmManagerError, message string, r *helmmanager.RepositoryListData) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
