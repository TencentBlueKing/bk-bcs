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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/convert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// UpdateQuotaAction action for update project quota
type UpdateQuotaAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.UpdateProjectQuotaRequest
	resp  *proto.ProjectQuotaResponse
	user  string

	sQuota *quota.ProjectQuota
	task   *types.Task
	pQuota *proto.ProjectQuota
}

// NewUpdateQuotaAction new update project quota action
func NewUpdateQuotaAction(model store.ProjectModel) *UpdateQuotaAction {
	return &UpdateQuotaAction{
		model: model,
	}
}

// validate check project quota request
func (ua *UpdateQuotaAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	// 从 context 中获取 username
	authUser, err := middleware.GetUserFromContext(ua.ctx)
	if err == nil {
		ua.user = authUser.GetUsername()
	}

	// check quota exist
	sQuota, err := ua.model.GetProjectQuotaById(ua.ctx, ua.req.GetQuotaId())
	if err != nil {
		return err
	}
	ua.sQuota = sQuota

	return nil
}

// createProjectQuota update project quota && associate with provider
func (ua *UpdateQuotaAction) updateProjectQuota() error {
	ua.sQuota.UpdateTime = time.Now().Unix()
	ua.sQuota.Updater = ua.user

	if ua.req.GetName() != "" {
		ua.sQuota.QuotaName = ua.req.Name
	}
	if ua.req.GetQuota() != nil {
		ua.sQuota.Quota = quota.TransPorto2StoreQuota(ua.req.Quota)
	}

	err := ua.model.UpdateProjectQuota(ua.ctx, ua.sQuota)
	if err != nil {
		return errorx.NewDBErr(err.Error())
	}

	ua.pQuota = quota.TransStore2ProtoQuota(ua.sQuota)
	return nil
}

// dispatchTask dispatch quota task
func (ua *UpdateQuotaAction) dispatchTask() error {
	quotaMgr, err := manager.GetQuotaManager(ua.sQuota.Provider)
	if err != nil {
		return err
	}

	task, err := quotaMgr.UpdateProjectQuota(ua.pQuota.QuotaId, &provider.UpdateProjectQuotaOptions{})
	if err != nil {
		return err
	}
	ua.task = task

	if task != nil {
		err = manager.GetTaskServer().Dispatch(task)
		if err != nil {
			return err
		}
	}

	return nil
}

// Do update project request
func (ua *UpdateQuotaAction) Do(ctx context.Context,
	req *proto.UpdateProjectQuotaRequest, resp *proto.ProjectQuotaResponse) error {
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		return errorx.NewReadableErr(errorx.ParamErr, err.Error())
	}

	if err := ua.updateProjectQuota(); err != nil {
		return err
	}
	if err := ua.dispatchTask(); err != nil {
		return errorx.NewBuildTaskErr(err.Error())
	}

	// set resp data
	if ua.task != nil {
		task, err := convert.MarshalInterfaceToValue(ua.task)
		if err != nil {
			return err
		}
		resp.Task = task
	}
	resp.Data = ua.pQuota

	return nil
}
