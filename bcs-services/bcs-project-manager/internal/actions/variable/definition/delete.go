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

package definition

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// DeleteAction action for delete variable definition
type DeleteAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.DeleteVariableDefinitionsRequest
	resp  *proto.DeleteVariableDefinitionsResponse
}

// NewDeleteAction new delete variable definition action
func NewDeleteAction(model store.ProjectModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

// Do delete variable definition request
func (ca *DeleteAction) Do(ctx context.Context,
	req *proto.DeleteVariableDefinitionsRequest, resp *proto.DeleteVariableDefinitionsResponse) error {
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	total, err := ca.deleteVariable()
	if err != nil {
		return errorx.NewDBErr(err.Error())
	}
	data := &proto.DeleteVariableDefinitionsData{
		Total: uint32(total),
	}
	ca.resp.Code = 0
	ca.resp.Message = "ok"
	ca.resp.Data = data
	return nil
}

func (ca *DeleteAction) deleteVariable() (int64, error) {
	// check if key exists in project
	return ca.model.DeleteVariableDefinitions(ca.ctx, stringx.SplitString(ca.req.GetIdList()))
}
