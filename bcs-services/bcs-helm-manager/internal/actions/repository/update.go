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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewUpdateRepositoryAction return a new UpdateRepositoryAction instance
func NewUpdateRepositoryAction(model store.HelmManagerModel) *UpdateRepositoryAction {
	return &UpdateRepositoryAction{
		model: model,
	}
}

// UpdateRepositoryAction provides the action to do update repository
type UpdateRepositoryAction struct {
	ctx context.Context

	model store.HelmManagerModel

	req  *helmmanager.UpdateRepositoryReq
	resp *helmmanager.UpdateRepositoryResp
}

// Handle the updating process
func (u *UpdateRepositoryAction) Handle(ctx context.Context,
	req *helmmanager.UpdateRepositoryReq, resp *helmmanager.UpdateRepositoryResp) error {

	if req == nil || resp == nil {
		blog.Errorf("update repository failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	u.ctx = ctx
	u.req = req
	u.resp = resp

	if err := u.req.Validate(); err != nil {
		blog.Errorf("update repository failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	username := auth.GetUserFromCtx(ctx)
	return u.update(contextx.GetProjectCodeFromCtx(ctx), u.req.GetName(),
		(&entity.Repository{}).LoadFromProto(&helmmanager.Repository{
			Type:      u.req.Type,
			Remote:    u.req.Remote,
			RemoteURL: u.req.RemoteURL,
			Username:  u.req.Username,
			Password:  u.req.Password,
			UpdateBy:  &username,
		}))
}

func (u *UpdateRepositoryAction) update(projectCode, name string, m entity.M) error {
	if err := u.model.UpdateRepository(u.ctx, projectCode, name, m); err != nil {
		blog.Errorf("update repository failed, %s, projectCode: %s, name: %s", err.Error(), projectCode, name)
		u.setResp(common.ErrHelmManagerUpdateActionFailed, err.Error(), nil)
		return nil
	}

	r, err := u.model.GetProjectRepository(u.ctx, projectCode, name)
	if err != nil {
		blog.Errorf("update repository failed, %s, projectCode: %s, name: %s", err.Error(), projectCode, name)
		u.setResp(common.ErrHelmManagerUpdateActionFailed, err.Error(), nil)
		return nil
	}

	u.setResp(common.ErrHelmManagerSuccess, "ok", r.Transfer2Proto(u.ctx))
	blog.Infof("update repository successfully, projectCode: %s, name: %s", projectCode, name)
	return nil
}

func (u *UpdateRepositoryAction) setResp(err common.HelmManagerError, message string, r *helmmanager.Repository) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	u.resp.Code = &code
	u.resp.Message = &msg
	u.resp.Result = err.OK()
	u.resp.Data = r
}
