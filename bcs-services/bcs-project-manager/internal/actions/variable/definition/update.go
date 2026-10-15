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

package definition

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
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
		return nil, errorx.NewDBErr(err.Error())
	}
	return vd, nil
}

func (ca *UpdateAction) updateVariable() (*vdm.VariableDefinition, error) {
	// check if key exists in project
	old, err := ca.model.GetVariableDefinition(ca.ctx, ca.req.GetVariableID())
	if err != nil {
		return nil, err
	}
	if old.Key != ca.req.GetKey() {
		// confim new key available
		_, err := ca.model.GetVariableDefinitionByKey(ca.ctx, ca.req.GetProjectCode(), ca.req.GetKey())
		if err == nil {
			return nil, fmt.Errorf("variable key %s alread exists in project", ca.req.GetKey())
		} else if err != drivers.ErrTableRecordNotFound {
			logging.Error("get variable definition from db failed, err: %s", err.Error())
			return nil, err
		}
	}
	// construct variable definition and update
	vd := entity.M{
		vdm.FieldKeyID:          ca.req.GetVariableID(),
		vdm.FieldKeyKey:         ca.req.GetKey(),
		vdm.FieldKeyDefault:     ca.req.GetDefault(),
		vdm.FieldKeyName:        ca.req.GetName(),
		vdm.FieldKeyDescription: ca.req.GetDesc(),
		vdm.FieldKeyScope:       ca.req.GetScope(),
	}
	if authUser, err := middleware.GetUserFromContext(ca.ctx); err == nil {
		vd[vdm.FieldKeyUpdater] = authUser.GetUsername()
	}
	return ca.model.UpdateVariableDefinition(ca.ctx, vd)
}
