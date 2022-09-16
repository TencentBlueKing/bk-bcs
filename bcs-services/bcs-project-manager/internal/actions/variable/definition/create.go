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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

// CreateAction action for create variable definition
type CreateAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.CreateVariableRequest
}

// NewCreateAction new create variable definition action
func NewCreateAction(model store.ProjectModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

// Do create variable definition request
func (ca *CreateAction) Do(ctx context.Context, req *proto.CreateVariableRequest) (*vdm.VariableDefinition, error) {
	ca.ctx = ctx
	ca.req = req

	vd, err := ca.createVariable()
	if err != nil {
		return nil, errorx.NewDBErr(err)
	}
	return vd, nil
}

func (ca *CreateAction) createVariable() (*vdm.VariableDefinition, error) {
	// check if key exists in project
	_, err := ca.model.GetVariableDefinitionByKey(ca.ctx, ca.req.GetProjectCode(), ca.req.GetKey())
	if err == nil {
		return nil, fmt.Errorf("variable key %s alread exists in project", ca.req.GetKey())
	} else if err != drivers.ErrTableRecordNotFound {
		logging.Error("get variable definition from db failed, err: ", err.Error())
		return nil, err
	}
	// construct variable definition and create
	vd := &vdm.VariableDefinition{
		Key:         ca.req.GetKey(),
		Default:     ca.req.GetDefault(),
		Name:        ca.req.GetName(),
		Description: ca.req.GetDesc(),
		ProjectCode: ca.req.GetProjectCode(),
		Scope:       ca.req.GetScope(),
		Category:    constant.VariableCategoryCustom,
		CreateTime:  time.Now().Format(time.RFC3339),
	}
	if authUser, err := middleware.GetUserFromContext(ca.ctx); err == nil {
		vd.Creator = authUser.GetUsername()
	}
	err = ca.tryGenerateIDAndDoCreate(vd)
	if err != nil {
		return nil, err
	}
	return vd, nil
}

func (ca *CreateAction) tryGenerateIDAndDoCreate(definition *vdm.VariableDefinition) error {
	var count = 3
	var err error
	for i := 0; i < count; i++ {
		definition.ID = stringx.RandomString("variable-", 8)
		err = ca.model.CreateVariableDefinition(ca.ctx, definition)
		if err == nil {
			return nil
		}
		if err != drivers.ErrTableRecordDuplicateKey {
			return err
		}
	}
	return err
}
