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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// ImportAction action for import variables
type ImportAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ImportVariablesRequest
}

// NewImportVariablesAction new import variables action
func NewImportVariablesAction(model store.ProjectModel) *ImportAction {
	return &ImportAction{
		model: model,
	}
}

// Do import variables request
func (ca *ImportAction) Do(ctx context.Context, req *proto.ImportVariablesRequest) error {
	ca.ctx = ctx
	ca.req = req
	var username string
	if authUser, err := middleware.GetUserFromContext(ca.ctx); err == nil {
		username = authUser.GetUsername()
	}
	for _, variable := range ca.req.GetData() {
		definition, err := ca.model.GetVariableDefinitionByKey(ca.ctx, ca.req.GetProjectCode(), variable.Key)
		if err != nil && err != drivers.ErrTableRecordNotFound {
			return err
		}
		// create variable definition if not exists
		if err == drivers.ErrTableRecordNotFound {
			if iErr := ca.createVariableDefinition(definition, variable, username); iErr != nil {
				return iErr
			}
		}
		// update variable definition if exists
		if err == nil {
			// check if old scope equals new
			if uErr := ca.updateVariableDefinition(definition, variable, username); uErr != nil {
				return uErr
			}
		}
		if err := ca.upsertVariableValues(definition, variable, username); err != nil {
			return err
		}
	}
	return nil
}

func (ca *ImportAction) createVariableDefinition(definition *vdm.VariableDefinition,
	variable *proto.ImportVariableData, username string) error {
	definition = &vdm.VariableDefinition{}
	definition.ProjectCode = ca.req.GetProjectCode()
	definition.Key = variable.Key
	definition.Name = variable.Name
	definition.Category = vdm.VariableDefinitionCategoryCustom
	definition.Scope = variable.Scope
	definition.Description = variable.Desc
	definition.Default = variable.Value
	definition.Creator = username
	definition.CreateTime = time.Now().Format(time.RFC3339)
	err := ca.tryGenerateIDAndDoCreate(definition)
	if err != nil {
		logging.Error("try generate id and insert variable definition projectCode [%s], key [%s] failed, err: %s",
			definition.ProjectCode, definition.Key, err.Error())
		return err
	}
	return nil
}

func (ca *ImportAction) updateVariableDefinition(definition *vdm.VariableDefinition,
	variable *proto.ImportVariableData, username string) error {
	// check if old scope equals new
	if definition.Scope != variable.Scope {
		return fmt.Errorf("不能更改原有变量key(%s)的作用范围(%s)",
			variable.Key, variable.Scope)
	}
	definition.Name = variable.Name
	definition.Description = variable.Desc
	definition.Default = variable.Value
	// 覆盖之前导入没有设置类型的变量
	definition.Category = vdm.VariableDefinitionCategoryCustom
	vd := entity.M{
		vdm.FieldKeyID:          definition.ID,
		vdm.FieldKeyName:        variable.Name,
		vdm.FieldKeyDescription: variable.Desc,
		vdm.FieldKeyDefault:     variable.Value,
		vdm.FieldKeyUpdateTime:  time.Now().Format(time.RFC3339),
		vdm.FieldKeyUpdater:     username,
	}
	definition, err := ca.model.UpdateVariableDefinition(ca.ctx, vd)
	if err != nil {
		logging.Error("failed to import variables to db, err: %s",
			err.Error())
		return err
	}
	return nil
}

func (ca *ImportAction) upsertVariableValues(definition *vdm.VariableDefinition,
	variable *proto.ImportVariableData, username string) error {
	for _, entry := range variable.Vars {
		entity := &vvm.VariableValue{
			VariableID: definition.ID,
			Key:        definition.Key,
			Scope:      definition.Scope,
			ClusterID:  entry.ClusterID,
			Namespace:  entry.Namespace,
			Value:      entry.Value,
			UpdateTime: time.Now().Format(time.RFC3339),
			Updater:    username,
		}
		if err := ca.model.UpsertVariableValue(ca.ctx, entity); err != nil {
			logging.Error("failed to import variables to db, err: %s",
				err.Error())
			return err
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
