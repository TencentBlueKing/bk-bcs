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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ImportAction action for import variables
type ImportAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ImportVariablesRequest
	resp  *proto.ImportVariablesResponse
}

// NewImportVariablesAction new import variables action
func NewImportVariablesAction(model store.ProjectModel) *ImportAction {
	return &ImportAction{
		model: model,
	}
}

// Do import variables request
func (ca *ImportAction) Do(ctx context.Context,
	req *proto.ImportVariablesRequest, resp *proto.ImportVariablesResponse) error {
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp
	var username string
	if authUser, err := middleware.GetUserFromContext(ca.ctx); err == nil {
		username = authUser.GetUsername()
	}
	if err := ca.validateParam(); err != nil {
		return err
	}
	for _, variable := range ca.req.GetData() {
		if _, ok := vdm.SystemVariables[variable.GetKey()]; ok {
			return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("key 不能与系统变量[%s] 相同", variable.GetKey()))
		}
		definition, err := ca.upsertVariableDefinition(variable, username)
		if err != nil {
			return err
		}
		if err := ca.upsertVariableValues(definition, variable, username); err != nil {
			return err
		}
	}
	return nil
}

func (ca *ImportAction) validateParam() error {
	for _, variable := range ca.req.GetData() {
		if _, ok := vdm.SystemVariables[variable.GetKey()]; ok {
			return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("不能与系统变量 key[%s] 重复", variable.GetKey()))
		}
		if !stringx.StringInSlice(variable.GetScope(),
			[]string{vdm.VariableScopeGlobal, vdm.VariableScopeCluster, vdm.VariableScopeNamespace}) {
			return errorx.NewReadableErr(errorx.ParamErr, "作用域只能为 [global,cluster,namespace]")
		}
	}
	return nil
}

func (ca *ImportAction) upsertVariableDefinition(variable *proto.ImportVariableData, username string) (
	*vdm.VariableDefinition, error) {
	definition, err := ca.model.GetVariableDefinitionByKey(ca.ctx, ca.req.GetProjectCode(), variable.Key)
	if err != nil && err != drivers.ErrTableRecordNotFound {
		logging.Error("get variable key %s in project %s failed, err: %s", variable.Key, ca.req.GetProjectCode(), err.Error())
		return nil, errorx.NewDBErr(err.Error())
	}
	if err == drivers.ErrTableRecordNotFound {
		// create if not exists
		newDefinition, cErr := ca.createVariableDefinition(variable, username)
		if cErr != nil {
			return nil, cErr
		}
		definition = newDefinition
	}
	// update if exists
	if err := ca.updateVariableDefinition(definition, variable, username); err != nil {
		return nil, err
	}
	return definition, nil
}

func (ca *ImportAction) createVariableDefinition(variable *proto.ImportVariableData, username string) (
	*vdm.VariableDefinition, error) {
	definition := &vdm.VariableDefinition{}
	definition.ProjectCode = ca.req.GetProjectCode()
	definition.Key = variable.Key
	definition.Name = variable.Name
	definition.Category = vdm.VariableCategoryCustom
	definition.Scope = variable.Scope
	definition.Description = variable.Desc
	definition.Default = variable.Value
	definition.Creator = username
	err := ca.tryGenerateIDAndDoCreate(definition)
	if err != nil {
		logging.Error("try generate id and insert variable definition projectCode [%s], key [%s] failed, err: %s",
			definition.ProjectCode, definition.Key, err.Error())
		return nil, errorx.NewDBErr(err.Error())
	}
	return definition, nil
}

func (ca *ImportAction) updateVariableDefinition(definition *vdm.VariableDefinition,
	variable *proto.ImportVariableData, username string) error {
	// check if old scope equals new
	if definition.Scope != variable.Scope {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("不能更改原有变量key(%s)的作用范围(%s)", variable.Key, variable.Scope))
	}
	definition.Name = variable.Name
	definition.Description = variable.Desc
	definition.Default = variable.Value
	// 覆盖之前导入没有设置类型的变量
	definition.Category = vdm.VariableCategoryCustom
	vd := entity.M{
		vdm.FieldKeyID:          definition.ID,
		vdm.FieldKeyName:        variable.Name,
		vdm.FieldKeyDescription: variable.Desc,
		vdm.FieldKeyDefault:     variable.Value,
		vdm.FieldKeyUpdater:     username,
	}
	definition, err := ca.model.UpdateVariableDefinition(ca.ctx, vd)
	if err != nil {
		logging.Error("update variable definition %s db failed, err: %s", definition.ID, err.Error())
		return errorx.NewDBErr(err.Error())
	}
	return nil
}

func (ca *ImportAction) upsertVariableValues(definition *vdm.VariableDefinition,
	variable *proto.ImportVariableData, username string) error {
	for _, entry := range variable.GetVars() {
		entity := &vvm.VariableValue{
			VariableID: definition.ID,
			Scope:      definition.Scope,
			ClusterID:  entry.ClusterID,
			Namespace:  entry.Namespace,
			Value:      entry.Value,
			UpdateTime: time.Now().Format(time.RFC3339),
			Updater:    username,
		}
		if err := ca.model.UpsertVariableValue(ca.ctx, entity); err != nil {
			logging.Error("upsert variable value %s in db failed, err: %s", entity.VariableID, err.Error())
			return errorx.NewDBErr(err.Error())
		}
	}
	return nil
}

func (ca *ImportAction) tryGenerateIDAndDoCreate(definition *vdm.VariableDefinition) error {
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
