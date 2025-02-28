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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ScaleDownQuotaAction action for scale down project quota
type ScaleDownQuotaAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ScaleDownProjectQuotaRequest
	resp  *proto.ScaleDownProjectQuotaResponse
	user  string

	sQuota *quota.ProjectQuota
	task   *types.Task
}

// NewScaleDownQuotaAction new scale down project quota action
func NewScaleDownQuotaAction(model store.ProjectModel) *ScaleDownQuotaAction {
	return &ScaleDownQuotaAction{
		model: model,
	}
}

// validate check project quota request
func (sa *ScaleDownQuotaAction) validate() error {
	err := sa.req.Validate()
	if err != nil {
		return err
	}

	// 从 context 中获取 username
	authUser, err := middleware.GetUserFromContext(sa.ctx)
	if err == nil {
		sa.user = authUser.GetUsername()
	}

	// check quota exist
	sQuota, err := sa.model.GetProjectQuotaById(sa.ctx, sa.req.GetQuotaId())
	if err != nil {
		return err
	}
	sa.sQuota = sQuota

	// check update params
	switch sQuota.QuotaType {
	case quota.Host:
		if sa.req.GetQuota().GetZoneResources() == nil {
			return errorx.NewParamErr("zoneResources is required")
		}
	case quota.Shared:
		if sa.req.GetQuota().GetCpu() == nil && sa.req.GetQuota().GetMem() == nil {
			return errorx.NewParamErr("cpu or mem is required")
		}
	case quota.Federation:
		if sa.req.GetQuota().GetCpu() == nil && sa.req.GetQuota().GetMem() == nil &&
			sa.req.GetQuota().GetGpu() == nil {
			return errorx.NewParamErr("cpu or mem or gpu is required")
		}
	default:
		return errorx.NewParamErr("quotaType is invalid")
	}

	// quota validate

	return nil
}

// dispatchTask dispatch quota task
func (sa *ScaleDownQuotaAction) dispatchTask() error {
	quotaMgr, err := manager.GetQuotaManager(sa.sQuota.Provider)
	if err != nil {
		return err
	}

	task, err := quotaMgr.ScaleDownProjectQuota(sa.sQuota.QuotaId, sa.req.GetQuota(),
		&provider.ScaleDownProjectQuotaOptions{
			Operator: sa.user,
		})
	if err != nil {
		return err
	}
	sa.task = task

	err = manager.GetTaskServer().Dispatch(task)
	if err != nil {
		return err
	}

	return nil
}

// Do scale up project request
func (sa *ScaleDownQuotaAction) Do(ctx context.Context,
	req *proto.ScaleDownProjectQuotaRequest, resp *proto.ScaleDownProjectQuotaResponse) error {
	sa.ctx = ctx
	sa.req = req
	sa.resp = resp

	if err := sa.validate(); err != nil {
		return errorx.NewReadableErr(errorx.ParamErr, err.Error())
	}

	if err := sa.dispatchTask(); err != nil {
		return errorx.NewBuildTaskErr(err.Error())
	}

	t := getTaskWithSN(sa.task.TaskID)
	if t != nil {
		sa.task = t
	}

	// set resp data
	task, err := convert.MarshalInterfaceToValue(sa.task)
	if err != nil {
		return err
	}
	resp.Task = task

	return nil
}
