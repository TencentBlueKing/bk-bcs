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

// Package quota xxx
package quota

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/convert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// DeleteQuotaAction action for delete project quota
type DeleteQuotaAction struct {
	ctx    context.Context
	model  store.ProjectModel
	req    *proto.DeleteProjectQuotaRequest
	resp   *proto.ProjectQuotaResponse
	user   string
	sQuota *quota.ProjectQuota
	task   *types.Task
	pQuota *proto.ProjectQuota
}

// NewDeleteQuotaAction new delete project quota action
func NewDeleteQuotaAction(model store.ProjectModel) *DeleteQuotaAction {
	return &DeleteQuotaAction{
		model: model,
	}
}

// validate check project quota request
func (da *DeleteQuotaAction) validate() error {
	err := da.req.Validate()
	if err != nil {
		return err
	}

	// 从 context 中获取 username
	authUser, err := middleware.GetUserFromContext(da.ctx)
	if err == nil {
		da.user = authUser.GetUsername()
	}

	// check quota exist
	sQuota, err := da.model.GetProjectQuotaById(da.ctx, da.req.GetQuotaId())
	if err != nil {
		return err
	}
	da.sQuota = sQuota

	// check quota status
	err = da.checkProjectQuotaStatus()
	if err != nil {
		return err
	}

	return nil
}

func (da *DeleteQuotaAction) checkProjectQuotaStatus() error {
	if da.sQuota.Status == quota.Deleting {
		return errorx.NewCheckQuotaStatusErr("project quota status is DELETING")
	}

	if da.sQuota.Status != quota.Running {
		return errorx.NewCheckQuotaStatusErr("project quota status is not RUNNING")
	}

	return nil
}

// createProjectQuota create project quota && associate with provider
func (da *DeleteQuotaAction) updateProjectQuota() error {
	da.sQuota.Status = quota.Deleting

	err := da.model.UpdateProjectQuotaByField(da.ctx, entity.M{
		quota.FieldKeyQuotaId: da.sQuota.QuotaId,
		quota.FieldKeyStatus:  quota.Deleting.String(),
	})
	if err != nil {
		return errorx.NewDBErr(err.Error())
	}

	da.pQuota = quota.TransStore2ProtoQuota(da.sQuota)

	return nil
}

// dispatchTask dispatch quota task
func (da *DeleteQuotaAction) dispatchTask() error {
	quotaMgr, err := manager.GetQuotaManager(da.sQuota.Provider)
	if err != nil {
		return err
	}

	task, err := quotaMgr.DeleteProjectQuota(da.pQuota.QuotaId, &provider.DeleteProjectQuotaOptions{
		Operator: da.user,
	})
	if err != nil {
		return err
	}
	da.task = task

	err = manager.GetTaskServer().Dispatch(task)
	if err != nil {
		return err
	}

	return nil
}

// Do delete project request
func (da *DeleteQuotaAction) Do(ctx context.Context,
	req *proto.DeleteProjectQuotaRequest, resp *proto.ProjectQuotaResponse) error {
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		return errorx.NewReadableErr(errorx.ParamErr, err.Error())
	}

	if err := da.updateProjectQuota(); err != nil {
		return err
	}
	if err := da.dispatchTask(); err != nil {
		return errorx.NewBuildTaskErr(err.Error())
	}

	// set resp data
	task, err := convert.MarshalInterfaceToValue(da.task)
	if err != nil {
		return err
	}
	resp.Task = task
	resp.Data = da.pQuota

	return nil
}
