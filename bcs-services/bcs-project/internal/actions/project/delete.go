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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// DeleteAction action for delete project
type DeleteAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.DeleteProjectRequest
}

// NewDeleteAction delete project action
func NewDeleteAction(model store.ProjectModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

// Do delete project
func (da *DeleteAction) Do(ctx context.Context, req *proto.DeleteProjectRequest) *errorx.ProjectError {
	da.ctx = ctx
	da.req = req

	if err := da.model.DeleteProject(ctx, req.ProjectID); err != nil {
		return errorx.New(errcode.DBErr, errcode.DbErrMsg, err)
	}

	return errorx.New(errcode.Success, errcode.SuccessMsg)
}
