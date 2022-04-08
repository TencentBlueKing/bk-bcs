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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
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

// Do get project info
func (ga *GetAction) Do(ctx context.Context, req *proto.GetProjectRequest) (*pm.Project, *errorx.ProjectError) {
	ga.ctx = ctx
	ga.req = req

	p, err := ga.model.GetProject(ctx, req.ProjectIDOrCode)
	if err != nil {
		return nil, errorx.New(errcode.DBErr, errcode.DbErrMsg, err)
	}

	return p, errorx.New(errcode.Success, errcode.SuccessMsg)
}
