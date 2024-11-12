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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetAction action for get project
type GetAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.GetProjectQuotaRequest
	resp  *proto.ProjectQuotaResponse
}

// NewGetAction new get project action
func NewGetAction(model store.ProjectModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

// Do get project info
func (ga *GetAction) Do(ctx context.Context, req *proto.GetProjectQuotaRequest,
	resp *proto.ProjectQuotaResponse) error {
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	sQuota, err := ga.model.GetProjectQuotaById(ga.ctx, ga.req.GetQuotaId())
	if err != nil {
		return errorx.NewDBErr(err.Error())
	}
	pQuota := quota.TransStore2ProtoQuota(sQuota)

	ga.resp.Data = pQuota

	return nil
}
