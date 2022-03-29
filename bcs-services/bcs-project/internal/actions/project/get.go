/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// GetAction action for get project
type GetAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.GetProjectRequest
}

// NewGetAction new get project action
func NewGetAction(model store.ProjectModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

// Handle get project info
func (ga *GetAction) Handle(ctx context.Context, req *proto.GetProjectRequest, resp *proto.ProjectResponse) {
	if req == nil || resp == nil {
		return
	}
	ga.ctx = ctx
	ga.req = req

	p, err := ga.model.GetProject(ctx, req.ProjectIdOrCode)
	if err != nil {
		setResp(resp, common.BcsProjectDbErr, common.BcsProjectDbErrMsg, err.Error(), nil)
		return
	}

	setResp(resp, common.BcsProjectSuccess, "", common.BcsProjectSuccessMsg, p)
	return
}
