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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

// UpdateAction action for update variable definition
type UpdateAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.UpdateVariableRequest
}

// NewUpdateAction new update variable definition action
func NewUpdateAction(model store.ProjectModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

// Do update variable definition request
func (ca *UpdateAction) Do(ctx context.Context, req *proto.UpdateVariableRequest) (*vdm.VariableDefinition, error) {
	ca.ctx = ctx
	ca.req = req

	vd, err := ca.updateVariable()
	if err != nil {
		return nil, errorx.NewDBErr(err)
	}
	return vd, nil
}

func (ca *UpdateAction) updateVariable() (*vdm.VariableDefinition, error) {
	// check if key exists in project
	_, err := ca.model.GetVariableDefinitionByKey(ca.ctx, ca.req.GetProjectCode(), ca.req.GetKey())
	if err == nil {
		return nil, fmt.Errorf("variable key %s alread exists in project", ca.req.GetKey())
	} else if err != drivers.ErrTableRecordNotFound {
		logging.Error("get variable definition from db failed, err: ", err.Error())
		return nil, err
	}
	// construct variable definition and update
	timeStr := time.Now().Format(time.RFC3339)
	// 从 context 中获取 username
	username := auth.GetUserFromCtx(ca.ctx)
	vd := &vdm.VariableDefinition{
		ID:          ca.req.GetVariableID(),
		Key:         ca.req.GetKey(),
		Default:     ca.req.GetDefault(),
		Name:        ca.req.GetName(),
		Description: ca.req.GetDesc(),
		Scope:       ca.req.GetScope(),
		Updater:     username,
		UpdateTime:  timeStr,
	}
	err = ca.model.UpdateVariableDefinition(ca.ctx, vd)
	return vd, nil
}
