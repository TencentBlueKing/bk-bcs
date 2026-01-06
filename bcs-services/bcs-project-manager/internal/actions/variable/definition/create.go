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

// Package definition xxx
package definition

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// CreateAction action for create variable definition
type CreateAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.CreateVariableRequest
	resp  *proto.CreateVariableResponse
}

// NewCreateAction new create variable definition action
func NewCreateAction(model store.ProjectModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

// Do create variable definition request
func (ca *CreateAction) Do(ctx context.Context,
	req *proto.CreateVariableRequest, resp *proto.CreateVariableResponse) error {
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	err := ca.createVariable()
	if err != nil {
		return err
	}
	return nil
}

func (ca *CreateAction) createVariable() error {
	// check if key is system variables
	if _, ok := vdm.SystemVariables[ca.req.GetKey()]; ok {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("不能与系统变量 key[%s] 重复", ca.req.GetKey()))
	}
	// check if key is valid
	if !stringx.StringInSlice(ca.req.GetScope(),
		[]string{vdm.VariableScopeGlobal, vdm.VariableScopeCluster, vdm.VariableScopeNamespace}) {
		return errorx.NewReadableErr(errorx.ParamErr, "作用域只能为 [global,cluster,namespace]")
	}
	// check if key exists in project
	_, err := ca.model.GetVariableDefinitionByKey(ca.ctx, ca.req.GetProjectCode(), ca.req.GetKey())
	if err == nil {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("变量 key[%s] 在项目中已存在", ca.req.GetKey()))
	} else if err != drivers.ErrTableRecordNotFound {
		logging.Error("get variable definition from db failed, err: %s", err.Error())
		return err
	}
	// construct variable definition and create
	vd := &vdm.VariableDefinition{
		Key:         ca.req.GetKey(),
		Default:     ca.req.GetDefault(),
		Name:        ca.req.GetName(),
		Description: ca.req.GetDesc(),
		ProjectCode: ca.req.GetProjectCode(),
		Scope:       ca.req.GetScope(),
		Category:    vdm.VariableCategoryCustom,
	}
	if authUser, e := middleware.GetUserFromContext(ca.ctx); e == nil {
		vd.Creator = authUser.GetUsername()
	}
	err = ca.tryGenerateIDAndDoCreate(vd)
	if err != nil {
		return err
	}
	ca.resp.Message = "ok"
	ca.resp.Data = &proto.CreateVariableData{
		Id:          vd.ID,
		ProjectCode: vd.ProjectCode,
		Name:        vd.Name,
		Key:         vd.Key,
		Scope:       vd.Scope,
		Default:     vd.Default,
		Desc:        vd.Description,
	}
	return nil
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
